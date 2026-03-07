package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupTrainingHandlersTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.TrainingCardRepository, *repository.UserCardRepository, *repository.SessionRepository) {
	db := testutil.SetupTestDB(t)
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	return db, userRepo, trainingCardRepo, userCardRepo, sessionRepo
}

func TestHandleTrainingStart_NoCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, sessionRepo := setupTrainingHandlersTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(55555)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
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

	userCardRepo := repository.NewUserCardRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
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

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	// Verify error message
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["error"] == nil {
		t.Error("Response should contain error message")
	}
}

func TestGetTrainingDelaysForUser_ConfigDefaults(t *testing.T) {
	cfg := &config.Config{
		Training: config.TrainingConfig{
			OptionsDelayMS:          3000,
			WrongAnswerDelaySeconds: 5,
		},
	}
	router := &Router{config: cfg}

	optsMS, wrongSec := router.getTrainingDelaysForUser(999)
	if optsMS != 3000 || wrongSec != 5 {
		t.Errorf("getTrainingDelaysForUser() = (%d, %d), want (3000, 5)", optsMS, wrongSec)
	}
}

func TestGetTrainingDelaysForUser_NilUserRepo(t *testing.T) {
	cfg := &config.Config{
		Training: config.TrainingConfig{
			OptionsDelayMS:          2000,
			WrongAnswerDelaySeconds: 3,
		},
	}
	router := &Router{config: cfg, userRepo: nil}

	optsMS, wrongSec := router.getTrainingDelaysForUser(1)
	if optsMS != 2000 || wrongSec != 3 {
		t.Errorf("getTrainingDelaysForUser() with nil userRepo = (%d, %d), want (2000, 3)", optsMS, wrongSec)
	}
}

func TestGetTrainingDelaysForUser_FromUserSettings(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)

	user, err := userRepo.GetOrCreateUser(88888)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	optsDelay := 4
	wrongDelay := 7
	settings := models.UserSettings{
		OptionsDelaySeconds:     &optsDelay,
		WrongAnswerDelaySeconds: &wrongDelay,
	}
	settingsJSON, _ := json.Marshal(settings)
	if err := userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		t.Fatalf("UpdateUserSettings: %v", err)
	}

	cfg := &config.Config{
		Training: config.TrainingConfig{
			OptionsDelayMS:          2000,
			WrongAnswerDelaySeconds: 3,
		},
	}
	router := &Router{
		config:   cfg,
		userRepo: userRepo,
		logger:   logger,
	}

	optsMS, wrongSec := router.getTrainingDelaysForUser(user.ID)
	wantOpts := 4000 // 4 * 1000
	if optsMS != wantOpts || wrongSec != 7 {
		t.Errorf("getTrainingDelaysForUser() = (%d, %d), want (%d, 7)", optsMS, wrongSec, wantOpts)
	}
}

func TestHandleTrainingStart_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)

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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create GET request (should fail)
	req := httptest.NewRequest("GET", "/api/training/start", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleTrainingStart(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleTrainingCurrent_NoSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(66666)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with user context
	req := httptest.NewRequest("GET", "/api/training/current", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleTrainingCurrent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response["active"] != false {
		t.Errorf("Expected active=false, got %v", response["active"])
	}
}

func TestHandleTrainingReveal_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)

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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create GET request (should fail, needs POST)
	req := httptest.NewRequest("GET", "/api/training/reveal", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleTrainingReveal(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleTrainingAnswer_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)

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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create GET request (should fail, needs POST)
	req := httptest.NewRequest("GET", "/api/training/answer", nil)
	w := httptest.NewRecorder()

	// Call handler
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleTrainingAnswer_MissingParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(77777)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request without answer
	req := httptest.NewRequest("POST", "/api/training/answer", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleTrainingAnswer_NoSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)

	user, err := userRepo.GetOrCreateUser(181818)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:        "test-secret",
			JWTTTLHours:      24,
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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/training/answer", strings.NewReader("option_index=0&user_card_id=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 when no session, got %d", w.Code)
	}
}

// errReader is a body that fails on Read (to trigger ParseForm error).
type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestHandleTrainingAnswer_ParseFormError(t *testing.T) {
	logger := zap.NewNop()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)

	user, err := userRepo.GetOrCreateUser(191919)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:        "test-secret",
			JWTTTLHours:      24,
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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/training/answer", io.NopCloser(errReader{}))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 on ParseForm error, got %d", w.Code)
	}
}
