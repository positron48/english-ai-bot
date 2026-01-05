package service

import (
	"database/sql"
	"encoding/json"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupOptionsServiceTestDB(t *testing.T) (*sql.DB, *repository.TrainingCardRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS word_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT UNIQUE NOT NULL,
		definition TEXT NOT NULL,
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
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)

	return db, trainingCardRepo
}

func TestOptionsService_GenerateOptions_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)
	defer db.Close()

	// Create a word card
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "generate", "to generate")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training cards for the same word
	distractorsEN, _ := json.Marshal([]string{"create", "make", "build"})
	card1 := &models.TrainingCard{
		WordCardID:    1,
		WordEN:        "generate",
		SenseIndex:    0,
		WordRU:        "генерировать",
		MeaningEN:     "to generate",
		DistractorsEN: string(distractorsEN),
	}
	card1ID, err := trainingCardRepo.CreateTrainingCard(card1)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger)

	userCard := &models.UserCardWithTraining{
		UserCard: models.UserCard{
			ID:              1,
			Direction:       models.DirectionRUtoEN,
			WrongAnswersJSON: "",
		},
		TrainingCard: models.TrainingCard{
			ID:            card1ID,
			WordCardID:    1,
			WordEN:        "generate",
			WordRU:        "генерировать",
			DistractorsEN: string(distractorsEN),
		},
	}

	options, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{})
	if err != nil {
		t.Fatalf("GenerateOptions() error = %v", err)
	}
	if len(options) < 2 {
		t.Errorf("Expected at least 2 options, got %d", len(options))
	}
	if correctAnswer != "generate" {
		t.Errorf("Expected correct answer 'generate', got %q", correctAnswer)
	}
	// Verify correct answer is in options
	found := false
	for _, opt := range options {
		if opt == correctAnswer {
			found = true
			break
		}
	}
	if !found {
		t.Error("Correct answer should be in options")
	}
}

func TestOptionsService_GenerateOptions_UsesCardDistractors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)
	defer db.Close()

	// Create a word card
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "spy", "to spy")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card with specific distractors
	cardDistractors := []string{"watch", "observe", "monitor"}
	distractorsEN, _ := json.Marshal(cardDistractors)
	card1 := &models.TrainingCard{
		WordCardID:    1,
		WordEN:        "spy",
		SenseIndex:    0,
		WordRU:        "шпионить",
		MeaningEN:     "to spy",
		DistractorsEN: string(distractorsEN),
	}
	card1ID, err := trainingCardRepo.CreateTrainingCard(card1)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger)

	userCard := &models.UserCardWithTraining{
		UserCard: models.UserCard{
			ID:              1,
			Direction:       models.DirectionRUtoEN,
			WrongAnswersJSON: "",
		},
		TrainingCard: models.TrainingCard{
			ID:            card1ID,
			WordCardID:    1,
			WordEN:        "spy",
			WordRU:        "шпионить",
			DistractorsEN: string(distractorsEN),
		},
	}

	// Generate options without session words to ensure card distractors are used
	options, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{})
	if err != nil {
		t.Fatalf("GenerateOptions() error = %v", err)
	}
	if len(options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(options))
	}
	if correctAnswer != "spy" {
		t.Errorf("Expected correct answer 'spy', got %q", correctAnswer)
	}

	// Verify that at least one card distractor is in the options
	cardDistractorsSet := make(map[string]bool)
	for _, d := range cardDistractors {
		cardDistractorsSet[d] = true
	}

	foundCardDistractor := false
	for _, opt := range options {
		if opt != correctAnswer && cardDistractorsSet[opt] {
			foundCardDistractor = true
			break
		}
	}

	if !foundCardDistractor {
		t.Errorf("Expected at least one card distractor (%v) in options, got: %v", cardDistractors, options)
	}

	// For 4 options, we should have at least 1 card distractor
	// Let's count how many card distractors are used
	cardDistractorsUsed := 0
	for _, opt := range options {
		if opt != correctAnswer && cardDistractorsSet[opt] {
			cardDistractorsUsed++
		}
	}
	if cardDistractorsUsed < 1 {
		t.Errorf("Expected at least 1 card distractor to be used, found %d", cardDistractorsUsed)
	}
}
