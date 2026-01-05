package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

func TestRouter_getRateLimitPolicy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			RateLimitWindowMinutes: 5,
			RateLimitBurstMultiplier: 3,
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	policy := router.getRateLimitPolicy(100, 2)
	if policy.RequestsPerWindow != 100 {
		t.Errorf("Expected RequestsPerWindow 100, got %d", policy.RequestsPerWindow)
	}
	if policy.WindowDuration != 5*time.Minute {
		t.Errorf("Expected WindowDuration 5 minutes, got %v", policy.WindowDuration)
	}
	if policy.BurstSize != 200 {
		t.Errorf("Expected BurstSize 200, got %d", policy.BurstSize)
	}
}

func TestRouter_getRateLimitPolicy_Defaults(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			RateLimitWindowMinutes: 0,
			RateLimitBurstMultiplier: 0,
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	policy := router.getRateLimitPolicy(0, 0)
	if policy.RequestsPerWindow != 60 {
		t.Errorf("Expected default RequestsPerWindow 60, got %d", policy.RequestsPerWindow)
	}
	if policy.WindowDuration != 1*time.Minute {
		t.Errorf("Expected default WindowDuration 1 minute, got %v", policy.WindowDuration)
	}
	if policy.BurstSize < 60 {
		t.Errorf("Expected BurstSize >= 60, got %d", policy.BurstSize)
	}
}

func TestRouter_ServeHTTP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
