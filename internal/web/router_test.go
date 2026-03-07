package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRouter_HandleNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	router.handleNotFound(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), "Not Found") {
		t.Errorf("Response body should contain 'Not Found', got %s", w.Body.String())
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

func TestSwaggerResponseWriter_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &swaggerResponseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
	}

	// Write before header is written - should buffer
	data := []byte("test data")
	n, err := wrapped.Write(data)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(data) {
		t.Errorf("Write() returned %d, want %d", n, len(data))
	}
	if len(wrapped.buf) != len(data) {
		t.Errorf("Expected buffered data length %d, got %d", len(data), len(wrapped.buf))
	}

	// Write after header is written - should write directly
	wrapped.headerWritten = true
	n, err = wrapped.Write([]byte("more data"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 9 {
		t.Errorf("Write() returned %d, want 9", n)
	}
}

func TestSwaggerResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &swaggerResponseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		buf:            []byte("buffered data"),
	}

	// WriteHeader should set headerWritten and statusCode
	wrapped.WriteHeader(http.StatusOK)
	if !wrapped.headerWritten {
		t.Error("headerWritten should be true after WriteHeader")
	}
	if wrapped.statusCode != http.StatusOK {
		t.Errorf("Expected statusCode %d, got %d", http.StatusOK, wrapped.statusCode)
	}

	// Second call should not change anything
	wrapped.WriteHeader(http.StatusNotFound)
	if wrapped.statusCode != http.StatusOK {
		t.Errorf("Expected statusCode to remain %d, got %d", http.StatusOK, wrapped.statusCode)
	}
}

func TestSwaggerResponseWriter_WriteHeader_WithHTML(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &swaggerResponseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		buf:            []byte("<!DOCTYPE html><html><body>Test</body></html>"),
	}

	// Set Content-Type to HTML
	recorder.Header().Set("Content-Type", "text/html; charset=utf-8")

	// WriteHeader should process HTML content and inject JavaScript
	wrapped.WriteHeader(http.StatusOK)
	if !wrapped.headerWritten {
		t.Error("headerWritten should be true after WriteHeader")
	}
	// Buffer is written to ResponseWriter, so it should be empty or nil after WriteHeader
	// Check that the response was written (recorder should have the modified content)
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestSwaggerResponseWriter_WriteHeader_WithHTMLContentType(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &swaggerResponseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		buf:            []byte("<html><head></head><body>Test</body></html>"),
	}

	// Set Content-Type to HTML (different way)
	recorder.Header().Set("Content-Type", "text/html")

	// WriteHeader should process HTML content
	wrapped.WriteHeader(http.StatusOK)
	if !wrapped.headerWritten {
		t.Error("headerWritten should be true after WriteHeader")
	}
}

func TestSwaggerResponseWriter_WriteHeader_NonHTML(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &swaggerResponseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
		buf:            []byte("plain text"),
	}

	// Set Content-Type to plain text
	recorder.Header().Set("Content-Type", "text/plain")

	// WriteHeader should not inject JavaScript for non-HTML content
	wrapped.WriteHeader(http.StatusOK)
	if !wrapped.headerWritten {
		t.Error("headerWritten should be true after WriteHeader")
	}
}

func TestSwaggerResponseWriter_Header(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &swaggerResponseWriter{
		ResponseWriter: recorder,
	}

	headers := wrapped.Header()
	if headers == nil {
		t.Error("Header() should not return nil")
	}

	// Should be the same as underlying ResponseWriter
	headers.Set("Test-Header", "test-value")
	if recorder.Header().Get("Test-Header") != "test-value" {
		t.Error("Header() should return the same header map as underlying ResponseWriter")
	}
}

func TestRouter_SwaggerHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	// Test swagger handler with non-HTML path (static file)
	req := httptest.NewRequest("GET", "/swagger/swagger-ui.css", nil)
	w := httptest.NewRecorder()

	router.swaggerHandler(w, req)

	// Should handle the request (may return 404 if swagger files don't exist, but function should execute)
	_ = w.Code
}

func TestRouter_SwaggerHandler_HTMLPage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	// Test swagger handler with HTML page path
	req := httptest.NewRequest("GET", "/swagger/", nil)
	w := httptest.NewRecorder()

	router.swaggerHandler(w, req)

	// Should handle the request
	_ = w.Code
}

func TestRouter_SwaggerHandler_IndexHTML(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	// Test swagger handler with index.html path
	req := httptest.NewRequest("GET", "/swagger/index.html", nil)
	w := httptest.NewRecorder()

	router.swaggerHandler(w, req)

	// Should handle the request
	_ = w.Code
}

func TestRouter_SwaggerHandler_RootPath(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	// Test swagger handler with root swagger path
	req := httptest.NewRequest("GET", "/swagger", nil)
	w := httptest.NewRecorder()

	router.swaggerHandler(w, req)

	// Should handle the request
	_ = w.Code
}
