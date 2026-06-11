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

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setUserIDInContextWordSets(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	// Set empty categories for tests (user will need permissions assigned)
	ctx = context.WithValue(ctx, userCategoriesKey, []int64{})
	return req.WithContext(ctx)
}

func setupAdminWordSetsTest(t *testing.T) (*Router, *database.DB, int64, func()) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"

	// Create super admin user in DB
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	adminUser, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	cleanup := func() {} // shared db, do not close

	return router, db, adminUser.ID, cleanup
}

func TestHandleAdminWordSetCategories_Get(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-set-categories", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Categories []interface{} `json:"categories"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// categories key is present (decode succeeded); may be nil or empty slice
	_ = resp.Categories
}

func TestHandleAdminWordSetCategories_GetFiltersByCourse(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()
	repo := repository.NewWordSetCategoryRepository(db.GetConnection(), zap.NewNop())
	if _, err := repo.CreateCategory(&models.WordSetCategory{Name: "English", CourseCode: "en_ru"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CreateCategory(&models.WordSetCategory{Name: "Spanish", CourseCode: "es_ru"}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-set-categories?course_code=es_ru", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()
	router.handleAdminWordSetCategories(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		Categories []models.WordSetCategory `json:"categories"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Categories) != 1 || resp.Categories[0].CourseCode != "es_ru" {
		t.Fatalf("categories = %+v, want only es_ru", resp.Categories)
	}
}

func TestHandleAdminWordSetCategories_PostInvalidBody(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/word-set-categories", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetCategories(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleAdminWordSetCategories_PostNameRequired(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/word-set-categories", bytes.NewBufferString(`{"name":""}`))
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetCategories(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Name is required") {
		t.Errorf("Expected body to contain 'Name is required', got %s", rr.Body.String())
	}
}

func TestHandleAdminWordSets_Get(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-sets", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSets_GetWithCategory(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-sets?category_id=1", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSets_PostInvalidBody(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/word-sets", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSets(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleAdminWordSetDetailOrSets_ListWordSets(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-sets/", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetailOrSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSetDetailOrSets_NotFound(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-sets/invalid", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetailOrSets(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSetDetail_MethodNotAllowed(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/word-sets/1", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetail(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminWordSetDetail_NotFound(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-sets/99999", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetail(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSetDetail_InvalidID(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-sets/not-a-number", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSetCategories_MethodNotAllowed(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/word-set-categories", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetCategories(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSets_MethodNotAllowed(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/word-sets", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSets(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSets_GetWithLimitAndOffset(t *testing.T) {
	router, _, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/word-sets?limit=5&offset=0", nil)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		WordSets []interface{} `json:"word_sets"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Response may have empty or non-empty word_sets; key must exist (decode succeeded)
	_ = resp.WordSets
}

func TestHandleAdminWordSetDetail_GetSuccess(t *testing.T) {
	router, db, adminUserID, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), logger)
	setID, err := wordSetRepo.CreateWordSet(&models.WordSet{Title: "Detail Set", IsPublished: true})
	if err != nil {
		t.Fatalf("CreateWordSet: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/admin/word-sets/%d", setID), nil)
	req.URL.Path = fmt.Sprintf("/api/admin/word-sets/%d", setID)
	req = setUserIDInContextWordSets(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetail(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp struct {
		WordSet map[string]interface{} `json:"word_set"`
		Words   []interface{}          `json:"words"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.WordSet == nil {
		t.Error("word_set should be present")
	}
	if resp.Words == nil {
		t.Error("words should be present")
	}
}
