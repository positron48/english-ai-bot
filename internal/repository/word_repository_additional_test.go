package repository

import (
	"testing"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestWordRepository_UpsertWordCardLemma_OnConflictUpdate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	repo := NewWordRepository(db, logger)

	card := &models.WordCard{Word: "upsertword", Definition: "first"}
	id1, err := repo.UpsertWordCardLemma(card)
	if err != nil {
		t.Fatalf("UpsertWordCardLemma() first error = %v", err)
	}
	card.Definition = "updated"
	id2, err := repo.UpsertWordCardLemma(card)
	if err != nil {
		t.Fatalf("UpsertWordCardLemma() second error = %v", err)
	}
	if id1 != id2 {
		t.Errorf("ON CONFLICT update should return same id: %d vs %d", id1, id2)
	}
	got, _ := repo.GetWordCardByLemma("upsertword")
	if got == nil || got.Definition != "updated" {
		t.Errorf("expected definition 'updated', got %v", got)
	}
}

func TestWordRepository_ListPronunciationCandidates_ZeroLimit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	repo := NewWordRepository(db, logger)

	cands, err := repo.ListPronunciationCandidates("", 0)
	if err != nil {
		t.Fatalf("ListPronunciationCandidates(\"\", 0) error = %v", err)
	}
	if cands == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestWordRepository_AddWordRequestHistoryWithCard_WithCardAndWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(2000)
	repo := NewWordRepository(db, logger)

	_ = repo.SaveWordCard("lemma", "definition")
	card, _ := repo.GetWordCard("lemma")
	if card == nil {
		t.Fatal("word card not created")
	}
	wordCardID := card.ID
	inputWord := "inputword"
	word := "lemma"
	err := repo.AddWordRequestHistoryWithCard(user.ID, inputWord, &wordCardID, &word)
	if err != nil {
		t.Fatalf("AddWordRequestHistoryWithCard() error = %v", err)
	}
}

func TestWordRepository_GetUserIDsByWord_WordNotInWordCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(2100)
	repo := NewWordRepository(db, logger)

	// Add history for word that does not exist in word_cards (only input_word/word column)
	_ = repo.AddWordRequestHistoryWithCard(user.ID, "onlyinput", nil, nil)
	userIDs, err := repo.GetUserIDsByWord("onlyinput")
	if err != nil {
		t.Fatalf("GetUserIDsByWord() error = %v", err)
	}
	if len(userIDs) != 1 || userIDs[0] != user.ID {
		t.Errorf("expected [user.ID] when searching by input_word only, got %v", userIDs)
	}
}

func TestWordRepository_GetUserIDsByWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	userRepo := NewUserRepository(db, logger)
	u1, _ := userRepo.GetOrCreateUser(1100)
	u2, _ := userRepo.GetOrCreateUser(1200)
	u3, _ := userRepo.GetOrCreateUser(1300)

	repo := NewWordRepository(db, logger)

	// Add word request history for multiple users
	repo.AddWordRequestHistory(u1.ID, "uniqueword")
	repo.AddWordRequestHistory(u2.ID, "uniqueword")
	repo.AddWordRequestHistory(u3.ID, "uniqueword")
	repo.AddWordRequestHistory(u1.ID, "otherword") // Same user, different word

	// Get user IDs for the word
	userIDs, err := repo.GetUserIDsByWord("uniqueword")
	if err != nil {
		t.Fatalf("GetUserIDsByWord() error = %v", err)
	}
	if len(userIDs) != 3 {
		t.Errorf("Expected 3 unique user IDs, got %d", len(userIDs))
	}
	// Verify all users are in the list
	userMap := make(map[int64]bool)
	for _, id := range userIDs {
		userMap[id] = true
	}
	if !userMap[u1.ID] || !userMap[u2.ID] || !userMap[u3.ID] {
		t.Error("Expected all three user IDs in results")
	}
}

func TestWordRepository_GetUserIDsByWord_CaseInsensitive(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	userRepo := NewUserRepository(db, logger)
	u1, _ := userRepo.GetOrCreateUser(1400)
	u2, _ := userRepo.GetOrCreateUser(1500)
	u3, _ := userRepo.GetOrCreateUser(1600)

	repo := NewWordRepository(db, logger)

	// Add word request history with different cases
	repo.AddWordRequestHistory(u1.ID, "CaseWord")
	repo.AddWordRequestHistory(u2.ID, "caseword")
	repo.AddWordRequestHistory(u3.ID, "CASEWORD")

	// Get user IDs (should be case-insensitive)
	userIDs, err := repo.GetUserIDsByWord("caseword")
	if err != nil {
		t.Fatalf("GetUserIDsByWord() error = %v", err)
	}
	if len(userIDs) < 3 {
		t.Errorf("Expected at least 3 user IDs (case-insensitive), got %d", len(userIDs))
	}
}

func TestWordRepository_GetUserIDsByWord_NoUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)

	repo := NewWordRepository(db, logger)

	// Get user IDs for word with no history
	userIDs, err := repo.GetUserIDsByWord("nonexistent")
	if err != nil {
		t.Fatalf("GetUserIDsByWord() error = %v", err)
	}
	if len(userIDs) != 0 {
		t.Errorf("Expected 0 user IDs, got %d", len(userIDs))
	}
}
