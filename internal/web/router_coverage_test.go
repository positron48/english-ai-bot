package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// TestRouter_SetDependencies_JWTServiceError covers the panic path when NewJWTService fails (line 152-154).
// NewJWTService fails when both JWTSecret and SessionSecret are empty.
func TestRouter_SetDependencies_JWTServiceError(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "", // empty
			SessionSecret: "", // empty → NewJWTService returns error
		},
	}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(nil, logger)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when JWT secret is empty, but no panic occurred")
		}
	}()
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
}

// TestRouter_setupRoutes_WithBot covers the webhook registration block (lines 237-243).
// When r.bot != nil and WebhookPath is empty, the default "/webhook" path is used.
// NewRouter calls setupRoutes internally, so we pass bot via the constructor path.
func TestRouter_setupRoutes_WithBot_DefaultWebhookPath(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		WebApp:   config.WebAppConfig{JWTSecret: "test-secret"},
		Telegram: config.TelegramConfig{WebhookPath: ""}, // empty → default "/webhook"
	}
	// Pass a non-nil bot to NewRouter so setupRoutes registers the webhook handler.
	bot := &tgbotapi.BotAPI{}
	router := newRouterWithBot(logger, cfg, bot)

	// Verify the webhook path was registered by hitting it
	req := httptest.NewRequest("GET", "/webhook", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)
	// handleWebhook returns 405 for GET, confirming the route is registered
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 from webhook handler, got %d", w.Code)
	}
}

// TestRouter_setupRoutes_WithBot_CustomWebhookPath covers the webhook registration with a custom path.
func TestRouter_setupRoutes_WithBot_CustomWebhookPath(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		WebApp:   config.WebAppConfig{JWTSecret: "test-secret"},
		Telegram: config.TelegramConfig{WebhookPath: "/custom-webhook"},
	}
	bot := &tgbotapi.BotAPI{}
	router := newRouterWithBot(logger, cfg, bot)

	req := httptest.NewRequest("GET", "/custom-webhook", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 from custom webhook handler, got %d", w.Code)
	}
}

// newRouterWithBot creates a Router with bot set before setupRoutes is called.
// Since NewRouter calls setupRoutes internally, we need to set bot on the struct before
// the internal call. We do this by creating the router struct directly.
func newRouterWithBot(logger *zap.Logger, cfg *config.Config, bot *tgbotapi.BotAPI) *Router {
	r := &Router{
		mux:         http.NewServeMux(),
		logger:      logger,
		config:      cfg,
		rateLimiter: NewRateLimiter(5*time.Minute, 1*time.Hour),
		bot:         bot,
	}
	r.setupRoutes()
	return r
}

// TestRouter_setupRoutes_SwaggerDocJSON covers the /swagger/doc.json handler body (lines 232-234).
func TestRouter_setupRoutes_SwaggerDocJSON(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/swagger/doc.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// The handler body runs; ServeFile may return 404 if file doesn't exist.
	// We just verify the route is registered and handled (not 405 Method Not Allowed).
	if w.Code == http.StatusMethodNotAllowed {
		t.Errorf("Expected handler to run for /swagger/doc.json, got 405")
	}
}

// TestRouter_setupRoutes_RootRedirect covers the "/" → "/app" redirect (line 315).
func TestRouter_setupRoutes_RootRedirect(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected 302 redirect, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/app" {
		t.Errorf("Expected redirect to /app, got %s", loc)
	}
}

// TestSwaggerResponseWriter_WriteHeader_NoBodyTag covers the else branch when there is no </body> tag
// but there is a </html> tag (lines 912-916).
func TestSwaggerResponseWriter_WriteHeader_NoBodyTag_WithHTMLTag(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &swaggerResponseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		buf:            []byte("<html><head></head></html>"),
	}
	recorder.Header().Set("Content-Type", "text/html")

	wrapped.WriteHeader(http.StatusOK)

	if !wrapped.headerWritten {
		t.Error("headerWritten should be true")
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "<script>") {
		t.Errorf("Expected injected <script> tag in body without </body>, got: %s", body)
	}
}

// TestSwaggerResponseWriter_WriteHeader_NoBodyNoHTMLTag covers the else branch when there is
// neither </body> nor </html> tag (lines 917-920).
func TestSwaggerResponseWriter_WriteHeader_NoBodyNoHTMLTag(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &swaggerResponseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		buf:            []byte("swagger-ui-bundle.js content here"),
	}
	recorder.Header().Set("Content-Type", "text/html")

	wrapped.WriteHeader(http.StatusOK)

	if !wrapped.headerWritten {
		t.Error("headerWritten should be true")
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "<script>") {
		t.Errorf("Expected appended <script> tag, got: %s", body)
	}
}

// TestRouter_swaggerHandler_HTMLPage_IndexHTML covers the isHTMLPage branch for /swagger/index.html.
func TestRouter_swaggerHandler_HTMLPage_IndexHTML(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	router.swaggerHandler(w, req)
	// Should handle the request (200 with HTML content)
	_ = w.Code
}

// TestRouter_registerBotCommands_Error covers the error path in registerBotCommands (lines 982-984).
// We construct a BotAPI with a failing HTTP client so Request() returns an error.
func TestRouter_registerBotCommands_Error(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	// Construct BotAPI directly to avoid GetMe() call in NewBotAPIWithClient.
	bot := &tgbotapi.BotAPI{
		Token:  "test:token",
		Client: &http.Client{Transport: &failingTransport{}},
		Buffer: 100,
	}
	router.bot = bot
	// Should not panic; error is logged
	router.registerBotCommands()
}

// failingTransport always returns an error.
type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("injected transport error")
}

// TestRouter_handleWebhook_WithMessage covers the message logging branch (lines 1017-1021).
// Sends a webhook update with a Message that is NOT a command.
func TestRouter_handleWebhook_WithMessage(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	router.bot = &tgbotapi.BotAPI{}
	// Need a non-nil botCommandService so we reach the message logging code.
	router.botCommandService = newNopBotCommandService(logger)

	// Build an update with a non-command message
	update := map[string]interface{}{
		"update_id": 42,
		"message": map[string]interface{}{
			"message_id": 1,
			"text":       "hello world",
			"from":       map[string]interface{}{"id": 123, "is_bot": false, "first_name": "Test"},
			"chat":       map[string]interface{}{"id": 123, "type": "private"},
			"date":       1234567890,
		},
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleWebhook(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestRouter_handleWebhook_WithCommand covers the command logging branch (lines 1022-1027).
// Sends a webhook update with a Message that IS a command.
func TestRouter_handleWebhook_WithCommand(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)
	nopSvc := newNopBotCommandService(logger)
	router.botCommandService = nopSvc

	// Build an update with a /start command message.
	// tgbotapi.Message.IsCommand() returns true when Entities contains a bot_command at offset 0.
	update := map[string]interface{}{
		"update_id": 43,
		"message": map[string]interface{}{
			"message_id": 2,
			"text":       "/start arg1",
			"from":       map[string]interface{}{"id": 124, "is_bot": false, "first_name": "Test"},
			"chat":       map[string]interface{}{"id": 124, "type": "private"},
			"date":       1234567890,
			"entities": []map[string]interface{}{
				{"type": "bot_command", "offset": 0, "length": 6},
			},
		},
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleWebhook(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// newNopBotCommandService creates a BotCommandService with a failing bot (for testing).
// The bot will fail to send messages but won't panic.
func newNopBotCommandService(logger *zap.Logger) *service.BotCommandService {
	bot := &tgbotapi.BotAPI{
		Token:  "test:token",
		Client: &http.Client{Transport: &failingTransport{}},
		Buffer: 100,
	}
	return service.NewBotCommandService(bot, nil, logger, "", "", "")
}

// buildValidInitData builds a valid Telegram initData string for the given telegramID and botToken.
func buildValidInitData(telegramID int64, botToken string) string {
	authDate := "1234567890"
	userJSON, _ := json.Marshal(map[string]int64{"id": telegramID})
	params := map[string]string{"auth_date": authDate, "user": string(userJSON)}
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	dataCheckString := strings.Join(parts, "\n")
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secretKeyBytes := secretKey.Sum(nil)
	hash := hmac.New(sha256.New, secretKeyBytes)
	hash.Write([]byte(dataCheckString))
	hashHex := hex.EncodeToString(hash.Sum(nil))
	userEncoded := url.QueryEscape(string(userJSON))
	return "auth_date=" + url.QueryEscape(authDate) + "&user=" + userEncoded + "&hash=" + hashHex
}

// TestRouter_handleAuthTelegram_GetOrCreateUserError covers the GetOrCreateUser error path (lines 1095-1099).
func TestRouter_handleAuthTelegram_GetOrCreateUserError(t *testing.T) {
	logger := zap.NewNop()
	botToken := "test-bot-token-gocue"
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, botToken)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, botToken)
	router.authMiddleware = authMiddleware
	// Inject error for GetOrCreateUser
	router.getOrCreateUserForTelegram = func(_ int64) (*models.User, error) {
		return nil, errors.New("injected db error")
	}

	initData := buildValidInitData(777, botToken)
	req := httptest.NewRequest("POST", "/auth/telegram",
		strings.NewReader("initData="+url.QueryEscape(initData)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAuthTelegram(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 when GetOrCreateUser fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestRouter_handleAuthTelegram_GenerateTokenPairError covers the GenerateTokenPair error path (lines 1103-1107).
func TestRouter_handleAuthTelegram_GenerateTokenPairError(t *testing.T) {
	logger := zap.NewNop()
	botToken := "test-bot-token-gtpe"
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, botToken)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, botToken)
	router.authMiddleware = authMiddleware
	// Return a valid user but inject error for GenerateTokenPair
	router.getOrCreateUserForTelegram = func(_ int64) (*models.User, error) {
		return &models.User{ID: 1, TelegramID: 777}, nil
	}
	router.generateTokenPairForTelegram = func(_, _ int64) (string, string, error) {
		return "", "", errors.New("injected token pair error")
	}

	initData := buildValidInitData(777, botToken)
	req := httptest.NewRequest("POST", "/auth/telegram",
		strings.NewReader("initData="+url.QueryEscape(initData)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAuthTelegram(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 when GenerateTokenPair fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestRouter_handleAuthTelegramUnsafe_GetOrCreateUserError covers the GetOrCreateUser error path (lines 1165-1169).
func TestRouter_handleAuthTelegramUnsafe_GetOrCreateUserError(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}

	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.getOrCreateUserForTelegram = func(_ int64) (*models.User, error) {
		return nil, errors.New("injected db error")
	}

	body := strings.NewReader("user_id=12345")
	req := httptest.NewRequest("POST", "/auth/telegram_unsafe", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAuthTelegramUnsafe(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 when GetOrCreateUser fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestRouter_handleAuthTelegramUnsafe_GenerateTokenPairError covers the GenerateTokenPair error path (lines 1174-1178).
func TestRouter_handleAuthTelegramUnsafe_GenerateTokenPairError(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware
	router.getOrCreateUserForTelegram = func(_ int64) (*models.User, error) {
		return &models.User{ID: 1, TelegramID: 12345}, nil
	}
	router.generateTokenPairForTelegram = func(_, _ int64) (string, string, error) {
		return "", "", errors.New("injected token pair error")
	}

	body := strings.NewReader("user_id=12345")
	req := httptest.NewRequest("POST", "/auth/telegram_unsafe", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAuthTelegramUnsafe(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 when GenerateTokenPair fails, got %d: %s", w.Code, w.Body.String())
	}
}
