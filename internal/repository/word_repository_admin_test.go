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

func TestWordRepository_CountWordCardsAdmin_WithFilterUserID(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	defer db.Close()

	// Create word cards first
	repo.SaveWordCard("userword1", "definition 1")
	repo.SaveWordCard("userword2", "definition 2")

	// Get word cards to get their IDs
	card1, err := repo.GetWordCard("userword1")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}
	card2, err := repo.GetWordCard("userword2")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	// Add request history for user 100 with word_card_id
	word1 := "userword1"
	word2 := "userword2"
	repo.AddWordRequestHistoryWithCard(100, "userword1", &card1.ID, &word1)
	repo.AddWordRequestHistoryWithCard(100, "userword2", &card2.ID, &word2)

	// Count word cards for user 100
	userID := int64(100)
	count, err := repo.CountWordCardsAdmin(&userID, false, "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 cards for user, got %d", count)
	}
}

func TestWordRepository_CountWordCardsAdmin_WithErrors(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	defer db.Close()

	// Create word card with error
	repo.SaveWordCard("errorword", "definition")
	card, err := repo.GetWordCard("errorword")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}
	
	// Mark as processed with error
	err = repo.MarkWordCardProcessedError(card.ID, "test error")
	if err != nil {
		t.Fatalf("Failed to mark error: %v", err)
	}

	// Count word cards with errors
	count, err := repo.CountWordCardsAdmin(nil, true, "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 card with error, got %d", count)
	}
}

func TestWordRepository_CountWordCardsAdmin_WithSearch(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	defer db.Close()

	// Create word cards
	repo.SaveWordCard("searchable", "definition")
	repo.SaveWordCard("other", "definition")

	// Count word cards with search
	count, err := repo.CountWordCardsAdmin(nil, false, "search")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 card matching search, got %d", count)
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
