package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
)

func TestRequireFeature_FreeUserForbidden(t *testing.T) {
	router := &Router{
		config:   &config.Config{Speaking: config.SpeakingConfig{Enabled: true}},
		userRepo: &stubUserRepo{tier: models.TierFree},
		logger:   zap.NewNop(),
	}
	handler := router.RequireFeature("speaking")(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/learning/speaking/categories", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	rr := httptest.NewRecorder()
	handler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestRequireFeature_ProUserAllowed(t *testing.T) {
	router := &Router{
		config:   &config.Config{Speaking: config.SpeakingConfig{Enabled: true}},
		userRepo: &stubUserRepo{tier: models.TierPro},
		logger:   zap.NewNop(),
	}
	called := false
	handler := router.RequireFeature("speaking")(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/learning/speaking/categories", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	rr := httptest.NewRecorder()
	handler(rr, req)
	if rr.Code != http.StatusOK || !called {
		t.Fatalf("expected 200, got %d called=%v", rr.Code, called)
	}
}

type stubUserRepo struct {
	tier models.UserTier
}

func (s *stubUserRepo) GetUserByID(userID int64) (*models.User, error) {
	return &models.User{ID: userID, SubscriptionTier: s.tier}, nil
}
