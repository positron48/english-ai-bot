package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupTrainingHandlersTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *repository.TrainingCardRepository, *repository.UserCardRepository, *repository.SessionRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_id INTEGER UNIQUE NOT NULL,
		telegram_username TEXT,
		username TEXT,
		timezone TEXT DEFAULT '',
		preferred_training_time TEXT DEFAULT '',
		settings_json TEXT DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
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
	);
	
	CREATE TABLE IF NOT EXISTS review_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER,
		user_id INTEGER NOT NULL,
		user_card_id INTEGER NOT NULL,
		direction TEXT NOT NULL,
		shown_at TEXT NOT NULL,
		options_shown_at TEXT,
		answered_at TEXT,
		t_delay_ms INTEGER,
		early_reveal INTEGER NOT NULL DEFAULT 0,
		option_count INTEGER NOT NULL,
		options_json TEXT,
		chosen_option TEXT,
		is_correct INTEGER NOT NULL DEFAULT 0,
		quality INTEGER NOT NULL,
		metrics_json TEXT,
		srs_before_json TEXT,
		srs_after_json TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

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
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	userCardRepo := repository.NewUserCardRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	router := NewRouter(logger, cfg, db, trainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with user context
	req := httptest.NewRequest("POST", "/app/training/start", nil)
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

func TestHandleTrainingStart_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)
	defer db.Close()

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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create GET request (should fail)
	req := httptest.NewRequest("GET", "/app/training/start", nil)
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
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with user context
	req := httptest.NewRequest("GET", "/app/training/current", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler
	router.handleTrainingCurrent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleTrainingReveal_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)
	defer db.Close()

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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create GET request (should fail, needs POST)
	req := httptest.NewRequest("GET", "/app/training/reveal", nil)
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
	defer db.Close()

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

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create GET request (should fail, needs POST)
	req := httptest.NewRequest("GET", "/app/training/answer", nil)
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
	defer db.Close()

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
	authMiddleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request without answer
	req := httptest.NewRequest("POST", "/app/training/answer", strings.NewReader(""))
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
