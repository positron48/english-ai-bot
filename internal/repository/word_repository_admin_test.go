package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupWordAdminTestDB(t *testing.T) (*sql.DB, *WordRepository) {
	db := testutil.SetupTestDB(t)
	logger, _ := zap.NewDevelopment()
	wordRepo := NewWordRepository(db, logger)
	return db, wordRepo
}

func TestWordRepository_ListWordCardsAdmin(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word cards
	repo.SaveWordCard("admin1", "definition 1")
	repo.SaveWordCard("admin2", "definition 2")

	// List word cards
	cards, err := repo.ListWordCardsAdmin(nil, false, nil, "", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	if len(cards) < 2 {
		t.Errorf("Expected at least 2 cards, got %d", len(cards))
	}
}

func TestWordRepository_ListWordCardsAdmin_WithSearch(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word cards
	repo.SaveWordCard("searchword", "definition")
	repo.SaveWordCard("otherword", "definition")

	// List word cards with search
	cards, err := repo.ListWordCardsAdmin(nil, false, nil, "search", "", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	if len(cards) == 0 {
		t.Error("Expected at least one card matching search")
	}
}

func TestWordRepository_CountWordCardsAdmin(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word cards
	repo.SaveWordCard("count1", "definition 1")
	repo.SaveWordCard("count2", "definition 2")
	repo.SaveWordCard("count3", "definition 3")

	// Count word cards
	count, err := repo.CountWordCardsAdmin(nil, false, nil, "", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 3 {
		t.Errorf("Expected at least 3 cards, got %d", count)
	}
}

func TestWordRepository_CountWordCardsAdmin_WithFilterUserID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, repo := setupWordAdminTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(100)

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

	// Add request history for user with word_card_id
	word1 := "userword1"
	word2 := "userword2"
	repo.AddWordRequestHistoryWithCard(user.ID, "userword1", &card1.ID, &word1)
	repo.AddWordRequestHistoryWithCard(user.ID, "userword2", &card2.ID, &word2)

	// Count word cards for user
	userID := user.ID
	count, err := repo.CountWordCardsAdmin(&userID, false, nil, "", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 cards for user, got %d", count)
	}
}

func TestWordRepository_CountWordCardsAdmin_WithErrors(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

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
	count, err := repo.CountWordCardsAdmin(nil, true, nil, "", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 card with error, got %d", count)
	}
}

func TestWordRepository_CountWordCardsAdmin_WithSearch(t *testing.T) {
	_, repo := setupWordAdminTestDB(t)

	// Create word cards
	repo.SaveWordCard("searchable", "definition")
	repo.SaveWordCard("other", "definition")

	// Count word cards with search
	count, err := repo.CountWordCardsAdmin(nil, false, nil, "search", "")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 card matching search, got %d", count)
	}
}

func TestWordRepository_ListWordCardsAdmin_MissingTrainingPOS(t *testing.T) {
	db, repo := setupWordAdminTestDB(t)
	logger, _ := zap.NewDevelopment()
	tcRepo := NewTrainingCardRepository(db, logger)

	// Word A: has a training card with pos=noun
	repo.SaveWordCard("wordwithnoun", "definition")
	cardA, _ := repo.GetWordCard("wordwithnoun")
	noun := "noun"
	_, err := tcRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: cardA.ID, WordEN: "wordwithnoun", SenseIndex: 0,
		WordRU: "слово", MeaningEN: "meaning", POS: &noun,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	// Word B: no training card with pos=noun (no cards at all)
	repo.SaveWordCard("wordwithoutnoun", "definition")

	// Filter: missing card for noun -> only words that have no training card with pos=noun
	list, err := repo.ListWordCardsAdmin(nil, false, nil, "", "noun", 10, 0, "", "desc")
	if err != nil {
		t.Fatalf("ListWordCardsAdmin() error = %v", err)
	}
	// Should contain only wordwithoutnoun (wordwithnoun has a noun card)
	var found bool
	for _, w := range list {
		if w.Word == "wordwithoutnoun" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find wordwithoutnoun in list when filtering by missing_training_pos=noun, got %d words", len(list))
	}
	for _, w := range list {
		if w.Word == "wordwithnoun" {
			t.Error("wordwithnoun has a noun card and must not appear when filtering by missing_training_pos=noun")
			break
		}
	}

	count, err := repo.CountWordCardsAdmin(nil, false, nil, "", "noun")
	if err != nil {
		t.Fatalf("CountWordCardsAdmin() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 word without noun card, got count %d", count)
	}
}

func TestWordRepository_GetWordCardRequestingUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, repo := setupWordAdminTestDB(t)
	userRepo := NewUserRepository(db, logger)
	u1, _ := userRepo.GetOrCreateUser(100)
	u2, _ := userRepo.GetOrCreateUser(200)

	// Create a word card
	repo.SaveWordCard("requesting", "definition")

	// Get word card to get its ID
	card, err := repo.GetWordCard("requesting")
	if err != nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	// Add request history
	repo.AddWordRequestHistory(u1.ID, "requesting")
	repo.AddWordRequestHistory(u2.ID, "requesting")

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
	_, repo := setupWordAdminTestDB(t)

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
