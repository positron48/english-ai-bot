package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

func TestIsAPIEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Dashboard endpoint",
			path:     "/api/dashboard",
			expected: true,
		},
		{
			name:     "Vocab endpoint",
			path:     "/api/vocab",
			expected: true,
		},
		{
			name:     "Training endpoint",
			path:     "/api/training/start",
			expected: true,
		},
		{
			name:     "Chat endpoint",
			path:     "/api/chat",
			expected: true,
		},
		{
			name:     "Admin endpoint",
			path:     "/api/admin",
			expected: true,
		},
		{
			name:     "Any API path",
			path:     "/api/some-page",
			expected: true,
		},
		{
			name:     "Root app path (SPA route)",
			path:     "/app",
			expected: false,
		},
		{
			name:     "App dashboard (SPA route)",
			path:     "/app/dashboard",
			expected: false,
		},
		{
			name:     "Asset path under /app",
			path:     "/app/assets/main.js",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAPIEndpoint(tt.path)
			if result != tt.expected {
				t.Errorf("isAPIEndpoint(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestHasStaticAssetExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "JavaScript file",
			path:     "/app/assets/main.js",
			expected: true,
		},
		{
			name:     "CSS file",
			path:     "/app/assets/style.css",
			expected: true,
		},
		{
			name:     "Image file",
			path:     "/app/favicon.svg",
			expected: true,
		},
		{
			name:     "SPA route param contains dots (should NOT be treated as file)",
			path:     "/app/learning/grammar/en.grammar.first_sentences_be_as",
			expected: false,
		},
		{
			name:     "No extension",
			path:     "/app/dashboard",
			expected: false,
		},
		{
			name:     "Root path",
			path:     "/app/",
			expected: false,
		},
		{
			name:     "Single dot",
			path:     "/api/file.",
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasStaticAssetExtension(tt.path)
			if result != tt.expected {
				t.Errorf("hasStaticAssetExtension(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestSetupDevProxy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			ViteDevServerURL: "http://localhost:5173",
		},
	}

	// Create router manually to avoid route conflicts
	// setupDevProxy is called internally by setupWebappRoutes when files aren't embedded
	// We just verify it doesn't panic with valid config
	router := &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		config: cfg,
	}
	
	router.setupDevProxy()

	// Test that the proxy handler is set up
	// We can't easily test the redirect without making actual HTTP requests,
	// but we can verify the function doesn't panic and sets up handlers
	_ = router
}

func TestSetupDevProxy_InvalidURL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			ViteDevServerURL: "://invalid-url",
		},
	}

	router := &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		config: cfg,
	}
	
	// Should handle invalid URL gracefully (logs error but doesn't panic)
	router.setupDevProxy()
	_ = router
}

func TestSetupDevProxy_EmptyURL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			ViteDevServerURL: "",
		},
	}

	router := &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		config: cfg,
	}
	
	// Should use default URL
	router.setupDevProxy()
	_ = router
}

func TestHandleNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	w := httptest.NewRecorder()

	router.handleNotFound(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}
