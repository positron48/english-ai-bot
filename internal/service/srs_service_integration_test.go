package service

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupSRSServiceTestDB(t *testing.T) (*sql.DB, *repository.UserCardRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
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
	userCardRepo := repository.NewUserCardRepository(db, logger)

	return db, userCardRepo
}

func TestSRSService_GradeCard_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo := setupSRSServiceTestDB(t)
	defer db.Close()

	// Create a training card
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "grade", 0, "оценить", "to grade")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         111,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		Reps:           0,
		IntervalDays:   0,
		LearningStep:   0,
		LapseCount:     0,
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get the card
	userCard, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get user card: %v", err)
	}

	// Create SRS service
	service := NewSRSService(userCardRepo, logger)

	// Grade the card with correct answer
	attemptData := models.AttemptData{
		Correct:      true,
		EarlyReveal:  false,
		AnswerTimeMS: 3000, // Normal speed
	}

	err = service.GradeCard(userCard, attemptData)
	if err != nil {
		t.Fatalf("GradeCard() error = %v", err)
	}

	// Verify the card was updated
	updated, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get updated user card: %v", err)
	}
	if updated.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, updated.State)
	}
	if updated.LastQuality == nil {
		t.Error("LastQuality should be set after grading")
	}
}

func TestSRSService_GradeCard_WrongAnswer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo := setupSRSServiceTestDB(t)
	defer db.Close()

	// Create a training card
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "wrong", 0, "неправильно", "wrong")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card in review state
	now := time.Now()
	card := &models.UserCard{
		UserID:         222,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           5,
		IntervalDays:   10,
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get the card
	userCard, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get user card: %v", err)
	}

	// Create SRS service
	service := NewSRSService(userCardRepo, logger)

	// Grade the card with wrong answer
	attemptData := models.AttemptData{
		Correct: false,
	}

	err = service.GradeCard(userCard, attemptData)
	if err != nil {
		t.Fatalf("GradeCard() error = %v", err)
	}

	// Verify the card was reset to learning
	updated, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get updated user card: %v", err)
	}
	if updated.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, updated.State)
	}
	if updated.LapseCount == 0 {
		t.Error("LapseCount should be incremented after wrong answer")
	}
}
