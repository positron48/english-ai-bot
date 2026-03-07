package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

// errWriter implements http.ResponseWriter and fails on Write to trigger encode error path in Wrap.
type errWriter struct {
	http.ResponseWriter
}

func (w *errWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

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

func TestRateLimitMiddleware_Wrap_OPTIONSBypass(t *testing.T) {
	logger := zap.NewNop()
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 1,
		WindowDuration:    1 * time.Minute,
		BurstSize:         1,
	}

	middleware := NewRateLimitMiddleware(rl, logger, policy, KeyFuncIP)

	handlerCalled := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.RemoteAddr = "127.0.0.1:99999"
	w := httptest.NewRecorder()
	wrapped(w, req)

	if !handlerCalled {
		t.Error("OPTIONS request should bypass rate limit and call handler")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_Wrap_EmptyKeyAllowsRequest(t *testing.T) {
	logger := zap.NewNop()
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 1,
		WindowDuration:    1 * time.Minute,
		BurstSize:         1,
	}

	// Key function that returns empty string
	emptyKeyFn := func(r *http.Request) string { return "" }
	middleware := NewRateLimitMiddleware(rl, logger, policy, emptyKeyFn)

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
		t.Error("Request with empty key should be allowed and call handler")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_Wrap_EncodeError(t *testing.T) {
	logger := zap.NewNop()
	rl := NewRateLimiter(1*time.Minute, 1*time.Hour)
	defer rl.Stop()

	policy := RateLimitPolicy{
		RequestsPerWindow: 1,
		WindowDuration:     1 * time.Minute,
		BurstSize:         1,
	}

	middleware := NewRateLimitMiddleware(rl, logger, policy, KeyFuncIP)
	handler := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }
	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	// Exhaust the limit so next request gets 429
	rec := httptest.NewRecorder()
	wrapped(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first request should succeed, got %d", rec.Code)
	}

	// Second request is rate-limited; use a writer that fails on Write so json.Encode fails
	w := httptest.NewRecorder()
	wrapped(&errWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}
