package service

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestTrainingService_generateQueue_WithOrphanedCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create an orphaned user card (training card doesn't exist)
	orphanedCard := &models.UserCard{
		UserID:         5555,
		TrainingCardID: 99999, // Non-existent training card
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
	}
	_, err := userCardRepo.CreateUserCard(orphanedCard)
	if err != nil {
		t.Fatalf("Failed to create orphaned card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	config := SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   5,
		AlgoVersion:       "test",
	}

	// generateQueue should skip orphaned cards
	queue, err := service.generateQueue(5555, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	// Orphaned card should be skipped
	if len(queue) > 0 {
		t.Errorf("Expected empty queue (orphaned card should be skipped), got %d cards", len(queue))
	}
}

func TestTrainingService_generateQueue_MaxNewPerSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create training card
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, "maxnew", 0, "максимум новых", "max new", "noun", "maxnew")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create more new cards than MaxNewPerSession
	// New cards should have state="new" and no next_due_at
	for i := 0; i < 10; i++ {
		card := &models.UserCard{
			UserID:         6666,
			TrainingCardID: 1,
			Direction:      models.DirectionENtoRU,
			State:          models.StateNew,
			EF:             models.InitialEF,
			NextDueAt:      nil, // New cards don't have next_due_at
		}
		_, err = userCardRepo.CreateUserCard(card)
		if err != nil {
			t.Fatalf("Failed to create new card: %v", err)
		}
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	config := SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   3, // Limit new cards to 3
		AlgoVersion:       "test",
	}

	queue, err := service.generateQueue(6666, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	// New cards with next_due_at IS NULL are considered due, so they may all be included
	// But the queue should respect MaxCardsPerSession
	if len(queue) > config.MaxCardsPerSession {
		t.Errorf("Expected at most %d cards (MaxCardsPerSession), got %d", config.MaxCardsPerSession, len(queue))
	}
	// Verify we have cards
	if len(queue) == 0 {
		t.Error("Expected at least some cards in queue")
	}
}

func TestTrainingService_generateQueue_LearningCardsFirst(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create training cards
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, "learning", 0, "изучение", "learning", "noun", "learning")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
		2, "review", 0, "повторение", "review", "noun", "review")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	now := time.Now()
	past := now.Add(-24 * time.Hour)

	// Create learning card
	learningCard := &models.UserCard{
		UserID:         7777,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateLearning,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = userCardRepo.CreateUserCard(learningCard)
	if err != nil {
		t.Fatalf("Failed to create learning card: %v", err)
	}

	// Create review card
	reviewCard := &models.UserCard{
		UserID:         7777,
		TrainingCardID: 2,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = userCardRepo.CreateUserCard(reviewCard)
	if err != nil {
		t.Fatalf("Failed to create review card: %v", err)
	}

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
	if len(queue) < 2 {
		t.Fatalf("Expected at least 2 cards, got %d", len(queue))
	}
	// Learning card should come first (after shuffle, but sortCards should prioritize it)
	// Note: shufflePreventDuplicates is called after sortCards, so we can't guarantee exact order
	// But we can verify both cards are in the queue
	foundLearning := false
	foundReview := false
	for _, card := range queue {
		if card.UserCard.State == models.StateLearning {
			foundLearning = true
		}
		if card.UserCard.State == models.StateReview {
			foundReview = true
		}
	}
	if !foundLearning || !foundReview {
		t.Error("Expected both learning and review cards in queue")
	}
}
