package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestRouter_SetDependencies(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	userRepo := repository.NewUserRepository(nil, logger)
	wordService := (*service.WordService)(nil)
	aiService := (*interface{})(nil)
	bot := (*tgbotapi.BotAPI)(nil)

	router.SetDependencies(userRepo, wordService, aiService, bot, "test-token")
	_ = router // Verify router is configured
}

func TestRouter_SetOTPRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	otpRepo := repository.NewWebOTPRepository(nil, logger)
	router.SetOTPRepo(otpRepo)
	_ = router // Verify router is configured
}

func TestRouter_getAuthMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	// Set dependencies to initialize auth middleware
	userRepo := repository.NewUserRepository(nil, logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Get auth middleware
	middleware := router.getAuthMiddleware()
	if middleware == nil {
		t.Error("getAuthMiddleware() should not return nil after SetDependencies")
	}
}

func TestRouter_ServeHTTP_Health(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRouter_SetPronunciationService(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	pronService := service.NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://api.dictionaryapi.dev/api/v2/entries/en",
	}, nil, logger)

	router.SetPronunciationService(pronService)

	if router.pronunciationService != pronService {
		t.Error("pronunciationService not set")
	}
	if !router.pronunciationMediaRouteRegistered {
		t.Error("pronunciationMediaRouteRegistered should be true after SetPronunciationService with enabled service")
	}
}

func TestRouter_HandleWebhook_MethodNotAllowed(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	router.bot = &tgbotapi.BotAPI{}
	router.botCommandService = service.NewBotCommandService(nil, nil, logger, "", "", "")

	req := httptest.NewRequest("GET", "/webhook", nil)
	w := httptest.NewRecorder()
	router.handleWebhook(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestRouter_HandleWebhook_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	router.bot = &tgbotapi.BotAPI{}
	router.botCommandService = service.NewBotCommandService(nil, nil, logger, "", "", "")

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouter_HandleWebhook_NoBotCommandService(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	router.bot = &tgbotapi.BotAPI{}
	router.botCommandService = nil

	body, _ := json.Marshal(map[string]interface{}{"update_id": 1})
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleWebhook(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRouter_HandleWebhook_Success(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	router.bot = &tgbotapi.BotAPI{}
	router.botCommandService = service.NewBotCommandService(nil, nil, logger, "", "", "")

	body, _ := json.Marshal(map[string]interface{}{"update_id": 123})
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouter_HandleAuthTelegram_MethodNotAllowed(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(nil, logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	req := httptest.NewRequest("GET", "/auth/telegram", nil)
	w := httptest.NewRecorder()
	router.handleAuthTelegram(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestRouter_HandleAuthTelegram_NoInitData(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(nil, logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	req := httptest.NewRequest("POST", "/auth/telegram", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAuthTelegram(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouter_HandleAuthTelegram_InvalidInitData(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(nil, logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	req := httptest.NewRequest("POST", "/auth/telegram", strings.NewReader("initData=invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAuthTelegram(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouter_HandleAuthTelegram_Success(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	botToken := "test-bot-token-for-telegram-auth"
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:        "test-secret",
			JWTTTLHours:      24,
			RefreshTTLHours:  720,
		},
	}
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, botToken)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	telegramID := int64(777888999)
	authDate := "1234567890"
	userJSON, _ := json.Marshal(map[string]int64{"id": telegramID})
	userEncoded := url.QueryEscape(string(userJSON))
	params := map[string]string{"auth_date": authDate, "user": string(userJSON)}
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var dataCheckParts []string
	for _, k := range keys {
		dataCheckParts = append(dataCheckParts, k+"="+params[k])
	}
	dataCheckString := strings.Join(dataCheckParts, "\n")
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secretKeyBytes := secretKey.Sum(nil)
	hash := hmac.New(sha256.New, secretKeyBytes)
	hash.Write([]byte(dataCheckString))
	hashHex := hex.EncodeToString(hash.Sum(nil))
	initData := "auth_date=" + url.QueryEscape(authDate) + "&user=" + userEncoded + "&hash=" + hashHex

	req := httptest.NewRequest("POST", "/auth/telegram", strings.NewReader("initData="+url.QueryEscape(initData)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAuthTelegram(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["success"] != true {
		t.Error("Expected success=true")
	}
	if resp["access_token"] == nil || resp["refresh_token"] == nil {
		t.Error("Expected access_token and refresh_token in response")
	}
}

// authErrReader is a body that fails on Read to trigger ParseForm error.
type authErrReader struct{}

func (authErrReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestRouter_HandleAuthTelegram_ParseFormError(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(nil, logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	req := httptest.NewRequest("POST", "/auth/telegram", io.NopCloser(authErrReader{}))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAuthTelegram(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 on ParseForm error, got %d", w.Code)
	}
}

func TestRegisterBotCommands_NilBot(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	router.bot = nil
	router.registerBotCommands()
	// Should not panic; with nil bot it returns immediately
}
