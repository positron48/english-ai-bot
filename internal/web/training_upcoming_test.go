package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func setupTrainingUpcomingTest(t *testing.T) (*Router, *database.DB, func()) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	cleanup := func() {} // shared db, do not close

	return router, db, cleanup
}

func setTrainingUpcomingUserContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}

func TestHandleTrainingUpcoming_Get(t *testing.T) {
	router, db, cleanup := setupTrainingUpcomingTest(t)
	defer cleanup()

	// Create user
	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	req := httptest.NewRequest(http.MethodGet, "/api/training/upcoming", nil)
	req = setTrainingUpcomingUserContext(req, user.ID)
	rr := httptest.NewRecorder()

	router.handleTrainingUpcoming(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify response is valid JSON
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	// Should have 7 date keys (one for each day)
	if len(response) != 7 {
		t.Errorf("Expected 7 date entries, got %d", len(response))
	}
}

func TestHandleTrainingUpcoming_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupTrainingUpcomingTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/training/upcoming", nil)
	req = setTrainingUpcomingUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleTrainingUpcoming(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleTrainingUpcoming_Unauthorized(t *testing.T) {
	router, _, cleanup := setupTrainingUpcomingTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/training/upcoming", nil)
	// No user context
	rr := httptest.NewRecorder()

	router.handleTrainingUpcoming(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestHandleTrainingUpcoming_WithUserCards(t *testing.T) {
	router, db, cleanup := setupTrainingUpcomingTest(t)
	defer cleanup()

	// Create user
	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	// Create word card and training card
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("testword", "test definition")
	wordCard, _ := wordRepo.GetWordCard("testword")

	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), router.logger)
	tc := &models.TrainingCard{
		WordCardID:    wordCard.ID,
		WordEN:        "testword",
		SenseIndex:    0,
		WordRU:        "тест",
		MeaningEN:     "test meaning",
		DistractorsRU: `[]`,
		DistractorsEN: `[]`,
	}
	tcID, _ := tcRepo.CreateTrainingCard(tc)

	// Create user card with future due date
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), router.logger)
	dueDate := time.Now().AddDate(0, 0, 2) // Due in 2 days
	uc := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: tcID,
		State:          models.StateReview,
		NextDueAt:      &dueDate,
	}
	ucRepo.CreateUserCard(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/training/upcoming", nil)
	req = setTrainingUpcomingUserContext(req, user.ID)
	rr := httptest.NewRecorder()

	router.handleTrainingUpcoming(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify response contains upcoming cards
	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	// Should have date entries
	if len(response) == 0 {
		t.Error("Expected date entries in response")
	}
}

func TestHandleTrainingUpcoming_WithTimezone(t *testing.T) {
	router, db, cleanup := setupTrainingUpcomingTest(t)
	defer cleanup()

	// Create user with timezone (set via SQL since there's no direct method)
	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	// Set timezone via SQL
	tz := "America/New_York"
	db.GetConnection().Exec("UPDATE users SET timezone = ? WHERE id = ?", tz, user.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/training/upcoming", nil)
	req = setTrainingUpcomingUserContext(req, user.ID)
	rr := httptest.NewRecorder()

	router.handleTrainingUpcoming(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
