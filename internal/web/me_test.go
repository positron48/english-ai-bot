package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupMeTest(t *testing.T) (*Router, int64) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-me-secret"},
	}
	router := NewRouter(logger, cfg, conn, nil, nil, nil, nil)
	userRepo := repository.NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(900101)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := userRepo.UpdateSubscriptionTier(user.ID, models.TierPro); err != nil {
		t.Fatalf("set tier: %v", err)
	}
	router.SetDependencies(userRepo, nil, nil, nil, "bot-token")
	return router, user.ID
}

func TestMe_OK(t *testing.T) {
	router, userID := setupMeTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleMe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		ID               int64  `json:"id"`
		TelegramID       int64  `json:"telegram_id"`
		SubscriptionTier string `json:"subscription_tier"`
		Features         map[string]bool `json:"features"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ID != userID || body.TelegramID != 900101 {
		t.Fatalf("unexpected profile: %+v", body)
	}
	if body.SubscriptionTier != string(models.TierPro) {
		t.Fatalf("tier = %q, want pro", body.SubscriptionTier)
	}
	if !body.Features["speaking"] {
		t.Fatal("pro user should have speaking feature")
	}
}

func TestMe_Errors(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		noUser     bool
		setup      func(*Router)
		wantCode   int
	}{
		{
			name:     "unauthorized",
			method:   http.MethodGet,
			noUser:   true,
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "method not allowed",
			method:   http.MethodPost,
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name: "user repo missing",
			setup: func(r *Router) {
				r.userRepo = nil
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "user repo wrong type",
			setup: func(r *Router) {
				r.userRepo = "not-a-repo"
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router, userID := setupMeTest(t)
			if tc.setup != nil {
				tc.setup(router)
			}
			method := tc.method
			if method == "" {
				method = http.MethodGet
			}
			req := httptest.NewRequest(method, "/api/me", nil)
			if !tc.noUser {
				req = setUserIDInContext(req, userID)
			}
			w := httptest.NewRecorder()
			router.handleMe(w, req)
			if w.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d: %s", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}