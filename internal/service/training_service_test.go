package service

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupTrainingServiceTestDB(t *testing.T) (*sql.DB, *repository.UserCardRepository, *repository.TrainingCardRepository, *repository.SessionRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS word_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT UNIQUE NOT NULL,
		definition TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS training_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word_card_id INTEGER NOT NULL,
		word_en TEXT NOT NULL,
		transcription TEXT,
		sense_index INTEGER NOT NULL,
		word_ru TEXT NOT NULL,
		meaning_en TEXT NOT NULL,
		example_en TEXT,
		example_ru TEXT,
		distractors_ru TEXT,
		distractors_en TEXT,
		hint TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS user_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		training_card_id INTEGER NOT NULL,
		direction TEXT NOT NULL,
		state TEXT NOT NULL,
		ef REAL NOT NULL DEFAULT 2.5,
		reps INTEGER NOT NULL DEFAULT 0,
		interval_days INTEGER NOT NULL DEFAULT 0,
		learning_step INTEGER NOT NULL DEFAULT 0,
		lapse_count INTEGER NOT NULL DEFAULT 0,
		next_due_at TEXT,
		last_review_at TEXT,
		last_quality INTEGER,
		last_options_json TEXT,
		wrong_answers_json TEXT,
		stats_json TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS training_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ended_at TEXT,
		source TEXT NOT NULL,
		planned_count INTEGER NOT NULL DEFAULT 0,
		done_count INTEGER NOT NULL DEFAULT 0,
		session_json TEXT DEFAULT ''
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	userCardRepo := repository.NewUserCardRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)

	return db, userCardRepo, trainingCardRepo, sessionRepo
}

func TestTrainingService_GetDueCount(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userCardRepo, _, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create a training card
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, ?, ?, ?, ?)",
		1, "test", 0, "тест", "test")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create due cards
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	for i := 0; i < 2; i++ {
		card := &models.UserCard{
			UserID:         333,
			TrainingCardID: 1,
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

	service := NewTrainingService(userCardRepo, nil, nil, logger)
	count, err := service.GetDueCount(333)
	if err != nil {
		t.Fatalf("GetDueCount() error = %v", err)
	}
	if count < 2 {
		t.Errorf("Expected at least 2 due cards, got %d", count)
	}
}

func TestTrainingService_GetSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create a session
	session := &models.TrainingSession{
		UserID:       444,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service := NewTrainingService(nil, nil, sessionRepo, logger)
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
	db, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create an active session
	session := &models.TrainingSession{
		UserID:       555,
		Source:       models.SourceManual,
		PlannedCount: 3,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	_, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service := NewTrainingService(nil, nil, sessionRepo, logger)
	active, err := service.GetActiveSession(555)
	if err != nil {
		t.Fatalf("GetActiveSession() error = %v", err)
	}
	if active == nil {
		t.Fatal("GetActiveSession() should not return nil")
	}
	if active.UserID != 555 {
		t.Errorf("Expected UserID 555, got %d", active.UserID)
	}
}

func TestTrainingService_FinishSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create a session
	session := &models.TrainingSession{
		UserID:       666,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service := NewTrainingService(nil, nil, sessionRepo, logger)
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
	db, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	defer db.Close()

	// Create a session
	session := &models.TrainingSession{
		UserID:       777,
		Source:       models.SourceManual,
		PlannedCount: 5,
		DoneCount:    0,
		SessionJSON:  `{}`,
	}
	id, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	service := NewTrainingService(nil, nil, sessionRepo, logger)
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
	db, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	defer db.Close()

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
		UserID:         888,
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
		UserID:         888,
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

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, logger)
	queue, err := service.RestoreQueue(888, []int64{userCardID1, userCardID2})
	if err != nil {
		t.Fatalf("RestoreQueue() error = %v", err)
	}
	if len(queue) != 2 {
		t.Errorf("Expected 2 cards in queue, got %d", len(queue))
	}
}
