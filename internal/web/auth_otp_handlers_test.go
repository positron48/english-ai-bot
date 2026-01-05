package web

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupAuthOTPHandlersTestDB(t *testing.T) (*sql.DB, *repository.WebOTPRepository, *repository.UserRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_id INTEGER UNIQUE NOT NULL,
		telegram_username TEXT,
		username TEXT,
		timezone TEXT DEFAULT '',
		preferred_training_time TEXT DEFAULT '',
		settings_json TEXT DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS web_otps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		code_hash TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		consumed_at TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	otpRepo := repository.NewWebOTPRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)

	return db, otpRepo, userRepo
}

func TestHandleAuthRequestOTP_UserNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, otpRepo, userRepo := setupAuthOTPHandlersTestDB(t)
	defer db.Close()

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
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

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
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

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
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

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
	defer db.Close()

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
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

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
