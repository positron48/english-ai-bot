package service

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
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

	service := NewTrainingService(userCardRepo, nil, nil, nil, config.DefaultLearningConfig(), logger)
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

	service := NewTrainingService(nil, nil, sessionRepo, nil, config.DefaultLearningConfig(), logger)
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

	service := NewTrainingService(nil, nil, sessionRepo, nil, config.DefaultLearningConfig(), logger)
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

	service := NewTrainingService(nil, nil, sessionRepo, nil, config.DefaultLearningConfig(), logger)
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
	service := NewTrainingService(nil, nil, sessionRepo, masteringRepo, config.DefaultLearningConfig(), logger)

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
	service := NewTrainingService(nil, nil, sessionRepo, masteringRepo, config.DefaultLearningConfig(), logger)

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
	service := NewTrainingService(nil, nil, sessionRepo, nil, config.DefaultLearningConfig(), logger)

	err := service.FinishSession(999999, 2)
	if err != nil {
		t.Errorf("FinishSession(non-existent id) should not error (UPDATE succeeds), got: %v", err)
	}
}

// mockMasteringRepoForSession returns error from GetWordCardIDsBySessionID or GetWordMasteringStatsBatch to trigger FinishSession/updateMasteringScoresForSession paths.
type mockMasteringRepoForSession struct {
	getWordCardIDsBySessionIDFunc  func(sessionID int64) ([]repository.UserWordPair, error)
	getWordMasteringStatsBatchFunc func(pairs []repository.UserWordPair) (map[repository.UserWordPair]repository.WordMasteringStatsRow, error)
	getKnownForPairsFunc           func(pairs []repository.UserWordPair) (map[repository.UserWordPair]bool, error)
	upsertBatchFunc                func(entries []struct {
		UserID, WordCardID int64
		Score              int
	}) error
	getScoreFunc func(userID, wordCardID int64) (int, error)
}

func (m *mockMasteringRepoForSession) GetWordCardIDsBySessionID(sessionID int64) ([]repository.UserWordPair, error) {
	if m.getWordCardIDsBySessionIDFunc != nil {
		return m.getWordCardIDsBySessionIDFunc(sessionID)
	}
	return nil, nil
}

func (m *mockMasteringRepoForSession) GetWordMasteringStatsBatch(pairs []repository.UserWordPair) (map[repository.UserWordPair]repository.WordMasteringStatsRow, error) {
	if m.getWordMasteringStatsBatchFunc != nil {
		return m.getWordMasteringStatsBatchFunc(pairs)
	}
	return nil, nil
}

func (m *mockMasteringRepoForSession) GetKnownForPairs(pairs []repository.UserWordPair) (map[repository.UserWordPair]bool, error) {
	if m.getKnownForPairsFunc != nil {
		return m.getKnownForPairsFunc(pairs)
	}
	return nil, nil
}

func (m *mockMasteringRepoForSession) UpsertBatch(entries []struct {
	UserID, WordCardID int64
	Score              int
}) error {
	if m.upsertBatchFunc != nil {
		return m.upsertBatchFunc(entries)
	}
	return nil
}

func (m *mockMasteringRepoForSession) GetScore(userID, wordCardID int64) (int, error) {
	if m.getScoreFunc != nil {
		return m.getScoreFunc(userID, wordCardID)
	}
	return 0, nil
}

// TestTrainingService_FinishSession_MasteringUpdateFails verifies that when updateMasteringScoresForSession fails, FinishSession still returns nil and logs Warn.
func TestTrainingService_FinishSession_MasteringUpdateFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(777)
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 2,
		DoneCount:    1,
		SessionJSON:  "{}",
	}
	sessionID, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	mockMastering := &mockMasteringRepoForSession{
		getWordCardIDsBySessionIDFunc: func(_ int64) ([]repository.UserWordPair, error) {
			return nil, fmt.Errorf("mock db error")
		},
	}
	service := NewTrainingService(nil, nil, sessionRepo, mockMastering, config.DefaultLearningConfig(), logger)

	err = service.FinishSession(sessionID, 1)
	if err != nil {
		t.Errorf("FinishSession should return nil when only mastering update fails (Warn path), got: %v", err)
	}
}

// TestTrainingService_FinishSession_MasteringStatsBatchFails verifies that when updateMasteringScoresForSession fails
// due to GetWordMasteringStatsBatch error, FinishSession still returns nil and logs Warn.
func TestTrainingService_FinishSession_MasteringStatsBatchFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(778)
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 2,
		DoneCount:    1,
		SessionJSON:  "{}",
	}
	sessionID, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	mockMastering := &mockMasteringRepoForSession{
		getWordCardIDsBySessionIDFunc: func(_ int64) ([]repository.UserWordPair, error) {
			return []repository.UserWordPair{{UserID: user.ID, WordCardID: 1}}, nil
		},
		getWordMasteringStatsBatchFunc: func(_ []repository.UserWordPair) (map[repository.UserWordPair]repository.WordMasteringStatsRow, error) {
			return nil, fmt.Errorf("mock stats batch error")
		},
	}
	service := NewTrainingService(nil, nil, sessionRepo, mockMastering, config.DefaultLearningConfig(), logger)

	err = service.FinishSession(sessionID, 1)
	if err != nil {
		t.Errorf("FinishSession should return nil when mastering stats batch fails (Warn path), got: %v", err)
	}
	// Session should still be finished
	sess, _ := service.GetSession(sessionID)
	if sess != nil && sess.EndedAt == nil {
		t.Error("session should have ended_at set after FinishSession")
	}
}

// TestTrainingService_FinishSession_GetKnownForPairsFails verifies that when updateMasteringScoresForSession fails
// due to GetKnownForPairs error, FinishSession still returns nil and logs Warn.
func TestTrainingService_FinishSession_GetKnownForPairsFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(779)
	session := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 2,
		DoneCount:    1,
		SessionJSON:  "{}",
	}
	sessionID, err := sessionRepo.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	mockMastering := &mockMasteringRepoForSession{
		getWordCardIDsBySessionIDFunc: func(_ int64) ([]repository.UserWordPair, error) {
			return []repository.UserWordPair{{UserID: user.ID, WordCardID: 1}}, nil
		},
		getWordMasteringStatsBatchFunc: func(_ []repository.UserWordPair) (map[repository.UserWordPair]repository.WordMasteringStatsRow, error) {
			return map[repository.UserWordPair]repository.WordMasteringStatsRow{
				{UserID: user.ID, WordCardID: 1}: {UserID: user.ID, WordCardID: 1, Total: 10, Correct: 8, RecentTotal: 5, RecentCorrect: 4},
			}, nil
		},
		getKnownForPairsFunc: func(_ []repository.UserWordPair) (map[repository.UserWordPair]bool, error) {
			return nil, fmt.Errorf("mock get known error")
		},
	}
	service := NewTrainingService(nil, nil, sessionRepo, mockMastering, config.DefaultLearningConfig(), logger)

	err = service.FinishSession(sessionID, 1)
	if err != nil {
		t.Errorf("FinishSession should return nil when GetKnownForPairs fails (Warn path), got: %v", err)
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

	service := NewTrainingService(nil, nil, sessionRepo, nil, config.DefaultLearningConfig(), logger)
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

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, config.DefaultLearningConfig(), logger)
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
	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, config.DefaultLearningConfig(), logger)

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

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, config.DefaultLearningConfig(), logger)
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

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, config.DefaultLearningConfig(), logger)
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
	service := NewTrainingService(nil, nil, sessionRepo, nil, config.DefaultLearningConfig(), logger)

	err := service.UpdateSessionState(999999, `{"state":1}`)
	if err == nil {
		t.Fatal("expected error when session not found")
	}
	if !strings.Contains(err.Error(), "session not found") {
		t.Errorf("expected session not found error, got: %v", err)
	}
}

// TestTrainingService_StartSession_Success_NoActiveSession starts a session when user has cards and no active session.
func TestTrainingService_StartSession_Success_NoActiveSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(7001)

	_, _ = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (1, 'start1', 'def1')")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, 'start1', 0, 'старт', 'def1', 'noun', 'start1')")
	past := time.Now().Add(-24 * time.Hour)
	_, err := userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	})
	if err != nil {
		t.Fatalf("create user card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, config.DefaultLearningConfig(), logger)
	config := &SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   5,
		AlgoVersion:        "test",
	}
	session, queue, err := service.StartSession(user.ID, models.SourceManual, config)
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}
	if session == nil {
		t.Fatal("StartSession() session should not be nil")
	}
	if session.ID == 0 {
		t.Error("StartSession() session ID should be set")
	}
	if session.UserID != user.ID {
		t.Errorf("session UserID = %d, want %d", session.UserID, user.ID)
	}
	if len(queue) == 0 {
		t.Error("StartSession() queue should not be empty")
	}
	if session.PlannedCount != len(queue) {
		t.Errorf("PlannedCount = %d, want %d", session.PlannedCount, len(queue))
	}
}

// TestTrainingService_StartSession_Success_NilConfig uses default config when config is nil.
func TestTrainingService_StartSession_Success_NilConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(7002)

	_, _ = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (1, 'nilcfg', 'def')")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, 'nilcfg', 0, 'нил', 'def', 'noun', 'nilcfg')")
	past := time.Now().Add(-24 * time.Hour)
	_, _ = userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	})

	service := NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, config.DefaultLearningConfig(), logger)
	session, queue, err := service.StartSession(user.ID, models.SourceManual, nil)
	if err != nil {
		t.Fatalf("StartSession(nil config) error = %v", err)
	}
	if session == nil || len(queue) == 0 {
		t.Fatal("StartSession with nil config should succeed and return session and queue")
	}
}

// TestTrainingService_StartSession_Success_FinishesOldSession finishes existing active session then starts new one.
func TestTrainingService_StartSession_Success_FinishesOldSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(7003)

	_, _ = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (1, 'oldnew', 'def')")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, 'oldnew', 0, 'старый', 'def', 'noun', 'oldnew')")
	past := time.Now().Add(-24 * time.Hour)
	_, _ = userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	})

	// Create an active session first
	oldSession := &models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 2,
		DoneCount:    0,
		SessionJSON:  "{}",
	}
	oldID, err := sessionRepo.CreateSession(oldSession)
	if err != nil {
		t.Fatalf("create old session: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, config.DefaultLearningConfig(), logger)
	session, queue, err := service.StartSession(user.ID, models.SourceManual, &SessionConfig{MaxCardsPerSession: 10, MaxNewPerSession: 5, AlgoVersion: "test"})
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}
	if session == nil {
		t.Fatal("StartSession() session should not be nil")
	}
	if session.ID == oldID {
		t.Error("StartSession() should create new session, not reuse old ID")
	}
	if len(queue) == 0 {
		t.Error("StartSession() queue should not be empty")
	}
	// Old session should be finished
	old, _ := sessionRepo.GetSession(oldID)
	if old != nil && old.EndedAt == nil {
		t.Error("old session should have ended_at set after StartSession")
	}
}

// TestTrainingService_StartSession_NoCardsAvailable returns error when user has no cards for training.
func TestTrainingService_StartSession_NoCardsAvailable(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, userCardRepo, trainingCardRepo, sessionRepo := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(7004)
	// No user_cards created

	service := NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, config.DefaultLearningConfig(), logger)
	_, _, err := service.StartSession(user.ID, models.SourceManual, &SessionConfig{MaxCardsPerSession: 10, MaxNewPerSession: 5, AlgoVersion: "test"})
	if err == nil {
		t.Fatal("StartSession() expected error when no cards available")
	}
	if !strings.Contains(err.Error(), "no cards available") {
		t.Errorf("error should mention no cards available, got: %v", err)
	}
}

// TestTrainingService_StartSession_GetActiveSessionFails returns error when GetActiveSession fails (e.g. DB unavailable).
func TestTrainingService_StartSession_GetActiveSessionFails(t *testing.T) {
	testutil.SetupTestDB(t) // ensure DB is up and postgres_compat driver is registered
	dsn := testutil.GetTestDSN(t)
	closedConn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	closedConn.Close()
	logger, _ := zap.NewDevelopment()
	sessionRepo := repository.NewSessionRepository(closedConn, logger)
	service := NewTrainingService(nil, nil, sessionRepo, nil, config.DefaultLearningConfig(), logger)

	_, _, err = service.StartSession(7005, models.SourceManual, nil)
	if err == nil {
		t.Fatal("StartSession() expected error when GetActiveSession fails")
	}
	if !strings.Contains(err.Error(), "failed to check active session") {
		t.Errorf("error should mention failed to check active session, got: %v", err)
	}
}
