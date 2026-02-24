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
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type mockTelegramClient struct {
	lastParams url.Values
}

func (c *mockTelegramClient) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	_ = req.Body.Close()
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
