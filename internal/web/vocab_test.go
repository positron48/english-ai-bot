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

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupVocabTestDB(t *testing.T) (*sql.DB, *repository.UserRepository) {
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
	);
	
	CREATE TABLE IF NOT EXISTS word_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT UNIQUE NOT NULL,
		definition TEXT NOT NULL,
		pos TEXT,
		transcription TEXT,
		definition_ru TEXT,
		examples_json TEXT,
		verb_forms_json TEXT,
		display_en TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS training_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word_card_id INTEGER NOT NULL,
		word_en TEXT NOT NULL,
		transcription TEXT,
		sense_index INTEGER NOT NULL,
		word_ru TEXT NOT NULL,
		meaning_en TEXT NOT NULL,
		example_en TEXT,
		example_ru TEXT,
		distractors_ru TEXT,
		distractors_en TEXT,
		hint TEXT,
		pos TEXT,
		display_word TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS user_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		training_card_id INTEGER NOT NULL,
		direction TEXT NOT NULL,
		state TEXT NOT NULL,
		ef REAL NOT NULL DEFAULT 2.5,
		reps INTEGER NOT NULL DEFAULT 0,
		interval_days INTEGER NOT NULL DEFAULT 0,
		learning_step INTEGER NOT NULL DEFAULT 0,
		lapse_count INTEGER NOT NULL DEFAULT 0,
		next_due_at TEXT,
		last_review_at TEXT,
		last_quality INTEGER,
		last_options_json TEXT,
		wrong_answers_json TEXT,
		stats_json TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS review_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER,
		user_id INTEGER NOT NULL,
		user_card_id INTEGER NOT NULL,
		direction TEXT NOT NULL,
		shown_at TEXT NOT NULL,
		options_shown_at TEXT,
		answered_at TEXT,
		t_delay_ms INTEGER,
		early_reveal INTEGER NOT NULL DEFAULT 0,
		option_count INTEGER NOT NULL,
		options_json TEXT,
		chosen_option TEXT,
		is_correct INTEGER NOT NULL DEFAULT 0,
		quality INTEGER NOT NULL,
		metrics_json TEXT,
		srs_before_json TEXT,
		srs_after_json TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS user_word_knowledge (
		user_id INTEGER NOT NULL,
		word_card_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'known' CHECK(status IN ('known')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE,
		UNIQUE(user_id, word_card_id)
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)

	return db, userRepo
}

func TestHandleVocab_Basic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with user context
	req := httptest.NewRequest("GET", "/app/vocab", nil)
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
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create POST request (should fail)
	req := httptest.NewRequest("POST", "/app/vocab", nil)
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
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request without user context
	req := httptest.NewRequest("GET", "/app/vocab", nil)
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
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with search parameter
	req := httptest.NewRequest("GET", "/app/vocab?search=test", nil)
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
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with pagination
	req := httptest.NewRequest("GET", "/app/vocab?page=2&limit=10", nil)
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

func TestHandleVocab_GroupByLemma(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupVocabTestDB(t)
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request
	req := httptest.NewRequest("GET", "/app/vocab", nil)
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
