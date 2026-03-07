package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestNudgeRepository_CreateNudge(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	// Create user first
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	nudgeRepo := NewNudgeRepository(conn, logger)
	localDate := time.Now().Format("2006-01-02")

	nudge := &models.TrainingNudge{
		UserID:          user.ID,
		LocalDate:       localDate,
		DueCountAtSend:  5,
	}

	_, err := nudgeRepo.CreateNudge(nudge)
	if err != nil {
		t.Fatalf("CreateNudge() error = %v", err)
	}
}

func TestNudgeRepository_HasNudgeToday(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	nudgeRepo := NewNudgeRepository(conn, logger)
	localDate := time.Now().Format("2006-01-02")

	// Initially no nudge
	hasNudge, err := nudgeRepo.HasNudgeToday(user.ID, localDate)
	if err != nil {
		t.Fatalf("HasNudgeToday() error = %v", err)
	}
	if hasNudge {
		t.Error("Expected no nudge today initially")
	}

	// Create nudge
	nudge := &models.TrainingNudge{
		UserID:          user.ID,
		LocalDate:       localDate,
		DueCountAtSend:  5,
	}
	_, _ = nudgeRepo.CreateNudge(nudge)

	// Should have nudge now
	hasNudge, _ = nudgeRepo.HasNudgeToday(user.ID, localDate)
	if !hasNudge {
		t.Error("Expected to have nudge today")
	}
}

func TestNudgeRepository_ConsumeNudge(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	nudgeRepo := NewNudgeRepository(conn, logger)
	localDate := time.Now().Format("2006-01-02")

	// Create nudge
	nudge := &models.TrainingNudge{
		UserID:          user.ID,
		LocalDate:       localDate,
		DueCountAtSend:  5,
	}
	_, _ = nudgeRepo.CreateNudge(nudge)

	// Consume nudge
	err := nudgeRepo.ConsumeNudge(user.ID, localDate)
	if err != nil {
		t.Fatalf("ConsumeNudge() error = %v", err)
	}
}

func TestNudgeRepository_GetUnconsumedNudge(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	nudgeRepo := NewNudgeRepository(conn, logger)
	localDate := time.Now().Format("2006-01-02")

	// Initially no unconsumed nudge
	nudge, err := nudgeRepo.GetUnconsumedNudge(user.ID, localDate)
	if err != nil {
		t.Fatalf("GetUnconsumedNudge() error = %v", err)
	}
	if nudge != nil {
		t.Error("Expected no unconsumed nudge initially")
	}

	// Create nudge
	newNudge := &models.TrainingNudge{
		UserID:          user.ID,
		LocalDate:       localDate,
		DueCountAtSend:  5,
	}
	_, _ = nudgeRepo.CreateNudge(newNudge)

	// Should find unconsumed nudge
	nudge, _ = nudgeRepo.GetUnconsumedNudge(user.ID, localDate)
	if nudge == nil {
		t.Error("Expected to find unconsumed nudge")
	}
}

func TestNudgeRepository_GetUnconsumedNudge_WithMessageID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(67890)
	nudgeRepo := NewNudgeRepository(conn, logger)
	localDate := time.Now().Format("2006-01-02")

	nudge := &models.TrainingNudge{
		UserID:         user.ID,
		LocalDate:      localDate,
		DueCountAtSend: 3,
	}
	id, err := nudgeRepo.CreateNudge(nudge)
	if err != nil {
		t.Fatalf("CreateNudge() error = %v", err)
	}

	// Set message_id via raw update (simulating telegram message id)
	_, err = conn.Exec(`UPDATE training_nudges SET message_id = $1 WHERE id = $2`, 12345, id)
	if err != nil {
		t.Fatalf("UPDATE message_id: %v", err)
	}

	found, err := nudgeRepo.GetUnconsumedNudge(user.ID, localDate)
	if err != nil {
		t.Fatalf("GetUnconsumedNudge() error = %v", err)
	}
	if found == nil {
		t.Fatal("Expected to find unconsumed nudge")
	}
	if found.MessageID == nil {
		t.Error("MessageID should be set when stored in DB")
	}
	if *found.MessageID != 12345 {
		t.Errorf("Expected MessageID 12345, got %d", *found.MessageID)
	}
}

func TestNewNudgeRepository(t *testing.T) {
	logger := zap.NewNop()
	conn := testutil.SetupTestDB(t)
	repo := NewNudgeRepository(conn, logger)
	if repo == nil {
		t.Fatal("NewNudgeRepository() returned nil")
	}
	if repo.db != conn {
		t.Error("repo.db should be the passed connection")
	}
	if repo.logger != logger {
		t.Error("repo.logger should be the passed logger")
	}
}

func TestNudgeRepository_CreateNudge_WithMessageID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(99999)
	nudgeRepo := NewNudgeRepository(conn, logger)
	localDate := time.Now().Format("2006-01-02")
	msgID := 42
	nudge := &models.TrainingNudge{
		UserID:          user.ID,
		LocalDate:       localDate,
		DueCountAtSend:  1,
		MessageID:       &msgID,
	}
	id, err := nudgeRepo.CreateNudge(nudge)
	if err != nil {
		t.Fatalf("CreateNudge() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateNudge() id = %d", id)
	}
	// Verify MessageID round-trip via GetUnconsumedNudge
	found, err := nudgeRepo.GetUnconsumedNudge(user.ID, localDate)
	if err != nil {
		t.Fatalf("GetUnconsumedNudge() error = %v", err)
	}
	if found == nil || found.MessageID == nil || *found.MessageID != 42 {
		t.Errorf("expected MessageID 42, got %v", found)
	}
}

func TestNudgeRepository_CreateNudge_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewNudgeRepository(db, zap.NewNop())
	_, err = repo.CreateNudge(&models.TrainingNudge{UserID: 1, LocalDate: "2025-01-01", DueCountAtSend: 0})
	if err == nil {
		t.Error("CreateNudge() expected error with invalid DSN")
	}
}

func TestNudgeRepository_HasNudgeToday_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewNudgeRepository(db, zap.NewNop())
	_, err = repo.HasNudgeToday(1, "2025-01-01")
	if err == nil {
		t.Error("HasNudgeToday() expected error with invalid DSN")
	}
}

func TestNudgeRepository_ConsumeNudge_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewNudgeRepository(db, zap.NewNop())
	err = repo.ConsumeNudge(1, "2025-01-01")
	if err == nil {
		t.Error("ConsumeNudge() expected error with invalid DSN")
	}
}

func TestNudgeRepository_GetUnconsumedNudge_Error(t *testing.T) {
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", "postgres://x:x@invalid.invalid:1/db?connect_timeout=1")
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	defer db.Close()
	repo := NewNudgeRepository(db, zap.NewNop())
	_, err = repo.GetUnconsumedNudge(1, "2025-01-01")
	if err == nil {
		t.Error("GetUnconsumedNudge() expected error with invalid DSN")
	}
}
