package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"tgbot-skeleton/internal/testutil"
	"go.uber.org/zap"
)

func setupDashboardTestDB(t *testing.T) (*sql.DB, *repository.UserRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)

	return db, userRepo
}

func TestHandleDashboard_Basic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupDashboardTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(11111)
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
	router.authMiddleware = authMiddleware

	// Create request with user context
	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response is JSON
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["due_count"] == nil {
		t.Error("Response should contain due_count")
	}
}

func TestHandleDashboard_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupDashboardTestDB(t)

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
	router.authMiddleware = authMiddleware

	// Create POST request (should fail)
	req := httptest.NewRequest("POST", "/api/dashboard", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleDashboard(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleDashboard_Unauthorized(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupDashboardTestDB(t)

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
	router.authMiddleware = authMiddleware

	// Create request without user context
	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleDashboard(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
