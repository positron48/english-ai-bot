package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/learning"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// @title           English Bot API
// @version         1.0
// @description     API для English Bot - системы изучения английского языка с использованием SRS (Spaced Repetition System)
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8184
// @BasePath  /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer JWT token authentication. В Swagger UI введите только токен (без "Bearer "), префикс будет добавлен автоматически.

// @tag.name Auth
// @tag.description Методы аутентификации (Telegram WebApp и OTP)

// @tag.name Dashboard
// @tag.description Главная панель пользователя

// @tag.name Vocab
// @tag.description Управление словарным запасом

// @tag.name Training
// @tag.description Система тренировок и обучения

// @tag.name Chat
// @tag.description AI чат для помощи в изучении языка

// @tag.name Admin
// @tag.description Административные функции

// pronunciationServiceInterface is the subset of PronunciationService used by web handlers.
// Defined in web so tests can inject mocks for full coverage of tts.go branches.
type pronunciationServiceInterface interface {
	IsEnabled() bool
	PublicBasePath() string
	AudioDir() string
	Lookup(word string) service.PronunciationLookupResult
	GetStatus(word string) (service.TTSStatusResult, error)
	ForceRegenerate(word string) (service.TTSStatusResult, error)
	Recheck(word string) (service.TTSStatusResult, error)
	ScheduleWord(word string) bool
	ListPendingExternal(limit int) ([]service.ExternalPendingWord, error)
	StoreExternalAudio(word, provider, format string, audio []byte) (service.TTSStatusResult, error)
	MarkExternalFailure(word, provider, state, errorCode, errorMessage string) (service.TTSStatusResult, error)
}

// srsServiceInterface is the subset of SRSService used by web handlers.
// Defined in web so tests can inject mocks for full coverage of error branches.
type srsServiceInterface interface {
	GradeCard(card *models.UserCard, attempt models.AttemptData) error
	RecordWrongAnswer(card *models.UserCard, wrongOption string) error
}

// optionsServiceInterface is the subset of OptionsService used by web handlers.
// Defined in web so tests can inject mocks for full coverage of error branches.
type optionsServiceInterface interface {
	GenerateOptions(card *models.UserCardWithTraining, optionCount int, sessionWords []string, sessionWordENs map[string]bool, sessionWordRUs map[string]bool) ([]string, string, error)
}

// Router handles web routes
type Router struct {
	mux                               *http.ServeMux
	logger                            *zap.Logger
	config                            *config.Config
	db                                *sql.DB
	userRepo                          interface{} // Will be properly typed later
	accessCategoryRepo                *repository.UserAccessCategoryRepository
	trainingService                   *service.TrainingService
	srsService                        srsServiceInterface
	optionsService                    optionsServiceInterface
	wordService                       interface{} // Will be properly typed later
	grammarService                    *service.GrammarService
	grammarServices                   map[string]*service.GrammarService // keyed by bundle ID: "en", "es", ...
	cbService                         *service.CircuitBreakerService
	ttsCbService                      *service.CircuitBreakerService
	pronunciationService              pronunciationServiceInterface
	pronunciationServices             map[string]pronunciationServiceInterface // keyed by bundle ID: "en", "es", ...
	aiService                         interface{}                              // Will be properly typed later
	bot                               *tgbotapi.BotAPI
	authMiddleware                    *AuthMiddleware
	otpRepo                           *repository.WebOTPRepository
	courseRepo                        *repository.CourseRepository
	linglowEventRepo                  *repository.LinglowEventRepository
	linglowSRSMirrorRepo              *repository.LinglowSRSMirrorRepository
	readingCatalogRepo                *repository.ReadingCatalogRepository
	speakingCatalogRepo               *repository.SpeakingCatalogRepository
	speakingSessionRepo               *repository.SpeakingSessionRepository
	conversationRepo                  *repository.ConversationRepository
	speakingEvaluator                 *service.SpeakingEvaluatorService
	botToken                          string
	webTrainingHandler                *WebTrainingHandler
	rateLimiter                       *RateLimiter
	botCommandService                 *service.BotCommandService
	pronunciationMediaRouteRegistered bool
	internalTTSEnabled                bool
	internalTTSTokens                 map[string]string
	internalTTSMaxPendingLimit        int
	internalTTSMaxUploadBytes         int
	internalServiceTokens             map[string]string
	// generateTokenPairForRefresh if set is used in handleAuthRefresh instead of auth.GenerateTokenPair (for testing).
	generateTokenPairForRefresh func(userID, telegramID int64) (access, refresh string, err error)
	// getOrCreateUserForTelegram if set is used in handleAuthTelegram/handleAuthTelegramUnsafe instead of userRepo.GetOrCreateUser (for testing).
	getOrCreateUserForTelegram func(telegramID int64) (*models.User, error)
	// generateTokenPairForTelegram if set is used in handleAuthTelegram/handleAuthTelegramUnsafe instead of auth.GenerateTokenPair (for testing).
	generateTokenPairForTelegram func(userID, telegramID int64) (access, refresh string, err error)
}

// NewRouter creates a new web router
func NewRouter(
	logger *zap.Logger,
	cfg *config.Config,
	db *sql.DB,
	trainingService *service.TrainingService,
	srsService srsServiceInterface,
	optionsService optionsServiceInterface,
	cbService *service.CircuitBreakerService,
) *Router {
	// Initialize rate limiter (cleanup every 5 minutes, TTL 1 hour)
	rateLimiter := NewRateLimiter(5*time.Minute, 1*time.Hour)

	r := &Router{
		mux:                        http.NewServeMux(),
		logger:                     logger,
		config:                     cfg,
		db:                         db,
		trainingService:            trainingService,
		srsService:                 srsService,
		optionsService:             optionsService,
		cbService:                  cbService,
		rateLimiter:                rateLimiter,
		internalTTSEnabled:         cfg.TTS.InternalEnabled,
		internalTTSTokens:          parseTTSTokens(cfg.TTS.InternalTokensJSON, logger),
		internalTTSMaxPendingLimit: maxInt(cfg.TTS.InternalMaxPendingLimit, 500),
		internalTTSMaxUploadBytes:  maxInt(cfg.TTS.InternalMaxUploadMB, 10) * 1024 * 1024,
		internalServiceTokens:      parseInternalServiceTokens(cfg, logger),
	}
	if db != nil {
		r.courseRepo = repository.NewCourseRepository(db, logger)
		r.linglowEventRepo = repository.NewLinglowEventRepository(db)
		r.linglowSRSMirrorRepo = repository.NewLinglowSRSMirrorRepository(db)
		r.readingCatalogRepo = repository.NewReadingCatalogRepository(db)
		r.speakingCatalogRepo = repository.NewSpeakingCatalogRepository(db)
		r.speakingSessionRepo = repository.NewSpeakingSessionRepository(db)
		r.conversationRepo = repository.NewConversationRepository(db, logger)
	}

	// Setup routes
	r.setupRoutes()

	return r
}

// SetDependencies sets additional dependencies
func (r *Router) SetDependencies(
	userRepo interface{},
	wordService interface{},
	aiService interface{},
	bot *tgbotapi.BotAPI,
	botToken string,
) {
	r.userRepo = userRepo
	r.wordService = wordService
	r.aiService = aiService
	r.bot = bot
	r.botToken = botToken

	// Create JWT service
	jwtService, err := NewJWTService(r.config, r.logger)
	if err != nil {
		panic(fmt.Sprintf("failed to create JWT service: %v", err))
	}

	// Create access category repository
	r.accessCategoryRepo = repository.NewUserAccessCategoryRepository(r.db, r.logger)

	// Create auth middleware
	r.authMiddleware = NewAuthMiddleware(
		userRepo.(*repository.UserRepository),
		r.accessCategoryRepo,
		jwtService,
		r.logger,
		r.config,
		botToken,
	)

	// Initialize bot command service if bot is available
	if bot != nil {
		r.botCommandService = service.NewBotCommandService(
			bot,
			userRepo.(*repository.UserRepository),
			r.logger,
			r.config.Bot.HelpMessage,
			r.config.Bot.StartMessage,
			r.config.Bot.UnknownCommandMessage,
		)
		// Register bot commands
		r.registerBotCommands()
	}

	// Setup protected routes now that auth middleware is initialized
	r.setupProtectedRoutes()
}

// SetGrammarService sets the grammar service (default/fallback).
func (r *Router) SetGrammarService(grammarService *service.GrammarService) {
	r.grammarService = grammarService
}

// SetGrammarServices registers per-bundle grammar services. The map key is
// the bundle ID ("en", "es"). When set, grammar handlers select the service
// matching the user's current course (en_ru→"en", es_ru→"es").
func (r *Router) SetGrammarServices(services map[string]*service.GrammarService) {
	r.grammarServices = services
}

// grammarBundleForCourse returns the grammar bundle ID for a given course code.
func grammarBundleForCourse(courseCode string) string {
	parts := strings.SplitN(courseCode, "_", 2)
	if len(parts) > 0 && parts[0] != "" {
		return parts[0] // "en_ru" → "en", "es_ru" → "es"
	}
	return ""
}

// grammarServiceForRequest returns the GrammarService appropriate for the
// current user's course (from ?course_code= param or stored preference).
// Falls back to the default grammarService if multi-bundle is not configured.
func (r *Router) grammarServiceForRequest(req *http.Request, userID int64) *service.GrammarService {
	if len(r.grammarServices) == 0 {
		return r.grammarService
	}
	courseCode := req.URL.Query().Get("course_code")
	if courseCode == "" {
		courseCode = r.currentCourseCodeForUser(req.Context(), userID)
	}
	bundleID := grammarBundleForCourse(courseCode)
	if svc, ok := r.grammarServices[bundleID]; ok {
		return svc
	}
	return r.grammarService
}

// SetPronunciationService sets pronunciation/TTS service and registers media route.
func (r *Router) SetPronunciationService(pronunciationService *service.PronunciationService) {
	r.pronunciationService = pronunciationService
	r.setupPronunciationMediaRoute()
}

// SetPronunciationServices registers per-bundle pronunciation services. The map key is
// the bundle ID (e.g. "en", "es"). The primary service must also be set via SetPronunciationService.
func (r *Router) SetPronunciationServices(services map[string]*service.PronunciationService) {
	m := make(map[string]pronunciationServiceInterface, len(services))
	for k, v := range services {
		m[k] = v
	}
	r.pronunciationServices = m
}

// SetTTSCircuitBreaker wires the pronunciation/TTS circuit breaker for the admin
// status/reset endpoints. Optional: nil leaves those endpoints reporting "not configured".
func (r *Router) SetTTSCircuitBreaker(ttsCbService *service.CircuitBreakerService) {
	r.ttsCbService = ttsCbService
}

// pronunciationServiceForRequest returns the PronunciationService for the current user's course.
func (r *Router) pronunciationServiceForRequest(req *http.Request, userID int64) pronunciationServiceInterface {
	if len(r.pronunciationServices) > 0 {
		courseCode := req.URL.Query().Get("course_code")
		if courseCode == "" {
			courseCode = r.currentCourseCodeForUser(req.Context(), userID)
		}
		bundleID := grammarBundleForCourse(courseCode)
		if svc, ok := r.pronunciationServices[bundleID]; ok {
			return svc
		}
	}
	return r.pronunciationService
}

// pronunciationServiceForLang returns the PronunciationService registered for the
// given target language/bundle ID (e.g. "en", "es"). Falls back to the primary
// service when targetLang is empty or has no dedicated entry.
func (r *Router) pronunciationServiceForLang(targetLang string) pronunciationServiceInterface {
	targetLang = strings.ToLower(strings.TrimSpace(targetLang))
	if targetLang != "" {
		bundleID := grammarBundleForCourse(targetLang)
		if bundleID == "" {
			bundleID = targetLang
		}
		if svc, ok := r.pronunciationServices[bundleID]; ok {
			return svc
		}
		if svc, ok := r.pronunciationServices[targetLang]; ok {
			return svc
		}
	}
	return r.pronunciationService
}

// SetSpeakingEvaluator sets the speaking evaluation service.
func (r *Router) SetSpeakingEvaluator(evaluator *service.SpeakingEvaluatorService) {
	r.speakingEvaluator = evaluator
}

// SetOTPRepo sets the OTP repository
func (r *Router) SetOTPRepo(otpRepo *repository.WebOTPRepository) {
	r.otpRepo = otpRepo
}

// getAuthMiddleware returns the auth middleware
func (r *Router) getAuthMiddleware() *AuthMiddleware {
	return r.authMiddleware
}

// getRateLimitPolicy creates a rate limit policy from config values
func (r *Router) getRateLimitPolicy(requestsPerWindow, burstMultiplier int) RateLimitPolicy {
	// Use defaults if config values are 0 or negative
	if requestsPerWindow <= 0 {
		requestsPerWindow = 60 // Safe default
	}
	windowMinutes := r.config.WebApp.RateLimitWindowMinutes
	if windowMinutes <= 0 {
		windowMinutes = 1
	}
	if burstMultiplier <= 0 {
		burstMultiplier = 2
	}
	burst := requestsPerWindow * burstMultiplier
	return RateLimitPolicy{
		RequestsPerWindow: requestsPerWindow,
		WindowDuration:    time.Duration(windowMinutes) * time.Minute,
		BurstSize:         burst,
	}
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	lc := r.config.Learning
	if lc.TargetLang == "" {
		lc = config.DefaultLearningConfig()
	}
	spanishVerbForms := r.config.Training.SpanishVerbFormsEnabled && strings.EqualFold(lc.TargetLang, "es")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"learning": map[string]interface{}{
			"pair":                       lc.Pair,
			"native_lang":                lc.NativeLang,
			"target_lang":                lc.TargetLang,
			"app_code":                   lc.AppCode,
			"grammar_bundle_id":          lc.GrammarBundleID,
			"target_lang_name_ru":        learning.TargetLangNameRUAccusative(lc.TargetLang),
			"target_lang_name_en":        learning.TargetLangNameEN(lc.TargetLang),
			"spanish_verb_forms_enabled": spanishVerbForms,
		},
	})
}

// setupRoutes configures all routes
func (r *Router) setupRoutes() {
	// Swagger documentation with custom UI that auto-adds "Bearer " prefix
	r.mux.HandleFunc("/swagger/", r.swaggerHandler)
	r.mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, req, "docs/swagger/swagger.json")
	})

	// Telegram webhook handler (always register if bot is available, webhook can be enabled later)
	if r.bot != nil {
		webhookPath := r.config.Telegram.WebhookPath
		if webhookPath == "" {
			webhookPath = "/webhook"
		}
		r.mux.HandleFunc(webhookPath, r.handleWebhook)
		r.logger.Info("webhook handler registered", zap.String("path", webhookPath), zap.Bool("webhook_enabled", r.config.Telegram.WebhookEnable))
	}

	// Auth routes with rate limiting
	// POST /auth/telegram - moderate limit per IP
	telegramPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthTelegramPerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	telegramMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, telegramPolicy, KeyFuncIP)
	r.mux.HandleFunc("/auth/telegram", telegramMiddleware.Wrap(r.handleAuthTelegram))

	// POST /auth/telegram_unsafe - strict limit per IP, stricter per IP+user_id
	telegramUnsafePolicyIP := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthTelegramUnsafePerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	telegramUnsafePolicyIPUser := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthTelegramUnsafePerIPUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	telegramUnsafeMiddlewareIP := NewRateLimitMiddleware(r.rateLimiter, r.logger, telegramUnsafePolicyIP, KeyFuncIP)
	telegramUnsafeMiddlewareIPUser := NewRateLimitMiddleware(r.rateLimiter, r.logger, telegramUnsafePolicyIPUser, KeyFuncIPAndUserID)
	r.mux.HandleFunc("/auth/telegram_unsafe", telegramUnsafeMiddlewareIP.Wrap(telegramUnsafeMiddlewareIPUser.Wrap(r.handleAuthTelegramUnsafe)))

	// POST /auth/request_otp - strict limit per IP, stricter per IP+username
	requestOTPPolicyIP := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthRequestOTPPerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	requestOTPPolicyIPUser := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthRequestOTPPerIPUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	requestOTPMiddlewareIP := NewRateLimitMiddleware(r.rateLimiter, r.logger, requestOTPPolicyIP, KeyFuncIP)
	requestOTPMiddlewareIPUser := NewRateLimitMiddleware(r.rateLimiter, r.logger, requestOTPPolicyIPUser, KeyFuncIPAndUsername)
	r.mux.HandleFunc("/auth/request_otp", requestOTPMiddlewareIP.Wrap(requestOTPMiddlewareIPUser.Wrap(r.handleAuthRequestOTP)))

	// POST /auth/otp - moderate limit per IP, stricter per IP+user_id
	otpPolicyIP := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthOTPPerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	otpPolicyIPUser := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthOTPPerIPUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	otpMiddlewareIP := NewRateLimitMiddleware(r.rateLimiter, r.logger, otpPolicyIP, KeyFuncIP)
	otpMiddlewareIPUser := NewRateLimitMiddleware(r.rateLimiter, r.logger, otpPolicyIPUser, KeyFuncIPAndUserID)
	r.mux.HandleFunc("/auth/otp", otpMiddlewareIP.Wrap(otpMiddlewareIPUser.Wrap(r.handleAuthOTP)))

	// POST /auth/refresh - moderate limit per IP
	refreshPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthRefreshPerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	refreshMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, refreshPolicy, KeyFuncIP)
	r.mux.HandleFunc("/auth/refresh", refreshMiddleware.Wrap(r.handleAuthRefresh))

	// Health check (includes public learning metadata for webapp title / UI before auth)
	r.mux.HandleFunc("/health", r.handleHealth)

	// Redirect site root to webapp entrypoint.
	r.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			r.handleNotFound(w, req)
			return
		}
		http.Redirect(w, req, "/app", http.StatusFound)
	})

	// Webapp static files (must be registered after API routes to avoid conflicts)
	r.setupWebappRoutes()
}

// setupProtectedRoutes configures protected routes (called after SetDependencies)
func (r *Router) setupProtectedRoutes() {
	auth := r.getAuthMiddleware()

	// Rate limit policies for /app/* routes
	appAPIPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAppAPIPerUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	appChatPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAppChatPerUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	speakingPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitSpeakingPerUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	appAPIMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, appAPIPolicy, KeyFuncIPAndUserIDFromContext)
	appChatMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, appChatPolicy, KeyFuncIPAndUserIDFromContext)
	speakingMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, speakingPolicy, KeyFuncIPAndUserIDFromContext)

	// Protected user routes (wrapped with auth middleware and rate limiting)
	r.mux.HandleFunc("/api/dashboard", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleDashboard)))
	r.mux.HandleFunc("/api/courses", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleCourses)))
	r.mux.HandleFunc("/api/user/courses/current", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleCurrentCourse)))
	r.mux.HandleFunc("/api/user/courses/select", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleSelectCourse)))
	r.mux.HandleFunc("/api/learning/course", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningCourse)))
	r.mux.HandleFunc("/api/linglow/city", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowCity)))
	r.mux.HandleFunc("/api/linglow/daily-route", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowDailyRoute)))
	r.mux.HandleFunc("/api/linglow/review", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowReview)))
	r.mux.HandleFunc("/api/linglow/progress", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowProgress)))
	r.mux.HandleFunc("/api/linglow/srs-shadow", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowSRSShadow)))
	r.mux.HandleFunc("/api/linglow/exercise-attempts", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowExerciseAttempts)))
	r.mux.HandleFunc("/api/linglow/words", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowWords)))
	r.mux.HandleFunc("/api/linglow/word-progress", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowWordLevelProgress)))
	r.mux.HandleFunc("/api/linglow/history", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowHistory)))
	r.mux.HandleFunc("/api/linglow/activity", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowActivity)))
	r.mux.HandleFunc("/api/linglow/stats", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowStats)))
	r.mux.HandleFunc("/api/me", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleMe)))
	r.mux.HandleFunc("/api/linglow/lumi-fact", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLumiFact)))
	r.mux.HandleFunc("/api/linglow/district-extras", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowDistrictExtras)))
	// District conversations (quest + free chat) are available to everyone; only the
	// standalone general AI chat (/api/chat) stays behind the conversation feature.
	r.mux.HandleFunc("/api/linglow/conversation/scenarios", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowConversationScenarios)))
	r.mux.HandleFunc("/api/linglow/conversation/sessions", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowConversationSessions)))
	r.mux.HandleFunc("/api/linglow/conversation/sessions/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLinglowConversationSessionByID)))
	r.mux.HandleFunc("/api/vocab", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVocab)))
	r.mux.HandleFunc("/api/vocab/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVocabDelete)))

	// Learning word sets routes
	r.mux.HandleFunc("/api/learning/words/categories", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningWordsCategories)))
	r.mux.HandleFunc("/api/learning/words/sets", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningWordsSets)))
	r.mux.HandleFunc("/api/learning/words/sets/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningWordsSetDetailOrStudy)))

	// Learning grammar routes
	r.mux.HandleFunc("/api/learning/grammar/offline/manifest", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarOfflineManifest)))
	r.mux.HandleFunc("/api/learning/grammar/offline/training-pack", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarOfflineTrainingPack)))
	r.mux.HandleFunc("/api/learning/grammar/offline/chapters/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarOfflineChapter)))
	r.mux.HandleFunc("/api/learning/grammar/offline/sync-attempts", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarOfflineSyncAttempts)))
	r.mux.HandleFunc("/api/learning/grammar/offline/sync-training-attempts", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarOfflineSyncTrainingAttempts)))
	r.mux.HandleFunc("/api/learning/grammar/categories", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarCategories)))
	r.mux.HandleFunc("/api/learning/grammar/categories/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarChapters)))
	r.mux.HandleFunc("/api/learning/grammar/chapters/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarChapterOrTest)))
	// Reading segment audio is fetched by plain <audio>/new Audio(), so keep it public.
	r.mux.HandleFunc("/api/learning/reading/audio", appAPIMiddleware.Wrap(r.handleLearningReadingAudio))
	r.mux.HandleFunc("/api/learning/reading/image", appAPIMiddleware.Wrap(r.handleLearningReadingImage))
	r.mux.HandleFunc("/api/learning/grammar/tests/submit", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarSubmitTest)))
	r.mux.HandleFunc("/api/learning/grammar/statistics", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarStatistics)))
	r.mux.HandleFunc("/api/learning/grammar/training/availability", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarTrainingAvailability)))
	r.mux.HandleFunc("/api/learning/grammar/training/session/start", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarTrainingStart)))
	r.mux.HandleFunc("/api/learning/grammar/training/session/answer", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarTrainingAnswer)))
	r.mux.HandleFunc("/api/learning/grammar/training/report", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarTrainingReport)))
	r.mux.HandleFunc("/api/learning/grammar/chapter/report", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarChapterReport)))
	r.mux.HandleFunc("/api/learning/grammar/test/report", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarTestReport)))
	r.mux.HandleFunc("/api/content-reports/offline/sync-reports", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleContentReportsOfflineSync)))
	r.mux.HandleFunc("/api/learning/grammar/placement-test", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarPlacementTest)))
	r.mux.HandleFunc("/api/learning/grammar/placement-test/submit", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningGrammarSubmitPlacementTest)))
	r.mux.HandleFunc("/api/learning/reading/categories", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningReadingCategories)))
	r.mux.HandleFunc("/api/learning/reading/categories/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningReadingCategoryTexts)))
	r.mux.HandleFunc("/api/learning/reading/texts/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningReadingTexts)))
	r.mux.HandleFunc("/api/reading/word-lookup", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleReadingWordLookup)))

	speakingAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return speakingMiddleware.Wrap(auth.RequireAuth(r.RequireFeature("speaking")(h)))
	}
	r.mux.HandleFunc("/api/learning/speaking/availability", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningSpeakingAvailability)))
	r.mux.HandleFunc("/api/learning/speaking/categories", speakingAuth(r.handleLearningSpeakingCategories))
	r.mux.HandleFunc("/api/learning/speaking/categories/", speakingAuth(r.handleLearningSpeakingCategoryTasks))
	r.mux.HandleFunc("/api/learning/speaking/sessions", speakingAuth(r.handleLearningSpeakingSessions))
	r.mux.HandleFunc("/api/learning/speaking/sessions/", speakingAuth(r.handleLearningSpeakingSessionByID))

	r.mux.HandleFunc("/api/training/start", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingStart)))
	r.mux.HandleFunc("/api/training/current", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingCurrent)))
	r.mux.HandleFunc("/api/training/reveal", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingReveal)))
	r.mux.HandleFunc("/api/training/answer", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingAnswer)))
	r.mux.HandleFunc("/api/training/report", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingReport)))
	r.mux.HandleFunc("/api/training/upcoming", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingUpcoming)))
	r.mux.HandleFunc("/api/training/offline/pack", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingOfflinePack)))
	r.mux.HandleFunc("/api/training/offline/sync-attempts", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingOfflineSyncAttempts)))
	r.mux.HandleFunc("/api/verb-training/start", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVerbTrainingStart)))
	r.mux.HandleFunc("/api/verb-training/current", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVerbTrainingCurrent)))
	r.mux.HandleFunc("/api/verb-training/answer", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVerbTrainingAnswer)))
	r.mux.HandleFunc("/api/verb-training/upcoming", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVerbTrainingUpcoming)))
	r.mux.HandleFunc("/api/verb-training/forms-by-lemma", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVerbTrainingLemmaForms)))
	r.mux.HandleFunc("/api/chat", appChatMiddleware.Wrap(auth.RequireAuth(r.RequireFeature("conversation")(r.handleChat))))
	r.mux.HandleFunc("/api/settings", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleSettings)))
	r.mux.HandleFunc("/api/settings/notifications", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleNotificationSettings)))
	r.mux.HandleFunc("/api/settings/language", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLanguageSettings)))
	r.mux.HandleFunc("/api/settings/training", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingSettings)))
	r.mux.HandleFunc("/api/tts/word", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTTSWord)))
	r.mux.HandleFunc("/api/internal/tts/pending", r.handleInternalTTSPending)
	r.mux.HandleFunc("/api/internal/tts/audio", r.handleInternalTTSAudio)
	r.mux.HandleFunc("/api/internal/tts/fail", r.handleInternalTTSFail)
	r.mux.HandleFunc("/api/internal/content-reports", r.handleInternalContentReports)
	r.mux.HandleFunc("/api/internal/content-reports/", r.handleInternalContentReportsSubpath)
	r.mux.HandleFunc("/api/internal/content-reports/grammar", r.handleInternalGrammarContentReports)
	r.mux.HandleFunc("/api/internal/content-reports/grammar/resolve-bulk", r.handleInternalGrammarContentReportsResolveBulk)
	r.mux.HandleFunc("/api/internal/training/card/", r.handleInternalTrainingCard)
	r.mux.HandleFunc("/api/internal/tts/regenerate", r.handleInternalTTSRegenerate)
	r.mux.HandleFunc("/api/internal/tts/status", r.handleInternalTTSStatus)
	r.mux.HandleFunc("/api/internal/verb-training/pending", r.handleInternalVerbTrainingPending)

	// Access control routes
	r.mux.HandleFunc("/api/access/me", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleAccessMe)))

	// Admin routes (wrapped with permission guards and rate limiting)
	adminAuth := auth.RequireAuth

	// Words admin routes
	// GET /api/admin/words - read all words
	r.mux.HandleFunc("/api/admin/words", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionWordsReadAll)(r.handleAdminWords))))
	// PUT/DELETE /api/admin/words/{id} - edit words
	r.mux.HandleFunc("/api/admin/words/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionWordsEditAll)(r.handleAdminWord))))
	// TTS status routes: GET requires read, POST actions require edit (checked in handler)
	r.mux.HandleFunc("/api/admin/tts/", appAPIMiddleware.Wrap(adminAuth(r.RequireAnyPermission(PermissionWordsReadAll, PermissionWordsEditAll)(r.handleAdminTTS))))

	// Training admin routes (GET requires words.read_all, POST/PUT/DELETE require words.edit_all)
	r.mux.HandleFunc("/api/admin/training/", appAPIMiddleware.Wrap(adminAuth(r.RequireAnyPermission(PermissionWordsReadAll, PermissionWordsEditAll)(r.handleAdminTraining))))
	r.mux.HandleFunc("/api/admin/training/card/", appAPIMiddleware.Wrap(adminAuth(r.RequireAnyPermission(PermissionWordsReadAll, PermissionWordsEditAll)(r.handleAdminTrainingCard))))
	// Verb-form training cards inventory (read-only, Spanish DB content)
	r.mux.HandleFunc("/api/admin/verb-training/lemmas", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionWordsReadAll)(r.handleAdminVerbTrainingLemmas))))
	r.mux.HandleFunc("/api/admin/verb-training/cards", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionWordsReadAll)(r.handleAdminVerbTrainingCards))))

	// Users admin routes
	r.mux.HandleFunc("/api/admin/users", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionUsersReadAll)(r.handleAdminUsers))))
	r.mux.HandleFunc("/api/admin/users/", appAPIMiddleware.Wrap(adminAuth(r.RequireAnyPermission(PermissionFullAccess, PermissionUsersReadAll)(r.handleAdminUserSubroutes))))

	// Word sets admin routes
	// Categories: GET requires read, POST/PUT/DELETE require edit
	r.mux.HandleFunc("/api/admin/word-set-categories", appAPIMiddleware.Wrap(adminAuth(r.RequireAnyPermission(PermissionWordSetsRead, PermissionWordSetsEdit)(r.handleAdminWordSetCategories))))
	r.mux.HandleFunc("/api/admin/word-set-categories/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionWordSetsEdit)(r.handleAdminWordSetCategories))))
	// Word sets: GET requires read, POST/PUT/DELETE require edit
	r.mux.HandleFunc("/api/admin/word-sets", appAPIMiddleware.Wrap(adminAuth(r.RequireAnyPermission(PermissionWordSetsRead, PermissionWordSetsEdit)(r.handleAdminWordSets))))
	// Word set detail: GET requires read, PUT items requires edit
	r.mux.HandleFunc("/api/admin/word-sets/", appAPIMiddleware.Wrap(adminAuth(r.RequireAnyPermission(PermissionWordSetsRead, PermissionWordSetsEdit)(r.handleAdminWordSetDetailOrSets))))
	// Manual re-sync of legacy word sets into Linglow v2 districts/modules (requires edit)
	r.mux.HandleFunc("/api/admin/word-sets/sync-districts", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionWordSetsEdit)(r.handleAdminWordSetsSyncDistricts))))

	// Access control admin routes (require full_access)
	r.mux.HandleFunc("/api/admin/access/available-permissions", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminAccessAvailablePermissions))))
	r.mux.HandleFunc("/api/admin/access/categories", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminAccessCategories))))
	r.mux.HandleFunc("/api/admin/access/categories/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminAccessCategoryRoutes))))
	r.mux.HandleFunc("/api/admin/access/users/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminAccessUsers))))

	// Stats admin route (require stats.read)
	r.mux.HandleFunc("/api/admin/stats", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionStatsRead)(r.handleAdminStats))))

	// Other admin routes (require full_access)
	r.mux.HandleFunc("/api/admin", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdmin))))
	r.mux.HandleFunc("/api/admin/circuit/reset", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminCircuitReset))))
	r.mux.HandleFunc("/api/admin/circuit/open", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminCircuitOpen))))
	r.mux.HandleFunc("/api/admin/circuit/tts", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminTTSCircuitStatus))))
	r.mux.HandleFunc("/api/admin/circuit/tts/reset", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminTTSCircuitReset))))
	r.mux.HandleFunc("/api/admin/content-reports", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminContentReports))))
	r.mux.HandleFunc("/api/admin/content-reports/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminContentReportByID))))
	r.mux.HandleFunc("/api/admin/linglow/srs-readiness", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminLinglowSRSReadiness))))
	r.mux.HandleFunc("/api/admin/lumi-facts", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminLumiFacts))))

	// Grammar admin routes (require full_access)
	r.mux.HandleFunc("/api/admin/grammar/categories", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminGrammarCategories))))
	r.mux.HandleFunc("/api/admin/grammar/categories/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminGrammarCategoryPublish))))
	r.mux.HandleFunc("/api/admin/grammar/chapters", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminGrammarChapters))))
	r.mux.HandleFunc("/api/admin/grammar/chapters/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminGrammarChapterPublish))))
	r.mux.HandleFunc("/api/admin/grammar/items/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminGrammarItemRename))))
	r.mux.HandleFunc("/api/admin/reading/texts", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminReadingTexts))))
	r.mux.HandleFunc("/api/admin/reading/texts/", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminReadingTexts))))

	// App settings admin routes (require full_access)
	r.mux.HandleFunc("/api/admin/app-settings", appAPIMiddleware.Wrap(adminAuth(r.RequirePermission(PermissionFullAccess)(r.handleAdminAppSettings))))
}

func parseTTSTokens(raw string, logger *zap.Logger) map[string]string {
	m := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return m
	}
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		if logger != nil {
			logger.Warn("failed to parse TTS_INTERNAL_TOKENS_JSON", zap.Error(err))
		}
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		key := strings.ToLower(strings.TrimSpace(k))
		val := strings.TrimSpace(v)
		if key == "" || val == "" {
			continue
		}
		out[key] = val
	}
	return out
}

func maxInt(v, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}

func parseInternalServiceTokens(cfg *config.Config, logger *zap.Logger) map[string]string {
	tokens := parseTTSTokens(cfg.WebApp.InternalServiceTokensJSON, logger)
	raw := strings.TrimSpace(cfg.WebApp.ComplaintsServiceToken)
	if raw == "" {
		return tokens
	}
	if tokens == nil {
		tokens = map[string]string{}
	}
	// COMPLAINTS_SERVICE_TOKEN is the primary release variable.
	if _, exists := tokens["default"]; !exists {
		tokens["default"] = raw
		return tokens
	}
	// Keep backward compatibility if default is already set via WEBAPP_INTERNAL_SERVICE_TOKENS_JSON.
	tokens["complaints"] = raw
	return tokens
}

// corsMiddleware adds CORS headers to allow Swagger UI to make requests
func (r *Router) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Set CORS headers
		origin := req.Header.Get("Origin")

		// Allow requests from localhost origins (for Swagger UI)
		allowedOrigins := []string{
			"http://localhost:8184",
			"http://127.0.0.1:8184",
			"http://localhost:8080",
			"http://127.0.0.1:8080",
		}

		if origin != "" {
			// Check if origin is in allowed list
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
			// If origin matches localhost pattern, allow it
			if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			// Allow Telegram Web App origins (web.telegram.org, etc.)
			if strings.Contains(origin, "telegram.org") || strings.Contains(origin, "t.me") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			// Allow same-origin requests (when app is served from same domain)
			if strings.HasPrefix(origin, "https://") || strings.HasPrefix(origin, "http://") {
				// For production, allow requests from the same domain
				// This handles cases where webapp is served from the same domain as API
				if w.Header().Get("Access-Control-Allow-Origin") == "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		} else {
			// For requests without Origin header (same-origin), allow all
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		// Only set credentials if we have a specific origin (not *)
		if origin != "" && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")) {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, req)
	}
}

// swaggerResponseWriter wraps http.ResponseWriter to inject JavaScript for auto Bearer prefix
type swaggerResponseWriter struct {
	http.ResponseWriter
	buf           []byte
	headerWritten bool
	statusCode    int
}

func (w *swaggerResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.buf = append(w.buf, b...)
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

func (w *swaggerResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.statusCode = statusCode
	w.headerWritten = true

	// Process buffered content only if it's HTML
	if len(w.buf) > 0 {
		contentType := w.Header().Get("Content-Type")
		content := string(w.buf)

		// Only modify HTML content, not JSON, CSS, JS, etc.
		isHTML := strings.Contains(contentType, "text/html") ||
			strings.Contains(content, "<html") ||
			strings.Contains(content, "<!DOCTYPE") ||
			strings.Contains(content, "swagger-ui-bundle.js")

		if isHTML {
			// Inject JavaScript before closing body tag
			js := `<script>
(function() {
    'use strict';
    
    // Function to add Bearer prefix if missing
    function ensureBearerPrefix(value) {
        if (!value) return value;
        const trimmed = value.trim();
        if (trimmed && !trimmed.startsWith('Bearer ')) {
            return 'Bearer ' + trimmed;
        }
        return trimmed;
    }
    
    // Function to setup input field watcher
    function setupInputWatcher(input) {
        if (input.dataset.bearerWatcherAdded) return;
        input.dataset.bearerWatcherAdded = 'true';
        
        // Watch for input changes
        input.addEventListener('input', function() {
            const value = this.value;
            if (value && !value.startsWith('Bearer ') && value.length > 10) {
                // Don't auto-add while typing, wait for blur or authorize
            }
        });
        
        // Add Bearer on blur
        input.addEventListener('blur', function() {
            const value = this.value;
            if (value && !value.startsWith('Bearer ')) {
                this.value = ensureBearerPrefix(value);
            }
        });
        
        // Add Bearer on Enter key
        input.addEventListener('keydown', function(e) {
            if (e.key === 'Enter') {
                const value = this.value;
                if (value && !value.startsWith('Bearer ')) {
                    this.value = ensureBearerPrefix(value);
                }
            }
        });
    }
    
    // Function to intercept Swagger UI authorize
    function interceptSwaggerAuthorize() {
        // Try multiple ways to access Swagger UI system
        let system = null;
        if (window.ui && window.ui.getSystem) {
            system = window.ui.getSystem();
        } else if (window.SwaggerUIBundle) {
            // Try to get system from SwaggerUIBundle
            const swaggerUI = document.querySelector('#swagger-ui');
            if (swaggerUI && swaggerUI.swaggerUISystem) {
                system = swaggerUI.swaggerUISystem;
            }
        }
        
        if (system && system.authActions) {
            const originalAuthorize = system.authActions.authorize;
            if (originalAuthorize) {
                system.authActions.authorize = function(payload) {
                    // Process all auth schemes
                    if (payload) {
                        Object.keys(payload).forEach(function(key) {
                            if (payload[key] && payload[key].value) {
                                const value = payload[key].value;
                                if (typeof value === 'string' && value.trim() && !value.trim().startsWith('Bearer ')) {
                                    payload[key].value = ensureBearerPrefix(value);
                                }
                            }
                        });
                    }
                    return originalAuthorize.call(this, payload);
                };
                return true;
            }
        }
        return false;
    }
    
    // Store refresh token for automatic token refresh
    let refreshToken = null;
    
    // Function to get refresh token from localStorage or Swagger UI
    function getRefreshToken() {
        // Try to get from localStorage first
        const stored = localStorage.getItem('refresh_token');
        if (stored) return stored;
        
        // Try to get from Swagger UI auth state
        try {
            const system = window.ui ? window.ui.getSystem() : null;
            if (system && system.getState) {
                const state = system.getState();
                if (state && state.auth && state.auth.authorized) {
                    // Check if refresh token is stored somewhere in auth state
                    // This is a fallback - refresh token should be stored separately
                }
            }
        } catch (e) {
            console.warn('Could not get refresh token from Swagger UI state', e);
        }
        
        return null;
    }
    
    // Function to refresh access token using refresh token
    async function refreshAccessToken() {
        const refresh = getRefreshToken();
        if (!refresh) {
            console.warn('[Swagger] No refresh token available for automatic token refresh');
            return null;
        }
        
        try {
            const response = await fetch('/auth/refresh', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ refresh_token: refresh })
            });
            
            if (response.ok) {
                const data = await response.json();
                if (data.access_token && data.refresh_token) {
                    // Update tokens
                    refreshToken = data.refresh_token;
                    localStorage.setItem('refresh_token', data.refresh_token);
                    
                    // Update Swagger UI authorization
                    const system = window.ui ? window.ui.getSystem() : null;
                    if (system && system.authActions) {
                        system.authActions.authorize({
                            ApiKeyAuth: {
                                name: 'ApiKeyAuth',
                                value: 'Bearer ' + data.access_token
                            }
                        });
                    }
                    
                    console.log('[Swagger] Access token refreshed automatically');
                    return data.access_token;
                }
            } else {
                console.warn('[Swagger] Failed to refresh token:', response.status);
                // Clear invalid refresh token
                localStorage.removeItem('refresh_token');
                refreshToken = null;
            }
        } catch (error) {
            console.error('[Swagger] Error refreshing token:', error);
        }
        
        return null;
    }
    
    // Function to intercept fetch/XHR requests and add Bearer if needed, and handle 401 errors
    function interceptRequests() {
        // Intercept fetch - only for API requests, not for Swagger UI assets
        const originalFetch = window.fetch;
        window.fetch = function(...args) {
            const url = args[0];
            // Don't intercept requests to Swagger UI assets or doc.json
            if (typeof url === 'string' && (url.includes('/swagger/') || url.includes('swagger-ui') || url.includes('doc.json'))) {
                return originalFetch.apply(this, args);
            }
            
            // Add Bearer prefix if needed
            if (args[1] && args[1].headers) {
                const authHeader = args[1].headers.get ? args[1].headers.get('Authorization') : args[1].headers['Authorization'];
                if (authHeader && !authHeader.startsWith('Bearer ')) {
                    if (args[1].headers instanceof Headers) {
                        args[1].headers.set('Authorization', ensureBearerPrefix(authHeader));
                    } else {
                        args[1].headers['Authorization'] = ensureBearerPrefix(authHeader);
                    }
                }
            }
            
            // Make the request
            return originalFetch.apply(this, args).then(async (response) => {
                // If 401 Unauthorized, try to refresh token and retry once
                if (response.status === 401 && !url.includes('/auth/')) {
                    const newAccessToken = await refreshAccessToken();
                    if (newAccessToken) {
                        // Retry the request with new token
                        const newHeaders = args[1] ? { ...args[1].headers } : {};
                        if (newHeaders instanceof Headers) {
                            newHeaders.set('Authorization', 'Bearer ' + newAccessToken);
                        } else {
                            newHeaders['Authorization'] = 'Bearer ' + newAccessToken;
                        }
                        const retryArgs = [args[0], { ...args[1], headers: newHeaders }];
                        return originalFetch.apply(this, retryArgs);
                    }
                }
                return response;
            });
        };
        
        // Intercept XMLHttpRequest - only for API requests
        const originalSetRequestHeader = XMLHttpRequest.prototype.setRequestHeader;
        const originalSend = XMLHttpRequest.prototype.send;
        
        XMLHttpRequest.prototype.setRequestHeader = function(header, value) {
            // Don't intercept if it's a Swagger UI asset request
            const url = this.responseURL || this._url || '';
            if (url.includes('/swagger/') || url.includes('swagger-ui') || url.includes('doc.json')) {
                return originalSetRequestHeader.call(this, header, value);
            }
            
            if (header.toLowerCase() === 'authorization' && value && !value.startsWith('Bearer ')) {
                value = ensureBearerPrefix(value);
            }
            return originalSetRequestHeader.call(this, header, value);
        };
        
        // Also intercept send to handle 401 responses
        XMLHttpRequest.prototype.send = function(...args) {
            const xhr = this;
            const originalOnReadyStateChange = xhr.onreadystatechange;
            
            xhr.onreadystatechange = async function() {
                if (xhr.readyState === 4 && xhr.status === 401 && !xhr.responseURL.includes('/auth/')) {
                    const newAccessToken = await refreshAccessToken();
                    if (newAccessToken) {
                        // Create new request with refreshed token
                        const newXhr = new XMLHttpRequest();
                        newXhr.open(xhr._method || 'GET', xhr.responseURL);
                        newXhr.setRequestHeader('Authorization', 'Bearer ' + newAccessToken);
                        // Copy other headers
                        for (let header of xhr.getAllResponseHeaders().split('\r\n')) {
                            const parts = header.split(': ');
                            if (parts.length === 2 && parts[0].toLowerCase() !== 'authorization') {
                                newXhr.setRequestHeader(parts[0], parts[1]);
                            }
                        }
                        newXhr.send(args[0]);
                        // Replace xhr properties
                        Object.defineProperty(xhr, 'status', { value: newXhr.status, writable: true });
                        Object.defineProperty(xhr, 'responseText', { value: newXhr.responseText, writable: true });
                        Object.defineProperty(xhr, 'response', { value: newXhr.response, writable: true });
                    }
                }
                if (originalOnReadyStateChange) {
                    originalOnReadyStateChange.apply(this, arguments);
                }
            };
            
            return originalSend.apply(this, args);
        };
    }
    
    // Function to save refresh token when user authorizes
    function setupRefreshTokenSaving() {
        // Monitor Swagger UI authorize actions
        const system = window.ui ? window.ui.getSystem() : null;
        if (system && system.authActions) {
            const originalAuthorize = system.authActions.authorize;
            system.authActions.authorize = function(payload) {
                // Check if this is setting a token
                if (payload && payload.ApiKeyAuth && payload.ApiKeyAuth.value) {
                    const token = payload.ApiKeyAuth.value;
                    // If token starts with Bearer, extract it
                    const actualToken = token.startsWith('Bearer ') ? token.substring(7) : token;
                    // Try to decode JWT to check if it's an access token
                    try {
                        const parts = actualToken.split('.');
                        if (parts.length === 3) {
                            const payload = JSON.parse(atob(parts[1]));
                            // Check if it has exp claim (access token) vs longer exp (refresh token)
                            // For now, we'll rely on user to save refresh token manually or via response
                        }
                    } catch (e) {
                        // Not a valid JWT, ignore
                    }
                }
                return originalAuthorize.call(this, payload);
            };
        }
        
        // Also intercept responses from auth endpoints to save refresh token
        const originalFetch = window.fetch;
        window.fetch = function(...args) {
            const url = args[0];
            if (typeof url === 'string' && (url.includes('/auth/otp') || url.includes('/auth/telegram'))) {
                return originalFetch.apply(this, args).then(async (response) => {
                    if (response.ok) {
                        const data = await response.json();
                        if (data.refresh_token) {
                            refreshToken = data.refresh_token;
                            localStorage.setItem('refresh_token', data.refresh_token);
                            console.log('[Swagger] Refresh token saved for automatic token refresh');
                        }
                    }
                    return response;
                });
            }
            return originalFetch.apply(this, args);
        };
    }
    
    // Wait for Swagger UI to load
    function init() {
        // Load refresh token from localStorage
        refreshToken = getRefreshToken();
        if (refreshToken) {
            console.log('[Swagger] Refresh token loaded from storage');
        }
        
        // Setup request interceptors immediately
        interceptRequests();
        
        // Setup refresh token saving
        setupRefreshTokenSaving();
        
        // Try to intercept Swagger authorize
        if (interceptSwaggerAuthorize()) {
            console.log('[Swagger] Bearer prefix auto-addition enabled');
        }
        
        // Setup MutationObserver to watch for new input fields
        const observer = new MutationObserver(function(mutations) {
            // Look for input fields
            const inputs = document.querySelectorAll('input[type="password"], input[type="text"][name*="token" i], input[placeholder*="token" i], input[placeholder*="Token" i], input[placeholder*="authorization" i], input[placeholder*="Authorization" i]');
            inputs.forEach(setupInputWatcher);
            
            // Also look for authorize buttons and intercept clicks
            const authButtons = document.querySelectorAll('button.authorize, .authorize-btn, [class*="authorize"]');
            authButtons.forEach(function(btn) {
                if (!btn.dataset.bearerClickAdded) {
                    btn.dataset.bearerClickAdded = 'true';
                    btn.addEventListener('click', function() {
                        setTimeout(function() {
                            const inputs = document.querySelectorAll('input[type="password"], input[type="text"]');
                            inputs.forEach(function(input) {
                                const value = input.value;
                                if (value && !value.startsWith('Bearer ')) {
                                    input.value = ensureBearerPrefix(value);
                                }
                            });
                        }, 100);
                    });
                }
            });
        });
        
        observer.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: false
        });
        
        // Initial setup
        setTimeout(function() {
            const inputs = document.querySelectorAll('input[type="password"], input[type="text"]');
            inputs.forEach(setupInputWatcher);
        }, 1000);
    }
    
    // Start initialization
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
    
    // Also try after a delay to catch late-loading Swagger UI
    setTimeout(init, 2000);
})();
</script>`
			// Insert before </body> tag
			if idx := strings.LastIndex(content, "</body>"); idx > 0 {
				content = content[:idx] + js + "\n" + content[idx:]
				w.buf = []byte(content)
			} else {
				// If no </body> tag, append before </html> or at the end
				if idx := strings.LastIndex(content, "</html>"); idx > 0 {
					content = content[:idx] + js + "\n" + content[idx:]
					w.buf = []byte(content)
				} else {
					// Append at the end
					w.buf = append(w.buf, []byte(js)...)
				}
			}
		}
	}

	// Write header and buffered content
	w.ResponseWriter.WriteHeader(w.statusCode)
	if len(w.buf) > 0 {
		w.ResponseWriter.Write(w.buf)
		w.buf = nil
	}
}

func (w *swaggerResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// swaggerHandler serves Swagger UI with custom JavaScript to auto-add "Bearer " prefix
func (r *Router) swaggerHandler(w http.ResponseWriter, req *http.Request) {
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)

	// For HTML pages only (not static files), wrap response writer to inject JavaScript
	path := req.URL.Path
	isHTMLPage := (path == "/swagger/" || path == "/swagger/index.html" || path == "/swagger") &&
		!strings.Contains(path, ".") &&
		!strings.HasSuffix(path, ".js") &&
		!strings.HasSuffix(path, ".css") &&
		!strings.HasSuffix(path, ".json") &&
		!strings.HasSuffix(path, ".png") &&
		!strings.HasSuffix(path, ".ico")

	if isHTMLPage {
		wrapped := &swaggerResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		swaggerHandler.ServeHTTP(wrapped, req)
		return
	}

	swaggerHandler.ServeHTTP(w, req)
}

// registerBotCommands registers bot commands with Telegram
func (r *Router) registerBotCommands() {
	if r.bot == nil {
		return
	}

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начать работу с ботом"},
		{Command: "help", Description: "Показать справку"},
		{Command: "unsubscribe", Description: "Отписаться от уведомлений"},
		{Command: "notification", Description: "Настроить периодичность уведомлений"},
	}

	cmdConfig := tgbotapi.NewSetMyCommands(commands...)
	if _, err := r.bot.Request(cmdConfig); err != nil {
		r.logger.Error("failed to register bot commands", zap.Error(err))
	} else {
		r.logger.Info("bot commands registered successfully")
	}
}

// handleWebhook handles Telegram webhook updates
func (r *Router) handleWebhook(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.botCommandService == nil {
		r.logger.Warn("bot command service not initialized, webhook request ignored")
		http.Error(w, "Bot command service not initialized", http.StatusInternalServerError)
		return
	}

	var update tgbotapi.Update
	if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
		r.logger.Error("failed to decode webhook update", zap.Error(err))
		http.Error(w, "Invalid update", http.StatusBadRequest)
		return
	}

	// Log update details
	updateLog := []zap.Field{
		zap.Int("update_id", update.UpdateID),
		zap.Bool("has_message", update.Message != nil),
		zap.Bool("has_callback", update.CallbackQuery != nil),
	}
	if update.Message != nil {
		updateLog = append(updateLog,
			zap.String("message_text", update.Message.Text),
			zap.Bool("is_command", update.Message.IsCommand()),
		)
		if update.Message.IsCommand() {
			updateLog = append(updateLog,
				zap.String("command", update.Message.Command()),
				zap.String("command_args", update.Message.CommandArguments()),
			)
		}
	}
	r.logger.Info("received webhook update", updateLog...)

	// Handle update synchronously to ensure proper error handling
	// Telegram expects quick response, but we need to process commands
	r.botCommandService.HandleUpdate(update)

	// Respond to Telegram
	w.WriteHeader(http.StatusOK)
}

// handleNotFound handles 404 errors
func (r *Router) handleNotFound(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error": "Not Found", "message": "The requested resource was not found", "path": "%s"}`, req.URL.Path)
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply CORS middleware to all requests
	r.corsMiddleware(func(w http.ResponseWriter, req *http.Request) {
		r.mux.ServeHTTP(w, req)
	})(w, req)
}

// handleAuthTelegram handles Telegram WebApp initData authentication
// @Summary      Аутентификация через Telegram WebApp
// @Description  Аутентификация пользователя через Telegram WebApp initData. Возвращает пару JWT токенов (access и refresh) для авторизованного пользователя.
// @Tags         Auth
// @Accept       application/x-www-form-urlencoded
// @Produce      application/json
// @Param        initData  formData  string  true  "Telegram WebApp initData (подписанные данные от Telegram)"
// @Success      200  {object}  map[string]interface{}  "Успешная аутентификация с JWT токенами"
// @Failure      400  {string}  string  "Неверный запрос (отсутствует initData)"
// @Failure      401  {string}  string  "Неавторизован (неверный initData)"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /auth/telegram [post]
func (r *Router) handleAuthTelegram(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := req.ParseForm(); err != nil {
		r.logger.Error("failed to parse form", zap.Error(err))
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	initData := req.FormValue("initData")
	if initData == "" {
		http.Error(w, "initData is required", http.StatusBadRequest)
		return
	}

	// Validate initData and get user ID
	auth := r.getAuthMiddleware()
	telegramID, err := auth.ValidateTelegramInitData(initData)
	if err != nil {
		r.logger.Warn("failed to validate initData", zap.Error(err))
		http.Error(w, "Invalid initData", http.StatusUnauthorized)
		return
	}

	// Get or create user
	var user *models.User
	if r.getOrCreateUserForTelegram != nil {
		user, err = r.getOrCreateUserForTelegram(telegramID)
	} else {
		userRepo := r.userRepo.(*repository.UserRepository)
		user, err = userRepo.GetOrCreateUser(telegramID)
	}
	if err != nil {
		r.logger.Error("failed to get/create user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate JWT token pair
	var accessToken, refreshToken string
	if r.generateTokenPairForTelegram != nil {
		accessToken, refreshToken, err = r.generateTokenPairForTelegram(user.ID, user.TelegramID)
	} else {
		accessToken, refreshToken, err = auth.GenerateTokenPair(user.ID, user.TelegramID)
	}
	if err != nil {
		r.logger.Error("failed to generate JWT tokens", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response with JWT tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Authentication successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// handleAuthTelegramUnsafe handles Telegram WebApp authentication using initDataUnsafe (less secure, fallback)
// @Summary      Аутентификация через Telegram WebApp (небезопасный метод)
// @Description  Аутентификация пользователя через user_id напрямую (менее безопасный метод, используется как fallback). Возвращает пару JWT токенов (access и refresh) для авторизованного пользователя.
// @Tags         Auth
// @Accept       application/x-www-form-urlencoded
// @Produce      application/json
// @Param        user_id  formData  string  true  "Telegram User ID"
// @Success      200  {object}  map[string]interface{}  "Успешная аутентификация с JWT токенами"
// @Failure      400  {string}  string  "Неверный запрос (отсутствует user_id)"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /auth/telegram_unsafe [post]
func (r *Router) handleAuthTelegramUnsafe(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("handleAuthTelegramUnsafe called",
		zap.String("method", req.Method),
		zap.String("path", req.URL.Path),
		zap.String("remote_addr", req.RemoteAddr))

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID directly from initDataUnsafe (less secure, but works when initData is not available)
	userIDStr := req.FormValue("user_id")
	r.logger.Info("received user_id", zap.String("user_id", userIDStr))

	if userIDStr == "" {
		r.logger.Warn("user_id is empty")
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	telegramID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		r.logger.Warn("invalid user_id", zap.String("user_id", userIDStr), zap.Error(err))
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	r.logger.Info("authenticating via initDataUnsafe", zap.Int64("telegram_id", telegramID))

	// Get or create user
	var user *models.User
	if r.getOrCreateUserForTelegram != nil {
		user, err = r.getOrCreateUserForTelegram(telegramID)
	} else {
		userRepo := r.userRepo.(*repository.UserRepository)
		user, err = userRepo.GetOrCreateUser(telegramID)
	}
	if err != nil {
		r.logger.Error("failed to get/create user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate JWT token pair
	auth := r.getAuthMiddleware()
	var accessToken, refreshToken string
	if r.generateTokenPairForTelegram != nil {
		accessToken, refreshToken, err = r.generateTokenPairForTelegram(user.ID, user.TelegramID)
	} else {
		accessToken, refreshToken, err = auth.GenerateTokenPair(user.ID, user.TelegramID)
	}
	if err != nil {
		r.logger.Error("failed to generate JWT tokens", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response with JWT tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Authentication successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// handleAuthRequestOTP and handleAuthOTP are implemented in auth_otp.go

// handleDashboard and handleChat are implemented in dashboard.go
// handleVocab and handleVocabDelete are implemented in vocab.go
// handleTrainingStart, handleTrainingCurrent, handleTrainingReveal, handleTrainingAnswer are implemented in training.go
// handleAdmin, handleAdminCircuitReset, handleAdminTraining are implemented in admin.go
