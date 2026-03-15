package web

// Tests to cover error paths in admin_tts.go not covered by admin_tts_test.go.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/service"
)

// errPronunciationService is a mock that always returns errors from GetStatus/ForceRegenerate/Recheck.
type errPronunciationService struct {
	mockPronunciationService
	statusErr     error
	regenerateErr error
	recheckErr    error
}

func (m *errPronunciationService) GetStatus(word string) (service.TTSStatusResult, error) {
	return service.TTSStatusResult{}, m.statusErr
}

func (m *errPronunciationService) ForceRegenerate(word string) (service.TTSStatusResult, error) {
	return service.TTSStatusResult{}, m.regenerateErr
}

func (m *errPronunciationService) Recheck(word string) (service.TTSStatusResult, error) {
	return service.TTSStatusResult{}, m.recheckErr
}

// makeTTSReqWithCategories creates a request with user ID and categories in context.
func makeTTSReqWithCategories(method, path string, userID int64, categories []int64) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.URL.Path = path
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	ctx = context.WithValue(ctx, userCategoriesKey, categories)
	return req.WithContext(ctx)
}

// TestHandleAdminTTS_GET_Forbidden covers lines 32-35:
// HasPermission returns false for GET → 403 Forbidden.
func TestHandleAdminTTS_GET_Forbidden(t *testing.T) {
	router, userRepo, _ := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(9902)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	// User has no permissions (empty categories)
	req := makeTTSReqWithCategories(http.MethodGet, "/api/admin/tts/hello", user.ID, []int64{})
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTTS_GET_GetStatusError covers lines 41-44:
// GetStatus returns an error → 400 Bad Request.
func TestHandleAdminTTS_GET_GetStatusError(t *testing.T) {
	svc := &errPronunciationService{
		mockPronunciationService: mockPronunciationService{enabled: true},
		statusErr:                errors.New("status lookup failed"),
	}
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(9903)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsReadAll))
	router.pronunciationService = svc

	req := makeTTSReqWithCategories(http.MethodGet, "/api/admin/tts/hello", user.ID, categories)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTTS_POST_ForceRegenerateError covers lines 55-58:
// ForceRegenerate returns an error → 400 Bad Request.
func TestHandleAdminTTS_POST_ForceRegenerateError(t *testing.T) {
	svc := &errPronunciationService{
		mockPronunciationService: mockPronunciationService{enabled: true},
		regenerateErr:            errors.New("regenerate failed"),
	}
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(9904)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsEditAll))
	router.pronunciationService = svc

	req := makeTTSReqWithCategories(http.MethodPost, "/api/admin/tts/hello/regenerate", user.ID, categories)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTTS_POST_DefaultAction covers lines 69-72 (inner switch default):
// POST with an unknown action → "Method not allowed" 405.
func TestHandleAdminTTS_POST_DefaultAction(t *testing.T) {
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(9906)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsEditAll))

	req := makeTTSReqWithCategories(http.MethodPost, "/api/admin/tts/hello/unknown_action", user.ID, categories)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTTS_DefaultMethod covers lines 73-75 (outer switch default):
// Non-GET, non-POST method → "Method not allowed" 405.
func TestHandleAdminTTS_DefaultMethod(t *testing.T) {
	svc := &mockPronunciationService{enabled: true}
	router := &Router{logger: zap.NewNop(), pronunciationService: svc}

	req := makeTTSReqWithCategories(http.MethodDelete, "/api/admin/tts/hello", 1, []int64{})
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTTS_POST_RecheckError covers lines 63-66:
// Recheck returns an error → 400 Bad Request.
func TestHandleAdminTTS_POST_RecheckError(t *testing.T) {
	svc := &errPronunciationService{
		mockPronunciationService: mockPronunciationService{enabled: true},
		recheckErr:               errors.New("recheck failed"),
	}
	router, userRepo, accessRepo := setupAdminTTSRouter(t)
	user, err := userRepo.GetOrCreateUser(9905)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	categories := grantCategory(t, accessRepo, user.ID, string(PermissionWordsEditAll))
	router.pronunciationService = svc

	req := makeTTSReqWithCategories(http.MethodPost, "/api/admin/tts/hello/recheck", user.ID, categories)
	w := httptest.NewRecorder()

	router.handleAdminTTS(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
