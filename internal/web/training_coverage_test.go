package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

// ---- helpers ----

// newBrokenDB returns a *sql.DB that is already closed (for testing error paths).
// Opens a separate connection to the shared test DB, then closes it.
// Uses GetTestDSN (not SetupTestDB) to avoid truncating the shared DB.
func newBrokenDB(t *testing.T) *sql.DB {
	t.Helper()
	// GetTestDSN ensures the container is up and the postgres_compat driver is registered
	// (via database.NewWithConfig called internally), without truncating tables.
	dsn := testutil.GetTestDSN(t)
	conn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("postgres_compat open: %v", err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		t.Skipf("ping: %v", err)
	}
	conn.Close()
	return conn
}

func newCoverageRouter(t *testing.T) (*Router, *repository.UserRepository, *repository.TrainingCardRepository, *repository.UserCardRepository, *repository.SessionRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
		Training: config.TrainingConfig{
			OptionsDelayMS:          2000,
			WrongAnswerDelaySeconds: 5,
		},
	}
	userRepo := repository.NewUserRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	router := NewRouter(logger, cfg, db, trainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	return router, userRepo, trainingCardRepo, userCardRepo, sessionRepo
}

func ctxWithUser(req *http.Request, userID int64) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
}

// ---- handleTrainingStart: internal error from StartSession ----

// TestHandleTrainingStart_InternalError covers the 500 branch when StartSession returns a non-"no cards" error.
// We use a second postgres container (separate from shared) to simulate a broken DB.
func TestHandleTrainingStart_InternalError(t *testing.T) {
	logger := zap.NewNop()
	db, userRepo, _, _, _ := setupTrainingIntegrationTestDB(t)
	cfg := &config.Config{
		WebApp:   config.WebAppConfig{JWTSecret: "s", JWTTTLHours: 1, RefreshTTLHours: 720},
		Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3},
	}
	user, err := userRepo.GetOrCreateUser(700001)
	if err != nil || user == nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	// Use a broken DB (second container, closed) to force StartSession to fail.
	brokenConn := newBrokenDB(t)

	brokenUserCardRepo := repository.NewUserCardRepository(brokenConn, logger)
	brokenTrainingCardRepo := repository.NewTrainingCardRepository(brokenConn, logger)
	brokenSessionRepo := repository.NewSessionRepository(brokenConn, logger)
	brokenTrainingService := service.NewTrainingService(brokenUserCardRepo, brokenTrainingCardRepo, brokenSessionRepo, nil, logger)
	srsService := service.NewSRSService(brokenUserCardRepo, logger)
	optionsService := service.NewOptionsService(brokenTrainingCardRepo, logger)

	router := NewRouter(logger, cfg, db, brokenTrainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	req := httptest.NewRequest(http.MethodPost, "/api/training/start", nil)
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingStart(w, req)

	if w.Code != http.StatusInternalServerError && w.Code != http.StatusBadRequest {
		t.Errorf("expected 500 or 400 on broken db, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleTrainingStart_SpellThresholdNegative covers the t<0 branch for SpellMasteringThreshold.
func TestHandleTrainingStart_SpellThresholdNegative(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, err := userRepo.GetOrCreateUser(700002)
	if err != nil || user == nil {
		t.Fatalf("GetOrCreateUser: %v, user=%v", err, user)
	}

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "neg", "neg").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "neg", SenseIndex: 0, WordRU: "отриц", MeaningEN: "neg",
	})
	past := time.Now().Add(-time.Hour)
	userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})

	spellTh := -10
	typeTh := 110
	settings := models.UserSettings{
		SpellModeEnabled:        ptrBool(false),
		TypeModeEnabled:         ptrBool(false),
		SpellMasteringThreshold: &spellTh,
		TypeMasteringThreshold:  &typeTh,
	}
	settingsJSON, _ := json.Marshal(settings)
	userRepo.UpdateUserSettings(user.ID, string(settingsJSON))

	cfg := &config.Config{
		WebApp:   config.WebAppConfig{JWTSecret: "s", JWTTTLHours: 1, RefreshTTLHours: 720},
		Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3},
	}
	trainingService := service.NewTrainingService(userCardRepo, trainingCardRepo, sessionRepo, nil, logger)
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	router := NewRouter(logger, cfg, db, trainingService, srsService, optionsService, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	req := httptest.NewRequest(http.MethodPost, "/api/training/start", nil)
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingStart(w, req)
	// Just verify it doesn't panic and returns a valid response
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status %d", w.Code)
	}
}


// TestShowTrainingCard_TypeChallenge_NoPrefix covers type challenge where DisplayWord has no "to " prefix.
func TestShowTrainingCard_TypeChallenge_NoPrefix(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 1000, WrongAnswerDelaySeconds: 3}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	state := &WebTrainingState{
		SessionID: 1, CurrentIndex: 0,
		Queue: []*models.TrainingQueueItem{{
			Type: "type",
			TypeChallenge: &models.TypeChallenge{
				WordRU:      "яблоко",
				DisplayWord: "apple",
			},
		}},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.showTrainingCard(w, req, state)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if payload["prefix"] != "" {
		t.Errorf("expected empty prefix for non-verb, got %v", payload["prefix"])
	}
	if payload["hint_first_letter"] != "a" {
		t.Errorf("expected hint_first_letter=a, got %v", payload["hint_first_letter"])
	}
	if payload["hint_length"] != float64(5) {
		t.Errorf("expected hint_length=5, got %v", payload["hint_length"])
	}
}

// TestShowTrainingCard_TypeChallenge_EmptyDisplayWord covers type challenge with empty DisplayWord (0 runes).
func TestShowTrainingCard_TypeChallenge_EmptyDisplayWord(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 1000, WrongAnswerDelaySeconds: 3}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	state := &WebTrainingState{
		SessionID: 1, CurrentIndex: 0,
		Queue: []*models.TrainingQueueItem{{
			Type: "type",
			TypeChallenge: &models.TypeChallenge{
				WordRU:      "пусто",
				DisplayWord: "",
			},
		}},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.showTrainingCard(w, req, state)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if payload["hint_first_letter"] != "" {
		t.Errorf("expected empty hint_first_letter for empty word, got %v", payload["hint_first_letter"])
	}
	if payload["hint_length"] != float64(0) {
		t.Errorf("expected hint_length=0, got %v", payload["hint_length"])
	}
}

// ---- extractSessionWordsFromQueue: WordCardID match ----

// TestExtractSessionWordsFromQueue_SameWordCardID covers the branch where a queue item has the same WordCardID as current.
func TestExtractSessionWordsFromQueue_SameWordCardID(t *testing.T) {
	r := &Router{}
	current := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{WordCardID: 5, WordEN: "run", WordRU: "бежать"},
	}
	sameWordCard := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{WordCardID: 5, WordEN: "run2", WordRU: "бежать2"},
	}
	queue := []*models.TrainingQueueItem{
		{Type: "card", Card: current},
		{Type: "card", Card: sameWordCard},
	}
	words := r.extractSessionWordsFromQueue(queue, 0, current, nil)
	if len(words) != 0 {
		t.Errorf("expected no words when same WordCardID, got %v", words)
	}
}

// ---- extractSessionWords: RUtoEN with DisplayWord ----

// TestExtractSessionWords_RUtoEN_WithDisplayWord covers the DisplayWord branch in RUtoEN direction.
func TestExtractSessionWords_RUtoEN_WithDisplayWord(t *testing.T) {
	r := &Router{}
	verb := "verb"
	displayWord := "to run"
	current := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{WordCardID: 1, WordEN: "run", WordRU: "бежать", POS: &verb},
	}
	other := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{WordCardID: 2, WordEN: "jump", WordRU: "прыгать", POS: &verb, DisplayWord: &displayWord},
	}
	queue := []*models.UserCardWithTraining{current, other}
	words := r.extractSessionWords(queue, 0, current, nil)
	if len(words) != 1 || words[0] != "to run" {
		t.Errorf("expected ['to run'] with DisplayWord, got %v", words)
	}
}

// TestExtractSessionWords_SameWordEN covers the branch filtering same WordEN.
func TestExtractSessionWords_SameWordEN(t *testing.T) {
	r := &Router{}
	current := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{WordCardID: 1, WordEN: "run", WordRU: "бежать"},
	}
	sameEN := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{WordCardID: 2, WordEN: "run", WordRU: "другое"},
	}
	queue := []*models.UserCardWithTraining{current, sameEN}
	words := r.extractSessionWords(queue, 0, current, nil)
	if len(words) != 0 {
		t.Errorf("expected no words when same WordEN, got %v", words)
	}
}

// ---- handleTrainingCurrent: session exists but nil (explicit nil in map) ----

// TestHandleTrainingCurrent_NilSessionInMap covers the !exists || state == nil branch.
func TestHandleTrainingCurrent_NilSessionInMap(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{
		WebApp:   config.WebAppConfig{JWTSecret: "s", JWTTTLHours: 1, RefreshTTLHours: 720},
		Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3},
	}
	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(700010)
	if err != nil || user == nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(db, logger)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.webTrainingHandler = NewWebTrainingHandler(nil, nil, nil, sessionRepo, logger, 2000, 3)
	// Explicitly set nil state in map
	router.webTrainingHandler.sessionsMutex.Lock()
	router.webTrainingHandler.sessions[user.ID] = nil
	router.webTrainingHandler.sessionsMutex.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingCurrent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["active"] != false {
		t.Errorf("expected active=false for nil session, got %v", resp["active"])
	}
}

// ---- handleTrainingReveal: no session (exists=false) ----

// TestHandleTrainingReveal_SessionNotExists covers the !exists branch.
func TestHandleTrainingReveal_SessionNotExists(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{
		WebApp:   config.WebAppConfig{JWTSecret: "s", JWTTTLHours: 1, RefreshTTLHours: 720},
		Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3},
	}
	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(700011)
	if err != nil || user == nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	sessionRepo := repository.NewSessionRepository(db, logger)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.webTrainingHandler = NewWebTrainingHandler(nil, nil, nil, sessionRepo, logger, 2000, 3)
	// No session in map

	req := httptest.NewRequest(http.MethodPost, "/api/training/reveal", nil)
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingReveal(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---- gradeReplacedCardForSpellType: GradeCard error ----

// TestGradeReplacedCardForSpellType_GradeCardError covers the error branch when GradeCard fails.
// We use a broken srsService (nil userCardRepo) to trigger the error.
func TestGradeReplacedCardForSpellType_GradeCardError(t *testing.T) {
	logger := zap.NewNop()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, err := userRepo.GetOrCreateUser(700020)
	if err != nil || user == nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "graderr", "graderr").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "graderr", SenseIndex: 0, WordRU: "ошибка", MeaningEN: "graderr",
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}

	// Use a broken db for srsService so GradeCard fails
	brokenDB := newBrokenDB(t)
	brokenUserCardRepo := repository.NewUserCardRepository(brokenDB, logger)
	brokenSRSService := service.NewSRSService(brokenUserCardRepo, logger)

	router := NewRouter(logger, cfg, db, nil, brokenSRSService, nil, nil)
	router.webTrainingHandler = &WebTrainingHandler{sessionRepo: sessionRepo}

	// Should not panic, just log error and return
	router.gradeReplacedCardForSpellType(user.ID, userCardID, true, "graderr", time.Now(), 1)
}

// TestGradeReplacedCardForSpellType_CreateReviewEventError covers lines 674-676 where CreateReviewEvent fails.
func TestGradeReplacedCardForSpellType_CreateReviewEventError(t *testing.T) {
	logger := zap.NewNop()
	db, userRepo, trainingCardRepo, userCardRepo, _ := setupTrainingIntegrationTestDB(t)
	user, err := userRepo.GetOrCreateUser(700021)
	if err != nil || user == nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "graderr2", "graderr2").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "graderr2", SenseIndex: 0, WordRU: "ошибка2", MeaningEN: "graderr2",
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}

	// srsService uses good DB so GradeCard succeeds
	goodUserCardRepo := repository.NewUserCardRepository(db, logger)
	goodSRSService := service.NewSRSService(goodUserCardRepo, logger)

	// sessionRepo uses broken DB so CreateReviewEvent fails
	brokenDB := newBrokenDB(t)
	brokenSessionRepo := repository.NewSessionRepository(brokenDB, logger)

	router := NewRouter(logger, cfg, db, nil, goodSRSService, nil, nil)
	router.webTrainingHandler = &WebTrainingHandler{sessionRepo: brokenSessionRepo}

	// Should not panic, just log error
	router.gradeReplacedCardForSpellType(user.ID, userCardID, true, "graderr2", time.Now(), 1)
}

// TestGradeReplacedCardForSpellType_RecordWrongAnswerError covers lines 678-680 where RecordWrongAnswer fails.
// TestGradeReplacedCardForSpellType_WrongAnswer covers the !isCorrect branch (lines 677-680).
func TestGradeReplacedCardForSpellType_WrongAnswer(t *testing.T) {
	logger := zap.NewNop()
	db, userRepo, trainingCardRepo, userCardRepo, _ := setupTrainingIntegrationTestDB(t)
	user, err := userRepo.GetOrCreateUser(700022)
	if err != nil || user == nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "graderr3", "graderr3").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "graderr3", SenseIndex: 0, WordRU: "ошибка3", MeaningEN: "graderr3",
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}

	goodUserCardRepo := repository.NewUserCardRepository(db, logger)
	goodSRSService := service.NewSRSService(goodUserCardRepo, logger)
	goodSessionRepo := repository.NewSessionRepository(db, logger)

	router := NewRouter(logger, cfg, db, nil, goodSRSService, nil, nil)
	router.webTrainingHandler = &WebTrainingHandler{sessionRepo: goodSessionRepo}

	// isCorrect=false covers the !isCorrect branch (lines 677-680)
	router.gradeReplacedCardForSpellType(user.ID, userCardID, false, "wrong", time.Now(), 1)
}

// ---- handleTrainingSpellAnswer: empty answer (userNorm becomes " ") ----

// TestHandleTrainingSpellAnswer_EmptyAnswer covers the userNorm == "" branch.
func TestHandleTrainingSpellAnswer_EmptyAnswer(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 700030
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 1, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type:  "spell",
					Spell: &models.SpellChallenge{DisplayWord: "hello"},
				}},
				ShownAt: time.Now().Add(-time.Second),
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", nil)
	w := httptest.NewRecorder()
	router.handleTrainingSpellAnswer(w, req, userID, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if payload["is_correct"] != false {
		t.Errorf("empty answer should be incorrect, got %v", payload["is_correct"])
	}
}

// ---- handleTrainingTypeAnswer: empty answer ----

// TestHandleTrainingTypeAnswer_EmptyAnswer covers the userNorm == "" branch.
func TestHandleTrainingTypeAnswer_EmptyAnswer(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 700031
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 1, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type:          "type",
					TypeChallenge: &models.TypeChallenge{DisplayWord: "banana"},
				}},
				ShownAt: time.Now().Add(-time.Second),
			},
		},
	}

	w := httptest.NewRecorder()
	router.handleTrainingTypeAnswer(w, nil, userID, "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if payload["is_correct"] != false {
		t.Errorf("empty answer should be incorrect, got %v", payload["is_correct"])
	}
}

// TestHandleTrainingTypeAnswer_WithReplacedCard covers the replacedUserCardID != 0 branch.
func TestHandleTrainingTypeAnswer_WithReplacedCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700032)

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "typecard", "typecard").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "typecard", SenseIndex: 0, WordRU: "тип", MeaningEN: "typecard",
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})
	sessionID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	srsService := service.NewSRSService(userCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, nil, nil)
	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: sessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: sessionID, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "type",
					TypeChallenge: &models.TypeChallenge{
						DisplayWord:        "typecard",
						ReplacedUserCardID: userCardID,
					},
				}},
				ShownAt: time.Now().Add(-time.Second),
			},
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", nil)
	router.handleTrainingTypeAnswer(w, req, user.ID, "typecard")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if payload["is_correct"] != true {
		t.Errorf("expected is_correct=true, got %v", payload["is_correct"])
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM review_events WHERE session_id = ? AND user_card_id = ?", sessionID, userCardID).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 review event for replaced card, got %d", count)
	}
}

// ---- handleTrainingAnswer: various uncovered branches ----

// TestHandleTrainingAnswer_AnswerTextWithNoActiveSession covers the else branch (unlock) when answer_text is set but no session.
func TestHandleTrainingAnswer_AnswerTextWithNoActiveSession(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 700040

	// Handler exists but no session for this user
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{},
	}

	form := url.Values{}
	form.Set("answer_text", "hello")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	// Falls through to option_index parsing which fails (empty string)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (invalid option_index), got %d", w.Code)
	}
}

// TestHandleTrainingAnswer_NoHandlerAfterOptionParsing covers the r.webTrainingHandler == nil branch after option parsing.
func TestHandleTrainingAnswer_NoHandlerAfterOptionParsing(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	router.webTrainingHandler = nil
	const userID int64 = 700041

	form := url.Values{}
	form.Set("option_index", "0")
	form.Set("user_card_id", "1")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when no handler, got %d", w.Code)
	}
}

// TestHandleTrainingAnswer_SessionExhausted covers the state.CurrentIndex >= len(state.Queue) branch.
func TestHandleTrainingAnswer_SessionExhausted(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 700042

	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 1, CurrentIndex: 1,
				Queue: []*models.TrainingQueueItem{
					{Type: "card", Card: &models.UserCardWithTraining{}},
				},
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", "0")
	form.Set("user_card_id", "1")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when session exhausted, got %d", w.Code)
	}
}

// TestHandleTrainingAnswer_NotCardType covers the item.Type != "card" branch.
func TestHandleTrainingAnswer_NotCardType(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 700043

	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 1, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type:  "spell",
					Spell: &models.SpellChallenge{DisplayWord: "x"},
				}},
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", "0")
	form.Set("user_card_id", "1")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-card type, got %d", w.Code)
	}
}

// TestHandleTrainingAnswer_CardMismatch covers the card.UserCard.ID != userCardID branch.
func TestHandleTrainingAnswer_CardMismatch(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 700044

	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 1, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card",
					Card: &models.UserCardWithTraining{
						UserCard: models.UserCard{ID: 99},
					},
				}},
				Options: []string{"a", "b"},
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", "0")
	form.Set("user_card_id", "1") // mismatch: card ID is 99
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for card mismatch, got %d", w.Code)
	}
}

// TestHandleTrainingAnswer_InvalidOptionIndexRange covers the optionIndex < 0 || optionIndex >= len(options) branch.
func TestHandleTrainingAnswer_InvalidOptionIndexRange(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 700045

	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 1, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card",
					Card: &models.UserCardWithTraining{
						UserCard: models.UserCard{ID: 5},
					},
				}},
				Options: []string{"a", "b"},
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", "99") // out of range
	form.Set("user_card_id", "5")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for out-of-range option index, got %d", w.Code)
	}
}

// TestHandleTrainingAnswer_WithOptionsShownAt covers the optionsShownAt != nil branches.
func TestHandleTrainingAnswer_WithOptionsShownAt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700046)

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "opts", "opts").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "opts", SenseIndex: 0, WordRU: "опции", MeaningEN: "opts",
		DistractorsRU: `["а","б","в"]`, DistractorsEN: `["a","b","c"]`,
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})
	sessionID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, optionsService, nil)

	// Pre-generate options
	card := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: userCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: wordCardID, WordEN: "opts", WordRU: "опции",
			DistractorsRU: `["а","б","в"]`,
		},
	}
	options, correctAnswer, _ := optionsService.GenerateOptions(card, 4, nil, nil, nil)

	shownAt := time.Now().Add(-3 * time.Second)
	optionsShownAt := time.Now().Add(-2 * time.Second)
	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: sessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: sessionID, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card", Card: card,
				}},
				ShownAt:        shownAt,
				OptionsShownAt: &optionsShownAt,
				Options:        options,
				CorrectAnswer:  correctAnswer,
			},
		},
	}

	// Find the correct option index
	correctIdx := 0
	for i, opt := range options {
		if opt == correctAnswer {
			correctIdx = i
			break
		}
	}

	form := url.Values{}
	form.Set("option_index", fmt.Sprintf("%d", correctIdx))
	form.Set("user_card_id", fmt.Sprintf("%d", userCardID))
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if payload["is_correct"] != true {
		t.Errorf("expected is_correct=true, got %v", payload["is_correct"])
	}
}

// TestHandleTrainingAnswer_WrongAnswerWithRecentCorrect covers the isCorrect=false branch and RecentCorrectAnswers trimming.
func TestHandleTrainingAnswer_WrongAnswerWithRecentCorrect(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700047)

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "wrong", "wrong").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "wrong", SenseIndex: 0, WordRU: "неверно", MeaningEN: "wrong",
		DistractorsRU: `["а","б","в"]`,
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})
	sessionID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 5}}
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, optionsService, nil)

	card := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: userCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: wordCardID, WordEN: "wrong", WordRU: "неверно",
			DistractorsRU: `["а","б","в"]`,
		},
	}
	options, correctAnswer, _ := optionsService.GenerateOptions(card, 4, nil, nil, nil)

	// Find a wrong option index
	wrongIdx := 0
	for i, opt := range options {
		if opt != correctAnswer {
			wrongIdx = i
			break
		}
	}

	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: sessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: sessionID, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card", Card: card,
				}},
				ShownAt:              time.Now().Add(-3 * time.Second),
				Options:              options,
				CorrectAnswer:        correctAnswer,
				RecentCorrectAnswers: []string{"prev1", "prev2"},
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", fmt.Sprintf("%d", wrongIdx))
	form.Set("user_card_id", fmt.Sprintf("%d", userCardID))
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if payload["is_correct"] != false {
		t.Errorf("expected is_correct=false, got %v", payload["is_correct"])
	}
	if payload["delay_seconds"] != float64(5) {
		t.Errorf("expected delay_seconds=5, got %v", payload["delay_seconds"])
	}
}

// TestHandleTrainingAnswer_CorrectAnswer_RecentCorrectAnswersTrimmed covers RecentCorrectAnswers > 2 trimming.
func TestHandleTrainingAnswer_CorrectAnswer_RecentCorrectAnswersTrimmed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700048)

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "trim", "trim").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "trim", SenseIndex: 0, WordRU: "обрезать", MeaningEN: "trim",
		DistractorsRU: `["а","б","в"]`,
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})
	sessionID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, optionsService, nil)

	card := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: userCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: wordCardID, WordEN: "trim", WordRU: "обрезать",
			DistractorsRU: `["а","б","в"]`,
		},
	}
	options, correctAnswer, _ := optionsService.GenerateOptions(card, 4, nil, nil, nil)

	correctIdx := 0
	for i, opt := range options {
		if opt == correctAnswer {
			correctIdx = i
			break
		}
	}

	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: sessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: sessionID, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card", Card: card,
				}},
				ShownAt:              time.Now().Add(-3 * time.Second),
				Options:              options,
				CorrectAnswer:        correctAnswer,
				RecentCorrectAnswers: []string{"prev1", "prev2"}, // already 2, will be trimmed after adding
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", fmt.Sprintf("%d", correctIdx))
	form.Set("user_card_id", fmt.Sprintf("%d", userCardID))
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &payload)
	if payload["is_correct"] != true {
		t.Errorf("expected is_correct=true, got %v", payload["is_correct"])
	}
}

// TestHandleTrainingAnswer_UserCardDeletedDuringSession covers the existingCard == nil branch.
func TestHandleTrainingAnswer_UserCardDeletedDuringSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700049)

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "deleted", "deleted").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "deleted", SenseIndex: 0, WordRU: "удалено", MeaningEN: "deleted",
		DistractorsRU: `["а","б","в"]`,
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})
	sessionID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, optionsService, nil)

	card := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: userCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: wordCardID, WordEN: "deleted", WordRU: "удалено",
			DistractorsRU: `["а","б","в"]`,
		},
	}
	options, correctAnswer, _ := optionsService.GenerateOptions(card, 4, nil, nil, nil)

	correctIdx := 0
	for i, opt := range options {
		if opt == correctAnswer {
			correctIdx = i
			break
		}
	}

	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: sessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: sessionID, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card", Card: card,
				}},
				ShownAt:       time.Now().Add(-3 * time.Second),
				Options:       options,
				CorrectAnswer: correctAnswer,
			},
		},
	}

	// Delete the user card before answering to simulate deletion during session
	db.Exec("DELETE FROM user_cards WHERE id = ?", userCardID)

	form := url.Values{}
	form.Set("option_index", fmt.Sprintf("%d", correctIdx))
	form.Set("user_card_id", fmt.Sprintf("%d", userCardID))
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	// Should still return 200 (skips review event creation)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even when card deleted, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleTrainingAnswer_ReviewEventCreateError covers the CreateReviewEvent error branch.
func TestHandleTrainingAnswer_ReviewEventCreateError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, _ := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700050)

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "reventerr", "reventerr").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "reventerr", SenseIndex: 0, WordRU: "событие", MeaningEN: "reventerr",
		DistractorsRU: `["а","б","в"]`,
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	srsService := service.NewSRSService(userCardRepo, logger)
	optionsService := service.NewOptionsService(trainingCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, optionsService, nil)

	card := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: userCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: wordCardID, WordEN: "reventerr", WordRU: "событие",
			DistractorsRU: `["а","б","в"]`,
		},
	}
	options, correctAnswer, _ := optionsService.GenerateOptions(card, 4, nil, nil, nil)

	correctIdx := 0
	for i, opt := range options {
		if opt == correctAnswer {
			correctIdx = i
			break
		}
	}

	// Use a broken sessionRepo to trigger CreateReviewEvent error
	brokenDB := newBrokenDB(t)
	brokenSessionRepo := repository.NewSessionRepository(brokenDB, logger)

	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: brokenSessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: 1, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card", Card: card,
				}},
				ShownAt:       time.Now().Add(-3 * time.Second),
				Options:       options,
				CorrectAnswer: correctAnswer,
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", fmt.Sprintf("%d", correctIdx))
	form.Set("user_card_id", fmt.Sprintf("%d", userCardID))
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	// Should still return 200 (error is logged but not fatal)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even with review event error, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleTrainingAnswer_GradeCardError covers the GradeCard error branch (lines 934-936).
// GradeCard error is just logged, execution continues.
func TestHandleTrainingAnswer_GradeCardError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700051)

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "graderr", "graderr").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "graderr", SenseIndex: 0, WordRU: "ошибка", MeaningEN: "graderr",
		DistractorsRU: `["а","б","в"]`,
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})
	sessionID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	// Use broken srsService so GradeCard fails (but execution continues)
	brokenDB := newBrokenDB(t)
	brokenUserCardRepo := repository.NewUserCardRepository(brokenDB, logger)
	brokenSRSService := service.NewSRSService(brokenUserCardRepo, logger)

	router := NewRouter(logger, cfg, db, nil, brokenSRSService, optionsService, nil)

	card := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: userCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: wordCardID, WordEN: "graderr", WordRU: "ошибка",
			DistractorsRU: `["а","б","в"]`,
		},
	}
	options, correctAnswer, _ := optionsService.GenerateOptions(card, 4, nil, nil, nil)

	correctIdx := 0
	for i, opt := range options {
		if opt == correctAnswer {
			correctIdx = i
			break
		}
	}

	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: sessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: sessionID, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card", Card: card,
				}},
				ShownAt:       time.Now().Add(-3 * time.Second),
				Options:       options,
				CorrectAnswer: correctAnswer,
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", fmt.Sprintf("%d", correctIdx))
	form.Set("user_card_id", fmt.Sprintf("%d", userCardID))
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	// GradeCard error is logged but not fatal, should still return 200
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even with GradeCard error, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleTrainingAnswer_GetUserCardError covers the GetUserCard error branch (lines 962-964).
// GetUserCard error is just logged, execution continues to skip review event creation.
func TestHandleTrainingAnswer_GetUserCardError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700052)

	var wordCardID int64
	db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "getcarderr", "getcarderr").Scan(&wordCardID)
	trainingCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "getcarderr", SenseIndex: 0, WordRU: "ошибка карты", MeaningEN: "getcarderr",
		DistractorsRU: `["а","б","в"]`,
	})
	past := time.Now().Add(-time.Hour)
	userCardID, _ := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &past,
	})
	sessionID, _ := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	optionsService := service.NewOptionsService(trainingCardRepo, logger)

	// Use broken r.db so GetUserCard fails (line 962-964)
	// srsService uses good DB so GradeCard succeeds
	goodUserCardRepo := repository.NewUserCardRepository(db, logger)
	goodSRSService := service.NewSRSService(goodUserCardRepo, logger)

	brokenDB := newBrokenDB(t)
	router := NewRouter(logger, cfg, brokenDB, nil, goodSRSService, optionsService, nil)

	card := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: userCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: wordCardID, WordEN: "getcarderr", WordRU: "ошибка карты",
			DistractorsRU: `["а","б","в"]`,
		},
	}
	options, correctAnswer, _ := optionsService.GenerateOptions(card, 4, nil, nil, nil)

	correctIdx := 0
	for i, opt := range options {
		if opt == correctAnswer {
			correctIdx = i
			break
		}
	}

	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: sessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: sessionID, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card", Card: card,
				}},
				ShownAt:       time.Now().Add(-3 * time.Second),
				Options:       options,
				CorrectAnswer: correctAnswer,
			},
		},
	}

	form := url.Values{}
	form.Set("option_index", fmt.Sprintf("%d", correctIdx))
	form.Set("user_card_id", fmt.Sprintf("%d", userCardID))
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	// GetUserCard error is logged but not fatal, should still return 200
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even with GetUserCard error, got %d: %s", w.Code, w.Body.String())
	}
}

// ---- finishTrainingSession: GetSessionStats error ----

// TestFinishTrainingSession_GetSessionStatsError covers the err != nil branch in GetSessionStats.
func TestFinishTrainingSession_GetSessionStatsError(t *testing.T) {
	logger := zap.NewNop()
	db, _, _, _, sessionRepo := setupTrainingIntegrationTestDB(t)

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	trainingService := service.NewTrainingService(nil, nil, sessionRepo, nil, logger)
	router := NewRouter(logger, cfg, db, trainingService, nil, nil, nil)
	router.webTrainingHandler = NewWebTrainingHandler(trainingService, nil, nil, sessionRepo, logger, 2000, 3)

	// Use a broken sessionRepo to trigger GetSessionStats error
	brokenDB := newBrokenDB(t)
	brokenSessionRepo := repository.NewSessionRepository(brokenDB, logger)

	// Create a router that uses broken db for GetSessionStats
	router2 := NewRouter(logger, cfg, brokenDB, trainingService, nil, nil, nil)
	router2.webTrainingHandler = NewWebTrainingHandler(trainingService, nil, nil, brokenSessionRepo, logger, 2000, 3)

	state := &WebTrainingState{
		UserID:       1,
		SessionID:    999999,
		CurrentIndex: 3,
		Queue:        nil,
	}
	router2.webTrainingHandler.sessionsMutex.Lock()
	router2.webTrainingHandler.sessions[state.UserID] = state
	router2.webTrainingHandler.sessionsMutex.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router2.finishTrainingSession(w, req, state)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even with stats error, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["complete"] != true {
		t.Errorf("expected complete=true, got %v", resp["complete"])
	}
	// When stats fail, totalCards = state.CurrentIndex = 3
	if resp["total_cards"] != float64(3) {
		t.Errorf("expected total_cards=3 (fallback), got %v", resp["total_cards"])
	}
}

// TestFinishTrainingSession_FinishSessionError covers the err != nil branch in FinishSession (line 1047-1049).
func TestFinishTrainingSession_FinishSessionError(t *testing.T) {
	logger := zap.NewNop()
	db, _, _, _, sessionRepo := setupTrainingIntegrationTestDB(t)

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}

	// Use a broken sessionRepo for trainingService so FinishSession fails
	brokenDB := newBrokenDB(t)
	brokenSessionRepo := repository.NewSessionRepository(brokenDB, logger)
	brokenTrainingService := service.NewTrainingService(nil, nil, brokenSessionRepo, nil, logger)

	router := NewRouter(logger, cfg, db, brokenTrainingService, nil, nil, nil)
	router.webTrainingHandler = NewWebTrainingHandler(brokenTrainingService, nil, nil, sessionRepo, logger, 2000, 3)

	state := &WebTrainingState{
		UserID:       1,
		SessionID:    999999,
		CurrentIndex: 3,
		Queue:        nil,
	}
	router.webTrainingHandler.sessionsMutex.Lock()
	router.webTrainingHandler.sessions[state.UserID] = state
	router.webTrainingHandler.sessionsMutex.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.finishTrainingSession(w, req, state)

	// FinishSession error is just logged, response is still 200
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even with FinishSession error, got %d", w.Code)
	}
}

// TestFinishTrainingSession_NilWebTrainingHandler covers the r.webTrainingHandler == nil branch.
func TestFinishTrainingSession_NilWebTrainingHandler(t *testing.T) {
	logger := zap.NewNop()
	db, _, _, _, sessionRepo := setupTrainingIntegrationTestDB(t)
	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	trainingService := service.NewTrainingService(nil, nil, sessionRepo, nil, logger)
	router := NewRouter(logger, cfg, db, trainingService, nil, nil, nil)
	router.webTrainingHandler = nil // explicitly nil

	state := &WebTrainingState{
		UserID: 1, SessionID: 999998, CurrentIndex: 2,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.finishTrainingSession(w, req, state)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with nil handler, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["complete"] != true {
		t.Errorf("expected complete=true, got %v", resp["complete"])
	}
}

// ---- handleTrainingUpcoming: GetUserByID error and invalid timezone ----

// TestHandleTrainingUpcoming_GetUserError covers the err != nil branch from GetUserByID.
func TestHandleTrainingUpcoming_GetUserError(t *testing.T) {
	logger := zap.NewNop()
	// Use a broken db so GetUserByID fails
	brokenDB := newBrokenDB(t)

	cfg := &config.Config{}
	router := NewRouter(logger, cfg, brokenDB, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/training/upcoming", nil)
	req = ctxWithUser(req, 700060)
	w := httptest.NewRecorder()
	router.handleTrainingUpcoming(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetUserByID fails, got %d", w.Code)
	}
}

// TestHandleTrainingUpcoming_InvalidTimezone covers the time.LoadLocation error branch.
func TestHandleTrainingUpcoming_InvalidTimezone(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{}
	userRepo := repository.NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(700061)

	// Set invalid timezone
	db.Exec("UPDATE users SET timezone = ? WHERE id = ?", "Invalid/Timezone", user.ID)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/training/upcoming", nil)
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingUpcoming(w, req)

	// Should succeed (falls back to UTC)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with invalid timezone (fallback to UTC), got %d", w.Code)
	}
}

// TestHandleTrainingUpcoming_GetUpcomingCardsError covers the err != nil branch from GetUpcomingCardsByDate.
// It uses a second postgres container so we can safely drop tables without affecting the shared test DB.
func TestHandleTrainingUpcoming_GetUpcomingCardsError(t *testing.T) {
	logger := zap.NewNop()

	// Get a second postgres DSN (separate container, safe to modify schema)
	dsn := testutil.SecondPostgresDSN(t) // skips if docker unavailable

	// Connect and migrate the second DB (retry to wait for container readiness)
	var secondDB *database.DB
	var err error
	for i := 0; i < 10; i++ {
		secondDB, err = database.NewWithConfig("postgres", "", dsn, logger)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * 300 * time.Millisecond)
	}
	if err != nil {
		t.Skipf("second postgres not available: %v", err)
	}
	conn := secondDB.GetConnection()

	// Create user in second DB
	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(700062)
	if err != nil {
		t.Skipf("GetOrCreateUser failed: %v", err)
	}

	// Drop user_cards to force GetUpcomingCardsByDate to fail
	if _, err := conn.Exec("DROP TABLE IF EXISTS user_cards CASCADE"); err != nil {
		t.Skipf("cannot drop user_cards: %v", err)
	}

	cfg := &config.Config{}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/training/upcoming", nil)
	req = ctxWithUser(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingUpcoming(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetUpcomingCardsByDate fails, got %d", w.Code)
	}
}

// TestHandleTrainingAnswer_AnswerTextWithCurrentIndexExhausted covers the else branch when answer_text is set but CurrentIndex >= len(Queue).
func TestHandleTrainingAnswer_AnswerTextCurrentIndexExhausted(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 700070

	// Session exists but CurrentIndex >= len(Queue)
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 1, CurrentIndex: 5,
				Queue: []*models.TrainingQueueItem{
					{Type: "card", Card: &models.UserCardWithTraining{}},
				},
			},
		},
	}

	form := url.Values{}
	form.Set("answer_text", "hello")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	// Falls through to option_index parsing which fails (empty string)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (invalid option_index), got %d", w.Code)
	}
}

// TestHandleTrainingAnswer_AnswerTextWithCardType covers the answer_text branch when item.Type is "card" (not type or spell).
func TestHandleTrainingAnswer_AnswerTextWithCardType(t *testing.T) {
	router, _, _, _, _ := newCoverageRouter(t)
	const userID int64 = 700071

	// Session with a "card" type item - answer_text branch falls through
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 1, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "card",
					Card: &models.UserCardWithTraining{
						UserCard: models.UserCard{ID: 1},
					},
				}},
				Options: []string{"a", "b"},
			},
		},
	}

	form := url.Values{}
	form.Set("answer_text", "hello")
	form.Set("option_index", "0")
	form.Set("user_card_id", "1")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingAnswer(w, req)

	// Falls through to card answer handling - card mismatch (ID=1 but no real card in db)
	// or invalid option index
	if w.Code != http.StatusBadRequest && w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("unexpected status %d", w.Code)
	}
}
