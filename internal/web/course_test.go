package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestHandleLearningCourse_OK(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)

	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(100500)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := repository.NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(t.Context(), cfg.Learning); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
		SELECT c.id, d.id, l.id, 'reading_category:test', 'reading', 'Reading Test', 'reading_category', 'read.test', 1, 'published'
		FROM courses c
		JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
		JOIN locations l ON l.district_id = d.id AND l.location_type = 'reading'
		WHERE c.code = 'es_ru'
	`); err != nil {
		t.Fatalf("insert module: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'reading_text', 'reading_text', 'read.text.test', 'Text Test', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'reading_category:test'
	`); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/learning/course", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningCourse(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Course struct {
			Code string `json:"code"`
		} `json:"course"`
		UserCourse *struct {
			ID int64 `json:"id"`
		} `json:"user_course"`
		Totals struct {
			Districts int            `json:"districts"`
			Locations int            `json:"locations"`
			Modules   int            `json:"modules"`
			Items     int            `json:"items"`
			ByType    map[string]int `json:"by_type"`
		} `json:"totals"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Course.Code != "es_ru" || body.UserCourse == nil {
		t.Fatalf("unexpected course payload: %+v", body)
	}
	if body.Totals.Districts != 6 || body.Totals.Locations != 36 || body.Totals.Modules != 1 || body.Totals.Items != 1 {
		t.Fatalf("unexpected totals: %+v", body.Totals)
	}
	if body.Totals.ByType["reading_text"] != 1 {
		t.Fatalf("reading_text total = %d, want 1", body.Totals.ByType["reading_text"])
	}
}

func TestHandleLearningCourse_UsesCurrentCourseAndExplicitOverride(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)

	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(100501)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := repository.NewCourseRepository(conn, logger)
	if _, err := courseRepo.SelectCurrentCourse(t.Context(), user.ID, "en_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/learning/course", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningCourse(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("current status=%d body=%s", w.Code, w.Body.String())
	}
	var currentBody struct {
		Course struct {
			Code string `json:"code"`
		} `json:"course"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &currentBody); err != nil {
		t.Fatalf("decode current: %v", err)
	}
	if currentBody.Course.Code != "en_ru" {
		t.Fatalf("current course = %q, want en_ru", currentBody.Course.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/city?course_code=es_ru", nil)
	req = setUserIDInContext(req, user.ID)
	w = httptest.NewRecorder()
	router.handleLinglowCity(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("explicit status=%d body=%s", w.Code, w.Body.String())
	}
	var explicitBody struct {
		Course struct {
			Code string `json:"code"`
		} `json:"course"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &explicitBody); err != nil {
		t.Fatalf("decode explicit: %v", err)
	}
	if explicitBody.Course.Code != "es_ru" {
		t.Fatalf("explicit course = %q, want es_ru", explicitBody.Course.Code)
	}

	resolved, err := courseRepo.ResolveCurrentCourseCode(t.Context(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("ResolveCurrentCourseCode: %v", err)
	}
	if resolved != "en_ru" {
		t.Fatalf("explicit read changed current course to %q", resolved)
	}
}

func TestHandleLearningCourse_ExplicitUnknownCourse(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	router := NewRouter(zap.NewNop(), &config.Config{Learning: config.DefaultLearningConfig()}, conn, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(100502)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/city?course_code=xx_ru", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLinglowCity(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleLearningCourse_Unauthorized(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	router := NewRouter(zap.NewNop(), &config.Config{Learning: config.DefaultLearningConfig()}, conn, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/course", nil)
	w := httptest.NewRecorder()
	router.handleLearningCourse(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
