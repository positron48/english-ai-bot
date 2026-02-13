package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func TestFinishTrainingSession_CompleteSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	defer db.Close()

	// Create a user
	user, err := userRepo.GetOrCreateUser(464646)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card first (required for training card)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "finish", "finish")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	trainingCard := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "finish",
		SenseIndex: 0,
		WordRU:     "завершить",
		MeaningEN:  "finish",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a due card
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = userCardRepo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Disable spell and type challenges so the queue has only the card (deterministic test)
	settings := models.UserSettings{
		SpellModeEnabled: ptrBool(false),
		TypeModeEnabled:  ptrBool(false),
	}
	settingsJSON, _ := json.Marshal(settings)
	if err := userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		t.Fatalf("Failed to update user settings: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
		Training: config.TrainingConfig{
			OptionsDelayMS:         2000,
			WrongAnswerDelaySeconds: 3,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	router := NewRouter(logger, cfg, db, trainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Start a session
	startReq := httptest.NewRequest("POST", "/api/training/start", nil)
	startCtx := context.WithValue(startReq.Context(), userIDKey, user.ID)
	startReq = startReq.WithContext(startCtx)
	startW := httptest.NewRecorder()
	router.handleTrainingStart(startW, startReq)

	if startW.Code != http.StatusOK {
		t.Fatalf("Failed to start session: got status %d", startW.Code)
	}

	// Get session state
	router.webTrainingHandler.sessionsMutex.Lock()
	state, exists := router.webTrainingHandler.sessions[user.ID]
	router.webTrainingHandler.sessionsMutex.Unlock()

	if !exists || state == nil {
		t.Fatal("Session state should exist")
	}

	// Create a review event to simulate a correct answer (first item is a card)
	if state.Queue[0].Type != "card" || state.Queue[0].Card == nil {
		t.Fatal("Expected first queue item to be a card")
	}
	reviewEvent := &models.ReviewEvent{
		SessionID:  &state.SessionID,
		UserID:     user.ID,
		UserCardID: state.Queue[0].Card.UserCard.ID,
		Direction:  state.Queue[0].Card.UserCard.Direction,
		ShownAt:    time.Now(),
		AnsweredAt: &time.Time{},
		IsCorrect:  true,
		Quality:    4,
		OptionCount: 4,
	}
	_, err = sessionRepo.CreateReviewEvent(reviewEvent)
	if err != nil {
		t.Fatalf("Failed to create review event: %v", err)
	}

	// Reveal options
	revealReq := httptest.NewRequest("POST", "/api/training/reveal", nil)
	revealCtx := context.WithValue(revealReq.Context(), userIDKey, user.ID)
	revealReq = revealReq.WithContext(revealCtx)
	revealW := httptest.NewRecorder()
	router.handleTrainingReveal(revealW, revealReq)

	if revealW.Code != http.StatusOK {
		t.Fatalf("Failed to reveal options: got status %d", revealW.Code)
	}

	// Answer the card to complete it
	time.Sleep(10 * time.Millisecond)
	answerReq := httptest.NewRequest("POST", "/api/training/answer", strings.NewReader(fmt.Sprintf("option_index=0&user_card_id=%d", state.Queue[0].Card.UserCard.ID)))
	answerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	answerCtx := context.WithValue(answerReq.Context(), userIDKey, user.ID)
	answerReq = answerReq.WithContext(answerCtx)
	answerW := httptest.NewRecorder()
	router.handleTrainingAnswer(answerW, answerReq)

	if answerW.Code != http.StatusOK {
		t.Fatalf("Failed to answer: got status %d", answerW.Code)
	}

	// Now get the current card - this should trigger finishTrainingSession
	// since CurrentIndex >= len(Queue) after answering
	currentReq := httptest.NewRequest("GET", "/api/training/current", nil)
	currentCtx := context.WithValue(currentReq.Context(), userIDKey, user.ID)
	currentReq = currentReq.WithContext(currentCtx)
	currentW := httptest.NewRecorder()
	router.handleTrainingCurrent(currentW, currentReq)

	if currentW.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", currentW.Code)
	}

	// Verify response contains completion data
	var response map[string]interface{}
	if err := json.NewDecoder(currentW.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["complete"] != true {
		t.Error("Response should indicate session is complete")
	}

	if response["cards_completed"] == nil {
		t.Error("Response should contain cards_completed")
	}

	// Verify session was removed from memory
	router.webTrainingHandler.sessionsMutex.Lock()
	_, exists = router.webTrainingHandler.sessions[user.ID]
	router.webTrainingHandler.sessionsMutex.Unlock()

	if exists {
		t.Error("Session should be removed from memory after completion")
	}
}

func ptrBool(b bool) *bool {
	return &b
}
