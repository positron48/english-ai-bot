package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupWordSetsRouter(t *testing.T) (*Router, *database.DB, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	cleanup := func() {} // shared db, do not close

	return router, db, cleanup
}

func TestHandleLearningWordsCategories(t *testing.T) {
	router, db, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	catRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), router.logger)

	rootID, err := catRepo.CreateCategory(&models.WordSetCategory{Name: "root", IsPublished: true})
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}
	_, err = catRepo.CreateCategory(&models.WordSetCategory{Name: "child", ParentID: &rootID, IsPublished: true})
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/categories", nil)
	w := httptest.NewRecorder()
	router.handleLearningWordsCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string][]map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(payload["categories"]) != 1 {
		t.Fatalf("expected 1 root category")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/learning/words/categories?all=true", nil)
	w = httptest.NewRecorder()
	router.handleLearningWordsCategories(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(payload["categories"]) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(payload["categories"]))
	}
}

func TestHandleLearningWordsSets(t *testing.T) {
	router, db, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), router.logger)

	setID, err := wordSetRepo.CreateWordSet(&models.WordSet{Title: "Set A", IsPublished: true})
	if err != nil {
		t.Fatalf("CreateWordSet error: %v", err)
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	user, err := userRepo.GetOrCreateUser(123)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningWordsSets(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	sets, ok := payload["sets"].([]interface{})
	if !ok {
		t.Fatalf("expected sets payload")
	}
	if len(sets) != 1 {
		t.Fatalf("expected 1 set, got %d", len(sets))
	}

	_ = setID

	// Unauthorized
	req = httptest.NewRequest(http.MethodGet, "/api/learning/words/sets", nil)
	w = httptest.NewRecorder()
	router.handleLearningWordsSets(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleLearningWordsSets_WithCategory(t *testing.T) {
	router, db, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	catRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), router.logger)
	catID, err := catRepo.CreateCategory(&models.WordSetCategory{Name: "cat", IsPublished: true})
	if err != nil {
		t.Fatalf("CreateCategory error: %v", err)
	}

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), router.logger)
	_, err = wordSetRepo.CreateWordSet(&models.WordSet{Title: "Set A", CategoryID: &catID, IsPublished: true})
	if err != nil {
		t.Fatalf("CreateWordSet error: %v", err)
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	user, err := userRepo.GetOrCreateUser(123)
	if err != nil {
		t.Fatalf("GetOrCreateUser error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets?category_id="+strconv.FormatInt(catID, 10), nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningWordsSets(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	sets, ok := payload["sets"].([]interface{})
	if !ok {
		t.Fatalf("expected sets payload")
	}
	if len(sets) != 1 {
		t.Fatalf("expected 1 set, got %d", len(sets))
	}
}

func TestHandleLearningWordsSetDetail_FromWordSetsTest_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(router.db, router.logger)
	user, _ := userRepo.GetOrCreateUser(99999)

	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/sets/1", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningWordsSetDetail(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleLearningWordsSetDetail_FromWordSetsTest_Unauthorized(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/1", nil)
	w := httptest.NewRecorder()
	router.handleLearningWordsSetDetail(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandleLearningWordsSetDetail_FromWordSetsTest_InvalidSetID(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(router.db, router.logger)
	user, _ := userRepo.GetOrCreateUser(99998)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/notanumber", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningWordsSetDetail(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid set ID, got %d", w.Code)
	}
}

func TestHandleLearningWordsSetDetail_FromWordSetsTest_NotFound(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(router.db, router.logger)
	user, _ := userRepo.GetOrCreateUser(99997)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/999999", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningWordsSetDetail(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent set, got %d", w.Code)
	}
}

func TestHandleLearningWordsSetDetail_FromWordSetsTest_Success(t *testing.T) {
	router, db, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), router.logger)
	setID, err := wordSetRepo.CreateWordSet(&models.WordSet{Title: "Detail Set", IsPublished: true})
	if err != nil {
		t.Fatalf("CreateWordSet: %v", err)
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	user, _ := userRepo.GetOrCreateUser(99996)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/"+strconv.FormatInt(setID, 10), nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningWordsSetDetail(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Response has word_set (object) and words (array)
	if _, ok := payload["word_set"]; !ok {
		t.Error("expected word_set in response")
	}
	if _, ok := payload["words"]; !ok {
		t.Error("expected words in response")
	}
}

func TestHandleLearningWordsSets_FromWordSetsTest_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(router.db, router.logger)
	user, _ := userRepo.GetOrCreateUser(99995)

	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/sets", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningWordsSets(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleLearningWordsSets_FromWordSetsTest_LimitOffset(t *testing.T) {
	router, db, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), router.logger)
	for i := 0; i < 3; i++ {
		_, err := wordSetRepo.CreateWordSet(&models.WordSet{Title: "Set " + strconv.Itoa(i), IsPublished: true})
		if err != nil {
			t.Fatalf("CreateWordSet: %v", err)
		}
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), router.logger)
	user, _ := userRepo.GetOrCreateUser(99994)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets?limit=2&offset=1", nil)
	req = setUserIDInContext(req, user.ID)
	w := httptest.NewRecorder()
	router.handleLearningWordsSets(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	sets, _ := payload["sets"].([]interface{})
	if len(sets) != 2 {
		t.Errorf("expected 2 sets with limit=2 offset=1, got %d", len(sets))
	}
}

// TestRouter_GetWordSetService ensures getWordSetService returns a non-nil service (covers getWordSetService with aiService nil).
func TestRouter_GetWordSetService(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	svc := router.getWordSetService()
	if svc == nil {
		t.Error("getWordSetService returned nil")
	}
}

// TestGetWordSetService_WithAIService covers the branch where r.aiService is set and type-asserts to *ai.Service.
func TestGetWordSetService_WithAIService(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	// Use real ai.Service so type assertion in getWordSetService succeeds
	aiSvc := ai.NewService("http://localhost", "test-model", "test-key", "test-prompt", zap.NewNop())
	router.aiService = aiSvc

	svc := router.getWordSetService()
	if svc == nil {
		t.Error("getWordSetService returned nil when aiService was set")
	}
}

// TestHandleLearningWordsCategories_MethodNotAllowed_WordSets covers POST -> 405.
func TestHandleLearningWordsCategories_MethodNotAllowed_WordSets(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/categories", nil)
	w := httptest.NewRecorder()
	router.handleLearningWordsCategories(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

// TestHandleLearningWordsCategories_WithParentID covers filtering by parent_id (children of root).
func TestHandleLearningWordsCategories_WithParentID(t *testing.T) {
	router, db, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	catRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), router.logger)
	rootID, err := catRepo.CreateCategory(&models.WordSetCategory{Name: "root2", IsPublished: true})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	_, err = catRepo.CreateCategory(&models.WordSetCategory{Name: "child2", ParentID: &rootID, IsPublished: true, SortOrder: 1})
	if err != nil {
		t.Fatalf("CreateCategory child: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/categories?parent_id="+strconv.FormatInt(rootID, 10), nil)
	w := httptest.NewRecorder()
	router.handleLearningWordsCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string][]map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload["categories"]) != 1 {
		t.Errorf("expected 1 child category, got %d", len(payload["categories"]))
	}
}

// TestHandleLearningWordsSets_InvalidLimitAndOffset covers limit/offset parsing (clamp and invalid).
func TestHandleLearningWordsSets_InvalidLimitAndOffset(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(router.db, router.logger)
	user, _ := userRepo.GetOrCreateUser(88881)

	tests := []struct {
		name     string
		query    string
		wantCode int
	}{
		{"limit over 200 clamped", "limit=300&offset=0", http.StatusOK},
		{"offset negative ignored", "offset=-1", http.StatusOK},
		{"limit zero ignored", "limit=0", http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets?"+tt.query, nil)
			req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
			w := httptest.NewRecorder()
			router.handleLearningWordsSets(w, req)
			if w.Code != tt.wantCode {
				t.Errorf("got status %d, want %d", w.Code, tt.wantCode)
			}
		})
	}
}

// TestHandleLearningWordsSets_CategoryIDNonNumeric covers category_id present but invalid (ListWordSets path).
func TestHandleLearningWordsSets_CategoryIDNonNumeric(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(router.db, router.logger)
	user, _ := userRepo.GetOrCreateUser(88882)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets?category_id=abc", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleLearningWordsSets(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (list with nil categoryID), got %d", w.Code)
	}
}

// TestHandleLearningWordsSetDetail_EmptyPath covers path that yields empty set ID (400).
func TestHandleLearningWordsSetDetail_EmptyPath(t *testing.T) {
	router, _, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	userRepo := repository.NewUserRepository(router.db, router.logger)
	user, _ := userRepo.GetOrCreateUser(88883)

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/", nil)
	req.URL.Path = "/api/learning/words/sets/"
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()
	router.handleLearningWordsSetDetail(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty path, got %d", w.Code)
	}
}

// TestHandleLearningWordsCategories_SortOrderSwap covers the sort swap branch (two root categories with different sort_order).
func TestHandleLearningWordsCategories_SortOrderSwap(t *testing.T) {
	router, db, cleanup := setupWordSetsRouter(t)
	defer cleanup()

	catRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), router.logger)
	_, err := catRepo.CreateCategory(&models.WordSetCategory{Name: "second", IsPublished: true, SortOrder: 2})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	_, err = catRepo.CreateCategory(&models.WordSetCategory{Name: "first", IsPublished: true, SortOrder: 1})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/categories", nil)
	w := httptest.NewRecorder()
	router.handleLearningWordsCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string][]map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload["categories"]) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(payload["categories"]))
	}
	if payload["categories"][0]["name"] != "first" {
		t.Errorf("expected first category name 'first' after sort, got %v", payload["categories"][0]["name"])
	}
}

// setupWordSetsRouterWithBadDB creates a router with a closed DB for error-path coverage.
func setupWordSetsRouterWithBadDB(t *testing.T) *Router {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	return NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)
}

// TestHandleLearningWordsCategories_GetPublishedCategoriesError covers 500 when GetPublishedCategories fails.
func TestHandleLearningWordsCategories_GetPublishedCategoriesError(t *testing.T) {
	router := setupWordSetsRouterWithBadDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/categories", nil)
	w := httptest.NewRecorder()
	router.handleLearningWordsCategories(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetPublishedCategories fails, got %d", w.Code)
	}
}

// TestHandleLearningWordsSets_QueryFails covers 500 when listing sets without category fails (raw query).
func TestHandleLearningWordsSets_QueryFails(t *testing.T) {
	router := setupWordSetsRouterWithBadDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	w := httptest.NewRecorder()
	router.handleLearningWordsSets(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when query fails, got %d", w.Code)
	}
}

// TestHandleLearningWordsSets_ListWordSetsFails covers 500 when ListWordSets fails (category_id set).
func TestHandleLearningWordsSets_ListWordSetsFails(t *testing.T) {
	router := setupWordSetsRouterWithBadDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets?category_id=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	w := httptest.NewRecorder()
	router.handleLearningWordsSets(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when ListWordSets fails, got %d", w.Code)
	}
}

// TestHandleLearningWordsSetDetail_GetWordSetError covers 500 when GetWordSet fails.
func TestHandleLearningWordsSetDetail_GetWordSetError(t *testing.T) {
	router := setupWordSetsRouterWithBadDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/1", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	w := httptest.NewRecorder()
	router.handleLearningWordsSetDetail(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetWordSet fails, got %d", w.Code)
	}
}

// TestHandleLearningWordsSetStudy_GetWordSetWordsFails covers 500 when GetWordSetWords fails in study.
func TestHandleLearningWordsSetStudy_GetWordSetWordsFails(t *testing.T) {
	router := setupWordSetsRouterWithBadDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/1/study?word_card_id=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	w := httptest.NewRecorder()
	router.handleLearningWordsSetStudy(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetWordSetWords fails, got %d", w.Code)
	}
}

// TestHandleLearningWordsSetStudyLearn_GetWordSetWordsFails covers 500 when GetWordSetWords fails in learn.
func TestHandleLearningWordsSetStudyLearn_GetWordSetWordsFails(t *testing.T) {
	router := setupWordSetsRouterWithBadDB(t)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/sets/1/study/learn", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	req.Body = io.NopCloser(strings.NewReader(`{"word_card_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleLearningWordsSetStudyLearn(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetWordSetWords fails in learn, got %d", w.Code)
	}
}

// TestHandleLearningWordsSetStudyKnow_GetWordSetWordsFails covers 500 when GetWordSetWords fails in know.
func TestHandleLearningWordsSetStudyKnow_GetWordSetWordsFails(t *testing.T) {
	router := setupWordSetsRouterWithBadDB(t)
	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/sets/1/study/know", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	req.Body = io.NopCloser(strings.NewReader(`{"word_card_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleLearningWordsSetStudyKnow(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetWordSetWords fails in know, got %d", w.Code)
	}
}
