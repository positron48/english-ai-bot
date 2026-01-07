package main

import (
	"context"
	"fmt"

	"tgbot-skeleton/internal/bot"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/logger"

	"go.uber.org/zap"
)

// Version information set during build
var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("config error: %w", err))
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		panic(fmt.Errorf("logger init error: %w", err))
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting application",
		zap.String("version", version),
		zap.String("buildTime", buildTime),
		zap.String("commit", commit),
	)

	// Create bot instance (Telegram bot initialization is optional)
	telegramBot, err := bot.New(cfg, log)
	if err != nil {
		log.Fatal("failed to initialize application", zap.Error(err))
	}

	// Start bot (includes web server)
	ctx := context.Background()
	if err := telegramBot.Start(ctx); err != nil {
		log.Fatal("application error", zap.Error(err))
	}
}
