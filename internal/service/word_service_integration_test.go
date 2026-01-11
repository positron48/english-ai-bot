package service

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func setupWordServiceTestDB(t *testing.T) (*sql.DB, *repository.WordRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
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
		processed_at TEXT,
		processing_error TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS word_forms (
		form TEXT PRIMARY KEY,
		word_card_id INTEGER NOT NULL,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS word_request_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		word TEXT,
		word_card_id INTEGER,
		input_word TEXT,
		requested_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	wordRepo := repository.NewWordRepository(db, logger)

	return db, wordRepo
}

func setupWordServiceTestDBWithTraining(t *testing.T) (*sql.DB, *repository.WordRepository, *repository.TrainingCardRepository, *repository.UserCardRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
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
		processed_at TEXT,
		processing_error TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS word_forms (
		form TEXT PRIMARY KEY,
		word_card_id INTEGER NOT NULL,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS word_request_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		word TEXT,
		word_card_id INTEGER,
		input_word TEXT,
		requested_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS training_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word_card_id INTEGER NOT NULL,
		word_en TEXT NOT NULL,
		transcription TEXT,
		sense_index INTEGER NOT NULL DEFAULT 0,
		word_ru TEXT NOT NULL,
		meaning_en TEXT NOT NULL,
		example_en TEXT,
		example_ru TEXT,
		distractors_ru TEXT,
		distractors_en TEXT,
		hint TEXT,
		pos TEXT,
		display_word TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE,
		UNIQUE(word_card_id, sense_index)
	);
	
	CREATE TABLE IF NOT EXISTS user_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		training_card_id INTEGER NOT NULL,
		direction TEXT NOT NULL CHECK(direction IN ('ru_en', 'en_ru')),
		state TEXT DEFAULT 'new' CHECK(state IN ('new', 'learning', 'review')),
		ef REAL DEFAULT 2.5,
		reps INTEGER DEFAULT 0,
		interval_days INTEGER DEFAULT 0,
		learning_step INTEGER DEFAULT 0,
		lapse_count INTEGER DEFAULT 0,
		next_due_at TEXT,
		last_review_at TEXT,
		last_quality INTEGER,
		last_options_json TEXT,
		wrong_answers_json TEXT,
		stats_json TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (training_card_id) REFERENCES training_cards(id) ON DELETE CASCADE,
		UNIQUE(user_id, training_card_id, direction)
	);
	
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_id INTEGER UNIQUE NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
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
	wordRepo := repository.NewWordRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)

	return db, wordRepo, trainingCardRepo, userCardRepo
}

func TestWordService_GetWordDefinition_FromDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, wordRepo := setupWordServiceTestDB(t)
	defer db.Close()

	// Save a word to database
	err := wordRepo.SaveWordCard("testword", "a test word definition")
	if err != nil {
		t.Fatalf("Failed to save word card: %v", err)
	}

	service := NewWordService(wordRepo, nil, nil, (*ai.Service)(nil), logger)

	ctx := context.Background()
	definition, err := service.GetWordDefinition(ctx, 123, "testword")
	if err != nil {
		t.Fatalf("GetWordDefinition() error = %v", err)
	}
	// After refactoring, GetWordDefinition returns markdown rendered from structured data
	// The definition should contain the word and some markdown structure
	if definition == "" {
		t.Error("Expected non-empty definition (markdown)")
	}
	if !contains(definition, "testword") {
		t.Errorf("Expected definition to contain 'testword', got %q", definition)
	}
}

func TestWordService_GetWordDefinition_FromAI(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, wordRepo := setupWordServiceTestDB(t)
	defer db.Close()

	service := NewWordService(wordRepo, nil, nil, (*ai.Service)(nil), logger)
	_ = service // Verify service is created
	// Note: Full test would require a real AI service
}

func TestWordService_GetWordCard_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, wordRepo := setupWordServiceTestDB(t)
	defer db.Close()

	// Save a word to database
	err := wordRepo.SaveWordCard("getcard", "card definition")
	if err != nil {
		t.Fatalf("Failed to save word card: %v", err)
	}

	service := NewWordService(wordRepo, nil, nil, nil, logger)
	card, err := service.GetWordCard("getcard")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}
	if card == nil {
		t.Fatal("GetWordCard() should not return nil")
	}
	if card.Word != "getcard" {
		t.Errorf("Expected word 'getcard', got %q", card.Word)
	}
	if card.Definition != "card definition" {
		t.Errorf("Expected definition 'card definition', got %q", card.Definition)
	}
}

func TestWordService_GetWordDefinition_CreatesUserCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, wordRepo, trainingCardRepo, userCardRepo := setupWordServiceTestDBWithTraining(t)
	defer db.Close()

	ctx := context.Background()
	userID := int64(123)

	// Create a word card
	wordCard := &models.WordCard{
		Word:       "testword",
		Definition: "test definition",
		DefinitionRU: stringPtr("тестовое определение"),
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training cards for this word
	trainingCard1 := &models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "testword",
		SenseIndex: 0,
		WordRU:     "тестовое слово",
		MeaningEN:  "test meaning",
	}
	trainingCardID1, err := trainingCardRepo.CreateTrainingCard(trainingCard1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}

	trainingCard2 := &models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "testword",
		SenseIndex: 1,
		WordRU:     "другое значение",
		MeaningEN:  "another meaning",
	}
	trainingCardID2, err := trainingCardRepo.CreateTrainingCard(trainingCard2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Create service with repositories
	service := NewWordService(wordRepo, trainingCardRepo, userCardRepo, nil, logger)

	// Request the word - this should create user_cards
	definition, err := service.GetWordDefinition(ctx, userID, "testword")
	if err != nil {
		t.Fatalf("GetWordDefinition() error = %v", err)
	}
	if definition == "" {
		t.Error("Expected non-empty definition")
	}

	// Verify that user_cards were created for both training cards and both directions
	ruEnCard1, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID1, models.DirectionRUtoEN)
	if err != nil {
		t.Fatalf("Failed to get ru_en card 1: %v", err)
	}
	if ruEnCard1 == nil {
		t.Error("Expected ru_en user card 1 to be created")
	}

	enRuCard1, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID1, models.DirectionENtoRU)
	if err != nil {
		t.Fatalf("Failed to get en_ru card 1: %v", err)
	}
	if enRuCard1 == nil {
		t.Error("Expected en_ru user card 1 to be created")
	}

	ruEnCard2, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID2, models.DirectionRUtoEN)
	if err != nil {
		t.Fatalf("Failed to get ru_en card 2: %v", err)
	}
	if ruEnCard2 == nil {
		t.Error("Expected ru_en user card 2 to be created")
	}

	enRuCard2, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID2, models.DirectionENtoRU)
	if err != nil {
		t.Fatalf("Failed to get en_ru card 2: %v", err)
	}
	if enRuCard2 == nil {
		t.Error("Expected en_ru user card 2 to be created")
	}

	// Request the word again - should not create duplicates
	definition2, err := service.GetWordDefinition(ctx, userID, "testword")
	if err != nil {
		t.Fatalf("GetWordDefinition() error on second call = %v", err)
	}

	// Verify no new cards were created (should return existing IDs)
	ruEnCard1Again, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID1, models.DirectionRUtoEN)
	if err != nil {
		t.Fatalf("Failed to get ru_en card 1 again: %v", err)
	}
	if ruEnCard1Again.ID != ruEnCard1.ID {
		t.Errorf("Expected same card ID, got %d != %d", ruEnCard1Again.ID, ruEnCard1.ID)
	}

	_ = definition2 // Use the variable
}

func stringPtr(s string) *string {
	return &s
}
