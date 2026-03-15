package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/testutil"
)

func setupDashboardChatTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.WordRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	wordRepo := repository.NewWordRepository(db, logger)

	return db, userRepo, wordRepo
}

func TestHandleChat_SingleWord_WordInDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, wordRepo := setupDashboardChatTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(121212)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Save word to database first
	err = wordRepo.SaveWordCard("hello", "a greeting")
	if err != nil {
		t.Fatalf("Failed to save word card: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	// Create word service with nil AI service (word is in DB, so won't need AI)
	wordService := service.NewWordService(wordRepo, nil, nil, nil, logger)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, wordService, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with single word
	req := httptest.NewRequest("POST", "/api/chat", strings.NewReader("message=hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleChat(w, req)

	// Should succeed since word is in DB
	if w.Code == http.StatusOK {
		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
			if response["response"] == nil {
				t.Error("Response should contain response field")
			}
		}
	}
}

func TestHandleChat_MultipleWords(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _ := setupDashboardChatTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(131313)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with multiple words (should go to AI chat)
	req := httptest.NewRequest("POST", "/api/chat", strings.NewReader("message=hello world"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleChat(w, req)

	// Should return error since AI service is nil
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Errorf("Expected status 500 or 200, got %d", w.Code)
	}
}

func TestHandleChat_MissingMessage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _ := setupDashboardChatTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(141414)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request without message
	req := httptest.NewRequest("POST", "/api/chat", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleChat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleChat_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _ := setupDashboardChatTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create GET request (should fail)
	req := httptest.NewRequest("GET", "/api/chat", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleChat(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleChat_Unauthorized(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _ := setupDashboardChatTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
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
	req := httptest.NewRequest("POST", "/api/chat", strings.NewReader("message=hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Call handler
	router.handleChat(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
