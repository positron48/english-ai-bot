package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func newSpellTypeRouter(t *testing.T, db *sql.DB) *Router {
	t.Helper()

	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
		Training: config.TrainingConfig{
			OptionsDelayMS:          2000,
			WrongAnswerDelaySeconds: 7,
		},
	}

	return NewRouter(logger, cfg, db, nil, nil, nil, nil)
}

func TestHandleTrainingAnswer_SpellAnswerText(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)

	const userID int64 = 9001
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID:    userID,
				SessionID: 11,
				Queue: []*models.TrainingQueueItem{{
					Type: "spell",
					Spell: &models.SpellChallenge{
						DisplayWord: "apple",
					},
				}},
				ShownAt: time.Now().Add(-2 * time.Second),
			},
		},
	}

	form := url.Values{}
	form.Set("answer_text", " APPLE ")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()

	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["is_correct"] != true {
		t.Fatalf("expected is_correct=true, got %v", payload["is_correct"])
	}
	if payload["chosen_option"] != "apple" {
		t.Fatalf("expected normalized chosen_option=apple, got %v", payload["chosen_option"])
	}
	if payload["correct_answer"] != "apple" {
		t.Fatalf("expected correct_answer=apple, got %v", payload["correct_answer"])
	}

	state := router.webTrainingHandler.sessions[userID]
	if state.CurrentIndex != 1 {
		t.Fatalf("expected current index to advance to 1, got %d", state.CurrentIndex)
	}
}

func TestHandleTrainingAnswer_TypeAnswerTextWrong(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)

	const userID int64 = 9002
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID:    userID,
				SessionID: 12,
				Queue: []*models.TrainingQueueItem{{
					Type: "type",
					TypeChallenge: &models.TypeChallenge{
						DisplayWord: "banana",
					},
				}},
				ShownAt: time.Now().Add(-2 * time.Second),
			},
		},
	}

	form := url.Values{}
	form.Set("answer_text", "wrong")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()

	router.handleTrainingAnswer(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["is_correct"] != false {
		t.Fatalf("expected is_correct=false, got %v", payload["is_correct"])
	}
	if payload["chosen_option"] != "wrong" {
		t.Fatalf("expected chosen_option=wrong, got %v", payload["chosen_option"])
	}
	if payload["delay_seconds"] != float64(7) {
		t.Fatalf("expected delay_seconds=7, got %v", payload["delay_seconds"])
	}

	state := router.webTrainingHandler.sessions[userID]
	if state.CurrentIndex != 1 {
		t.Fatalf("expected current index to advance to 1, got %d", state.CurrentIndex)
	}
}

func TestHandleTrainingSpellAnswer_NotSpellChallenge(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)

	const userID int64 = 9003
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				Queue: []*models.TrainingQueueItem{{
					Type: "card",
					Card: &models.UserCardWithTraining{},
				}},
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", nil)
	w := httptest.NewRecorder()
	router.handleTrainingSpellAnswer(w, req, userID, "word")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleTrainingTypeAnswer_NoActiveSession(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)

	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", nil)
	w := httptest.NewRecorder()
	router.handleTrainingTypeAnswer(w, req, 9999, "word")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestHandleTrainingSpellAnswer_NoHandler(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	router.webTrainingHandler = nil

	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", nil)
	w := httptest.NewRecorder()
	router.handleTrainingSpellAnswer(w, req, 9004, "word")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 when handler is nil, got %d", w.Code)
	}
}

func TestHandleTrainingSpellAnswer_NoSession(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	router.webTrainingHandler = &WebTrainingHandler{sessions: map[int64]*WebTrainingState{}}

	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", nil)
	w := httptest.NewRecorder()
	router.handleTrainingSpellAnswer(w, req, 9005, "word")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 when no session, got %d", w.Code)
	}
}

func TestHandleTrainingSpellAnswer_WrongAnswer(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 9006
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				UserID: userID, SessionID: 13,
				Queue: []*models.TrainingQueueItem{{
					Type:  "spell",
					Spell: &models.SpellChallenge{DisplayWord: "correct"},
				}},
				ShownAt: time.Now().Add(-2 * time.Second),
			},
		},
	}

	form := url.Values{}
	form.Set("answer_text", "wrong")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingSpellAnswer(w, req, userID, "wrong")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["is_correct"] != false {
		t.Fatalf("expected is_correct=false, got %v", payload["is_correct"])
	}
	if payload["delay_seconds"] != float64(7) {
		t.Fatalf("expected delay_seconds=7, got %v", payload["delay_seconds"])
	}
}

func TestHandleTrainingTypeAnswer_NoHandler(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	router.webTrainingHandler = nil

	w := httptest.NewRecorder()
	router.handleTrainingTypeAnswer(w, nil, 9007, "word")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 when handler is nil, got %d", w.Code)
	}
}

func TestHandleTrainingTypeAnswer_NoSession(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	router.webTrainingHandler = &WebTrainingHandler{sessions: map[int64]*WebTrainingState{}}

	w := httptest.NewRecorder()
	router.handleTrainingTypeAnswer(w, nil, 9008, "word")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 when no session, got %d", w.Code)
	}
}

func TestHandleTrainingTypeAnswer_NotTypeChallenge(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	router := newSpellTypeRouter(t, db)
	const userID int64 = 9009
	router.webTrainingHandler = &WebTrainingHandler{
		sessions: map[int64]*WebTrainingState{
			userID: {
				Queue: []*models.TrainingQueueItem{{
					Type: "card",
					Card: &models.UserCardWithTraining{},
				}},
			},
		},
	}

	w := httptest.NewRecorder()
	router.handleTrainingTypeAnswer(w, nil, userID, "word")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for non-type challenge, got %d", w.Code)
	}
}

// TestHandleTrainingSpellAnswer_WithReplacedCard calls gradeReplacedCardForSpellType from the handler when ReplacedUserCardID is set.
func TestHandleTrainingSpellAnswer_WithReplacedCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)
	user, err := userRepo.GetOrCreateUser(505051)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	var wordCardID int64
	if err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "spell", "spell").Scan(&wordCardID); err != nil {
		t.Fatalf("create word card: %v", err)
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "spell", SenseIndex: 0, WordRU: "орфография", MeaningEN: "spell",
	})
	if err != nil {
		t.Fatalf("create training card: %v", err)
	}
	nextDueAt := time.Now().Add(-time.Hour)
	userCardID, err := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: user.ID, TrainingCardID: trainingCardID, Direction: models.DirectionENtoRU,
		State: models.StateReview, EF: 2.0, NextDueAt: &nextDueAt,
	})
	if err != nil {
		t.Fatalf("create user card: %v", err)
	}
	sessionID, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID: user.ID, Source: models.SourceManual, PlannedCount: 1, DoneCount: 0,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	srsService := service.NewSRSService(userCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, nil, nil)
	router.webTrainingHandler = &WebTrainingHandler{
		sessionRepo: sessionRepo,
		sessions: map[int64]*WebTrainingState{
			user.ID: {
				UserID: user.ID, SessionID: sessionID, CurrentIndex: 0,
				Queue: []*models.TrainingQueueItem{{
					Type: "spell",
					Spell: &models.SpellChallenge{
						DisplayWord:        "spell",
						ReplacedUserCardID: userCardID,
					},
				}},
				ShownAt: time.Now().Add(-2 * time.Second),
			},
		},
	}
	form := url.Values{}
	form.Set("answer_text", "spell")
	req := httptest.NewRequest(http.MethodPost, "/api/training/answer", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleTrainingSpellAnswer(w, req, user.ID, "spell")
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM review_events WHERE session_id = ? AND user_card_id = ?", sessionID, userCardID).Scan(&count); err != nil {
		t.Fatalf("count review events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 review event for replaced card, got %d", count)
	}
}

func TestGradeReplacedCardForSpellType_CreatesReviewEventAndWrongAnswer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)

	user, err := userRepo.GetOrCreateUser(505050)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	var wordCardID int64
	if err := db.QueryRow(
		"INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id",
		"grade",
		"grade",
	).Scan(&wordCardID); err != nil {
		t.Fatalf("create word card: %v", err)
	}

	trainingCardID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "grade",
		SenseIndex: 0,
		WordRU:     "оценка",
		MeaningEN:  "grade",
	})
	if err != nil {
		t.Fatalf("create training card: %v", err)
	}

	nextDueAt := time.Now().Add(-time.Hour)
	userCardID, err := userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &nextDueAt,
	})
	if err != nil {
		t.Fatalf("create user card: %v", err)
	}

	sessionID, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	srsService := service.NewSRSService(userCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, nil, nil)
	router.webTrainingHandler = &WebTrainingHandler{sessionRepo: sessionRepo}

	shownAt := time.Now().Add(-2 * time.Second)
	answeredAt := time.Now()
	router.gradeReplacedCardForSpellType(user.ID, userCardID, false, "mistake", shownAt, answeredAt, sessionID, "spell", 7)

	var reviewEvents int
	if err := db.QueryRow("SELECT COUNT(*) FROM review_events WHERE session_id = ? AND user_card_id = ?", sessionID, userCardID).Scan(&reviewEvents); err != nil {
		t.Fatalf("count review events: %v", err)
	}
	if reviewEvents != 1 {
		t.Fatalf("expected 1 review event, got %d", reviewEvents)
	}

	var chosenOption string
	var isCorrect int
	if err := db.QueryRow(
		"SELECT chosen_option, is_correct FROM review_events WHERE session_id = ? AND user_card_id = ? LIMIT 1",
		sessionID,
		userCardID,
	).Scan(&chosenOption, &isCorrect); err != nil {
		t.Fatalf("load review event payload: %v", err)
	}
	if chosenOption != "mistake" {
		t.Fatalf("expected chosen_option=mistake, got %q", chosenOption)
	}
	if isCorrect != 0 {
		t.Fatalf("expected is_correct=0, got %d", isCorrect)
	}
}

func TestGradeReplacedCardForSpellType_UserCardMissing(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, _, userCardRepo, sessionRepo := setupTrainingIntegrationTestDB(t)

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	srsService := service.NewSRSService(userCardRepo, logger)
	router := NewRouter(logger, cfg, db, nil, srsService, nil, nil)
	router.webTrainingHandler = &WebTrainingHandler{sessionRepo: sessionRepo}

	now := time.Now()
	router.gradeReplacedCardForSpellType(1, 999999, true, "ok", now, now, 123, "spell", 2)

	var reviewEvents int
	if err := db.QueryRow("SELECT COUNT(*) FROM review_events").Scan(&reviewEvents); err != nil {
		t.Fatalf("count review events: %v", err)
	}
	if reviewEvents != 0 {
		t.Fatalf("expected no review events when user card is missing, got %d", reviewEvents)
	}
}
