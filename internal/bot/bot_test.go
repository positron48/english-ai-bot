package bot

import (
	"context"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestBot_registerCommands_skipsWhenAPINil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	b := &Bot{api: nil, logger: logger}
	b.registerCommands()
	// Should not panic; logs "skipping bot commands registration"
}

func TestBot_registerCommands_registersWhenAPISet(t *testing.T) {
	client := &mockTelegramClient{}
	logger, _ := zap.NewDevelopment()
	b := &Bot{api: newTestBot(client), logger: logger}
	b.registerCommands()
	if client.DoCount < 1 {
		t.Error("expected at least one API request when registering commands")
	}
	body := string(client.LastBody)
	if body != "" && !strings.Contains(body, "setMyCommands") && client.lastParams.Get("method") != "setMyCommands" {
		t.Logf("registerCommands sent %d request(s); body may be JSON", client.DoCount)
	}
}

func TestBot_Start_cancelledContext_returnsNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	dsn := testutil.GetTestDSN(t)
	cfg := &config.Config{}
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = dsn
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Telegram.Token = ""
	cfg.Telegram.WebhookEnable = false
	cfg.Training.WorkerEnabled = false
	cfg.AI.URL = "http://localhost"
	cfg.AI.Model = "test"
	cfg.Training.PromptFile = ""
	cfg.Training.CircuitBreakerThreshold = 5
	cfg.Training.OptionsDelayMS = 0
	cfg.Training.WrongAnswerDelaySeconds = 0
	cfg.Admin.TelegramID = 0
	cfg.Bot.StartMessage = "start"
	cfg.Bot.HelpMessage = "help"
	cfg.Bot.UnknownCommandMessage = "unknown"
	cfg.WebApp.JWTSecret = "test-secret"

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = bot.Start(ctx)
	if err != nil {
		t.Errorf("Start() with cancelled context error = %v", err)
	}
}
