package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupGrammarOfflineTest(t *testing.T) (*Router, int64, string) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{
		Learning: config.DefaultLearningConfig(),
		WebApp:   config.WebAppConfig{PublicURL: "https://test.example/app"},
	}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(900201)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	contentRepo := repository.NewGrammarContentRepository(logger)
	publishRepo := repository.NewGrammarPublishRepository(conn, logger)
	attemptRepo := repository.NewGrammarAttemptRepository(conn, logger)
	grammarService := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, cfg.Learning, logger)
	router.SetGrammarService(grammarService)
	return router, user.ID, ""
}

func publishFirstGrammarChapter(t *testing.T, router *Router) (sectionID, chapterID string) {
	t.Helper()
	sectionsData, err := router.grammarService.ContentRepo.GetSections()
	if err != nil || len(sectionsData.Sections) == 0 {
		t.Fatalf("sections: %v", err)
	}
	section := sectionsData.Sections[0]
	if len(section.ChapterIDs) == 0 {
		t.Fatal("section has no chapters")
	}
	chapterID = section.ChapterIDs[0]
	if err := router.grammarService.PublishRepo.SetPublished("section", section.SectionID, true, nil); err != nil {
		t.Fatalf("publish section: %v", err)
	}
	if err := router.grammarService.PublishRepo.SetPublished("chapter", chapterID, true, nil); err != nil {
		t.Fatalf("publish chapter: %v", err)
	}
	return section.SectionID, chapterID
}

func TestGrammarOfflineManifest_OK(t *testing.T) {
	router, userID, _ := setupGrammarOfflineTest(t)
	publishFirstGrammarChapter(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/offline/manifest", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningGrammarOfflineManifest(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Sections []map[string]interface{} `json:"sections"`
		Total    int                        `json:"total_chapters"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Sections) == 0 {
		t.Fatal("expected at least one published section")
	}
}

func TestGrammarOfflineManifest_Errors(t *testing.T) {
	router, userID, _ := setupGrammarOfflineTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/offline/manifest", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningGrammarOfflineManifest(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/learning/grammar/offline/manifest", nil)
	w = httptest.NewRecorder()
	router.handleLearningGrammarOfflineManifest(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGrammarOfflineTrainingPack(t *testing.T) {
	router, userID, _ := setupGrammarOfflineTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/offline/training-pack", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningGrammarOfflineTrainingPack(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/learning/grammar/offline/training-pack", nil)
	w = httptest.NewRecorder()
	router.handleLearningGrammarOfflineTrainingPack(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGrammarOfflineChapter(t *testing.T) {
	router, _, _ := setupGrammarOfflineTest(t)
	_, chapterID := publishFirstGrammarChapter(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/grammar/offline/chapters/"+chapterID, nil)
	w := httptest.NewRecorder()
	router.handleLearningGrammarOfflineChapter(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/learning/grammar/offline/chapters/", nil)
	w = httptest.NewRecorder()
	router.handleLearningGrammarOfflineChapter(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty chapter expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/learning/grammar/offline/chapters/missing-chapter", nil)
	w = httptest.NewRecorder()
	router.handleLearningGrammarOfflineChapter(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing chapter expected 404, got %d", w.Code)
	}
}

func TestGrammarOfflineSyncAttempts(t *testing.T) {
	router, userID, _ := setupGrammarOfflineTest(t)
	sectionID, chapterID := publishFirstGrammarChapter(t, router)

	payload := map[string]interface{}{
		"attempts": []map[string]interface{}{
			{"client_attempt_id": "", "scope": "chapter", "scope_id": chapterID},
			{"client_attempt_id": "offline-grammar-1", "scope": "", "scope_id": chapterID},
			{
				"client_attempt_id": "offline-grammar-2",
				"scope":             "chapter",
				"scope_id":          "nonexistent-chapter",
				"answers":           []map[string]interface{}{},
			},
		},
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/offline/sync-attempts", bytes.NewReader(raw))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningGrammarOfflineSyncAttempts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Results []map[string]interface{} `json:"results"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Results) != 3 {
		t.Fatalf("results len = %d", len(body.Results))
	}
	for _, r := range body.Results[:2] {
		if r["synced"] == true {
			t.Fatalf("expected validation failure, got %+v", r)
		}
	}

	req = httptest.NewRequest(http.MethodPost, "/api/learning/grammar/offline/sync-attempts", bytes.NewReader([]byte("{")))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningGrammarOfflineSyncAttempts(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/learning/grammar/offline/sync-attempts", bytes.NewReader([]byte(`{"attempts":[]}`)))
	w = httptest.NewRecorder()
	router.handleLearningGrammarOfflineSyncAttempts(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	_ = sectionID
}

func TestGrammarOfflineSyncTrainingAttempts(t *testing.T) {
	router, userID, _ := setupGrammarOfflineTest(t)

	payload := map[string]interface{}{
		"attempts": []map[string]interface{}{
			{"client_attempt_id": "", "question_id": "q1"},
			{"client_attempt_id": "train-offline-1", "question_id": ""},
			{"client_attempt_id": "train-offline-2", "question_id": "missing-q", "answer": "x"},
		},
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/offline/sync-training-attempts", bytes.NewReader(raw))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningGrammarOfflineSyncTrainingAttempts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/learning/grammar/offline/sync-training-attempts", bytes.NewReader([]byte("invalid")))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningGrammarOfflineSyncTrainingAttempts(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json expected 400, got %d", w.Code)
	}
}
