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

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(111)

	userCardRepo := NewUserCardRepository(db, logger)
	// Create word_card, training_card, user_card for review event
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "evword", "def")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, 'evword', 0, 'слово', 'word')")
	now := time.Now()
	uc := &models.UserCard{UserID: user.ID, TrainingCardID: 1, Direction: models.DirectionENtoRU, State: models.StateReview, EF: 2.0, NextDueAt: &now}
	ucID, err := userCardRepo.CreateUserCard(uc)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	repo := NewSessionRepository(db, logger)

	// Create a session first
	session := &models.TrainingSession{
		UserID:       user.ID,
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
	event := &models.ReviewEvent{
		SessionID:    &sessionID,
		UserID:       user.ID,
		UserCardID:   ucID,
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

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(222)

	userCardRepo := NewUserCardRepository(db, logger)
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "statword", "def")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (1, 'statword', 0, 'слово', 'word')")
	now := time.Now()
	uc1 := &models.UserCard{UserID: user.ID, TrainingCardID: 1, Direction: models.DirectionENtoRU, State: models.StateReview, EF: 2.0, NextDueAt: &now}
	uc2 := &models.UserCard{UserID: user.ID, TrainingCardID: 1, Direction: models.DirectionRUtoEN, State: models.StateReview, EF: 2.0, NextDueAt: &now}
	uc1ID, _ := userCardRepo.CreateUserCard(uc1)
	uc2ID, _ := userCardRepo.CreateUserCard(uc2)

	repo := NewSessionRepository(db, logger)

	// Create a session
	session := &models.TrainingSession{
		UserID:       user.ID,
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
	event1 := &models.ReviewEvent{
		SessionID:  &sessionID,
		UserID:     user.ID,
		UserCardID: uc1ID,
		Direction:  models.DirectionENtoRU,
		ShownAt:    now,
		AnsweredAt: &now,
		IsCorrect:  true,
		Quality:    2,
	}
	event2 := &models.ReviewEvent{
		SessionID:  &sessionID,
		UserID:     user.ID,
		UserCardID: uc2ID,
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

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(333)

	repo := NewSessionRepository(db, logger)

	// Create sessions for today
	today := time.Now().UTC().Format("2006-01-02")
	for i := 0; i < 3; i++ {
		session := &models.TrainingSession{
			UserID:       user.ID,
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
	count, err := repo.GetTodaySessionCount(user.ID, today)
	if err != nil {
		t.Fatalf("GetTodaySessionCount() error = %v", err)
	}
	if count < 3 {
		t.Errorf("Expected at least 3 sessions, got %d", count)
	}
}

func TestSessionRepository_GetTrainingStreak(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionAdditionalTestDB(t)

	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(444)

	repo := NewSessionRepository(db, logger)

	// Insert sessions on specific dates (UTC). Use "as of" date 2025-02-24 so yesterday = 2025-02-23.
	// Sessions: 2025-02-21, 2025-02-22, 2025-02-23 (three consecutive days ending yesterday)
	for _, dateStr := range []string{"2025-02-21 12:00:00+00", "2025-02-22 12:00:00+00", "2025-02-23 14:00:00+00"} {
		_, err := db.Exec(`INSERT INTO training_sessions (user_id, source, planned_count, done_count, session_json, started_at)
			VALUES ($1, $2, $3, $4, $5, $6::timestamptz)`,
			user.ID, "manual", 5, 0, "{}", dateStr)
		if err != nil {
			t.Fatalf("insert session: %v", err)
		}
	}

	streak, trainedYesterday, err := repo.GetTrainingStreak(user.ID, "UTC", "2025-02-24")
	if err != nil {
		t.Fatalf("GetTrainingStreak() error = %v", err)
	}
	if streak != 3 {
		t.Errorf("Expected streak 3, got %d", streak)
	}
	if !trainedYesterday {
		t.Error("Expected trainedYesterday true")
	}

	// No session on 2025-02-23: streak 0, trainedYesterday false
	_, _ = db.Exec("DELETE FROM training_sessions WHERE user_id = $1", user.ID)
	_, _ = db.Exec(`INSERT INTO training_sessions (user_id, source, planned_count, done_count, session_json, started_at)
		VALUES ($1, $2, $3, $4, $5, $6::timestamptz)`,
		user.ID, "manual", 5, 0, "{}", "2025-02-22 12:00:00+00")
	_, _ = db.Exec(`INSERT INTO training_sessions (user_id, source, planned_count, done_count, session_json, started_at)
		VALUES ($1, $2, $3, $4, $5, $6::timestamptz)`,
		user.ID, "manual", 5, 0, "{}", "2025-02-21 12:00:00+00")

	streak2, trainedYesterday2, err := repo.GetTrainingStreak(user.ID, "UTC", "2025-02-24")
	if err != nil {
		t.Fatalf("GetTrainingStreak() second call error = %v", err)
	}
	if streak2 != 0 {
		t.Errorf("Expected streak 0 when yesterday missing, got %d", streak2)
	}
	if trainedYesterday2 {
		t.Error("Expected trainedYesterday false when no session on 2025-02-23")
	}
}

func TestSessionRepository_GetTrainingStreak_InvalidTimezone(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionAdditionalTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(445)
	repo := NewSessionRepository(db, logger)

	// Invalid timezone falls back to UTC
	streak, trainedYesterday, err := repo.GetTrainingStreak(user.ID, "Invalid/Timezone", "2025-02-24")
	if err != nil {
		t.Fatalf("GetTrainingStreak(invalid TZ) should fallback to UTC: %v", err)
	}
	if streak != 0 || trainedYesterday {
		t.Errorf("expected zero streak when no sessions: streak=%d trainedYesterday=%v", streak, trainedYesterday)
	}
}

func TestSessionRepository_GetTrainingStreak_InvalidDate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupSessionAdditionalTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(446)
	repo := NewSessionRepository(db, logger)

	_, _, err := repo.GetTrainingStreak(user.ID, "UTC", "not-a-date")
	if err == nil {
		t.Fatal("GetTrainingStreak(invalid date) expected error")
	}
}
