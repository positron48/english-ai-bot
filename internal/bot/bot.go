package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/web"
	"tgbot-skeleton/internal/wordsetimport"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Bot represents the Telegram bot
type Bot struct {
	api                  *tgbotapi.BotAPI
	config               *config.Config
	logger               *zap.Logger
	handler              *Handler
	db                   *database.DB
	trainingWorker       *service.TrainingWorker
	pronunciationService *service.PronunciationService
	notificationService  *service.NotificationService
	webRouter            *web.Router
	// shutdownCtxFn returns the context used for http.Server.Shutdown; defaults to context.Background.
	// Overridable in tests to exercise the shutdown-error path.
	shutdownCtxFn func() context.Context
	// dbCloseFn closes the database; defaults to b.db.Close.
	// Overridable in tests to exercise the db-close-error path.
	dbCloseFn func() error
}

// New creates a new bot instance
func New(cfg *config.Config, log *zap.Logger) (*Bot, error) {
	// Initialize Telegram bot (optional - app can work without it)
	var bot *tgbotapi.BotAPI
	var err error

	if cfg.Telegram.Token != "" {
		endpoint := tgbotapi.APIEndpoint
		if cfg.Telegram.APIBaseURL != "" {
			endpoint = normalizeAPIEndpoint(cfg.Telegram.APIBaseURL)
		}

		var httpClient tgbotapi.HTTPClient = &http.Client{}
		if cfg.Telegram.Socks5ProxyAddr != "" {
			telegramClient, clientErr := newTelegramHTTPClientWithSocks5Proxy(cfg.Telegram.Socks5ProxyAddr, cfg.Telegram.UpdatesTimeout, log)
			if clientErr != nil {
				err = clientErr
			} else {
				httpClient = telegramClient
			}
		}

		if err == nil {
			bot, err = tgbotapi.NewBotAPIWithClient(cfg.Telegram.Token, endpoint, httpClient)
		}
		if err != nil {
			log.Warn("failed to initialize Telegram bot, continuing without it",
				zap.Error(err),
				zap.String("note", "Web application will continue to work, but Telegram features will be disabled"))
			bot = nil
		} else {
			// Disable debug mode to avoid verbose Telegram API logs
			bot.Debug = false
			log.Info("authorized on account", zap.String("username", bot.Self.UserName))
		}
	} else {
		log.Info("Telegram token not provided, running without Telegram bot")
	}

	// Initialize database
	db, err := database.NewWithConfig(cfg.Database.Driver, cfg.Database.Path, cfg.Database.URL, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create AI service
	aiHTTPTimeout := ai.ParseHTTPTimeout(cfg.AI.RequestTimeout)
	aiService := ai.NewServiceWithTimeout(
		cfg.AI.URL,
		cfg.AI.Model,
		cfg.AI.APIKey,
		cfg.AI.Prompt,
		aiHTTPTimeout,
		log,
	)

	// Load training prompt (same template engine as AI_PROMPT_FILE: {{native_lang}}, {{target_lang}}, {{pair}})
	if cfg.Training.PromptFile != "" {
		trainingPrompt, err := ai.LoadRenderedPromptFile(cfg.Training.PromptFile, cfg.Learning.NativeLang, cfg.Learning.TargetLang, cfg.Learning.Pair)
		if err != nil {
			log.Warn("failed to load training prompt file, training worker will not work",
				zap.String("file", cfg.Training.PromptFile),
				zap.Error(err),
			)
		} else {
			aiService.SetTrainingPrompt(trainingPrompt)
		}
	}

	// Create repositories
	conn := db.GetConnection()

	// Zero-touch bootstrap for English frequency word sets.
	if err := wordsetimport.AutoSyncEnglishWordSets(context.Background(), cfg, conn, log); err != nil {
		log.Warn("english word sets bootstrap skipped due to error", zap.Error(err))
	}
	if err := wordsetimport.AutoSyncEnglishMustHaveWordSets(context.Background(), cfg, conn, log); err != nil {
		log.Warn("english must-have word sets bootstrap skipped due to error", zap.Error(err))
	}
	if err := wordsetimport.AutoSyncSpanishMustHaveWordSets(context.Background(), cfg, conn, log); err != nil {
		log.Warn("spanish must-have word sets bootstrap skipped due to error", zap.Error(err))
	}
	courseRepo := repository.NewCourseRepository(conn, log)
	if summary, err := courseRepo.BackfillUserCoursesForLearning(context.Background(), cfg.Learning); err != nil {
		log.Warn("linglow user_courses bootstrap skipped due to error", zap.Error(err))
	} else {
		log.Info("linglow user_courses bootstrap completed",
			zap.String("course_code", summary.CourseCode),
			zap.Int64("course_id", summary.CourseID),
			zap.Int64("users_scanned", summary.UsersScanned),
			zap.Int64("existing", summary.Existing),
			zap.Int64("created", summary.Created),
		)
	}

	wordRepo := repository.NewWordRepository(conn, log)
	userRepo := repository.NewUserRepository(conn, log)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, log)
	userCardRepo := repository.NewUserCardRepository(conn, log)
	sessionRepo := repository.NewSessionRepository(conn, log)
	cbRepo := repository.NewCircuitBreakerRepository(conn, log)
	nudgeRepo := repository.NewNudgeRepository(conn, log)

	// Create services
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, log)
	wordService := service.NewWordServiceWithMastering(wordRepo, trainingCardRepo, userCardRepo, userWordMasteringRepo, aiService, cfg.Learning, log)
	pronunciationService := service.NewPronunciationService(cfg.TTS, cfg.Learning, wordRepo, log)
	wordService.SetPronunciationService(pronunciationService)
	verbFormsRepoForSync := repository.NewVerbFormsRepository(conn, log)
	verbTrainingForWordSync := service.NewVerbTrainingService(verbFormsRepoForSync, cfg.Learning, cfg.Training, log)
	wordService.SetVerbFormCardsSync(cfg.Training, func(userID int64) error {
		u, err := userRepo.GetUserByID(userID)
		var scopes []string
		switch {
		case err != nil || u == nil || strings.TrimSpace(u.SettingsJSON) == "":
			scopes = models.DefaultSpanishVerbScopes()
		default:
			var settings models.UserSettings
			if err := json.Unmarshal([]byte(u.SettingsJSON), &settings); err != nil {
				scopes = models.DefaultSpanishVerbScopes()
			} else {
				scopes = service.ResolveVerbScopes(&settings, cfg.Learning)
			}
		}
		if len(scopes) == 0 {
			scopes = models.DefaultSpanishVerbScopes()
		}
		return verbTrainingForWordSync.EnsureVerbFormUserCards(userID, scopes)
	})
	srsService := service.NewSRSService(userCardRepo, cfg.Learning, log)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, cfg.Learning, log)
	targetLang := strings.ToLower(strings.TrimSpace(cfg.Learning.TargetLang))
	optionsService := service.NewOptionsService(trainingCardRepo, log, targetLang)
	cbService := service.NewCircuitBreakerService(cbRepo, cfg.Training.CircuitBreakerThreshold, log)

	// Create training handler
	trainingHandler := NewTrainingHandler(bot, trainingService, srsService, optionsService, sessionRepo, log, cfg.Training.OptionsDelayMS, cfg.Training.WrongAnswerDelaySeconds, conn)

	// Create handler
	handler := NewHandler(bot, log, aiService, wordService, trainingHandler, userRepo, trainingCardRepo, userCardRepo, cbService, cfg, conn)

	// Create training worker
	var trainingWorker *service.TrainingWorker
	if cfg.Training.WorkerEnabled {
		workerInterval, err := parseDuration(cfg.Training.WorkerInterval)
		if err != nil {
			log.Warn("invalid worker interval, using default 30s", zap.Error(err))
			workerInterval = 30 * time.Second
		}

		trainingWorker = service.NewTrainingWorker(
			aiService,
			wordRepo,
			trainingCardRepo,
			userCardRepo,
			userRepo,
			pronunciationService,
			cbService,
			bot,
			cfg.Admin.TelegramID,
			cfg.Training.WorkerBatchSize,
			cfg.Training.LLMWorkers,
			workerInterval,
			cfg.AI.ModelHigh,
			cfg.Learning,
			log,
		)
	}

	// Create notification service
	notificationService := service.NewNotificationService(
		bot,
		userRepo,
		userCardRepo,
		nudgeRepo,
		sessionRepo,
		log,
	)

	// Create web repositories
	otpRepo := repository.NewWebOTPRepository(conn, log)

	// Create grammar repositories and service
	var grammarContentRepo *repository.GrammarContentRepository
	if cfg.Learning.ContentSource == "db" {
		grammarContentRepo = repository.NewGrammarContentRepositoryFromDB(conn, cfg.Learning.GrammarBundleID, log)
	} else {
		grammarContentRepo, err = repository.NewGrammarContentRepositoryForLearning(cfg.Learning, log)
		if err != nil {
			return nil, fmt.Errorf("grammar content repository: %w", err)
		}
	}
	grammarPublishRepo := repository.NewGrammarPublishRepository(conn, log)
	grammarAttemptRepo := repository.NewGrammarAttemptRepository(conn, log)
	var grammarTrainingPackRepo *repository.GrammarTrainingPackRepository
	if cfg.Learning.ContentSource == "db" {
		grammarTrainingPackRepo = repository.NewGrammarTrainingPackRepositoryFromDB(conn, cfg.Learning.GrammarBundleID, log)
	} else {
		grammarTrainingPackRepo, err = repository.NewGrammarTrainingPackRepositoryForLearning(cfg.Learning, log)
		if err != nil {
			return nil, fmt.Errorf("grammar training pack repository: %w", err)
		}
	}
	grammarSRSRepo := repository.NewGrammarSRSRepository(conn, log)
	grammarService := service.NewGrammarService(
		grammarContentRepo,
		grammarPublishRepo,
		grammarAttemptRepo,
		cfg.Learning,
		log,
	)
	grammarService.SetTrainingPackRepository(grammarTrainingPackRepo)
	grammarService.SetSRSRepository(grammarSRSRepo)

	// Create web router
	webRouter := web.NewRouter(
		log,
		cfg,
		conn,
		trainingService,
		srsService,
		optionsService,
		cbService,
	)
	webRouter.SetDependencies(userRepo, wordService, aiService, bot, cfg.Telegram.Token)
	webRouter.SetOTPRepo(otpRepo)
	webRouter.SetGrammarService(grammarService)
	webRouter.SetPronunciationService(pronunciationService)
	speakingEvaluator := service.NewSpeakingEvaluatorService(cfg, log)
	webRouter.SetSpeakingEvaluator(speakingEvaluator)

	b := &Bot{
		api:                  bot,
		config:               cfg,
		logger:               log,
		handler:              handler,
		db:                   db,
		trainingWorker:       trainingWorker,
		pronunciationService: pronunciationService,
		notificationService:  notificationService,
		webRouter:            webRouter,
		shutdownCtxFn:        context.Background,
	}
	b.dbCloseFn = b.db.Close
	return b, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start training worker
	if b.trainingWorker != nil {
		go b.trainingWorker.Start(ctx)
		b.logger.Info("training worker started")
	}

	// Start pronunciation prefetch/backfill worker.
	if b.pronunciationService != nil {
		go b.pronunciationService.Start(ctx)
		b.logger.Info("pronunciation service started")
	}

	// Start notification service
	if b.notificationService != nil {
		go b.notificationService.Start(ctx)
		b.logger.Info("notification service started")
	}

	// Reading/speaking catalog: in bundle mode sync bundle → DB on startup; in DB-first mode
	// content must already be imported by cmd/import_learning_content.
	if b.webRouter != nil {
		if b.config.Learning.ContentSource != "db" {
			if err := b.webRouter.SyncReadingCatalogFromBundle(ctx); err != nil {
				b.logger.Warn("reading catalog sync failed", zap.Error(err))
			}
			if err := b.webRouter.SyncSpeakingCatalogFromBundle(ctx); err != nil {
				b.logger.Warn("speaking catalog sync failed", zap.Error(err))
			}
		}
		go b.webRouter.BootstrapReadingWordCards(ctx)
	}

	// Register bot commands (will be updated by webRouter.SetDependencies)
	b.registerCommands()

	// Webhook mode vs long polling
	if b.config.Telegram.WebhookEnable {
		return b.startWebhook(ctx)
	}

	return b.startLongPolling(ctx)
}

// registerCommands registers bot commands
func (b *Bot) registerCommands() {
	if b.api == nil {
		b.logger.Info("skipping bot commands registration (Telegram bot not initialized)")
		return
	}

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начать работу с ботом"},
		{Command: "help", Description: "Помощь"},
		{Command: "train", Description: "Начать тренировку слов"},
		{Command: "stats", Description: "Статистика по карточкам"},
		{Command: "unsubscribe", Description: "Отписаться от уведомлений"},
		{Command: "notification", Description: "Настроить периодичность уведомлений"},
	}

	if _, err := b.api.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		b.logger.Warn("failed to set bot commands", zap.Error(err))
	} else {
		b.logger.Info("bot commands registered")
	}
}

// healthHandler handles the /health HTTP endpoint.
func (b *Bot) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		b.logger.Error("failed to write health response", zap.Error(err))
	}
}

// startWebhook starts the bot in webhook mode
func (b *Bot) startWebhook(ctx context.Context) error {
	if b.api == nil {
		b.logger.Warn("webhook mode requested but Telegram bot not initialized, starting web server only")
		return b.startWebServerOnly(ctx)
	}

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

	// Create main mux
	mux := http.NewServeMux()

	// Serve webhook
	mux.HandleFunc(b.config.Telegram.WebhookPath, func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/health", b.healthHandler)

	// Web app routes
	mux.Handle("/", b.webRouter)

	b.logger.Info("starting HTTP server for webhook", zap.String("address", b.config.Server.Address))
	go func() {
		if err := http.ListenAndServe(b.config.Server.Address, mux); err != nil {
			b.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Clean up webhook
	if b.api != nil {
		b.logger.Info("cleaning up webhook")
		if _, err := b.api.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
			b.logger.Warn("failed to delete webhook", zap.Error(err))
		} else {
			b.logger.Info("webhook deleted successfully")
		}
	}

	// Close database connection
	if b.db != nil {
		if err := b.dbCloseFn(); err != nil {
			b.logger.Warn("failed to close database", zap.Error(err))
		}
	}

	b.logger.Info("shutting down")
	return nil
}

// startLongPolling starts the bot in long polling mode
func (b *Bot) startLongPolling(ctx context.Context) error {
	if b.api == nil {
		b.logger.Warn("long polling mode requested but Telegram bot not initialized, starting web server only")
		return b.startWebServerOnly(ctx)
	}

	// Create main mux
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", b.healthHandler)

	// Web app routes
	mux.Handle("/", b.webRouter)

	// Start HTTP server
	go func() {
		if err := http.ListenAndServe(b.config.Server.Address, mux); err != nil {
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
				if err := b.dbCloseFn(); err != nil {
					b.logger.Warn("failed to close database", zap.Error(err))
				}
			}
			return nil
		case update := <-updates:
			b.handler.HandleUpdate(ctx, update)
		}
	}
}

// startWebServerOnly starts only the web server without Telegram bot functionality
func (b *Bot) startWebServerOnly(ctx context.Context) error {
	// Create main mux
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", b.healthHandler)

	// Web app routes
	mux.Handle("/", b.webRouter)

	b.logger.Info("starting HTTP server (without Telegram bot)", zap.String("address", b.config.Server.Address))

	// Start HTTP server
	server := &http.Server{
		Addr:    b.config.Server.Address,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			b.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	b.logger.Info("shutting down HTTP server")
	if err := server.Shutdown(b.shutdownCtxFn()); err != nil {
		b.logger.Warn("error shutting down HTTP server", zap.Error(err))
	}

	// Close database connection
	if b.db != nil {
		if err := b.dbCloseFn(); err != nil {
			b.logger.Warn("failed to close database", zap.Error(err))
		}
	}

	b.logger.Info("shutdown complete")
	return nil
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

// parseDuration parses a duration string (e.g., "30s", "5m")
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
