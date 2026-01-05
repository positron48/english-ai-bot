package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupUserCardTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS training_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word_card_id INTEGER NOT NULL,
		word_en TEXT NOT NULL,
		sense_index INTEGER NOT NULL,
		word_ru TEXT NOT NULL,
		meaning_en TEXT NOT NULL,
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

	return db
}

func TestNewUserCardRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)
	_ = repo // Verify repository is created
}

func TestUserCardRepository_CreateUserCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)

	// Create a training card first
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "hello", 0, "привет", "greeting")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	now := time.Now()
	card := &models.UserCard{
		UserID:         123,
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

	id, err := repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("CreateUserCard() error = %v", err)
	}
	if id == 0 {
		t.Error("CreateUserCard() should return non-zero ID")
	}
}

func TestUserCardRepository_GetUserCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)

	// Create a training card first
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "world", 0, "мир", "earth")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         456,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateLearning,
		EF:             2.0,
		Reps:           1,
		IntervalDays:   3,
		LearningStep:   1,
		NextDueAt:      &now,
	}
	id, err := repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get the user card
	found, err := repo.GetUserCard(id)
	if err != nil {
		t.Fatalf("GetUserCard() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetUserCard() should not return nil")
	}
	if found.UserID != 456 {
		t.Errorf("Expected UserID 456, got %d", found.UserID)
	}
	if found.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, found.State)
	}
}

func TestUserCardRepository_GetDueCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)

	// Create a training card first
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "test", 0, "тест", "test")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user cards - one due, one not due
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	dueCard := &models.UserCard{
		UserID:         789,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = repo.CreateUserCard(dueCard)
	if err != nil {
		t.Fatalf("Failed to create due card: %v", err)
	}

	notDueCard := &models.UserCard{
		UserID:         789,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &future,
	}
	_, err = repo.CreateUserCard(notDueCard)
	if err != nil {
		t.Fatalf("Failed to create not due card: %v", err)
	}

	// Get due cards
	dueCards, err := repo.GetDueCards(789, now, 10)
	if err != nil {
		t.Fatalf("GetDueCards() error = %v", err)
	}
	if len(dueCards) == 0 {
		t.Error("GetDueCards() should return at least one card")
	}
}

func TestUserCardRepository_GetNewCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)

	// Create a training card first
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "new", 0, "новый", "new")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a new card
	newCard := &models.UserCard{
		UserID:         999,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	_, err = repo.CreateUserCard(newCard)
	if err != nil {
		t.Fatalf("Failed to create new card: %v", err)
	}

	// Get new cards
	newCards, err := repo.GetNewCards(999, 10)
	if err != nil {
		t.Fatalf("GetNewCards() error = %v", err)
	}
	if len(newCards) == 0 {
		t.Error("GetNewCards() should return at least one card")
	}
	if newCards[0].State != models.StateNew {
		t.Errorf("Expected State %v, got %v", models.StateNew, newCards[0].State)
	}
}

func TestUserCardRepository_UpdateUserCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)

	// Create a training card first
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "update", 0, "обновить", "update")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	card := &models.UserCard{
		UserID:         111,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	id, err := repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Update the card
	card.ID = id
	card.State = models.StateReview
	card.EF = 2.2
	card.Reps = 5
	card.IntervalDays = 10

	err = repo.UpdateUserCard(card)
	if err != nil {
		t.Fatalf("UpdateUserCard() error = %v", err)
	}

	// Verify update
	updated, err := repo.GetUserCard(id)
	if err != nil {
		t.Fatalf("GetUserCard() error = %v", err)
	}
	if updated.State != models.StateReview {
		t.Errorf("Expected State %v, got %v", models.StateReview, updated.State)
	}
	if updated.EF != 2.2 {
		t.Errorf("Expected EF 2.2, got %f", updated.EF)
	}
}

func TestUserCardRepository_GetDueCount(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)

	// Create a training card first
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "count", 0, "считать", "count")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create due cards
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	for i := 0; i < 3; i++ {
		card := &models.UserCard{
			UserID:         222,
			TrainingCardID: 1,
			Direction:      models.DirectionENtoRU,
			State:          models.StateReview,
			EF:             2.0,
			NextDueAt:      &past,
		}
		_, err = repo.CreateUserCard(card)
		if err != nil {
			t.Fatalf("Failed to create user card: %v", err)
		}
	}

	// Get due count
	count, err := repo.GetDueCount(222, now)
	if err != nil {
		t.Fatalf("GetDueCount() error = %v", err)
	}
	if count < 3 {
		t.Errorf("Expected at least 3 due cards, got %d", count)
	}
}
