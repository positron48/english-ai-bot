package web

// Tests to cover error paths in auth_otp.go that are not covered by existing tests.

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// TestHandleAuthRequestOTP_GenerateOTPError covers lines 87-95 in auth_otp.go:
// the GenerateOTP error path when the OTP repository DB is broken.
func TestHandleAuthRequestOTP_GenerateOTPError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	goodDB := testutil.SetupTestDB(t)
	badDB := badDBConn(t)

	userRepo := repository.NewUserRepository(goodDB, logger)
	otpRepoBad := repository.NewWebOTPRepository(badDB, logger)

	// Create a real user so GetUserByUsernameOrID succeeds
	user, err := userRepo.GetOrCreateUser(55551)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	usernameOrID := strconv.FormatInt(user.TelegramID, 10)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
			OTPTTLSeconds:   300,
		},
	}

	router := NewRouter(logger, cfg, goodDB, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepoBad)

	req := httptest.NewRequest("POST", "/auth/request_otp", strings.NewReader("username="+usernameOrID))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GenerateOTP fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAuthRequestOTP_BotSendError covers lines 109-117 in auth_otp.go:
// the bot.Send error path when the Telegram bot API returns an error.
func TestHandleAuthRequestOTP_BotSendError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	user, err := userRepo.GetOrCreateUser(55552)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	usernameOrID := strconv.FormatInt(user.TelegramID, 10)

	// Mock server that returns success for getMe (bot init) but error for sendMessage
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.Path, "getMe") {
			// Return valid bot info for initialization
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":123456,"is_bot":true,"first_name":"TestBot","username":"testbot"}}`))
		} else {
			// Return Telegram API error for sendMessage and other calls
			_, _ = w.Write([]byte(`{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`))
		}
	}))
	defer mockServer.Close()

	endpoint := mockServer.URL + "/bot%s/%s"
	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint("test-token", endpoint)
	if err != nil {
		t.Fatalf("NewBotAPIWithAPIEndpoint: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
			OTPTTLSeconds:   300,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, bot, "test-token")
	router.SetOTPRepo(otpRepo)

	req := httptest.NewRequest("POST", "/auth/request_otp", strings.NewReader("username="+usernameOrID))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when bot.Send fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAuthOTP_GetUserByIDNilUser covers lines 214-222 in auth_otp.go:
// the GetUserByID error path after OTP validation succeeds.
// We use a broken user repo so GetUserByID returns an error.
func TestHandleAuthOTP_GetUserByIDNilUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Call badDBConn FIRST to ensure truncation happens before OTP generation
	badDB := badDBConn(t)

	goodDB := testutil.SetupTestDB(t)

	userRepoGood := repository.NewUserRepository(goodDB, logger)
	otpRepo := repository.NewWebOTPRepository(goodDB, logger)

	// Create user and generate OTP after badDBConn (which truncates shared DB)
	user, err := userRepoGood.GetOrCreateUser(55553)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	code, _, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(goodDB, logger)

	// Use a broken user repo so GetUserByID returns an error
	userRepoBad := repository.NewUserRepository(badDB, logger)
	authMiddleware := NewAuthMiddleware(userRepoBad, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, goodDB, nil, nil, nil, nil)
	router.SetDependencies(userRepoBad, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/auth/otp",
		strings.NewReader("user_id="+strconv.FormatInt(user.ID, 10)+"&code="+code))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthOTP(w, req)

	// GetUserByID fails → 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetUserByID fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAuthRequestOTP_ParseFormError covers lines 42-49 in auth_otp.go:
// the ParseForm error path when the URL contains invalid percent-encoding.
func TestHandleAuthRequestOTP_ParseFormError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	// Malformed URL with invalid percent-encoding triggers ParseForm error
	req := httptest.NewRequest("POST", "/auth/request_otp?username=%ZZ", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when ParseForm fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAuthOTP_ParseFormError covers lines 153-160 in auth_otp.go:
// the ParseForm error path when the URL contains invalid percent-encoding.
func TestHandleAuthOTP_ParseFormError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	// Malformed URL with invalid percent-encoding triggers ParseForm error
	req := httptest.NewRequest("POST", "/auth/otp?user_id=%ZZ", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when ParseForm fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAuthOTP_ValidateOTPError covers the ValidateOTP error path (lines 196-208):
// when the OTP repository DB is broken, ValidateOTP fails.
func TestHandleAuthOTP_ValidateOTPError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	goodDB := testutil.SetupTestDB(t)
	badDB := badDBConn(t)

	userRepo := repository.NewUserRepository(goodDB, logger)
	otpRepoBad := repository.NewWebOTPRepository(badDB, logger)

	user, err := userRepo.GetOrCreateUser(55554)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(goodDB, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, goodDB, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepoBad)
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/auth/otp",
		strings.NewReader("user_id="+strconv.FormatInt(user.ID, 10)+"&code=123456"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthOTP(w, req)

	// ValidateOTP fails with broken DB → 401 (invalid OTP error)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when ValidateOTP fails, got %d: %s", w.Code, w.Body.String())
	}
}
