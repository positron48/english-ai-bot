package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupSessionAdditionalTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS training_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ended_at TEXT,
		source TEXT NOT NULL,
		planned_count INTEGER NOT NULL DEFAULT 0,
		done_count INTEGER NOT NULL DEFAULT 0,
		session_json TEXT DEFAULT ''
	);
	
	CREATE TABLE IF NOT EXISTS review_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER,
		user_id INTEGER NOT NULL,
		user_card_id INTEGER NOT NULL,
		direction TEXT NOT NULL,
		shown_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		options_shown_at TEXT,
		answered_at TEXT,
		t_delay_ms INTEGER NOT NULL DEFAULT 0,
		early_reveal INTEGER NOT NULL DEFAULT 0,
		option_count INTEGER NOT NULL DEFAULT 0,
		options_json TEXT,
		chosen_option TEXT,
		is_correct INTEGER NOT NULL DEFAULT 0,
		quality INTEGER NOT NULL DEFAULT 0,
		metrics_json TEXT,
		srs_before_json TEXT,
		srs_after_json TEXT
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	return db
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
	today := time.Now().Format("2006-01-02")
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
