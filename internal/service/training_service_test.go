package service

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupTrainingServiceTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.UserCardRepository, *repository.TrainingCardRepository, *repository.SessionRepository) {
	db := testutil.SetupTestDB(t)
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	return db, userRepo, userCardRepo, trainingCardRepo, sessionRepo
}

func TestTrainingService_GetDueCount(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, _, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(333)

	// Create word cards first
	var err error
	for i := 1; i <= 2; i++ {
		_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", i, fmt.Sprintf("test%d", i), fmt.Sprintf("test %d", i))
		if err != nil {
			t.Fatalf("Failed to create word card: %v", err)
		}
	}

	// Create training cards
	for i := 1; i <= 2; i++ {
		_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
			i, "test", 0, "тест", "test", "noun", "test")
		if err != nil {
			t.Fatalf("Failed to create training card: %v", err)
		}
	}

	// Create due cards
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	for i := 1; i <= 2; i++ {
		card := &models.UserCard{
			UserID:         user.ID,
			TrainingCardID: int64(i),
			Direction:      models.DirectionENtoRU,
			State:          models.StateReview,
			EF:             2.0,
			NextDueAt:      &past,
		}
		_, err = userCardRepo.CreateUserCard(card)
		if err != nil {
			t.Fatalf("Failed to create user card: %v", err)
		}
	}

	service := NewTrainingService(userCardRepo, nil, nil, nil, logger)
	count, err := service.GetDueCount(user.ID)
	if err != nil {
		t.Fatalf("GetDueCount() error = %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 due cards, got %d", count)
	}
}

func TestTrainingService_GetSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(444)

	// Create a session
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service := NewTrainingService(nil, nil, sessionRepo, nil, logger)
	found, err := service.GetSession(id)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetSession() should not return nil")
	}
	if found.ID != id {
		t.Errorf("Expected session ID %d, got %d", id, found.ID)
	}
}

func TestTrainingService_GetActiveSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(555)

	// Create an active session
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 3,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	_, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service := NewTrainingService(nil, nil, sessionRepo, nil, logger)
	active, err := service.GetActiveSession(user.ID)
	if err != nil {
		t.Fatalf("GetActiveSession() error = %v", err)
	}
	if active == nil {
		t.Fatal("GetActiveSession() should not return nil")
	}
	if active.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, active.UserID)
	}
}

func TestTrainingService_FinishSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(666)

	// Create a session
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service := NewTrainingService(nil, nil, sessionRepo, nil, logger)
	err = service.FinishSession(id, 3)
	if err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}

	// Verify session is finished
	finished, err := service.GetSession(id)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if finished.EndedAt == nil {
		t.Error("Session should have ended_at set after FinishSession")
	}
}

func TestTrainingService_UpdateSessionState(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(777)

	// Create a session
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service := NewTrainingService(nil, nil, sessionRepo, nil, logger)
	err = service.UpdateSessionState(id, `{"updated": true}`)
	if err != nil {
		t.Fatalf("UpdateSessionState() error = %v", err)
	}

	// Verify update
	updated, err := service.GetSession(id)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if updated.SessionJSON != `{"updated": true}` {
		t.Errorf("Expected SessionJSON %q, got %q", `{"updated": true}`, updated.SessionJSON)
	}
}

func TestTrainingService_RestoreQueue(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(888)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "restore", "to restore")
	// Create a training card
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "restore",
		SenseIndex: 0,
		WordRU:     "восстановить",
		MeaningEN:  "to restore",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user cards
	now := time.Now()
	userCard1 := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &now,
	}
	userCardID1, err := userCardRepo.CreateUserCard(userCard1)
	if err != nil {
		t.Fatalf("Failed to create user card 1: %v", err)
	}

	userCard2 := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateLearning,
		EF:             2.0,
		NextDueAt:      &now,
	}
	userCardID2, err := userCardRepo.CreateUserCard(userCard2)
	if err != nil {
		t.Fatalf("Failed to create user card 2: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	queue, err := service.RestoreQueue(user.ID, []int64{userCardID1, userCardID2})
	if err != nil {
		t.Fatalf("RestoreQueue() error = %v", err)
	}
	if len(queue) != 2 {
		t.Errorf("Expected 2 cards in queue, got %d", len(queue))
	}
}
