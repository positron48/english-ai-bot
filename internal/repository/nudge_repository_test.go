package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupNudgeTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS training_nudges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		local_date TEXT NOT NULL,
		sent_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		consumed_at TEXT,
		due_count_at_send INTEGER NOT NULL DEFAULT 0,
		message_id INTEGER
	)`

	_, err = db.Exec(createTable)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	return db
}

func TestNewNudgeRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupNudgeTestDB(t)
	defer db.Close()

	repo := NewNudgeRepository(db, logger)
	_ = repo // Verify repository is created
}

func TestNudgeRepository_CreateNudge(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupNudgeTestDB(t)
	defer db.Close()

	repo := NewNudgeRepository(db, logger)

	msgID := 123
	nudge := &models.TrainingNudge{
		UserID:         456,
		LocalDate:       "2024-01-01",
		DueCountAtSend: 5,
		MessageID:      &msgID,
	}

	id, err := repo.CreateNudge(nudge)
	if err != nil {
		t.Fatalf("CreateNudge() error = %v", err)
	}
	if id == 0 {
		t.Error("CreateNudge() should return non-zero ID")
	}
}

func TestNudgeRepository_HasNudgeToday(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupNudgeTestDB(t)
	defer db.Close()

	repo := NewNudgeRepository(db, logger)

	// Check before creating nudge
	has, err := repo.HasNudgeToday(789, "2024-01-01")
	if err != nil {
		t.Fatalf("HasNudgeToday() error = %v", err)
	}
	if has {
		t.Error("HasNudgeToday() should return false before creating nudge")
	}

	// Create a nudge
	nudge := &models.TrainingNudge{
		UserID:         789,
		LocalDate:       "2024-01-01",
		DueCountAtSend: 3,
	}
	_, err = repo.CreateNudge(nudge)
	if err != nil {
		t.Fatalf("Failed to create nudge: %v", err)
	}

	// Check after creating nudge
	has, err = repo.HasNudgeToday(789, "2024-01-01")
	if err != nil {
		t.Fatalf("HasNudgeToday() error = %v", err)
	}
	if !has {
		t.Error("HasNudgeToday() should return true after creating nudge")
	}
}

func TestNudgeRepository_ConsumeNudge(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupNudgeTestDB(t)
	defer db.Close()

	repo := NewNudgeRepository(db, logger)

	// Create a nudge
	nudge := &models.TrainingNudge{
		UserID:         111,
		LocalDate:       "2024-01-02",
		DueCountAtSend: 2,
	}
	_, err := repo.CreateNudge(nudge)
	if err != nil {
		t.Fatalf("Failed to create nudge: %v", err)
	}

	// Consume the nudge
	err = repo.ConsumeNudge(111, "2024-01-02")
	if err != nil {
		t.Fatalf("ConsumeNudge() error = %v", err)
	}

	// Verify it's consumed
	unconsumed, err := repo.GetUnconsumedNudge(111, "2024-01-02")
	if err != nil {
		t.Fatalf("GetUnconsumedNudge() error = %v", err)
	}
	if unconsumed != nil {
		t.Error("GetUnconsumedNudge() should return nil after consuming")
	}
}

func TestNudgeRepository_GetUnconsumedNudge(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupNudgeTestDB(t)
	defer db.Close()

	repo := NewNudgeRepository(db, logger)

	// Create a nudge
	nudge := &models.TrainingNudge{
		UserID:         222,
		LocalDate:       "2024-01-03",
		DueCountAtSend: 4,
	}
	_, err := repo.CreateNudge(nudge)
	if err != nil {
		t.Fatalf("Failed to create nudge: %v", err)
	}

	// Get unconsumed nudge
	found, err := repo.GetUnconsumedNudge(222, "2024-01-03")
	if err != nil {
		t.Fatalf("GetUnconsumedNudge() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetUnconsumedNudge() should not return nil")
	}
	if found.UserID != 222 {
		t.Errorf("Expected UserID 222, got %d", found.UserID)
	}
	if found.DueCountAtSend != 4 {
		t.Errorf("Expected DueCountAtSend 4, got %d", found.DueCountAtSend)
	}
}
