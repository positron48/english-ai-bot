package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type mockTelegramClient struct {
	lastParams url.Values
	DoCount    int
	LastBody   []byte
}

func (c *mockTelegramClient) Do(req *http.Request) (*http.Response, error) {
	c.DoCount++
	body, _ := io.ReadAll(req.Body)
	_ = req.Body.Close()
	c.LastBody = body
	params, _ := url.ParseQuery(string(body))
	c.lastParams = params

	resp := `{"ok": true, "result": {}}`
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(resp)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

// failingTelegramClient returns an error on Do for coverage of send error paths.
type failingTelegramClient struct{}

func (c *failingTelegramClient) Do(req *http.Request) (*http.Response, error) {
	return nil, io.EOF
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newAIServiceWithResponse(t *testing.T, logger *zap.Logger, content string) *ai.Service {
	t.Helper()
	aiService := ai.NewService("http://example.com", "model", "key", "prompt", logger)
	mockClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			payload, _ := json.Marshal(ai.ChatResponse{
				Choices: []ai.Choice{{Message: ai.Message{Role: "assistant", Content: content}}},
			})
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(payload)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}
	aiValue := reflect.ValueOf(aiService).Elem()
	clientField := aiValue.FieldByName("client")
	if !clientField.IsValid() {
		t.Fatalf("ai service client field not found")
	}
	reflect.NewAt(clientField.Type(), unsafe.Pointer(clientField.UnsafeAddr())).Elem().Set(reflect.ValueOf(mockClient))
	return aiService
}

// newAIServiceWithError returns an AI service whose GenerateResponse will fail (e.g. for handleMessage error path).
func newAIServiceWithError(t *testing.T, logger *zap.Logger) *ai.Service {
	t.Helper()
	aiService := ai.NewService("http://example.com", "model", "key", "prompt", logger)
	mockClient := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("error")),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}
	aiValue := reflect.ValueOf(aiService).Elem()
	clientField := aiValue.FieldByName("client")
	if !clientField.IsValid() {
		t.Fatalf("ai service client field not found")
	}
	reflect.NewAt(clientField.Type(), unsafe.Pointer(clientField.UnsafeAddr())).Elem().Set(reflect.ValueOf(mockClient))
	return aiService
}

func newTestBot(client interface{ Do(*http.Request) (*http.Response, error) }) *tgbotapi.BotAPI {
	bot := &tgbotapi.BotAPI{Token: "test", Client: client, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")
	return bot
}

func commandMessage(text string) *tgbotapi.Message {
	entities := []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(strings.Fields(text)[0])}}
	return &tgbotapi.Message{
		Text:     text,
		Entities: entities,
		Chat:     &tgbotapi.Chat{ID: 10},
		From:     &tgbotapi.User{ID: 42, UserName: "tester"},
	}
}

const testDefaultCourse = "en_ru"

// newHandlerForTest builds a Handler wired with the given AI/word service and a real
// course repository against the test DB, mirroring the production NewHandler signature.
func newHandlerForTest(t *testing.T, client *mockTelegramClient, aiService *ai.Service, wordService *service.WordService, cfg *config.Config) (*Handler, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	courseRepo := repository.NewCourseRepository(db.GetConnection(), logger)
	bot := newTestBot(client)

	wordServices := map[string]*service.WordService{}
	if wordService != nil {
		wordServices[testDefaultCourse] = wordService
	}

	if cfg == nil {
		cfg = &config.Config{}
	}

	h := NewHandler(
		bot,
		logger,
		aiService,
		wordServices,
		testDefaultCourse,
		courseRepo,
		userRepo,
		nil,
		nil,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger),
		cfg,
		db.GetConnection(),
	)
	return h, db
}

func setupHandler(t *testing.T, client *mockTelegramClient) (*Handler, *database.DB) {
	t.Helper()
	cfg := &config.Config{}
	cfg.Bot.StartMessage = "start msg"
	cfg.Bot.HelpMessage = "help msg"
	cfg.Bot.UnknownCommandMessage = "unknown"
	return newHandlerForTest(t, client, nil, nil, cfg)
}

func setupHandlerWithWordService(t *testing.T, client *mockTelegramClient, aiResponse string) (*Handler, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	courseRepo := repository.NewCourseRepository(db.GetConnection(), logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	bot := newTestBot(client)

	aiService := newAIServiceWithResponse(t, logger, aiResponse)
	wordService := service.NewWordService(wordRepo, trainingCardRepo, userCardRepo, aiService, config.DefaultLearningConfig(), logger)

	cfg := &config.Config{}
	cfg.Bot.EmptyMessage = "empty"
	cfg.Bot.ErrorMessage = "error"

	h := NewHandler(
		bot,
		logger,
		aiService,
		map[string]*service.WordService{testDefaultCourse: wordService},
		testDefaultCourse,
		courseRepo,
		userRepo,
		trainingCardRepo,
		userCardRepo,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger),
		cfg,
		db.GetConnection(),
	)
	return h, db
}

func lastText(client *mockTelegramClient) string {
	if client.lastParams == nil {
		return ""
	}
	return client.lastParams.Get("text")
}

// editFailingTelegramClient fails editMessageText but allows other Telegram API calls.
type editFailingTelegramClient struct {
	mockTelegramClient
}

func (c *editFailingTelegramClient) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if strings.Contains(req.URL.Path, "editMessageText") {
		return nil, io.EOF
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return c.mockTelegramClient.Do(req)
}

type stubUserRepo struct {
	getErr error
	user   *models.User
}

func (s *stubUserRepo) GetOrCreateUser(telegramID int64) (*models.User, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.user != nil {
		return s.user, nil
	}
	return &models.User{ID: 1, TelegramID: telegramID}, nil
}

func (s *stubUserRepo) UpdateUsername(int64, string) error { return nil }
func (s *stubUserRepo) UpdateUserSettings(int64, string) error { return nil }

func callbackQuery(data string, userID int64) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{
		ID:   "cb-1",
		Data: data,
		From: &tgbotapi.User{ID: userID},
		Message: &tgbotapi.Message{
			MessageID: 99,
			Chat:      &tgbotapi.Chat{ID: 10},
		},
	}
}

func TestHandleCallbackQuery_SetLanguage(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.handleCallbackQuery(context.Background(), callbackQuery("setlang:en_ru", 42))

	if client.DoCount == 0 {
		t.Fatal("expected Telegram API calls")
	}
}

func TestHandleCallbackQuery_NotificationUnsubscribe(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.handleCallbackQuery(context.Background(), callbackQuery("notification_unsubscribe", 42))

	if client.DoCount == 0 {
		t.Fatal("expected bot command service to call Telegram API")
	}
}

func TestHandleSetLanguage_InvalidCourse(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.handleSetLanguage(context.Background(), callbackQuery("setlang:invalid_course", 42), "invalid_course")

	got := lastText(client)
	if got == "" || !strings.Contains(got, "Не удалось переключить язык") {
		t.Fatalf("got %q", got)
	}
}

func TestHandleSetLanguage_UserRepoError(t *testing.T) {
	client := &mockTelegramClient{}
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	courseRepo := repository.NewCourseRepository(db.GetConnection(), logger)
	cfg := &config.Config{}

	h := NewHandler(
		newTestBot(client),
		logger,
		nil,
		nil,
		testDefaultCourse,
		courseRepo,
		nil,
		nil,
		nil,
		nil,
		cfg,
		db.GetConnection(),
	)
	h.userRepo = &stubUserRepo{getErr: errors.New("db down")}

	h.handleSetLanguage(context.Background(), callbackQuery("setlang:en_ru", 42), "en_ru")

	got := lastText(client)
	if got == "" || !strings.Contains(got, "Произошла ошибка") {
		t.Fatalf("got %q", got)
	}
}

func TestHandleSetLanguage_EditFailureFallsBackToSend(t *testing.T) {
	client := &editFailingTelegramClient{}
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	userRepo := repository.NewUserRepository(conn, logger)
	courseRepo := repository.NewCourseRepository(conn, logger)
	cfg := &config.Config{}

	h := NewHandler(
		newTestBot(client),
		logger,
		nil,
		nil,
		testDefaultCourse,
		courseRepo,
		userRepo,
		nil,
		nil,
		nil,
		cfg,
		conn,
	)

	h.handleSetLanguage(context.Background(), callbackQuery("setlang:en_ru", 42), "en_ru")

	got := lastText(&client.mockTelegramClient)
	if got == "" || !strings.Contains(got, "Язык переключён") {
		t.Fatalf("got %q", got)
	}
}

func TestHandleResetCircuitCommand(t *testing.T) {
	t.Run("non admin silent", func(t *testing.T) {
		client := &mockTelegramClient{}
		cfg := &config.Config{}
		cfg.Admin.TelegramID = 999
		h, _ := newHandlerForTest(t, client, nil, nil, cfg)

		h.handleResetCircuitCommand(10, 42)

		if client.DoCount != 0 {
			t.Fatalf("non-admin should not send messages, DoCount=%d", client.DoCount)
		}
	})

	t.Run("admin success", func(t *testing.T) {
		client := &mockTelegramClient{}
		cfg := &config.Config{}
		cfg.Admin.TelegramID = 42
		h, _ := newHandlerForTest(t, client, nil, nil, cfg)

		h.handleResetCircuitCommand(10, 42)

		got := lastText(client)
		if got == "" || !strings.Contains(got, "Circuit breaker сброшен") {
			t.Fatalf("got %q", got)
		}
	})
}

func TestLanguageFlag(t *testing.T) {
	if languageFlag("en") == "" || languageFlag("es") == "" {
		t.Fatal("expected flags for en/es")
	}
	if languageFlag("de") != "" {
		t.Fatalf("unexpected flag for de: %q", languageFlag("de"))
	}
}

func TestHandleUpdate_StartCommand(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.HandleUpdate(context.Background(), tgbotapi.Update{Message: commandMessage("/start")})

	if got := lastText(client); got != "start msg" {
		t.Fatalf("/start: got %q, want %q", got, "start msg")
	}
}

func TestHandleUpdate_HelpCommand(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.HandleUpdate(context.Background(), tgbotapi.Update{Message: commandMessage("/help")})

	if got := lastText(client); got != "help msg" {
		t.Fatalf("/help: got %q, want %q", got, "help msg")
	}
}

func TestHandleUpdate_UnknownCommand(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.HandleUpdate(context.Background(), tgbotapi.Update{Message: commandMessage("/bogus")})

	if got := lastText(client); got != "unknown" {
		t.Fatalf("/bogus: got %q, want %q", got, "unknown")
	}
}

func TestHandleUpdate_EmptyMessage(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithWordService(t, client, "")

	msg := &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 10}, From: &tgbotapi.User{ID: 42}}
	h.HandleUpdate(context.Background(), tgbotapi.Update{Message: msg})

	if got := lastText(client); got != "empty" {
		t.Fatalf("empty message: got %q, want %q", got, "empty")
	}
}

func TestHandleLanguageCommand_ShowsCourses(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.handleLanguageCommand(context.Background(), 10, 42)

	// Either the course menu prompt or the "no languages" fallback is acceptable depending
	// on whether the test DB seeds courses; both confirm the command runs without panic.
	got := lastText(client)
	if got == "" {
		t.Fatal("/language: expected a response message")
	}
	if !strings.Contains(got, "язык") && !strings.Contains(got, "Сейчас нет") {
		t.Fatalf("/language: unexpected response %q", got)
	}
}

func TestResolveUserCourse_FallsBackToDefault(t *testing.T) {
	client := &mockTelegramClient{}
	logger, _ := zap.NewDevelopment()

	ws := service.NewWordService(nil, nil, nil, nil, config.DefaultLearningConfig(), logger)
	h, _ := newHandlerForTest(t, client, nil, ws, nil)

	code, resolved := h.resolveUserCourse(context.Background(), 0)
	if code != testDefaultCourse {
		t.Fatalf("resolveUserCourse: got course %q, want %q", code, testDefaultCourse)
	}
	if resolved != ws {
		t.Fatal("resolveUserCourse: expected default word service")
	}
}

func TestCourseDisplayName(t *testing.T) {
	got := courseDisplayName(repository.CourseSummary{Title: "English RU", TargetLanguage: "en"})
	if !strings.Contains(got, "English RU") {
		t.Fatalf("courseDisplayName: got %q", got)
	}

	// Empty title falls back to code.
	got = courseDisplayName(repository.CourseSummary{Code: "es_ru", TargetLanguage: "es"})
	if !strings.Contains(got, "es_ru") {
		t.Fatalf("courseDisplayName fallback: got %q", got)
	}
}

func TestResolveUserCourse_PreservesSelectedCourseWhenServiceMissing(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	userRepo := repository.NewUserRepository(conn, logger)
	courseRepo := repository.NewCourseRepository(conn, logger)
	ws := service.NewWordService(nil, nil, nil, nil, config.DefaultLearningConfig(), logger)

	user, err := userRepo.GetOrCreateUser(8001)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	if _, err := courseRepo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	h := NewHandler(
		nil, logger, nil,
		map[string]*service.WordService{testDefaultCourse: ws},
		testDefaultCourse,
		courseRepo, userRepo, nil, nil, nil, nil, conn,
	)

	code, resolved := h.resolveUserCourse(context.Background(), user.ID)
	if code != "es_ru" {
		t.Fatalf("resolveUserCourse course: got %q, want es_ru", code)
	}
	if resolved != ws {
		t.Fatal("resolveUserCourse: expected default word service fallback")
	}
}

func TestHandleMessage_SingleWordLookup_UsesSelectedCourse(t *testing.T) {
	client := &mockTelegramClient{}
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	userRepo := repository.NewUserRepository(conn, logger)
	courseRepo := repository.NewCourseRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	bot := newTestBot(client)

	user, err := userRepo.GetOrCreateUser(9001)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	if _, err := courseRepo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	posNoun := "noun"
	defAlgo := "алгоритм"
	displayAlgorithm := "algorithm"
	algoCardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word:         "algorithm",
		CourseCode:   "en_ru",
		POS:          &posNoun,
		DefinitionRU: &defAlgo,
		DisplayEN:    &displayAlgorithm,
	})
	if err != nil {
		t.Fatalf("upsert algorithm: %v", err)
	}
	if err := wordRepo.UpsertWordFormMapping("algo", algoCardID); err != nil {
		t.Fatalf("UpsertWordFormMapping algo->algorithm: %v", err)
	}

	defSomething := "что-то"
	displayAlgo := "algo"
	_, err = wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word:         "algo",
		CourseCode:   "es_ru",
		POS:          &posNoun,
		DefinitionRU: &defSomething,
		DisplayEN:    &displayAlgo,
	})
	if err != nil {
		t.Fatalf("upsert es algo: %v", err)
	}

	// Only en_ru WordService registered — lookup must still honor user's es_ru selection.
	wordService := service.NewWordService(wordRepo, nil, nil, nil, config.DefaultLearningConfig(), logger)

	cfg := &config.Config{}
	cfg.Bot.ErrorMessage = "error"

	h := NewHandler(
		bot, logger, nil,
		map[string]*service.WordService{testDefaultCourse: wordService},
		testDefaultCourse,
		courseRepo, userRepo, nil, nil,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(conn, logger), 5, logger),
		cfg, conn,
	)

	msg := &tgbotapi.Message{
		Text: "algo",
		Chat: &tgbotapi.Chat{ID: 10},
		From: &tgbotapi.User{ID: 9001},
	}
	h.handleMessage(context.Background(), msg)

	got := lastText(client)
	if got == "" {
		t.Fatal("expected bot response")
	}
	if strings.Contains(got, "алгоритм") || strings.Contains(strings.ToLower(got), "algorithm") {
		t.Fatalf("expected Spanish course card, got English algorithm: %q", got)
	}
	if !strings.Contains(got, "что-то") {
		t.Fatalf("expected Spanish definition, got %q", got)
	}
}

func TestHandleMessage_AIErrorSendsErrorMessage(t *testing.T) {
	client := &mockTelegramClient{}
	logger, _ := zap.NewDevelopment()

	aiService := newAIServiceWithError(t, logger)
	ws := service.NewWordService(nil, nil, nil, aiService, config.DefaultLearningConfig(), logger)
	cfg := &config.Config{}
	cfg.Bot.ErrorMessage = "error"

	h, _ := newHandlerForTest(t, client, aiService, ws, cfg)

	// Multi-word text takes the AI (sentence-correction) path.
	msg := &tgbotapi.Message{Text: "this is a sentence", Chat: &tgbotapi.Chat{ID: 10}, From: &tgbotapi.User{ID: 42}}
	h.handleMessage(context.Background(), msg)

	if got := lastText(client); got != "error" {
		t.Fatalf("AI error path: got %q, want %q", got, "error")
	}
}
