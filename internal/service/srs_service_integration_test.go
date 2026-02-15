package service

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"tgbot-skeleton/internal/testutil"
	"go.uber.org/zap"
)

func setupSRSServiceTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.UserCardRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)

	return db, userRepo, userCardRepo
}

func TestSRSService_GradeCard_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo := setupSRSServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(111)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "grade", "to grade")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, $1, $2, $3, $4)",
		"grade", 0, "оценить", "to grade")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		Reps:           0,
		IntervalDays:   0,
		LearningStep:   0,
		LapseCount:     0,
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get the card
	userCard, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get user card: %v", err)
	}

	// Create SRS service
	service := NewSRSService(userCardRepo, logger)

	// Grade the card with correct answer
	attemptData := models.AttemptData{
		Correct:      true,
		EarlyReveal:  false,
		AnswerTimeMS: 3000, // Normal speed
	}

	err = service.GradeCard(userCard, attemptData)
	if err != nil {
		t.Fatalf("GradeCard() error = %v", err)
	}

	// Verify the card was updated
	updated, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get updated user card: %v", err)
	}
	if updated.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, updated.State)
	}
	if updated.LastQuality == nil {
		t.Error("LastQuality should be set after grading")
	}
}

func TestSRSService_GradeCard_WrongAnswer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo := setupSRSServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(222)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "wrong", "wrong")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, $1, $2, $3, $4)",
		"wrong", 0, "неправильно", "wrong")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card in review state
	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           5,
		IntervalDays:   10,
		NextDueAt:      &now,
	}
	id, err := userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get the card
	userCard, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get user card: %v", err)
	}

	// Create SRS service
	service := NewSRSService(userCardRepo, logger)

	// Grade the card with wrong answer
	attemptData := models.AttemptData{
		Correct: false,
	}

	err = service.GradeCard(userCard, attemptData)
	if err != nil {
		t.Fatalf("GradeCard() error = %v", err)
	}

	// Verify the card uses gentle approach (stays in review, interval reduced)
	updated, err := userCardRepo.GetUserCard(id)
	if err != nil {
		t.Fatalf("Failed to get updated user card: %v", err)
	}
	// Should stay in review for first error
	if updated.State != models.StateReview {
		t.Errorf("Expected State %v, got %v (should stay in review for first error)", models.StateReview, updated.State)
	}
	// Interval should be reduced (10 / 2 = 5)
	if updated.IntervalDays != 5 {
		t.Errorf("Expected IntervalDays 5 (10/2), got %d", updated.IntervalDays)
	}
	// Reps should be preserved
	if updated.Reps != 5 {
		t.Errorf("Expected Reps 5 (preserved), got %d", updated.Reps)
	}
	if updated.LapseCount != 1 {
		t.Errorf("Expected LapseCount 1, got %d", updated.LapseCount)
	}
	// EF should be reduced
	if updated.EF >= 2.0 {
		t.Errorf("Expected EF < 2.0, got %f", updated.EF)
	}
}
