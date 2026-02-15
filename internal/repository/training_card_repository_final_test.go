package repository

import (
	"testing"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestTrainingCardRepository_DeleteAllTrainingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)

	repo := NewTrainingCardRepository(db, logger)

	// Create word cards
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "deleteall", "to delete all")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training cards
	card1 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "deleteall",
		SenseIndex: 0,
		WordRU:     "удалить все",
		MeaningEN:  "to delete all",
	}
	card2 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "deleteall",
		SenseIndex: 1,
		WordRU:     "удалить все 2",
		MeaningEN:  "to delete all 2",
	}

	_, err = repo.CreateTrainingCard(card1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}
	_, err = repo.CreateTrainingCard(card2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Delete all training cards
	deleted, err := repo.DeleteAllTrainingCards()
	if err != nil {
		t.Fatalf("DeleteAllTrainingCards() error = %v", err)
	}
	if deleted < 2 {
		t.Errorf("Expected at least 2 cards deleted, got %d", deleted)
	}

	// Verify deletion
	cards, err := repo.GetTrainingCardsByWordCardID(1)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID() error = %v", err)
	}
	if len(cards) != 0 {
		t.Errorf("Expected 0 cards after deletion, got %d", len(cards))
	}
}
