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

func setupVocabDeleteTestDB(t *testing.T) (*sql.DB, *repository.UserRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)

	return db, userRepo
}

func TestHandleVocabDelete_ConfirmDelete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(88888)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "deleteword", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "deleteword", 0, "удалить", "to delete")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Get training card ID
	var trainingCardID int64
	err = db.QueryRow("SELECT id FROM training_cards WHERE word_en = ?", "deleteword").Scan(&trainingCardID)
	if err != nil {
		t.Fatalf("Failed to get training card ID: %v", err)
	}

	// Create user card
	_, err = db.Exec("INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, ?, ?, ?)",
		user.ID, trainingCardID, "en_ru", "new", 2.5)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
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

	// Create request for confirm_delete
	req := httptest.NewRequest("GET", "/api/vocab/deleteword/confirm_delete", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	// Response should contain lemma or error
	if response["lemma"] == nil && response["error"] == nil {
		t.Error("Response should contain lemma or error")
	}
	if response["lemma"] != nil && response["lemma"].(string) != "deleteword" {
		t.Errorf("Expected lemma 'deleteword', got %v", response["lemma"])
	}
}

func TestHandleVocabDelete_WordNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(99999)
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

	// Create request for non-existent word
	req := httptest.NewRequest("GET", "/api/vocab/nonexistent/confirm_delete", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleVocabDelete_Unauthorized(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

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
	req := httptest.NewRequest("GET", "/api/vocab/testword/confirm_delete", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleVocabDelete_InvalidPath(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(101010)
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

	// Create request with empty word
	req := httptest.NewRequest("GET", "/api/vocab/", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	// Should redirect to vocab list
	if w.Code != http.StatusFound {
		t.Errorf("Expected status 302 (redirect), got %d", w.Code)
	}
}
