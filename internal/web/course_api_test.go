package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupCourseAPITest(t *testing.T) (*Router, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{Learning: config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}}
	userRepo := repository.NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(99701)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return NewRouter(logger, cfg, db, nil, nil, nil, nil), user.ID
}

func TestCourseAPI_ListCurrentAndSelect(t *testing.T) {
	router, userID := setupCourseAPITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/courses", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w := httptest.NewRecorder()
	router.handleCourses(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("handleCourses status=%d body=%s", w.Code, w.Body.String())
	}
	var listResp struct {
		Courses []struct {
			Code      string `json:"code"`
			IsCurrent bool   `json:"is_current"`
		} `json:"courses"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode courses: %v", err)
	}
	if len(listResp.Courses) < 2 {
		t.Fatalf("courses = %+v, want at least seeded en/es", listResp.Courses)
	}

	selectReq := httptest.NewRequest(http.MethodPost, "/api/user/courses/select", strings.NewReader(`{"course_code":"en_ru"}`))
	selectReq = selectReq.WithContext(context.WithValue(selectReq.Context(), userIDKey, userID))
	w = httptest.NewRecorder()
	router.handleSelectCourse(w, selectReq)
	if w.Code != http.StatusOK {
		t.Fatalf("handleSelectCourse status=%d body=%s", w.Code, w.Body.String())
	}
	var selectResp struct {
		Course struct {
			Code      string `json:"code"`
			IsCurrent bool   `json:"is_current"`
		} `json:"course"`
		UserCourse struct {
			ID int64 `json:"id"`
		} `json:"user_course"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &selectResp); err != nil {
		t.Fatalf("decode selected course: %v", err)
	}
	if selectResp.Course.Code != "en_ru" || !selectResp.Course.IsCurrent || selectResp.UserCourse.ID == 0 {
		t.Fatalf("selected response = %+v", selectResp)
	}

	currentReq := httptest.NewRequest(http.MethodGet, "/api/user/courses/current", nil)
	currentReq = currentReq.WithContext(context.WithValue(currentReq.Context(), userIDKey, userID))
	w = httptest.NewRecorder()
	router.handleCurrentCourse(w, currentReq)
	if w.Code != http.StatusOK {
		t.Fatalf("handleCurrentCourse status=%d body=%s", w.Code, w.Body.String())
	}
	var currentResp struct {
		Course struct {
			Code string `json:"code"`
		} `json:"course"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &currentResp); err != nil {
		t.Fatalf("decode current course: %v", err)
	}
	if currentResp.Course.Code != "en_ru" {
		t.Fatalf("current course = %q, want en_ru", currentResp.Course.Code)
	}
}

func TestCourseAPI_SelectRejectsUnknownCourse(t *testing.T) {
	router, userID := setupCourseAPITest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/user/courses/select", strings.NewReader(`{"course_code":"xx_ru"}`))
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	w := httptest.NewRecorder()
	router.handleSelectCourse(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s, want 404", w.Code, w.Body.String())
	}
}
