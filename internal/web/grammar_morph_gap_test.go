package web

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

const grammarMorphGapUserTelegram int64 = 900001

func setupGrammarMorphGapRouter(t *testing.T, linglow config.LinglowConfig) (*Router, int64) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	cfg := &config.Config{
		Learning: lc,
		Linglow:  linglow,
	}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(grammarMorphGapUserTelegram)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := repository.NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(context.Background(), lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	return router, user.ID
}

func seedGrammarMorphGapLearningItem(t *testing.T, conn *sql.DB) {
	t.Helper()
	if _, err := conn.Exec(`
		INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
		SELECT c.id, d.id, l.id, 'grammar_section:morph-gap', 'grammar', 'Morph Gap', 'grammar_section', 'morph.section', 1, 'published'
		FROM courses c
		JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
		JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
		WHERE c.code = 'es_ru'
		ON CONFLICT DO NOTHING
	`); err != nil {
		t.Fatalf("insert module: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'grammar_chapter', 'grammar_chapter', 'morph.chapter', 'Morph Chapter', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'grammar_section:morph-gap'
		ON CONFLICT DO NOTHING
	`); err != nil {
		t.Fatalf("insert chapter item: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'grammar_theory_block', 'grammar_theory_block', 'morph.block', 'Morph Block', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'grammar_section:morph-gap'
		ON CONFLICT DO NOTHING
	`); err != nil {
		t.Fatalf("insert block item: %v", err)
	}
}

func TestGrammarMorphGap_00_MorphologyHelpers(t *testing.T) {
	t.Run("firstNonEmpty", func(t *testing.T) {
		cases := []struct {
			name string
			in   []string
			want string
		}{
			{name: "first value", in: []string{"a", "b"}, want: "a"},
			{name: "skip blank", in: []string{"  ", "b"}, want: "b"},
			{name: "all empty", in: []string{"", "  "}, want: ""},
			{name: "none", in: nil, want: ""},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if got := firstNonEmpty(tc.in...); got != tc.want {
					t.Fatalf("firstNonEmpty() = %q, want %q", got, tc.want)
				}
			})
		}
	})

	t.Run("linglowAnsweredAt", func(t *testing.T) {
		fixed := time.Date(2024, 6, 15, 12, 30, 0, 0, time.FixedZone("MSK", 3*3600))
		got := linglowAnsweredAt(fixed)
		if !got.Equal(fixed.UTC()) {
			t.Fatalf("non-zero: got %v want %v", got, fixed.UTC())
		}
		before := time.Now().UTC()
		fresh := linglowAnsweredAt(time.Time{})
		after := time.Now().UTC()
		if fresh.Before(before) || fresh.After(after) {
			t.Fatalf("zero time should become now UTC, got %v", fresh)
		}
	})

	t.Run("normalizeNounGenderValue", func(t *testing.T) {
		cases := []struct {
			in, want string
		}{
			{"M", "m"}, {" F ", "f"}, {"mf", "mf"}, {"n", "n"},
			{"unknown", ""}, {"", ""},
		}
		for _, tc := range cases {
			if got := normalizeNounGenderValue(tc.in); got != tc.want {
				t.Fatalf("normalizeNounGenderValue(%q) = %q, want %q", tc.in, got, tc.want)
			}
		}
	})

	t.Run("nounArticleForTarget", func(t *testing.T) {
		cases := []struct {
			lang, gender, want string
		}{
			{"es", "m", "el"}, {"es", "f", "la"}, {"es", "mf", "el/la"}, {"es", "n", "lo"},
			{"es", "x", ""}, {"en", "m", ""}, {"ES", "f", "la"},
		}
		for _, tc := range cases {
			if got := nounArticleForTarget(tc.lang, tc.gender); got != tc.want {
				t.Fatalf("nounArticleForTarget(%q,%q) = %q, want %q", tc.lang, tc.gender, got, tc.want)
			}
		}
	})
}

func grammarMorphGapWriteLexicon(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "lexicon.tsv")
	content := "lemma\tgender\tarticle\topposite_gender_word\tsource\tnotes\n" +
		"perro\tm\tel\tperra\tunit\tdog\n" +
		"libro\tm\tel\tlibra\tunit\tbook\n" +
		"agua\tmf\tel/la\t\tunit\tboth\n" +
		"it\tf\tla\t\tunit\tfeminine\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write lexicon: %v", err)
	}
	t.Setenv("SPANISH_GENDER_LEXICON_PATH", path)
	return path
}

func TestGrammarMorphGap_01_SpanishGenderLexicon(t *testing.T) {
	grammarMorphGapWriteLexicon(t)

	if _, ok := lookupSpanishGenderLexiconByLemma(""); ok {
		t.Fatal("empty lemma should miss")
	}
	if _, ok := lookupSpanishGenderLexiconByLemma("  "); ok {
		t.Fatal("whitespace lemma should miss")
	}

	entry, ok := lookupSpanishGenderLexiconByLemma("Perro")
	if !ok || entry.Gender != "m" || entry.Article != "el" || entry.OppositeGenderWord != "perra" {
		t.Fatalf("perro entry: %+v ok=%v", entry, ok)
	}
	libroEntry, ok := lookupSpanishGenderLexiconByLemma("libro")
	if !ok || libroEntry.Gender != "m" {
		t.Fatalf("libro entry: %+v ok=%v", libroEntry, ok)
	}
	if _, ok := lookupSpanishGenderLexiconByLemma("missing-lemma-xyz"); ok {
		t.Fatal("unknown lemma should miss")
	}
}

func TestGrammarMorphGap_02_BuildCompactMorphFromWordCard(t *testing.T) {
	grammarMorphGapWriteLexicon(t)

	nounPOS := "noun"
	mGender := "m"
	fGender := "f"
	opp := "hermana"
	verbPOS := "verb"
	verbJSON := `{"v1":"hablar","gerund":"hablando"}`

	t.Run("nil card and pos", func(t *testing.T) {
		if got := buildCompactMorphFromWordCard("es", nil, nil); got != nil {
			t.Fatalf("expected nil, got %+v", got)
		}
	})

	t.Run("noun with card gender", func(t *testing.T) {
		card := &models.WordCard{Word: "perro", NounGender: &mGender}
		got := buildCompactMorphFromWordCard("es", card, &nounPOS)
		if got == nil || got.NounGender != "m" || got.Article != "el" {
			t.Fatalf("unexpected morph: %+v", got)
		}
	})

	t.Run("noun mf gender article", func(t *testing.T) {
		card := &models.WordCard{Word: "agua"}
		got := buildCompactMorphFromWordCard("es", card, &nounPOS)
		if got == nil || got.NounGender != "mf" || got.Article != "el/la" {
			t.Fatalf("mf gender morph: %+v", got)
		}
	})

	t.Run("noun opposite from lexicon only", func(t *testing.T) {
		card := &models.WordCard{Word: "libro"}
		got := buildCompactMorphFromWordCard("es", card, &nounPOS)
		if got == nil || got.NounGender != "m" || got.Article != "el" || got.OppositeGenderWord != "libra" {
			t.Fatalf("lexicon fallback morph: %+v", got)
		}
	})

	t.Run("noun with opposite from card", func(t *testing.T) {
		card := &models.WordCard{Word: "hermano", NounGender: &mGender, OppositeGenderWord: &opp}
		got := buildCompactMorphFromWordCard("es", card, &nounPOS)
		if got == nil || got.OppositeGenderWord != "hermana" {
			t.Fatalf("opposite gender: %+v", got)
		}
	})

	t.Run("noun non-es target skips lexicon", func(t *testing.T) {
		card := &models.WordCard{Word: "libro"}
		got := buildCompactMorphFromWordCard("en", card, &nounPOS)
		if got == nil || got.POS != "noun" || got.NounGender != "" || got.Article != "" {
			t.Fatalf("english noun without gender metadata: %+v", got)
		}
	})

	t.Run("verb forms", func(t *testing.T) {
		card := &models.WordCard{Word: "hablar", VerbFormsJSON: &verbJSON}
		got := buildCompactMorphFromWordCard("es", card, &verbPOS)
		if got == nil || got.VerbForms == nil || got.VerbForms.V1 != "hablar" {
			t.Fatalf("verb morph: %+v", got)
		}
	})

	t.Run("fallback pos from card", func(t *testing.T) {
		cardPOS := "Noun"
		card := &models.WordCard{Word: "mesa", NounGender: &fGender, POS: &cardPOS}
		got := buildCompactMorphFromWordCard("es", card, nil)
		if got == nil || got.POS != "noun" || got.Article != "la" {
			t.Fatalf("card pos fallback: %+v", got)
		}
	})
}

func TestGrammarMorphGap_03_RecordLinglowEarlyReturns(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	result := &service.TestResult{AttemptID: 1, AnsweredAt: time.Now()}
	training := &service.GrammarSrsAnswerResult{AttemptID: 1, ChapterID: "c", TheoryBlockID: "b", AnsweredAt: time.Now()}

	var nilRouter *Router
	nilRouter.recordLinglowGrammarTestAttempt(req, 1, "chapter", "id", "", nil, result)
	nilRouter.recordLinglowGrammarTrainingAttempt(req, 1, "", training)
	nilRouter.mirrorLegacyGrammarSRS(req.Context(), 1, "c", "b")

	disabled, _ := setupGrammarMorphGapRouter(t, config.LinglowConfig{EventsWriteEnabled: false})
	disabled.recordLinglowGrammarTestAttempt(req, 1, "chapter", "id", "", nil, result)
	disabled.recordLinglowGrammarTrainingAttempt(req, 1, "", training)

	enabled, _ := setupGrammarMorphGapRouter(t, config.LinglowConfig{EventsWriteEnabled: true})
	enabled.recordLinglowGrammarTestAttempt(req, 1, "chapter", "id", "", nil, nil)
	enabled.recordLinglowGrammarTestAttempt(req, 1, "chapter", "id", "", nil, &service.TestResult{})
	enabled.recordLinglowGrammarTrainingAttempt(req, 1, "", nil)
	enabled.recordLinglowGrammarTrainingAttempt(req, 1, "", &service.GrammarSrsAnswerResult{})

	enabled.linglowEventRepo = nil
	enabled.recordLinglowGrammarTestAttempt(req, 1, "chapter", "id", "", nil, result)
	enabled.recordLinglowGrammarTrainingAttempt(req, 1, "", training)

	enabled.mirrorLegacyGrammarSRS(req.Context(), 0, "", "")
	enabled.config.Linglow.SRSReadEnabled = false
	enabled.mirrorLegacyGrammarSRS(req.Context(), 1, "c", "b")
	enabled.config.Linglow.SRSReadEnabled = true
	enabled.linglowSRSMirrorRepo = nil
	enabled.mirrorLegacyGrammarSRS(req.Context(), 1, "c", "b")
}

func TestGrammarMorphGap_04_RecordLinglowTestAttempt(t *testing.T) {
	router, userID := setupGrammarMorphGapRouter(t, config.LinglowConfig{EventsWriteEnabled: true})
	seedGrammarMorphGapLearningItem(t, router.db)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	answered := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC)
	result := &service.TestResult{
		AttemptID:  88001,
		Score:      90,
		Passed:     true,
		Correct:    9,
		Total:      10,
		Results:    []interface{}{map[string]interface{}{"question_id": "q1", "correct": true}},
		AnsweredAt: answered,
	}
	answers := []service.AnswerItem{{QuestionID: "q1", Answer: "a"}}

	router.recordLinglowGrammarTestAttempt(req, userID, "chapter", "morph.chapter", "client-gap", answers, result)

	broken := newBrokenDB(t)
	router.linglowEventRepo = repository.NewLinglowEventRepository(broken)
	router.recordLinglowGrammarTestAttempt(req, userID, "chapter", "morph.chapter", "client-gap", answers, result)
}

func TestGrammarMorphGap_05_RecordLinglowTrainingAttempt(t *testing.T) {
	router, userID := setupGrammarMorphGapRouter(t, config.LinglowConfig{
		EventsWriteEnabled: true,
		SRSReadEnabled:     true,
	})
	seedGrammarMorphGapLearningItem(t, router.db)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	answered := time.Date(2025, 3, 4, 8, 0, 0, 0, time.UTC)
	result := &service.GrammarSrsAnswerResult{
		AttemptID:       88002,
		ChapterID:       "morph.chapter",
		TheoryBlockID:   "morph.block",
		ConceptID:       "morph.concept",
		QuestionID:      "morph-q1",
		Correct:         true,
		UserAnswer:      "si",
		CorrectAnswer:   "sí",
		ClientAttemptID: "from-result",
		AnsweredAt:      answered,
	}

	router.recordLinglowGrammarTrainingAttempt(req, userID, "  client-gap  ", result)

	broken := newBrokenDB(t)
	router.linglowSRSMirrorRepo = repository.NewLinglowSRSMirrorRepository(broken)
	router.recordLinglowGrammarTrainingAttempt(req, userID, "", result)

	router.linglowEventRepo = repository.NewLinglowEventRepository(broken)
	router.linglowSRSMirrorRepo = repository.NewLinglowSRSMirrorRepository(router.db)
	router.recordLinglowGrammarTrainingAttempt(req, userID, "only-client", result)
}

func TestGrammarMorphGap_06_MirrorLegacyGrammarSRS(t *testing.T) {
	router, userID := setupGrammarMorphGapRouter(t, config.LinglowConfig{SRSReadEnabled: true})
	seedGrammarMorphGapLearningItem(t, router.db)

	ctx := context.Background()
	router.mirrorLegacyGrammarSRS(ctx, userID, "morph.chapter", "morph.block")

	broken := newBrokenDB(t)
	router.linglowSRSMirrorRepo = repository.NewLinglowSRSMirrorRepository(broken)
	router.mirrorLegacyGrammarSRS(ctx, userID, "morph.chapter", "morph.block")
}

func setupGrammarMorphGapAdminRouter(t *testing.T) *Router {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	return NewRouter(logger, cfg, conn, nil, nil, nil, nil)
}

func TestGrammarMorphGap_07_AdminConversationsList(t *testing.T) {
	router := setupGrammarMorphGapAdminRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/conversations/scenarios?course_code=es_ru", nil)
	w := httptest.NewRecorder()
	router.adminConversationsList(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["course_code"] != "es_ru" {
		t.Fatalf("course_code: %+v", body["course_code"])
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/conversations/scenarios", nil)
	w = httptest.NewRecorder()
	router.adminConversationsList(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("default course list expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/conversations/scenarios?course_code=missing_course_xyz", nil)
	w = httptest.NewRecorder()
	router.adminConversationsList(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown course expected 404, got %d", w.Code)
	}
}

func TestGrammarMorphGap_08_AdminConversationHandlers(t *testing.T) {
	router := setupGrammarMorphGapAdminRouter(t)

	createBody := map[string]interface{}{
		"code": "morph_gap_cafe", "title": "Morph Gap Cafe", "cefr_level": "A0",
		"npc_name": "Luis", "scene_setup": "Morning cafe", "status": "draft",
	}
	raw, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios?course_code=es_ru", bytes.NewReader(raw))
	w := httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create scenario expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var created map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	scenarioID := int64(created["id"].(float64))

	req = httptest.NewRequest(http.MethodGet, "/api/admin/conversations/scenarios?course_code=es_ru", nil)
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarios(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list via handler expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/scenarios/not-a-number", bytes.NewReader([]byte(`{}`)))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid scenario id expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPatch, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10), nil)
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("patch scenario expected 405, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"?course_code=es_ru", bytes.NewReader([]byte(`{"code":"x","title":"","cefr_level":""}`)))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid scenario update expected 400, got %d: %s", w.Code, w.Body.String())
	}

	updateBody := map[string]interface{}{
		"code": "morph_gap_cafe", "title": "Updated Morph Cafe", "cefr_level": "A0",
		"npc_name": "Luis", "scene_setup": "Updated scene", "status": "active",
	}
	raw, _ = json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"?course_code=es_ru", bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update scenario expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"?course_code=es_ru", bytes.NewReader([]byte("{")))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json on update expected 400, got %d", w.Code)
	}

	taskBody := map[string]interface{}{
		"code": "order", "title": "Order coffee", "completion_criteria": "user orders drink",
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

	req = httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"/tasks", bytes.NewReader([]byte(`{"code":"","title":"","completion_criteria":""}`)))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid task payload expected 400, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"/tasks", bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate task expected 409, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"/tasks", bytes.NewReader([]byte("{")))
	w = httptest.NewRecorder()
	router.adminConversationTaskCreate(w, req, scenarioID)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json on task create expected 400, got %d", w.Code)
	}

	updateTaskBody := map[string]interface{}{
		"code": "order", "title": "Order politely", "completion_criteria": "polite order", "is_required": true,
	}
	raw, _ = json.Marshal(updateTaskBody)
	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/tasks/"+strconv.FormatInt(taskID, 10), bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationTaskByID(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update task expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/tasks/not-id", bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationTaskByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid task id expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/tasks/"+strconv.FormatInt(taskID, 10), bytes.NewReader([]byte("{")))
	w = httptest.NewRecorder()
	router.handleAdminConversationTaskByID(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json on task update expected 400, got %d", w.Code)
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

	router.conversationRepo = nil
	req = httptest.NewRequest(http.MethodGet, "/api/admin/conversations/scenarios/"+strconv.FormatInt(scenarioID, 10)+"/tasks", nil)
	w = httptest.NewRecorder()
	router.handleAdminConversationScenarioByID(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil repo scenario handler expected 503, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/conversations/tasks/1", bytes.NewReader(raw))
	w = httptest.NewRecorder()
	router.handleAdminConversationTaskByID(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil repo task handler expected 503, got %d", w.Code)
	}
}
