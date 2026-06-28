package web

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

const linglowConvTelegramID int64 = 900301

type linglowConvFixture struct {
	userID       int64
	userCourseID int64
	scenarioCode string
	districtCode string
	courseCode   string
}

func insertLinglowConvFixture(t *testing.T, conn *sql.DB, userID int64) linglowConvFixture {
	t.Helper()
	const courseCode = "es_ru"
	const districtCode = "a0_spark_gate"
	const scenarioCode = "handler_test_cafe"

	var courseID, districtID, locationID int64
	if err := conn.QueryRow(`SELECT id FROM courses WHERE code = ?`, courseCode).Scan(&courseID); err != nil {
		t.Fatalf("get course: %v", err)
	}
	if err := conn.QueryRow(`SELECT id FROM districts WHERE course_id = ? AND code = ?`, courseID, districtCode).Scan(&districtID); err != nil {
		t.Fatalf("get district: %v", err)
	}
	if err := conn.QueryRow(`SELECT id FROM locations WHERE district_id = ? AND code = 'conversation'`, districtID).Scan(&locationID); err != nil {
		t.Fatalf("get location: %v", err)
	}

	var userCourseID int64
	if err := conn.QueryRow(`
		INSERT INTO user_courses (user_id, course_id, status) VALUES (?, ?, 'active') RETURNING id`,
		userID, courseID).Scan(&userCourseID); err != nil {
		t.Fatalf("insert user_course: %v", err)
	}

	var learningItemID int64
	if err := conn.QueryRow(`
		INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		VALUES (?, ?, ?, 'speaking_task', 'conversation_scenario', ?, 'Handler Test Cafe', 'A0', 'published')
		RETURNING id`, courseID, districtID, locationID, scenarioCode).Scan(&learningItemID); err != nil {
		t.Fatalf("insert learning_item: %v", err)
	}

	if _, err := conn.Exec(`
		INSERT INTO conversation_scenarios
			(course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
			 title, npc_name, npc_persona, scene_setup, is_quest, max_turns, token_budget, status)
		VALUES (?, ?, ?, ?, ?, 'cafe', 'A0', 'Handler Test Cafe', 'Mara', 'barista', 'A cozy cafe.', true, 16, 6000, 'active')`,
		courseID, districtID, locationID, learningItemID, scenarioCode); err != nil {
		t.Fatalf("insert scenario: %v", err)
	}

	var scenarioID int64
	if err := conn.QueryRow(`SELECT id FROM conversation_scenarios WHERE course_id = ? AND code = ?`, courseID, scenarioCode).Scan(&scenarioID); err != nil {
		t.Fatalf("get scenario id: %v", err)
	}

	for _, tk := range []struct {
		code     string
		order    int
		required bool
	}{
		{"greet", 0, true},
		{"order", 1, true},
		{"thank", 2, false},
	} {
		if _, err := conn.Exec(`
			INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
			VALUES (?, ?, ?, ?, ?, 'criteria')`, scenarioID, tk.code, tk.order, tk.required, tk.code); err != nil {
			t.Fatalf("insert task %s: %v", tk.code, err)
		}
	}

	return linglowConvFixture{
		userID:       userID,
		userCourseID: userCourseID,
		scenarioCode: scenarioCode,
		districtCode: districtCode,
		courseCode:   courseCode,
	}
}

func newConversationMockAI(t *testing.T) *ai.Service {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(body, &req)

		kind := "npc"
		if len(req.Messages) > 0 {
			sys := req.Messages[0].Content
			switch {
			case strings.Contains(sys, "quest evaluator"):
				kind = "quest"
			case strings.Contains(sys, "correction evaluator"):
				kind = "correction"
			}
		}

		var content string
		switch kind {
		case "quest":
			content = `{"completed_task_codes":["greet"],"all_done":false}`
		case "correction":
			content = `{"corrections":[{"original":"helo","corrected":"hello","explanation":"опечатка"}]}`
		default:
			content = "¡Hola! Bienvenido al café."
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"role": "assistant", "content": content}},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     10,
				"completion_tokens": 5,
				"total_tokens":      15,
			},
		})
	}))
	t.Cleanup(srv.Close)

	aiSvc := ai.NewService(srv.URL, "conv-model", "test-key", "prompt", zap.NewNop())
	aiSvc.SetConversationQuestPromptForCourse("es_ru", "quest evaluator")
	aiSvc.SetConversationCorrectionPromptForCourse("es_ru", "correction evaluator")
	aiSvc.SetConversationNPCPromptForCourse("es_ru", "npc character")
	return aiSvc
}

func setupLinglowConversationHandlersTest(t *testing.T) (*Router, linglowConvFixture) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)

	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(linglowConvTelegramID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	fx := insertLinglowConvFixture(t, conn, user.ID)
	router.aiService = newConversationMockAI(t)
	return router, fx
}

func startConversationSession(t *testing.T, router *Router, fx linglowConvFixture) int64 {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"course_code":   fx.courseCode,
		"scenario_code": fx.scenarioCode,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start session: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		SessionID int64 `json:"session_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if resp.SessionID == 0 {
		t.Fatal("expected session_id")
	}
	return resp.SessionID
}

func TestHandleLinglowConversationScenarios_OK(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", fx.districtCode, fx.courseCode), nil)
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationScenarios(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Scenarios []map[string]interface{} `json:"scenarios"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Scenarios) == 0 {
		t.Fatal("expected at least one scenario")
	}
	found := false
	for _, sc := range body.Scenarios {
		if sc["code"] == fx.scenarioCode {
			found = true
			if sc["npc_name"] != "Mara" {
				t.Errorf("npc_name = %v", sc["npc_name"])
			}
			tasks, _ := sc["tasks"].([]interface{})
			if len(tasks) != 3 {
				t.Errorf("tasks len = %d, want 3", len(tasks))
			}
		}
	}
	if !found {
		t.Fatalf("scenario %q not found in %+v", fx.scenarioCode, body.Scenarios)
	}
}

func TestHandleLinglowConversationScenarios_ValidationAndAuth(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/scenarios", nil)
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationScenarios(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing district_code: got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/scenarios?district_code="+fx.districtCode, nil)
	w = httptest.NewRecorder()
	router.handleLinglowConversationScenarios(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("missing auth: got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/scenarios?district_code="+fx.districtCode, nil)
	req = setUserIDInContext(req, fx.userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationScenarios(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("wrong method: got %d", w.Code)
	}
}

func TestHandleLinglowConversationSessions_StartAndResume(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)

	sessionID := startConversationSession(t, router, fx)

	req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader([]byte(fmt.Sprintf(
		`{"course_code":"%s","scenario_code":"%s"}`, fx.courseCode, fx.scenarioCode))))
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("resume session: got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		SessionID  int64                    `json:"session_id"`
		OpeningLine string                  `json:"opening_line"`
		Messages   []map[string]interface{} `json:"messages"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.SessionID != sessionID {
		t.Errorf("resume session_id = %d, want %d", resp.SessionID, sessionID)
	}
	if resp.OpeningLine == "" {
		t.Error("expected opening_line from AI mock")
	}
	if len(resp.Messages) == 0 {
		t.Error("expected stored opening message")
	}
}

func TestHandleLinglowConversationSessions_Validation(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader([]byte(`{`)))
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("bad json: got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader([]byte(`{"scenario_code":""}`)))
	req = setUserIDInContext(req, fx.userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing scenario_code: got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader([]byte(`{"scenario_code":"missing"}`)))
	req = setUserIDInContext(req, fx.userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("unknown scenario: got %d", w.Code)
	}

	router.aiService = nil
	req = httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader([]byte(`{"scenario_code":"x"}`)))
	req = setUserIDInContext(req, fx.userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("no ai service: got %d", w.Code)
	}
}

func TestHandleLinglowConversationSessionByID_Get(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)
	sessionID := startConversationSession(t, router, fx)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/sessions/%d", sessionID), nil)
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("get session: got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		ScenarioCode string `json:"scenario_code"`
		Status       string `json:"status"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ScenarioCode != fx.scenarioCode || resp.Status != "open" {
		t.Errorf("unexpected payload: %+v", resp)
	}
}

func TestHandleLinglowConversationSessionByID_Message(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)
	sessionID := startConversationSession(t, router, fx)

	body, _ := json.Marshal(map[string]string{"text": "helo"})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", sessionID), bytes.NewReader(body))
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("message: got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Reply       string                   `json:"reply"`
		Corrections []map[string]interface{} `json:"corrections"`
		Tasks       []map[string]interface{} `json:"tasks"`
		TurnCount   int                      `json:"turn_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Reply == "" {
		t.Error("expected NPC reply")
	}
	if len(resp.Corrections) == 0 {
		t.Error("expected corrections from mock")
	}
	if resp.TurnCount != 1 {
		t.Errorf("turn_count = %d, want 1", resp.TurnCount)
	}
	greetDone := false
	for _, task := range resp.Tasks {
		if task["code"] == "greet" && task["completed"] == true {
			greetDone = true
		}
	}
	if !greetDone {
		t.Errorf("expected greet task completed, tasks=%v", resp.Tasks)
	}
}

func TestHandleLinglowConversationSessionByID_MessageValidation(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)
	sessionID := startConversationSession(t, router, fx)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", sessionID), bytes.NewReader([]byte(`{"text":"  "}`)))
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("empty text: got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions/999999/message", bytes.NewReader([]byte(`{"text":"hi"}`)))
	req = setUserIDInContext(req, fx.userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("unknown session: got %d", w.Code)
	}
}

func TestHandleLinglowConversationSessionByID_End(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)
	sessionID := startConversationSession(t, router, fx)

	body, _ := json.Marshal(map[string]string{"status": "completed"})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/end", sessionID), bytes.NewReader(body))
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("end session: got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("status = %q, want completed", resp.Status)
	}

	// Closed session cannot accept messages.
	msgBody, _ := json.Marshal(map[string]string{"text": "hi again"})
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", sessionID), bytes.NewReader(msgBody))
	req = setUserIDInContext(req, fx.userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("message on closed session: got %d", w.Code)
	}
}

func TestHandleLinglowConversationSessionByID_Reset(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)
	sessionID := startConversationSession(t, router, fx)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/reset", sessionID), nil)
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("reset session: got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		SessionID int64  `json:"session_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.SessionID == sessionID {
		t.Errorf("expected new session id after reset, got same %d", sessionID)
	}
	if resp.Status != "open" {
		t.Errorf("status = %q, want open", resp.Status)
	}
}

func TestHandleLinglowConversationSessionByID_RoutingErrors(t *testing.T) {
	router, fx := setupLinglowConversationHandlersTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/sessions/not-a-number", nil)
	req = setUserIDInContext(req, fx.userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("invalid id: got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/linglow/conversation/sessions/1", nil)
	req = setUserIDInContext(req, fx.userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("unsupported method: got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/sessions/", nil)
	req = setUserIDInContext(req, fx.userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("empty id: got %d", w.Code)
	}
}

func TestFilterCorrectionsToMessage_DropsStale(t *testing.T) {
	corrections := []ai.ChatCorrection{
		{Original: "helo", Corrected: "hello", Explanation: "typo"},
		{Original: "old mistake", Corrected: "fixed", Explanation: "stale"},
	}
	out := filterCorrectionsToMessage(corrections, "helo there")
	if len(out) != 1 {
		t.Fatalf("expected 1 correction kept, got %d: %+v", len(out), out)
	}
	if out[0].Original != "helo" {
		t.Errorf("kept correction = %+v", out[0])
	}
}

func TestScenarioLockState_Cooldown(t *testing.T) {
	router, _ := setupLinglowConversationHandlersTest(t)

	sc := &repository.ConversationScenario{PrerequisiteCode: "prev"}
	passedCodes := map[string]bool{"prev": true}
	passedAt := map[string]time.Time{"prev": time.Now().Add(-30 * time.Minute)}
	locked, until := router.scenarioLockState(sc, passedCodes, passedAt, time.Now())
	if !locked || until == nil {
		t.Fatalf("expected cooldown lock, locked=%v until=%v", locked, until)
	}
}
