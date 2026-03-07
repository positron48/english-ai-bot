package service

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type mockTelegramClient struct {
	lastParams url.Values
	lastPath   string
}

func (c *mockTelegramClient) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	_ = req.Body.Close()
	params, _ := url.ParseQuery(string(body))
	c.lastParams = params
	c.lastPath = req.URL.Path

	resp := `{"ok": true, "result": {}}`
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(resp)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

func newTestBot(client *mockTelegramClient) *tgbotapi.BotAPI {
	bot := &tgbotapi.BotAPI{Token: "test", Client: client, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")
	return bot
}

func setupBotCommandService(t *testing.T) (*BotCommandService, *repository.UserRepository, *mockTelegramClient, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	client := &mockTelegramClient{}
	bot := newTestBot(client)

	svc := NewBotCommandService(bot, userRepo, logger, "help", "start", "unknown")

	cleanup := func() {} // shared db, do not close

	return svc, userRepo, client, cleanup
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

func TestBotCommandService_HandleNotificationDaily(t *testing.T) {
	svc, userRepo, client, cleanup := setupBotCommandService(t)
	defer cleanup()

	if _, err := userRepo.GetOrCreateUser(42); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	msg := commandMessage("/notification daily")
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)

	user, err := userRepo.GetUserByTelegramID(42)
	if err != nil {
		t.Fatalf("GetUserByTelegramID error: %v", err)
	}
	if user == nil || !strings.Contains(user.SettingsJSON, "daily") {
		t.Fatalf("expected settings to include daily")
	}

	if client.lastParams.Get("text") == "" {
		t.Fatalf("expected a message to be sent")
	}
}

func TestBotCommandService_HandleNotificationInvalid(t *testing.T) {
	svc, userRepo, client, cleanup := setupBotCommandService(t)
	defer cleanup()

	if _, err := userRepo.GetOrCreateUser(42); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	msg := commandMessage("/notification wrong")
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)

	if !strings.Contains(client.lastParams.Get("text"), "Неверный формат") {
		t.Fatalf("expected invalid format response, got %s", client.lastParams.Get("text"))
	}
}

func TestBotCommandService_HandleUnsubscribe(t *testing.T) {
	svc, userRepo, client, cleanup := setupBotCommandService(t)
	defer cleanup()

	if _, err := userRepo.GetOrCreateUser(42); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	svc.handleUnsubscribe(42, 10)

	user, err := userRepo.GetUserByTelegramID(42)
	if err != nil {
		t.Fatalf("GetUserByTelegramID error: %v", err)
	}
	if user == nil || !strings.Contains(user.SettingsJSON, "never") {
		t.Fatalf("expected settings to include never")
	}

	if client.lastParams.Get("text") == "" {
		t.Fatalf("expected a message to be sent")
	}
}

func TestBotCommandService_HandleUnsubscribe_UserNotFound(t *testing.T) {
	svc, _, client, cleanup := setupBotCommandService(t)
	defer cleanup()
	// Do not create user — telegram ID 999 has no user
	svc.handleUnsubscribe(999, 10)
	text := client.lastParams.Get("text")
	if text == "" {
		t.Fatalf("expected a message to be sent")
	}
	if !strings.Contains(text, "не найден") && !strings.Contains(text, "Пользователь") {
		t.Errorf("expected user-not-found message, got %q", text)
	}
}

func TestBotCommandService_HandleNotification_ShowCurrent(t *testing.T) {
	svc, userRepo, client, cleanup := setupBotCommandService(t)
	defer cleanup()
	// commandMessage uses From.ID 42
	if _, err := userRepo.GetOrCreateUser(42); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	msg := commandMessage("/notification ")
	msg.Text = "/notification"
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)
	text := client.lastParams.Get("text")
	if text == "" {
		t.Fatalf("expected a message with current settings")
	}
	if !strings.Contains(text, "Периодичность") && !strings.Contains(text, "daily") && !strings.Contains(text, "never") {
		t.Errorf("expected current notification info, got %q", text)
	}
}

func TestBotCommandService_HandleNotification_Never(t *testing.T) {
	svc, userRepo, client, cleanup := setupBotCommandService(t)
	defer cleanup()
	if _, err := userRepo.GetOrCreateUser(42); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	msg := commandMessage("/notification never")
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)
	user, _ := userRepo.GetUserByTelegramID(42)
	if user == nil || !strings.Contains(user.SettingsJSON, "never") {
		t.Fatalf("expected settings to include never")
	}
	if !strings.Contains(client.lastParams.Get("text"), "отключены") {
		t.Errorf("expected confirmation about notifications disabled")
	}
}

func TestBotCommandService_HandleNotification_ValidDays(t *testing.T) {
	svc, userRepo, client, cleanup := setupBotCommandService(t)
	defer cleanup()
	if _, err := userRepo.GetOrCreateUser(42); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	msg := commandMessage("/notification 3")
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)
	user, _ := userRepo.GetUserByTelegramID(42)
	if user == nil || !strings.Contains(user.SettingsJSON, "3") {
		t.Fatalf("expected settings to include 3")
	}
	text := client.lastParams.Get("text")
	if !strings.Contains(text, "3") && !strings.Contains(text, "дн") {
		t.Errorf("expected confirmation with period, got %q", text)
	}
}

func TestBotCommandService_HandleStart(t *testing.T) {
	svc, _, client, cleanup := setupBotCommandService(t)
	defer cleanup()

	msg := commandMessage("/start")
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)

	if got := client.lastParams.Get("text"); got != "start" {
		t.Fatalf("expected start message, got %q", got)
	}
}

func TestBotCommandService_HandleHelp(t *testing.T) {
	svc, _, client, cleanup := setupBotCommandService(t)
	defer cleanup()

	msg := commandMessage("/help")
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)

	if got := client.lastParams.Get("text"); got != "help" {
		t.Fatalf("expected help message, got %q", got)
	}
}

func TestBotCommandService_HandleUnknown(t *testing.T) {
	svc, _, client, cleanup := setupBotCommandService(t)
	defer cleanup()

	msg := commandMessage("/unknown")
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)

	if got := client.lastParams.Get("text"); got != "unknown" {
		t.Fatalf("expected unknown message, got %q", got)
	}
}

func TestBotCommandService_HandleStartUnsubscribe(t *testing.T) {
	svc, userRepo, client, cleanup := setupBotCommandService(t)
	defer cleanup()

	if _, err := userRepo.GetOrCreateUser(42); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	msg := commandMessage("/start unsubscribe")
	update := tgbotapi.Update{Message: msg}
	svc.HandleUpdate(update)

	user, err := userRepo.GetUserByTelegramID(42)
	if err != nil {
		t.Fatalf("GetUserByTelegramID error: %v", err)
	}
	if user == nil || !strings.Contains(user.SettingsJSON, "never") {
		t.Fatalf("expected settings to include never")
	}
	if client.lastParams.Get("text") == "" {
		t.Fatalf("expected a message to be sent")
	}
}

func TestBotCommandService_HandleCallbackUnsubscribe(t *testing.T) {
	svc, userRepo, client, cleanup := setupBotCommandService(t)
	defer cleanup()

	if _, err := userRepo.GetOrCreateUser(42); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	update := tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb1",
			Data: "notification_unsubscribe",
			From: &tgbotapi.User{ID: 42, UserName: "tester"},
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 10},
			},
		},
	}
	svc.HandleUpdate(update)

	user, err := userRepo.GetUserByTelegramID(42)
	if err != nil {
		t.Fatalf("GetUserByTelegramID error: %v", err)
	}
	if user == nil || !strings.Contains(user.SettingsJSON, "never") {
		t.Fatalf("expected settings to include never")
	}
	if client.lastPath == "" {
		t.Fatalf("expected callback query to be answered")
	}
}

func TestPluralizeDays(t *testing.T) {
	cases := map[int]string{
		1:  "день",
		2:  "дня",
		5:  "дней",
		11: "дней",
		21: "день",
		-2: "дня",
	}

	for input, expected := range cases {
		if got := pluralizeDays(input); got != expected {
			t.Fatalf("expected %s for %d, got %s", expected, input, got)
		}
	}
}
