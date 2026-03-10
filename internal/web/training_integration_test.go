package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupTrainingIntegrationTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.TrainingCardRepository, *repository.UserCardRepository, *repository.SessionRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)

	return db, userRepo, trainingCardRepo, userCardRepo, sessionRepo
}

func TestHandleTrainingStart_WithCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(434343)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card first (required for training card)
	var wordCardID int64
	err = db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "integration", "integration").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	trainingCard := &models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "integration",
		SenseIndex: 0,
		WordRU:     "интеграция",
		MeaningEN:  "integration",
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

	// Disable spell and type challenges so the first queue item is always a card (deterministic test)
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

	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	router := NewRouter(logger, cfg, db, trainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with user context
	req := httptest.NewRequest("POST", "/api/training/start", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleTrainingStart(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["question"] == nil {
		t.Error("Response should contain question")
	}
	if response["user_card_id"] == nil {
		t.Error("Response should contain user_card_id")
	}
}

// TestHandleTrainingStart_WithUserThresholds covers sessionConfig built from user settings with threshold clamping.
func TestHandleTrainingStart_WithUserThresholds(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, err := userRepo.GetOrCreateUser(434344)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	var wordCardID int64
	err = db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "thresh", "thresh").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "thresh", SenseIndex: 0, WordRU: "порог", MeaningEN: "thresh",
	})
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}
	past := time.Now().Add(-24 * time.Hour)
	_, err = userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}
	// SpellMasteringThreshold 150 -> clamped to 100, TypeMasteringThreshold -5 -> clamped to 0
	spellTh := 150
	typeTh := -5
	settings := models.UserSettings{
		SpellModeEnabled:        ptrBool(false),
		TypeModeEnabled:        ptrBool(false),
		SpellMasteringThreshold: &spellTh,
		TypeMasteringThreshold:  &typeTh,
	}
	settingsJSON, _ := json.Marshal(settings)
	if err := userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		t.Fatalf("UpdateUserSettings: %v", err)
	}
	cfg := &config.Config{
		WebApp:   config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
		Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	router := NewRouter(logger, cfg, db, trainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware
	req := httptest.NewRequest("POST", "/api/training/start", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleTrainingStart(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleTrainingReveal_WithSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(444444)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card first (required for training card)
	var wordCardID int64
	err = db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "reveal", "reveal").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	trainingCard := &models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "reveal",
		SenseIndex: 0,
		WordRU:     "раскрыть",
		MeaningEN:  "reveal",
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

	// Disable spell and type challenges so the first queue item is always a card (deterministic test)
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

	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, logger)
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

	// Now test reveal
	req := httptest.NewRequest("POST", "/api/training/reveal", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleTrainingReveal(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["options"] == nil {
		t.Error("Response should contain options")
	}
	if response["user_card_id"] == nil {
		t.Error("Response should contain user_card_id")
	}
}
