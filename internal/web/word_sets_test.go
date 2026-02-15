package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

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
