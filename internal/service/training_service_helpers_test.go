package service

import (
	"testing"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestTrainingService_RestoreQueue_EmptyIDs(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	// Restore queue with empty IDs
	queue, err := service.RestoreQueue(292929, []int64{})
	if err != nil {
		t.Fatalf("RestoreQueue() error = %v", err)
	}
	if len(queue) > 0 {
		t.Errorf("Expected empty queue, got %d cards", len(queue))
	}
}

func TestTrainingService_RestoreQueue_ValidCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create training cards
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "restore1", 0, "восстановить1", "restore 1")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		2, "restore2", 0, "восстановить2", "restore 2")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user cards
	userCard1 := &models.UserCard{
		UserID:         303030,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
	}
	userCardID1, err := userCardRepo.CreateUserCard(userCard1)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	userCard2 := &models.UserCard{
		UserID:         303030,
		TrainingCardID: 2,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
	}
	userCardID2, err := userCardRepo.CreateUserCard(userCard2)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	// Restore queue
	queue, err := service.RestoreQueue(303030, []int64{userCardID1, userCardID2})
	if err != nil {
		t.Fatalf("RestoreQueue() error = %v", err)
	}
	if len(queue) != 2 {
		t.Errorf("Expected 2 cards in queue, got %d", len(queue))
	}
}

func TestTrainingService_RestoreQueue_NonExistentUserCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	// Restore queue with non-existent user card ID
	queue, err := service.RestoreQueue(313131, []int64{99999})
	if err != nil {
		t.Fatalf("RestoreQueue() error = %v", err)
	}
	// Should return empty queue (non-existent card skipped)
	if len(queue) > 0 {
		t.Errorf("Expected empty queue (non-existent card should be skipped), got %d cards", len(queue))
	}
}

func TestTrainingService_RestoreQueue_WrongUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create training card
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "wronguser", 0, "неправильный пользователь", "wrong user")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user card for user 323232
	userCard := &models.UserCard{
		UserID:         323232,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
	}
	userCardID, err := userCardRepo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	// Try to restore queue for different user (333333) with card from user 323232
	queue, err := service.RestoreQueue(333333, []int64{userCardID})
	if err != nil {
		t.Fatalf("RestoreQueue() error = %v", err)
	}
	// Should return empty queue (card belongs to different user)
	if len(queue) > 0 {
		t.Errorf("Expected empty queue (card belongs to different user), got %d cards", len(queue))
	}
}

func TestTrainingService_RestoreQueue_NonExistentTrainingCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create user card with non-existent training_card_id
	userCard := &models.UserCard{
		UserID:         343434,
		TrainingCardID: 99999, // Non-existent training card
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
	}
	userCardID, err := userCardRepo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)

	// Restore queue
	queue, err := service.RestoreQueue(343434, []int64{userCardID})
	if err != nil {
		t.Fatalf("RestoreQueue() error = %v", err)
	}
	// Should return empty queue (training card doesn't exist)
	if len(queue) > 0 {
		t.Errorf("Expected empty queue (training card doesn't exist), got %d cards", len(queue))
	}
}
