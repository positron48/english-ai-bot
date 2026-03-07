package bot

import (
	"context"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func testBotConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	cfg.Telegram.Token = ""
	cfg.Telegram.WebhookPath = "/webhook"
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = testutil.GetTestDSN(t)
	cfg.WebApp.JWTSecret = "test-jwt-secret-for-bot-tests"
	return cfg
}

func TestNew_NoToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
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

func TestNew_InvalidDatabase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.Token = ""
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = "invalid-dsn"

	bot, err := New(cfg, logger)
	if err == nil {
		t.Fatal("New() expected error for invalid database")
	}
	if bot != nil {
		t.Error("New() expected nil bot on error")
	}
	if !strings.Contains(err.Error(), "database") {
		t.Errorf("error should mention database, got %q", err.Error())
	}
}

func TestRegisterCommands_ApiNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Should not panic; logs "skipping bot commands registration"
	b.registerCommands()
}

func TestRegisterCommands_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	b.api = newTestBot(&mockTelegramClient{})
	b.registerCommands()
}

func TestRegisterCommands_RequestFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	api := &tgbotapi.BotAPI{Token: "test", Client: &failingTelegramClient{}, Buffer: 1}
	api.SetAPIEndpoint("http://example.com/bot%s/%s")
	b.api = api
	b.registerCommands()
	// Should not panic; logs "failed to set bot commands"
}

func TestNormalizeAPIEndpoint(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"two placeholders kept", "https://api.telegram.org/bot%s/%s", "https://api.telegram.org/bot%s/%s"},
		{"encoded placeholder fixed", "https://x/bot%25s/%25s", "https://x/bot%s/%s"},
		{"url with path", "https://host.example/path", "https://host.example/path/bot%s/%s"},
		{"url with path trailing slash", "https://host.example/path/", "https://host.example/path/bot%s/%s"},
		{"no scheme fallback append", "invalid", "invalid/bot%s/%s"},
		{"whitespace trimmed", "  https://host/  ", "https://host/bot%s/%s"},
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
		in      string
		wantErr bool
	}{
		{"30s", false},
		{"5m", false},
		{"1h", false},
		{"0s", false},
		{"invalid", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			d, err := parseDuration(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDuration(%q) expected error", tt.in)
				}
				return
			}
			if err != nil {
				t.Errorf("parseDuration(%q) error = %v", tt.in, err)
				return
			}
			if tt.in == "30s" && d != 30*time.Second {
				t.Errorf("parseDuration(30s) = %v", d)
			}
		})
	}
}

func TestStart_WebhookMode_ApiNil_CanceledContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.WebhookEnable = true
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStart_LongPollingMode_ApiNil_CanceledContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.WebhookEnable = false
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}
