package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupSessionAdditionalTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestSessionRepository_CreateReviewEvent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionAdditionalTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)

	// Create a session first
	session := &models.TrainingSession{
		UserID:       111,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	sessionID, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create a review event
	now := time.Now()
	event := &models.ReviewEvent{
		SessionID:    &sessionID,
		UserID:       111,
		UserCardID:   1,
		Direction:    models.DirectionENtoRU,
		ShownAt:      now,
		AnsweredAt:   &now,
		IsCorrect:    true,
		Quality:      2,
		OptionCount:  4,
		ChosenOption: "correct",
	}

	id, err := repo.CreateReviewEvent(event)
	if err != nil {
		t.Fatalf("CreateReviewEvent() error = %v", err)
	}
	if id == 0 {
		t.Error("CreateReviewEvent() should return non-zero ID")
	}
}

func TestSessionRepository_GetSessionStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionAdditionalTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)

	// Create a session
	session := &models.TrainingSession{
		UserID:       222,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	sessionID, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create review events
	now := time.Now()
	event1 := &models.ReviewEvent{
		SessionID:  &sessionID,
		UserID:     222,
		UserCardID: 1,
		Direction:  models.DirectionENtoRU,
		ShownAt:    now,
		AnsweredAt: &now,
		IsCorrect:  true,
		Quality:    2,
	}
	event2 := &models.ReviewEvent{
		SessionID:  &sessionID,
		UserID:     222,
		UserCardID: 2,
		Direction:  models.DirectionENtoRU,
		ShownAt:    now,
		AnsweredAt: &now,
		IsCorrect:  false,
		Quality:    0,
	}

	_, err = repo.CreateReviewEvent(event1)
	if err != nil {
		t.Fatalf("Failed to create review event 1: %v", err)
	}
	_, err = repo.CreateReviewEvent(event2)
	if err != nil {
		t.Fatalf("Failed to create review event 2: %v", err)
	}

	// Get session stats
	total, correct, err := repo.GetSessionStats(sessionID)
	if err != nil {
		t.Fatalf("GetSessionStats() error = %v", err)
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
	if correct != 1 {
		t.Errorf("Expected correct 1, got %d", correct)
	}
}

func TestSessionRepository_GetTodaySessionCount(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionAdditionalTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)

	// Create sessions for today
	// Use UTC time to match SQLite's CURRENT_TIMESTAMP which returns UTC
	today := time.Now().UTC().Format("2006-01-02")
	for i := 0; i < 3; i++ {
		session := &models.TrainingSession{
			UserID:       333,
			Source:       models.SourceManual,
			PlannedCount: 5,
			DoneCount:    0,
			SessionJSON:  `{}`,
		}
		_, err := repo.CreateSession(session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}

	// Get today's session count
	count, err := repo.GetTodaySessionCount(333, today)
	if err != nil {
		t.Fatalf("GetTodaySessionCount() error = %v", err)
	}
	if count < 3 {
		t.Errorf("Expected at least 3 sessions, got %d", count)
	}
}
