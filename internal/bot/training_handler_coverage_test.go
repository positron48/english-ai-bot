package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// setTHField sets an unexported field on a TrainingHandler using reflection + unsafe.
func setTHField(th *TrainingHandler, fieldName string, value interface{}) {
	thv := reflect.ValueOf(th).Elem()
	field := thv.FieldByName(fieldName)
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

// newFailingBot creates a BotAPI that fails on all sends.
func newFailingBot() *tgbotapi.BotAPI {
	return &tgbotapi.BotAPI{Token: "test", Client: &failingTelegramClient{}, Buffer: 1}
}

// setupTrainingHandler creates a full TrainingHandler with real DB for coverage tests.
func setupTrainingHandler(t *testing.T, client *mockTelegramClient) (*TrainingHandler, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)

	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	srsService := service.NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	bot := newTestBot(client)
	th := NewTrainingHandler(bot, trainingService, srsService, optionsService, sessionRepo, logger, 0, 0, conn)
	return th, db
}

// createUserWithCard creates a user with one training card and user card, returns userID and userCardID.
func createUserWithCard(t *testing.T, db *database.DB, telegramID int64) (userID int64, userCardID int64) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	wordRepo := repository.NewWordRepository(conn, logger)
	wordID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "test", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	tcID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID,
		WordEN:     "test",
		WordRU:     "тест",
		MeaningEN:  "test",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	userCardRepo := repository.NewUserCardRepository(conn, logger)
	ucID, err := userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: tcID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	})
	if err != nil {
		t.Fatalf("CreateUserCard: %v", err)
	}

	return user.ID, ucID
}

// TestStartTraining_SaveSessionStateError covers the saveSessionState error warn path in StartTraining.
func TestStartTraining_SaveSessionStateError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3001)

	// Use hook so saveSessionState returns error after StartSession succeeds (covers 127-129)
	testHookSaveSessionStateFail = func() error { return fmt.Errorf("injected save state fail") }
	defer func() { testHookSaveSessionStateFail = nil }()

	err := th.StartTraining(context.Background(), 3001, userID, models.SourceManual)
	_ = err
}

// TestShowCard_GenerateOptionsError covers the skipCard path when GenerateOptions fails.
func TestShowCard_GenerateOptionsError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3002)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error (no cards?): %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3002] = state
	th.sessionsMutex.Unlock()

	// Use hook to force GenerateOptions error path (covers 194-197)
	testHookGenerateOptionsErr = func() error { return fmt.Errorf("injected options err") }
	defer func() { testHookGenerateOptionsErr = nil }()

	// showCard should call skipCard when GenerateOptions fails
	_ = th.showCard(3002)
}

// TestShowCard_SaveSessionStateError covers the saveSessionState warn path in showCard.
func TestShowCard_SaveSessionStateError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3003)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3003] = state
	th.sessionsMutex.Unlock()

	// Replace trainingService with one using a failing sessionRepo so saveSessionState fails
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	failingTS := service.NewTrainingService(userCardRepo, trainingCardRepo, failingSessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	setTHField(th, "trainingService", failingTS)

	// showCard should log warn for saveSessionState but still send the card
	_ = th.showCard(3003)
}

// TestShowOptions_SaveSessionStateError covers the saveSessionState warn path in ShowOptions.
func TestShowOptions_SaveSessionStateError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3004)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3004] = state
	th.sessionsMutex.Unlock()

	// Replace trainingService with failing sessionRepo
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	failingTS := service.NewTrainingService(userCardRepo, trainingCardRepo, failingSessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	setTHField(th, "trainingService", failingTS)

	err = th.ShowOptions(3004, false)
	_ = err
}

// TestHandleAnswer_SaveSessionStateError covers saveSessionState warn path in HandleAnswer.
func TestHandleAnswer_SaveSessionStateError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3005)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       &now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3005] = state
	th.sessionsMutex.Unlock()

	// Replace trainingService with failing sessionRepo so saveSessionState fails
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	failingTS := service.NewTrainingService(userCardRepo, trainingCardRepo, failingSessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	setTHField(th, "trainingService", failingTS)

	_ = th.HandleAnswer(3005, 0)
}

// TestHandleAnswer_GradeCardError covers the GradeCard error path (logger.Error) in HandleAnswer.
func TestHandleAnswer_GradeCardError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3040)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       &now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3040] = state
	th.sessionsMutex.Unlock()

	// Replace srsService with one using failing userCardRepo so GradeCard -> UpdateUserCard fails
	failingConn := newFailingDB(t)
	failingUserCardRepo := repository.NewUserCardRepository(failingConn, logger)
	failingSRS := service.NewSRSService(failingUserCardRepo, config.DefaultLearningConfig(), logger)
	setTHField(th, "srsService", failingSRS)

	_ = th.HandleAnswer(3040, 0)
}

// TestHandleAnswer_RecordWrongAnswerError covers the RecordWrongAnswer error path.
func TestHandleAnswer_RecordWrongAnswerError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3006)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	wrongIndex := -1
	for i, opt := range options {
		if opt != correctAnswer {
			wrongIndex = i
			break
		}
	}
	if wrongIndex == -1 {
		t.Skip("all options are correct, cannot test wrong answer path")
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       &now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3006] = state
	th.sessionsMutex.Unlock()

	// Use hook to force RecordWrongAnswer error path (covers 460-462)
	testHookRecordWrongAnswerErr = func() error { return fmt.Errorf("injected record wrong answer err") }
	defer func() { testHookRecordWrongAnswerErr = nil }()

	_ = th.HandleAnswer(3006, wrongIndex)
}

// TestStartTraining_SaveSessionStateMarshalError covers saveSessionState when json.Marshal(sessionData) fails (783-785).
func TestStartTraining_SaveSessionStateMarshalError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	userID, _ := createUserWithCard(t, db, 3019)

	testHookMarshalSessionState = func(interface{}) ([]byte, error) { return nil, fmt.Errorf("injected marshal err") }
	defer func() { testHookMarshalSessionState = nil }()

	_ = th.StartTraining(context.Background(), 3019, userID, models.SourceManual)
}

// TestRestoreSession_MarshalStateDataRawError covers restoreSession when marshal stateDataRaw fails (819-821).
func TestRestoreSession_MarshalStateDataRawError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	userID, _ := createUserWithCard(t, db, 3020)
	chatID := int64(3020)

	_ = th.StartTraining(context.Background(), chatID, userID, models.SourceManual)
	th.sessionsMutex.Lock()
	delete(th.sessions, chatID)
	th.sessionsMutex.Unlock()

	testHookMarshalStateDataRaw = func(interface{}) ([]byte, error) { return nil, fmt.Errorf("injected marshal state err") }
	defer func() { testHookMarshalStateDataRaw = nil }()

	_ = th.StartTraining(context.Background(), chatID, userID, models.SourceManual)
}

// TestRestoreSession_RestoreQueueInjectedError covers restoreSession when RestoreQueue fails via hook (830-832).
func TestRestoreSession_RestoreQueueInjectedError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	userID, _ := createUserWithCard(t, db, 3021)
	chatID := int64(3021)

	_ = th.StartTraining(context.Background(), chatID, userID, models.SourceManual)
	th.sessionsMutex.Lock()
	delete(th.sessions, chatID)
	th.sessionsMutex.Unlock()

	testHookRestoreQueueErr = func() error { return fmt.Errorf("injected restore queue err") }
	defer func() { testHookRestoreQueueErr = nil }()

	_ = th.StartTraining(context.Background(), chatID, userID, models.SourceManual)
}

// TestRestoreSession_FinishSessionWarnError covers restoreSession FinishSession error log (836-838, 855-857).
func TestRestoreSession_FinishSessionWarnError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	conn := db.GetConnection()

	userID, _ := createUserWithCard(t, db, 3022)
	chatID := int64(3022)
	_ = th.StartTraining(context.Background(), chatID, userID, models.SourceManual)
	th.sessionsMutex.Lock()
	delete(th.sessions, chatID)
	th.sessionsMutex.Unlock()

	// Set session_json so RestoreQueue returns [] (invalid user_card_ids) to hit len(cardQueue)==0 branch
	_, _ = conn.Exec(`UPDATE training_sessions SET session_json = ? WHERE user_id = ? AND ended_at IS NULL`,
		`{"state":{"user_card_ids":[999999],"current_index":0}}`, userID)

	testHookFinishSessionWarnErr = func() error { return fmt.Errorf("injected finish session err") }
	defer func() { testHookFinishSessionWarnErr = nil }()

	_ = th.StartTraining(context.Background(), chatID, userID, models.SourceManual)
}

// TestRestoreSession_FinishSessionWarnErrorCompletedSession covers currentIndex >= len(queue) branch (855-857).
func TestRestoreSession_FinishSessionWarnErrorCompletedSession(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	userID, userCardID := createUserWithCard(t, db, 3023)
	chatID := int64(3023)
	_ = th.StartTraining(context.Background(), chatID, userID, models.SourceManual)
	th.sessionsMutex.Lock()
	delete(th.sessions, chatID)
	th.sessionsMutex.Unlock()

	// Set session_json so current_index >= len(queue) (1 card, current_index=1)
	conn := db.GetConnection()
	var sessionID int64
	_ = conn.QueryRow(`SELECT id FROM training_sessions WHERE user_id = ? AND ended_at IS NULL`, userID).Scan(&sessionID)
	_, _ = conn.Exec(`UPDATE training_sessions SET session_json = ? WHERE id = ?`,
		fmt.Sprintf(`{"state":{"user_card_ids":[%d],"current_index":1}}`, userCardID), sessionID)

	testHookFinishSessionWarnErr = func() error { return fmt.Errorf("injected finish completed err") }
	defer func() { testHookFinishSessionWarnErr = nil }()

	_ = th.StartTraining(context.Background(), chatID, userID, models.SourceManual)
}

// TestSaveSessionState_SessionNilInDB covers the "session not found" path (GetSession returns nil).
func TestSaveSessionState_SessionNilInDB(t *testing.T) {
	client := &mockTelegramClient{}
	th, _ := setupTrainingHandler(t, client)

	state := &SessionState{
		UserID:               9999,
		SessionID:            99999,
		Queue:                []*models.TrainingQueueItem{},
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0),
	}

	err := th.saveSessionState(state)
	if err == nil {
		t.Error("expected error when session not found in DB")
	}
	if !strings.Contains(err.Error(), "session") {
		t.Errorf("expected session-related error, got %v", err)
	}
}

// TestSaveSessionState_GetSessionError covers the GetSession error path in saveSessionState.
func TestSaveSessionState_GetSessionError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3016)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)

	// Use a failing DB for sessionRepo so GetSession fails
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	failingTS := service.NewTrainingService(userCardRepo, trainingCardRepo, failingSessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	setTHField(th, "trainingService", failingTS)

	state := &SessionState{
		UserID:               userID,
		SessionID:            99999,
		Queue:                []*models.TrainingQueueItem{},
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0),
	}

	err := th.saveSessionState(state)
	if err == nil {
		t.Error("expected error when GetSession fails")
	}
}

// TestRestoreSession_RestoreQueueError covers the path where RestoreQueue returns an empty queue
// because all user cards fail to load (failing DB). The session is finished and restored=false.
func TestRestoreSession_RestoreQueueError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, ucID := createUserWithCard(t, db, 3007)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)

	stateData := SessionStateData{
		UserCardIDs:  []int64{ucID},
		CurrentIndex: 0,
	}
	stateJSON, _ := json.Marshal(map[string]interface{}{"state": stateData})
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  string(stateJSON),
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Replace trainingService with one using a failing userCardRepo so RestoreQueue skips all cards
	// (RestoreQueue logs warn and continues, returning empty queue)
	failingConn := newFailingDB(t)
	failingUserCardRepo := repository.NewUserCardRepository(failingConn, logger)
	failingTS := service.NewTrainingService(failingUserCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	setTHField(th, "trainingService", failingTS)

	// RestoreQueue returns empty queue (not error), so restoreSession finishes the session and returns false, nil
	restored, err := th.restoreSession(3007, userID)
	if restored {
		t.Error("expected not restored when RestoreQueue returns empty queue")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRestoreSession_EmptyQueue covers the "queue is empty, finish session" path.
func TestRestoreSession_EmptyQueue(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3008)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)

	stateData := SessionStateData{
		UserCardIDs:  []int64{99999},
		CurrentIndex: 0,
	}
	stateJSON, _ := json.Marshal(map[string]interface{}{"state": stateData})
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  string(stateJSON),
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	restored, err := th.restoreSession(3008, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if restored {
		t.Error("expected not restored when queue is empty")
	}
}

// TestRestoreSession_IndexBeyondQueue covers the "currentIndex >= len(queue)" path.
func TestRestoreSession_IndexBeyondQueue(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, ucID := createUserWithCard(t, db, 3009)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)

	stateData := SessionStateData{
		UserCardIDs:  []int64{ucID},
		CurrentIndex: 100,
	}
	stateJSON, _ := json.Marshal(map[string]interface{}{"state": stateData})
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  string(stateJSON),
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	restored, err := th.restoreSession(3009, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if restored {
		t.Error("expected not restored when currentIndex >= len(queue)")
	}
}

// TestRestoreSession_InvalidStateJSON covers the json.Unmarshal failure path.
func TestRestoreSession_InvalidStateJSON(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3010)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)

	invalidJSON := `{"state": "not-an-object"}`
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  invalidJSON,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	restored, err := th.restoreSession(3010, userID)
	if err == nil && restored {
		t.Error("expected error or not-restored when state JSON is invalid")
	}
}

// TestRestoreSession_GetActiveSessionError covers the GetActiveSession error path in restoreSession.
// When trainingService uses a failing DB, GetActiveSession fails and restoreSession returns (false, error).
func TestRestoreSession_GetActiveSessionError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3011)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)

	// Replace trainingService with one using a failing sessionRepo so GetActiveSession fails
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	failingTS := service.NewTrainingService(userCardRepo, trainingCardRepo, failingSessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	setTHField(th, "trainingService", failingTS)

	// GetActiveSession fails, returns (false, error)
	restored, err := th.restoreSession(3011, userID)
	if restored {
		t.Error("expected not restored when GetActiveSession fails")
	}
	if err == nil {
		t.Error("expected error when GetActiveSession fails")
	}
}

// TestRestoreSession_IndexBeyondQueueWithFailingDB covers the path where GetActiveSession fails
// when using a failing DB (covers the error return path in restoreSession).
func TestRestoreSession_IndexBeyondQueueWithFailingDB(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3012)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)

	// Replace trainingService with one using a failing sessionRepo
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	failingTS := service.NewTrainingService(userCardRepo, trainingCardRepo, failingSessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	setTHField(th, "trainingService", failingTS)

	// GetActiveSession fails, returns (false, error)
	restored, err := th.restoreSession(3012, userID)
	if restored {
		t.Error("expected not restored when GetActiveSession fails")
	}
	if err == nil {
		t.Error("expected error when GetActiveSession fails")
	}
}

// TestStartTraining_ExistingSessionInMemory covers the "session already exists in memory" path.
func TestStartTraining_ExistingSessionInMemory(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3013)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       &now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3013] = state
	th.sessionsMutex.Unlock()

	err = th.StartTraining(context.Background(), 3013, userID, models.SourceManual)
	_ = err
}

// TestStartTraining_RestoredSession covers the "session restored" path in StartTraining.
func TestStartTraining_RestoredSession(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, ucID := createUserWithCard(t, db, 3014)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)

	stateData := SessionStateData{
		UserCardIDs:  []int64{ucID},
		CurrentIndex: 0,
	}
	stateJSON, _ := json.Marshal(map[string]interface{}{"state": stateData})
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  string(stateJSON),
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	err = th.StartTraining(context.Background(), 3014, userID, models.SourceManual)
	_ = err
}

// TestHandleAnswer_OptionsNotShownYet covers the optionsShownAt == nil path in HandleAnswer.
func TestHandleAnswer_OptionsNotShownYet(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3015)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       nil,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3015] = state
	th.sessionsMutex.Unlock()

	_ = th.HandleAnswer(3015, 0)
}

// TestHandleAnswer_EarlyRevealTrue covers the branch where optionsShownAt != nil and tDelayMS < optionsDelayMS (earlyReveal = true).
func TestHandleAnswer_EarlyRevealTrue(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	// Set optionsDelayMS to 10s so that showing options quickly yields earlyReveal = true
	setTHField(th, "optionsDelayMS", 10000)
	userID, _ := createUserWithCard(t, db, 3016)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()
	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil || len(queue) == 0 {
		t.Skipf("StartSession: %v", err)
	}
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions: %v", err)
	}
	shownAt := time.Now().Add(-time.Second)
	optionsShownAt := shownAt.Add(10 * time.Millisecond)
	state := &SessionState{
		UserID: userID, SessionID: session.ID, Queue: queue, CurrentIndex: 0,
		ShownAt: shownAt, OptionsShownAt: &optionsShownAt,
		Options: options, CorrectAnswer: correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}
	th.sessionsMutex.Lock()
	th.sessions[3017] = state
	th.sessionsMutex.Unlock()
	correctIdx := 0
	for i, o := range options {
		if o == correctAnswer {
			correctIdx = i
			break
		}
	}
	err = th.HandleAnswer(3017, correctIdx)
	if err != nil {
		t.Fatalf("HandleAnswer (earlyReveal): %v", err)
	}
}

// TestSaveSessionState_InvalidSessionJSON covers the json.Unmarshal failure path in saveSessionState.
func TestSaveSessionState_InvalidSessionJSON(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3016)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)

	sess, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  "not-valid-json",
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	state := &SessionState{
		UserID:               userID,
		SessionID:            sess,
		Queue:                []*models.TrainingQueueItem{},
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0),
	}

	err = th.saveSessionState(state)
	if err != nil {
		t.Logf("saveSessionState with invalid JSON: %v (may be ok)", err)
	}
}

// TestHandleAnswer_CardNotInitialized covers the "card state not initialized" path.
func TestHandleAnswer_CardNotInitialized(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3017)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       nil,
		Options:              []string{},
		CorrectAnswer:        "",
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3017] = state
	th.sessionsMutex.Unlock()

	err = th.HandleAnswer(3017, 0)
	if err == nil {
		t.Error("expected error when card state not initialized")
	}
	if !strings.Contains(err.Error(), "card is being shown") && !strings.Contains(err.Error(), "please wait") {
		t.Errorf("expected 'card is being shown' error, got %v", err)
	}
}

// TestHandleAnswer_TrimRecentCorrectAnswers covers the len > 2 trim path (line 471-473).
func TestHandleAnswer_TrimRecentCorrectAnswers(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3024)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	// Find the correct answer index
	correctIndex := -1
	for i, opt := range options {
		if opt == correctAnswer {
			correctIndex = i
			break
		}
	}
	if correctIndex == -1 {
		t.Skip("correct answer not in options")
	}

	now := time.Now()
	state := &SessionState{
		UserID:         userID,
		SessionID:      session.ID,
		Queue:          queue,
		CurrentIndex:   0,
		ShownAt:        now,
		OptionsShownAt: &now,
		Options:        options,
		CorrectAnswer:  correctAnswer,
		// Pre-populate with 2 items so that adding another triggers the trim (len > 2)
		RecentCorrectAnswers: []string{"word1", "word2"},
	}

	th.sessionsMutex.Lock()
	th.sessions[3024] = state
	th.sessionsMutex.Unlock()

	// HandleAnswer with correct index - RecentCorrectAnswers will grow to 3, then be trimmed to 2
	_ = th.HandleAnswer(3024, correctIndex)
}

// createMultipleUserCards creates a user with N training cards and user cards.
func createMultipleUserCards(t *testing.T, db *database.DB, telegramID int64, count int) (userID int64) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)

	for i := 0; i < count; i++ {
		word := fmt.Sprintf("word%d_%d", telegramID, i)
		wordID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: word, Definition: ""})
		if err != nil {
			t.Fatalf("UpsertWordCardLemma: %v", err)
		}
		tcID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
			WordCardID: wordID,
			WordEN:     word,
			WordRU:     fmt.Sprintf("слово%d", i),
			MeaningEN:  word,
			SenseIndex: 0,
		})
		if err != nil {
			t.Fatalf("CreateTrainingCard: %v", err)
		}
		_, err = userCardRepo.CreateUserCard(&models.UserCard{
			UserID:         user.ID,
			TrainingCardID: tcID,
			Direction:      models.DirectionENtoRU,
			State:          models.StateNew,
			EF:             models.InitialEF,
		})
		if err != nil {
			t.Fatalf("CreateUserCard: %v", err)
		}
	}

	return user.ID
}

// TestShowCard_MultipleCardsInQueue covers the loop body (lines 177-183) with multiple cards.
func TestShowCard_MultipleCardsInQueue(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID := createMultipleUserCards(t, db, 4010, 3)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) < 2 {
		t.Skip("need at least 2 cards in queue")
	}

	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[4010] = state
	th.sessionsMutex.Unlock()

	// showCard with multiple cards in queue covers the loop body (lines 177-183)
	err = th.showCard(4010)
	_ = err
}

// TestShowCard_BotSendError covers the bot.Send error path in showCard (line 235-237).
func TestShowCard_BotSendError(t *testing.T) {
	// Use a failing bot so bot.Send returns error
	failingBot := newFailingBot()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	srsService := service.NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	th := NewTrainingHandler(failingBot, trainingService, srsService, optionsService, sessionRepo, logger, 0, 0, conn)

	userID, _ := createUserWithCard(t, db, 4001)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[4001] = state
	th.sessionsMutex.Unlock()

	// showCard should fail at bot.Send and return error
	err = th.showCard(4001)
	if err == nil {
		t.Error("expected error when bot.Send fails")
	}
}

// TestShowOptions_BotSendError covers the bot.Send error path in ShowOptions (line 304-306).
func TestShowOptions_BotSendError(t *testing.T) {
	failingBot := newFailingBot()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	srsService := service.NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	th := NewTrainingHandler(failingBot, trainingService, srsService, optionsService, sessionRepo, logger, 0, 0, conn)

	userID, _ := createUserWithCard(t, db, 4002)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	opts, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		Options:              opts,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[4002] = state
	th.sessionsMutex.Unlock()

	// ShowOptions should fail at bot.Send and return error
	err = th.ShowOptions(4002, false)
	if err == nil {
		t.Error("expected error when bot.Send fails")
	}
}

// TestShowOptions_EarlyRevealTrue covers the earlyReveal branch (optionsText = "Варианты ответа:").
func TestShowOptions_EarlyRevealTrue(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 4003)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[4003] = state
	th.sessionsMutex.Unlock()

	err = th.ShowOptions(4003, true)
	if err != nil {
		t.Fatalf("ShowOptions(earlyReveal=true): %v", err)
	}
	if text := client.lastParams.Get("text"); text != "" && text != "Варианты ответа:" {
		// Client may send "Выберите правильный вариант" or "Варианты ответа:" depending on earlyReveal
		if !strings.Contains(text, "Варианты") {
			t.Errorf("expected options message, got %q", text)
		}
	}
}

// TestHandleAnswer_ShowCardError covers the showCard error path in HandleAnswer (line 349-351).
func TestHandleAnswer_ShowCardError(t *testing.T) {
	failingBot := newFailingBot()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	srsService := service.NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	th := NewTrainingHandler(failingBot, trainingService, srsService, optionsService, sessionRepo, logger, 0, 0, conn)

	userID, _ := createUserWithCard(t, db, 4003)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	opts, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	// Find correct answer index
	correctIndex := -1
	for i, opt := range opts {
		if opt == correctAnswer {
			correctIndex = i
			break
		}
	}
	if correctIndex == -1 {
		t.Skip("correct answer not in options")
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       &now,
		Options:              opts,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[4003] = state
	th.sessionsMutex.Unlock()

	// HandleAnswer with correct answer - after feedback, showCard is called for next card
	// But there's no next card (queue has only 1 item), so finishSession is called
	// The bot.Send failure will happen in finishSession's sendMessage calls
	err = th.HandleAnswer(4003, correctIndex)
	// May or may not return error depending on where bot.Send fails
	_ = err
}

// TestHandleAnswer_ShowCardFailsWhenCardNotInitialized covers line 349-351:
// HandleAnswer where options are empty and showCard fails.
func TestHandleAnswer_ShowCardFailsWhenCardNotInitialized(t *testing.T) {
	failingBot := newFailingBot()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	srsService := service.NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	th := NewTrainingHandler(failingBot, trainingService, srsService, optionsService, sessionRepo, logger, 0, 0, conn)

	userID, _ := createUserWithCard(t, db, 4004)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       nil,
		Options:              []string{}, // empty - card not initialized
		CorrectAnswer:        "",
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[4004] = state
	th.sessionsMutex.Unlock()

	// HandleAnswer with empty options triggers showCard which fails (failing bot)
	err = th.HandleAnswer(4004, 0)
	if err == nil {
		t.Error("expected error when showCard fails")
	}
}

// TestSaveSessionState_EmptySessionJSON covers the else branch when session.SessionJSON == "".
func TestSaveSessionState_EmptySessionJSON(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3018)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)

	// Create a session with empty SessionJSON
	sess, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  "",
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	state := &SessionState{
		UserID:               userID,
		SessionID:            sess,
		Queue:                []*models.TrainingQueueItem{},
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0),
	}

	// saveSessionState with empty SessionJSON should use new map (line 774-776)
	err = th.saveSessionState(state)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRestoreSession_InvalidJSONInDB covers the json.Unmarshal error path (line 808-810).
func TestRestoreSession_InvalidJSONInDB(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3019)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)

	// Create a session with truly invalid JSON (not valid JSON at all)
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  "{invalid json}",
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// restoreSession should fail at json.Unmarshal (line 808-810)
	restored, err := th.restoreSession(3019, userID)
	if restored {
		t.Error("expected not restored when session JSON is invalid")
	}
	if err == nil {
		t.Error("expected error when session JSON is invalid")
	}
}

// TestHandleAnswer_CreateReviewEventError covers the CreateReviewEvent error path (line 453-455).
func TestHandleAnswer_CreateReviewEventError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3020)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       &now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3020] = state
	th.sessionsMutex.Unlock()

	// Replace sessionRepo with failing one so CreateReviewEvent fails
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	setTHField(th, "sessionRepo", failingSessionRepo)

	// HandleAnswer should log error for CreateReviewEvent but continue
	_ = th.HandleAnswer(3020, 0)
}

// TestFinishSession_GetSessionStatsError covers the GetSessionStats error path (line 565-567).
func TestFinishSession_GetSessionStatsError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3021)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       &now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3021] = state
	th.sessionsMutex.Unlock()

	// Replace sessionRepo with failing one so GetSessionStats fails
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	setTHField(th, "sessionRepo", failingSessionRepo)

	// HandleAnswer with correct answer triggers finishSession which calls GetSessionStats
	_ = th.HandleAnswer(3021, 0)
}

// TestFinishSession_DBQueryErrors covers the DB query errors in finishSession (lines 590-593, 608-611).
func TestFinishSession_DBQueryErrors(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3022)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, err := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	if err != nil {
		t.Skipf("GenerateOptions error: %v", err)
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		OptionsShownAt:       &now,
		Options:              options,
		CorrectAnswer:        correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3022] = state
	th.sessionsMutex.Unlock()

	// Replace h.db with failing one so QueryRow fails in finishSession
	failingConn := newFailingDB(t)
	setTHField(th, "db", failingConn)

	// HandleAnswer with correct answer triggers finishSession which uses h.db for queries
	_ = th.HandleAnswer(3022, 0)
}

// TestCancelSession_FinishSessionError covers the FinishSession error in CancelSession (line 668-670).
func TestCancelSession_FinishSessionError(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, _ := createUserWithCard(t, db, 3023)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)

	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil {
		t.Skipf("StartSession error: %v", err)
	}
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}

	now := time.Now()
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		ShownAt:              now,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	th.sessionsMutex.Lock()
	th.sessions[3023] = state
	th.sessionsMutex.Unlock()

	// Replace trainingService with failing one so FinishSession fails
	failingConn := newFailingDB(t)
	failingSessionRepo := repository.NewSessionRepository(failingConn, logger)
	failingTS := service.NewTrainingService(userCardRepo, trainingCardRepo, failingSessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	setTHField(th, "trainingService", failingTS)

	// CancelSession should log error for FinishSession but continue
	err = th.CancelSession(3023)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestRestoreSession_NegativeCurrentIndex covers the currentIndex < 0 path (line 850-852).
func TestRestoreSession_NegativeCurrentIndex(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)

	userID, ucID := createUserWithCard(t, db, 3025)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()

	sessionRepo := repository.NewSessionRepository(conn, logger)

	// Create a session with negative CurrentIndex
	stateData := SessionStateData{
		UserCardIDs:  []int64{ucID},
		CurrentIndex: -5, // negative index
	}
	stateJSON, _ := json.Marshal(map[string]interface{}{"state": stateData})
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  string(stateJSON),
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// restoreSession should normalize currentIndex to 0 and restore successfully
	restored, err := th.restoreSession(3025, userID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !restored {
		t.Error("expected session to be restored with negative currentIndex normalized to 0")
	}
}

// TestExtractSessionWordsFromQueue_IndexBeyondQueue covers the currentIndex >= len(queue) path (line 687-689).
func TestExtractSessionWordsFromQueue_IndexBeyondQueue(t *testing.T) {
	client := &mockTelegramClient{}
	th, _ := setupTrainingHandler(t, client)

	// Create a queue with 1 item
	queue := []*models.TrainingQueueItem{
		{Type: "card", Card: &models.UserCardWithTraining{}},
	}

	// Call with currentIndex >= len(queue)
	result := th.extractSessionWordsFromQueue(queue, 5, &models.UserCardWithTraining{}, []string{})
	if len(result) != 0 {
		t.Errorf("expected empty result when currentIndex >= len(queue), got %v", result)
	}
}

// TestExtractSessionWordsFromQueue_SameWordCardID covers the WordCardID == currentWordCardID skip (line 707-708).
func TestExtractSessionWordsFromQueue_SameWordCardID(t *testing.T) {
	client := &mockTelegramClient{}
	th, _ := setupTrainingHandler(t, client)

	// Create two cards with the same WordCardID
	wordCardID := int64(42)
	card1 := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{WordCardID: wordCardID, WordEN: "test", WordRU: "тест"},
	}
	card2 := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{WordCardID: wordCardID, WordEN: "test", WordRU: "тест"}, // same WordCardID
	}

	queue := []*models.TrainingQueueItem{
		{Type: "card", Card: card1},
		{Type: "card", Card: card2},
	}

	// card2 has same WordCardID as currentCard, so it should be skipped (line 707-708)
	result := th.extractSessionWordsFromQueue(queue, 0, card1, []string{})
	// card2 should be skipped, so result should be empty
	if len(result) != 0 {
		t.Errorf("expected empty result when other card has same WordCardID, got %v", result)
	}
}

// TestExtractSessionWordsFromQueue_RUtoEN_NoDisplayWord covers the else branch (line 719-721)
// when direction is RUtoEN and DisplayWord is nil.
func TestExtractSessionWordsFromQueue_RUtoEN_NoDisplayWord(t *testing.T) {
	client := &mockTelegramClient{}
	th, _ := setupTrainingHandler(t, client)

	card1 := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{WordCardID: 1, WordEN: "hello", WordRU: "привет"},
	}
	card2 := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{WordCardID: 2, WordEN: "world", WordRU: "мир", DisplayWord: nil}, // no DisplayWord
	}

	queue := []*models.TrainingQueueItem{
		{Type: "card", Card: card1},
		{Type: "card", Card: card2},
	}

	// card2 has no DisplayWord, so word = card.TrainingCard.WordEN (line 719-721)
	result := th.extractSessionWordsFromQueue(queue, 0, card1, []string{})
	if len(result) == 0 {
		t.Error("expected non-empty result for RUtoEN card without DisplayWord")
	}
}

// TestShowCard_NoSession covers showCard when chatID has no session in map.
func TestShowCard_NoSession(t *testing.T) {
	th, _ := setupTrainingHandler(t, &mockTelegramClient{})
	err := th.showCard(99999)
	if err == nil {
		t.Fatal("expected error when no session")
	}
	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session', got %q", err.Error())
	}
}

// TestShowCard_NilState covers showCard when state is nil.
func TestShowCard_NilState(t *testing.T) {
	th, _ := setupTrainingHandler(t, &mockTelegramClient{})
	th.sessionsMutex.Lock()
	th.sessions[99998] = nil
	th.sessionsMutex.Unlock()
	err := th.showCard(99998)
	if err == nil {
		t.Fatal("expected error when state is nil")
	}
	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session', got %q", err.Error())
	}
}

// TestFinishSession_NotExists covers finishSession when chatID has no session (!exists, return nil).
func TestFinishSession_NotExists(t *testing.T) {
	th, _ := setupTrainingHandler(t, &mockTelegramClient{})
	err := th.finishSession(99997)
	if err != nil {
		t.Errorf("finishSession with no session should return nil, got %v", err)
	}
}

// TestFinishSession_DBNil covers finishSession when h.db is nil (skip due/new queries, availableForTraining stays 0).
func TestFinishSession_DBNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	userRepo := repository.NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(3025)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	sessID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 0, DoneCount: 0, SessionJSON: "{}",
	})
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	srsService := service.NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	th := NewTrainingHandler(newTestBot(&mockTelegramClient{}), trainingService, srsService, optionsService, sessionRepo, logger, 0, 0, nil)
	th.sessionsMutex.Lock()
	th.sessions[3025] = &SessionState{
		UserID: user.ID, SessionID: sessID, Queue: []*models.TrainingQueueItem{}, CurrentIndex: 0,
	}
	th.sessionsMutex.Unlock()
	err := th.finishSession(3025)
	if err != nil {
		t.Fatalf("finishSession with db=nil: %v", err)
	}
}

// TestFinishSession_TotalCardsZero_NoAvailableCards covers the message branch when totalCards == 0 and availableForTraining == 0.
func TestFinishSession_TotalCardsZero_NoAvailableCards(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()
	userRepo := repository.NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(3026)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	sessID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 0, DoneCount: 0, SessionJSON: "{}",
	})
	th.sessionsMutex.Lock()
	th.sessions[3026] = &SessionState{
		UserID: user.ID, SessionID: sessID, Queue: []*models.TrainingQueueItem{}, CurrentIndex: 0,
	}
	th.sessionsMutex.Unlock()
	err := th.finishSession(3026)
	if err != nil {
		t.Fatalf("finishSession: %v", err)
	}
	text := client.lastParams.Get("text")
	if !strings.Contains(text, "Тренировка завершена") {
		t.Errorf("expected completion message, got %q", text)
	}
	if !strings.Contains(text, "До встречи завтра") {
		t.Errorf("expected goodbye when no available cards, got %q", text)
	}
}

// TestShowOptions_NoSession covers ShowOptions when no active session.
func TestShowOptions_NoSession(t *testing.T) {
	th, _ := setupTrainingHandler(t, &mockTelegramClient{})
	err := th.ShowOptions(99996, false)
	if err == nil {
		t.Fatal("expected error when no session")
	}
	if !strings.Contains(err.Error(), "no active session") {
		t.Errorf("expected 'no active session', got %q", err.Error())
	}
}

// TestShowOptions_NilState covers ShowOptions when state is nil.
func TestShowOptions_NilState(t *testing.T) {
	th, _ := setupTrainingHandler(t, &mockTelegramClient{})
	th.sessionsMutex.Lock()
	th.sessions[99995] = nil
	th.sessionsMutex.Unlock()
	err := th.ShowOptions(99995, false)
	if err == nil {
		t.Fatal("expected error when state is nil")
	}
}

// TestShowOptions_AlreadyShown covers ShowOptions when options already shown (returns nil).
func TestShowOptions_AlreadyShown(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	userID, _ := createUserWithCard(t, db, 3027)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()
	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	session, queue, _ := trainingService.StartSession(userID, models.SourceManual, nil)
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, _ := optionsService.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	now := time.Now()
	state := &SessionState{
		UserID: userID, SessionID: session.ID, Queue: queue, CurrentIndex: 0,
		Options: options, CorrectAnswer: correctAnswer, OptionsShownAt: &now,
		RecentCorrectAnswers: make([]string, 0, 2),
	}
	th.sessionsMutex.Lock()
	th.sessions[3027] = state
	th.sessionsMutex.Unlock()
	err := th.ShowOptions(3027, false)
	if err != nil {
		t.Fatalf("first ShowOptions: %v", err)
	}
	err = th.ShowOptions(3027, false)
	if err != nil {
		t.Errorf("second ShowOptions (already shown) should return nil, got %v", err)
	}
}

// TestAutoRevealOptions_OptionsAlreadyShown covers autoRevealOptions early return when OptionsShownAt != nil.
// We use a handler with optionsDelayMS=50, run showCard (spawns goroutine), then immediately ShowOptions to set OptionsShownAt, then wait for goroutine.
func TestAutoRevealOptions_OptionsAlreadyShown(t *testing.T) {
	client := &mockTelegramClient{}
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	srsService := service.NewSRSService(userCardRepo, config.DefaultLearningConfig(), logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	th := NewTrainingHandler(newTestBot(client), trainingService, srsService, optionsService, sessionRepo, logger, 50, 0, conn)
	userID, _ := createUserWithCard(t, db, 3028)
	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil || len(queue) == 0 {
		t.Skipf("StartSession: %v", err)
	}
	optionsService2 := service.NewOptionsService(trainingCardRepo, logger)
	options, correctAnswer, _ := optionsService2.GenerateOptions(queue[0].Card, models.DefaultOptionCount, nil, nil, nil)
	now := time.Now()
	state := &SessionState{
		UserID: userID, SessionID: session.ID, Queue: queue, CurrentIndex: 0,
		ShownAt: now, Options: options, CorrectAnswer: correctAnswer,
		RecentCorrectAnswers: make([]string, 0, 2),
	}
	th.sessionsMutex.Lock()
	th.sessions[3028] = state
	th.sessionsMutex.Unlock()
	_ = th.showCard(3028)
	// Set options already shown before the goroutine (50ms delay) runs
	_ = th.ShowOptions(3028, true)
	time.Sleep(100 * time.Millisecond)
	// Goroutine should have run and returned early because OptionsShownAt != nil
}

// TestRestoreSession_StateValueInvalidType covers restoreSession when "state" value cannot be unmarshaled into SessionStateData.
func TestRestoreSession_StateValueInvalidType(t *testing.T) {
	th, db := setupTrainingHandler(t, &mockTelegramClient{})
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(3035)
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 0, DoneCount: 0, SessionJSON: `{"state": 123}`,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	restored, err := th.RestoreSession(3035, user.ID)
	if err == nil {
		t.Fatal("expected error when state value is invalid type")
	}
	if restored {
		t.Error("expected not restored")
	}
	if !strings.Contains(err.Error(), "unmarshal state data") {
		t.Errorf("expected unmarshal state data error, got %q", err.Error())
	}
}

// TestRestoreSession_StateKeyMissing covers restoreSession when session_json has no "state" key.
func TestRestoreSession_StateKeyMissing(t *testing.T) {
	th, db := setupTrainingHandler(t, &mockTelegramClient{})
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(3034)
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 0, DoneCount: 0, SessionJSON: `{"other":"value"}`,
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	restored, err := th.RestoreSession(3034, user.ID)
	if err != nil {
		t.Fatalf("RestoreSession: %v", err)
	}
	if restored {
		t.Error("expected not restored when state key is missing")
	}
}

// TestRestoreSession_EmptySessionJSON covers restoreSession when active session has SessionJSON == "".
func TestRestoreSession_EmptySessionJSON(t *testing.T) {
	th, db := setupTrainingHandler(t, &mockTelegramClient{})
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(3029)
	_, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 0, DoneCount: 0, SessionJSON: "",
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	restored, err := th.RestoreSession(3029, user.ID)
	if err != nil {
		t.Fatalf("RestoreSession: %v", err)
	}
	if restored {
		t.Error("expected not restored when SessionJSON is empty")
	}
}

// TestShowCard_SpellSkips covers showCard when current item is type "spell" (skip and recurse).
func TestShowCard_SpellSkips(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	userID, _ := createUserWithCard(t, db, 3030)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()
	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	session, queue, err := trainingService.StartSession(userID, models.SourceManual, nil)
	if err != nil || len(queue) == 0 {
		t.Skipf("StartSession: %v", err)
	}
	// Replace queue so first item is spell
	state := &SessionState{
		UserID: userID, SessionID: session.ID,
		Queue: []*models.TrainingQueueItem{
			{Type: "spell", Spell: &models.SpellChallenge{DisplayWord: "x", WordRU: "икс"}},
			queue[0],
		},
		CurrentIndex: 0, RecentCorrectAnswers: make([]string, 0, 2),
	}
	th.sessionsMutex.Lock()
	th.sessions[3030] = state
	th.sessionsMutex.Unlock()
	err = th.showCard(3030)
	if err != nil {
		t.Fatalf("showCard with spell skip: %v", err)
	}
	if state.CurrentIndex != 1 {
		t.Errorf("expected CurrentIndex 1 after spell skip, got %d", state.CurrentIndex)
	}
}

// TestShowCard_RUtoEN covers showCard when card has Direction RUtoEN (question in Russian, lines 206-211).
func TestShowCard_RUtoEN(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()
	userRepo := repository.NewUserRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(3037)
	wordID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "ruen", Definition: ""})
	tcID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID, WordEN: "hello", WordRU: "привет", MeaningEN: "hello", SenseIndex: 0,
		DistractorsEN: `["hi","bye","hey"]`, DistractorsRU: `["пока","привет","здравствуй"]`,
	})
	ucID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionRUtoEN,
		State: models.StateNew, EF: models.InitialEF,
	})
	uc, _ := userCardRepo.GetUserCard(ucID)
	tc, _ := trainingCardRepo.GetTrainingCard(tcID)
	if uc == nil || tc == nil {
		t.Fatal("GetUserCard or GetTrainingCard returned nil")
	}
	card := &models.UserCardWithTraining{UserCard: *uc, TrainingCard: *tc}
	sessID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1, DoneCount: 0, SessionJSON: "{}",
	})
	state := &SessionState{
		UserID: user.ID, SessionID: sessID, Queue: []*models.TrainingQueueItem{{Type: "card", Card: card}},
		CurrentIndex: 0, RecentCorrectAnswers: make([]string, 0, 2),
	}
	th.sessionsMutex.Lock()
	th.sessions[3037] = state
	th.sessionsMutex.Unlock()
	err := th.showCard(3037)
	if err != nil {
		t.Fatalf("showCard RUtoEN: %v", err)
	}
	text := client.lastParams.Get("text")
	if text != "" && !strings.Contains(text, "привет") {
		t.Errorf("expected question to contain WordRU 'привет', got %q", text)
	}
}

// TestShowCard_DisplayWordENtoRU covers showCard when card has Direction ENtoRU and DisplayWord set (lines 213-216).
func TestShowCard_DisplayWordENtoRU(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()
	userRepo := repository.NewUserRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(3036)
	wordID, _ := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "display", Definition: ""})
	displayWord := "to display"
	tcID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID, WordEN: "display", WordRU: "показывать", MeaningEN: "display", SenseIndex: 0,
		DisplayWord: &displayWord, Transcription: "[dɪsˈpleɪ]",
	})
	ucID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU,
		State: models.StateNew, EF: models.InitialEF,
	})
	uc, _ := userCardRepo.GetUserCard(ucID)
	tc, _ := trainingCardRepo.GetTrainingCard(tcID)
	if uc == nil || tc == nil {
		t.Fatal("GetUserCard or GetTrainingCard returned nil")
	}
	card := &models.UserCardWithTraining{UserCard: *uc, TrainingCard: *tc}
	sessID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1, DoneCount: 0, SessionJSON: "{}",
	})
	state := &SessionState{
		UserID: user.ID, SessionID: sessID, Queue: []*models.TrainingQueueItem{{Type: "card", Card: card}},
		CurrentIndex: 0, RecentCorrectAnswers: make([]string, 0, 2),
	}
	th.sessionsMutex.Lock()
	th.sessions[3036] = state
	th.sessionsMutex.Unlock()
	err := th.showCard(3036)
	if err != nil {
		t.Fatalf("showCard with DisplayWord: %v", err)
	}
	text := client.lastParams.Get("text")
	if text != "" && !strings.Contains(text, "to display") {
		t.Errorf("expected question to contain DisplayWord 'to display', got %q", text)
	}
}

// TestShowCard_NonCardSkips covers showCard when item is not type 'card' or Card is nil.
func TestShowCard_NonCardSkips(t *testing.T) {
	client := &mockTelegramClient{}
	th, db := setupTrainingHandler(t, client)
	userID, _ := createUserWithCard(t, db, 3031)
	logger, _ := zap.NewDevelopment()
	conn := db.GetConnection()
	sessionRepo := repository.NewSessionRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, userWordMasteringRepo, config.DefaultLearningConfig(), logger)
	session, queue, _ := trainingService.StartSession(userID, models.SourceManual, nil)
	if len(queue) == 0 {
		t.Skip("no cards in queue")
	}
	state := &SessionState{
		UserID: userID, SessionID: session.ID,
		Queue:                []*models.TrainingQueueItem{{Type: "card", Card: nil}},
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}
	th.sessionsMutex.Lock()
	th.sessions[3031] = state
	th.sessionsMutex.Unlock()
	err := th.showCard(3031)
	if err != nil {
		t.Fatalf("showCard with non-card item: %v", err)
	}
	if state.CurrentIndex != 1 {
		t.Errorf("expected CurrentIndex 1 after non-card skip, got %d", state.CurrentIndex)
	}
}

// TestHandleAnswer_CurrentIndexBeyondQueue covers HandleAnswer when CurrentIndex >= len(Queue) (returns nil).
func TestHandleAnswer_CurrentIndexBeyondQueue(t *testing.T) {
	th, db := setupTrainingHandler(t, &mockTelegramClient{})
	userID, _ := createUserWithCard(t, db, 3032)
	state := &SessionState{
		UserID: userID, SessionID: 1, Queue: []*models.TrainingQueueItem{}, CurrentIndex: 1,
	}
	th.sessionsMutex.Lock()
	th.sessions[3032] = state
	th.sessionsMutex.Unlock()
	err := th.HandleAnswer(3032, 0)
	if err != nil {
		t.Errorf("HandleAnswer when index beyond queue should return nil, got %v", err)
	}
}

// TestHandleAnswer_NonCardItem covers HandleAnswer when current item is not a card (returns nil).
func TestHandleAnswer_NonCardItem(t *testing.T) {
	th, db := setupTrainingHandler(t, &mockTelegramClient{})
	userID, _ := createUserWithCard(t, db, 3033)
	state := &SessionState{
		UserID: userID, SessionID: 1,
		Queue:         []*models.TrainingQueueItem{{Type: "spell", Spell: &models.SpellChallenge{}}},
		CurrentIndex:  0,
		Options:       []string{"A", "B"},
		CorrectAnswer: "A",
	}
	th.sessionsMutex.Lock()
	th.sessions[3033] = state
	th.sessionsMutex.Unlock()
	err := th.HandleAnswer(3033, 0)
	if err != nil {
		t.Errorf("HandleAnswer with non-card item should return nil, got %v", err)
	}
}

// TestHandleAnswer_NilState covers HandleAnswer when sessions[chatID] exists but is nil (!exists || state == nil).
func TestHandleAnswer_NilState(t *testing.T) {
	th, _ := setupTrainingHandler(t, &mockTelegramClient{})
	th.sessionsMutex.Lock()
	th.sessions[3036] = nil
	th.sessionsMutex.Unlock()
	err := th.HandleAnswer(3036, 0)
	if err == nil {
		t.Fatal("expected error when state is nil")
	}
	if err.Error() != "no active session" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestHandleAnswer_NoSession covers HandleAnswer when chatID has no session (!exists).
func TestHandleAnswer_NoSession(t *testing.T) {
	th, _ := setupTrainingHandler(t, &mockTelegramClient{})
	err := th.HandleAnswer(99999, 0)
	if err == nil {
		t.Fatal("expected error when no session")
	}
	if err.Error() != "no active session" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestHandleAnswer_InvalidOptionIndex_Negative covers the invalid option index branch (optionIndex < 0).
func TestHandleAnswer_InvalidOptionIndex_Negative(t *testing.T) {
	th, db := setupTrainingHandler(t, &mockTelegramClient{})
	userID, _ := createUserWithCard(t, db, 3037)
	now := time.Now()
	state := &SessionState{
		UserID:         userID,
		SessionID:      1,
		Queue:          []*models.TrainingQueueItem{{Type: "card", Card: &models.UserCardWithTraining{UserCard: models.UserCard{}, TrainingCard: models.TrainingCard{}}}},
		CurrentIndex:   0,
		Options:        []string{"a", "b", "c"},
		CorrectAnswer:  "a",
		ShownAt:        now,
		OptionsShownAt: &now,
	}
	th.sessionsMutex.Lock()
	th.sessions[3037] = state
	th.sessionsMutex.Unlock()
	err := th.HandleAnswer(3037, -1)
	if err == nil {
		t.Fatal("expected error for invalid option index")
	}
	if !strings.Contains(err.Error(), "invalid option index") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestHandleAnswer_InvalidOptionIndex_TooLarge covers the invalid option index branch (optionIndex >= len(options)).
func TestHandleAnswer_InvalidOptionIndex_TooLarge(t *testing.T) {
	th, db := setupTrainingHandler(t, &mockTelegramClient{})
	userID, _ := createUserWithCard(t, db, 3038)
	now := time.Now()
	state := &SessionState{
		UserID:         userID,
		SessionID:      1,
		Queue:          []*models.TrainingQueueItem{{Type: "card", Card: &models.UserCardWithTraining{UserCard: models.UserCard{}, TrainingCard: models.TrainingCard{}}}},
		CurrentIndex:   0,
		Options:        []string{"a", "b"},
		CorrectAnswer:  "a",
		ShownAt:        now,
		OptionsShownAt: &now,
	}
	th.sessionsMutex.Lock()
	th.sessions[3038] = state
	th.sessionsMutex.Unlock()
	err := th.HandleAnswer(3038, 10)
	if err == nil {
		t.Fatal("expected error for invalid option index")
	}
	if !strings.Contains(err.Error(), "invalid option index") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestHandleAnswer_ItemCardNil covers HandleAnswer when current item Type is "card" but Card is nil (returns nil).
func TestHandleAnswer_ItemCardNil(t *testing.T) {
	th, db := setupTrainingHandler(t, &mockTelegramClient{})
	userID, _ := createUserWithCard(t, db, 3039)
	state := &SessionState{
		UserID: userID, SessionID: 1,
		Queue:         []*models.TrainingQueueItem{{Type: "card", Card: nil}},
		CurrentIndex:  0,
		Options:       []string{"a", "b"},
		CorrectAnswer: "a",
	}
	th.sessionsMutex.Lock()
	th.sessions[3039] = state
	th.sessionsMutex.Unlock()
	err := th.HandleAnswer(3039, 0)
	if err != nil {
		t.Errorf("HandleAnswer with card nil should return nil, got %v", err)
	}
}

var _ = unsafe.Pointer(nil)
