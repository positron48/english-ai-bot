package repository

import (
	"testing"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestTrainingCardRepository_GetTrainingCardsByWordEN(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)

	repo := NewTrainingCardRepository(db, logger)

	// Create word cards
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "word1", "def1")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	_, err = db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "word2", "def2")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training cards for word1
	card1 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "word1",
		SenseIndex: 0,
		WordRU:     "слово1",
		MeaningEN:  "meaning 1",
	}
	card2 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "word1",
		SenseIndex: 1,
		WordRU:     "слово1-2",
		MeaningEN:  "meaning 2",
	}

	_, err = repo.CreateTrainingCard(card1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}
	_, err = repo.CreateTrainingCard(card2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Get training cards by word EN
	cards, err := repo.GetTrainingCardsByWordEN("word1")
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordEN() error = %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(cards))
	}
}

func TestTrainingCardRepository_DeleteTrainingCardsByWordEN(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)

	repo := NewTrainingCardRepository(db, logger)

	// Create word cards
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "delete", "to delete")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training cards
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "delete",
		SenseIndex: 0,
		WordRU:     "удалить",
		MeaningEN:  "to delete",
	}
	_, err = repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Delete training cards by word EN
	deleted, err := repo.DeleteTrainingCardsByWordEN("delete")
	if err != nil {
		t.Fatalf("DeleteTrainingCardsByWordEN() error = %v", err)
	}
	if deleted == 0 {
		t.Error("DeleteTrainingCardsByWordEN() should delete at least one card")
	}

	// Verify deletion
	cards, err := repo.GetTrainingCardsByWordEN("delete")
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordEN() error = %v", err)
	}
	if len(cards) != 0 {
		t.Errorf("Expected 0 cards after deletion, got %d", len(cards))
	}
}

func TestTrainingCardRepository_UpdateTrainingCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)

	repo := NewTrainingCardRepository(db, logger)

	// Create word card first
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "update", "to update")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create a training card
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "update",
		SenseIndex: 0,
		WordRU:     "обновить",
		MeaningEN:  "to update",
	}
	id, err := repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Update the card
	card.ID = id
	card.MeaningEN = "updated meaning"
	card.ExampleEN = "Updated example"

	err = repo.UpdateTrainingCard(card)
	if err != nil {
		t.Fatalf("UpdateTrainingCard() error = %v", err)
	}

	// Verify update
	updated, err := repo.GetTrainingCard(id)
	if err != nil {
		t.Fatalf("GetTrainingCard() error = %v", err)
	}
	if updated.MeaningEN != "updated meaning" {
		t.Errorf("Expected MeaningEN 'updated meaning', got %q", updated.MeaningEN)
	}
}

func TestTrainingCardRepository_DeleteTrainingCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)

	repo := NewTrainingCardRepository(db, logger)

	// Create word card first
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "delete", "to delete")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create a training card
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "delete",
		SenseIndex: 0,
		WordRU:     "удалить",
		MeaningEN:  "to delete",
	}
	id, err := repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Delete the card
	err = repo.DeleteTrainingCard(id)
	if err != nil {
		t.Fatalf("DeleteTrainingCard() error = %v", err)
	}

	// Verify deletion
	deleted, err := repo.GetTrainingCard(id)
	if err != nil {
		t.Fatalf("GetTrainingCard() error = %v", err)
	}
	if deleted != nil {
		t.Error("GetTrainingCard() should return nil for deleted card")
	}
}
