package service

import (
	"database/sql"
	"fmt"
	"strings"
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

// seedSessionWithReviewEvent creates user, word_card, training_card, user_card, session and one review_event linked to the session.
func seedSessionWithReviewEvent(t *testing.T, db *sql.DB, sessionRepo *repository.SessionRepository) (sessionID int64, userID, wordCardID int64) {
	t.Helper()
	var uid int64
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (90002) RETURNING id`).Scan(&uid); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	var wcid int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ('finishword', 'def') RETURNING id`).Scan(&wcid); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	var tcid int64
	if err := db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES ($1, 'finishword', 0, 'тест', 'meaning') RETURNING id`, wcid).Scan(&tcid); err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	var ucid int64
	if err := db.QueryRow(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'en_ru', 'review') RETURNING id`, uid, tcid).Scan(&ucid); err != nil {
		t.Fatalf("insert user_card: %v", err)
	}
	sess := &models.TrainingSession{UserID: uid, Source: models.SourceManual, PlannedCount: 1, DoneCount: 0, SessionJSON: "{}"}
	sid, err := sessionRepo.CreateSession(sess)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO review_events (session_id, user_id, user_card_id, direction, answered_at, is_correct) VALUES ($1, $2, $3, 'en_ru', CURRENT_TIMESTAMP, 1)`, sid, uid, ucid); err != nil {
		t.Fatalf("insert review_event: %v", err)
	}
	return sid, uid, wcid
}

func TestTrainingService_FinishSession_WithMasteringRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	sessionID, userID, wordCardID := seedSessionWithReviewEvent(t, db, sessionRepo)
	masteringRepo := repository.NewUserWordMasteringRepository(db, logger)
	service := NewTrainingService(nil, nil, sessionRepo, masteringRepo, logger)

	err := service.FinishSession(sessionID, 1)
	if err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}
	finished, err := service.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if finished.EndedAt == nil {
		t.Error("Session should have ended_at set after FinishSession")
	}
	// updateMasteringScoresForSession should have run: pairs from GetWordCardIDsBySessionID, then UpsertBatch
	var score int
	if err := db.QueryRow(`SELECT mastering_score FROM user_word_mastering WHERE user_id = $1 AND word_card_id = $2`, userID, wordCardID).Scan(&score); err != nil {
		t.Fatalf("expected mastering row after FinishSession: %v", err)
	}
	if score < 0 || score > 100 {
		t.Errorf("mastering_score should be 0-100, got %d", score)
	}
}

func TestTrainingService_FinishSession_WithMasteringRepo_NoReviewEvents(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(90003)
	sess := &models.TrainingSession{UserID: user.ID, Source: models.SourceManual, PlannedCount: 0, DoneCount: 0, SessionJSON: "{}"}
	sessionID, err := sessionRepo.CreateSession(sess)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	masteringRepo := repository.NewUserWordMasteringRepository(db, logger)
	service := NewTrainingService(nil, nil, sessionRepo, masteringRepo, logger)

	err = service.FinishSession(sessionID, 0)
	if err != nil {
		t.Fatalf("FinishSession() error = %v", err)
	}
	// updateMasteringScoresForSession: GetWordCardIDsBySessionID returns [], so we return early and never call UpsertBatch
	finished, err := service.GetSession(sessionID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if finished.EndedAt == nil {
		t.Error("Session should have ended_at set after FinishSession")
	}
}

// TestTrainingService_FinishSession_NonExistentSession verifies that finishing a non-existent session
// does not return an error (UPDATE affects 0 rows but Exec succeeds).
func TestTrainingService_FinishSession_NonExistentSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, _, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	service := NewTrainingService(nil, nil, sessionRepo, nil, logger)

	err := service.FinishSession(999999, 2)
	if err != nil {
		t.Errorf("FinishSession(non-existent id) should not error (UPDATE succeeds), got: %v", err)
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

func TestTrainingService_RestoreQueue_EmptyInput(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, _, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)

	queue, err := service.RestoreQueue(1, nil)
	if err != nil {
		t.Fatalf("RestoreQueue(nil) error = %v", err)
	}
	if queue != nil {
		t.Errorf("Expected nil queue for nil input, got len %d", len(queue))
	}

	queue, err = service.RestoreQueue(1, []int64{})
	if err != nil {
		t.Fatalf("RestoreQueue(empty) error = %v", err)
	}
	if queue != nil {
		t.Errorf("Expected nil queue for empty input, got len %d", len(queue))
	}
}

func TestTrainingService_RestoreQueue_PartialSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(889)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "partial", "def")
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "partial",
		SenseIndex: 0,
		WordRU:     "частичный",
		MeaningEN:  "partial",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
	}
	validID, _ := userCardRepo.CreateUserCard(userCard)

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	// One valid ID, one non-existent
	queue, err := service.RestoreQueue(user.ID, []int64{validID, 999999})
	if err != nil {
		t.Fatalf("RestoreQueue error = %v", err)
	}
	if len(queue) != 1 {
		t.Errorf("Expected 1 card in queue (one invalid id skipped), got %d", len(queue))
	}
}

func TestTrainingService_RestoreQueue_WrongUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user1, _ := userRepo.GetOrCreateUser(890)
	user2, _ := userRepo.GetOrCreateUser(891)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "wronguser", "def")
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "wronguser",
		SenseIndex: 0,
		WordRU:     "другой",
		MeaningEN:  "wrong user",
	}
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(card)
	userCard := &models.UserCard{
		UserID:         user1.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
	}
	userCardID, _ := userCardRepo.CreateUserCard(userCard)

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	// Restore as user2 - card belongs to user1, so it should be skipped
	queue, err := service.RestoreQueue(user2.ID, []int64{userCardID})
	if err != nil {
		t.Fatalf("RestoreQueue error = %v", err)
	}
	if len(queue) != 0 {
		t.Errorf("Expected 0 cards when requesting as different user, got %d", len(queue))
	}
}

func TestTrainingService_UpdateSessionState_SessionNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, _, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	service := NewTrainingService(nil, nil, sessionRepo, nil, logger)

	err := service.UpdateSessionState(999999, `{"state":1}`)
	if err == nil {
		t.Fatal("expected error when session not found")
	}
	if !strings.Contains(err.Error(), "session not found") {
		t.Errorf("expected session not found error, got: %v", err)
	}
}
