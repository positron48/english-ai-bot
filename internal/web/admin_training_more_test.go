package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func setupAdminAIService(t *testing.T, responseContent string) *ai.Service {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":` + strconv.Quote(responseContent) + `}}]}`))
	}))
	t.Cleanup(server.Close)

	svc := ai.NewService(server.URL, "test-model", "test-key", "test-prompt", zap.NewNop())
	svc.SetTrainingPrompt("PROMPT: ")
	return svc
}

func TestHandleAdminTraining_Generate_Success(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)

	response := `{"word_en":"run","transcription":"rʌn","senses":[{"pos":"verb","display_word":"run","word_ru":"бежать","meaning_en":"to move fast","example_en":"I run.","example_ru":"Я бегу.","distractors_ru":["идти","сидеть","стоять"],"distractors_en":["walk","sit","stand"],"hint":"movement"}]}`
	router.aiService = setupAdminAIService(t, response)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/run/generate", bytes.NewBufferString(`{"constraints":"verb only"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload["success"] != true {
		t.Fatalf("expected success=true, got %v", payload["success"])
	}
}

func TestHandleAdminTraining_Generate_LLMError(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	router.aiService = setupAdminAIService(t, `{"error":"llm failed"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/run/generate", bytes.NewBufferString(`{"constraints":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_Generate_FormData(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	response := `{"word_en":"run","transcription":"rʌn","senses":[{"pos":"verb","display_word":"run","word_ru":"бежать","meaning_en":"to move fast","example_en":"I run.","example_ru":"Я бегу.","distractors_ru":["идти","сидеть"],"distractors_en":["walk","sit"],"hint":""}]}`
	router.aiService = setupAdminAIService(t, response)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/run/generate", bytes.NewBufferString("constraints=verb+only"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for generate with form data, got %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload["success"] != true {
		t.Fatalf("expected success=true, got %v", payload["success"])
	}
}

func TestHandleAdminTraining_Generate_ValidationFailsMeaningTarget(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	response := `{"word_en":"casa","transcription":"kasa","senses":[{"pos":"noun","display_word":"casa","word_ru":"дом","meaning_en":"дом","example_en":"La casa es grande.","example_ru":"Дом большой.","distractors_ru":["квартира","жилище","здание"],"distractors_en":["hogar","edificio","vivienda"],"hint":"home"}]}`
	router.aiService = setupAdminAIService(t, response)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/casa/generate", bytes.NewBufferString(`{"constraints":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when validation fails, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_Generate_ValidationFailsEmptyTranscription(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	response := `{"word_en":"x","transcription":"","senses":[{"pos":"noun","display_word":"x","word_ru":"икс","meaning_en":"letter x","example_en":"This is x.","example_ru":"Это икс.","distractors_ru":["игрек","зет","альфа"],"distractors_en":["y","z","alpha"],"hint":"letter"}]}`
	router.aiService = setupAdminAIService(t, response)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/x/generate", bytes.NewBufferString(`{"constraints":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when transcription is empty, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_CreateCardJSON_SuccessCreatesUserCards(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db.GetConnection(), router.logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	userCardRepo := repository.NewUserCardRepository(db.GetConnection(), router.logger)

	if err := wordRepo.SaveWordCard("focus", "definition", ""); err != nil {
		t.Fatalf("SaveWordCard failed: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("focus")
	if err != nil || wordCard == nil {
		t.Fatalf("GetWordCard failed: %v", err)
	}

	baseCardID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID:    wordCard.ID,
		WordEN:        "focus",
		SenseIndex:    0,
		WordRU:        "фокус",
		MeaningEN:     "focus",
		DistractorsRU: `[]`,
		DistractorsEN: `[]`,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard failed: %v", err)
	}

	user, err := userRepo.GetOrCreateUser(555001)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}
	dueAt := time.Now().Add(-time.Hour)
	if _, err := userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: baseCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2,
		NextDueAt:      &dueAt,
	}); err != nil {
		t.Fatalf("CreateUserCard failed: %v", err)
	}

	createBody := `{"word_ru":"концентрация","meaning_en":"concentration","example_en":"Stay focused","example_ru":"Оставайся сосредоточенным","distractors_ru":"[\"внимание\",\"сила\",\"мысль\"]","distractors_en":"[\"attention\",\"power\",\"thought\"]"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/focus", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload["users_updated"] == nil || payload["user_cards_created"] == nil {
		t.Fatalf("expected users_updated and user_cards_created, got %v", payload)
	}
}

func TestHandleAdminTraining_CreateCardJSON_WordNotFound(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/missing", bytes.NewBufferString(`{"word_ru":"x","meaning_en":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_CreateCardJSON_MissingRequiredFields(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/word", bytes.NewBufferString(`{"word_ru":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_Delete_EmptyWord(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training//delete", nil)
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_ForbiddenWithoutPermissions(t *testing.T) {
	router, db, _ := setupAdminTrainingTest(t)
	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	nonAdmin, err := userRepo.GetOrCreateUser(888001)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/admin/training/test", nil)
	getReq = setUserIDInContext(getReq, nonAdmin.ID)
	getW := httptest.NewRecorder()
	router.handleAdminTraining(getW, getReq)
	if getW.Code != http.StatusForbidden {
		t.Fatalf("expected GET 403, got %d", getW.Code)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/api/admin/training/test", bytes.NewBufferString(`{"word_ru":"x","meaning_en":"y"}`))
	postReq.Header.Set("Content-Type", "application/json")
	postReq = setUserIDInContext(postReq, nonAdmin.ID)
	postW := httptest.NewRecorder()
	router.handleAdminTraining(postW, postReq)
	if postW.Code != http.StatusForbidden {
		t.Fatalf("expected POST 403, got %d", postW.Code)
	}
}

func TestHandleAdminTraining_Generate_NoSenses(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	router.aiService = setupAdminAIService(t, `{"word_en":"x","senses":[]}`)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/x/generate", bytes.NewBufferString(`{"constraints":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when no senses, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_Generate_ParseError(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	router.aiService = setupAdminAIService(t, `not json at all`)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/x/generate", bytes.NewBufferString(`{"constraints":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when parse fails, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_Generate_AIError(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	// Server returns 500 so GenerateAdditionalTrainingCard returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	t.Cleanup(server.Close)
	aiSvc := ai.NewService(server.URL, "test-model", "test-key", "PROMPT: ", zap.NewNop())
	aiSvc.SetTrainingPrompt("PROMPT: ")
	router.aiService = aiSvc

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/x/generate", bytes.NewBufferString(`{"constraints":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminTrainingUserContext(req, adminUserID)
	w := httptest.NewRecorder()

	router.handleAdminTraining(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when AI returns error, got %d: %s", w.Code, w.Body.String())
	}
}
