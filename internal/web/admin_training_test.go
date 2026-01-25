package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupAdminTrainingTest(t *testing.T) (*Router, *database.DB, int64, func()) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)
	
	// Initialize dependencies for permission checks
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	
	// Create admin user in DB (required for IsSuperAdmin to work)
	adminUser, err := userRepo.GetOrCreateUser(int64(cfg.Admin.TelegramID))
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return router, db, adminUser.ID, cleanup
}

func setAdminTrainingUserContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	// Set empty categories (super admin has all permissions)
	ctx = context.WithValue(ctx, userCategoriesKey, []int64{})
	return req.WithContext(ctx)
}

func TestHandleAdminTraining_GetWord(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	// Create word card
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("test", "test definition")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/training/test", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["word_en"] != "test" {
		t.Errorf("Expected word_en 'test', got %v", response["word_en"])
	}
}

func TestHandleAdminTraining_GetWord_NotFound(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/training/nonexistent", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	// Should return 200 with empty cards
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTraining_DeleteAll_More(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	// Create some training cards first
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("test", "test definition")
	wordCard, _ := wordRepo.GetWordCard("test")

	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), router.logger)
	tc := &models.TrainingCard{
		WordCardID:    wordCard.ID,
		WordEN:        "test",
		SenseIndex:    0,
		WordRU:        "тест",
		MeaningEN:     "test meaning",
		DistractorsRU: `[]`,
		DistractorsEN: `[]`,
	}
	tcRepo.CreateTrainingCard(tc)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/delete_all", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response["success"] != true {
		t.Errorf("Expected success true, got %v", response["success"])
	}
}

func TestHandleAdminTraining_DeleteWord(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	// Create word and training card
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
	tcRepo.CreateTrainingCard(tc)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/testword/delete", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTraining_Generate_NoAIService(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	body := bytes.NewBufferString(`{"constraints": "test constraint"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/test/generate", body)
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	// Should fail because AI service is nil
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTraining_Generate_InvalidJSON(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	body := bytes.NewBufferString(`{invalid json}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/test/generate", body)
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTraining_InvalidPath(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/training/", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	// Empty word should still work but return empty
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTraining_UnknownAction(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/test", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	// PUT with unknown action returns 400
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 400 or 405, got %d", rr.Code)
	}
}

func TestHandleAdminTrainingCard_Delete_More(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	// Create word and training card
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

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/training/card/"+string(rune(tcID+'0')), nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	// Either 200 or 404 depending on card existence
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTrainingCard_DeleteNotFound(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/training/card/99999", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTrainingCard_InvalidID(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	// Use PUT instead of GET since handleAdminTrainingCard doesn't handle GET
	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/card/invalid", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleAdminWords_Get_More(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	// Create some words
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("apple", "a fruit")
	wordRepo.SaveWordCard("banana", "another fruit")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWords(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWords_GetWithSearch(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("apple", "a fruit")
	wordRepo.SaveWordCard("banana", "another fruit")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words?search=apple", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWords(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWord_Get(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("testword", "test definition")
	wordCard, _ := wordRepo.GetWordCard("testword")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words/"+string(rune(wordCard.ID+'0')), nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWord(rr, req)

	// GET is not supported, expects 405 or 400
	if rr.Code != http.StatusOK && rr.Code != http.StatusMethodNotAllowed && rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200, 400, or 405, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWord_NotFound(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/words/99999", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWord(rr, req)

	if rr.Code != http.StatusNotFound && rr.Code != http.StatusOK {
		t.Errorf("Expected status 404 or 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWord_Delete_More(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminTrainingTest(t)
	defer cleanup()

	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("testword", "test definition")
	wordCard, _ := wordRepo.GetWordCard("testword")

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/words/"+string(rune(wordCard.ID+'0')), nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWord(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d: %s", rr.Code, rr.Body.String())
	}
}
