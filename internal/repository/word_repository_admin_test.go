package repository

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupWordAdminTestDB(t *testing.T) (*sql.DB, *WordRepository) {
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
	);
	
	CREATE TABLE IF NOT EXISTS training_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word_card_id INTEGER NOT NULL,
		word_en TEXT NOT NULL,
		sense_index INTEGER NOT NULL,
		word_ru TEXT NOT NULL,
		meaning_en TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	wordRepo := NewWordRepository(db, logger)

	return db, wordRepo
}

func TestWordRepository_ListWordCardsAdmin(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	defer db.Close()

	// Create word cards
	repo.SaveWordCard("admin1", "definition 1")
	repo.SaveWordCard("admin2", "definition 2")

	// List word cards
	cards, err := repo.ListWordCardsAdmin(nil, false, "", 10, 0)
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	if len(cards) < 2 {
		t.Errorf("Expected at least 2 cards, got %d", len(cards))
	}
}

func TestWordRepository_ListWordCardsAdmin_WithSearch(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	defer db.Close()

	// Create word cards
	repo.SaveWordCard("searchword", "definition")
	repo.SaveWordCard("otherword", "definition")

	// List word cards with search
	cards, err := repo.ListWordCardsAdmin(nil, false, "search", 10, 0)
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	if len(cards) == 0 {
		t.Error("Expected at least one card matching search")
	}
}

func TestWordRepository_CountWordCardsAdmin(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	defer db.Close()

	// Create word cards
	repo.SaveWordCard("count1", "definition 1")
	repo.SaveWordCard("count2", "definition 2")
	repo.SaveWordCard("count3", "definition 3")

	// Count word cards
	count, err := repo.CountWordCardsAdmin(nil, false, "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 3 {
		t.Errorf("Expected at least 3 cards, got %d", count)
	}
}

func TestWordRepository_GetWordCardRequestingUsers(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	defer db.Close()

	// Create a word card
	repo.SaveWordCard("requesting", "definition")

	// Get word card to get its ID
	card, err := repo.GetWordCard("requesting")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	// Add request history
	repo.AddWordRequestHistory(100, "requesting")
	repo.AddWordRequestHistory(200, "requesting")

	// Get requesting users
	users, err := repo.GetWordCardRequestingUsers(card.ID)
	if err != nil {
		t.Fatalf("GetWordCardRequestingUsers() error = %v", err)
	}
	if len(users) < 2 {
		t.Errorf("Expected at least 2 users, got %d", len(users))
	}
}

func TestWordRepository_DeleteWordCard(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	defer db.Close()

	// Create a word card
	repo.SaveWordCard("deletecard", "definition")

	// Get word card to get its ID
	card, err := repo.GetWordCard("deletecard")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	// Delete word card
	err = repo.DeleteWordCard(card.ID)
	if err != nil {
		t.Fatalf("DeleteWordCard() error = %v", err)
	}

	// Verify deletion
	deleted, err := repo.GetWordCard("deletecard")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}
	if deleted != nil {
		t.Error("GetWordCard() should return nil for deleted card")
	}
}
