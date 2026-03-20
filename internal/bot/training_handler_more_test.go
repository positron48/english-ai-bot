package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

type trainingHandlerDeps struct {
	handler          *TrainingHandler
	db               *database.DB
	userRepo         *repository.UserRepository
	trainingCardRepo *repository.TrainingCardRepository
	userCardRepo     *repository.UserCardRepository
	sessionRepo      *repository.SessionRepository
}

func strPtr(s string) *string {
	return &s
}

func setupTrainingHandlerDeps(t *testing.T, client *mockTelegramClient) *trainingHandlerDeps {
	t.Helper()

	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)

	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	handler := NewTrainingHandler(
		newTestBot(client),
		trainingService,
		srsService,
		optionsService,
		sessionRepo,
		logger,
		100000,
		0,
		db.GetConnection(),
	)

	return &trainingHandlerDeps{
		handler:          handler,
		db:               db,
		userRepo:         userRepo,
		trainingCardRepo: trainingCardRepo,
		userCardRepo:     userCardRepo,
		sessionRepo:      sessionRepo,
	}
}

func seedTrainableUserCard(t *testing.T, deps *trainingHandlerDeps, telegramID int64) (userID int64, userCardID int64) {
	t.Helper()

	user, err := deps.userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	spell := false
	typing := false
	settings := models.UserSettings{
		SpellModeEnabled: &spell,
		TypeModeEnabled:  &typing,
	}
	settingsJSON, _ := json.Marshal(settings)
	if err := deps.userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		t.Fatalf("UpdateUserSettings failed: %v", err)
	}

	var wordCardID int64
	if err := deps.db.GetConnection().QueryRow(
		"INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id",
		"coverage",
		"coverage",
	).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word card failed: %v", err)
	}

	trainingCardID, err := deps.trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "coverage",
		SenseIndex:    0,
		WordRU:        "покрытие",
		MeaningEN:     "coverage",
		ExampleEN:     "Coverage matters.",
		ExampleRU:     "Покрытие важно.",
		Hint:          "Think about tests",
		DistractorsRU: `["пример","код","анализ"]`,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard failed: %v", err)
	}

	dueAt := time.Now().Add(-2 * time.Hour)
	userCardID, err = deps.userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &dueAt,
	})
	if err != nil {
		t.Fatalf("CreateUserCard failed: %v", err)
	}

	return user.ID, userCardID
}

func TestTrainingHandler_StartTraining_NoCards(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)

	user, err := deps.userRepo.GetOrCreateUser(910001)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	err = deps.handler.StartTraining(context.Background(), 1001, user.ID, models.SourceManual)
	if err == nil {
		t.Fatal("expected StartTraining to fail when no cards are available")
	}
	if deps.handler.HasActiveSession(1001) {
		t.Fatal("did not expect active session after failed start")
	}
}

func TestTrainingHandler_StartTraining_HandleAnswer_FinishFlow(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)
	userID, _ := seedTrainableUserCard(t, deps, 910002)

	const chatID int64 = 1002
	if err := deps.handler.StartTraining(context.Background(), chatID, userID, models.SourceManual); err != nil {
		t.Fatalf("StartTraining failed: %v", err)
	}
	if !deps.handler.HasActiveSession(chatID) {
		t.Fatal("expected active session after StartTraining")
	}

	if err := deps.handler.ShowOptions(chatID, false); err != nil {
		t.Fatalf("ShowOptions failed: %v", err)
	}

	deps.handler.sessionsMutex.RLock()
	state := deps.handler.sessions[chatID]
	deps.handler.sessionsMutex.RUnlock()
	if state == nil || len(state.Options) == 0 {
		t.Fatal("expected options to be generated for active card")
	}

	correctIdx := -1
	for i, option := range state.Options {
		if option == state.CorrectAnswer {
			correctIdx = i
			break
		}
	}
	if correctIdx == -1 {
		t.Fatalf("could not locate correct option in %#v", state.Options)
	}

	if err := deps.handler.HandleAnswer(chatID, correctIdx); err != nil {
		t.Fatalf("HandleAnswer failed: %v", err)
	}

	if deps.handler.HasActiveSession(chatID) {
		t.Fatal("expected session to finish after answering single card")
	}
	if !strings.Contains(client.lastParams.Get("text"), "Тренировка завершена") {
		t.Fatalf("expected completion message, got %q", client.lastParams.Get("text"))
	}
}

func TestTrainingHandler_HandleAnswer_InvalidOptionIndex(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)
	userID, _ := seedTrainableUserCard(t, deps, 910003)

	const chatID int64 = 1003
	if err := deps.handler.StartTraining(context.Background(), chatID, userID, models.SourceManual); err != nil {
		t.Fatalf("StartTraining failed: %v", err)
	}
	if err := deps.handler.ShowOptions(chatID, false); err != nil {
		t.Fatalf("ShowOptions failed: %v", err)
	}

	err := deps.handler.HandleAnswer(chatID, 999)
	if err == nil || !strings.Contains(err.Error(), "invalid option index") {
		t.Fatalf("expected invalid option index error, got %v", err)
	}
}

func TestTrainingHandler_ShowOptionsAndCancelErrors(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)

	if err := deps.handler.ShowOptions(1111, false); err == nil {
		t.Fatal("expected ShowOptions error when no active session")
	}
	if err := deps.handler.CancelSession(1111); err == nil {
		t.Fatal("expected CancelSession error when no active session")
	}
}

func TestTrainingHandler_CancelSession_Active(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)
	userID, _ := seedTrainableUserCard(t, deps, 910004)

	const chatID int64 = 1004
	if err := deps.handler.StartTraining(context.Background(), chatID, userID, models.SourceManual); err != nil {
		t.Fatalf("StartTraining failed: %v", err)
	}

	if err := deps.handler.CancelSession(chatID); err != nil {
		t.Fatalf("CancelSession failed: %v", err)
	}
	if deps.handler.HasActiveSession(chatID) {
		t.Fatal("expected no active session after cancel")
	}
	if !strings.Contains(client.lastParams.Get("text"), "Тренировка отменена") {
		t.Fatalf("expected cancel message, got %q", client.lastParams.Get("text"))
	}
}

func TestTrainingHandler_SaveAndRestoreSessionState(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)
	userID, userCardID := seedTrainableUserCard(t, deps, 910005)

	sessionID, err := deps.sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  `{}`,
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	state := &SessionState{
		UserID:       userID,
		SessionID:    sessionID,
		CurrentIndex: 0,
		Queue: []*models.TrainingQueueItem{{
			Type: "card",
			Card: &models.UserCardWithTraining{
				UserCard: models.UserCard{ID: userCardID},
			},
		}},
	}

	if err := deps.handler.saveSessionState(state); err != nil {
		t.Fatalf("saveSessionState failed: %v", err)
	}

	stored, err := deps.handler.trainingService.GetSession(sessionID)
	if err != nil || stored == nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if !strings.Contains(stored.SessionJSON, "user_card_ids") {
		t.Fatalf("expected state to be persisted in session_json, got %s", stored.SessionJSON)
	}

	restored, err := deps.handler.RestoreSession(2005, userID)
	if err != nil {
		t.Fatalf("RestoreSession failed: %v", err)
	}
	if !restored {
		t.Fatal("expected RestoreSession to return restored=true")
	}
	if !deps.handler.HasActiveSession(2005) {
		t.Fatal("expected restored session to be active in memory")
	}
}

func TestTrainingHandler_RestoreSession_NoStateData(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)
	userID, _ := seedTrainableUserCard(t, deps, 910006)

	_, err := deps.sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  `{"config":{"max_cards_per_session":20}}`,
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	restored, err := deps.handler.RestoreSession(2006, userID)
	if err != nil {
		t.Fatalf("RestoreSession returned error: %v", err)
	}
	if restored {
		t.Fatal("expected RestoreSession to return restored=false when state is missing")
	}
}

func TestTrainingHandler_extractSessionWordsFromQueue(t *testing.T) {
	verb := "verb"
	h := &TrainingHandler{}

	queue := []*models.TrainingQueueItem{
		{Type: "card", Card: &models.UserCardWithTraining{UserCard: models.UserCard{Direction: models.DirectionRUtoEN}, TrainingCard: models.TrainingCard{WordCardID: 1, POS: &verb}}},
		{Type: "card", Card: &models.UserCardWithTraining{TrainingCard: models.TrainingCard{WordCardID: 2, WordEN: "jump", DisplayWord: strPtr("to jump"), POS: &verb}}},
		{Type: "card", Card: &models.UserCardWithTraining{TrainingCard: models.TrainingCard{WordCardID: 3, WordEN: "walk", POS: strPtr("noun")}}},
		{Type: "spell", Spell: &models.SpellChallenge{DisplayWord: "ignored"}},
	}

	words := h.extractSessionWordsFromQueue(queue, 0, queue[0].Card, []string{"to jump"})
	if len(words) != 0 {
		t.Fatalf("expected recent correct answer to be excluded, got %v", words)
	}

	words = h.extractSessionWordsFromQueue(queue, 0, queue[0].Card, nil)
	if len(words) != 1 || words[0] != "to jump" {
		t.Fatalf("expected one matching word, got %v", words)
	}
}

func TestTrainingHandler_sendFeedbackAndSendMessage_NoPanic(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)

	deps.handler.sendFeedback(3001, &models.TrainingCard{WordEN: "word", WordRU: "слово", Hint: "hint"}, true, "", "")
	if !strings.Contains(client.lastParams.Get("text"), "Правильно") {
		t.Fatalf("expected positive feedback, got %q", client.lastParams.Get("text"))
	}

	deps.handler.sendFeedback(3001, &models.TrainingCard{WordEN: "word", WordRU: "слово", ExampleEN: "Example", ExampleRU: "Пример", Hint: "hint"}, false, "ошибка", "верно")
	if !strings.Contains(client.lastParams.Get("text"), "Неправильно") {
		t.Fatalf("expected negative feedback, got %q", client.lastParams.Get("text"))
	}

	deps.handler.sendMessage(3001, fmt.Sprintf("msg-%d", time.Now().UnixNano()))
	if !strings.Contains(client.lastParams.Get("text"), "msg-") {
		t.Fatalf("expected sendMessage text to be delivered, got %q", client.lastParams.Get("text"))
	}
}

func TestTrainingHandler_SkipCard(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)
	userID, _ := seedTrainableUserCard(t, deps, 910007)

	const chatID int64 = 1007
	if err := deps.handler.StartTraining(context.Background(), chatID, userID, models.SourceManual); err != nil {
		t.Fatalf("StartTraining failed: %v", err)
	}

	if err := deps.handler.skipCard(chatID, "test reason"); err != nil {
		t.Fatalf("skipCard failed: %v", err)
	}
	if deps.handler.HasActiveSession(chatID) {
		t.Fatal("expected session to be finished after skipping the only card")
	}
	if !strings.Contains(client.lastParams.Get("text"), "Тренировка завершена") {
		t.Fatalf("expected completion message after skip, got %q", client.lastParams.Get("text"))
	}
}

func TestTrainingHandler_HandleAnswer_UninitializedState(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)
	userID, userCardID := seedTrainableUserCard(t, deps, 910008)

	sessionID, err := deps.sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       userID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
		SessionJSON:  `{}`,
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	userCard, err := deps.userCardRepo.GetUserCard(userCardID)
	if err != nil || userCard == nil {
		t.Fatalf("GetUserCard failed: %v", err)
	}
	trainingCard, err := deps.trainingCardRepo.GetTrainingCard(userCard.TrainingCardID)
	if err != nil || trainingCard == nil {
		t.Fatalf("GetTrainingCard failed: %v", err)
	}
	card := &models.UserCardWithTraining{UserCard: *userCard, TrainingCard: *trainingCard}

	const chatID int64 = 1008
	deps.handler.sessions[chatID] = &SessionState{
		UserID:       userID,
		SessionID:    sessionID,
		Queue:        []*models.TrainingQueueItem{{Type: "card", Card: card}},
		CurrentIndex: 0,
		// options/correct answer intentionally empty to trigger recovery branch
	}

	err = deps.handler.HandleAnswer(chatID, 0)
	if err == nil || !strings.Contains(err.Error(), "please wait for options") {
		t.Fatalf("expected uninitialized-state error, got %v", err)
	}

	deps.handler.sessionsMutex.RLock()
	state := deps.handler.sessions[chatID]
	deps.handler.sessionsMutex.RUnlock()
	if state == nil || len(state.Options) == 0 || state.CorrectAnswer == "" {
		t.Fatalf("expected card state to be initialized after recovery, got %+v", state)
	}
}

func TestTrainingHandler_AutoRevealOptions(t *testing.T) {
	client := &mockTelegramClient{}
	deps := setupTrainingHandlerDeps(t, client)
	userID, _ := seedTrainableUserCard(t, deps, 910009)

	const chatID int64 = 1009
	if err := deps.handler.StartTraining(context.Background(), chatID, userID, models.SourceManual); err != nil {
		t.Fatalf("StartTraining failed: %v", err)
	}

	deps.handler.autoRevealOptions(chatID, 0)

	deps.handler.sessionsMutex.RLock()
	state := deps.handler.sessions[chatID]
	deps.handler.sessionsMutex.RUnlock()
	if state == nil || state.OptionsShownAt == nil {
		t.Fatal("expected autoRevealOptions to set OptionsShownAt")
	}
}
