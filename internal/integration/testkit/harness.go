package testkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/web"

	"go.uber.org/zap"
)

// Harness holds DB, Router and helpers for integration tests
type Harness struct {
	T      *testing.T
	DB     *database.DB
	Router *web.Router
	Logger *zap.Logger
	Cfg    *config.Config
}

// HarnessOpt configures Harness creation
type HarnessOpt func(*harnessConfig)

type harnessConfig struct {
	aiServiceURL string
	aiAPIKey     string
	aiPrompt     string
}

// WithAIService sets AI service URL (for mock or real). If empty, a no-op mock is used.
func WithAIService(baseURL, apiKey, prompt string) HarnessOpt {
	return func(c *harnessConfig) {
		c.aiServiceURL = baseURL
		c.aiAPIKey = apiKey
		c.aiPrompt = prompt
	}
}

// NewHarness creates a test harness with Postgres and full app router
func NewHarness(t *testing.T, opts ...HarnessOpt) *Harness {
	t.Helper()

	pc, cleanup := StartPostgres(t)
	t.Cleanup(cleanup)

	logger := zap.NewNop()
	cfg := &harnessConfig{
		aiServiceURL: "http://localhost:9999", // placeholder, won't be called if wordService not used
		aiAPIKey:     "test-key",
		aiPrompt:     "You are a test assistant.",
	}
	for _, o := range opts {
		o(cfg)
	}

	var db *database.DB
	var connErr error
	for attempt := 0; attempt < 10; attempt++ {
		db, connErr = database.NewWithConfig("postgres", "", pc.DSN(), logger)
		if connErr == nil {
			break
		}
		if attempt < 9 {
			time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
		}
	}
	if connErr != nil {
		t.Fatalf("failed to connect/migrate postgres: %v", connErr)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn := db.GetConnection()

	// Repositories
	userRepo := repository.NewUserRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	cbRepo := repository.NewCircuitBreakerRepository(conn, logger)
	otpRepo := repository.NewWebOTPRepository(conn, logger)
	grammarContentRepo := repository.NewGrammarContentRepository(logger)
	grammarPublishRepo := repository.NewGrammarPublishRepository(conn, logger)
	grammarAttemptRepo := repository.NewGrammarAttemptRepository(conn, logger)

	// Services
	aiService := ai.NewService(cfg.aiServiceURL, "test-model", cfg.aiAPIKey, cfg.aiPrompt, logger)
	wordService := service.NewWordService(wordRepo, trainingCardRepo, userCardRepo, aiService, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	grammarService := service.NewGrammarService(grammarContentRepo, grammarPublishRepo, grammarAttemptRepo, logger)

	appCfg := testConfig()
	router := web.NewRouter(logger, appCfg, conn, trainingService, srsService, optionsService, cbService)
	router.SetDependencies(userRepo, wordService, aiService, nil, "test-bot-token")
	router.SetOTPRepo(otpRepo)
	router.SetGrammarService(grammarService)

	return &Harness{
		T:      t,
		DB:     db,
		Router: router,
		Logger: logger,
		Cfg:   appCfg,
	}
}

func testConfig() *config.Config {
	return &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:                 "test-jwt-secret-for-integration",
			JWTTTLHours:               24,
			RefreshTTLHours:           720,
			RateLimitAppAPIPerUser:    2000, // high limit for integration tests (60 cards × 4 req/card)
			RateLimitAppChatPerUser:   200,
			RateLimitBurstMultiplier:  2,
			RateLimitWindowMinutes:    1,
		},
		Admin: config.AdminConfig{TelegramID: 0},
	}
}

// AuthAsUser authenticates via /auth/telegram_unsafe and returns Bearer token
func (h *Harness) AuthAsUser(telegramID int64) string {
	h.T.Helper()

	form := url.Values{}
	form.Set("user_id", strconv.FormatInt(telegramID, 10))
	req := httptest.NewRequest(http.MethodPost, "/auth/telegram_unsafe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		h.T.Fatalf("auth failed: status %d, body %s", w.Code, w.Body.String())
	}

	var resp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		h.T.Fatalf("auth response decode: %v", err)
	}
	if resp.AccessToken == "" {
		h.T.Fatal("auth response missing access_token")
	}
	return "Bearer " + resp.AccessToken
}

// GetConnection returns the underlying *sql.DB for direct queries
func (h *Harness) GetConnection() *database.DB {
	return h.DB
}
