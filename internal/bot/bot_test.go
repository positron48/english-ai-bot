package bot

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestNormalizeAPIEndpoint(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty",
			in:   "",
			want: "/bot%s/%s",
		},
		{
			name: "trimmed with two placeholders kept as-is",
			in:   "https://api.telegram.org/bot%s/%s",
			want: "https://api.telegram.org/bot%s/%s",
		},
		{
			name: "encoded placeholders fixed",
			in:   "https://example.com/bot%25s/%25s",
			want: "https://example.com/bot%s/%s",
		},
		{
			name: "full URL without placeholders gets suffix",
			in:   "https://proxy.example.com/telegram",
			want: "https://proxy.example.com/telegram/bot%s/%s",
		},
		{
			name: "URL with trailing slash",
			in:   "https://proxy.example.com/",
			want: "https://proxy.example.com/bot%s/%s",
		},
		{
			name: "whitespace trimmed",
			in:   "  https://api.telegram.org  ",
			want: "https://api.telegram.org/bot%s/%s",
		},
		{
			name: "single placeholder gets /bot%s/%s appended to path",
			in:   "https://api.telegram.org/bot%s",
			want: "https://api.telegram.org/bot%s/bot%s/%s",
		},
		{
			name: "invalid URL fallback append",
			in:   "not-a-url",
			want: "not-a-url/bot%s/%s",
		},
		{
			name: "invalid URL with trailing slash",
			in:   "not-a-url/",
			want: "not-a-url/bot%s/%s",
		},
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

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantDur time.Duration
		wantErr bool
	}{
		{"valid seconds", "30s", 30 * time.Second, false},
		{"valid minutes", "5m", 5 * time.Minute, false},
		{"valid hours", "1h", time.Hour, false},
		{"valid combined", "1h30m", 90 * time.Minute, false},
		{"empty", "", 0, true},
		{"invalid", "x1m", 0, true},
		{"zero", "0s", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantDur {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.in, got, tt.wantDur)
			}
		})
	}
}

func TestStartWebhook_ErrorNoURLNorDomain(t *testing.T) {
	logger := zap.NewNop()
	client := &mockTelegramClient{}
	api := newTestBot(client)
	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = ""
	cfg.Telegram.WebhookDomain = ""
	cfg.Telegram.WebhookPath = "/webhook"
	cfg.Server.Address = ":0"

	b := &Bot{
		api:    api,
		config: cfg,
		logger: logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := b.startWebhook(ctx)
	if err == nil {
		t.Fatal("startWebhook expected error when neither webhook_url nor webhook_domain set")
	}
	if !strings.Contains(err.Error(), "webhook_url") && !strings.Contains(err.Error(), "webhook_domain") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStartWebhook_WebhookDomainUsedWhenURLEmpty(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	client := &mockTelegramClient{}
	api := newTestBot(client)

	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = ""
	cfg.Telegram.WebhookDomain = "https://example.com"
	cfg.Telegram.WebhookPath = "/webhook"
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Bot.StartMessage = "start"
	cfg.Bot.HelpMessage = "help"
	cfg.Bot.UnknownCommandMessage = "unknown"

	userRepo := repository.NewUserRepository(conn, logger)
	cbRepo := repository.NewCircuitBreakerRepository(conn, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	handler := NewHandler(api, logger, nil, nil, nil, userRepo, nil, nil, cbService, cfg, conn)

	trainingService := service.NewTrainingService(nil, nil, nil, nil, logger)
	srsService := service.NewSRSService(nil, logger)
	optionsService := service.NewOptionsService(nil, logger)
	webRouter := web.NewRouter(logger, cfg, conn, trainingService, srsService, optionsService, cbService)

	b := &Bot{
		api:       api,
		config:    cfg,
		logger:    logger,
		handler:   handler,
		db:        nil, // do not close shared test DB on shutdown
		webRouter: webRouter,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.startWebhook(ctx) }()
	time.Sleep(80 * time.Millisecond)
	cancel()
	err := <-done
	if err != nil {
		t.Errorf("startWebhook with webhook_domain should succeed until shutdown: %v", err)
	}
	if client.DoCount < 1 {
		t.Errorf("expected setWebhook API call, got DoCount=%d", client.DoCount)
	}
}

func TestStartWebhook_ErrorSetWebhookFails(t *testing.T) {
	logger := zap.NewNop()
	api := &tgbotapi.BotAPI{Token: "test", Client: &failingTelegramClient{}, Buffer: 1}
	api.SetAPIEndpoint("http://example.com/bot%s/%s")
	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/webhook"
	cfg.Telegram.WebhookPath = "/webhook"
	cfg.Server.Address = ":0"

	b := &Bot{
		api:    api,
		config: cfg,
		logger: logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := b.startWebhook(ctx)
	if err == nil {
		t.Fatal("startWebhook expected error when setWebhook fails")
	}
	if !strings.Contains(err.Error(), "webhook") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStartWebhook_Success_Shutdown(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	client := &mockTelegramClient{}
	api := newTestBot(client)

	cfg := &config.Config{}
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/webhook"
	cfg.Telegram.WebhookPath = "/webhook"
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Bot.StartMessage = "start"
	cfg.Bot.HelpMessage = "help"
	cfg.Bot.UnknownCommandMessage = "unknown"

	userRepo := repository.NewUserRepository(conn, logger)
	cbRepo := repository.NewCircuitBreakerRepository(conn, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	handler := NewHandler(api, logger, nil, nil, nil, userRepo, nil, nil, cbService, cfg, conn)

	trainingService := service.NewTrainingService(nil, nil, nil, nil, logger)
	srsService := service.NewSRSService(nil, logger)
	optionsService := service.NewOptionsService(nil, logger)
	webRouter := web.NewRouter(logger, cfg, conn, trainingService, srsService, optionsService, cbService)

	b := &Bot{
		api:       api,
		config:    cfg,
		logger:    logger,
		handler:   handler,
		db:        nil, // do not close shared test DB on shutdown
		webRouter: webRouter,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.startWebhook(ctx) }()

	time.Sleep(100 * time.Millisecond)
	cancel()

	err := <-done
	if err != nil {
		t.Errorf("startWebhook after shutdown: %v", err)
	}
	if client.DoCount < 1 {
		t.Errorf("expected at least one API call (setWebhook), got DoCount=%d", client.DoCount)
	}
}

func TestStartWebhook_APINil_StartsWebServerOnly(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	cfg := &config.Config{}
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Bot.StartMessage = "start"
	cfg.Bot.HelpMessage = "help"
	cfg.Bot.UnknownCommandMessage = "unknown"

	cbRepo := repository.NewCircuitBreakerRepository(conn, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	trainingService := service.NewTrainingService(nil, nil, nil, nil, logger)
	srsService := service.NewSRSService(nil, logger)
	optionsService := service.NewOptionsService(nil, logger)
	webRouter := web.NewRouter(logger, cfg, conn, trainingService, srsService, optionsService, cbService)

	b := &Bot{
		api:       nil,
		config:    cfg,
		logger:    logger,
		db:        nil, // do not close shared test DB on shutdown
		webRouter: webRouter,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.startWebhook(ctx) }()

	time.Sleep(100 * time.Millisecond)
	cancel()

	err := <-done
	if err != nil {
		t.Errorf("startWebhook (api nil): %v", err)
	}
}

func TestStartLongPolling_APINil_StartsWebServerOnly(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	cfg := &config.Config{}
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Telegram.UpdatesTimeout = 10
	cfg.Bot.StartMessage = "start"
	cfg.Bot.HelpMessage = "help"
	cfg.Bot.UnknownCommandMessage = "unknown"

	cbRepo := repository.NewCircuitBreakerRepository(conn, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	trainingService := service.NewTrainingService(nil, nil, nil, nil, logger)
	srsService := service.NewSRSService(nil, logger)
	optionsService := service.NewOptionsService(nil, logger)
	webRouter := web.NewRouter(logger, cfg, conn, trainingService, srsService, optionsService, cbService)

	b := &Bot{
		api:       nil,
		config:    cfg,
		logger:    logger,
		db:        nil, // do not close shared test DB on shutdown
		webRouter: webRouter,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.startLongPolling(ctx) }()

	time.Sleep(100 * time.Millisecond)
	cancel()

	err := <-done
	if err != nil {
		t.Errorf("startLongPolling (api nil): %v", err)
	}
}

func TestStartLongPolling_Success_Shutdown(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// getMe expects result to be a user object; getUpdates expects result to be an array
		if strings.Contains(r.URL.Path, "getMe") {
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test","username":"test_bot"}}`))
		} else {
			_, _ = w.Write([]byte(`{"ok":true,"result":[]}`))
		}
	}))
	defer srv.Close()

	api, err := tgbotapi.NewBotAPIWithAPIEndpoint("test-token", normalizeAPIEndpoint(srv.URL))
	if err != nil {
		t.Fatalf("NewBotAPIWithAPIEndpoint: %v", err)
	}
	api.Debug = false

	cfg := &config.Config{}
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Telegram.UpdatesTimeout = 1
	cfg.Bot.StartMessage = "start"
	cfg.Bot.HelpMessage = "help"
	cfg.Bot.UnknownCommandMessage = "unknown"

	userRepo := repository.NewUserRepository(conn, logger)
	cbRepo := repository.NewCircuitBreakerRepository(conn, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	handler := NewHandler(api, logger, nil, nil, nil, userRepo, nil, nil, cbService, cfg, conn)

	trainingService := service.NewTrainingService(nil, nil, nil, nil, logger)
	srsService := service.NewSRSService(nil, logger)
	optionsService := service.NewOptionsService(nil, logger)
	webRouter := web.NewRouter(logger, cfg, conn, trainingService, srsService, optionsService, cbService)

	b := &Bot{
		api:       api,
		config:    cfg,
		logger:    logger,
		handler:   handler,
		db:        nil, // do not close shared test DB on shutdown
		webRouter: webRouter,
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.startLongPolling(ctx) }()

	time.Sleep(200 * time.Millisecond)
	cancel()

	err = <-done
	if err != nil {
		t.Errorf("startLongPolling after shutdown: %v", err)
	}
}

