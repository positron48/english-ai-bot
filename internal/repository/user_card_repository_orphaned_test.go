package repository

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestUserCardRepository_ListUserCardsWithOrphanedTrainingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12358)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card and training card
	wordCard := &models.WordCard{
		Word:       "orphantraining",
		Definition: "definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	pos := "noun"
	displayWord := "orphantraining"
	trainingCard := &models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "orphantraining",
		SenseIndex:  0,
		WordRU:      "сирота",
		MeaningEN:   "orphan",
		POS:         &pos,
		DisplayWord: &displayWord,
	}
	trainingCardID, err := trainingRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user card
	now := time.Now()
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Delete word card to make training card orphaned
	_, err = db.Exec("DELETE FROM word_cards WHERE id = ?", wordCardID)
	if err != nil {
		t.Fatalf("Failed to delete word card: %v", err)
	}

	// List user cards with orphaned training cards
	orphaned, err := repo.ListUserCardsWithOrphanedTrainingCards(10, 0)
	if err != nil {
		t.Fatalf("ListUserCardsWithOrphanedTrainingCards() error = %v", err)
	}
	if len(orphaned) == 0 {
		t.Error("Expected at least one user card with orphaned training card")
	}

	// Verify structure
	found := false
	for _, item := range orphaned {
		if item.TrainingCardID == trainingCardID {
			found = true
			if item.UserID == 0 {
				t.Error("UserID should not be zero")
			}
			if item.TrainingCardID == 0 {
				t.Error("TrainingCardID should not be zero")
			}
		}
	}
	if !found {
		t.Error("Expected user card with orphaned training card not found")
	}
}

func TestUserCardRepository_CountUserCardsWithOrphanedTrainingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12359)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card and training card
	wordCard := &models.WordCard{
		Word:       "countorphan",
		Definition: "definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	pos := "noun"
	displayWord := "countorphan"
	trainingCard := &models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "countorphan",
		SenseIndex:  0,
		WordRU:      "сирота",
		MeaningEN:   "orphan",
		POS:         &pos,
		DisplayWord: &displayWord,
	}
	trainingCardID, err := trainingRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user card
	now := time.Now()
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Delete word card to make training card orphaned
	_, err = db.Exec("DELETE FROM word_cards WHERE id = ?", wordCardID)
	if err != nil {
		t.Fatalf("Failed to delete word card: %v", err)
	}

	// Count user cards with orphaned training cards
	count, err := repo.CountUserCardsWithOrphanedTrainingCards()
	if err != nil {
		t.Fatalf("CountUserCardsWithOrphanedTrainingCards() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 user card with orphaned training card, got %d", count)
	}
}
