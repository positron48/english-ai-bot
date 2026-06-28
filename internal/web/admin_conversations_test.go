package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupAdminConversationsTest(t *testing.T) (*Router, *repository.ConversationRepository) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	return router, router.conversationRepo
}

func TestAdminConversationsScenarios(t *testing.T) {
	router, repo := setupAdminConversationsTest(t)
	ctx := context.Background()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/conversations/scenarios?course_code=es_ru", nil)
	w := httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var listed struct {
		CourseCode string                   `json:"course_code"`
		Scenarios  []map[string]interface{} `json:"scenarios"`
		Levels     []map[string]string      `json:"levels"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if listed.CourseCode != "es_ru" || len(listed.Levels) == 0 {
		t.Fatalf("unexpected list payload: %+v", listed)
	}

	createBody := map[string]interface{}{
		"code": "web_admin_cafe", "title": "Web Admin Cafe", "cefr_level": "A0",
		"npc_name": "Mara", "scene_setup": "At the cafe", "status": "draft",
	}
	raw, _ := json.Marshal(createBody)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios?course_code=es_ru", bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	scenarioID := int64(created["id"].(float64))

	req = httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios?course_code=es_ru", bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate expected 409, got %d: %s", w.Code, w.Body.String())
	}

	updateBody := map[string]interface{}{
		"code": "web_admin_cafe", "title": "Updated Cafe", "cefr_level": "A0",
		"npc_name": "Mara", "scene_setup": "Updated scene", "status": "active",
	}
	raw, _ = json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"?course_code=es_ru", bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update expected 200, got %d: %s", w.Code, w.Body.String())
	}

	taskBody := map[string]interface{}{
		"code": "greet", "title": "Greet", "completion_criteria": "user greets npc",
	}
	raw, _ = json.Marshal(taskBody)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"/tasks", bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create task expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var taskCreated map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &taskCreated); err != nil {
		t.Fatalf("decode task: %v", err)
	}
	taskID := int64(taskCreated["id"].(float64))

	updateTaskBody := map[string]interface{}{
		"code": "greet", "title": "Greet warmly", "completion_criteria": "warm greeting", "is_required": true,
	}
	raw, _ = json.Marshal(updateTaskBody)
	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/tasks/"+strconv.FormatInt(taskID, 10), bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationTaskByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update task expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/npcs/mara/image?course_code=es_ru", bytes.NewReader([]byte(`{"image_url":"https://cdn.example/mara.png"}`)))
	w = httptest.NewRecorder()
	router.handleAdminConversationNPCImage(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("npc image expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/admin/conversations/tasks/"+strconv.FormatInt(taskID, 10), nil)
	w = httptest.NewRecorder()
	router.handleAdminConversationTaskByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete task expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10), nil)
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete scenario expected 200, got %d: %s", w.Code, w.Body.String())
	}

	images, err := repo.GetNPCImages(ctx, mustCourseID(t, repo, "es_ru"))
	if err != nil {
		t.Fatalf("GetNPCImages: %v", err)
	}
	if images["mara"] != "https://cdn.example/mara.png" {
		t.Fatalf("npc image not saved: %+v", images)
	}
}

func mustCourseID(t *testing.T, repo *repository.ConversationRepository, code string) int64 {
	t.Helper()
	id, err := repo.CourseIDByCode(context.Background(), code)
	if err != nil {
		t.Fatalf("CourseIDByCode: %v", err)
	}
	return id
}

func TestAdminConversationsValidationErrors(t *testing.T) {
	router, _ := setupAdminConversationsTest(t)

	router.conversationRepo = nil
	req := httptest.NewRequest(http.MethodGet, "/api/admin/conversations/scenarios", nil)
	w := httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil repo expected 503, got %d", w.Code)
	}

	router, _ = setupAdminConversationsTest(t)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios?course_code=es_ru", bytes.NewReader([]byte("{")))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios?course_code=es_ru", bytes.NewReader([]byte(`{"code":"","title":"","cefr_level":""}`)))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing fields expected 400, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/conversations/scenarios?course_code=missing_course", nil)
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown course expected 404, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/scenarios/not-id", bytes.NewReader([]byte(`{}`)))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid scenario id expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/npcs/mara/wrong", bytes.NewReader([]byte(`{"image_url":"/x"}`)))
	w = httptest.NewRecorder()
	router.handleAdminConversationNPCImage(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid npc path expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPatch, "/api/admin/conversations/tasks/1", nil)
	w = httptest.NewRecorder()
	router.handleAdminConversationTaskByID(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("patch task expected 405, got %d", w.Code)
	}
}
