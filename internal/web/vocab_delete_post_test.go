package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupVocabDeletePostTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.TrainingCardRepository, *repository.UserCardRepository) {
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
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)

	return db, userRepo, trainingCardRepo, userCardRepo
}

func TestHandleVocabDelete_Delete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo := setupVocabDeletePostTestDB(t)
	defer db.Close()

	// Create a user
	user, err := userRepo.GetOrCreateUser(414141)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word_card (lemma)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "deletepost", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	trainingCard := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "deletepost",
		SenseIndex: 0,
		WordRU:     "удалить пост",
		MeaningEN:  "delete post",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user card
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	_, err = userCardRepo.CreateUserCard(userCard)
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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create POST request for delete using lemma
	req := httptest.NewRequest("POST", "/app/vocab/deletepost/delete", nil)
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
	if response["success"] == nil {
		t.Error("Response should contain success field")
	}
	if response["lemma"] == nil {
		t.Error("Response should contain lemma field")
	}
	if response["lemma"].(string) != "deletepost" {
		t.Errorf("Expected lemma 'deletepost', got %v", response["lemma"])
	}
}

func TestHandleVocabDelete_Delete_NoCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _ := setupVocabDeletePostTestDB(t)
	defer db.Close()

	// Create a user
	user, err := userRepo.GetOrCreateUser(424242)
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

	// Create POST request for delete (word doesn't exist)
	req := httptest.NewRequest("POST", "/app/vocab/nonexistent/delete", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleVocabDelete(w, req)

	// Should return 404 (word not found)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleVocabDelete_DeleteByWordCardID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo := setupVocabDeletePostTestDB(t)
	defer db.Close()

	// Create a user
	user, err := userRepo.GetOrCreateUser(434343)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word_card (lemma)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "testword", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create two training cards with same word_card_id
	trainingCard1 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "testword",
		SenseIndex: 0,
		WordRU:     "тестовое слово",
		MeaningEN:  "test word",
	}
	trainingCardID1, err := trainingCardRepo.CreateTrainingCard(trainingCard1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}

	trainingCard2 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "testword",
		SenseIndex: 1,
		WordRU:     "тестовое слово 2",
		MeaningEN:  "test word 2",
	}
	trainingCardID2, err := trainingCardRepo.CreateTrainingCard(trainingCard2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Create user cards for both training cards
	userCard1 := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	_, err = userCardRepo.CreateUserCard(userCard1)
	if err != nil {
		t.Fatalf("Failed to create user card 1: %v", err)
	}

	userCard2 := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID2,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateLearning,
		EF:             models.InitialEF,
	}
	_, err = userCardRepo.CreateUserCard(userCard2)
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

	// Create POST request for delete using lemma
	req := httptest.NewRequest("POST", "/app/vocab/testword/delete", nil)
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
	if response["success"] == nil || !response["success"].(bool) {
		t.Error("Response should contain success=true")
	}
	if response["rows_affected"] == nil {
		t.Error("Response should contain rows_affected field")
	}
	rowsAffected := int64(response["rows_affected"].(float64))
	if rowsAffected != 2 {
		t.Errorf("Expected 2 rows affected (both user_cards deleted), got %d", rowsAffected)
	}

	// Verify both user_cards are deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user_cards WHERE user_id = ? AND (training_card_id = ? OR training_card_id = ?)",
		user.ID, trainingCardID1, trainingCardID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check user cards: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 user_cards remaining, got %d", count)
	}
}
