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

	"go.uber.org/zap"
	"tgbot-skeleton/internal/testutil"
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

func TestHandleVocabDelete_PostDelete_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	user, err := userRepo.GetOrCreateUser(77777)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 100, "todelete", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		100, "todelete", 0, "удалить", "to delete")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}
	var trainingCardID int64
	err = db.QueryRow("SELECT id FROM training_cards WHERE word_en = ?", "todelete").Scan(&trainingCardID)
	if err != nil {
		t.Fatalf("Failed to get training card ID: %v", err)
	}
	_, err = db.Exec("INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, ?, ?, ?)",
		user.ID, trainingCardID, "en_ru", "new", 2.5)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
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

	req := httptest.NewRequest(http.MethodPost, "/api/vocab/todelete/delete", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["success"] != true {
		t.Error("Expected success true")
	}
	if response["lemma"] != "todelete" {
		t.Errorf("Expected lemma 'todelete', got %v", response["lemma"])
	}
}

func TestHandleVocabDelete_InvalidAction(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	user, err := userRepo.GetOrCreateUser(55555)
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

	req := httptest.NewRequest("GET", "/api/vocab/hello/invalid_action", nil)
	req.URL.Path = "/api/vocab/hello/invalid_action"
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleVocabDelete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid action, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocabDelete_GetNoAction_InvalidRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	user, err := userRepo.GetOrCreateUser(66666)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, _ = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 50, "lemma1", "def")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		50, "lemma1", 0, "лемма", "lemma")

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

	// GET /api/vocab/lemma1 with no action falls through to Invalid request
	req := httptest.NewRequest("GET", "/api/vocab/lemma1", nil)
	req.URL.Path = "/api/vocab/lemma1"
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleVocabDelete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for GET with no valid action, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocabDelete_ConfirmDelete_WordNotInUserVocab(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	user, err := userRepo.GetOrCreateUser(44444)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Word exists in word_cards but user has no user_cards for it
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 200, "notinvocab", "def")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
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

	req := httptest.NewRequest("GET", "/api/vocab/notinvocab/confirm_delete", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleVocabDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 when word not in user vocab (count=0), got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocabDelete_EmptyLemmaRedirect(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)
	user, err := userRepo.GetOrCreateUser(88882)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/vocab//confirm_delete", nil)
	req.URL.Path = "/api/vocab//confirm_delete"
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status 302 redirect when lemma empty, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/app/vocab" {
		t.Errorf("Expected Location /app/vocab, got %s", loc)
	}
}

// TestHandleVocabDelete_MarkKnown_Success covers POST .../mark_known: marks word as known and removes user_cards.
func TestHandleVocabDelete_MarkKnown_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	user, err := userRepo.GetOrCreateUser(88883)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	wordCardID := int64(301)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", wordCardID, "markknown", "def")
	if err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		wordCardID, "markknown", 0, "отметить", "to mark")
	if err != nil {
		t.Fatalf("insert training_cards: %v", err)
	}
	var tcID int64
	if err := db.QueryRow("SELECT id FROM training_cards WHERE word_en = ?", "markknown").Scan(&tcID); err != nil {
		t.Fatalf("get training_card id: %v", err)
	}
	_, err = db.Exec("INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, ?, ?, ?)",
		user.ID, tcID, "en_ru", "new", 2.5)
	if err != nil {
		t.Fatalf("insert user_cards: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest(http.MethodPost, "/api/vocab/markknown/mark_known", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["success"] != true {
		t.Error("Expected success true")
	}
	if resp["lemma"] != "markknown" {
		t.Errorf("Expected lemma markknown, got %v", resp["lemma"])
	}
}

// TestHandleVocabDelete_MoveToTraining_Success covers POST .../move_to_training for a known word (no user_cards).
func TestHandleVocabDelete_MoveToTraining_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)

	user, err := userRepo.GetOrCreateUser(88884)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	wordCardID := int64(302)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", wordCardID, "movetraining", "def")
	if err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		wordCardID, "movetraining", 0, "перевести в тренировку", "to move to training")
	if err != nil {
		t.Fatalf("insert training_cards: %v", err)
	}
	_, err = db.Exec("INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES (?, ?, ?)", user.ID, wordCardID, "known")
	if err != nil {
		t.Fatalf("insert user_word_knowledge: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest(http.MethodPost, "/api/vocab/movetraining/move_to_training", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["success"] != true {
		t.Error("Expected success true")
	}
	if resp["lemma"] != "movetraining" {
		t.Errorf("Expected lemma movetraining, got %v", resp["lemma"])
	}
}

func TestHandleVocabDelete_DBGetWordCardIDFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabDeleteTestDB(t)
	user, err := userRepo.GetOrCreateUser(99998)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	badDB := badDBConn(t)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/vocab/someword/confirm_delete", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when get word card ID fails, got %d: %s", w.Code, w.Body.String())
	}
}
