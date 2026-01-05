package service

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestTrainingService_generateQueue_WithDueAndNewCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create training cards
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "queue1", 0, "очередь1", "queue 1")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		2, "queue2", 0, "очередь2", "queue 2")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create due cards
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	dueCard := &models.UserCard{
		UserID:         9999,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = userCardRepo.CreateUserCard(dueCard)
	if err != nil {
		t.Fatalf("Failed to create due card: %v", err)
	}

	// Create new cards
	newCard := &models.UserCard{
		UserID:         9999,
		TrainingCardID: 2,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	_, err = userCardRepo.CreateUserCard(newCard)
	if err != nil {
		t.Fatalf("Failed to create new card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	config := SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   5,
		AlgoVersion:       "test",
	}

	queue, err := service.generateQueue(9999, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) < 2 {
		t.Errorf("Expected at least 2 cards in queue, got %d", len(queue))
	}
}

func TestTrainingService_generateQueue_OnlyDueCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create training card
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "onlydue", 0, "только просроченные", "only due")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create only due cards
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	for i := 0; i < 3; i++ {
		card := &models.UserCard{
			UserID:         8888,
			TrainingCardID: 1,
			Direction:      models.DirectionENtoRU,
			State:          models.StateReview,
			EF:             2.0,
			NextDueAt:      &past,
		}
		_, err = userCardRepo.CreateUserCard(card)
		if err != nil {
			t.Fatalf("Failed to create due card: %v", err)
		}
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	config := SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   5,
		AlgoVersion:       "test",
	}

	queue, err := service.generateQueue(8888, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) < 3 {
		t.Errorf("Expected at least 3 cards in queue, got %d", len(queue))
	}
}

func TestTrainingService_generateQueue_Empty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	config := SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   5,
		AlgoVersion:       "test",
	}

	queue, err := service.generateQueue(7777, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != 0 {
		t.Errorf("Expected empty queue, got %d cards", len(queue))
	}
}
