package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupAuthTelegramUnsafeTestDB(t *testing.T) (*sql.DB, *repository.UserRepository) {
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
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)

	return db, userRepo
}

func TestHandleAuthTelegramUnsafe_ValidUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthTelegramUnsafeTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret-key-for-telegram-unsafe-testing",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with valid user_id
	telegramID := int64(12345)
	req := httptest.NewRequest("POST", "/auth/telegram_unsafe", strings.NewReader("user_id="+strconv.FormatInt(telegramID, 10)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthTelegramUnsafe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response contains tokens
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["success"] != true {
		t.Error("Response should contain success=true")
	}
	if response["access_token"] == nil {
		t.Error("Response should contain access_token")
	}
	if response["refresh_token"] == nil {
		t.Error("Response should contain refresh_token")
	}
}

func TestHandleAuthTelegramUnsafe_MissingUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthTelegramUnsafeTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create request without user_id
	req := httptest.NewRequest("POST", "/auth/telegram_unsafe", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthTelegramUnsafe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAuthTelegramUnsafe_InvalidUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthTelegramUnsafeTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create request with invalid user_id (not a number)
	req := httptest.NewRequest("POST", "/auth/telegram_unsafe", strings.NewReader("user_id=not-a-number"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthTelegramUnsafe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAuthTelegramUnsafe_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthTelegramUnsafeTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create GET request (should fail)
	req := httptest.NewRequest("GET", "/auth/telegram_unsafe", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthTelegramUnsafe(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleAuthTelegramUnsafe_ExistingUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthTelegramUnsafeTestDB(t)
	defer db.Close()

	// Create a user first
	telegramID := int64(99999)
	_, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret-key-for-existing-user",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with existing user_id
	req := httptest.NewRequest("POST", "/auth/telegram_unsafe", strings.NewReader("user_id="+strconv.FormatInt(telegramID, 10)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthTelegramUnsafe(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response contains tokens
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["success"] != true {
		t.Error("Response should contain success=true")
	}
	if response["access_token"] == nil {
		t.Error("Response should contain access_token")
	}
}
