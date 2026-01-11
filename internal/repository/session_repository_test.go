package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupSessionTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewSessionRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)
	if repo == nil {
		t.Error("NewSessionRepository() should not return nil")
	}
}

func TestSessionRepository_CreateSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)

	session := &models.TrainingSession{
		UserID:       123,
		Source:       models.SourceManual,
		PlannedCount: 10,
		DoneCount:    0,
		SessionJSON:  `{"max_cards": 10}`,
	}

	id, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if id == 0 {
		t.Error("CreateSession() should return non-zero ID")
	}
}

func TestSessionRepository_GetSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)

	// Create a session first
	session := &models.TrainingSession{
		UserID:       456,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get the session
	found, err := repo.GetSession(id)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetSession() should not return nil")
	}
	if found.ID != id {
		t.Errorf("Expected ID %d, got %d", id, found.ID)
	}
	if found.UserID != 456 {
		t.Errorf("Expected UserID 456, got %d", found.UserID)
	}
}

func TestSessionRepository_GetActiveSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)

	// Create an active session (no ended_at)
	session := &models.TrainingSession{
		UserID:       789,
		Source:       models.SourceManual,
		PlannedCount: 3,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	_, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get active session
	active, err := repo.GetActiveSession(789)
	if err != nil {
		t.Fatalf("GetActiveSession() error = %v", err)
	}
	if active == nil {
		t.Fatal("GetActiveSession() should not return nil")
	}
	if active.UserID != 789 {
		t.Errorf("Expected UserID 789, got %d", active.UserID)
	}
}

func TestSessionRepository_FinishSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)

	// Create a session
	session := &models.TrainingSession{
		UserID:       999,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Finish the session
	err = repo.FinishSession(id, 3)
	if err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}

	// Verify it's finished
	finished, err := repo.GetSession(id)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if finished.EndedAt == nil {
		t.Error("Session should have ended_at set after FinishSession")
	}
	if finished.DoneCount != 3 {
		t.Errorf("Expected DoneCount 3, got %d", finished.DoneCount)
	}
}

func TestSessionRepository_UpdateSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)
	defer db.Close()

	repo := NewSessionRepository(db, logger)

	// Create a session
	session := &models.TrainingSession{
		UserID:       111,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := repo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Update the session
	session.ID = id
	session.SessionJSON = `{"updated": true}`
	err = repo.UpdateSession(session)
	if err != nil {
		t.Fatalf("UpdateSession() error = %v", err)
	}

	// Verify update
	updated, err := repo.GetSession(id)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if updated.SessionJSON != `{"updated": true}` {
		t.Errorf("Expected SessionJSON %q, got %q", `{"updated": true}`, updated.SessionJSON)
	}
}
