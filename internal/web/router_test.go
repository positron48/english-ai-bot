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

func TestRouter_corsMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	tests := []struct {
		name           string
		origin         string
		method         string
		expectedOrigin string
		expectedStatus int
	}{
		{
			name:           "Localhost origin",
			origin:         "http://localhost:8184",
			method:         "GET",
			expectedOrigin: "http://localhost:8184",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "127.0.0.1 origin",
			origin:         "http://127.0.0.1:8080",
			method:         "GET",
			expectedOrigin: "http://127.0.0.1:8080",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Telegram origin",
			origin:         "https://web.telegram.org",
			method:         "GET",
			expectedOrigin: "https://web.telegram.org",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OPTIONS preflight",
			origin:         "http://localhost:8184",
			method:         "OPTIONS",
			expectedOrigin: "http://localhost:8184",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "No origin header",
			origin:         "",
			method:         "GET",
			expectedOrigin: "*",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HTTPS origin",
			origin:         "https://example.com",
			method:         "GET",
			expectedOrigin: "https://example.com",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			handler := func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}

			wrapped := router.corsMiddleware(handler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			wrapped(w, req)

			if tt.method == "OPTIONS" {
				if w.Code != http.StatusNoContent {
					t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusNoContent, w.Code)
				}
				if handlerCalled {
					t.Error("Handler should not be called for OPTIONS request")
				}
			} else {
				if !handlerCalled {
					t.Error("Handler should be called")
				}
				if w.Code != tt.expectedStatus {
					t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				}
			}

			gotOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if gotOrigin != tt.expectedOrigin {
				t.Errorf("Expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, gotOrigin)
			}

			gotMethods := w.Header().Get("Access-Control-Allow-Methods")
			if gotMethods == "" && tt.method == "OPTIONS" {
				t.Error("Access-Control-Allow-Methods should be set for OPTIONS request")
			}
		})
	}
}
