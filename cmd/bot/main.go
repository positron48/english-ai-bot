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

	logFields := []zap.Field{
		zap.String("version", version),
		zap.String("buildTime", buildTime),
		zap.String("commit", commit),
		zap.String("learning_pair", cfg.Learning.Pair),
		zap.String("learning_native_lang", cfg.Learning.NativeLang),
		zap.String("learning_target_lang", cfg.Learning.TargetLang),
		zap.String("learning_app_code", cfg.Learning.AppCode),
		zap.String("grammar_bundle_id", cfg.Learning.GrammarBundleID),
	}
	if cfg.Learning.GrammarBundleDir != "" {
		logFields = append(logFields, zap.String("grammar_bundle_dir", cfg.Learning.GrammarBundleDir))
	}
	log.Info("starting application", logFields...)

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
