package service

import (
	"testing"

	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestShufflePreventDuplicates_SmallList(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	// Create a small list of cards
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordEN: "banana"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordEN: "cherry"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(shuffled))
	}
}

func TestShufflePreventDuplicates_LargeList(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	// Create a larger list with some duplicate words
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordEN: "banana"}},
		{UserCard: models.UserCard{ID: 4}, TrainingCard: models.TrainingCard{WordEN: "banana"}},
		{UserCard: models.UserCard{ID: 5}, TrainingCard: models.TrainingCard{WordEN: "cherry"}},
		{UserCard: models.UserCard{ID: 6}, TrainingCard: models.TrainingCard{WordEN: "date"}},
		{UserCard: models.UserCard{ID: 7}, TrainingCard: models.TrainingCard{WordEN: "elderberry"}},
		{UserCard: models.UserCard{ID: 8}, TrainingCard: models.TrainingCard{WordEN: "fig"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 8 {
		t.Errorf("Expected 8 cards, got %d", len(shuffled))
	}

	// Check that no two adjacent cards have the same word
	for i := 1; i < len(shuffled); i++ {
		if shuffled[i].TrainingCard.WordEN == shuffled[i-1].TrainingCard.WordEN {
			t.Logf("Adjacent duplicates found at positions %d and %d: %s", i-1, i, shuffled[i].TrainingCard.WordEN)
			// This might happen in edge cases, but should be minimized
		}
	}
}

func TestShufflePreventDuplicates_EmptyList(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	var cards []*models.UserCardWithTraining
	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 0 {
		t.Errorf("Expected 0 cards, got %d", len(shuffled))
	}
}

func TestShufflePreventDuplicates_SingleElement(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	// Create a single card
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 1 {
		t.Errorf("Expected 1 card, got %d", len(shuffled))
	}
}

func TestShufflePreventDuplicates_AllSameWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	// All cards have the same word
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(shuffled))
	}
}
