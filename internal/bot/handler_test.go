package bot

import (
	"bytes"
	"context"
	"encoding/json"
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

func newTestBot(client *mockTelegramClient) *tgbotapi.BotAPI {
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
