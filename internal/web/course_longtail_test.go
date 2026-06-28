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
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestHandleLinglowWordsAndHistory(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(998801)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := repository.NewCourseRepository(conn, logger)
	if _, err := courseRepo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var wordCardID int64
	if err := conn.QueryRow(`
		INSERT INTO word_cards (word, definition, display_en, definition_ru)
		VALUES ('gato', 'cat', 'cat', 'кот')
		RETURNING id
	`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	var itemID int64
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'word_set:lt-words', 'word_set', 'LT Words', 'word_set', 'lt-words', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', $1, 'gato', 'A0', 'published'
		FROM module
		RETURNING id
	`, strconv.FormatInt(wordCardID, 10)).Scan(&itemID); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}
	correct := true
	quality := 4
	if _, err := courseRepo.RecordExerciseAttempt(context.Background(), repository.ExerciseAttemptInput{
		UserID: user.ID, DefaultCourse: "es_ru", LearningItemID: itemID,
		Mode: "card", ClientAttemptID: "lt-word-1", IsCorrect: &correct, Quality: &quality,
	}); err != nil {
		t.Fatalf("RecordExerciseAttempt: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/words?limit=10&q=кот", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLinglowWords(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("words status=%d body=%s", w.Code, w.Body.String())
	}
	var wordsResp struct {
		Total int `json:"total"`
		Words []struct {
			Lemma string `json:"lemma"`
		} `json:"words"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &wordsResp); err != nil {
		t.Fatalf("decode words: %v", err)
	}
	if wordsResp.Total != 1 || len(wordsResp.Words) != 1 || wordsResp.Words[0].Lemma != "gato" {
		t.Fatalf("words response = %+v", wordsResp)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/words?limit=bad", nil)
	req = setUserIDInContext(req, user.ID)
	w = httptest.NewRecorder()
	router.handleLinglowWords(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad limit expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/history?days=7", nil)
	req = setUserIDInContext(req, user.ID)
	w = httptest.NewRecorder()
	router.handleLinglowHistory(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", w.Code, w.Body.String())
	}
	var histResp struct {
		TotalAttempts   int `json:"total_attempts"`
		CorrectAttempts int `json:"correct_attempts"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &histResp); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	if histResp.TotalAttempts != 1 || histResp.CorrectAttempts != 1 {
		t.Fatalf("history = %+v", histResp)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/history?course_code=xx_ru", nil)
	req = setUserIDInContext(req, user.ID)
	w = httptest.NewRecorder()
	router.handleLinglowHistory(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown course expected 404, got %d", w.Code)
	}
}

func TestCourseHelpers_MultiCourseScope(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "en", GrammarBundleID: "en"}}
	router := NewRouter(logger, cfg, conn, nil, nil, service.NewOptionsService(repository.NewTrainingCardRepository(conn, logger), logger, "en"), nil)
	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(998802)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := repository.NewCourseRepository(conn, logger)
	if _, err := courseRepo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	// Tag training cards with two course codes so HasMultipleWordCourses is true.
	for _, pair := range []struct{ word, course string }{
		{"apple", "en_ru"},
		{"manzana", "es_ru"},
	} {
		var wcID int64
		if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id`, pair.word, pair.word).Scan(&wcID); err != nil {
			t.Fatalf("insert word_cards: %v", err)
		}
		if _, err := conn.Exec(`
			INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, course_code)
			VALUES ($1, $2, 0, $3, $4, $5)
		`, wcID, pair.word, pair.word+"_ru", pair.word, pair.course); err != nil {
			t.Fatalf("insert training_cards: %v", err)
		}
	}

	ctx := context.Background()
	if code := router.currentCourseCodeForUser(ctx, user.ID); code != "es_ru" {
		t.Fatalf("currentCourseCodeForUser = %q, want es_ru", code)
	}
	lc := router.learningConfigForUser(ctx, user.ID)
	if lc.TargetLang != "es" {
		t.Fatalf("learningConfigForUser target = %q, want es", lc.TargetLang)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/words?course_code=en_ru", nil)
	if code := router.requestedCourseCodeForUser(req, user.ID); code != "en_ru" {
		t.Fatalf("requestedCourseCodeForUser explicit = %q", code)
	}
}

func TestCourseAPI_GuardsAndValidation(t *testing.T) {
	router, userID := setupCourseAPITest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/courses", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleCourses(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("handleCourses POST expected 405, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/user/courses/current", nil)
	w = httptest.NewRecorder()
	router.handleCurrentCourse(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("handleCurrentCourse no auth expected 401, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/user/courses/select", bytes.NewReader([]byte("{")))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleSelectCourse(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/user/courses/select", bytes.NewReader([]byte(`{"course_code":""}`)))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleSelectCourse(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty course_code expected 400, got %d", w.Code)
	}
}

func TestEnrichTheoryBlockTitles(t *testing.T) {
	router, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()

	items := []repository.DailyRouteItem{
		{SourceKind: "reading_text", SourceID: "x:y", Title: "keep"},
		{SourceKind: "grammar_theory_block", SourceID: "ch1:block-a", Title: "block-a"},
	}
	router.enrichTheoryBlockTitles(httptest.NewRequest(http.MethodGet, "/", nil), userID, items)
	if items[0].Title != "keep" {
		t.Fatalf("non-grammar title changed: %q", items[0].Title)
	}
	// Grammar service may or may not have TheoryIndex entry; call should not panic.
	router.enrichTheoryBlockTitles(httptest.NewRequest(http.MethodGet, "/", nil), userID, nil)
}

func TestHandleLinglowExerciseAttempts_Validation(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	router := NewRouter(zap.NewNop(), &config.Config{Learning: config.DefaultLearningConfig()}, conn, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(conn, zap.NewNop())
	user, _ := userRepo.GetOrCreateUser(998803)

	req := httptest.NewRequest(http.MethodPost, "/api/linglow/exercise-attempts", bytes.NewReader([]byte(`{"prompt":{},"answer":{},"result":{}}`)))
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLinglowExerciseAttempts(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing mode expected 400, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/exercise-attempts", bytes.NewReader([]byte(`{"mode":"x","prompt":[],"answer":{},"result":{}}`)))
	req = setUserIDInContext(req, user.ID)
	w = httptest.NewRecorder()
	router.handleLinglowExerciseAttempts(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("prompt not object expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/exercise-attempts", bytes.NewReader([]byte(`{"mode":"x","prompt":{},"answer":{},"result":{},"answered_at":"not-rfc3339"}`)))
	req = setUserIDInContext(req, user.ID)
	w = httptest.NewRecorder()
	router.handleLinglowExerciseAttempts(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad answered_at expected 400, got %d", w.Code)
	}
}
