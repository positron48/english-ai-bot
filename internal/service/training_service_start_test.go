package service

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestTrainingService_StartSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(1000)

	// Create word card first
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "start", "to start")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create a training card
	trainingCard := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "start",
		SenseIndex: 0,
		WordRU:     "начать",
		MeaningEN:  "to start",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a due user card
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = userCardRepo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, config.DefaultLearningConfig(), logger)
	session, queue, err := service.StartSession(user.ID, models.SourceManual, nil)
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}
	if session == nil {
		t.Fatal("StartSession() should not return nil session")
	}
	if len(queue) == 0 {
		t.Error("StartSession() should return non-empty queue")
	}
	if session.PlannedCount != len(queue) {
		t.Errorf("Expected PlannedCount %d, got %d", len(queue), session.PlannedCount)
	}
}

func TestTrainingService_StartSession_NoCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, userCardRepo, trainingCardRepo, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(9999)

	service := NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, config.DefaultLearningConfig(), logger)
	_, _, err := service.StartSession(user.ID, models.SourceManual, nil)
	if err == nil {
		t.Error("StartSession() should return error when no cards available")
	}
}

func TestTrainingService_StartSession_FinishOldSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(2000)

	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "finish", "to finish")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create a training card
	trainingCard := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "finish",
		SenseIndex: 0,
		WordRU:     "закончить",
		MeaningEN:  "to finish",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a due user card
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = userCardRepo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, config.DefaultLearningConfig(), logger)

	// Start first session
	session1, _, err := service.StartSession(user.ID, models.SourceManual, nil)
	if err != nil {
		t.Fatalf("Failed to start first session: %v", err)
	}

	// Start second session (should finish the first one)
	session2, _, err := service.StartSession(user.ID, models.SourceManual, nil)
	if err != nil {
		t.Fatalf("Failed to start second session: %v", err)
	}

	// Verify first session is finished
	finished, err := service.GetSession(session1.ID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if finished.EndedAt == nil {
		t.Error("First session should be finished")
	}

	// Verify second session is active
	if session2.ID == session1.ID {
		t.Error("Second session should have different ID")
	}
}
