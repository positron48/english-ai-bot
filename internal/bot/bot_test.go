package bot

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/web"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
		wantSec float64
	}{
		{"valid seconds", "30s", false, 30},
		{"valid minutes", "5m", false, 300},
		{"valid hours", "1h", false, 3600},
		{"valid combined", "1m30s", false, 90},
		{"empty", "", true, 0},
		{"invalid", "invalid", true, 0},
		{"number only", "30", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Seconds() != tt.wantSec {
				t.Errorf("parseDuration(%q) = %v, want %v sec", tt.in, got, tt.wantSec)
			}
		})
	}
}

func TestNormalizeAPIEndpoint(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"already two placeholders", "https://api.example.com/bot%s/%s", "https://api.example.com/bot%s/%s"},
		{"trim space", "  https://api.example.com  ", "https://api.example.com/bot%s/%s"},
		{"replace encoded placeholder", "https://api.example.com/bot%25s/%25s", "https://api.example.com/bot%s/%s"},
		{"url with path no placeholders", "https://custom.api/telegram", "https://custom.api/telegram/bot%s/%s"},
		{"trailing slash", "https://host/path/", "https://host/path/bot%s/%s"},
		{"no scheme fallback append", "plain", "plain/bot%s/%s"},
		{"ends with slash append", "https://host/", "https://host/bot%s/%s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAPIEndpoint(tt.in)
			if got != tt.want {
				t.Errorf("normalizeAPIEndpoint(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNew_WithoutTelegramToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.Token = ""
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = testutil.GetTestDSN(t)
	cfg.Server.Address = "127.0.0.1:0"
	cfg.WebApp.JWTSecret = "test-jwt-secret-for-bot-test"

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
	if bot.api != nil {
		t.Error("expected api nil when token empty")
	}
	if bot.db == nil {
		t.Error("expected db initialized")
	}
}

func TestNew_WithWorkerEnabled_InvalidInterval(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.Token = ""
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = testutil.GetTestDSN(t)
	cfg.Server.Address = "127.0.0.1:0"
	cfg.WebApp.JWTSecret = "test-jwt-secret-for-bot-test"
	cfg.Training.WorkerEnabled = true
	cfg.Training.WorkerInterval = "invalid-duration"
	cfg.Training.WorkerBatchSize = 1
	cfg.Training.LLMWorkers = 1
	cfg.Admin.TelegramID = 0

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
	// Worker should still be created with default 30s interval (covered by parseDuration error path)
	if bot.trainingWorker == nil {
		t.Error("expected training worker created despite invalid interval")
	}
}

func TestNew_WithWorkerEnabled_ValidInterval(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.Token = ""
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = testutil.GetTestDSN(t)
	cfg.Server.Address = "127.0.0.1:0"
	cfg.WebApp.JWTSecret = "test-jwt-secret-for-bot-test"
	cfg.Training.WorkerEnabled = true
	cfg.Training.WorkerInterval = "1m"
	cfg.Training.WorkerBatchSize = 1
	cfg.Training.LLMWorkers = 1
	cfg.Admin.TelegramID = 0

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
	if bot.trainingWorker == nil {
		t.Error("expected training worker created")
	}
}

func TestNew_DatabaseFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.Token = ""
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = "invalid-dsn"

	_, err := New(cfg, logger)
	if err == nil {
		t.Fatal("New() expected error when database config invalid")
	}
}

func TestStartWebhook_ApiNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/hook"
	cfg.Telegram.WebhookPath = "/hook"
	cfg.Server.Address = "127.0.0.1:0"

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so startWebServerOnly returns right away

	b := &Bot{
		api:     nil,
		config:  cfg,
		logger:  logger,
		db:      nil,
		webRouter: web.NewRouter(logger, cfg, nil, nil, nil, nil, nil),
	}

	err := b.startWebhook(ctx)
	if err != nil {
		t.Errorf("startWebhook() err = %v", err)
	}
}

func TestStartWebhook_NoURLConfigured(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = ""
	cfg.Telegram.WebhookDomain = ""

	client := &mockTelegramClient{}
	b := &Bot{
		api:    newTestBot(client),
		config: cfg,
		logger: logger,
	}

	err := b.startWebhook(context.Background())
	if err == nil {
		t.Fatal("startWebhook() expected error when neither webhook_url nor webhook_domain set")
	}
}

func TestStartWebhook_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/hook"
	cfg.Telegram.WebhookPath = "/hook"
	cfg.Server.Address = "127.0.0.1:0"

	client := &mockTelegramClient{}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	b := &Bot{
		api:       newTestBot(client),
		config:    cfg,
		logger:    logger,
		db:        nil,
		webRouter: web.NewRouter(logger, cfg, nil, nil, nil, nil, nil),
		handler:   nil,
	}

	err := b.startWebhook(ctx)
	if err != nil {
		t.Errorf("startWebhook() err = %v", err)
	}
}

func TestStartWebhook_SetWebhookFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/hook"
	cfg.Telegram.WebhookPath = "/hook"

	// Client that returns Telegram API error (ok: false) so Request returns error
	failClient := &http.Client{
		Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok": false, "description": "bad request"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		}),
	}
	bot := &tgbotapi.BotAPI{Token: "test", Client: failClient, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")

	b := &Bot{api: bot, config: cfg, logger: logger}

	err := b.startWebhook(context.Background())
	if err == nil {
		t.Fatal("startWebhook() expected error when set webhook fails")
	}
}

func TestStartLongPolling_ApiNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Server.Address = "127.0.0.1:0"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	b := &Bot{
		api:        nil,
		config:     cfg,
		logger:     logger,
		db:         nil,
		webRouter:  web.NewRouter(logger, cfg, nil, nil, nil, nil, nil),
	}

	err := b.startLongPolling(ctx)
	if err != nil {
		t.Errorf("startLongPolling() err = %v", err)
	}
}

func TestStartLongPolling_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Telegram.UpdatesTimeout = 1

	client := &mockTelegramClient{}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	b := &Bot{
		api:       newTestBot(client),
		config:    cfg,
		logger:    logger,
		db:        nil,
		webRouter: web.NewRouter(logger, cfg, nil, nil, nil, nil, nil),
		handler:   nil,
	}

	err := b.startLongPolling(ctx)
	if err != nil {
		t.Errorf("startLongPolling() err = %v", err)
	}
}
