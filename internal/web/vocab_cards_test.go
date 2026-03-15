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

func setupVocabCardsTestDB(t *testing.T) (*sql.DB, *repository.UserRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)

	return db, userRepo
}

func TestHandleVocabDelete_Cards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabCardsTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(353535)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word_card (lemma)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "cards", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "cards", 0, "карточки", "cards")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Get training card ID
	var trainingCardID int64
	err = db.QueryRow("SELECT id FROM training_cards WHERE word_en = ?", "cards").Scan(&trainingCardID)
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

	// Create request for cards using lemma
	req := httptest.NewRequest("GET", "/api/vocab/cards/cards", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler (path .../cards delegates to handleVocabWordCards)
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["cards"] == nil {
		t.Error("Response should contain cards")
	}
	if response["lemma"] == nil {
		t.Error("Response should contain lemma field")
	}
	if response["lemma"].(string) != "cards" {
		t.Errorf("Expected lemma 'cards', got %v", response["lemma"])
	}
}

func TestHandleVocabDelete_InvalidRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabCardsTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(363636)
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

	// Create request with invalid action
	req := httptest.NewRequest("GET", "/api/vocab/testword/invalid", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleVocabWordCards_WordNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabCardsTestDB(t)

	user, err := userRepo.GetOrCreateUser(454545)
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

	req := httptest.NewRequest("GET", "/api/vocab/nonexistentlemma123/cards", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleVocabDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleVocabWordCards_KnownWordWithoutUserCards covers the branch where the word is marked "known"
// but has no user_cards; handler returns training_cards as card details for both directions.
func TestHandleVocabWordCards_KnownWordWithoutUserCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabCardsTestDB(t)

	user, err := userRepo.GetOrCreateUser(464646)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 2, "knownonly", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		2, "knownonly", 0, "только известное", "known only")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}
	_, err = db.Exec("INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES (?, ?, ?)", user.ID, 2, "known")
	if err != nil {
		t.Fatalf("Failed to create user_word_knowledge: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/vocab/knownonly/cards", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if response["cards"] == nil {
		t.Error("Response should contain cards")
	}
	cards, _ := response["cards"].([]interface{})
	if len(cards) == 0 {
		t.Error("Expected at least one card for known word without user_cards")
	}
	if response["is_known"] != true {
		t.Error("Expected is_known true")
	}
}

// TestHandleVocabWordCards_NoUserCardsNotKnown covers word in word_cards but no user_cards and not in user_word_knowledge -> 404.
func TestHandleVocabWordCards_NoUserCardsNotKnown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabCardsTestDB(t)

	user, err := userRepo.GetOrCreateUser(474747)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 10, "nocards", "definition")
	if err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		10, "nocards", 0, "без карточек", "no cards")
	if err != nil {
		t.Fatalf("insert training_cards: %v", err)
	}
	// No user_cards, no user_word_knowledge for this user

	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/vocab/nocards/cards", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 when word has no user_cards and not known, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleVocabWordCards_ResponseHasVerbFormsAndPos covers response with verb_forms and pos when word_cards has pos=verb and verb_forms_json.
func TestHandleVocabWordCards_ResponseHasVerbFormsAndPos(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabCardsTestDB(t)

	user, err := userRepo.GetOrCreateUser(484848)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	verbFormsJSON := `{"past":"went","past_participle":"gone"}`
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition, pos, verb_forms_json) VALUES (?, ?, ?, ?, ?)",
		11, "go", "definition", "verb", verbFormsJSON)
	if err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		11, "go", 0, "идти", "to go")
	if err != nil {
		t.Fatalf("insert training_cards: %v", err)
	}
	var tcID int64
	if err := db.QueryRow("SELECT id FROM training_cards WHERE word_card_id = 11 LIMIT 1").Scan(&tcID); err != nil {
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

	req := httptest.NewRequest("GET", "/api/vocab/go/cards", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if response["verb_forms"] == nil {
		t.Error("Expected verb_forms in response when word has verb_forms_json")
	}
	if response["pos"] != "verb" {
		t.Errorf("Expected pos=verb, got %v", response["pos"])
	}
}

// TestHandleVocabWordCards_InvalidVerbFormsJSON covers the branch where verb_forms_json fails to parse (Warn).
func TestHandleVocabWordCards_InvalidVerbFormsJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabCardsTestDB(t)

	user, err := userRepo.GetOrCreateUser(505050)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	_, err = db.Exec("INSERT INTO word_cards (id, word, definition, pos, verb_forms_json) VALUES (?, ?, ?, ?, ?)",
		12, "badverb", "definition", "verb", "invalid json {{{")
	if err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		12, "badverb", 0, "глагол", "bad verb")
	if err != nil {
		t.Fatalf("insert training_cards: %v", err)
	}
	var tcID int64
	if err := db.QueryRow("SELECT id FROM training_cards WHERE word_card_id = 12 LIMIT 1").Scan(&tcID); err != nil {
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

	req := httptest.NewRequest("GET", "/api/vocab/badverb/cards", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (invalid verb_forms is non-fatal), got %d: %s", w.Code, w.Body.String())
	}
	// Response should still have cards and pos; verb_forms may be absent due to parse error
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if response["cards"] == nil {
		t.Error("Expected cards in response")
	}
}

func TestHandleVocabWordCards_DBGetWordCardIDFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabCardsTestDB(t)
	user, err := userRepo.GetOrCreateUser(494949)
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

	req := httptest.NewRequest("GET", "/api/vocab/anyword/cards", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocabDelete(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when get word card ID fails in handleVocabWordCards, got %d: %s", w.Code, w.Body.String())
	}
}
