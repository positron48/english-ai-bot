package web

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"tgbot-skeleton/internal/testutil"
	"go.uber.org/zap"
)

func setupAuthRefreshTestDB(t *testing.T) (*sql.DB, *repository.UserRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)

	return db, userRepo
}

func TestHandleAuthRefresh_ValidToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthRefreshTestDB(t)

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

// TestHandleAuthRefresh_FormFallback sends invalid JSON body; handler falls back to ParseForm.
// FormValue("refresh_token") is populated from URL query (ParseForm parses query + body).
func TestHandleAuthRefresh_FormFallback(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupAuthRefreshTestDB(t)
	user, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret-key-form-fallback",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	_, refreshToken, err := authMiddleware.GenerateTokenPair(user.ID, user.TelegramID)
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Invalid JSON body so Decode fails; refresh_token in query so ParseForm populates Form
	req := httptest.NewRequest("POST", "/auth/refresh?refresh_token="+url.QueryEscape(refreshToken), bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.handleAuthRefresh(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with form fallback, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["access_token"] == nil || resp["refresh_token"] == nil {
		t.Error("Expected access_token and refresh_token in response")
	}
}
