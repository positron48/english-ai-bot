package repository

import (
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestTrainingCardRepository_ListOrphanedTrainingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)

	// Create a word card
	wordCard := &models.WordCard{
		Word:       "orphan",
		Definition: "definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create a training card
	pos := "noun"
	displayWord := "orphan"
	trainingCard := &models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "orphan",
		SenseIndex:  0,
		WordRU:      "сирота",
		MeaningEN:   "orphan",
		POS:         &pos,
		DisplayWord: &displayWord,
	}
	trainingCardID, err := repo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Delete the word card to make training card orphaned
	_, err = db.Exec("DELETE FROM word_cards WHERE id = ?", wordCardID)
	if err != nil {
		t.Fatalf("Failed to delete word card: %v", err)
	}

	// List orphaned training cards
	orphaned, err := repo.ListOrphanedTrainingCards(10, 0)
	if err != nil {
		t.Fatalf("ListOrphanedTrainingCards() error = %v", err)
	}
	if len(orphaned) == 0 {
		t.Error("Expected at least one orphaned training card")
	}

	// Verify structure
	found := false
	for _, card := range orphaned {
		if card.TrainingCardID == trainingCardID {
			found = true
			if card.WordEN == "" {
				t.Error("WordEN should not be empty")
			}
			if card.WordRU == "" {
				t.Error("WordRU should not be empty")
			}
		}
	}
	if !found {
		t.Error("Expected orphaned training card not found in results")
	}
}

func TestTrainingCardRepository_CountOrphanedTrainingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)

	// Create word cards and training cards
	wordCard1 := &models.WordCard{
		Word:       "orphan1",
		Definition: "definition1",
	}
	wordCardID1, err := wordRepo.UpsertWordCardLemma(wordCard1)
	if err != nil {
		t.Fatalf("Failed to create word card 1: %v", err)
	}

	wordCard2 := &models.WordCard{
		Word:       "orphan2",
		Definition: "definition2",
	}
	wordCardID2, err := wordRepo.UpsertWordCardLemma(wordCard2)
	if err != nil {
		t.Fatalf("Failed to create word card 2: %v", err)
	}

	// Create training cards
	pos := "noun"
	displayWord1 := "orphan1"
	trainingCard1 := &models.TrainingCard{
		WordCardID:  wordCardID1,
		WordEN:      "orphan1",
		SenseIndex:  0,
		WordRU:      "сирота1",
		MeaningEN:   "orphan1",
		POS:         &pos,
		DisplayWord: &displayWord1,
	}
	_, err = repo.CreateTrainingCard(trainingCard1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}

	displayWord2 := "orphan2"
	trainingCard2 := &models.TrainingCard{
		WordCardID:  wordCardID2,
		WordEN:      "orphan2",
		SenseIndex:  0,
		WordRU:      "сирота2",
		MeaningEN:   "orphan2",
		POS:         &pos,
		DisplayWord: &displayWord2,
	}
	_, err = repo.CreateTrainingCard(trainingCard2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Delete word cards to make training cards orphaned
	_, err = db.Exec("DELETE FROM word_cards WHERE id IN (?, ?)", wordCardID1, wordCardID2)
	if err != nil {
		t.Fatalf("Failed to delete word cards: %v", err)
	}

	// Count orphaned training cards
	count, err := repo.CountOrphanedTrainingCards()
	if err != nil {
		t.Fatalf("CountOrphanedTrainingCards() error = %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 orphaned training cards, got %d", count)
	}
}

func TestTrainingCardRepository_ListOrphanedTrainingCards_Empty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)

	// List orphaned training cards when there are none
	orphaned, err := repo.ListOrphanedTrainingCards(10, 0)
	if err != nil {
		t.Fatalf("ListOrphanedTrainingCards() error = %v", err)
	}
	if len(orphaned) != 0 {
		t.Errorf("Expected 0 orphaned training cards, got %d", len(orphaned))
	}
}

func TestTrainingCardRepository_CountOrphanedTrainingCards_Empty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)

	// Count orphaned training cards when there are none
	count, err := repo.CountOrphanedTrainingCards()
	if err != nil {
		t.Fatalf("CountOrphanedTrainingCards() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 orphaned training cards, got %d", count)
	}
}
