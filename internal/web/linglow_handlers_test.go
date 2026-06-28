package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupLinglowHandlersTest(t *testing.T, telegramID int64) (*Router, int64) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := repository.NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(t.Context(), cfg.Learning); err != nil {
		t.Fatalf("backfill user courses: %v", err)
	}
	return router, user.ID
}

func TestLinglowActivity(t *testing.T) {
	router, userID := setupLinglowHandlersTest(t, 900301)

	body := map[string]interface{}{
		"course_code": "es_ru",
		"client_day":  time.Now().Format("2006-01-02"),
		"seconds":     30,
		"mode":        "words",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/activity", bytes.NewReader(raw))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowActivity(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var accepted map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &accepted); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if accepted["accepted"] != true {
		t.Fatalf("expected accepted=true, got %+v", accepted)
	}

	body["seconds"] = 999
	raw, _ = json.Marshal(body)
	req = httptest.NewRequest(http.MethodPost, "/api/linglow/activity", bytes.NewReader(raw))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowActivity(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("capped ping expected 200, got %d", w.Code)
	}

	body["seconds"] = 0
	raw, _ = json.Marshal(body)
	req = httptest.NewRequest(http.MethodPost, "/api/linglow/activity", bytes.NewReader(raw))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowActivity(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("zero seconds expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/activity", bytes.NewReader([]byte("{")))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowActivity(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/activity", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowActivity(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET expected 405, got %d", w.Code)
	}
}

func TestLinglowActivity_DailyCap(t *testing.T) {
	router, userID := setupLinglowHandlersTest(t, 900302)
	statsRepo := repository.NewLinglowDailyStatsRepository(router.db)
	userCourseID, err := statsRepo.ResolveUserCourseID(t.Context(), userID, "es_ru")
	if err != nil {
		t.Fatalf("ResolveUserCourseID: %v", err)
	}
	day := time.Now().Format("2006-01-02")
	if err := statsRepo.Bump(t.Context(), repository.DailyBump{
		UserCourseID: userCourseID, Day: day, Mode: "other", ActiveSeconds: maxActivityDailySeconds,
	}); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	body := map[string]interface{}{"course_code": "es_ru", "client_day": day, "seconds": 10, "mode": "other"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/activity", bytes.NewReader(raw))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowActivity(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["accepted"] != false || resp["reason"] != "daily_cap" {
		t.Fatalf("expected daily_cap rejection, got %+v", resp)
	}
}

func TestLinglowDistrictExtras(t *testing.T) {
	router, userID := setupLinglowHandlersTest(t, 900303)
	conn := router.db

	if _, err := conn.Exec(`
		INSERT INTO reading_texts (text_id, title, category_id, level, target_language, reading_passage)
		VALUES ('read.discovery.test', 'Discovery text', 'cat.a0', 'A0', 'es', 'Hola mundo.')
		ON CONFLICT (text_id) DO NOTHING
	`); err != nil {
		t.Fatalf("insert reading text: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/district-extras?district_code=a0_spark_gate&course_code=es_ru", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowDistrictExtras(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Discovery map[string]interface{}   `json:"discovery"`
		Tasks     []map[string]interface{} `json:"tasks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Discovery == nil || body.Discovery["text_id"] != "read.discovery.test" {
		t.Fatalf("expected discovery, got %+v", body.Discovery)
	}
	if len(body.Tasks) == 0 {
		t.Fatal("expected tasks block")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/district-extras", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowDistrictExtras(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing district expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/district-extras?district_code=missing_district", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowDistrictExtras(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown district expected 404, got %d", w.Code)
	}
}

func TestLinglowStats(t *testing.T) {
	router, userID := setupLinglowHandlersTest(t, 900304)
	statsRepo := repository.NewLinglowDailyStatsRepository(router.db)
	userCourseID, err := statsRepo.ResolveUserCourseID(t.Context(), userID, "es_ru")
	if err != nil {
		t.Fatalf("ResolveUserCourseID: %v", err)
	}
	day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if err := statsRepo.Bump(t.Context(), repository.DailyBump{
		UserCourseID: userCourseID, Day: day, Mode: "grammar", Attempts: 3, Correct: 2, ActiveSeconds: 120,
	}); err != nil {
		t.Fatalf("Bump: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/stats?course_code=es_ru", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowStats(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var stats map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stats["streak"] == nil {
		t.Fatal("expected streak in stats")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/stats?course_code=xx_ru", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowStats(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown course expected 404, got %d", w.Code)
	}

	router.courseRepo = nil
	req = httptest.NewRequest(http.MethodGet, "/api/linglow/stats", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowStats(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("nil course repo expected 503, got %d", w.Code)
	}
}
