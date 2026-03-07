package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/database"
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

	repo := NewSessionRepository(db, logger)
	if repo == nil {
		t.Error("NewSessionRepository() should not return nil")
	}
}

func TestSessionRepository_CreateSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(123)

	repo := NewSessionRepository(db, logger)

	session := &models.TrainingSession{
		UserID:       user.ID,
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

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(456)

	repo := NewSessionRepository(db, logger)

	// Create a session first
	session := &models.TrainingSession{
		UserID:       user.ID,
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
	if found.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, found.UserID)
	}
}

func TestSessionRepository_GetActiveSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(789)
	otherUser, _ := userRepo.GetOrCreateUser(790)

	repo := NewSessionRepository(db, logger)

	t.Run("no active session returns nil", func(t *testing.T) {
		active, err := repo.GetActiveSession(otherUser.ID)
		if err != nil {
			t.Fatalf("GetActiveSession() error = %v", err)
		}
		if active != nil {
			t.Errorf("expected nil when user has no session, got %+v", active)
		}
	})

	t.Run("returns active session when exists", func(t *testing.T) {
		session := &models.TrainingSession{
			UserID:       user.ID,
			Source:       models.SourceManual,
			PlannedCount: 3,
			DoneCount:    0,
			SessionJSON:  `{}`,
		}
		_, err := repo.CreateSession(session)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		active, err := repo.GetActiveSession(user.ID)
		if err != nil {
			t.Fatalf("GetActiveSession() error = %v", err)
		}
		if active == nil {
			t.Fatal("GetActiveSession() should not return nil")
		}
		if active.UserID != user.ID {
			t.Errorf("Expected UserID %d, got %d", user.ID, active.UserID)
		}
	})

	t.Run("returns session with all fields populated", func(t *testing.T) {
		session := &models.TrainingSession{
			UserID:       user.ID,
			Source:       models.SourceManual,
			PlannedCount: 5,
			DoneCount:    2,
			SessionJSON:  `{"key": "value"}`,
		}
		id, err := repo.CreateSession(session)
		if err != nil {
			t.Fatalf("CreateSession() error = %v", err)
		}

		active, err := repo.GetActiveSession(user.ID)
		if err != nil {
			t.Fatalf("GetActiveSession() error = %v", err)
		}
		if active == nil {
			t.Fatal("GetActiveSession() should not return nil")
		}
		if active.ID != id {
			t.Errorf("Expected ID %d, got %d", id, active.ID)
		}
		if active.Source != models.SourceManual {
			t.Errorf("Expected Source %q, got %q", models.SourceManual, active.Source)
		}
		if active.PlannedCount != 5 {
			t.Errorf("Expected PlannedCount 5, got %d", active.PlannedCount)
		}
		if active.DoneCount != 2 {
			t.Errorf("Expected DoneCount 2, got %d", active.DoneCount)
		}
		if active.SessionJSON != `{"key": "value"}` {
			t.Errorf("Expected SessionJSON %q, got %q", `{"key": "value"}`, active.SessionJSON)
		}
		if active.EndedAt != nil {
			t.Error("EndedAt should be nil for active session")
		}
	})

	t.Run("returns error when DB fails", func(t *testing.T) {
		dsn := testutil.SecondPostgresDSN(t)
		dbWrap, err := database.NewWithConfig("postgres", "", dsn, logger)
		if err != nil {
			t.Skipf("second DB not available (e.g. Docker): %v", err)
		}
		conn := dbWrap.GetConnection()
		_ = dbWrap.Close()

		badRepo := NewSessionRepository(conn, logger)
		_, err = badRepo.GetActiveSession(user.ID)
		if err == nil {
			t.Error("GetActiveSession() expected error on closed DB")
		}
	})
}

func TestSessionRepository_FinishSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionTestDB(t)

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(999)

	repo := NewSessionRepository(db, logger)

	// Create a session
	session := &models.TrainingSession{
		UserID:       user.ID,
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

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(111)

	repo := NewSessionRepository(db, logger)

	// Create a session
	session := &models.TrainingSession{
		UserID:       user.ID,
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
