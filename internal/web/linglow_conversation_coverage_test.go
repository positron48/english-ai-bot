package web

import (
	"bytes"
	"context"
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

const (
	linglowCovTelegramID     int64 = 900401
	linglowCovLegacyTelegram int64 = 900402
	linglowCovChainTelegram  int64 = 900403
)

type linglowCovFixture struct {
	userID       int64
	userCourseID int64
	scenarioID   int64
	scenarioCode string
	districtCode string
	courseCode   string
	courseID     int64
}

func setupLinglowCovRouter(t *testing.T, telegramID int64, learning config.LearningConfig) (*Router, *sql.DB) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: learning}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	return router, conn
}

func insertLinglowCovScenario(t *testing.T, conn *sql.DB, userID int64, courseCode, districtCode, scenarioCode string, opts ...func(*scenarioInsertOpts)) linglowCovFixture {
	t.Helper()
	o := scenarioInsertOpts{
		npcName:    "Mara",
		npcCode:    "mara_cov",
		isQuest:    true,
		maxTurns:   16,
		tokenBudget: 6000,
	}
	for _, fn := range opts {
		fn(&o)
	}

	var courseID, districtID, locationID int64
	if err := conn.QueryRow(`SELECT id FROM courses WHERE code = ?`, courseCode).Scan(&courseID); err != nil {
		t.Fatalf("course %s: %v", courseCode, err)
	}
	if err := conn.QueryRow(`SELECT id FROM districts WHERE course_id = ? AND code = ?`, courseID, districtCode).Scan(&districtID); err != nil {
		t.Fatalf("district %s: %v", districtCode, err)
	}
	if err := conn.QueryRow(`SELECT id FROM locations WHERE district_id = ? AND code = 'conversation'`, districtID).Scan(&locationID); err != nil {
		t.Fatalf("location: %v", err)
	}

	var userCourseID int64
	err := conn.QueryRow(`
		SELECT id FROM user_courses WHERE user_id = ? AND course_id = ?`, userID, courseID).Scan(&userCourseID)
	if err == sql.ErrNoRows {
		if err := conn.QueryRow(`
			INSERT INTO user_courses (user_id, course_id, status) VALUES (?, ?, 'active') RETURNING id`,
			userID, courseID).Scan(&userCourseID); err != nil {
			t.Fatalf("user_course: %v", err)
		}
	} else if err != nil {
		t.Fatalf("user_course lookup: %v", err)
	}

	var learningItemID int64
	if err := conn.QueryRow(`
		INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		VALUES (?, ?, ?, 'speaking_task', 'conversation_scenario', ?, ?, 'A0', 'published')
		RETURNING id`, courseID, districtID, locationID, scenarioCode, o.title).Scan(&learningItemID); err != nil {
		t.Fatalf("learning_item: %v", err)
	}

	if _, err := conn.Exec(`
		INSERT INTO conversation_scenarios
			(course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
			 title, npc_name, npc_persona, npc_code, scene_setup, is_quest, max_turns, token_budget,
			 prerequisite_code, status)
		VALUES (?, ?, ?, ?, ?, 'cafe', 'A0', ?, ?, 'barista', ?, 'A cozy cafe.', ?, ?, ?, ?, 'active')`,
		courseID, districtID, locationID, learningItemID, scenarioCode, o.title, o.npcName, o.npcCode,
		o.isQuest, o.maxTurns, o.tokenBudget, o.prerequisiteCode); err != nil {
		t.Fatalf("scenario: %v", err)
	}

	var scenarioID int64
	if err := conn.QueryRow(`SELECT id FROM conversation_scenarios WHERE course_id = ? AND code = ?`, courseID, scenarioCode).Scan(&scenarioID); err != nil {
		t.Fatalf("scenario id: %v", err)
	}

	for _, tk := range o.tasks {
		if _, err := conn.Exec(`
			INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
			VALUES (?, ?, ?, ?, ?, 'criteria')`, scenarioID, tk.code, tk.order, tk.required, tk.code); err != nil {
			t.Fatalf("task %s: %v", tk.code, err)
		}
	}

	if o.npcImageURL != "" {
		if _, err := conn.Exec(`
			INSERT INTO conversation_npcs (course_id, npc_code, image_url) VALUES (?, ?, ?)
			ON CONFLICT (course_id, npc_code) DO UPDATE SET image_url = EXCLUDED.image_url`,
			courseID, o.npcCode, o.npcImageURL); err != nil {
			t.Fatalf("npc image: %v", err)
		}
	}

	return linglowCovFixture{
		userID:       userID,
		userCourseID: userCourseID,
		scenarioID:   scenarioID,
		scenarioCode: scenarioCode,
		districtCode: districtCode,
		courseCode:   courseCode,
		courseID:     courseID,
	}
}

type scenarioInsertOpts struct {
	title             string
	npcName           string
	npcCode           string
	isQuest           bool
	maxTurns          int
	tokenBudget       int
	prerequisiteCode  string
	npcImageURL       string
	tasks             []struct {
		code     string
		order    int
		required bool
	}
}

func linglowCovUser(t *testing.T, conn *sql.DB, telegramID int64) int64 {
	t.Helper()
	ur := repository.NewUserRepository(conn, zap.NewNop())
	user, err := ur.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	return user.ID
}

func newCovSplitMockAI(t *testing.T, reply string) *ai.Service {
	t.Helper()
	if reply == "" {
		reply = "¡Hola!"
	}
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
		content := reply
		switch kind {
		case "quest":
			content = `{"completed_task_codes":["greet"],"all_done":false}`
		case "correction":
			content = `{"corrections":[]}`
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"role": "assistant", "content": content}},
			},
			"usage": map[string]interface{}{"prompt_tokens": 10, "completion_tokens": 5},
		})
	}))
	t.Cleanup(srv.Close)
	svc := ai.NewService(srv.URL, "m", "k", "p", zap.NewNop())
	svc.SetConversationQuestPromptForCourse("es_ru", "quest evaluator")
	svc.SetConversationCorrectionPromptForCourse("es_ru", "correction evaluator")
	svc.SetConversationNPCPromptForCourse("es_ru", "npc character")
	return svc
}

func newCovLegacyMockAI(t *testing.T, visible string, withUsage bool) *ai.Service {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := visible + "\n###CONTROL###\n{\"completed_task_codes\":[],\"all_done\":false,\"corrections\":[]}"
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"role": "assistant", "content": raw}},
			},
		}
		if withUsage {
			resp["usage"] = map[string]interface{}{"prompt_tokens": 8, "completion_tokens": 4}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)
	svc := ai.NewService(srv.URL, "m", "k", "p", zap.NewNop())
	svc.SetConversationPromptForCourse("en_ru", "legacy npc prompt")
	return svc
}

func newCovFailingMockAI(t *testing.T) *ai.Service {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	return ai.NewService(srv.URL, "m", "k", "p", zap.NewNop())
}

func newCovEmptyMockAI(t *testing.T) *ai.Service {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"role": "assistant", "content": "   "}},
			},
		})
	}))
	t.Cleanup(srv.Close)
	svc := ai.NewService(srv.URL, "m", "k", "p", zap.NewNop())
	svc.SetConversationNPCPromptForCourse("es_ru", "npc")
	return svc
}

func covBreakRouterDB(t *testing.T, router *Router) {
	t.Helper()
	broken := newBrokenDB(t)
	router.db = broken
	router.conversationRepo = repository.NewConversationRepository(broken, router.logger)
	router.courseRepo = repository.NewCourseRepository(broken, router.logger)
	router.linglowEventRepo = repository.NewLinglowEventRepository(broken)
}

func TestLinglowConversationCoverage_ScenarioLockState(t *testing.T) {
	router, _ := setupLinglowCovRouter(t, linglowCovTelegramID, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	now := time.Now()

	cases := []struct {
		name     string
		sc       *repository.ConversationScenario
		passed   map[string]bool
		passedAt map[string]time.Time
		wantLock bool
		wantCD   bool
	}{
		{
			name:     "no prerequisite",
			sc:       &repository.ConversationScenario{},
			wantLock: false,
		},
		{
			name:     "prerequisite not passed",
			sc:       &repository.ConversationScenario{PrerequisiteCode: "prev"},
			passed:   map[string]bool{},
			wantLock: true,
		},
		{
			name:     "cooldown active",
			sc:       &repository.ConversationScenario{PrerequisiteCode: "prev"},
			passed:   map[string]bool{"prev": true},
			passedAt: map[string]time.Time{"prev": now.Add(-30 * time.Minute)},
			wantLock: true,
			wantCD:   true,
		},
		{
			name:     "cooldown elapsed",
			sc:       &repository.ConversationScenario{PrerequisiteCode: "prev"},
			passed:   map[string]bool{"prev": true},
			passedAt: map[string]time.Time{"prev": now.Add(-2 * time.Hour)},
			wantLock: false,
		},
		{
			name:     "passed without timestamp",
			sc:       &repository.ConversationScenario{PrerequisiteCode: "prev"},
			passed:   map[string]bool{"prev": true},
			passedAt: map[string]time.Time{},
			wantLock: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			locked, until := router.scenarioLockState(tc.sc, tc.passed, tc.passedAt, now)
			if locked != tc.wantLock {
				t.Fatalf("locked=%v want %v", locked, tc.wantLock)
			}
			if (until != nil) != tc.wantCD {
				t.Fatalf("cooldown until present=%v want %v", until != nil, tc.wantCD)
			}
		})
	}
}

func TestLinglowConversationCoverage_DerivedSessionStatus(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovTelegramID, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovTelegramID)
	fx := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", "cov_status_scenario", func(o *scenarioInsertOpts) {
		o.title = "Status Test"
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}}
	})
	ctx := context.Background()
	tasks, _ := router.conversationRepo.ListTasks(ctx, fx.scenarioID)

	if got := router.derivedSessionStatus(ctx, fx.userCourseID, fx.scenarioID, tasks, false, true); got != "completed" {
		t.Fatalf("allTasksDone: got %q", got)
	}
	if got := router.derivedSessionStatus(ctx, fx.userCourseID, fx.scenarioID, tasks, true, false); got != "passed" {
		t.Fatalf("questPassed: got %q", got)
	}

	sess, _, err := router.conversationRepo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
	if got := router.derivedSessionStatus(ctx, fx.userCourseID, fx.scenarioID, tasks, false, false); got != "open" {
		t.Fatalf("latest session: got %q want open", got)
	}
	_ = router.conversationRepo.CloseSession(ctx, sess.ID, fx.userCourseID, "abandoned")
	if got := router.latestSessionStatus(ctx, fx.userCourseID, fx.scenarioID); got != "abandoned" {
		t.Fatalf("latest after close: got %q", got)
	}
	if got := router.latestSessionStatus(ctx, fx.userCourseID, 999999); got != "" {
		t.Fatalf("missing session: got %q", got)
	}
}

func TestLinglowConversationCoverage_ScenarioProgressFlags(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovTelegramID, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovTelegramID)
	fx := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", "cov_progress_scenario", func(o *scenarioInsertOpts) {
		o.title = "Progress"
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}, {"bye", 1, false}}
	})
	ctx := context.Background()
	tasks, _ := router.conversationRepo.ListTasks(ctx, fx.scenarioID)

	questPassed, fullyDone := router.scenarioProgressFlags(ctx, fx.userCourseID, fx.scenarioID, tasks)
	if questPassed || fullyDone {
		t.Fatalf("no session yet: quest=%v fully=%v", questPassed, fullyDone)
	}

	sess, _, err := router.conversationRepo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	taskIDs := map[string]int64{}
	for _, tk := range tasks {
		taskIDs[tk.Code] = tk.ID
	}
	_ = router.conversationRepo.MarkTasksCompleted(ctx, sess.ID, taskIDs, []string{"greet"}, 1)
	questPassed, fullyDone = router.scenarioProgressFlags(ctx, fx.userCourseID, fx.scenarioID, tasks)
	if !questPassed || fullyDone {
		t.Fatalf("required only: quest=%v fully=%v", questPassed, fullyDone)
	}

	_ = router.conversationRepo.CloseSession(ctx, sess.ID, fx.userCourseID, "completed")
	questPassed, fullyDone = router.scenarioProgressFlags(ctx, fx.userCourseID, fx.scenarioID, tasks)
	if !fullyDone {
		t.Fatalf("completed session should set fullyDone, quest=%v fully=%v", questPassed, fullyDone)
	}

	// ScenarioEverPassed via exercise_attempt without active session tasks.
	userID2 := linglowCovUser(t, conn, linglowCovTelegramID+100)
	fx2 := insertLinglowCovScenario(t, conn, userID2, "es_ru", "a0_spark_gate", "cov_ever_passed", func(o *scenarioInsertOpts) {
		o.title = "Ever Passed"
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}}
	})
	tasks2, _ := router.conversationRepo.ListTasks(ctx, fx2.scenarioID)
	sess2, _, _ := router.conversationRepo.StartSession(ctx, fx2.userCourseID, fx2.scenarioID)
	_ = router.conversationRepo.RecordQuestCompletion(ctx, fx2.userCourseID, sql.NullInt64{}, fx2.scenarioCode, sess2.ID)
	_ = router.conversationRepo.CloseSession(ctx, sess2.ID, fx2.userCourseID, "abandoned")
	questPassed, _ = router.scenarioProgressFlags(ctx, fx2.userCourseID, fx2.scenarioID, tasks2)
	if !questPassed {
		t.Fatal("expected ScenarioEverPassed questPassed")
	}
}

func TestLinglowConversationCoverage_ScenariosHandler(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovChainTelegram, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovChainTelegram)
	prevCode := "cov_chain_prev"
	nextCode := "cov_chain_next"
	insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", prevCode, func(o *scenarioInsertOpts) {
		o.title = "Prev"
		o.npcImageURL = "https://example.com/mara.png"
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}}
	})
	fxNext := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", nextCode, func(o *scenarioInsertOpts) {
		o.title = "Next"
		o.prerequisiteCode = prevCode
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"order", 0, true}}
	})

	t.Run("nil repo", func(t *testing.T) {
		r := *router
		r.conversationRepo = nil
		req := httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/scenarios?district_code="+fxNext.districtCode, nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		r.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("course not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/scenarios?district_code="+fxNext.districtCode+"&course_code=missing_course", nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("district not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/scenarios?district_code=no_such_district&course_code="+fxNext.courseCode, nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("locked and cooldown", func(t *testing.T) {
		ctx := context.Background()
		prev, _ := router.conversationRepo.GetScenarioByCode(ctx, fxNext.courseID, prevCode)
		sess, _, _ := router.conversationRepo.StartSession(ctx, fxNext.userCourseID, prev.ID)
		_ = router.conversationRepo.RecordQuestCompletion(ctx, fxNext.userCourseID, sql.NullInt64{}, prevCode, sess.ID)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", fxNext.districtCode, fxNext.courseCode), nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("got %d: %s", w.Code, w.Body.String())
		}
		var body struct {
			Scenarios []map[string]interface{} `json:"scenarios"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		for _, sc := range body.Scenarios {
			if sc["code"] == nextCode {
				if sc["locked"] != true {
					t.Fatalf("expected locked next scenario: %+v", sc)
				}
				if sc["cooldown_until"] == nil {
					t.Fatal("expected cooldown_until")
				}
			}
			if sc["code"] == prevCode && sc["npc_image_url"] != "https://example.com/mara.png" {
				t.Fatalf("npc image: %+v", sc)
			}
		}
	})

	t.Run("npc images db error continues", func(t *testing.T) {
		if _, err := conn.Exec(`ALTER TABLE conversation_npcs RENAME TO conversation_npcs_cov_bak`); err != nil {
			t.Fatalf("rename npcs: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.Exec(`ALTER TABLE conversation_npcs_cov_bak RENAME TO conversation_npcs`)
		})
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", fxNext.districtCode, fxNext.courseCode), nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("list tasks error", func(t *testing.T) {
		if _, err := conn.Exec(`ALTER TABLE conversation_tasks RENAME TO conversation_tasks_cov_bak`); err != nil {
			t.Fatalf("rename tasks: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.Exec(`ALTER TABLE conversation_tasks_cov_bak RENAME TO conversation_tasks`)
		})
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", fxNext.districtCode, fxNext.courseCode), nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("broken db internal errors", func(t *testing.T) {
		r, _ := setupLinglowCovRouter(t, linglowCovChainTelegram, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		covBreakRouterDB(t, r)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=es_ru", fxNext.districtCode), nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		r.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("course resolve: got %d", w.Code)
		}
	})
}

func TestLinglowConversationCoverage_SessionsHandler(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovChainTelegram+10, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovChainTelegram+10)
	prevCode := "cov_sess_prev"
	nextCode := "cov_sess_next"
	insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", prevCode, func(o *scenarioInsertOpts) {
		o.title = "Prev"
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}}
	})
	fxNext := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", nextCode, func(o *scenarioInsertOpts) {
		o.title = "Next"
		o.prerequisiteCode = prevCode
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"order", 0, true}}
	})
	router.aiService = newCovSplitMockAI(t, "Hola")

	t.Run("method and auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/sessions", nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("got %d", w.Code)
		}
		req = httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader([]byte(`{"scenario_code":"x"}`)))
		w = httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("nil repo", func(t *testing.T) {
		r := *router
		r.conversationRepo = nil
		body, _ := json.Marshal(map[string]string{"scenario_code": fxNext.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("prerequisite locked", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"course_code": fxNext.courseCode, "scenario_code": nextCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("locked: got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("prerequisite cooldown", func(t *testing.T) {
		ctx := context.Background()
		prev, _ := router.conversationRepo.GetScenarioByCode(ctx, fxNext.courseID, prevCode)
		sess, _, _ := router.conversationRepo.StartSession(ctx, fxNext.userCourseID, prev.ID)
		_ = router.conversationRepo.RecordQuestCompletion(ctx, fxNext.userCourseID, sql.NullInt64{}, prevCode, sess.ID)

		body, _ := json.Marshal(map[string]string{"course_code": fxNext.courseCode, "scenario_code": nextCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "cooldown") {
			t.Fatalf("cooldown: got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("course not found", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"course_code": "nope", "scenario_code": prevCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("start ok after cooldown", func(t *testing.T) {
		_, _ = conn.Exec(`
			UPDATE exercise_attempts SET answered_at = answered_at - INTERVAL '2 hours'
			WHERE user_course_id = ? AND result_json->>'scenario_code' = ?`, fxNext.userCourseID, prevCode)
		body, _ := json.Marshal(map[string]string{"course_code": fxNext.courseCode, "scenario_code": nextCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("broken db", func(t *testing.T) {
		r, _ := setupLinglowCovRouter(t, linglowCovChainTelegram+10, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		r.aiService = newCovSplitMockAI(t, "Hi")
		covBreakRouterDB(t, r)
		body, _ := json.Marshal(map[string]string{"course_code": "es_ru", "scenario_code": prevCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", w.Code)
		}
	})
}

func TestLinglowConversationCoverage_GenerateOpeningLine(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovTelegramID, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovTelegramID)
	fx := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", "cov_opening", func(o *scenarioInsertOpts) {
		o.title = "Opening"
	})
	ctx := context.Background()
	scenario, _ := router.conversationRepo.GetScenarioByID(ctx, fx.scenarioID)
	sess, _, _ := router.conversationRepo.StartSession(ctx, fx.userCourseID, fx.scenarioID)

	t.Run("nil ai", func(t *testing.T) {
		r := *router
		r.aiService = nil
		r.generateOpeningLine(ctx, fx.courseCode, "es", scenario, sess.ID)
	})

	t.Run("split path success", func(t *testing.T) {
		router.aiService = newCovSplitMockAI(t, "Welcome!")
		router.generateOpeningLine(ctx, fx.courseCode, "es", scenario, sess.ID)
		msgs, _ := router.conversationRepo.ListMessages(ctx, sess.ID)
		if len(msgs) == 0 {
			t.Fatal("expected opening message")
		}
	})

	t.Run("legacy path", func(t *testing.T) {
		router.aiService = newCovLegacyMockAI(t, "Hello there", true)
		sess2, _, _ := router.conversationRepo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
		router.generateOpeningLine(ctx, "en_ru", "en", scenario, sess2.ID)
	})

	t.Run("ai error", func(t *testing.T) {
		router.aiService = newCovFailingMockAI(t)
		router.generateOpeningLine(ctx, fx.courseCode, "es", scenario, sess.ID)
	})

	t.Run("empty visible", func(t *testing.T) {
		router.aiService = newCovEmptyMockAI(t)
		router.generateOpeningLine(ctx, fx.courseCode, "es", scenario, sess.ID)
	})

	t.Run("append error", func(t *testing.T) {
		router.aiService = newCovSplitMockAI(t, "Oops store")
		router.generateOpeningLine(ctx, fx.courseCode, "es", scenario, 999999999)
	})
}

func TestLinglowConversationCoverage_LegacyMessagePath(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovLegacyTelegram, config.LearningConfig{NativeLang: "ru", TargetLang: "en", GrammarBundleID: "en"})
	userID := linglowCovUser(t, conn, linglowCovLegacyTelegram)
	fx := insertLinglowCovScenario(t, conn, userID, "en_ru", "a0_spark_gate", "cov_legacy_msg", func(o *scenarioInsertOpts) {
		o.title = "Legacy"
		o.maxTurns = 1
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}}
	})
	router.aiService = newCovLegacyMockAI(t, "Hi!", true)

	body, _ := json.Marshal(map[string]string{"course_code": fx.courseCode, "scenario_code": fx.scenarioCode})
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start: %d %s", w.Code, w.Body.String())
	}
	var startResp struct {
		SessionID int64 `json:"session_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &startResp)

	msgBody, _ := json.Marshal(map[string]string{"text": "hello"})
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("message: %d %s", w.Code, w.Body.String())
	}

	// Budget exhausted on next turn (max_turns=2, turn_count already 1).
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("budget turn: %d %s", w.Code, w.Body.String())
	}
	var budgetResp struct {
		BudgetExhausted bool `json:"budget_exhausted"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &budgetResp)
	if !budgetResp.BudgetExhausted {
		t.Fatal("expected budget exhausted")
	}
}

func TestLinglowConversationCoverage_FreeChatAndHelpers(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovTelegramID+50, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovTelegramID+50)
	fx := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", "cov_free_chat", func(o *scenarioInsertOpts) {
		o.title = "Free"
		o.isQuest = false
		o.maxTurns = 20
	})
	router.aiService = newCovSplitMockAI(t, "Chat")

	body, _ := json.Marshal(map[string]string{"course_code": fx.courseCode, "scenario_code": fx.scenarioCode})
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start: %d", w.Code)
	}
	var startResp struct {
		SessionID int64 `json:"session_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &startResp)

	msgBody, _ := json.Marshal(map[string]string{"text": "hola"})
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("message: %d %s", w.Code, w.Body.String())
	}

	// Helper coverage
	if allTasksDone(nil, nil) {
		t.Fatal("empty tasks not done")
	}
	if allRequiredDone([]repository.ConversationTask{{IsRequired: false}}, nil) {
		t.Fatal("no required tasks")
	}
	if out := filterCorrectionsToMessage(nil, "x"); out != nil {
		t.Fatal("nil corrections")
	}
	if out := filterCorrectionsToMessage([]ai.ChatCorrection{{Corrected: "x"}}, "y"); len(out) != 1 {
		t.Fatalf("keep empty original: %+v", out)
	}

	ctx := context.Background()
	if router.courseCodeByID(ctx, 999999) != "" {
		t.Fatal("bad course id")
	}
	if router.courseTargetLang(ctx, 999999) != "" {
		t.Fatal("bad target lang")
	}
	if router.aiSvc() == nil {
		t.Fatal("expected ai svc")
	}
	router.aiService = "not-a-service"
	if router.aiSvc() != nil {
		t.Fatal("wrong ai type")
	}

	scenario := &repository.ConversationScenario{
		NPCName: "N", NPCPersona: "p", SceneSetup: "s", IsQuest: true, CEFRLevel: "A0",
	}
	tasks := []repository.ConversationTask{
		{ID: 1, Code: "greet", CompletionCriteria: "say hi", IsRequired: true},
		{ID: 2, Code: "bye", CompletionCriteria: "bye", IsRequired: false},
	}
	if p := buildConversationSystemPrompt("", "en", scenario, tasks, map[int64]bool{1: true}, true); !strings.Contains(p, "stall") {
		t.Error("nudge missing")
	}
	if p := buildConversationSystemPrompt("", "en", &repository.ConversationScenario{NPCName: "N", NPCPersona: "p", IsQuest: false}, nil, nil, false); !strings.Contains(p, "free-chat") {
		t.Error("free chat prompt")
	}
	if p := buildOpeningSystemPrompt("base", "en", scenario); !strings.Contains(p, "base") {
		t.Error("opening system")
	}
	if p := buildQuestEvalPrompt("", "en", scenario, tasks, map[int64]bool{2: true}); !strings.Contains(p, "ALREADY DONE") {
		t.Error("quest eval optional done")
	}
	if p := buildCorrectionPrompt("", "es", "B1"); !strings.Contains(p, "Detect mistakes") {
		t.Error("correction fallback")
	}
	if p := buildNPCReplyPrompt("", "en", scenario, true); !strings.Contains(p, "stall") {
		t.Error("npc nudge")
	}

	code := router.conversationCourseCode(ctx, userID, "es_ru")
	if code != "es_ru" {
		t.Fatalf("explicit course: %q", code)
	}
	code = router.conversationCourseCode(ctx, userID, "")
	if code != "es_ru" {
		t.Fatalf("default course: %q", code)
	}
}

func TestLinglowConversationCoverage_SessionByIDExtras(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovTelegramID+60, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovTelegramID+60)
	fx := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", "cov_session_extras", func(o *scenarioInsertOpts) {
		o.title = "Extras"
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}}
	})
	router.aiService = newCovSplitMockAI(t, "Hi")

	body, _ := json.Marshal(map[string]string{"course_code": fx.courseCode, "scenario_code": fx.scenarioCode})
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	var startResp struct {
		SessionID int64 `json:"session_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &startResp)

	t.Run("end abandoned default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/end", startResp.SessionID), bytes.NewReader([]byte(`{}`)))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("end: %d", w.Code)
		}
		var resp struct {
			Status string `json:"status"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Status != "abandoned" {
			t.Fatalf("status=%q", resp.Status)
		}
	})

	t.Run("nil conversation repo", func(t *testing.T) {
		r := *router
		r.conversationRepo = nil
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/sessions/%d", startResp.SessionID), nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("write session with corrections", func(t *testing.T) {
		ctx := context.Background()
		sess, _, _ := router.conversationRepo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
		_ = router.conversationRepo.AppendMessageWithCorrections(ctx, sess.ID, 1, "assistant", "hi", 1, 1, `[{"original":"x","corrected":"y","explanation":"z"}]`)
		scenario, _ := router.conversationRepo.GetScenarioByID(ctx, fx.scenarioID)
		tasks, _ := router.conversationRepo.ListTasks(ctx, fx.scenarioID)
		w := httptest.NewRecorder()
		router.writeSessionState(w, ctx, userID, sess.ID, fx.userCourseID, scenario, tasks)
		if w.Code != http.StatusOK {
			t.Fatalf("write: %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("conversation history trim", func(t *testing.T) {
		ctx := context.Background()
		sess, _, _ := router.conversationRepo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
		_ = router.conversationRepo.CloseSession(ctx, sess.ID, fx.userCourseID, "abandoned")
		sess, _, _ = router.conversationRepo.StartSession(ctx, fx.userCourseID, fx.scenarioID)
		if err := router.conversationRepo.AppendMessage(ctx, sess.ID, 1, "system", "ignored", 0, 0); err != nil {
			t.Fatalf("append system: %v", err)
		}
		for i := 2; i <= 25; i++ {
			role := "user"
			if i%2 == 0 {
				role = "assistant"
			}
			if err := router.conversationRepo.AppendMessage(ctx, sess.ID, i, role, fmt.Sprintf("m%d", i), 0, 0); err != nil {
				t.Fatalf("append %d: %v", i, err)
			}
		}
		hist, err := router.conversationHistory(ctx, sess.ID)
		if err != nil {
			t.Fatalf("history: %v", err)
		}
		if len(hist) != conversationHistoryLimit {
			t.Fatalf("history len=%d want %d", len(hist), conversationHistoryLimit)
		}
	})

	t.Run("broken db get session", func(t *testing.T) {
		r, _ := setupLinglowCovRouter(t, linglowCovTelegramID+60, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		covBreakRouterDB(t, r)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/sessions/%d", startResp.SessionID), nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", w.Code)
		}
	})
}

func TestLinglowConversationCoverage_AllTasksComplete(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovTelegramID+70, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovTelegramID+70)
	fx := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", "cov_all_tasks", func(o *scenarioInsertOpts) {
		o.title = "All Tasks"
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}, {"bye", 1, false}}
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		content := "¡Hola!"
		if strings.Contains(string(body), "quest evaluator") {
			content = `{"completed_task_codes":["greet","bye"],"all_done":true}`
		} else if strings.Contains(string(body), "correction evaluator") {
			content = `{"corrections":[]}`
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]interface{}{"role": "assistant", "content": content}},
			},
		})
	}))
	t.Cleanup(srv.Close)
	svc := ai.NewService(srv.URL, "m", "k", "p", zap.NewNop())
	svc.SetConversationQuestPromptForCourse("es_ru", "quest evaluator")
	svc.SetConversationCorrectionPromptForCourse("es_ru", "correction evaluator")
	svc.SetConversationNPCPromptForCourse("es_ru", "npc")
	router.aiService = svc

	body, _ := json.Marshal(map[string]string{"course_code": fx.courseCode, "scenario_code": fx.scenarioCode})
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowConversationSessions(w, req)
	var startResp struct {
		SessionID int64 `json:"session_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &startResp)

	msgBody, _ := json.Marshal(map[string]string{"text": "hola y adios"})
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowConversationSessionByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("message: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Status != "completed" {
		t.Fatalf("status=%q want completed", resp.Status)
	}
}

func TestLinglowConversationCoverage_RemainingBranches(t *testing.T) {
	router, conn := setupLinglowCovRouter(t, linglowCovTelegramID+80, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
	userID := linglowCovUser(t, conn, linglowCovTelegramID+80)
	fx := insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", "cov_remaining", func(o *scenarioInsertOpts) {
		o.title = "Remaining"
		o.maxTurns = 20
		o.tasks = []struct {
			code     string
			order    int
			required bool
		}{{"greet", 0, true}}
	})
	ctx := context.Background()

	t.Run("conversationCourseCode from user setting", func(t *testing.T) {
		courseRepo := repository.NewCourseRepository(conn, zap.NewNop())
		for _, pair := range []struct{ word, course string }{
			{"covapple", "en_ru"},
			{"covmanzana", "es_ru"},
		} {
			var wcID int64
			_ = conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id`, pair.word, pair.word).Scan(&wcID)
			_, _ = conn.Exec(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, course_code) VALUES ($1,$2,0,$3,$4,$5)`,
				wcID, pair.word, pair.word+"_ru", pair.word, pair.course)
		}
		_, _ = courseRepo.SelectCurrentCourse(ctx, userID, "es_ru")
		if code := router.conversationCourseCode(ctx, userID, ""); code != "es_ru" {
			t.Fatalf("course code=%q want es_ru", code)
		}
	})

	t.Run("scenarios repo errors", func(t *testing.T) {
		t.Run("list scenarios", func(t *testing.T) {
			if _, err := conn.Exec(`ALTER TABLE conversation_scenarios RENAME TO conversation_scenarios_cov_err`); err != nil {
				t.Fatalf("rename: %v", err)
			}
			t.Cleanup(func() {
				_, _ = conn.Exec(`ALTER TABLE conversation_scenarios_cov_err RENAME TO conversation_scenarios`)
			})
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", fx.districtCode, fx.courseCode), nil)
			req = setUserIDInContext(req, userID)
			w := httptest.NewRecorder()
			router.handleLinglowConversationScenarios(w, req)
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("got %d", w.Code)
			}
		})
		t.Run("passed codes", func(t *testing.T) {
			if _, err := conn.Exec(`ALTER TABLE exercise_attempts RENAME TO exercise_attempts_cov_err`); err != nil {
				t.Fatalf("rename: %v", err)
			}
			if _, err := conn.Exec(`ALTER TABLE conversation_sessions RENAME TO conversation_sessions_cov_err`); err != nil {
				t.Fatalf("rename sessions: %v", err)
			}
			t.Cleanup(func() {
				_, _ = conn.Exec(`ALTER TABLE exercise_attempts_cov_err RENAME TO exercise_attempts`)
				_, _ = conn.Exec(`ALTER TABLE conversation_sessions_cov_err RENAME TO conversation_sessions`)
			})
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", fx.districtCode, fx.courseCode), nil)
			req = setUserIDInContext(req, userID)
			w := httptest.NewRecorder()
			router.handleLinglowConversationScenarios(w, req)
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("got %d", w.Code)
			}
		})
	})

	t.Run("sessions prerequisite and repo errors", func(t *testing.T) {
		router.aiService = newCovSplitMockAI(t, "Hi")
		prevCode := "cov_rem_prev"
		nextCode := "cov_rem_next"
		insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", prevCode, func(o *scenarioInsertOpts) {
			o.title = "Prev"
			o.tasks = []struct {
				code     string
				order    int
				required bool
			}{{"greet", 0, true}}
		})
		insertLinglowCovScenario(t, conn, userID, "es_ru", "a0_spark_gate", nextCode, func(o *scenarioInsertOpts) {
			o.title = "Next"
			o.prerequisiteCode = prevCode
		})

		if _, err := conn.Exec(`ALTER TABLE exercise_attempts RENAME TO exercise_attempts_sess_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.Exec(`ALTER TABLE exercise_attempts_sess_err RENAME TO exercise_attempts`)
		})
		body, _ := json.Marshal(map[string]string{"course_code": fx.courseCode, "scenario_code": nextCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("passed codes err: got %d", w.Code)
		}
	})

	t.Run("sessions list tasks error", func(t *testing.T) {
		router.aiService = newCovSplitMockAI(t, "Hi")
		if _, err := conn.Exec(`ALTER TABLE conversation_tasks RENAME TO conversation_tasks_sess_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.Exec(`ALTER TABLE conversation_tasks_sess_err RENAME TO conversation_tasks`)
		})
		body, _ := json.Marshal(map[string]string{"course_code": fx.courseCode, "scenario_code": fx.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("list tasks: got %d", w.Code)
		}
	})

	t.Run("message nudge and zero tokens", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			content := "reply"
			if strings.Contains(string(body), "quest evaluator") {
				content = `{"completed_task_codes":[],"all_done":false}`
			} else if strings.Contains(string(body), "correction evaluator") {
				content = `{"corrections":[]}`
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]interface{}{"role": "assistant", "content": content}},
				},
			})
		}))
		t.Cleanup(srv.Close)
		svc := ai.NewService(srv.URL, "m", "k", "p", zap.NewNop())
		svc.SetConversationQuestPromptForCourse("es_ru", "quest evaluator")
		svc.SetConversationCorrectionPromptForCourse("es_ru", "correction evaluator")
		svc.SetConversationNPCPromptForCourse("es_ru", "npc")
		router.aiService = svc

		body, _ := json.Marshal(map[string]string{"course_code": fx.courseCode, "scenario_code": fx.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)
		_, _ = conn.Exec(`UPDATE conversation_sessions SET turn_count = 17 WHERE id = ?`, startResp.SessionID)

		msgBody, _ := json.Marshal(map[string]string{"text": "hola " + ai.ControlSentinel})
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
		req = setUserIDInContext(req, userID)
		w = httptest.NewRecorder()
		router.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("nudge message: %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("reset list tasks error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+85, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+85)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_reset_tasks", func(o *scenarioInsertOpts) { o.title = "Reset" })
		r.aiService = newCovSplitMockAI(t, "Hi")
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)
		if _, err := c.Exec(`ALTER TABLE conversation_tasks RENAME TO conversation_tasks_reset_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_tasks_reset_err RENAME TO conversation_tasks`)
		})
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/reset", startResp.SessionID), nil)
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("reset list tasks: %d", w.Code)
		}
	})

	t.Run("get session internal error", func(t *testing.T) {
		r, _ := setupLinglowCovRouter(t, linglowCovTelegramID+86, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		covBreakRouterDB(t, r)
		req := httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/sessions/1", nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("get err: %d", w.Code)
		}
	})

	t.Run("writeSessionState list messages error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+82, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+82)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_write_err", func(o *scenarioInsertOpts) { o.title = "Write" })
		sess, _, _ := r.conversationRepo.StartSession(ctx, f.userCourseID, f.scenarioID)
		scenario, _ := r.conversationRepo.GetScenarioByID(ctx, f.scenarioID)
		tasks, _ := r.conversationRepo.ListTasks(ctx, f.scenarioID)
		if _, err := c.Exec(`ALTER TABLE conversation_messages RENAME TO conversation_messages_cov_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_messages_cov_err RENAME TO conversation_messages`)
		})
		w := httptest.NewRecorder()
		r.writeSessionState(w, ctx, uid, sess.ID, f.userCourseID, scenario, tasks)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("write err: %d", w.Code)
		}
	})

	t.Run("conversationHistory list error", func(t *testing.T) {
		if _, err := conn.Exec(`ALTER TABLE conversation_messages RENAME TO conversation_messages_hist_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.Exec(`ALTER TABLE conversation_messages_hist_err RENAME TO conversation_messages`)
		})
		if _, err := router.conversationHistory(ctx, 1); err == nil {
			t.Fatal("expected history error")
		}
	})

	t.Run("scenario progress completed without all tasks", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+83, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+83)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_completed_only", func(o *scenarioInsertOpts) {
			o.title = "Completed only"
			o.tasks = []struct {
				code     string
				order    int
				required bool
			}{{"greet", 0, true}, {"bye", 1, false}}
		})
		tasks, _ := r.conversationRepo.ListTasks(ctx, f.scenarioID)
		sess, _, _ := r.conversationRepo.StartSession(ctx, f.userCourseID, f.scenarioID)
		_ = r.conversationRepo.CloseSession(ctx, sess.ID, f.userCourseID, "completed")
		_, fullyDone := r.scenarioProgressFlags(ctx, f.userCourseID, f.scenarioID, tasks)
		if !fullyDone {
			t.Fatal("expected fullyDone from completed status")
		}
	})

	t.Run("message ai error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+84, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+84)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_ai_err", func(o *scenarioInsertOpts) { o.title = "AI err" })
		r.aiService = newCovFailingMockAI(t)
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)
		msgBody, _ := json.Marshal(map[string]string{"text": "hi"})
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusBadGateway {
			t.Fatalf("ai error: %d", w.Code)
		}
	})

	t.Run("district lookup db error", func(t *testing.T) {
		if _, err := conn.Exec(`ALTER TABLE districts RENAME TO districts_cov_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.Exec(`ALTER TABLE districts_cov_err RENAME TO districts`)
		})
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", fx.districtCode, fx.courseCode), nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("passedAt error", func(t *testing.T) {
		if _, err := conn.Exec(`ALTER TABLE conversation_sessions RENAME TO conversation_sessions_passed_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = conn.Exec(`ALTER TABLE conversation_sessions_passed_err RENAME TO conversation_sessions`)
		})
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", fx.districtCode, fx.courseCode), nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("get scenario db error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+87, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+87)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_get_scn_err", func(o *scenarioInsertOpts) { o.title = "Scn err" })
		r.aiService = newCovSplitMockAI(t, "Hi")
		if _, err := c.Exec(`ALTER TABLE conversation_scenarios RENAME TO conversation_scenarios_get_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_scenarios_get_err RENAME TO conversation_scenarios`)
		})
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("start session error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+88, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+88)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_start_err", func(o *scenarioInsertOpts) { o.title = "Start err" })
		r.aiService = newCovSplitMockAI(t, "Hi")
		if _, err := c.Exec(`ALTER TABLE conversation_sessions RENAME TO conversation_sessions_start_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_sessions_start_err RENAME TO conversation_sessions`)
		})
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("sessions passedAt prerequisite error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+89, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+89)
		prev := "cov_pat_prev"
		next := "cov_pat_next"
		insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", prev, func(o *scenarioInsertOpts) {
			o.title = "Prev"
			o.tasks = []struct {
				code     string
				order    int
				required bool
			}{{"greet", 0, true}}
		})
		fNext := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", next, func(o *scenarioInsertOpts) {
			o.title = "Next"
			o.prerequisiteCode = prev
		})
		r.aiService = newCovSplitMockAI(t, "Hi")
		ctx := context.Background()
		prevSc, _ := r.conversationRepo.GetScenarioByCode(ctx, fNext.courseID, prev)
		sess, _, _ := r.conversationRepo.StartSession(ctx, fNext.userCourseID, prevSc.ID)
		_ = r.conversationRepo.RecordQuestCompletion(ctx, fNext.userCourseID, sql.NullInt64{}, prev, sess.ID)
		if _, err := c.Exec(`ALTER TABLE conversation_sessions RENAME TO conversation_sessions_pat_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_sessions_pat_err RENAME TO conversation_sessions`)
		})
		body, _ := json.Marshal(map[string]string{"course_code": fNext.courseCode, "scenario_code": next})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("passedAt err: %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("conversation get paths", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+90, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+90)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_get_paths", func(o *scenarioInsertOpts) { o.title = "Get" })
		r.aiService = newCovSplitMockAI(t, "Hi")
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/sessions/%d", startResp.SessionID), nil)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("unauth get: %d", w.Code)
		}

		if _, err := c.Exec(`ALTER TABLE conversation_tasks RENAME TO conversation_tasks_get_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_tasks_get_err RENAME TO conversation_tasks`)
		})
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/sessions/%d", startResp.SessionID), nil)
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("list tasks get: %d", w.Code)
		}
	})

	t.Run("message store and validation errors", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+91, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+91)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_msg_err", func(o *scenarioInsertOpts) { o.title = "Msg err" })
		r.aiService = newCovSplitMockAI(t, "Hi")
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)

		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader([]byte(`{`)))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("bad json: %d", w.Code)
		}

		if _, err := c.Exec(`ALTER TABLE conversation_messages RENAME TO conversation_messages_msg_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_messages_msg_err RENAME TO conversation_messages`)
		})
		msgBody, _ := json.Marshal(map[string]string{"text": "hello"})
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("store user msg: %d", w.Code)
		}
	})
}

func TestLinglowConversationCoverage_HandlerEdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("get session not found and scenario missing", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+100, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+100)
		req := httptest.NewRequest(http.MethodGet, "/api/linglow/conversation/sessions/999999", nil)
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("missing session: %d", w.Code)
		}

		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_orphan_sess", func(o *scenarioInsertOpts) { o.title = "Orphan" })
		sess, _, _ := r.conversationRepo.StartSession(ctx, f.userCourseID, f.scenarioID)
		_, _ = c.Exec(`DELETE FROM conversation_scenarios WHERE id = ?`, f.scenarioID)
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/sessions/%d", sess.ID), nil)
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("missing scenario: %d", w.Code)
		}
	})

	t.Run("reset not found paths", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+101, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+101)
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions/999999/reset", nil)
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("reset missing: %d", w.Code)
		}

		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_reset_orphan", func(o *scenarioInsertOpts) { o.title = "Reset orphan" })
		sess, _, _ := r.conversationRepo.StartSession(ctx, f.userCourseID, f.scenarioID)
		_, _ = c.Exec(`DELETE FROM conversation_scenarios WHERE id = ?`, f.scenarioID)
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/reset", sess.ID), nil)
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("reset scenario missing: %d", w.Code)
		}
	})

	t.Run("message reply store error and invalid quest codes", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+102, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+102)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_reply_err", func(o *scenarioInsertOpts) {
			o.title = "Reply err"
			o.tasks = []struct {
				code     string
				order    int
				required bool
			}{{"greet", 0, true}}
		})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			body, _ := io.ReadAll(req.Body)
			content := "reply"
			if strings.Contains(string(body), "quest evaluator") {
				content = `{"completed_task_codes":["bogus"],"all_done":false}`
			} else if strings.Contains(string(body), "correction evaluator") {
				content = `{"corrections":[]}`
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]interface{}{"role": "assistant", "content": content}},
				},
			})
		}))
		t.Cleanup(srv.Close)
		svc := ai.NewService(srv.URL, "m", "k", "p", zap.NewNop())
		svc.SetConversationQuestPromptForCourse("es_ru", "quest evaluator")
		svc.SetConversationCorrectionPromptForCourse("es_ru", "correction evaluator")
		svc.SetConversationNPCPromptForCourse("es_ru", "npc")
		r.aiService = svc

		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)

		if _, err := c.Exec(`ALTER TABLE conversation_messages RENAME TO conversation_messages_reply_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_messages_reply_err RENAME TO conversation_messages`)
		})
		msgBody, _ := json.Marshal(map[string]string{"text": "hello"})
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("store reply: %d", w.Code)
		}
	})

	t.Run("conversation history skips non-dialog roles in window", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+103, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+103)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_hist_roles", func(o *scenarioInsertOpts) { o.title = "Hist" })
		sess, _, _ := r.conversationRepo.StartSession(ctx, f.userCourseID, f.scenarioID)
		_ = r.conversationRepo.AppendMessage(ctx, sess.ID, 1, "system", "skip-me", 0, 0)
		for i := 2; i <= 26; i++ {
			role := "user"
			if i%2 == 0 {
				role = "assistant"
			}
			_ = r.conversationRepo.AppendMessage(ctx, sess.ID, i, role, fmt.Sprintf("m%d", i), 0, 0)
		}
		hist, err := r.conversationHistory(ctx, sess.ID)
		if err != nil {
			t.Fatalf("history: %v", err)
		}
		if len(hist) != conversationHistoryLimit {
			t.Fatalf("len=%d want %d", len(hist), conversationHistoryLimit)
		}
		for _, m := range hist {
			if m.Role != "user" && m.Role != "assistant" {
				t.Fatalf("unexpected role %q", m.Role)
			}
		}
	})

	t.Run("message get session internal error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+104, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+104)
		covBreakRouterDB(t, r)
		r.aiService = newCovSplitMockAI(t, "Hi")
		msgBody, _ := json.Marshal(map[string]string{"text": "hello"})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions/1/message", bytes.NewReader(msgBody))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", w.Code)
		}
	})
}

func TestLinglowConversationCoverage_FinalBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("scenarios passedAt query error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+110, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+110)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_passed_at", func(o *scenarioInsertOpts) { o.title = "PassedAt" })
		if _, err := c.Exec(`ALTER TABLE exercise_attempts RENAME TO exercise_attempts_pat2_err`); err != nil {
			t.Fatalf("rename ea: %v", err)
		}
		if _, err := c.Exec(`ALTER TABLE conversation_sessions RENAME TO conversation_sessions_pat2_err`); err != nil {
			t.Fatalf("rename cs: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE exercise_attempts_pat2_err RENAME TO exercise_attempts`)
			_, _ = c.Exec(`ALTER TABLE conversation_sessions_pat2_err RENAME TO conversation_sessions`)
		})
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/scenarios?district_code=%s&course_code=%s", f.districtCode, f.courseCode), nil)
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationScenarios(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("get scenario by id error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+111, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+111)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_get_scn", func(o *scenarioInsertOpts) { o.title = "GetScn" })
		r.aiService = newCovSplitMockAI(t, "Hi")
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)
		if _, err := c.Exec(`ALTER TABLE conversation_scenarios RENAME TO conversation_scenarios_get2_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_scenarios_get2_err RENAME TO conversation_scenarios`)
		})
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/linglow/conversation/sessions/%d", startResp.SessionID), nil)
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("got %d", w.Code)
		}
	})

	t.Run("message list tasks error", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+116, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+116)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_list_tasks_msg", func(o *scenarioInsertOpts) { o.title = "ListTasks" })
		r.aiService = newCovSplitMockAI(t, "Hi")
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)
		if _, err := c.Exec(`ALTER TABLE conversation_tasks RENAME TO conversation_tasks_msg2_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_tasks_msg2_err RENAME TO conversation_tasks`)
		})
		msgBody, _ := json.Marshal(map[string]string{"text": "hello"})
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("list tasks msg: %d", w.Code)
		}
	})

	t.Run("quest record and mark task warnings", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+114, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+114)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_quest_warn", func(o *scenarioInsertOpts) {
			o.title = "QuestWarn"
			o.tasks = []struct {
				code     string
				order    int
				required bool
			}{{"greet", 0, true}}
		})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			body, _ := io.ReadAll(req.Body)
			content := "ok"
			if strings.Contains(string(body), "quest evaluator") {
				content = `{"completed_task_codes":["greet"],"all_done":true}`
			} else if strings.Contains(string(body), "correction evaluator") {
				content = `{"corrections":[]}`
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]interface{}{"role": "assistant", "content": content}},
				},
			})
		}))
		t.Cleanup(srv.Close)
		svc := ai.NewService(srv.URL, "m", "k", "p", zap.NewNop())
		svc.SetConversationQuestPromptForCourse("es_ru", "quest evaluator")
		svc.SetConversationCorrectionPromptForCourse("es_ru", "correction evaluator")
		svc.SetConversationNPCPromptForCourse("es_ru", "npc")
		r.aiService = svc

		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)

		if _, err := c.Exec(`ALTER TABLE exercise_attempts RENAME TO exercise_attempts_quest_err`); err != nil {
			t.Fatalf("rename: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE exercise_attempts_quest_err RENAME TO exercise_attempts`)
		})
		msgBody, _ := json.Marshal(map[string]string{"text": "hola"})
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("quest warn message: %d %s", w.Code, w.Body.String())
		}

		if _, err := c.Exec(`ALTER TABLE conversation_task_progress RENAME TO conversation_task_progress_err`); err != nil {
			t.Fatalf("rename tp: %v", err)
		}
		t.Cleanup(func() {
			_, _ = c.Exec(`ALTER TABLE conversation_task_progress_err RENAME TO conversation_task_progress`)
		})
		req = httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("mark warn message: %d %s", w.Code, w.Body.String())
		}
	})

	t.Run("message without ai on open session", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+117, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+117)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_msg_no_ai", func(o *scenarioInsertOpts) { o.title = "NoAI msg" })
		r.aiService = newCovSplitMockAI(t, "Hi")
		body, _ := json.Marshal(map[string]string{"course_code": f.courseCode, "scenario_code": f.scenarioCode})
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/conversation/sessions", bytes.NewReader(body))
		req = setUserIDInContext(req, uid)
		w := httptest.NewRecorder()
		r.handleLinglowConversationSessions(w, req)
		var startResp struct {
			SessionID int64 `json:"session_id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &startResp)
		r.aiService = nil
		msgBody, _ := json.Marshal(map[string]string{"text": "hello"})
		req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/linglow/conversation/sessions/%d/message", startResp.SessionID), bytes.NewReader(msgBody))
		req = setUserIDInContext(req, uid)
		w = httptest.NewRecorder()
		r.handleLinglowConversationSessionByID(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("no ai message: %d", w.Code)
		}
	})

	t.Run("history filters system in tail window", func(t *testing.T) {
		r, c := setupLinglowCovRouter(t, linglowCovTelegramID+115, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"})
		uid := linglowCovUser(t, c, linglowCovTelegramID+115)
		f := insertLinglowCovScenario(t, c, uid, "es_ru", "a0_spark_gate", "cov_hist_tail", func(o *scenarioInsertOpts) { o.title = "Tail" })
		sess, _, _ := r.conversationRepo.StartSession(ctx, f.userCourseID, f.scenarioID)
		for i := 1; i <= 23; i++ {
			role := "user"
			if i%2 == 0 {
				role = "assistant"
			}
			_ = r.conversationRepo.AppendMessage(ctx, sess.ID, i, role, "x", 0, 0)
		}
		_ = r.conversationRepo.AppendMessage(ctx, sess.ID, 24, "system", "in-window", 0, 0)
		_ = r.conversationRepo.AppendMessage(ctx, sess.ID, 25, "user", "last", 0, 0)
		hist, err := r.conversationHistory(ctx, sess.ID)
		if err != nil {
			t.Fatalf("history: %v", err)
		}
		if len(hist) != 11 {
			t.Fatalf("len=%d want 11 after filtering system", len(hist))
		}
	})
}