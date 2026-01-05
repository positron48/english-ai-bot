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

	// Create POST request for delete
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

	// Should still return 200 (deletion succeeds even if no cards found)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
