package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewRateLimitMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 10,
		WindowDuration:    1 * time.Minute,
		BurstSize:         10,
	}

	middleware := NewRateLimitMiddleware(rl, logger, policy, KeyFuncIP)
	_ = middleware // Verify middleware is created
}

func TestRateLimitMiddleware_Wrap(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 5,
		WindowDuration:    1 * time.Minute,
		BurstSize:         5,
	}

	middleware := NewRateLimitMiddleware(rl, logger, policy, KeyFuncIP)

	handlerCalled := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped(w, req)

	if !handlerCalled {
		t.Error("Handler should be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_RateLimit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 2,
		WindowDuration:    1 * time.Minute,
		BurstSize:         2,
	}

	middleware := NewRateLimitMiddleware(rl, logger, policy, KeyFuncIP)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	// Make requests up to limit
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		wrapped(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed, got status %d", i+1, w.Code)
		}
	}

	// Next request should be rate limited
	w := httptest.NewRecorder()
	wrapped(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}
