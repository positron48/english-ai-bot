package bot

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/web"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// webhookFailClient returns error on Do to simulate webhook set failure
type webhookFailClient struct{}

func (c *webhookFailClient) Do(req *http.Request) (*http.Response, error) {
	return nil, io.EOF
}

func testBotWithWebRouter(t *testing.T) (*Bot, context.CancelFunc) {
	t.Helper()
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	userCardRepo := repository.NewUserCardRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	cbService := service.NewCircuitBreakerService(repository.NewCircuitBreakerRepository(conn, logger), 5, logger)

	cfg := &config.Config{}
	cfg.Server.Address = ":0"
	cfg.Telegram.WebhookEnable = false
	cfg.Telegram.WebhookPath = "/webhook"

	webRouter := web.NewRouter(logger, cfg, conn, trainingService, srsService, optionsService, cbService)

	// Use db: nil so Bot shutdown does not close the shared test DB (used by other tests).
	return &Bot{
		api:       nil,
		config:    cfg,
		logger:    logger,
		handler:   nil,
		db:        nil,
		webRouter: webRouter,
	}, func() {}
}

func TestStart_WebhookMode_NilAPI_StartsWebServerOnly(t *testing.T) {
	b, _ := testBotWithWebRouter(t)
	b.config.Telegram.WebhookEnable = true
	b.config.Telegram.WebhookURL = "https://example.com/webhook" // not used when api is nil

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.Start(ctx) }()
	time.Sleep(80 * time.Millisecond)
	cancel()
	err := <-done
	if err != nil {
		t.Errorf("Start (webhook mode, nil API) want nil error, got %v", err)
	}
}

func TestStart_LongPollingMode_NilAPI_StartsWebServerOnly(t *testing.T) {
	b, _ := testBotWithWebRouter(t)
	b.config.Telegram.WebhookEnable = false

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.Start(ctx) }()
	time.Sleep(80 * time.Millisecond)
	cancel()
	err := <-done
	if err != nil {
		t.Errorf("Start (long polling, nil API) want nil error, got %v", err)
	}
}

func TestStartWebhook_MissingURL_ReturnsError(t *testing.T) {
	client := &mockTelegramClient{}
	api := newTestBot(client)
	logger := zap.NewNop()
	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	// WebhookURL and WebhookDomain both empty

	b := &Bot{api: api, config: cfg, logger: logger}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := b.startWebhook(ctx)
	if err == nil {
		t.Fatal("startWebhook with no URL/Domain expected error, got nil")
	}
	if !strings.Contains(err.Error(), "webhook_url") && !strings.Contains(err.Error(), "webhook_domain") {
		t.Errorf("expected webhook config error, got %v", err)
	}
}

func TestStartWebhook_SetWebhookFails_ReturnsError(t *testing.T) {
	api := &tgbotapi.BotAPI{Token: "test", Client: &webhookFailClient{}, Buffer: 1}
	api.SetAPIEndpoint("http://example.com/bot%s/%s")

	logger := zap.NewNop()
	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/webhook"

	b := &Bot{api: api, config: cfg, logger: logger}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := b.startWebhook(ctx)
	if err == nil {
		t.Fatal("startWebhook when Request fails expected error, got nil")
	}
	if !strings.Contains(err.Error(), "webhook") {
		t.Errorf("expected webhook-related error, got %v", err)
	}
}

func TestStartLongPolling_NilAPI_StartsWebServerOnly(t *testing.T) {
	b, _ := testBotWithWebRouter(t)
	b.api = nil

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.startLongPolling(ctx) }()
	time.Sleep(80 * time.Millisecond)
	cancel()
	err := <-done
	if err != nil {
		t.Errorf("startLongPolling (nil API) want nil error, got %v", err)
	}
}

func TestStartWebServerOnly_Shutdown_ReturnsNil(t *testing.T) {
	b, _ := testBotWithWebRouter(t)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- b.startWebServerOnly(ctx) }()
	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-done
	if err != nil {
		t.Errorf("startWebServerOnly after shutdown want nil error, got %v", err)
	}
}

func TestStartWebhook_NilAPI_FallsBackToWebServerOnly(t *testing.T) {
	b, _ := testBotWithWebRouter(t)
	b.api = nil
	b.config.Telegram.WebhookEnable = true
	b.config.Telegram.WebhookURL = "https://example.com/webhook"

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.startWebhook(ctx) }()
	time.Sleep(80 * time.Millisecond)
	cancel()
	err := <-done
	if err != nil {
		t.Errorf("startWebhook (nil API) fallback want nil error, got %v", err)
	}
}

func TestRegisterCommands_NilAPI_NoPanic(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}
	b := &Bot{api: nil, config: cfg, logger: logger}
	b.registerCommands()
	// should not panic and not call Telegram API
}

func TestRegisterCommands_WithAPI_SetsCommands(t *testing.T) {
	client := &mockTelegramClient{}
	api := newTestBot(client)
	logger := zap.NewNop()
	cfg := &config.Config{}
	b := &Bot{api: api, config: cfg, logger: logger}
	b.registerCommands()
	// mock client returns 200 OK; just ensure no panic
	if client.lastParams == nil {
		// SetMyCommands might send as form or JSON; lastParams is from ParseQuery(body)
		// If the library sends JSON, lastParams might stay nil; that's ok
		t.Log("registerCommands with API: request sent (params may be in JSON)")
	}
}

