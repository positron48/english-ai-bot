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

func setupVocabTestDB(t *testing.T) (*sql.DB, *repository.UserRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)

	return db, userRepo
}

func TestHandleVocab_Basic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(22222)
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
	req := httptest.NewRequest("GET", "/api/vocab", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response is JSON
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	// words can be empty array, but should exist
	if _, ok := response["words"]; !ok {
		t.Error("Response should contain words field")
	}
	if response["pagination"] == nil {
		t.Error("Response should contain pagination")
	}
}

func TestHandleVocab_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)

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
	req := httptest.NewRequest("POST", "/api/vocab", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocab(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleVocab_Unauthorized(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)

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
	req := httptest.NewRequest("GET", "/api/vocab", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocab(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleVocab_WithSearch(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(33333)
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

	// Create request with search parameter
	req := httptest.NewRequest("GET", "/api/vocab?search=test", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleVocab_WithPagination(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(44444)
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

	// Create request with pagination
	req := httptest.NewRequest("GET", "/api/vocab?page=2&limit=10", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	pagination, ok := response["pagination"].(map[string]interface{})
	if !ok {
		t.Fatal("Response should contain pagination object")
	}
	if pagination["page"] != float64(2) {
		t.Errorf("Expected page 2, got %v", pagination["page"])
	}
	if pagination["limit"] != float64(10) {
		t.Errorf("Expected limit 10, got %v", pagination["limit"])
	}
}

func TestHandleVocab_WithMasteryLevel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(55556)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
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

	req := httptest.NewRequest("GET", "/api/vocab?mastery_level=learning", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestHandleVocab_InvalidMasteryLevelFilter uses an invalid mastery_level so no filter is applied.
func TestHandleVocab_InvalidMasteryLevelFilter(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(55558)
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

	req := httptest.NewRequest("GET", "/api/vocab?mastery_level=invalid_level", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 when mastery_level is invalid (no filter), got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocab_WithSortParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(55557)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
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

	req := httptest.NewRequest("GET", "/api/vocab?sort_by=last_review&sort_order=desc&limit=50", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleVocab_GroupByLemma(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(55555)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word_card (lemma)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "spy", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create two training_cards with same word_card_id but different word_en (spy and to spy)
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, display_word) VALUES (?, ?, ?, ?, ?, ?)",
		1, "spy", 0, "шпионить", "to spy", "spy")
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}

	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, display_word) VALUES (?, ?, ?, ?, ?, ?)",
		1, "spy", 1, "шпион", "spy", "to spy")
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Get training card IDs
	var trainingCardID1, trainingCardID2 int64
	err = db.QueryRow("SELECT id FROM training_cards WHERE word_en = ? AND sense_index = 0", "spy").Scan(&trainingCardID1)
	if err != nil {
		t.Fatalf("Failed to get training card ID 1: %v", err)
	}
	err = db.QueryRow("SELECT id FROM training_cards WHERE word_en = ? AND sense_index = 1", "spy").Scan(&trainingCardID2)
	if err != nil {
		t.Fatalf("Failed to get training card ID 2: %v", err)
	}

	// Create user_cards for both training cards
	_, err = db.Exec("INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, ?, ?, ?)",
		user.ID, trainingCardID1, "en_ru", "new", 2.5)
	if err != nil {
		t.Fatalf("Failed to create user card 1: %v", err)
	}

	_, err = db.Exec("INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, ?, ?, ?)",
		user.ID, trainingCardID2, "ru_en", "learning", 2.5)
	if err != nil {
		t.Fatalf("Failed to create user card 2: %v", err)
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

	// Create request
	req := httptest.NewRequest("GET", "/api/vocab", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocab(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	words, ok := response["words"].([]interface{})
	if !ok {
		t.Fatal("Response should contain words array")
	}

	// Should have only 1 word (grouped by word_card_id/lemma), not 2
	if len(words) != 1 {
		t.Errorf("Expected 1 word (grouped by lemma), got %d", len(words))
	}

	// Verify the word has correct fields
	word := words[0].(map[string]interface{})
	if word["word_card_id"] == nil {
		t.Error("Word should have word_card_id field")
	}
	if word["lemma"] == nil {
		t.Error("Word should have lemma field")
	}
	if word["display_word"] == nil {
		t.Error("Word should have display_word field")
	}
	if word["lemma"].(string) != "spy" {
		t.Errorf("Expected lemma 'spy', got %v", word["lemma"])
	}
	// Should have 2 total_cards (one for each training_card)
	if word["total_cards"].(float64) != 2 {
		t.Errorf("Expected 2 total_cards, got %v", word["total_cards"])
	}
}

// TestHandleVocab_MasteringScoreQuery verifies the vocab query returns stored mastering_score from user_word_mastering.
func TestHandleVocab_MasteringScoreQuery(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)

	user, err := userRepo.GetOrCreateUser(77777)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "apple", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "apple", 0, "яблоко", "fruit")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}
	var tcID int64
	err = db.QueryRow("SELECT id FROM training_cards WHERE word_card_id = 1 LIMIT 1").Scan(&tcID)
	if err != nil {
		t.Fatalf("Failed to get training card ID: %v", err)
	}
	_, err = db.Exec("INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, reps) VALUES (?, ?, ?, ?, ?, ?)",
		user.ID, tcID, "en_ru", "review", 2.5, 10)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}
	// Stored mastering score (vocab reads from user_word_mastering)
	_, err = db.Exec("INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score) VALUES (?, ?, ?)",
		user.ID, 1, 85)
	if err != nil {
		t.Fatalf("Failed to insert mastering score: %v", err)
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

	req := httptest.NewRequest("GET", "/api/vocab", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	words, _ := response["words"].([]interface{})
	if len(words) < 1 {
		t.Fatalf("Expected at least 1 word, got %d", len(words))
	}
	word := words[0].(map[string]interface{})
	if word["mastering_score"] == nil {
		t.Error("Word should have mastering_score")
		return
	}
	score := int(word["mastering_score"].(float64))
	if score != 85 {
		t.Errorf("mastering_score should be 85 (stored), got %d", score)
	}
	if word["mastery_level"] != "mastered" {
		t.Errorf("Expected mastery_level mastered, got %v", word["mastery_level"])
	}
}

func TestHandleVocab_DBQueryFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(88881)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/vocab", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when DB query fails, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocab_SortByMasteryLevel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(88885)
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

	req := httptest.NewRequest("GET", "/api/vocab?sort_by=mastery_level&sort_order=asc", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocab_SortByMasteryLevelDesc(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(88886)
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

	req := httptest.NewRequest("GET", "/api/vocab?sort_by=mastery_level_desc&sort_order=desc", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocab_SortByMasteringScore(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(88887)
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

	req := httptest.NewRequest("GET", "/api/vocab?sort_by=mastering_score&sort_order=desc", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocab_SortByLemma(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(88888)
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

	req := httptest.NewRequest("GET", "/api/vocab?sort_by=lemma&sort_order=asc", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocab_SortByDisplayWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(88889)
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

	req := httptest.NewRequest("GET", "/api/vocab?sort_by=display_word&sort_order=asc", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleVocab_SortByTotalCards hits the default orderByClause branch (sort_by not in switch).
func TestHandleVocab_SortByTotalCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(88891)
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

	req := httptest.NewRequest("GET", "/api/vocab?sort_by=total_cards&sort_order=asc", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVocab_LimitOver1000KeepsDefault(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	user, err := userRepo.GetOrCreateUser(88890)
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

	req := httptest.NewRequest("GET", "/api/vocab?limit=2000", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleVocab(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	pagination, _ := response["pagination"].(map[string]interface{})
	if pagination["limit"] != float64(25) {
		t.Errorf("Expected limit 25 when limit=2000 (cap), got %v", pagination["limit"])
	}
}
