package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupWordTestDB(t *testing.T) *sql.DB {
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

	return db
}

func TestNewWordRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)
	if repo == nil {
		t.Error("NewWordRepository() should not return nil")
	}
}

func TestWordRepository_SaveWordCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	err := repo.SaveWordCard("hello", "a greeting")
	if err != nil {
		t.Fatalf("SaveWordCard() error = %v", err)
	}
}

func TestWordRepository_GetWordCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Save a word first
	err := repo.SaveWordCard("world", "the earth")
	if err != nil {
		t.Fatalf("Failed to save word: %v", err)
	}

	// Get the word
	card, err := repo.GetWordCard("world")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}
	if card == nil {
		t.Fatal("GetWordCard() should not return nil")
	}
	if card.Word != "world" {
		t.Errorf("Expected word 'world', got %q", card.Word)
	}
	if card.Definition != "the earth" {
		t.Errorf("Expected definition 'the earth', got %q", card.Definition)
	}
}

func TestWordRepository_AddWordRequestHistory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	err := repo.AddWordRequestHistory(123, "testword")
	if err != nil {
		t.Fatalf("AddWordRequestHistory() error = %v", err)
	}
}

func TestWordRepository_GetWordCard_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	card, err := repo.GetWordCard("nonexistent")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}
	if card != nil {
		t.Error("GetWordCard() should return nil for non-existent word")
	}
}

func TestWordRepository_SaveWordCard_UpdateExisting(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Save word first time
	err := repo.SaveWordCard("update", "first definition")
	if err != nil {
		t.Fatalf("Failed to save word: %v", err)
	}

	// Update with new definition
	err = repo.SaveWordCard("update", "updated definition")
	if err != nil {
		t.Fatalf("Failed to update word: %v", err)
	}

	// Verify update
	card, err := repo.GetWordCard("update")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}
	if card.Definition != "updated definition" {
		t.Errorf("Expected definition 'updated definition', got %q", card.Definition)
	}
}

func TestWordRepository_GetWordCardByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Save a word first
	err := repo.SaveWordCard("byid", "test definition")
	if err != nil {
		t.Fatalf("Failed to save word: %v", err)
	}

	// Get the word to get its ID
	card, err := repo.GetWordCard("byid")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}

	// Get by ID
	cardByID, err := repo.GetWordCardByID(card.ID)
	if err != nil {
		t.Fatalf("GetWordCardByID() error = %v", err)
	}
	if cardByID == nil {
		t.Fatal("GetWordCardByID() should not return nil")
	}
	if cardByID.Word != "byid" {
		t.Errorf("Expected word 'byid', got %q", cardByID.Word)
	}
}

func TestWordRepository_UpsertWordCardLemma(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	pos := "verb"
	transcription := "/spaɪ/"
	definitionRU := "шпионить"
	displayEN := "to spy"
	examplesJSON := `[{"example_en":"He likes to spy","gloss_ru":"Он любит шпионить"}]`
	verbFormsJSON := `{"v1":"spy","v2":"spied","v3":"spied"}`

	card := &models.WordCard{
		Word:          "spy",
		Definition:    "",
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU: &definitionRU,
		DisplayEN:     &displayEN,
		ExamplesJSON:  &examplesJSON,
		VerbFormsJSON: &verbFormsJSON,
	}

	id, err := repo.UpsertWordCardLemma(card)
	if err != nil {
		t.Fatalf("UpsertWordCardLemma() error = %v", err)
	}
	if id == 0 {
		t.Error("UpsertWordCardLemma() should return non-zero ID")
	}

	// Verify it was saved
	retrieved, err := repo.GetWordCardByID(id)
	if err != nil {
		t.Fatalf("GetWordCardByID() error = %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetWordCardByID() should not return nil")
	}
	if retrieved.Word != "spy" {
		t.Errorf("Expected word 'spy', got %q", retrieved.Word)
	}
	if retrieved.POS == nil || *retrieved.POS != "verb" {
		t.Errorf("Expected POS 'verb', got %v", retrieved.POS)
	}
}

func TestWordRepository_GetWordFormMapping(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Save a word first
	err := repo.SaveWordCard("spy", "to spy")
	if err != nil {
		t.Fatalf("Failed to save word: %v", err)
	}

	card, err := repo.GetWordCard("spy")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}

	// Create word form mapping
	err = repo.UpsertWordFormMapping("spied", card.ID)
	if err != nil {
		t.Fatalf("UpsertWordFormMapping() error = %v", err)
	}

	// Get the mapping
	wordForm, err := repo.GetWordFormMapping("spied")
	if err != nil {
		t.Fatalf("GetWordFormMapping() error = %v", err)
	}
	if wordForm == nil {
		t.Fatal("GetWordFormMapping() should not return nil")
	}
	if wordForm.Form != "spied" {
		t.Errorf("Expected form 'spied', got %q", wordForm.Form)
	}
	if wordForm.WordCardID != card.ID {
		t.Errorf("Expected word_card_id %d, got %d", card.ID, wordForm.WordCardID)
	}
}

func TestWordRepository_UpsertWordFormMapping(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Save a word first
	err := repo.SaveWordCard("run", "to run")
	if err != nil {
		t.Fatalf("Failed to save word: %v", err)
	}

	card, err := repo.GetWordCard("run")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}

	// Create word form mapping
	err = repo.UpsertWordFormMapping("ran", card.ID)
	if err != nil {
		t.Fatalf("UpsertWordFormMapping() error = %v", err)
	}

	// Verify it was created
	wordForm, err := repo.GetWordFormMapping("ran")
	if err != nil {
		t.Fatalf("GetWordFormMapping() error = %v", err)
	}
	if wordForm == nil {
		t.Fatal("GetWordFormMapping() should not return nil")
	}
	if wordForm.WordCardID != card.ID {
		t.Errorf("Expected word_card_id %d, got %d", card.ID, wordForm.WordCardID)
	}
}
