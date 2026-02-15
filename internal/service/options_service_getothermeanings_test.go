package service

import (
	"testing"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestOptionsService_getOtherMeaningsOfWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	// Create word cards
	_, err := db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "meanings", "def")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create multiple training cards for the same word
	card1 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "meanings",
		SenseIndex: 0,
		WordRU:     "значение1",
		MeaningEN:  "meaning 1",
	}
	card2 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "meanings",
		SenseIndex: 1,
		WordRU:     "значение2",
		MeaningEN:  "meaning 2",
	}

	_, err = trainingCardRepo.CreateTrainingCard(card1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}
	_, err = trainingCardRepo.CreateTrainingCard(card2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger)

	// Test for RU->EN direction
	meanings := service.getOtherMeaningsOfWord(1, models.DirectionRUtoEN)
	if len(meanings) < 1 {
		t.Errorf("Expected at least 1 meaning, got %d", len(meanings))
	}

	// Test for EN->RU direction
	meanings = service.getOtherMeaningsOfWord(1, models.DirectionENtoRU)
	if len(meanings) < 1 {
		t.Errorf("Expected at least 1 meaning, got %d", len(meanings))
	}
}
