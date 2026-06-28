package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"strconv"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

func seedWordTrainingReportCard(t *testing.T, router *Router, userID int64) (userCardID, trainingCardID int64) {
	t.Helper()
	trainingCardRepo := repository.NewTrainingCardRepository(router.db, router.logger)
	userCardRepo := repository.NewUserCardRepository(router.db, router.logger)
	var wordCardID int64
	if err := router.db.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id`, "reportword", "reportword").Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	tcID, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "reportword",
		SenseIndex: 0,
		WordRU:     "жалоба",
		MeaningEN:  "reportword",
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}
	ucID, err := userCardRepo.CreateUserCard(&models.UserCard{
		UserID: userID, TrainingCardID: tcID, Direction: models.DirectionRUtoEN, State: models.StateReview,
	})
	if err != nil {
		t.Fatalf("CreateUserCard: %v", err)
	}
	return ucID, tcID
}

func TestHandleTrainingReport_OKAndValidation(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	userCardID, trainingCardID := seedWordTrainingReportCard(t, router, userID)

	body := map[string]interface{}{
		"user_card_id":     userCardID,
		"training_card_id": trainingCardID,
		"report_category":  "bad_audio",
		"comment":          "audio clipped",
		"client_report_id": "wr-1",
	}
	raw, _ := json.Marshal(body)
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/training/report", bytes.NewReader(raw)), userID)
	w := httptest.NewRecorder()
	router.handleTrainingReport(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["success"] != true || resp["report_id"] == nil {
		t.Fatalf("unexpected response: %+v", resp)
	}

	// Duplicate client_report_id.
	req = withUserID(httptest.NewRequest(http.MethodPost, "/api/training/report", bytes.NewReader(raw)), userID)
	w = httptest.NewRecorder()
	router.handleTrainingReport(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("duplicate expected 200, got %d", w.Code)
	}
	var dup map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&dup)
	if dup["duplicate"] != true {
		t.Fatalf("expected duplicate=true, got %+v", dup)
	}

	req = withUserID(httptest.NewRequest(http.MethodPost, "/api/training/report", bytes.NewReader([]byte(`{"report_category":"bad_audio"}`))), userID)
	w = httptest.NewRecorder()
	router.handleTrainingReport(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing word expected 400, got %d", w.Code)
	}

	req = withUserID(httptest.NewRequest(http.MethodPost, "/api/training/report", bytes.NewReader([]byte(`{"word":"x","report_category":"other"}`))), userID)
	w = httptest.NewRecorder()
	router.handleTrainingReport(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("other without comment expected 400, got %d", w.Code)
	}
}

func TestHandleLearningGrammarTrainingReport_OK(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()

	body := map[string]interface{}{
		"question_id":     "es.chapter.1::b1::q9",
		"chapter_id":      "es.chapter.1",
		"theory_block_id": "b1",
		"report_category": "wrong_answer",
		"comment":         "wrong key",
		"client_report_id": "gr-tr-1",
	}
	raw, _ := json.Marshal(body)
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/report", bytes.NewReader(raw)), userID)
	w := httptest.NewRecorder()
	router.handleLearningGrammarTrainingReport(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = withUserID(httptest.NewRequest(http.MethodPost, "/api/learning/grammar/training/report", bytes.NewReader([]byte(`{"chapter_id":"x"}`))), userID)
	w = httptest.NewRecorder()
	router.handleLearningGrammarTrainingReport(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing question_id expected 400, got %d", w.Code)
	}
}

func TestHandleAdminContentReports_ListGetAndResolve(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	userCardID, trainingCardID := seedWordTrainingReportCard(t, router, userID)

	repo := repository.NewContentReportRepository(router.db, router.logger)
	reportID, err := repo.Create(repository.CreateContentReportInput{
		UserID: userID, SourceType: "word_training", Word: "reportword",
		TrainingCardID: &trainingCardID, UserCardID: &userCardID,
		ReportCategory: "bad_audio", CommentText: "test",
	})
	if err != nil {
		t.Fatalf("Create report: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/content-reports?status=active", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleAdminContentReports(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var listResp struct {
		Reports []map[string]interface{} `json:"reports"`
	}
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Reports) == 0 {
		t.Fatal("expected at least one report")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/content-reports/"+strconv.FormatInt(reportID, 10), nil)
	req.URL.Path = "/api/admin/content-reports/" + strconv.FormatInt(reportID, 10)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleAdminContentReportByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get by id expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resolveBody := []byte(`{}`)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/content-reports/"+strconv.FormatInt(reportID, 10)+"/resolve", bytes.NewReader(resolveBody))
	req.URL.Path = "/api/admin/content-reports/" + strconv.FormatInt(reportID, 10) + "/resolve"
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleAdminContentReportByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("resolve expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/content-reports/not-a-number", nil)
	req.URL.Path = "/api/admin/content-reports/not-a-number"
	w = httptest.NewRecorder()
	router.handleAdminContentReportByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid id expected 400, got %d", w.Code)
	}
}

func TestValidateReportComment(t *testing.T) {
	if !validateReportComment("word_training", "bad_audio", "") {
		t.Fatal("category without comment should be valid when normalized")
	}
	if validateReportComment("word_training", "other", "") {
		t.Fatal("other without comment should be invalid")
	}
}

func TestHandleInternalTrainingCardAndTTS(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	_ = userID
	_, trainingCardID := seedWordTrainingReportCard(t, router, userID)
	router.internalServiceTokens = map[string]string{"default": "tok"}
	router.pronunciationService = &mockPronunciationService{enabled: true}

	update := map[string]interface{}{
		"word_ru":     "обновлено",
		"meaning_en":  "updated",
		"display_word": "",
		"pos":         "",
	}
	raw, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/api/internal/training/card/"+strconv.FormatInt(trainingCardID, 10), bytes.NewReader(raw))
	req.URL.Path = "/api/internal/training/card/" + strconv.FormatInt(trainingCardID, 10)
	req.Header.Set("X-Service-Token", "tok")
	w := httptest.NewRecorder()
	router.handleInternalTrainingCard(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update card expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/internal/training/card/bad", bytes.NewReader(raw))
	req.URL.Path = "/api/internal/training/card/bad"
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalTrainingCard(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad card id expected 400, got %d", w.Code)
	}

	regenBody, _ := json.Marshal(map[string]string{"word": "reportword"})
	req = httptest.NewRequest(http.MethodPost, "/api/internal/tts/regenerate", bytes.NewReader(regenBody))
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalTTSRegenerate(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("regenerate expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/internal/tts/status?word=reportword", nil)
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalTTSStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status expected 200, got %d: %s", w.Code, w.Body.String())
	}

	disabled := NewRouter(router.logger, router.config, router.db, nil, nil, nil, nil)
	disabled.internalServiceTokens = map[string]string{"default": "tok"}
	req = httptest.NewRequest(http.MethodGet, "/api/internal/tts/status?word=x", nil)
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	disabled.handleInternalTTSStatus(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("no TTS expected 503, got %d", w.Code)
	}
}

func TestHandleInternalContentReportsSubpath(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	router.internalServiceTokens = map[string]string{"default": "tok"}

	reportID := seedInternalGrammarReport(t, router, userID, "en.chapter.sub", "bs", "en.chapter.sub::bs::q1")

	req := httptest.NewRequest(http.MethodGet, "/api/internal/content-reports/"+strconv.FormatInt(reportID, 10), nil)
	req.URL.Path = "/api/internal/content-reports/" + strconv.FormatInt(reportID, 10)
	req.Header.Set("X-Service-Token", "tok")
	w := httptest.NewRecorder()
	router.handleInternalContentReportsSubpath(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get by id expected 200, got %d: %s", w.Code, w.Body.String())
	}

	body, _ := json.Marshal(map[string]interface{}{"report_ids": []int64{reportID}, "reason": "test"})
	req = httptest.NewRequest(http.MethodPost, "/api/internal/content-reports/resolve-bulk", bytes.NewReader(body))
	req.URL.Path = "/api/internal/content-reports/resolve-bulk"
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalContentReportsSubpath(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("resolve bulk expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/internal/content-reports/unknown-sub", nil)
	req.URL.Path = "/api/internal/content-reports/unknown-sub"
	req.Header.Set("X-Service-Token", "tok")
	w = httptest.NewRecorder()
	router.handleInternalContentReportsSubpath(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown subpath expected 404, got %d", w.Code)
	}
}

func TestGrammarTrainingPackRelPath(t *testing.T) {
	rel, packID := grammarTrainingPackRelPath("", "b")
	if rel != "" || packID != "" {
		t.Fatalf("empty chapter should return empty, got %q %q", rel, packID)
	}
	rel, packID = grammarTrainingPackRelPath("es.chapter.1", "some-block")
	if packID != "es" {
		t.Fatalf("packID = %q, want es", packID)
	}
	_ = rel // may be empty if block not in embedded pack index
}
