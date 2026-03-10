package web

import (
	"database/sql"
	"encoding/json"
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

func setupAuthOTPHandlersTestDB(t *testing.T) (*sql.DB, *repository.WebOTPRepository, *repository.UserRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	otpRepo := repository.NewWebOTPRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)

	return db, otpRepo, userRepo
}

func TestHandleAuthRequestOTP_UserNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
			OTPTTLSeconds: 300,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)

	// Create request for non-existent user
	req := httptest.NewRequest("POST", "/auth/request_otp", strings.NewReader("username=nonexistent"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}


func TestHandleAuthOTP_ValidCode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Generate OTP
	code, _, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP() error = %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret-key-for-otp-testing",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.authMiddleware = authMiddleware

	// Create request with valid OTP
	req := httptest.NewRequest("POST", "/auth/otp", strings.NewReader("user_id="+strconv.FormatInt(user.ID, 10)+"&code="+code))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthOTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleAuthOTP_InvalidCode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(88888)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.authMiddleware = authMiddleware

	// Create request with invalid OTP
	req := httptest.NewRequest("POST", "/auth/otp", strings.NewReader("user_id="+strconv.FormatInt(user.ID, 10)+"&code=000000"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthOTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleAuthOTP_MissingParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.authMiddleware = authMiddleware

	// Create request without code
	req := httptest.NewRequest("POST", "/auth/otp", strings.NewReader("user_id=123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAuthRequestOTP_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
			OTPTTLSeconds: 300,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)

	// Create GET request (should fail)
	req := httptest.NewRequest("GET", "/auth/request_otp", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleAuthOTP_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.authMiddleware = authMiddleware

	// Create GET request (should fail)
	req := httptest.NewRequest("GET", "/auth/otp", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthOTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleAuthRequestOTP_EmptyUsername(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
			OTPTTLSeconds:   300,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)

	req := httptest.NewRequest("POST", "/auth/request_otp", strings.NewReader("username="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAuthOTP_InvalidUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/auth/otp", strings.NewReader("user_id=notnumeric&code=123456"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestHandleAuthOTP_InvalidFormData verifies 400 when form parsing fails for OTP verify.
func TestHandleAuthOTP_InvalidFormData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/auth/otp", strings.NewReader("invalid multipart body"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary")
	w := httptest.NewRecorder()

	router.handleAuthOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid form data, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] == nil {
		t.Error("Expected error key in response")
	}
}

// TestHandleAuthRequestOTP_InvalidFormData verifies 400 when form parsing fails.
func TestHandleAuthRequestOTP_InvalidFormData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
			OTPTTLSeconds:   300,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)

	// Invalid multipart body without proper boundary/parts causes ParseForm to fail
	req := httptest.NewRequest("POST", "/auth/request_otp", strings.NewReader("invalid multipart body"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundary")
	w := httptest.NewRecorder()

	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid form data, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["error"] == nil {
		t.Error("Expected error key in response")
	}
}

// TestHandleAuthRequestOTP_BotNil verifies 503 when user exists and OTP is generated but bot is not set.
func TestHandleAuthRequestOTP_BotNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	user, err := userRepo.GetOrCreateUser(77777)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	// GetUserByUsernameOrID accepts telegram_id as string
	usernameOrID := strconv.FormatInt(user.TelegramID, 10)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
			OTPTTLSeconds:   300,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.bot = nil

	req := httptest.NewRequest("POST", "/auth/request_otp", strings.NewReader("username="+usernameOrID))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 (bot nil), got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAuthRequestOTP_GetUserError verifies 500 when GetUserByUsernameOrID fails (e.g. DB error).
func TestHandleAuthRequestOTP_GetUserError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	goodDB := testutil.SetupTestDB(t)
	badDB := badDBConn(t)
	userRepoBad := repository.NewUserRepository(badDB, logger)
	otpRepo := repository.NewWebOTPRepository(goodDB, logger)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
			OTPTTLSeconds:   300,
		},
	}
	router := NewRouter(logger, cfg, goodDB, nil, nil, nil, nil)
	router.SetDependencies(userRepoBad, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)

	req := httptest.NewRequest("POST", "/auth/request_otp", strings.NewReader("username=12345"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthRequestOTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when GetUserByUsernameOrID fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAuthOTP_GetUserByIDFails verifies 500 when OTP is valid but GetUserByID fails.
func TestHandleAuthOTP_GetUserByIDFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	goodDB := testutil.SetupTestDB(t)
	badDB := badDBConn(t)
	userRepoGood := repository.NewUserRepository(goodDB, logger)
	userRepoBad := repository.NewUserRepository(badDB, logger)
	otpRepo := repository.NewWebOTPRepository(goodDB, logger)

	user, err := userRepoGood.GetOrCreateUser(45454)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	code, _, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(goodDB, logger)
	authMiddleware := NewAuthMiddleware(userRepoBad, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, goodDB, nil, nil, nil, nil)
	router.SetDependencies(userRepoBad, nil, nil, nil, "test-token")
	router.SetOTPRepo(otpRepo)
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/auth/otp", strings.NewReader("user_id="+strconv.FormatInt(user.ID, 10)+"&code="+code))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.handleAuthOTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when GetUserByID fails after OTP valid, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAuthRequestOTP_Success verifies 200 when user exists, OTP is generated, and bot Send succeeds (mock API).
func TestHandleAuthRequestOTP_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)

	user, err := userRepo.GetOrCreateUser(88888)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	usernameOrID := strconv.FormatInt(user.TelegramID, 10)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer mockServer.Close()

	// API endpoint format: base URL with placeholders for token and method (e.g. getMe, sendMessage)
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

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["success"] != true {
		t.Errorf("Expected success=true, got %v", resp["success"])
	}
}

