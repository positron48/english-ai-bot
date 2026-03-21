package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupAdminTTSRouter(t *testing.T) (*Router, *repository.UserRepository, *repository.UserAccessCategoryRepository) {
	t.Helper()
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	userRepo := repository.NewUserRepository(db, logger)
	accessRepo := repository.NewUserAccessCategoryRepository(db, logger)
	wordRepo := repository.NewWordRepository(db, logger)
	pron := service.NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
	}, config.DefaultLearningConfig(), wordRepo, logger)

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.pronunciationService = pron
	return router, userRepo, accessRepo
}

func grantCategory(t *testing.T, accessRepo *repository.UserAccessCategoryRepository, userID int64, perms ...string) []int64 {
	t.Helper()
	catID, err := accessRepo.CreateCategory(&models.UserAccessCategory{Name: "tts-test-" + perms[0]})
	if err != nil {
		t.Fatalf("CreateCategory() error = %v", err)
	}
	if err := accessRepo.SetCategoryPermissions(catID, perms); err != nil {
		t.Fatalf("SetCategoryPermissions() error = %v", err)
	}
	if err := accessRepo.SetUserCategories(userID, []int64{catID}); err != nil {
		t.Fatalf("SetUserCategories() error = %v", err)
	}
	return []int64{catID}
}

func TestHandleAdminTTS_GetContract(t *testing.T) {
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(1001)
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsReadAll))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/tts/spy", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	ctx = context.WithValue(ctx, userCategoriesKey, categories)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	required := []string{"state", "attempt_count", "max_attempts", "last_error_code", "last_error_message", "audio_url", "updated_at"}
	for _, key := range required {
		if _, ok := resp[key]; !ok {
			t.Fatalf("missing key %q in response", key)
		}
	}
}

func TestHandleAdminTTS_RegenerateRequiresEditPermission(t *testing.T) {
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(1002)
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsReadAll))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/tts/spy/regenerate", nil)
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	ctx = context.WithValue(ctx, userCategoriesKey, categories)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHandleAdminTTS_ServiceUnavailable(t *testing.T) {
	logger := zap.NewNop()
	db := testutil.SetupTestDB(t)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	userRepo := repository.NewUserRepository(db, logger)
	accessRepo := repository.NewUserAccessCategoryRepository(db, logger)
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.accessCategoryRepo = accessRepo
	// pronunciationService is nil

	req := httptest.NewRequest(http.MethodGet, "/api/admin/tts/word", nil)
	ctx := context.WithValue(req.Context(), userIDKey, int64(1))
	ctx = context.WithValue(ctx, userCategoriesKey, []int64{})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTTS_WordRequired(t *testing.T) {
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(1003)
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsReadAll))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/tts/", nil)
	req.URL.Path = "/api/admin/tts/"
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	ctx = context.WithValue(ctx, userCategoriesKey, categories)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTTS_GetWithActionMethodNotAllowed(t *testing.T) {
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(1004)
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsReadAll))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/tts/word/recheck", nil)
	req.URL.Path = "/api/admin/tts/word/recheck"
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	ctx = context.WithValue(ctx, userCategoriesKey, categories)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminTTS_PostWrongActionMethodNotAllowed(t *testing.T) {
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(1005)
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsEditAll))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/tts/word/unknown", nil)
	req.URL.Path = "/api/admin/tts/word/unknown"
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	ctx = context.WithValue(ctx, userCategoriesKey, categories)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminTTS_PostRegenerateSuccess(t *testing.T) {
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(1006)
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsEditAll))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/tts/hello/regenerate", nil)
	req.URL.Path = "/api/admin/tts/hello/regenerate"
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	ctx = context.WithValue(ctx, userCategoriesKey, categories)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["state"]; !ok {
		t.Fatalf("expected state in response")
	}
}

func TestHandleAdminTTS_PostRecheckSuccess(t *testing.T) {
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(1007)
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsEditAll))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/tts/hello/recheck", nil)
	req.URL.Path = "/api/admin/tts/hello/recheck"
	ctx := context.WithValue(req.Context(), userIDKey, user.ID)
	ctx = context.WithValue(ctx, userCategoriesKey, categories)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["state"]; !ok {
		t.Fatalf("expected state in response")
	}
}
