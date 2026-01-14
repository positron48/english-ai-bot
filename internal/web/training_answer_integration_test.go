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
	"tgbot-skeleton/internal/service"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func TestHandleTrainingAnswer_WithSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	defer db.Close()

	// Create a user
	user, err := userRepo.GetOrCreateUser(454545)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card first (required for training card)
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "answer", "answer")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	trainingCard := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "answer",
		SenseIndex: 0,
		WordRU:     "ответ",
		MeaningEN:  "answer",
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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	router := NewRouter(logger, cfg, db, trainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Start a session first
	startReq := httptest.NewRequest("POST", "/api/training/start", nil)
	startCtx := context.WithValue(startReq.Context(), userIDKey, user.ID)
	startReq = startReq.WithContext(startCtx)
	startW := httptest.NewRecorder()
	router.handleTrainingStart(startW, startReq)

	if startW.Code != http.StatusOK {
		t.Fatalf("Failed to start session: got status %d", startW.Code)
	}

	// Get user_card_id from response
	var startResponse map[string]interface{}
	if err := json.NewDecoder(startW.Body).Decode(&startResponse); err != nil {
		t.Fatalf("Failed to decode start response: %v", err)
	}
	userCardID, ok := startResponse["user_card_id"].(float64)
	if !ok {
		t.Fatal("Failed to get user_card_id from start response")
	}

	// Reveal options first
	revealReq := httptest.NewRequest("POST", "/api/training/reveal", nil)
	revealCtx := context.WithValue(revealReq.Context(), userIDKey, user.ID)
	revealReq = revealReq.WithContext(revealCtx)
	revealW := httptest.NewRecorder()
	router.handleTrainingReveal(revealW, revealReq)

	if revealW.Code != http.StatusOK {
		t.Fatalf("Failed to reveal options: got status %d", revealW.Code)
	}

	// Get options from response
	var revealResponse map[string]interface{}
	if err := json.NewDecoder(revealW.Body).Decode(&revealResponse); err != nil {
		t.Fatalf("Failed to decode reveal response: %v", err)
	}
	options, ok := revealResponse["options"].([]interface{})
	if !ok {
		t.Fatal("Failed to get options from reveal response")
	}
	if len(options) == 0 {
		t.Fatal("Options should not be empty")
	}

	// Now test answer (use first option)
	answerReq := httptest.NewRequest("POST", "/api/training/answer", strings.NewReader(fmt.Sprintf("option_index=0&user_card_id=%.0f", userCardID)))
	answerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	answerCtx := context.WithValue(answerReq.Context(), userIDKey, user.ID)
	answerReq = answerReq.WithContext(answerCtx)
	answerW := httptest.NewRecorder()
	
	// Small delay to ensure optionsShownAt is set
	time.Sleep(10 * time.Millisecond)

	// Call handler
	router.handleTrainingAnswer(answerW, answerReq)

	if answerW.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", answerW.Code)
	}

	// Verify response
	var answerResponse map[string]interface{}
	if err := json.NewDecoder(answerW.Body).Decode(&answerResponse); err != nil {
		t.Fatalf("Failed to decode answer response: %v", err)
	}
	if answerResponse["is_correct"] == nil {
		t.Error("Response should contain is_correct")
	}
}
