package web

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupAuthRefreshTestDB(t *testing.T) (*sql.DB, *repository.UserRepository) {
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

func TestHandleAuthRefresh_ValidToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthRefreshTestDB(t)
	defer db.Close()

	// Create a user
	user, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret-key-for-refresh-token-testing",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	// Generate refresh token
	_, refreshToken, err := authMiddleware.GenerateTokenPair(user.ID, user.TelegramID)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	// Create router
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create request
	body := map[string]interface{}{
		"refresh_token": refreshToken,
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthRefresh(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response contains new tokens
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["access_token"] == nil {
		t.Error("Response should contain access_token")
	}
	if response["refresh_token"] == nil {
		t.Error("Response should contain refresh_token")
	}
}

func TestHandleAuthRefresh_InvalidToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthRefreshTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create request with invalid token
	body := map[string]interface{}{
		"refresh_token": "invalid-token",
	}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthRefresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleAuthRefresh_MissingToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthRefreshTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create request without token
	body := map[string]interface{}{}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthRefresh(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAuthRefresh_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthRefreshTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create GET request (should fail)
	req := httptest.NewRequest("GET", "/auth/refresh", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleAuthRefresh(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
