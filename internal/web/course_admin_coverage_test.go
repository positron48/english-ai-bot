package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

const courseAdminCoverageTelegramID int64 = 900001

func setupCourseAdminCoverageRouter(t *testing.T, learning config.LearningConfig, linglow config.LinglowConfig) (*Router, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: learning, Linglow: linglow}
	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(courseAdminCoverageTelegramID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return NewRouter(logger, cfg, db, nil, nil, nil, nil), user.ID
}

func setupCourseAdminCoverageAdminRouter(t *testing.T) (*Router, int64, *repository.UserRepository) {
	t.Helper()
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{
		Admin:    config.AdminConfig{TelegramID: courseAdminCoverageTelegramID},
		WebApp:   config.WebAppConfig{JWTSecret: "test-secret"},
		Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"},
	}
	cbRepo := repository.NewCircuitBreakerRepository(db, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	userRepo := repository.NewUserRepository(db, logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	adminUser, err := userRepo.GetOrCreateUser(courseAdminCoverageTelegramID)
	if err != nil {
		t.Fatalf("create admin user: %v", err)
	}
	return router, adminUser.ID, userRepo
}

// ── course.go helpers ─────────────────────────────────────────────────────────

func TestCourseAdminCoverage_CourseHelperNilPaths(t *testing.T) {
	var nilRouter *Router
	ctx := context.Background()
	if got := nilRouter.learningConfigForUser(ctx, 1); got.TargetLang != "en" {
		t.Fatalf("nil router learning config = %+v", got)
	}
	if got := nilRouter.defaultCourseCode(); got != "" {
		t.Fatalf("nil router default course = %q", got)
	}
	if got := nilRouter.currentCourseCodeForUser(ctx, 1); got != "" {
		t.Fatalf("nil router current course = %q", got)
	}
	req := httptest.NewRequest(http.MethodGet, "/?course_code=en_ru", nil)
	if got := nilRouter.requestedCourseCodeForUser(req, 1); got != "" {
		t.Fatalf("nil router requested course = %q", got)
	}

	router, userID := setupCourseAdminCoverageRouter(t, config.DefaultLearningConfig(), config.LinglowConfig{})
	router.config = nil
	if got := router.learningConfigForUser(ctx, userID); got.TargetLang != "en" {
		t.Fatalf("nil config learning = %+v", got)
	}
	if got := router.defaultCourseCode(); got != "" {
		t.Fatalf("nil config default course = %q", got)
	}
	router.courseRepo = nil
	if got := router.currentCourseCodeForUser(ctx, userID); got != "" {
		t.Fatalf("nil courseRepo current course = %q", got)
	}
	if got := router.requestedCourseCodeForUser(req, userID); got != "" {
		t.Fatalf("nil courseRepo requested course = %q", got)
	}
	if got := router.requestedCourseCodeForUser(req, 0); got != "" {
		t.Fatalf("userID 0 requested course = %q", got)
	}
}

func TestCourseAdminCoverage_CourseAPINilRepoAndAuth(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, config.LinglowConfig{})

	handlers503 := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		method  string
		path    string
	}{
		{"courses", router.handleCourses, http.MethodGet, "/api/courses"},
		{"current", router.handleCurrentCourse, http.MethodGet, "/api/user/courses/current"},
		{"select", router.handleSelectCourse, http.MethodPost, "/api/user/courses/select"},
		{"city", router.handleCourseMap, http.MethodGet, "/api/linglow/city"},
		{"daily", router.handleLinglowDailyRoute, http.MethodGet, "/api/linglow/daily-route"},
		{"review", router.handleLinglowReview, http.MethodGet, "/api/linglow/review"},
		{"progress", router.handleLinglowProgress, http.MethodGet, "/api/linglow/progress"},
		{"shadow", router.handleLinglowSRSShadow, http.MethodGet, "/api/linglow/srs-shadow"},
		{"attempts", router.handleLinglowExerciseAttempts, http.MethodPost, "/api/linglow/exercise-attempts"},
		{"words", router.handleLinglowWords, http.MethodGet, "/api/linglow/words"},
		{"history", router.handleLinglowHistory, http.MethodGet, "/api/linglow/history"},
		{"srs-readiness", router.handleAdminLinglowSRSReadiness, http.MethodGet, "/api/admin/linglow/srs-readiness"},
	}
	router.courseRepo = nil
	for _, h := range handlers503 {
		t.Run(h.name+"_503", func(t *testing.T) {
			var req *http.Request
			if h.method == http.MethodPost {
				req = httptest.NewRequest(h.method, h.path, bytes.NewReader([]byte(`{"course_code":"es_ru"}`)))
			} else {
				req = httptest.NewRequest(h.method, h.path, nil)
			}
			req = setUserIDInContext(req, userID)
			w := httptest.NewRecorder()
			h.handler(w, req)
			if w.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
			}
		})
	}

	router2, _ := setupCourseAdminCoverageRouter(t, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, config.LinglowConfig{})
	handlers401 := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		method  string
		path    string
	}{
		{"courses", router2.handleCourses, http.MethodGet, "/api/courses"},
		{"current", router2.handleCurrentCourse, http.MethodGet, "/api/user/courses/current"},
		{"select", router2.handleSelectCourse, http.MethodPost, "/api/user/courses/select"},
		{"city", router2.handleCourseMap, http.MethodGet, "/api/linglow/city"},
		{"daily", router2.handleLinglowDailyRoute, http.MethodGet, "/api/linglow/daily-route"},
		{"review", router2.handleLinglowReview, http.MethodGet, "/api/linglow/review"},
		{"progress", router2.handleLinglowProgress, http.MethodGet, "/api/linglow/progress"},
		{"shadow", router2.handleLinglowSRSShadow, http.MethodGet, "/api/linglow/srs-shadow"},
		{"attempts", router2.handleLinglowExerciseAttempts, http.MethodPost, "/api/linglow/exercise-attempts"},
		{"words", router2.handleLinglowWords, http.MethodGet, "/api/linglow/words"},
		{"history", router2.handleLinglowHistory, http.MethodGet, "/api/linglow/history"},
	}
	for _, h := range handlers401 {
		t.Run(h.name+"_401", func(t *testing.T) {
			var req *http.Request
			if h.method == http.MethodPost {
				req = httptest.NewRequest(h.method, h.path, bytes.NewReader([]byte(`{"course_code":"es_ru","mode":"x"}`)))
			} else {
				req = httptest.NewRequest(h.method, h.path, nil)
			}
			w := httptest.NewRecorder()
			h.handler(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
		})
	}
}

func TestCourseAdminCoverage_CourseAPIMethodNotAllowed(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, config.LinglowConfig{})
	postOnly := []func(http.ResponseWriter, *http.Request){
		router.handleCourses,
		router.handleCurrentCourse,
		router.handleCourseMap,
		router.handleLinglowDailyRoute,
		router.handleLinglowReview,
		router.handleLinglowProgress,
		router.handleLinglowSRSShadow,
		router.handleLinglowWords,
		router.handleLinglowHistory,
		router.handleAdminLinglowSRSReadiness,
	}
	for i, h := range postOnly {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		h(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("handler %d: expected 405, got %d", i, w.Code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleSelectCourse(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("select GET expected 405, got %d", w.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowExerciseAttempts(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("attempts GET expected 405, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_CourseNotFoundPaths(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, config.LinglowConfig{})
	paths := []string{
		"/api/linglow/city?course_code=xx_ru",
		"/api/linglow/daily-route?course_code=xx_ru",
		"/api/linglow/review?course_code=xx_ru",
		"/api/linglow/progress?course_code=xx_ru",
		"/api/linglow/srs-shadow?course_code=xx_ru",
		"/api/linglow/words?course_code=xx_ru",
		"/api/linglow/history?course_code=xx_ru",
		"/api/admin/linglow/srs-readiness?course_code=xx_ru",
	}
	handlers := []func(http.ResponseWriter, *http.Request){
		router.handleCourseMap,
		router.handleLinglowDailyRoute,
		router.handleLinglowReview,
		router.handleLinglowProgress,
		router.handleLinglowSRSShadow,
		router.handleLinglowWords,
		router.handleLinglowHistory,
		router.handleAdminLinglowSRSReadiness,
	}
	for i, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		handlers[i](w, req)
		want := http.StatusNotFound
		if strings.Contains(path, "srs-readiness") {
			want = http.StatusInternalServerError
		}
		if w.Code != want {
			t.Fatalf("%s expected %d, got %d: %s", path, want, w.Code, w.Body.String())
		}
	}

	body := `{"course_code":"xx_ru","mode":"card","prompt":{},"answer":{},"result":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/exercise-attempts", strings.NewReader(body))
	req = setUserIDInContext(req, userID)
	w := awaitRecorder(router.handleLinglowExerciseAttempts, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("exercise attempt unknown course expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func awaitRecorder(h func(http.ResponseWriter, *http.Request), req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h(w, req)
	return w
}

func TestCourseAdminCoverage_CourseHandlersDBError(t *testing.T) {
	logger := zap.NewNop()
	brokenDB := newBrokenDB(t)
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	router := NewRouter(logger, cfg, brokenDB, nil, nil, nil, nil)
	userID := int64(900002)

	tests := []struct {
		method string
		path   string
		body   string
		fn     func(http.ResponseWriter, *http.Request)
	}{
		{http.MethodGet, "/api/courses", "", router.handleCourses},
		{http.MethodGet, "/api/user/courses/current", "", router.handleCurrentCourse},
		{http.MethodPost, "/api/user/courses/select", `{"course_code":"es_ru"}`, router.handleSelectCourse},
		{http.MethodGet, "/api/linglow/city", "", router.handleCourseMap},
		{http.MethodGet, "/api/linglow/daily-route", "", router.handleLinglowDailyRoute},
		{http.MethodGet, "/api/linglow/review", "", router.handleLinglowReview},
		{http.MethodGet, "/api/linglow/progress", "", router.handleLinglowProgress},
		{http.MethodGet, "/api/linglow/srs-shadow", "", router.handleLinglowSRSShadow},
		{http.MethodGet, "/api/admin/linglow/srs-readiness", "", router.handleAdminLinglowSRSReadiness},
		{http.MethodPost, "/api/linglow/exercise-attempts", `{"mode":"x","prompt":{},"answer":{},"result":{}}`, router.handleLinglowExerciseAttempts},
		{http.MethodGet, "/api/linglow/words", "", router.handleLinglowWords},
		{http.MethodGet, "/api/linglow/history", "", router.handleLinglowHistory},
	}
	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			req = setUserIDInContext(req, userID)
			w := httptest.NewRecorder()
			tc.fn(w, req)
			if w.Code != http.StatusInternalServerError && w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
				t.Fatalf("expected 5xx/404/400, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestCourseAdminCoverage_ExerciseAttemptsValidation(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, config.LinglowConfig{})

	cases := []struct {
		body       string
		wantStatus int
	}{
		{`{"mode":"x","prompt":{},"result":[]}`, http.StatusBadRequest},
		{`{"mode":"x","prompt":{},"answer":[]}`, http.StatusBadRequest},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/linglow/exercise-attempts", strings.NewReader(c.body))
		req = setUserIDInContext(req, userID)
		w := httptest.NewRecorder()
		router.handleLinglowExerciseAttempts(w, req)
		if w.Code != c.wantStatus {
			t.Fatalf("body %s: expected %d, got %d", c.body, c.wantStatus, w.Code)
		}
	}
}

func TestCourseAdminCoverage_DailyRouteSRSReadAndEnrich(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t,
		config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"},
		config.LinglowConfig{SRSReadEnabled: true},
	)

	contentRepo := repository.NewGrammarContentRepository(zap.NewNop())
	publishRepo := repository.NewGrammarPublishRepository(router.db, router.logger)
	attemptRepo := repository.NewGrammarAttemptRepository(router.db, router.logger)
	grammarSvc := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, router.logger)
	grammarSvc.TheoryIndex = &repository.TheoryBlockIndex{
		ByBlockID: map[string]*repository.TheoryBlockInfo{
			"block-a": {Title: "Present Tense"},
		},
	}
	router.SetGrammarService(grammarSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/daily-route?limit=5", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowDailyRoute(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("daily route status=%d body=%s", w.Code, w.Body.String())
	}

	items := []repository.DailyRouteItem{
		{SourceKind: "grammar_theory_block", SourceID: "ch:block-a", Title: "block-a"},
		{SourceKind: "grammar_theory_block", SourceID: "no-colon", Title: "unchanged"},
		{SourceKind: "reading_text", SourceID: "x:y", Title: "reading"},
	}
	router.enrichTheoryBlockTitles(req, userID, items)
	if items[0].Title != "Present Tense" {
		t.Fatalf("enriched title = %q", items[0].Title)
	}
	if items[1].Title != "unchanged" || items[2].Title != "reading" {
		t.Fatalf("unexpected titles: %+v", items)
	}
}

func TestCourseAdminCoverage_AdminSRSReadinessDefaultCourse(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t,
		config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"},
		config.LinglowConfig{SRSReadEnabled: true, SRSWriteEnabled: true},
	)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/linglow/srs-readiness", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleAdminLinglowSRSReadiness(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		SRSReadEnabled  bool `json:"srs_read_enabled"`
		SRSWriteEnabled bool `json:"srs_write_enabled"`
		CanEnableRead   bool `json:"can_enable_srs_read"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.SRSReadEnabled || !body.SRSWriteEnabled || body.CanEnableRead {
		t.Fatalf("unexpected flags: %+v", body)
	}
}

func TestCourseAdminCoverage_LinglowWordsAndHistoryFilters(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t,
		config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"},
		config.LinglowConfig{},
	)

	req := httptest.NewRequest(http.MethodGet, "/api/linglow/words?offset=-5&status=new&sort=lemma&q=test", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowWords(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("words status=%d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/history?days=30", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowHistory(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/linglow/history?days=9999", nil)
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowHistory(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("history capped days status=%d", w.Code)
	}
}

func TestCourseAdminCoverage_SelectCourseTrimLower(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t,
		config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"},
		config.LinglowConfig{},
	)
	req := httptest.NewRequest(http.MethodPost, "/api/user/courses/select", strings.NewReader(`{"course_code":" EN_RU "}`))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleSelectCourse(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("select status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_ParsePositiveLimitDefault(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	limit, ok := parsePositiveLimit(w, req, 42)
	if !ok || limit != 42 {
		t.Fatalf("default limit = %d ok=%v", limit, ok)
	}
}

func TestCourseAdminCoverage_RawObjectJSONEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	got, ok := rawObjectJSON(w, nil, "answer")
	if !ok || got != "{}" {
		t.Fatalf("empty raw = %q ok=%v", got, ok)
	}
}

// ── admin.go ──────────────────────────────────────────────────────────────────

func TestCourseAdminCoverage_ValidAdminDefinitionNative(t *testing.T) {
	if !requireNativeCyrillicForLearningPair("ru", "es") {
		t.Fatal("expected ru/es pair to require cyrillic")
	}
	if requireNativeCyrillicForLearningPair("ru", "en") {
		t.Fatal("expected ru/en pair not to require cyrillic")
	}
	if validAdminDefinitionNative("hello", "ru", "es") {
		t.Fatal("latin-only definition should fail for es/ru")
	}
	if !validAdminDefinitionNative("привет", "ru", "es") {
		t.Fatal("cyrillic definition should pass for es/ru")
	}
	if !validAdminDefinitionNative("hello", "ru", "en") {
		t.Fatal("any definition should pass for en target")
	}
}

func TestCourseAdminCoverage_TTSCircuitHandlers(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/circuit/tts", nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTTSCircuitStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("nil tts cb status=%d", w.Code)
	}
	var nilResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &nilResp)
	cb, _ := nilResp["circuit_breaker"].(map[string]interface{})
	if cb["state"] != "not_configured" {
		t.Fatalf("expected not_configured, got %v", cb["state"])
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/circuit/tts/reset", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminTTSCircuitReset(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("nil tts cb reset expected 404, got %d", w.Code)
	}

	cbRepo := repository.NewCircuitBreakerRepository(router.db, router.logger)
	ttsCb := service.NewCircuitBreakerService(cbRepo, 3, router.logger)
	router.SetTTSCircuitBreaker(ttsCb)

	req = httptest.NewRequest(http.MethodGet, "/api/admin/circuit/tts", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminTTSCircuitStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("configured tts status=%d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/circuit/tts/reset", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminTTSCircuitReset(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("tts reset status=%d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/circuit/tts", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminTTSCircuitStatus(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("tts status POST expected 405, got %d", w.Code)
	}

	brokenDB := newBrokenDB(t)
	brokenCbRepo := repository.NewCircuitBreakerRepository(brokenDB, router.logger)
	brokenTtsCb := service.NewCircuitBreakerService(brokenCbRepo, 3, router.logger)
	brokenRouter := NewRouter(router.logger, router.config, brokenDB, nil, nil, nil, nil)
	brokenRouter.SetTTSCircuitBreaker(brokenTtsCb)
	req = httptest.NewRequest(http.MethodGet, "/api/admin/circuit/tts", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	brokenRouter.handleAdminTTSCircuitStatus(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("broken tts status expected 500, got %d", w.Code)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/admin/circuit/tts/reset", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	brokenRouter.handleAdminTTSCircuitReset(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("broken tts reset expected 500, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_CircuitOpenSuccessAndError(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/circuit/open", nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminCircuitOpen(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("open status=%d body=%s", w.Code, w.Body.String())
	}

	brokenDB := newBrokenDB(t)
	brokenCbRepo := repository.NewCircuitBreakerRepository(brokenDB, router.logger)
	brokenCb := service.NewCircuitBreakerService(brokenCbRepo, 5, router.logger)
	brokenRouter := NewRouter(router.logger, router.config, brokenDB, nil, nil, nil, brokenCb)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/circuit/open", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	brokenRouter.handleAdminCircuitOpen(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("broken open expected 500, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminUsersWithGrammarPlacement(t *testing.T) {
	router, adminUserID, userRepo := setupCourseAdminCoverageAdminRouter(t)
	targetUser, err := userRepo.GetOrCreateUser(900003)
	if err != nil {
		t.Fatalf("create target user: %v", err)
	}
	_, err = router.db.Exec(`
		INSERT INTO grammar_placement_test (user_id, score, total_questions, opened_sections_json, completed_at, admin_override)
		VALUES ($1, 80, 10, $2, CURRENT_TIMESTAMP, true)
		ON CONFLICT (user_id) DO UPDATE SET score = 80, total_questions = 10, opened_sections_json = $2, admin_override = true
	`, targetUser.ID, `["A1"]`)
	if err != nil {
		t.Fatalf("insert placement: %v", err)
	}

	contentRepo := repository.NewGrammarContentRepository(router.logger)
	publishRepo := repository.NewGrammarPublishRepository(router.db, router.logger)
	attemptRepo := repository.NewGrammarAttemptRepository(router.db, router.logger)
	grammarSvc := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, router.config.Learning, router.logger)
	router.SetGrammarService(grammarSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminUsers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("users status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Users []map[string]interface{} `json:"users"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, u := range resp.Users {
		if int64(u["id"].(float64)) == targetUser.ID {
			if u["grammar_placement"] == nil {
				t.Fatalf("expected grammar_placement for user %d", targetUser.ID)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("target user not found in admin users list")
	}

	_, err = router.db.Exec(`
		UPDATE grammar_placement_test SET opened_sections_json = 'not-json' WHERE user_id = $1
	`, targetUser.ID)
	if err != nil {
		t.Fatalf("update placement json: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminUsers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("users after bad json status=%d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminUserSubroutes(t *testing.T) {
	router, adminUserID, userRepo := setupCourseAdminCoverageAdminRouter(t)
	targetUser, err := userRepo.GetOrCreateUser(900004)
	if err != nil {
		t.Fatalf("create target user: %v", err)
	}

	contentRepo := repository.NewGrammarContentRepository(router.logger)
	publishRepo := repository.NewGrammarPublishRepository(router.db, router.logger)
	attemptRepo := repository.NewGrammarAttemptRepository(router.db, router.logger)
	grammarSvc := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, router.config.Learning, router.logger)
	router.SetGrammarService(grammarSvc)

	cases := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{http.MethodGet, "/api/admin/users/", "", http.StatusNotFound},
		{http.MethodPut, "/api/admin/users/not-id/grammar-placement", `{}`, http.StatusBadRequest},
		{http.MethodPut, "/api/admin/users/999999/grammar-placement", `{}`, http.StatusNotFound},
		{http.MethodPut, fmt.Sprintf("/api/admin/users/%d/grammar-placement", targetUser.ID), `{bad`, http.StatusBadRequest},
		{http.MethodPut, fmt.Sprintf("/api/admin/users/%d/grammar-placement", targetUser.ID), `{"level":"nope"}`, http.StatusBadRequest},
		{http.MethodPut, fmt.Sprintf("/api/admin/users/%d/grammar-placement", targetUser.ID), `{"level":"A1"}`, http.StatusOK},
		{http.MethodGet, fmt.Sprintf("/api/admin/users/%d/grammar-placement", targetUser.ID), "", http.StatusMethodNotAllowed},
		{http.MethodPatch, fmt.Sprintf("/api/admin/users/%d/subscription-tier", targetUser.ID), `{bad`, http.StatusBadRequest},
		{http.MethodPatch, fmt.Sprintf("/api/admin/users/%d/subscription-tier", targetUser.ID), `{"tier":"nope"}`, http.StatusBadRequest},
		{http.MethodPatch, fmt.Sprintf("/api/admin/users/%d/subscription-tier", targetUser.ID), `{"tier":"pro"}`, http.StatusOK},
		{http.MethodGet, fmt.Sprintf("/api/admin/users/%d/unknown", targetUser.ID), "", http.StatusNotFound},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			var req *http.Request
			if c.body != "" {
				req = httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
			} else {
				req = httptest.NewRequest(c.method, c.path, nil)
			}
			req = adminCtx(req, adminUserID)
			w := httptest.NewRecorder()
			router.handleAdminUserSubroutes(w, req)
			if w.Code != c.want {
				t.Fatalf("expected %d, got %d: %s", c.want, w.Code, w.Body.String())
			}
		})
	}

	noGrammar := setupCourseAdminCoverageAdminRouterOnly(t)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/users/%d/grammar-placement", targetUser.ID), strings.NewReader(`{"level":"A1"}`))
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	noGrammar.handleAdminUserSubroutes(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("no grammar service expected 503, got %d", w.Code)
	}
}

func setupCourseAdminCoverageAdminRouterOnly(t *testing.T) *Router {
	t.Helper()
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{
		Admin:    config.AdminConfig{TelegramID: courseAdminCoverageTelegramID},
		WebApp:   config.WebAppConfig{JWTSecret: "test-secret"},
		Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"},
	}
	cbRepo := repository.NewCircuitBreakerRepository(db, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	userRepo := repository.NewUserRepository(db, logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	return router
}

func TestCourseAdminCoverage_AdminWordGenerateSpanishValidation(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	wordRepo := repository.NewWordRepository(router.db, router.logger)
	_ = wordRepo.SaveWordCard("hola", "hi", "")
	wc, _ := wordRepo.GetWordCard("hola")

	router.aiService = setupAdminAIService(t, `{"lemma":"hola","pos":"interj","transcription":"ola","definition_ru":"hello","examples":[]}`)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("non-cyrillic definition expected 400, got %d: %s", w.Code, w.Body.String())
	}

	router.aiService = setupAdminAIService(t, `{"lemma":"hola","pos":"interj","transcription":"","definition_ru":"привет","examples":[]}`)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty transcription expected 400, got %d", w.Code)
	}

	router.aiService = setupAdminAIService(t, `{"lemma":"hola","pos":"interj","transcription":"ola","definition_ru":"привет","examples":[{"en":"Hi","ru":"Привет"}]}`)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("valid generate expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminWordDuplicatePrecheck(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	wordRepo := repository.NewWordRepository(router.db, router.logger)
	_ = wordRepo.SaveWordCard("dupa", "one", "")
	_ = wordRepo.SaveWordCard("dupb", "two", "")
	wcA, _ := wordRepo.GetWordCard("dupa")

	body := `{"word":"dupb","definition":"one"}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wcA.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate precheck expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminWordsExtraFilters(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	paths := []string{
		"/api/admin/words?course_code=ES_RU&missing_training_pos=verb&search=test&has_audio=maybe&sort_order=asc&limit=10&offset=0&user_id=not-a-number",
		"/api/admin/words?only_errors=true&has_audio=0",
	}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req = adminCtx(req, adminUserID)
		w := httptest.NewRecorder()
		router.handleAdminWords(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d: %s", path, w.Code, w.Body.String())
		}
	}
}

func TestCourseAdminCoverage_AdminTrainingRemainingBranches(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	_ = wordRepo.SaveWordCard("genbranch", "def", "")

	router.aiService = setupAdminAIService(t, `{"word_en":"genbranch","transcription":"g","senses":[{"pos":"n","display_word":"genbranch","word_ru":"с","meaning_en":"m","example_en":"","example_ru":"","distractors_ru":[],"distractors_en":[],"hint":""}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/genbranch/generate", strings.NewReader(`{"constraints":""}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("generate status=%d body=%s", w.Code, w.Body.String())
	}

	router.aiService = &ai.Service{}
	req = httptest.NewRequest(http.MethodPost, "/api/admin/training/genbranch/generate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("wrong ai type expected 500, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/admin/training//delete", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty word delete expected 400, got %d", w.Code)
	}

	nonAdmin, _ := repository.NewUserRepository(db.GetConnection(), router.logger).GetOrCreateUser(900005)
	req = httptest.NewRequest(http.MethodGet, "/api/admin/training/genbranch", nil)
	req = setUserIDInContext(req, nonAdmin.ID)
	w = httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("non-admin GET expected 403, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminTrainingCreateWithExistingUserCards(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)
	conn := db.GetConnection()
	wordRepo := repository.NewWordRepository(conn, router.logger)
	_ = wordRepo.SaveWordCard("sharedword", "def", "")
	wc, _ := wordRepo.GetWordCard("sharedword")
	tcRepo := repository.NewTrainingCardRepository(conn, router.logger)
	cardID, _ := tcRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wc.ID, WordEN: "sharedword", SenseIndex: 0,
		WordRU: "с", MeaningEN: "m", DistractorsRU: "[]", DistractorsEN: "[]",
	})
	learner, _ := repository.NewUserRepository(conn, router.logger).GetOrCreateUser(900006)
	ucRepo := repository.NewUserCardRepository(conn, router.logger)
	_, _ = ucRepo.CreateUserCard(&models.UserCard{
		UserID: learner.ID, TrainingCardID: cardID, Direction: models.DirectionRUtoEN,
		State: models.StateNew, EF: models.InitialEF,
	})

	body := `{"word_ru":"новый","meaning_en":"new sense"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/sharedword", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create with existing users status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminTTSCircuitOpenState(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	cbRepo := repository.NewCircuitBreakerRepository(router.db, router.logger)
	ttsCb := service.NewCircuitBreakerService(cbRepo, 1, router.logger)
	_ = ttsCb.Open()
	router.SetTTSCircuitBreaker(ttsCb)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/circuit/tts", nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTTSCircuitStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("open tts status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	cb, _ := resp["circuit_breaker"].(map[string]interface{})
	if cb["state"] != "open" {
		t.Fatalf("expected open state, got %v", cb["state"])
	}
}

func TestCourseAdminCoverage_CurrentCourseCodeResolveError(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "en", GrammarBundleID: "en"}}
	router := NewRouter(logger, cfg, conn, nil, nil, service.NewOptionsService(repository.NewTrainingCardRepository(conn, logger), logger, "en"), nil)
	for _, pair := range []struct{ word, course string }{
		{"covapple", "en_ru"},
		{"covmanzana", "es_ru"},
	} {
		var wcID int64
		if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id`, pair.word, pair.word).Scan(&wcID); err != nil {
			t.Fatalf("insert word: %v", err)
		}
		if _, err := conn.Exec(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, course_code) VALUES ($1, $2, 0, $3, $4, $5)`,
			wcID, pair.word, pair.word, pair.word, pair.course); err != nil {
			t.Fatalf("insert training card: %v", err)
		}
	}
	if code := router.currentCourseCodeForUser(context.Background(), 999999); code != "" {
		t.Fatalf("expected empty code on resolve error, got %q", code)
	}
}

func TestCourseAdminCoverage_CourseMapEncodeError(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, config.LinglowConfig{})
	req := httptest.NewRequest(http.MethodGet, "/api/linglow/city", nil)
	req = setUserIDInContext(req, userID)
	rec := httptest.NewRecorder()
	router.handleCourseMap(&failingResponseWriter{ResponseWriter: rec}, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 before encode fail, got %d", rec.Code)
	}
}

func TestCourseAdminCoverage_EnrichTheoryBlockNoGrammarService(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, config.LinglowConfig{})
	items := []repository.DailyRouteItem{{SourceKind: "grammar_theory_block", SourceID: "ch:missing", Title: "slug"}}
	router.enrichTheoryBlockTitles(httptest.NewRequest(http.MethodGet, "/", nil), userID, items)
	if items[0].Title != "slug" {
		t.Fatalf("nil grammar service should leave title unchanged")
	}

	contentRepo := repository.NewGrammarContentRepository(router.logger)
	publishRepo := repository.NewGrammarPublishRepository(router.db, router.logger)
	attemptRepo := repository.NewGrammarAttemptRepository(router.db, router.logger)
	grammarSvc := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, router.config.Learning, router.logger)
	grammarSvc.TheoryIndex = &repository.TheoryBlockIndex{ByBlockID: map[string]*repository.TheoryBlockInfo{}}
	router.SetGrammarService(grammarSvc)
	router.enrichTheoryBlockTitles(httptest.NewRequest(http.MethodGet, "/", nil), userID, items)
	if items[0].Title != "slug" {
		t.Fatalf("missing block index entry should leave title unchanged")
	}
}

func TestCourseAdminCoverage_AdminSRSReadinessCourseNotFound(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t, config.DefaultLearningConfig(), config.LinglowConfig{})
	router.config = nil
	req := httptest.NewRequest(http.MethodGet, "/api/admin/linglow/srs-readiness", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleAdminLinglowSRSReadiness(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("empty course expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_ExerciseAttemptsSuccessWithAnsweredAt(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t,
		config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"},
		config.LinglowConfig{SRSWriteEnabled: true},
	)
	var itemID int64
	if err := router.db.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
			WHERE c.code = 'es_ru' LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'grammar:cov-at', 'grammar', 'Cov', 'grammar_category', 'cov-at', 1, 'published'
			FROM target RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'grammar_theory_block', 'grammar_theory_block', 'cov-at-item', 'Block', 'A0', 'published'
		FROM module RETURNING id
	`).Scan(&itemID); err != nil {
		t.Fatalf("insert item: %v", err)
	}
	body := fmt.Sprintf(`{"learning_item_id":%d,"mode":"grammar","client_attempt_id":"cov-at-1","is_correct":true,"prompt":{},"answer":{},"result":{},"answered_at":"2026-06-28T10:00:00Z"}`, itemID)
	req := httptest.NewRequest(http.MethodPost, "/api/linglow/exercise-attempts", strings.NewReader(body))
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLinglowExerciseAttempts(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attempt status=%d body=%s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/exercise-attempts", strings.NewReader(`{`))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowExerciseAttempts(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json expected 400, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/linglow/exercise-attempts", strings.NewReader(`{"mode":"grammar","learning_item_id":999999,"prompt":{},"answer":{},"result":{}}`))
	req = setUserIDInContext(req, userID)
	w = httptest.NewRecorder()
	router.handleLinglowExerciseAttempts(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid attempt expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminTTSCircuitLastFailureAndResetMethod(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	cbRepo := repository.NewCircuitBreakerRepository(router.db, router.logger)
	ttsCb := service.NewCircuitBreakerService(cbRepo, 1, router.logger)
	_ = ttsCb.RecordFailure("tts provider down")
	router.SetTTSCircuitBreaker(ttsCb)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/circuit/tts", nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTTSCircuitStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	cb, _ := resp["circuit_breaker"].(map[string]interface{})
	if cb["last_failure"] == nil {
		t.Fatalf("expected last_failure in response: %+v", cb)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/admin/circuit/tts/reset", nil)
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminTTSCircuitReset(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("reset GET expected 405, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminUserSubscriptionTierAllPaths(t *testing.T) {
	router, adminUserID, userRepo := setupCourseAdminCoverageAdminRouter(t)
	targetUser, _ := userRepo.GetOrCreateUser(900007)

	for _, method := range []string{http.MethodPut, http.MethodPost, http.MethodPatch} {
		req := httptest.NewRequest(method, fmt.Sprintf("/api/admin/users/%d/subscription-tier", targetUser.ID), strings.NewReader(`{"tier":"pro"}`))
		req = adminCtx(req, adminUserID)
		w := httptest.NewRecorder()
		router.handleAdminUserSubroutes(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d: %s", method, w.Code, w.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/admin/users/%d/subscription-tier", targetUser.ID), nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminUserSubroutes(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET tier expected 405, got %d", w.Code)
	}

	router.userRepo = struct{}{}
	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/admin/users/%d/subscription-tier", targetUser.ID), strings.NewReader(`{"tier":"pro"}`))
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	router.handleAdminUserSubroutes(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("bad userRepo expected 500, got %d", w.Code)
	}

	brokenDB := newBrokenDB(t)
	brokenRouter := NewRouter(router.logger, router.config, brokenDB, nil, nil, nil, nil)
	brokenRouter.userRepo = repository.NewUserRepository(brokenDB, router.logger)
	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/admin/users/%d/subscription-tier", targetUser.ID), strings.NewReader(`{"tier":"free"}`))
	req = adminCtx(req, adminUserID)
	w = httptest.NewRecorder()
	brokenRouter.handleAdminUserSubroutes(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("broken DB tier update expected 500, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminUserGrammarPlacementErrors(t *testing.T) {
	router, adminUserID, userRepo := setupCourseAdminCoverageAdminRouter(t)
	targetUser, _ := userRepo.GetOrCreateUser(900008)
	contentRepo := repository.NewGrammarContentRepository(router.logger)
	publishRepo := repository.NewGrammarPublishRepository(router.db, router.logger)
	attemptRepo := repository.NewGrammarAttemptRepository(router.db, router.logger)
	grammarSvc := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, router.config.Learning, router.logger)
	router.SetGrammarService(grammarSvc)

	brokenDB := newBrokenDB(t)
	brokenRouter := NewRouter(router.logger, router.config, brokenDB, nil, nil, nil, nil)
	brokenRouter.SetGrammarService(grammarSvc)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/users/%d/grammar-placement", targetUser.ID), strings.NewReader(`{"level":"A1"}`))
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	brokenRouter.handleAdminUserSubroutes(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("broken user lookup expected 500, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminWordsAndWordRemaining(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	wordRepo := repository.NewWordRepository(router.db, router.logger)
	_ = wordRepo.SaveWordCard("genderword", "def", "")
	wc, _ := wordRepo.GetWordCard("genderword")

	form := "word=genderword&definition=def&pos=noun&noun_gender=m&opposite_gender_word=genderword2&transcription=/g/"
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wc.ID), strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("form PUT expected 200, got %d: %s", w.Code, w.Body.String())
	}

	brokenRouter, brokenAdminID := setupAdminRouterWithBrokenDB(t)
	req = httptest.NewRequest(http.MethodGet, "/api/admin/words", nil)
	req = adminCtx(req, brokenAdminID)
	w = httptest.NewRecorder()
	brokenRouter.handleAdminWords(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("broken list words expected 500, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPut, "/api/admin/words/999", strings.NewReader(`{"word":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, brokenAdminID)
	w = httptest.NewRecorder()
	brokenRouter.handleAdminWord(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("broken get word expected 500, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminTrainingGenerateWordCardLookupWarn(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	router.aiService = setupAdminAIService(t, `{"word_en":"missing","transcription":"t","senses":[{"pos":"n","display_word":"missing","word_ru":"с","meaning_en":"m","example_en":"","example_ru":"","distractors_ru":[],"distractors_en":[],"hint":""}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/missingword/generate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("generate without word card expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminTrainingInvalidAction(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/someword/unknown", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unknown action expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminTrainingEmptyLemmaFallback(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	_ = wordRepo.SaveWordCard("lemmaback", "def", "")
	router.aiService = setupAdminAIService(t, `{"word_en":"","transcription":"l","senses":[{"pos":"n","display_word":"lemmaback","word_ru":"с","meaning_en":"m","example_en":"","example_ru":"","distractors_ru":[],"distractors_en":[],"hint":""}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/lemmaback/generate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("empty lemma fallback expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminTrainingGetWordCardLookupError(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	brokenDB := newBrokenDB(t)
	cfg := &config.Config{
		Admin:  config.AdminConfig{TelegramID: courseAdminCoverageTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret"},
	}
	cbRepo := repository.NewCircuitBreakerRepository(db, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, brokenDB, nil, nil, nil, cbService)
	userRepo := repository.NewUserRepository(db, logger)
	adminUser, _ := userRepo.GetOrCreateUser(courseAdminCoverageTelegramID)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.aiService = setupAdminAIService(t, `{"word_en":"brokenpath","transcription":"b","senses":[{"pos":"n","display_word":"brokenpath","word_ru":"с","meaning_en":"m","example_en":"","example_ru":"","distractors_ru":[],"distractors_en":[],"hint":""}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/brokenpath/generate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUser.ID)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("generate with broken word lookup expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminWordJSONGenderFields(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	wordRepo := repository.NewWordRepository(router.db, router.logger)
	_ = wordRepo.SaveWordCard("jsongender", "def", "")
	wc, _ := wordRepo.GetWordCard("jsongender")
	body := map[string]interface{}{
		"word":                 "jsongender",
		"definition":           "def",
		"pos":                  "noun",
		"noun_gender":          "m",
		"opposite_gender_word": "jsongender2",
		"transcription":        "/g/",
	}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wc.ID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("JSON gender PUT expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminWordGenerateSpanishVerbDisplay(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	wordRepo := repository.NewWordRepository(router.db, router.logger)
	_ = wordRepo.SaveWordCard("hablar", "def", "")
	wc, _ := wordRepo.GetWordCard("hablar")
	response := `{"lemma":"hablar","pos":"verb","transcription":"a","definition_ru":"говорить","examples":[{"en":"Hi","ru":"Привет"}],"verb_forms":{"v1":"hablar","v2":"habló","v3":"hablado"}}`
	router.aiService = setupAdminAIService(t, response)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Spanish verb generate expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	wc2, _ := resp["word_card"].(map[string]interface{})
	if wc2["display_en"] != "hablar" {
		t.Fatalf("expected bare infinitive display_en, got %v", wc2["display_en"])
	}
}

func TestCourseAdminCoverage_AdminWordUpdateUniqueConstraint(t *testing.T) {
	router, adminUserID, _ := setupCourseAdminCoverageAdminRouter(t)
	wordRepo := repository.NewWordRepository(router.db, router.logger)
	_ = wordRepo.SaveWordCard("uniq1", "def1", "")
	_ = wordRepo.SaveWordCard("uniq2", "def2", "")
	wc1, _ := wordRepo.GetWordCard("uniq1")
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wc1.ID), strings.NewReader(`{"word":"uniq2","definition":"def1"}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("unique constraint expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminWordsCountError(t *testing.T) {
	router, dbWrap, adminUserID := setupAdminRouterWithSecondDB(t)
	conn := dbWrap.GetConnection()

	_, err := conn.Exec(`INSERT INTO word_cards (word, definition) VALUES ('countcovword', 'test def')`)
	if err != nil {
		t.Skipf("cannot insert word card: %v", err)
	}
	_, err = conn.Exec(`
		CREATE OR REPLACE FUNCTION _wc_cov_fail_count_query() RETURNS boolean AS $$
		DECLARE q text;
		BEGIN
			SELECT query INTO q FROM pg_stat_activity WHERE pid = pg_backend_pid();
			IF q ILIKE '%COUNT(DISTINCT wc.id)%' THEN
				RAISE EXCEPTION 'count query blocked for testing';
			END IF;
			RETURN true;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		t.Skipf("cannot create function: %v", err)
	}
	_, err = conn.Exec(`ALTER TABLE word_cards RENAME TO word_cards_real`)
	if err != nil {
		t.Skipf("cannot rename word_cards: %v", err)
	}
	_, err = conn.Exec(`
		CREATE VIEW word_cards AS
		SELECT * FROM word_cards_real wc WHERE _wc_cov_fail_count_query()
	`)
	if err != nil {
		_, _ = conn.Exec(`ALTER TABLE word_cards_real RENAME TO word_cards`)
		t.Skipf("cannot create view: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words", nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWords(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("count error expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminWordGenerateEnglishVerbDisplay(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)
	router.config.Learning = config.LearningConfig{NativeLang: "ru", TargetLang: "en", GrammarBundleID: "en"}
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	_ = wordRepo.SaveWordCard("runword", "def", "")
	wc, _ := wordRepo.GetWordCard("runword")
	response := `{"lemma":"run","pos":"verb","transcription":"rʌn","definition_ru":"бежать","examples":[{"en":"I run","ru":"Я бегу"}],"verb_forms":{"v1":"run","v2":"ran","v3":"run"}}`
	router.aiService = setupAdminAIService(t, response)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("English verb generate expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	wc2, _ := resp["word_card"].(map[string]interface{})
	if wc2["display_en"] != "to run" {
		t.Fatalf("expected display_en 'to run', got %v", wc2["display_en"])
	}
}

func TestCourseAdminCoverage_AdminWordUpdateUniqueAfterLookupError(t *testing.T) {
	router, dbWrap, adminUserID := setupAdminRouterWithSecondDB(t)
	conn := dbWrap.GetConnection()
	wordRepo := repository.NewWordRepository(conn, router.logger)
	_ = wordRepo.SaveWordCard("dupname", "def1", "")
	_ = wordRepo.SaveWordCard("othername", "def2", "")
	wcOther, _ := wordRepo.GetWordCard("othername")

	_, err := conn.Exec(`
		CREATE OR REPLACE FUNCTION _wc_cov_fail_lemma_lookup(w text) RETURNS boolean AS $$
		BEGIN
			IF current_setting('local.fail_lemma', true) = '1' AND w = current_setting('local.target_lemma', true) THEN
				RAISE EXCEPTION 'lemma lookup failed for testing';
			END IF;
			RETURN true;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		t.Skipf("cannot create function: %v", err)
	}
	_, err = conn.Exec(`ALTER TABLE word_cards RENAME TO word_cards_real`)
	if err != nil {
		t.Skipf("cannot rename word_cards: %v", err)
	}
	_, err = conn.Exec(`
		CREATE VIEW word_cards AS
		SELECT * FROM word_cards_real wc WHERE _wc_cov_fail_lemma_lookup(wc.word)
	`)
	if err != nil {
		_, _ = conn.Exec(`ALTER TABLE word_cards_real RENAME TO word_cards`)
		t.Skipf("cannot create view: %v", err)
	}
	_, err = conn.Exec(`SELECT set_config('local.fail_lemma', '1', false)`)
	if err != nil {
		t.Skipf("set fail_lemma: %v", err)
	}
	_, err = conn.Exec(`SELECT set_config('local.target_lemma', 'dupname', false)`)
	if err != nil {
		t.Skipf("set target_lemma: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wcOther.ID), strings.NewReader(`{"word":"dupname","definition":"def2"}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("update unique after lookup error expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_AdminWordDuplicateLookupError(t *testing.T) {
	router, adminUserID := setupAdminRouterWithBrokenDB(t)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/words/1", strings.NewReader(`{"word":"x","definition":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("broken duplicate lookup expected 500, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminWordDeleteNotFound(t *testing.T) {
	router, _, adminUserID := setupAdminTrainingTest(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/words/99999", nil)
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("delete missing word expected 404, got %d", w.Code)
	}
}

func TestCourseAdminCoverage_AdminUserGrammarPlacementSetError(t *testing.T) {
	router, adminUserID, userRepo := setupCourseAdminCoverageAdminRouter(t)
	targetUser, _ := userRepo.GetOrCreateUser(900009)
	contentRepo := repository.NewGrammarContentRepository(router.logger)
	publishRepo := repository.NewGrammarPublishRepository(router.db, router.logger)
	attemptRepo := repository.NewGrammarAttemptRepository(router.db, router.logger)
	grammarSvc := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, router.config.Learning, router.logger)
	router.SetGrammarService(grammarSvc)
	if _, err := router.db.Exec(`DROP TABLE grammar_placement_test`); err != nil {
		t.Fatalf("drop placement table: %v", err)
	}
	t.Cleanup(func() {
		_, _ = router.db.Exec(`CREATE TABLE IF NOT EXISTS grammar_placement_test (
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			score INTEGER NOT NULL,
			total_questions INTEGER NOT NULL,
			opened_sections_json TEXT NOT NULL,
			completed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			admin_override BOOLEAN NOT NULL DEFAULT FALSE,
			UNIQUE(user_id)
		)`)
	})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/users/%d/grammar-placement", targetUser.ID), strings.NewReader(`{"level":"A1"}`))
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminUserSubroutes(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("placement set DB error expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCourseAdminCoverage_EnrichTheoryBlockEmptyTitle(t *testing.T) {
	router, userID := setupCourseAdminCoverageRouter(t, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}, config.LinglowConfig{})
	contentRepo := repository.NewGrammarContentRepository(router.logger)
	publishRepo := repository.NewGrammarPublishRepository(router.db, router.logger)
	attemptRepo := repository.NewGrammarAttemptRepository(router.db, router.logger)
	grammarSvc := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, router.config.Learning, router.logger)
	grammarSvc.TheoryIndex = &repository.TheoryBlockIndex{
		ByBlockID: map[string]*repository.TheoryBlockInfo{
			"empty-title": {Title: ""},
		},
	}
	router.SetGrammarService(grammarSvc)
	items := []repository.DailyRouteItem{{SourceKind: "grammar_theory_block", SourceID: "ch:empty-title", Title: "slug"}}
	router.enrichTheoryBlockTitles(httptest.NewRequest(http.MethodGet, "/", nil), userID, items)
	if items[0].Title != "slug" {
		t.Fatalf("empty theory title should not replace slug")
	}
}

func TestCourseAdminCoverage_AdminTrainingGenerateEmptyTranscription(t *testing.T) {
	router, db, adminUserID := setupAdminTrainingTest(t)
	wordRepo := repository.NewWordRepository(db.GetConnection(), router.logger)
	_ = wordRepo.SaveWordCard("notrans", "def", "")
	router.aiService = setupAdminAIService(t, `{"word_en":"notrans","transcription":"","senses":[{"pos":"n","display_word":"notrans","word_ru":"с","meaning_en":"m","example_en":"","example_ru":"","distractors_ru":[],"distractors_en":[],"hint":""}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/notrans/generate", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = adminCtx(req, adminUserID)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty transcription expected 400, got %d", w.Code)
	}
}
