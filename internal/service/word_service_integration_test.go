package service

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"tgbot-skeleton/internal/ai"
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

func TestWordService_GetWordDefinition_FromDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, wordRepo := setupWordServiceTestDB(t)
	defer db.Close()

	// Save a word to database
	err := wordRepo.SaveWordCard("testword", "a test word definition")
	if err != nil {
		t.Fatalf("Failed to save word card: %v", err)
	}

	service := NewWordService(wordRepo, (*ai.Service)(nil), logger)

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

	service := NewWordService(wordRepo, (*ai.Service)(nil), logger)
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

	service := NewWordService(wordRepo, nil, logger)
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
