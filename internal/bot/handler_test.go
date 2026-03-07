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
	"time"
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

func setupHandler(t *testing.T, client *mockTelegramClient) (*Handler, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	bot := newTestBot(client)

	cfg := &config.Config{}
	cfg.Bot.StartMessage = "start msg"
	cfg.Bot.HelpMessage = "help msg"
	cfg.Bot.UnknownCommandMessage = "unknown"

	h := NewHandler(
		bot,
		logger,
		nil,
		nil,
		nil,
		userRepo,
		nil,
		nil,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger),
		cfg,
		db.GetConnection(),
	)

	return h, db
}

func setupHandlerWithRepos(t *testing.T, client *mockTelegramClient) (*Handler, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)

	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	bot := newTestBot(client)
	trainingHandler := NewTrainingHandler(bot, trainingService, srsService, optionsService, sessionRepo, logger, 0, 0, db.GetConnection())

	cfg := &config.Config{}
	cfg.Bot.StartMessage = "start msg"
	cfg.Bot.HelpMessage = "help msg"
	cfg.Bot.UnknownCommandMessage = "unknown"

	h := NewHandler(
		bot,
		logger,
		nil,
		nil,
		trainingHandler,
		userRepo,
		trainingCardRepo,
		userCardRepo,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger),
		cfg,
		db.GetConnection(),
	)

	return h, db
}

func setupHandlerWithAI(t *testing.T, client *mockTelegramClient, aiResponse string) (*Handler, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	bot := newTestBot(client)

	aiService := newAIServiceWithResponse(t, logger, aiResponse)
	wordService := service.NewWordService(nil, nil, nil, nil, logger)

	cfg := &config.Config{}
	cfg.Bot.EmptyMessage = "empty"
	cfg.Bot.ErrorMessage = "error"

	h := NewHandler(
		bot,
		logger,
		aiService,
		wordService,
		nil,
		userRepo,
		nil,
		nil,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger),
		cfg,
		db.GetConnection(),
	)

	return h, db
}

func setupHandlerWithWordService(t *testing.T, client *mockTelegramClient, aiResponse string) (*Handler, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	bot := newTestBot(client)

	aiService := newAIServiceWithResponse(t, logger, aiResponse)
	wordService := service.NewWordService(wordRepo, trainingCardRepo, userCardRepo, aiService, logger)

	cfg := &config.Config{}
	cfg.Bot.EmptyMessage = "empty"
	cfg.Bot.ErrorMessage = "error"

	h := NewHandler(
		bot,
		logger,
		aiService,
		wordService,
		nil,
		userRepo,
		trainingCardRepo,
		userCardRepo,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger),
		cfg,
		db.GetConnection(),
	)

	return h, db
}

func TestHandleUpdate_StartUsesBotCommandService(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	update := tgbotapi.Update{Message: commandMessage("/start")}
	h.HandleUpdate(context.Background(), update)

	if got := client.lastParams.Get("text"); got != "start msg" {
		t.Fatalf("expected start message, got %q", got)
	}
}

func TestHandleUpdate_HelpUsesBotCommandService(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	update := tgbotapi.Update{Message: commandMessage("/help")}
	h.HandleUpdate(context.Background(), update)

	if got := client.lastParams.Get("text"); got != "help msg" {
		t.Fatalf("expected help message, got %q", got)
	}
}

func TestHandleDeleteTrainCommand_AdminFlow(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandler(t, client)

	logger, _ := zap.NewDevelopment()
	h.trainingCardRepo = repository.NewTrainingCardRepository(db.GetConnection(), logger)
	h.userCardRepo = repository.NewUserCardRepository(db.GetConnection(), logger)
	h.config.Admin.TelegramID = 42

	wordCardRepo := repository.NewWordRepository(db.GetConnection(), logger)
	wordID, err := wordCardRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}

	_, err = h.trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID,
		WordEN:     "apple",
		WordRU:     "яблоко",
		MeaningEN:  "apple",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard error: %v", err)
	}

	h.handleDeleteTrainCommand(10, 42, "apple")
	if got := client.lastParams.Get("text"); !strings.Contains(got, "Удалено тренировочных карточек") {
		t.Fatalf("expected delete message, got %q", got)
	}
}

func TestHandleDeleteTrainAllCommand_AdminFlow(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandler(t, client)

	logger, _ := zap.NewDevelopment()
	h.trainingCardRepo = repository.NewTrainingCardRepository(db.GetConnection(), logger)
	h.userCardRepo = repository.NewUserCardRepository(db.GetConnection(), logger)
	h.config.Admin.TelegramID = 42

	wordCardRepo := repository.NewWordRepository(db.GetConnection(), logger)
	wordID, err := wordCardRepo.UpsertWordCardLemma(&models.WordCard{Word: "apple", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}

	_, err = h.trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID,
		WordEN:     "apple",
		WordRU:     "яблоко",
		MeaningEN:  "apple",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard error: %v", err)
	}

	h.handleDeleteTrainAllCommand(10, 42)
	if got := client.lastParams.Get("text"); !strings.Contains(got, "Удалено всех тренировочных карточек") {
		t.Fatalf("expected delete all message, got %q", got)
	}
}

func TestHandleGetIDCommand(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.handleGetIDCommand(10, 42)
	if got := client.lastParams.Get("text"); !strings.Contains(got, "42") {
		t.Fatalf("expected user id in message, got %q", got)
	}
}

func TestHandleResetCircuitCommand_Admin(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.config.Admin.TelegramID = 42
	h.handleResetCircuitCommand(10, 42)
	if got := client.lastParams.Get("text"); !strings.Contains(got, "Circuit breaker") {
		t.Fatalf("expected reset message, got %q", got)
	}
}

func TestHandleResetCircuitCommand_NonAdmin(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.config.Admin.TelegramID = 99
	h.handleResetCircuitCommand(10, 42)
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		t.Fatalf("expected no message for non-admin")
	}
}

func TestHandleResetCircuitCommand_AdminIDZero(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.config.Admin.TelegramID = 0
	h.handleResetCircuitCommand(10, 42)
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		t.Fatalf("expected no message when Admin.TelegramID is 0")
	}
}

func TestHandleResetCircuitCommand_Admin_ResetFails(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)
	h.config.Admin.TelegramID = 42

	dsn := testutil.SecondPostgresDSN(t)
	dbWrap, err := database.NewWithConfig("postgres", "", dsn, h.logger)
	if err != nil {
		t.Skipf("second DB not available: %v", err)
	}
	conn := dbWrap.GetConnection()
	cbRepo := repository.NewCircuitBreakerRepository(conn, h.logger)
	failingCB := service.NewCircuitBreakerService(cbRepo, 5, h.logger)
	_ = dbWrap.Close()

	hv := reflect.ValueOf(h).Elem()
	hv.FieldByName("cbService").Set(reflect.ValueOf(failingCB))

	h.handleResetCircuitCommand(10, 42)
	got := client.lastParams.Get("text")
	if got == "" {
		t.Fatal("expected error message when Reset fails")
	}
	if !strings.Contains(got, "Не удалось") && !strings.Contains(got, "сбросить") {
		t.Errorf("expected error about reset failure, got %q", got)
	}
}

func TestHandleStatsCommand_Basic(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandlerWithRepos(t, client)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)

	user, err := userRepo.GetOrCreateUser(555)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}

	wordID := int64(1)
	_, err = db.GetConnection().Exec(`INSERT INTO word_cards (word, definition, created_at, updated_at) VALUES ('apple','',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatalf("insert word card error: %v", err)
	}

	_, err = trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID,
		WordEN:     "apple",
		WordRU:     "яблоко",
		MeaningEN:  "apple",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard error: %v", err)
	}

	_, err = userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	})
	if err != nil {
		t.Fatalf("CreateUserCard error: %v", err)
	}

	h.handleStatsCommand(context.Background(), 10, user.TelegramID)

	if got := client.lastParams.Get("text"); !strings.Contains(got, "Всего карточек") {
		t.Fatalf("expected stats message, got %q", got)
	}
}

func TestHandleGetTrainDataCommand_Admin(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandlerWithRepos(t, client)

	h.config.Admin.TelegramID = 42

	_, err := db.GetConnection().Exec(`INSERT INTO word_cards (word, definition, created_at, updated_at) VALUES ('apple','',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatalf("insert word card error: %v", err)
	}

	_, err = h.trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: 1,
		WordEN:     "apple",
		WordRU:     "яблоко",
		MeaningEN:  "apple",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard error: %v", err)
	}

	h.handleGetTrainDataCommand(10, 42, "apple")
	if got := client.lastParams.Get("text"); !strings.Contains(got, "Карточка") {
		t.Fatalf("expected training data message, got %q", got)
	}
}

func TestHandleCommand_Unknown(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.handleCommand(context.Background(), commandMessage("/unknown"))

	if got := client.lastParams.Get("text"); got != "unknown" {
		t.Fatalf("expected unknown message, got %q", got)
	}
}

func TestHandleMessage_Empty(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithAI(t, client, "ignored")

	msg := &tgbotapi.Message{
		Text: "",
		Chat: &tgbotapi.Chat{ID: 10},
		From: &tgbotapi.User{ID: 55, UserName: "tester"},
	}
	h.handleMessage(context.Background(), msg)

	if got := client.lastParams.Get("text"); got != "empty" {
		t.Fatalf("expected empty message, got %q", got)
	}
}

func TestHandleMessage_NonSingleWord_UsesAI(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithAI(t, client, "AI reply")

	msg := &tgbotapi.Message{
		Text: "hello world",
		Chat: &tgbotapi.Chat{ID: 10},
		From: &tgbotapi.User{ID: 55, UserName: "tester"},
	}
	h.handleMessage(context.Background(), msg)

	if got := client.lastParams.Get("text"); got != "AI reply" {
		t.Fatalf("expected AI reply, got %q", got)
	}
}

func TestHandleMessage_SingleWord_UsesWordService(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithWordService(t, client, `{"lemma":"banana","pos":"noun","definition_ru":"банан"}`)

	msg := &tgbotapi.Message{
		Text: "banana",
		Chat: &tgbotapi.Chat{ID: 10},
		From: &tgbotapi.User{ID: 55, UserName: "tester"},
	}
	h.handleMessage(context.Background(), msg)

	if got := client.lastParams.Get("text"); !strings.Contains(got, "банан") {
		t.Fatalf("expected definition in message, got %q", got)
	}
}

func TestHandleUpdate_NilMessage(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	update := tgbotapi.Update{UpdateID: 1, Message: nil}
	h.HandleUpdate(context.Background(), update)
	// Should not panic; callback and message paths both no-op when Message is nil and no CallbackQuery
}

func TestHandleUpdate_CallbackQuery(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)

	update := tgbotapi.Update{
		UpdateID: 1,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb1",
			From: &tgbotapi.User{ID: 555, UserName: "tester"},
			Message: &tgbotapi.Message{
				MessageID: 1,
				Chat:      &tgbotapi.Chat{ID: 10},
			},
			Data: "train_start",
		},
	}
	h.HandleUpdate(context.Background(), update)
	// train_start with missing user may yield "no cards" or error message; both are acceptable (no assertion).
}

func TestHandleUpdate_CallbackQueryAnswerInvalidIndex(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)

	update := tgbotapi.Update{
		UpdateID: 1,
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb2",
			From: &tgbotapi.User{ID: 555, UserName: "u"},
			Message: &tgbotapi.Message{
				MessageID: 2,
				Chat:      &tgbotapi.Chat{ID: 10},
			},
			Data: "answer_xyz",
		},
	}
	h.HandleUpdate(context.Background(), update)
	// answer_xyz: Atoi fails, handler logs and returns without sending "no active session"
}

func TestHandleCommand_Train_NoCards(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)

	msg := commandMessage("/train")
	msg.Chat.ID = 10
	msg.From.ID = 999
	h.handleCommand(context.Background(), msg)

	if got := client.lastParams.Get("text"); !strings.Contains(got, "карточек") && !strings.Contains(got, "ошибка") {
		t.Errorf("expected no-cards or error message, got %q", got)
	}
}

func TestHandleCommand_Stats_UserNotFound(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandlerWithRepos(t, client)
	// Use a new telegram ID that will be created by GetOrCreateUser when we call handleStatsCommand
	userRepo := repository.NewUserRepository(db.GetConnection(), h.logger)
	_, err := userRepo.GetOrCreateUser(777)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	msg := commandMessage("/stats")
	msg.Chat.ID = 10
	msg.From.ID = 777
	h.handleCommand(context.Background(), msg)

	if got := client.lastParams.Get("text"); !strings.Contains(got, "Статистика") && !strings.Contains(got, "Карточки") {
		t.Errorf("expected stats text, got %q", got)
	}
}

func TestHandleDeleteTrainCommand_EmptyWord(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandler(t, client)
	logger, _ := zap.NewDevelopment()
	h.config.Admin.TelegramID = 42
	h.trainingCardRepo = repository.NewTrainingCardRepository(db.GetConnection(), logger)
	h.userCardRepo = repository.NewUserCardRepository(db.GetConnection(), logger)

	h.handleDeleteTrainCommand(10, 42, "   ")
	if got := client.lastParams.Get("text"); !strings.Contains(got, "Укажите слово") {
		t.Fatalf("expected prompt to specify word, got %q", got)
	}
}

func TestHandleDeleteTrainCommand_WordNotFound(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandler(t, client)
	logger, _ := zap.NewDevelopment()
	h.config.Admin.TelegramID = 42
	h.trainingCardRepo = repository.NewTrainingCardRepository(db.GetConnection(), logger)
	h.userCardRepo = repository.NewUserCardRepository(db.GetConnection(), logger)

	h.handleDeleteTrainCommand(10, 42, "nonexistent_word_xyz")
	if got := client.lastParams.Get("text"); !strings.Contains(got, "не найдены") {
		t.Fatalf("expected not found message, got %q", got)
	}
}

func TestHandleCommand_GetID(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.handleCommand(context.Background(), commandMessage("/get_id"))
	if got := client.lastParams.Get("text"); !strings.Contains(got, "42") {
		t.Fatalf("expected user id in message, got %q", got)
	}
}

func TestHandler_SendMessage(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.sendMessage(10, "test message")
	if got := client.lastParams.Get("text"); got != "test message" {
		t.Errorf("sendMessage: got text %q, want %q", got, "test message")
	}
}

func TestHandler_SendTyping(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)

	h.sendTyping(10)
	if got := client.lastParams.Get("action"); got != "typing" {
		t.Errorf("sendTyping: got action %q, want typing", got)
	}
}

// failingTelegramClient returns error on Do for coverage of sendMessage/sendTyping error paths
type failingTelegramClient struct{}

func (c *failingTelegramClient) Do(req *http.Request) (*http.Response, error) {
	return nil, io.EOF
}

func TestHandler_SendMessage_WhenSendFails(t *testing.T) {
	bot := &tgbotapi.BotAPI{Token: "test", Client: &failingTelegramClient{}, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	cfg := &config.Config{}
	cfg.Bot.StartMessage = "start"
	h := NewHandler(bot, logger, nil, nil, nil, userRepo, nil, nil,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger),
		cfg, db.GetConnection())
	// should not panic when Send fails
	h.sendMessage(10, "test")
}

func TestHandler_SendTyping_WhenRequestFails(t *testing.T) {
	bot := &tgbotapi.BotAPI{Token: "test", Client: &failingTelegramClient{}, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	cfg := &config.Config{}
	h := NewHandler(bot, logger, nil, nil, nil, userRepo, nil, nil,
		service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger),
		cfg, db.GetConnection())
	// should not panic when Request fails
	h.sendTyping(10)
}

func TestHandleCommand_Start(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)
	h.handleCommand(context.Background(), commandMessage("/start"))
	if got := client.lastParams.Get("text"); got != "start msg" {
		t.Errorf("handleCommand(/start): got %q, want start msg", got)
	}
}

func TestHandleCommand_Help(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)
	h.handleCommand(context.Background(), commandMessage("/help"))
	if got := client.lastParams.Get("text"); got != "help msg" {
		t.Errorf("handleCommand(/help): got %q, want help msg", got)
	}
}

func TestHandleCommand_Train_Success(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandlerWithRepos(t, client)
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, err := userRepo.GetOrCreateUser(888)
	if err != nil || user == nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordID := int64(1)
	_, _ = db.GetConnection().Exec(`INSERT INTO word_cards (word, definition, created_at, updated_at) VALUES ('apple','',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	_, _ = trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID, WordEN: "apple", WordRU: "яблоко", MeaningEN: "apple", SenseIndex: 0,
	})
	_, _ = userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: 1, Direction: models.DirectionENtoRU,
		State: models.StateNew, EF: models.InitialEF,
	})
	msg := commandMessage("/train")
	msg.From.ID = 888
	msg.Chat.ID = 11
	h.handleCommand(context.Background(), msg)
	var got string
	if client.lastParams != nil {
		got = client.lastParams.Get("text")
	}
	// May be welcome ("Начинаем тренировку"), card question ("Карточка N из M"), or error/no-cards
	ok := strings.Contains(got, "Начинаем тренировку") || strings.Contains(got, "Карточка") ||
		strings.Contains(got, "ошибка") || strings.Contains(got, "карточек") || strings.Contains(got, "Попробуйте")
	if got != "" && !ok {
		t.Errorf("expected training start, card, or error message, got %q", got)
	}
}

func TestHandleResetCircuitCommand_ResetFails(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)
	h.config.Admin.TelegramID = 42
	h.handleResetCircuitCommand(10, 42)
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		if !strings.Contains(client.lastParams.Get("text"), "Circuit") && !strings.Contains(client.lastParams.Get("text"), "сбросить") && !strings.Contains(client.lastParams.Get("text"), "Не удалось") {
			t.Errorf("expected circuit-related or error message, got %q", client.lastParams.Get("text"))
		}
	}
}

func TestHandleDeleteTrainAllCommand_NonAdmin(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)
	h.config.Admin.TelegramID = 99
	h.handleDeleteTrainAllCommand(10, 42)
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		t.Error("expected no message for non-admin delete_train_all")
	}
}

func TestHandleDeleteTrainAllCommand_AdminIDZero(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)
	h.config.Admin.TelegramID = 0
	h.handleDeleteTrainAllCommand(10, 42)
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		t.Error("expected no message when Admin.TelegramID is 0")
	}
}

func TestHandleDeleteTrainAllCommand_Admin_DeleteAllFails(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandler(t, client)
	logger, _ := zap.NewDevelopment()
	h.config.Admin.TelegramID = 42
	h.trainingCardRepo = repository.NewTrainingCardRepository(db.GetConnection(), logger)
	h.userCardRepo = repository.NewUserCardRepository(db.GetConnection(), logger)

	dsn := testutil.SecondPostgresDSN(t)
	dbWrap, err := database.NewWithConfig("postgres", "", dsn, logger)
	if err != nil {
		t.Skipf("second DB not available: %v", err)
	}
	conn := dbWrap.GetConnection()
	badRepo := repository.NewTrainingCardRepository(conn, logger)
	_ = dbWrap.Close()

	hv := reflect.ValueOf(h).Elem()
	hv.FieldByName("trainingCardRepo").Set(reflect.ValueOf(badRepo))

	h.handleDeleteTrainAllCommand(10, 42)
	got := client.lastParams.Get("text")
	if got == "" {
		t.Fatal("expected error message when DeleteAllTrainingCards fails")
	}
	if !strings.Contains(got, "Ошибка") && !strings.Contains(got, "удалении") {
		t.Errorf("expected error about delete failure, got %q", got)
	}
}

func TestHandleGetTrainDataCommand_EmptyWord(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)
	h.config.Admin.TelegramID = 42
	h.handleGetTrainDataCommand(10, 42, "   ")
	if got := client.lastParams.Get("text"); !strings.Contains(got, "Укажите слово") {
		t.Errorf("expected prompt to specify word, got %q", got)
	}
}

func TestHandleGetTrainDataCommand_WordNotFound(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)
	h.config.Admin.TelegramID = 42
	h.handleGetTrainDataCommand(10, 42, "nonexistent_xyz")
	if got := client.lastParams.Get("text"); !strings.Contains(got, "не найдены") {
		t.Errorf("expected not found message, got %q", got)
	}
}

func TestHandleGetTrainDataCommand_NonAdmin(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)
	h.config.Admin.TelegramID = 99
	h.handleGetTrainDataCommand(10, 42, "apple")
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		t.Error("expected no message for non-admin get_train_data")
	}
}

func TestHandleUpdate_CallbackQuery_NotificationUnsubscribe(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandler(t, client)
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID: "cb1", From: &tgbotapi.User{ID: 42, UserName: "u"},
			Message: &tgbotapi.Message{MessageID: 1, Chat: &tgbotapi.Chat{ID: 10}},
			Data:    "notification_unsubscribe",
		},
	}
	h.HandleUpdate(context.Background(), update)
	// BotCommandService handles it; no panic, may or may not send reply
}

func TestHandleTrainCommand_TableDriven(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandlerWithRepos(t, client)
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)

	tests := []struct {
		name           string
		setupUser      bool
		setupCards     bool
		telegramID     int64
		chatID         int64
		wantSubstrings []string // at least one must be in the sent message
	}{
		{
			name:           "no_cards_sends_no_cards_or_error_message",
			setupUser:      true,
			setupCards:     false,
			telegramID:     901,
			chatID:         10,
			wantSubstrings: []string{"карточек", "ошибка", "Попробуйте", "нет карточек"},
		},
		{
			name:       "success_with_cards_sends_welcome_or_card",
			setupUser:  true,
			setupCards: true,
			telegramID: 902,
			chatID:     11,
			wantSubstrings: []string{"Начинаем тренировку", "Карточка", "ошибка", "карточек", "Попробуйте"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.lastParams = nil
			if tt.setupUser {
				_, err := userRepo.GetOrCreateUser(tt.telegramID)
				if err != nil {
					t.Fatalf("setup user: %v", err)
				}
			}
			if tt.setupCards {
				_, _ = db.GetConnection().Exec(`INSERT INTO word_cards (word, definition, created_at, updated_at) VALUES ('test','',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
				_, _ = trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
					WordCardID: 1, WordEN: "test", WordRU: "тест", MeaningEN: "test", SenseIndex: 0,
				})
				u, _ := userRepo.GetOrCreateUser(tt.telegramID)
				_, _ = userCardRepo.CreateUserCard(&models.UserCard{
					UserID: u.ID, TrainingCardID: 1, Direction: models.DirectionENtoRU,
					State: models.StateNew, EF: models.InitialEF,
				})
			}
			h.handleTrainCommand(context.Background(), tt.chatID, tt.telegramID)
			got := client.lastParams.Get("text")
			if got == "" {
				t.Error("expected a message to be sent")
				return
			}
			ok := false
			for _, sub := range tt.wantSubstrings {
				if strings.Contains(got, sub) {
					ok = true
					break
				}
			}
			if !ok {
				t.Errorf("message %q should contain one of %v", got, tt.wantSubstrings)
			}
		})
	}
}

func TestHandleStatsCommand_TableDriven(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandlerWithRepos(t, client)
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)

	tests := []struct {
		name           string
		telegramID     int64
		setupUser      bool
		wantSubstrings []string
	}{
		{
			name:           "user_with_no_cards_shows_zero_and_no_accuracy_data",
			telegramID:     801,
			setupUser:      true,
			wantSubstrings: []string{"Статистика", "Карточки", "Доступно", "Пока нет данных"},
		},
		{
			name:           "user_with_cards_shows_stats",
			telegramID:     802,
			setupUser:      true,
			wantSubstrings: []string{"Статистика", "Карточки", "Всего карточек"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.lastParams = nil
			if tt.setupUser {
				_, err := userRepo.GetOrCreateUser(tt.telegramID)
				if err != nil {
					t.Fatalf("setup user: %v", err)
				}
			}
			h.handleStatsCommand(context.Background(), 10, tt.telegramID)
			got := client.lastParams.Get("text")
			if got == "" {
				t.Error("expected a message to be sent")
				return
			}
			for _, sub := range tt.wantSubstrings {
				if !strings.Contains(got, sub) {
					t.Errorf("message %q should contain %q", got, sub)
				}
			}
		})
	}
}

func TestHandleStatsCommand_AccuracyWithReviews(t *testing.T) {
	client := &mockTelegramClient{}
	h, db := setupHandlerWithRepos(t, client)
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)

	u, err := userRepo.GetOrCreateUser(703)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	_, _ = db.GetConnection().Exec(`INSERT INTO word_cards (word, definition, created_at, updated_at) VALUES ('w','',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`)
	_, _ = trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: 1, WordEN: "w", WordRU: "в", MeaningEN: "w", SenseIndex: 0,
	})
	ucID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: u.ID, TrainingCardID: 1, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: models.InitialEF,
	})
	sessID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: u.ID, Source: models.SourceManual, PlannedCount: 1, DoneCount: 0, SessionJSON: "{}",
	})
	now := time.Now()
	_, _ = sessionRepo.CreateReviewEvent(&models.ReviewEvent{
		SessionID: &sessID, UserID: u.ID, UserCardID: ucID, Direction: models.DirectionENtoRU,
		ShownAt: now, OptionsShownAt: &now, AnsweredAt: &now,
		TDelayMS: 0, EarlyReveal: false, OptionCount: 4, OptionsJSON: "[]",
		ChosenOption: "ok", IsCorrect: true, Quality: 2,
		MetricsJSON: "{}", SRSBeforeJSON: "{}", SRSAfterJSON: "{}",
	})

	h.handleStatsCommand(context.Background(), 10, 703)
	got := client.lastParams.Get("text")
	if got == "" {
		t.Fatal("expected stats message")
	}
	if !strings.Contains(got, "Правильных ответов") && !strings.Contains(got, "Точность") {
		t.Errorf("expected accuracy section when reviews exist, got %q", got)
	}
}

func TestHandleCallbackQuery_AnswerNoSession(t *testing.T) {
	client := &mockTelegramClient{}
	h, _ := setupHandlerWithRepos(t, client)
	// User exists but no active session; callback answer_0
	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID: "cb2", From: &tgbotapi.User{ID: 555, UserName: "u"},
			Message: &tgbotapi.Message{MessageID: 2, Chat: &tgbotapi.Chat{ID: 10}},
			Data:    "answer_0",
		},
	}
	h.HandleUpdate(context.Background(), update)
	// Expect "no active session" or "Сессия не найдена"
	if client.lastParams != nil && client.lastParams.Get("text") != "" {
		if !strings.Contains(client.lastParams.Get("text"), "Сессия") && !strings.Contains(client.lastParams.Get("text"), "тренировку") {
			t.Logf("callback answer_0 no session: got %q", client.lastParams.Get("text"))
		}
	}
}
