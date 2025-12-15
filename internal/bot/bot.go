package bot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Bot represents the Telegram bot
type Bot struct {
	api     *tgbotapi.BotAPI
	config  *config.Config
	logger  *zap.Logger
	handler *Handler
	db      *database.DB
}

// New creates a new bot instance
func New(cfg *config.Config, log *zap.Logger) (*Bot, error) {
	// Initialize Telegram bot
	var bot *tgbotapi.BotAPI
	var err error

	if cfg.Telegram.APIBaseURL != "" {
		endpoint := normalizeAPIEndpoint(cfg.Telegram.APIBaseURL)
		bot, err = tgbotapi.NewBotAPIWithAPIEndpoint(cfg.Telegram.Token, endpoint)
	} else {
		bot, err = tgbotapi.NewBotAPI(cfg.Telegram.Token)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize bot: %w", err)
	}

	bot.Debug = cfg.Telegram.Debug

	log.Info("authorized on account", zap.String("username", bot.Self.UserName))

	// Initialize database
	dbPath := cfg.Database.Path
	// Ensure directory exists
	if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := database.New(dbPath, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create AI service
	aiService := ai.NewService(
		cfg.AI.URL,
		cfg.AI.Model,
		cfg.AI.APIKey,
		cfg.AI.Prompt,
		log,
	)

	// Create repository
	wordRepo := repository.NewWordRepository(db.GetConnection(), log)

	// Create word service
	wordService := service.NewWordService(wordRepo, aiService, log)

	// Create handler
	handler := NewHandler(bot, log, aiService, wordService, cfg)

	return &Bot{
		api:     bot,
		config:  cfg,
		logger:  log,
		handler: handler,
		db:      db,
	}, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Webhook mode vs long polling
	if b.config.Telegram.WebhookEnable {
		return b.startWebhook(ctx)
	}

	return b.startLongPolling(ctx)
}

// startWebhook starts the bot in webhook mode
func (b *Bot) startWebhook(ctx context.Context) error {
	// Determine webhook URL
	var webhookURL string
	if b.config.Telegram.WebhookURL != "" {
		webhookURL = b.config.Telegram.WebhookURL
	} else if b.config.Telegram.WebhookDomain != "" {
		webhookURL = strings.TrimSuffix(b.config.Telegram.WebhookDomain, "/") + b.config.Telegram.WebhookPath
	} else {
		return fmt.Errorf("webhook enabled but neither webhook_url nor webhook_domain is configured")
	}

	b.logger.Info("setting webhook", zap.String("url", webhookURL))

	// Set webhook
	whCfg, _ := tgbotapi.NewWebhook(webhookURL)
	if _, err := b.api.Request(whCfg); err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	b.logger.Info("webhook set successfully")

	// Serve webhook
	http.HandleFunc(b.config.Telegram.WebhookPath, func(w http.ResponseWriter, r *http.Request) {
		update, err := b.api.HandleUpdate(r)
		if err != nil {
			b.logger.Warn("webhook handle error", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if update != nil {
			go b.handler.HandleUpdate(context.Background(), *update)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			b.logger.Error("failed to write health response", zap.Error(err))
		}
	})

	b.logger.Info("starting HTTP server for webhook", zap.String("address", b.config.Server.Address))
	go func() {
		if err := http.ListenAndServe(b.config.Server.Address, nil); err != nil {
			b.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Clean up webhook
	b.logger.Info("cleaning up webhook")
	if _, err := b.api.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
		b.logger.Warn("failed to delete webhook", zap.Error(err))
	} else {
		b.logger.Info("webhook deleted successfully")
	}

	// Close database connection
	if b.db != nil {
		if err := b.db.Close(); err != nil {
			b.logger.Warn("failed to close database", zap.Error(err))
		}
	}

	b.logger.Info("shutting down")
	return nil
}

// startLongPolling starts the bot in long polling mode
func (b *Bot) startLongPolling(ctx context.Context) error {
	// Health endpoint
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				b.logger.Error("failed to write health response", zap.Error(err))
			}
		})
		if err := http.ListenAndServe(b.config.Server.Address, nil); err != nil {
			b.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Long polling loop
	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.config.Telegram.UpdatesTimeout
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("shutting down")
			// Close database connection
			if b.db != nil {
				if err := b.db.Close(); err != nil {
					b.logger.Warn("failed to close database", zap.Error(err))
				}
			}
			return nil
		case update := <-updates:
			b.handler.HandleUpdate(ctx, update)
		}
	}
}

// normalizeAPIEndpoint ensures endpoint string is a valid format expected by tgbotapi
func normalizeAPIEndpoint(base string) string {
	s := strings.TrimSpace(base)
	// Fix encoded placeholders
	s = strings.ReplaceAll(s, "%25s", "%s")
	// If it already has exactly two placeholders, keep as-is
	if strings.Count(s, "%s") == 2 {
		return s
	}
	// If the placeholder count is wrong or missing, rebuild using parsed URL
	if u, err := url.Parse(s); err == nil && u.Scheme != "" && u.Host != "" {
		path := strings.TrimSuffix(u.Path, "/")
		return u.Scheme + "://" + u.Host + path + "/bot%s/%s"
	}
	// Fallback: just append the correct suffix
	if strings.HasSuffix(s, "/") {
		return s + "bot%s/%s"
	}
	return s + "/bot%s/%s"
}
