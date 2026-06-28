package web

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupSpeakingTest(t *testing.T) (*Router, *sql.DB, int64) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{
		Learning: config.DefaultLearningConfig(),
		Speaking: config.SpeakingConfig{
			Enabled:            true,
			SessionTaskCount:   2,
			MaxAudioMB:         2,
			MaxAttemptsDefault: 3,
		},
		WebApp: config.WebAppConfig{JWTSecret: "test-speaking-secret"},
	}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(900001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := userRepo.UpdateSubscriptionTier(user.ID, models.TierPro); err != nil {
		t.Fatalf("set tier: %v", err)
	}
	router.SetDependencies(userRepo, nil, nil, nil, "bot-token")
	courseRepo := repository.NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(t.Context(), cfg.Learning); err != nil {
		t.Fatalf("backfill user courses: %v", err)
	}
	return router, conn, user.ID
}

func seedSpeakingCatalog(t *testing.T, db *sql.DB, categoryID string, taskIDs []string) {
	t.Helper()
	rawIDs, err := json.Marshal(taskIDs)
	if err != nil {
		t.Fatalf("marshal task ids: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO speaking_categories (category_id, title, level, sort_order, task_ids)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (category_id) DO UPDATE SET task_ids = excluded.task_ids`,
		categoryID, "Test speaking", "A0", 1, string(rawIDs),
	); err != nil {
		t.Fatalf("insert category: %v", err)
	}
	for i, taskID := range taskIDs {
		doc := map[string]interface{}{
			"id": taskID, "category_id": categoryID, "level": "A0", "type": "phrase",
			"target_language": "en", "title": "Hello", "prompt_ru": "Скажи привет",
			"display_text": "hello", "max_attempts": 3, "order": i,
		}
		raw, _ := json.Marshal(doc)
		if _, err := db.Exec(`
			INSERT INTO speaking_tasks (task_id, category_id, title, level, task_type, target_language, sort_order, task_json)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (task_id) DO UPDATE SET task_json = excluded.task_json`,
			taskID, categoryID, "Hello", "A0", "phrase", "en", i, string(raw),
		); err != nil {
			t.Fatalf("insert task %s: %v", taskID, err)
		}
	}
}

func TestSpeakingAppendUniqueString(t *testing.T) {
	list := appendUniqueString(nil, "A0")
	list = appendUniqueString(list, "A1")
	list = appendUniqueString(list, "A0")
	if len(list) != 2 || list[0] != "A0" || list[1] != "A1" {
		t.Fatalf("unexpected list: %v", list)
	}
}

func TestSpeakingSyncCatalogNilSafe(t *testing.T) {
	var nilRouter *Router
	if err := nilRouter.SyncSpeakingCatalogFromBundle(context.Background()); err != nil {
		t.Fatalf("nil router: %v", err)
	}
	router := NewRouter(zap.NewNop(), &config.Config{}, nil, nil, nil, nil, nil)
	if err := router.SyncSpeakingCatalogFromBundle(context.Background()); err != nil {
		t.Fatalf("nil repo: %v", err)
	}
}

func TestSpeakingAvailability(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		noUser   bool
		enabled  bool
		seed     bool
		wantCode int
	}{
		{name: "ok", method: http.MethodGet, enabled: true, seed: true, wantCode: http.StatusOK},
		{name: "unauthorized", method: http.MethodGet, noUser: true, enabled: true, wantCode: http.StatusUnauthorized},
		{name: "method not allowed", method: http.MethodPost, enabled: true, wantCode: http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router, conn, userID := setupSpeakingTest(t)
			router.config.Speaking.Enabled = tc.enabled
			if tc.seed {
				seedSpeakingCatalog(t, conn, "spk_cat_a", []string{"spk_task_a"})
				seedSpeakingCatalog(t, conn, "spk_cat_b", []string{"spk_task_b"})
				_, _ = conn.Exec(`UPDATE speaking_categories SET level = 'A1' WHERE category_id = 'spk_cat_b'`)
			}
			req := httptest.NewRequest(tc.method, "/api/learning/speaking/availability", nil)
			if !tc.noUser {
				req = setUserIDInContext(req, userID)
			}
			w := httptest.NewRecorder()
			router.handleLearningSpeakingAvailability(w, req)
			if w.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d: %s", tc.wantCode, w.Code, w.Body.String())
			}
			if tc.wantCode != http.StatusOK {
				return
			}
			var body struct {
				Available bool     `json:"available"`
				CanAccess bool     `json:"can_access"`
				Levels    []string `json:"levels"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !body.Available || !body.CanAccess {
				t.Fatalf("expected available pro access, got %+v", body)
			}
			if len(body.Levels) != 2 {
				t.Fatalf("levels = %v, want 2 unique levels", body.Levels)
			}
		})
	}
}

func TestSpeakingCategories(t *testing.T) {
	router, conn, userID := setupSpeakingTest(t)
	seedSpeakingCatalog(t, conn, "spk_list_cat", []string{"spk_list_task"})

	req := httptest.NewRequest(http.MethodGet, "/api/learning/speaking/categories", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningSpeakingCategories(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	router.speakingCatalogRepo = nil
	req = httptest.NewRequest(http.MethodGet, "/api/learning/speaking/categories", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingCategories(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("nil repo expected 404, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/learning/speaking/categories", nil)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingCategories(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestSpeakingCategoryTasks(t *testing.T) {
	router, conn, userID := setupSpeakingTest(t)
	categoryID := "spk_tasks_cat"
	seedSpeakingCatalog(t, conn, categoryID, []string{"spk_tasks_task"})

	req := httptest.NewRequest(http.MethodGet, "/api/learning/speaking/categories/"+categoryID+"/tasks", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningSpeakingCategoryTasks(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	badPaths := []string{
		"/api/learning/speaking/categories//tasks",
		"/api/learning/speaking/categories/x/y/tasks",
		"/api/learning/speaking/categories/x",
	}
	for _, path := range badPaths {
		req = httptest.NewRequest(http.MethodGet, path, nil)
		req = setUserIDInContext(req, userID)
		w = httptest.NewRecorder()
		router.handleLearningSpeakingCategoryTasks(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("path %q expected 400, got %d", path, w.Code)
		}
	}
}

func TestSpeakingSessions(t *testing.T) {
	router, conn, userID := setupSpeakingTest(t)
	categoryID := "spk_session_cat"
	seedSpeakingCatalog(t, conn, categoryID, []string{"spk_s1", "spk_s2", "spk_s3"})

	body := map[string]string{"category_id": categoryID}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/speaking/sessions", bytes.NewReader(raw))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningSpeakingSessions(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create session: %d %s", w.Code, w.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if created["total_tasks"].(float64) != 2 {
		t.Fatalf("session task count = %v, want 2", created["total_tasks"])
	}

	req = httptest.NewRequest(http.MethodPost, "/api/learning/speaking/sessions", bytes.NewReader([]byte("{")))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessions(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/learning/speaking/sessions", bytes.NewReader([]byte(`{"category_id":""}`)))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessions(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty category expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/learning/speaking/sessions", bytes.NewReader([]byte(`{"category_id":"missing"}`)))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessions(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing category expected 404, got %d", w.Code)
	}
}

func TestSpeakingSessionByID(t *testing.T) {
	router, conn, userID := setupSpeakingTest(t)
	categoryID := "spk_by_id_cat"
	taskID := "spk_by_id_task"
	seedSpeakingCatalog(t, conn, categoryID, []string{taskID})

	session, err := router.speakingSessionRepo.CreateSession(userID, categoryID, []string{taskID})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/learning/speaking/sessions/%d", session.ID), nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get session: %d %s", w.Code, w.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["current_task"] == nil {
		t.Fatal("expected current_task in session json")
	}

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/learning/speaking/sessions/%d/next", session.ID), nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("next: %d %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/learning/speaking/sessions/not-a-id", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid id expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/learning/speaking/sessions/%d", session.ID), nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("patch expected 405, got %d", w.Code)
	}
}

func speakingSubmitMultipart(t *testing.T, taskID string, audio []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("task_id", taskID)
	_ = writer.WriteField("mode", "initial")
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="audio"; filename="clip.wav"`)
	header.Set("Content-Type", "audio/wav")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("CreatePart: %v", err)
	}
	if _, err := part.Write(audio); err != nil {
		t.Fatalf("write audio: %v", err)
	}
	_ = writer.Close()
	return &body, writer.FormDataContentType()
}

func TestSpeakingSubmit(t *testing.T) {
	router, conn, userID := setupSpeakingTest(t)
	categoryID := "spk_submit_cat"
	taskID := "spk_submit_task"
	seedSpeakingCatalog(t, conn, categoryID, []string{taskID})
	session, err := router.speakingSessionRepo.CreateSession(userID, categoryID, []string{taskID})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	submitURL := fmt.Sprintf("/api/learning/speaking/sessions/%d/submit", session.ID)

	req := httptest.NewRequest(http.MethodPost, submitURL, strings.NewReader("not-multipart"))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid multipart expected 400, got %d", w.Code)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("task_id", taskID)
	_ = writer.Close()
	req = httptest.NewRequest(http.MethodPost, submitURL, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing audio expected 400, got %d: %s", w.Code, w.Body.String())
	}

	submitBody, contentType := speakingSubmitMultipart(t, taskID, []byte{0, 1, 2, 3})
	req = httptest.NewRequest(http.MethodPost, submitURL, submitBody)
	req.Header.Set("Content-Type", contentType)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("no evaluator expected 503, got %d: %s", w.Code, w.Body.String())
	}

	evalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": `{"understood_answer":"hello","meaning_score":4,"grammar_score":4,"pronunciation_score":4,"fluency_score":4,"is_acceptable":true,"audio_quality":"clear","short_feedback_ru":"Отлично","better_version":"","repeat_task":""}`}},
			},
		})
	}))
	defer evalServer.Close()

	router.config.Speaking.EvalBaseURL = evalServer.URL
	router.config.Speaking.EvalAPIKey = "test-key"
	router.config.Speaking.EvalModel = "test-model"
	router.config.Speaking.Enabled = true
	router.SetSpeakingEvaluator(service.NewSpeakingEvaluatorService(router.config, router.logger))

	submitBody, contentType = speakingSubmitMultipart(t, taskID, []byte{0, 1, 2, 3})
	req = httptest.NewRequest(http.MethodPost, submitURL, submitBody)
	req.Header.Set("Content-Type", contentType)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("submit expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode submit: %v", err)
	}
	if result["is_acceptable"] != true {
		t.Fatalf("expected acceptable result, got %+v", result)
	}
}

func TestSpeakingSubmit_EvalFailure(t *testing.T) {
	router, conn, userID := setupSpeakingTest(t)
	categoryID := "spk_eval_fail_cat"
	taskID := "spk_eval_fail_task"
	seedSpeakingCatalog(t, conn, categoryID, []string{taskID})
	session, err := router.speakingSessionRepo.CreateSession(userID, categoryID, []string{taskID})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	evalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer evalServer.Close()
	router.config.Speaking.EvalBaseURL = evalServer.URL
	router.config.Speaking.EvalAPIKey = "test-key"
	router.config.Speaking.EvalModel = "test-model"
	router.config.Speaking.Enabled = true
	router.SetSpeakingEvaluator(service.NewSpeakingEvaluatorService(router.config, router.logger))

	body, contentType := speakingSubmitMultipart(t, taskID, []byte{1, 2, 3})
	submitURL := fmt.Sprintf("/api/learning/speaking/sessions/%d/submit", session.ID)
	req := httptest.NewRequest(http.MethodPost, submitURL, body)
	req.Header.Set("Content-Type", contentType)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLearningSpeakingSessionByID(w, req)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("eval failure expected 502, got %d: %s", w.Code, w.Body.String())
	}
}
