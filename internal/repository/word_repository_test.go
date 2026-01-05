package repository

import (
	"database/sql"
	"testing"

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
		processed_at TEXT,
		processing_error TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS word_request_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		word TEXT NOT NULL,
		requested_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
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
