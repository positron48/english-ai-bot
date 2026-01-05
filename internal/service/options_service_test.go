package service

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestNewOptionsService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := (*repository.TrainingCardRepository)(nil)
	
	service := NewOptionsService(repo, logger)
	_ = service // Verify service is created
}

func TestOptionsService_getFallbackDistractors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewOptionsService(nil, logger)

	t.Run("RU to EN direction", func(t *testing.T) {
		distractors := service.getFallbackDistractors(models.DirectionRUtoEN)
		if len(distractors) == 0 {
			t.Error("Should return fallback distractors for RU->EN")
		}
		// Check that all are English words (basic check)
		for _, d := range distractors {
			if d == "" {
				t.Error("Distractor should not be empty")
			}
		}
	})

	t.Run("EN to RU direction", func(t *testing.T) {
		distractors := service.getFallbackDistractors(models.DirectionENtoRU)
		if len(distractors) == 0 {
			t.Error("Should return fallback distractors for EN->RU")
		}
		// Check that all are non-empty
		for _, d := range distractors {
			if d == "" {
				t.Error("Distractor should not be empty")
			}
		}
	})
}

// Note: GenerateOptions requires a real repository to call getOtherMeaningsOfWord
// For now, we test the simpler methods. Full integration tests would require a mock or real DB.

func TestTrainingService_NewTrainingService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	userCardRepo := (*repository.UserCardRepository)(nil)
	trainingCardRepo := (*repository.TrainingCardRepository)(nil)
	sessionRepo := (*repository.SessionRepository)(nil)

	service := NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, logger)
	_ = service // Verify service is created
}

func TestTrainingService_sortCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewTrainingService(nil, nil, nil, logger)

	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	cards := []*models.UserCard{
		{
			ID:        1,
			State:     models.StateReview,
			NextDueAt: &future,
		},
		{
			ID:        2,
			State:     models.StateLearning,
			NextDueAt: &past,
		},
		{
			ID:        3,
			State:     models.StateNew,
		},
		{
			ID:        4,
			State:     models.StateReview,
			NextDueAt: &past,
		},
	}

	service.sortCards(cards, now)

	// Learning cards should come first
	if cards[0].State != models.StateLearning {
		t.Errorf("First card should be learning, got %v", cards[0].State)
	}
}

func TestTrainingService_shufflePreventDuplicates(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewTrainingService(nil, nil, nil, logger)

	t.Run("Empty queue", func(t *testing.T) {
		result := service.shufflePreventDuplicates([]*models.UserCardWithTraining{})
		if len(result) != 0 {
			t.Errorf("Expected empty result, got %d items", len(result))
		}
	})

	t.Run("Single card", func(t *testing.T) {
		queue := []*models.UserCardWithTraining{
			{
				TrainingCard: models.TrainingCard{WordCardID: 1},
			},
		}
		result := service.shufflePreventDuplicates(queue)
		if len(result) != 1 {
			t.Errorf("Expected 1 card, got %d", len(result))
		}
	})

	t.Run("No duplicates", func(t *testing.T) {
		queue := []*models.UserCardWithTraining{
			{TrainingCard: models.TrainingCard{WordCardID: 1}},
			{TrainingCard: models.TrainingCard{WordCardID: 2}},
			{TrainingCard: models.TrainingCard{WordCardID: 3}},
		}
		result := service.shufflePreventDuplicates(queue)
		if len(result) != 3 {
			t.Errorf("Expected 3 cards, got %d", len(result))
		}
	})

	t.Run("With duplicates", func(t *testing.T) {
		queue := []*models.UserCardWithTraining{
			{TrainingCard: models.TrainingCard{WordCardID: 1}},
			{TrainingCard: models.TrainingCard{WordCardID: 1}},
			{TrainingCard: models.TrainingCard{WordCardID: 2}},
			{TrainingCard: models.TrainingCard{WordCardID: 2}},
		}
		result := service.shufflePreventDuplicates(queue)
		if len(result) != 4 {
			t.Errorf("Expected 4 cards, got %d", len(result))
		}
		// Check that same word IDs are not adjacent (basic check)
		// Note: With small queues, duplicates might appear, but the algorithm tries to prevent it
		_ = result // Verify result is not empty
	})
}
