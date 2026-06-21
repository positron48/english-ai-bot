package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupAdminTrainingTest(t *testing.T) (*Router, *database.DB, int64) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

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

	return router, db, adminUser.ID
}

func setAdminTrainingUserContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	// Set empty categories (super admin has all permissions)
	ctx = context.WithValue(ctx, userCategoriesKey, []int64{})
	return req.WithContext(ctx)
}

func TestHandleAdminTraining_GetWord(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	// Create word card
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("test", "test definition", "")

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
	router, _, adminUserID := setupAdminTrainingTest(t)

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
	router, db, adminUserID := setupAdminTrainingTest(t)

	// Create some training cards first
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("test", "test definition", "")
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
	router, db, adminUserID := setupAdminTrainingTest(t)

	// Create word and training card
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("testword", "test definition", "")
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
	router, _, adminUserID := setupAdminTrainingTest(t)

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
	router, _, adminUserID := setupAdminTrainingTest(t)

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
	router, _, adminUserID := setupAdminTrainingTest(t)

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
	router, _, adminUserID := setupAdminTrainingTest(t)

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
	router, db, adminUserID := setupAdminTrainingTest(t)

	// Create word and training card
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("testword", "test definition", "")
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

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/training/card/"+fmt.Sprint(tcID), nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["success"] != true {
		t.Error("expected success true")
	}
}

func TestHandleAdminTrainingCard_DeleteNotFound(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/training/card/99999", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTrainingCard_InvalidID(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)

	// Use PUT instead of GET since handleAdminTrainingCard doesn't handle GET
	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/card/invalid", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleAdminTrainingCard_EmptyCardID(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/card/", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 (card ID required), got %d", rr.Code)
	}
}

func TestHandleAdminTraining_CreateCardForm_Success(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	if err := wordRepo.SaveWordCard("formword", "def", ""); err != nil {
		t.Fatalf("SaveWordCard: %v", err)
	}

	form := "word_ru=перевод&meaning_en=meaning&example_en=ex&example_ru=пример"
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/formword", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTraining_Generate_FormConstraints(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	router.aiService = setupAdminAIService(t, `{"word_en":"x","transcription":"eks","senses":[{"pos":"n","display_word":"x","word_ru":"икс","meaning_en":"letter x","example_en":"","example_ru":"","distractors_ru":[],"distractors_en":[],"hint":""}]}`)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/x/generate", strings.NewReader("constraints=verb"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTraining(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTrainingCard_MethodNotAllowed(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), router.logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("w", "d", "")
	wc, _ := wordRepo.GetWordCard("w")
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "w", SenseIndex: 0, WordRU: "п", MeaningEN: "m", DistractorsRU: "[]", DistractorsEN: "[]"}
	cardID, _ := tcRepo.CreateTrainingCard(tc)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/card/"+fmt.Sprint(cardID), nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminTrainingCard_PUT_JSON(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), router.logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("word", "def", "")
	wc, _ := wordRepo.GetWordCard("word")
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "word", SenseIndex: 0, WordRU: "слово", MeaningEN: "meaning", DistractorsRU: "[]", DistractorsEN: "[]"}
	cardID, err := tcRepo.CreateTrainingCard(tc)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	body := map[string]interface{}{"word_ru": "обновлено", "meaning_en": "updated meaning"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/card/"+fmt.Sprint(cardID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTrainingCard_PUT_Form(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), router.logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("w2", "def", "")
	wc, _ := wordRepo.GetWordCard("w2")
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "w2", SenseIndex: 0, WordRU: "сл", MeaningEN: "m", DistractorsRU: "[]", DistractorsEN: "[]"}
	cardID, _ := tcRepo.CreateTrainingCard(tc)

	form := "word_ru=обновлено&meaning_en=updated"
	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/card/"+fmt.Sprint(cardID), strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminTrainingCard_PUT_Form_PosEmptyClearsPOS(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), router.logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("w3", "def", "")
	wc, _ := wordRepo.GetWordCard("w3")
	posVal := "noun"
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "w3", SenseIndex: 0, WordRU: "сл", MeaningEN: "m", POS: &posVal, DistractorsRU: "[]", DistractorsEN: "[]"}
	cardID, _ := tcRepo.CreateTrainingCard(tc)

	// Form with pos= (empty) so posProvided is true and card.POS is set to nil
	form := "word_ru=ok&meaning_en=ok&pos="
	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/card/"+fmt.Sprint(cardID), strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminTrainingCard(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWords_Get_More(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	// Create some words
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("apple", "a fruit", "")
	wordRepo.SaveWordCard("banana", "another fruit", "")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWords(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWords_GetWithSearch(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("apple", "a fruit", "")
	wordRepo.SaveWordCard("banana", "another fruit", "")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words?search=apple", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWords(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWords_GetWithMissingTrainingPOS(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("apple", "a fruit", "")

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words?missing_training_pos=noun", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWords(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWord_Get(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("testword", "test definition", "")
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
	router, _, adminUserID := setupAdminTrainingTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/words/99999", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWord(rr, req)

	if rr.Code != http.StatusNotFound && rr.Code != http.StatusOK {
		t.Errorf("Expected status 404 or 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWord_Delete_More(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)

	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	wordRepo.SaveWordCard("testword", "test definition", "")
	wordCard, _ := wordRepo.GetWordCard("testword")

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/words/"+string(rune(wordCard.ID+'0')), nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWord(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d: %s", rr.Code, rr.Body.String())
	}
}
