package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

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

// TestHandleDashboard_DBErrors covers branches where DB queries fail (counts set to 0, still 200).
// Uses a second connection that is closed so the shared DB is not affected.
func TestHandleDashboard_DBErrors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mainDB := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(mainDB, logger)
	user, err := userRepo.GetOrCreateUser(22222)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	conn2, err := sql.Open("postgres_compat", testutil.GetTestDSN(t))
	if err != nil {
		t.Skipf("second connection: %v (postgres_compat driver may not be registered)", err)
	}
	userRepo2 := repository.NewUserRepository(conn2, logger)
	_ = userRepo2 // same DB, user exists for conn2
	conn2.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:          "test-secret",
			JWTTTLHours:        24,
			RefreshTTLHours:     720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(mainDB, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, conn2, nil, nil, nil, nil)
	router.SetDependencies(userRepo2, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()

	router.handleDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 when DB errors (counts fallback to 0), got %d", w.Code)
	}
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["due_count"] != float64(0) {
		t.Errorf("Expected due_count 0 on DB error, got %v", response["due_count"])
	}
}

// TestHandleDashboard_WithGrammarStats covers grammarService != nil and grammar_stats in response.
func TestHandleDashboard_WithGrammarStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(33333)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	contentRepo := repository.NewGrammarContentRepository(logger)
	publishRepo := repository.NewGrammarPublishRepository(conn, logger)
	attemptRepo := repository.NewGrammarAttemptRepository(conn, logger)
	grammarService := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, logger)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:          "test-secret",
			JWTTTLHours:        24,
			RefreshTTLHours:     720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(conn, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware
	router.SetGrammarService(grammarService)

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()

	router.handleDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["grammar_stats"] == nil {
		t.Error("Expected grammar_stats in response when grammarService is set")
	}
}
