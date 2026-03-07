package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupUserWordKnowledgeTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewUserWordKnowledgeRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserWordKnowledgeTestDB(t)

	repo := NewUserWordKnowledgeRepository(db, logger)
	if repo == nil {
		t.Error("NewUserWordKnowledgeRepository() should not return nil")
	}
}

func TestUserWordKnowledgeRepository_MarkKnown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserWordKnowledgeTestDB(t)

	repo := NewUserWordKnowledgeRepository(db, logger)

	// Create a user and word card first
	userRepo := NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	wordRepo := NewWordRepository(db, logger)
	wordCard := &models.WordCard{
		Word:       "test",
		Definition: "test definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	t.Run("Mark word as known", func(t *testing.T) {
		err := repo.MarkKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("MarkKnown() error = %v", err)
		}

		// Verify it's marked as known
		isKnown, err := repo.IsKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("IsKnown() error = %v", err)
		}
		if !isKnown {
			t.Error("Word should be marked as known")
		}
	})

	t.Run("Mark same word as known again (should replace)", func(t *testing.T) {
		err := repo.MarkKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("MarkKnown() error = %v", err)
		}

		isKnown, err := repo.IsKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("IsKnown() error = %v", err)
		}
		if !isKnown {
			t.Error("Word should still be marked as known")
		}
	})
}

func TestUserWordKnowledgeRepository_RemoveKnown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserWordKnowledgeTestDB(t)

	repo := NewUserWordKnowledgeRepository(db, logger)

	// Create a user and word card first
	userRepo := NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(12346)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	wordRepo := NewWordRepository(db, logger)
	wordCard := &models.WordCard{
		Word:       "remove",
		Definition: "remove definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Mark as known first
	err = repo.MarkKnown(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("Failed to mark as known: %v", err)
	}

	t.Run("Remove known status", func(t *testing.T) {
		err := repo.RemoveKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("RemoveKnown() error = %v", err)
		}

		// Verify it's no longer known
		isKnown, err := repo.IsKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("IsKnown() error = %v", err)
		}
		if isKnown {
			t.Error("Word should not be marked as known")
		}
	})

	t.Run("Remove non-existent known status", func(t *testing.T) {
		// Should not error
		err := repo.RemoveKnown(user.ID, 99999)
		if err != nil {
			t.Fatalf("RemoveKnown() should not error for non-existent entry: %v", err)
		}
	})
}

func TestUserWordKnowledgeRepository_IsKnown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserWordKnowledgeTestDB(t)

	repo := NewUserWordKnowledgeRepository(db, logger)

	// Create a user and word card first
	userRepo := NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(12347)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	wordRepo := NewWordRepository(db, logger)
	wordCard := &models.WordCard{
		Word:       "check",
		Definition: "check definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	t.Run("Check unknown word", func(t *testing.T) {
		isKnown, err := repo.IsKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("IsKnown() error = %v", err)
		}
		if isKnown {
			t.Error("Word should not be known")
		}
	})

	t.Run("Check known word", func(t *testing.T) {
		err := repo.MarkKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("Failed to mark as known: %v", err)
		}

		isKnown, err := repo.IsKnown(user.ID, wordCardID)
		if err != nil {
			t.Fatalf("IsKnown() error = %v", err)
		}
		if !isKnown {
			t.Error("Word should be known")
		}
	})
}

func TestUserWordKnowledgeRepository_GetKnownWords(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserWordKnowledgeTestDB(t)

	repo := NewUserWordKnowledgeRepository(db, logger)

	// Create a user and word cards first
	userRepo := NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(12348)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	wordRepo := NewWordRepository(db, logger)
	wordCard1 := &models.WordCard{
		Word:       "word1",
		Definition: "definition1",
	}
	wordCardID1, err := wordRepo.UpsertWordCardLemma(wordCard1)
	if err != nil {
		t.Fatalf("Failed to create word card 1: %v", err)
	}

	wordCard2 := &models.WordCard{
		Word:       "word2",
		Definition: "definition2",
	}
	wordCardID2, err := wordRepo.UpsertWordCardLemma(wordCard2)
	if err != nil {
		t.Fatalf("Failed to create word card 2: %v", err)
	}

	t.Run("Get known words for user with no known words", func(t *testing.T) {
		words, err := repo.GetKnownWords(user.ID)
		if err != nil {
			t.Fatalf("GetKnownWords() error = %v", err)
		}
		if len(words) != 0 {
			t.Errorf("Expected 0 known words, got %d", len(words))
		}
	})

	t.Run("Get known words for user with known words", func(t *testing.T) {
		err := repo.MarkKnown(user.ID, wordCardID1)
		if err != nil {
			t.Fatalf("Failed to mark word 1 as known: %v", err)
		}

		err = repo.MarkKnown(user.ID, wordCardID2)
		if err != nil {
			t.Fatalf("Failed to mark word 2 as known: %v", err)
		}

		words, err := repo.GetKnownWords(user.ID)
		if err != nil {
			t.Fatalf("GetKnownWords() error = %v", err)
		}
		if len(words) != 2 {
			t.Errorf("Expected 2 known words, got %d", len(words))
		}

		// Check that both word IDs are present
		found1, found2 := false, false
		for _, id := range words {
			if id == wordCardID1 {
				found1 = true
			}
			if id == wordCardID2 {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Error("Expected word IDs not found in result")
		}
	})
}

// TestUserWordKnowledgeRepository_MarkKnown_Error covers the error path when Exec fails.
func TestUserWordKnowledgeRepository_MarkKnown_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserWordKnowledgeRepository(db, zap.NewNop())
	err = repo.MarkKnown(1, 1)
	if err == nil {
		t.Error("MarkKnown() expected error with invalid DSN")
	}
}

// TestUserWordKnowledgeRepository_RemoveKnown_Error covers the error path when Exec fails.
func TestUserWordKnowledgeRepository_RemoveKnown_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserWordKnowledgeRepository(db, zap.NewNop())
	err = repo.RemoveKnown(1, 1)
	if err == nil {
		t.Error("RemoveKnown() expected error with invalid DSN")
	}
}

// TestUserWordKnowledgeRepository_IsKnown_Error covers the error path when QueryRow/Scan fails.
func TestUserWordKnowledgeRepository_IsKnown_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserWordKnowledgeRepository(db, zap.NewNop())
	_, err = repo.IsKnown(1, 1)
	if err == nil {
		t.Error("IsKnown() expected error with invalid DSN")
	}
}

// TestUserWordKnowledgeRepository_GetKnownWords_Error covers the error path when Query fails.
func TestUserWordKnowledgeRepository_GetKnownWords_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewUserWordKnowledgeRepository(db, zap.NewNop())
	_, err = repo.GetKnownWords(1)
	if err == nil {
		t.Error("GetKnownWords() expected error with invalid DSN")
	}
}

// TestUserWordKnowledgeRepository_GetKnownWords_ScanError covers the branch where rows.Scan fails (logger.Warn + continue).
func TestUserWordKnowledgeRepository_GetKnownWords_ScanError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserWordKnowledgeTestDB(t)

	userRepo := NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(12349)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	wordRepo := NewWordRepository(db, logger)
	wordCard := &models.WordCard{Word: "scan", Definition: "scan def"}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("create word card: %v", err)
	}

	repo := NewUserWordKnowledgeRepository(db, logger)
	if err := repo.MarkKnown(user.ID, wordCardID); err != nil {
		t.Fatalf("mark known: %v", err)
	}

	// Allow NULL on word_card_id and insert a row that will fail Scan so the "continue" branch is covered.
	_, _ = db.Exec("ALTER TABLE user_word_knowledge ALTER COLUMN word_card_id DROP NOT NULL")
	defer func() {
		_, _ = db.Exec("DELETE FROM user_word_knowledge WHERE user_id = $1 AND word_card_id IS NULL", user.ID)
		_, _ = db.Exec("ALTER TABLE user_word_knowledge ALTER COLUMN word_card_id SET NOT NULL")
	}()

	_, err = db.Exec("INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES ($1, NULL, 'known')", user.ID)
	if err != nil {
		t.Fatalf("insert null row: %v", err)
	}

	words, err := repo.GetKnownWords(user.ID)
	if err != nil {
		t.Fatalf("GetKnownWords() error = %v", err)
	}
	// One row (valid wordCardID) should be returned; the NULL row is skipped via continue.
	if len(words) != 1 || words[0] != wordCardID {
		t.Errorf("expected [%d], got %v", wordCardID, words)
	}
}
