package database

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name(), logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if db == nil {
		t.Fatal("New() should not return nil")
	}

	// Test GetConnection
	conn := db.GetConnection()
	if conn == nil {
		t.Fatal("GetConnection() should not return nil")
	}

	// Test Close
	err = db.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestNew_InvalidPath(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Try to create database with invalid path (should still work, SQLite creates file)
	// But we'll test with a path that might fail
	db, err := New("/invalid/path/test.db", logger)
	// SQLite might still succeed, so we just check it doesn't panic
	_ = db
	_ = err
}

func TestMigrateWordCardsLemmaColumns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name(), logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Create word_cards table without new columns
	_, err = db.GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS word_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			word TEXT NOT NULL,
			definition TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create word_cards table: %v", err)
	}

	// Run migration
	err = db.migrateWordCardsLemmaColumns()
	if err != nil {
		t.Fatalf("migrateWordCardsLemmaColumns() error = %v", err)
	}

	// Verify columns were added
	columns := []string{"pos", "transcription", "definition_ru", "examples_json", "verb_forms_json", "display_en"}
	for _, col := range columns {
		var exists int
		err = db.GetConnection().QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('word_cards') 
			WHERE name=?
		`, col).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check column %s: %v", col, err)
		}
		if exists == 0 {
			t.Errorf("Column %s was not added", col)
		}
	}
}

func TestMigrateTrainingCardsDisplayColumns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name(), logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Create training_cards table without new columns
	_, err = db.GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS training_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			word_card_id INTEGER NOT NULL,
			word_en TEXT NOT NULL,
			word_ru TEXT NOT NULL,
			meaning_en TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create training_cards table: %v", err)
	}

	// Run migration
	err = db.migrateTrainingCardsDisplayColumns()
	if err != nil {
		t.Fatalf("migrateTrainingCardsDisplayColumns() error = %v", err)
	}

	// Verify columns were added
	columns := []string{"pos", "display_word"}
	for _, col := range columns {
		var exists int
		err = db.GetConnection().QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('training_cards') 
			WHERE name=?
		`, col).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check column %s: %v", col, err)
		}
		if exists == 0 {
			t.Errorf("Column %s was not added", col)
		}
	}
}

func TestMigrateWordRequestHistory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	db, err := New(tmpFile.Name(), logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Create word_request_history table without new columns
	_, err = db.GetConnection().Exec(`
		CREATE TABLE IF NOT EXISTS word_request_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			word TEXT,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create word_request_history table: %v", err)
	}

	// Run migration
	err = db.migrateWordRequestHistory()
	if err != nil {
		t.Fatalf("migrateWordRequestHistory() error = %v", err)
	}

	// Verify columns were added
	columns := []string{"word_card_id", "input_word"}
	for _, col := range columns {
		var exists int
		err = db.GetConnection().QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('word_request_history') 
			WHERE name=?
		`, col).Scan(&exists)
		if err != nil {
			t.Fatalf("Failed to check column %s: %v", col, err)
		}
		if exists == 0 {
			t.Errorf("Column %s was not added", col)
		}
	}
}
