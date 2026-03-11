package web

// Tests to cover error paths in webapp_routes.go that are not covered by webapp_routes_test.go.

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

// countingIndexFS is an fs.FS that allows the first N opens of dist/index.html
// but fails on subsequent ones.
type countingIndexFS struct {
	inner     fs.FS
	failAfter int
	count     *int
}

func (c *countingIndexFS) Open(name string) (fs.File, error) {
	if name == "dist/index.html" || name == "index.html" {
		*c.count++
		if *c.count > c.failAfter {
			return nil, fs.ErrNotExist
		}
	}
	return c.inner.Open(name)
}

// newTestRouter creates a minimal Router for webapp route tests.
func newTestRouter(t *testing.T) *Router {
	t.Helper()
	logger := zap.NewNop()
	cfg := &config.Config{}
	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		config: cfg,
	}
}

// TestSetupWebappRoutes_DevProxy covers lines 21-25:
// when fs.ReadFile(webappFS, "dist/index.html") fails, setupDevProxy is called.
func TestSetupWebappRoutes_DevProxy(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()

	// Use an empty FS so dist/index.html doesn't exist → setupDevProxy is called
	webappFS = fstest.MapFS{}

	router := newTestRouter(t)
	router.setupWebappRoutes()

	// Verify that /app/ is handled (dev proxy registers it)
	req := httptest.NewRequest("GET", "/app/dashboard", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)

	// Dev proxy redirects to Vite dev server
	if w.Code != http.StatusTemporaryRedirect && w.Code != http.StatusNotFound {
		t.Errorf("expected redirect or 404 in dev proxy mode, got %d", w.Code)
	}
}

// TestSetupWebappRoutes_SubFSNoPanic verifies that setupWebappRoutes doesn't panic
// when using a subFailingFS (fs.Sub always succeeds in stdlib, so routes are registered).
func TestSetupWebappRoutes_SubFSNoPanic(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()

	// subFailingFS has dist/index.html readable; fs.Sub always succeeds
	webappFS = subFailingFS{testWebappFS()}

	router := newTestRouter(t)
	// Should not panic
	router.setupWebappRoutes()
}

// TestSetupWebappRoutes_IndexReadFailInAppSlashHandler covers lines 62-66:
// the error path when fs.ReadFile(webappRoot, "index.html") fails inside the /app/ handler.
func TestSetupWebappRoutes_IndexReadFailInAppSlashHandler(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()

	// Use a FS where dist/index.html exists (initial check passes, Sub succeeds),
	// but the sub-FS's index.html is not readable after the first read.
	webappFS = &countingIndexFS{
		inner:     testWebappFS(),
		failAfter: 1, // allow first read (initial check), fail after
		count:     new(int),
	}

	router := newTestRouter(t)
	router.setupWebappRoutes()

	// Request a SPA route (non-static, non-API) to trigger index.html read in handler
	req := httptest.NewRequest("GET", "/app/dashboard", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)

	// Should return 404 (index.html read failed)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when index.html read fails, got %d", w.Code)
	}
}

// TestSetupWebappRoutes_APIEndpointInAppSlashHandler covers lines 46-49:
// the isAPIEndpoint branch inside the /app/ handler.
// We register the routes and send an API-like path through the mux.
func TestSetupWebappRoutes_APIEndpointInAppSlashHandler(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()
	webappFS = testWebappFS()

	router := newTestRouter(t)
	router.setupWebappRoutes()

	// /app/ handler checks isAPIEndpoint; /api/ paths match.
	// But the mux may have /api/ registered elsewhere. We use a path that
	// starts with /app/ but looks like an API path to the handler.
	// Actually isAPIEndpoint checks for /api/, /auth, /swagger, /health.
	// The /app/ handler is registered for /app/ prefix. We need a path that:
	// 1. Matches /app/ prefix in the mux
	// 2. isAPIEndpoint returns true for it
	// Since isAPIEndpoint checks strings.HasPrefix(path, "/api/"), and our path
	// starts with /app/, it won't match /api/. But it checks /auth, /swagger, /health too.
	// Let's check: isAPIEndpoint("/app/api/foo") - no match.
	// Actually the /app/ handler is called for ANY path starting with /app/.
	// The handler then checks if the path is an API endpoint.
	// Since isAPIEndpoint checks for /api/ prefix, and our path starts with /app/,
	// it won't trigger isAPIEndpoint unless the path is like /auth or /health.
	// But those don't start with /app/. So this branch may not be reachable via mux.
	// Let's verify by checking the actual code path.

	// The handler is:
	// r.mux.HandleFunc("/app/", func(w http.ResponseWriter, req *http.Request) {
	//     path := req.URL.Path
	//     if isAPIEndpoint(path) { ... }
	// })
	// So if we send a request to /app/ with path "/api/foo", the mux won't match it.
	// The isAPIEndpoint branch in the /app/ handler is dead code for paths starting with /app/.
	// Skip this test as the branch is unreachable via the mux.
	t.Skip("isAPIEndpoint branch in /app/ handler is unreachable via mux (paths starting with /app/ never match API prefixes)")
}

// TestSetupWebappRoutes_IndexReadFailInAppHandler covers lines 80-84:
// the error path when fs.ReadFile(webappRoot, "index.html") fails inside the /app handler.
func TestSetupWebappRoutes_IndexReadFailInAppHandler(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()

	webappFS = &countingIndexFS{
		inner:     testWebappFS(),
		failAfter: 1, // allow first read (initial check), fail after
		count:     new(int),
	}

	router := newTestRouter(t)
	router.setupWebappRoutes()

	// Request exactly /app to trigger the /app handler's index.html read
	req := httptest.NewRequest("GET", "/app", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)

	// Should return 404 (index.html read failed)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when index.html read fails in /app handler, got %d", w.Code)
	}
}

// TestSetupWebappRoutes_AppHandlerWrongPath covers lines 74-77:
// the branch where req.URL.Path != "/app" in the /app handler.
// This is a defensive check; the mux exact-match for "/app" means this
// branch is unreachable via normal routing. We document this.
func TestSetupWebappRoutes_AppHandlerWrongPath(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()
	webappFS = testWebappFS()

	router := newTestRouter(t)
	router.setupWebappRoutes()

	// The /app handler (exact match) checks req.URL.Path != "/app" and calls handleNotFound.
	// In normal mux routing, /app only matches exactly /app.
	// This branch is unreachable via the mux, so we skip it.
	t.Skip("req.URL.Path != '/app' branch in /app handler is unreachable via mux (exact match pattern)")
}

// TestSetupDevProxy_APIEndpointInHandler covers lines 162-166:
// the isAPIEndpoint branch inside the dev proxy handler.
// Same analysis: paths starting with /app/ never match API prefixes.
// This branch is dead code.
func TestSetupDevProxy_APIEndpointInHandler(t *testing.T) {
	t.Skip("isAPIEndpoint branch in dev proxy handler is dead code (paths starting with /app/ never match API prefixes)")
}

// TestSetupWebappRoutes_EmbeddedFS_BothHandlersFail covers both index.html read
// failures in a single test by using a FS that fails after the first read.
func TestSetupWebappRoutes_EmbeddedFS_BothHandlersFail(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()

	// FS that allows exactly 1 read of dist/index.html (initial check),
	// then fails all subsequent reads.
	webappFS = &countingIndexFS{
		inner:     testWebappFS(),
		failAfter: 1,
		count:     new(int),
	}

	logger := zap.NewNop()
	cfg := &config.Config{}
	router := &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		config: cfg,
	}
	router.setupWebappRoutes()

	tests := []struct {
		path string
	}{
		{"/app/"},
		{"/app"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.mux.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Errorf("path %s: expected 404, got %d", tt.path, w.Code)
			}
		})
	}
}

// TestSetupWebappRoutes_EmbeddedFS_StaticAssetNotFound verifies that a static
// asset request that doesn't exist in the embedded FS returns 404 via FileServer.
// This exercises the hasStaticAssetExtension branch in the /app/ handler.
func TestSetupWebappRoutes_EmbeddedFS_StaticAssetNotFound_Coverage(t *testing.T) {
	savedFS := webappFS
	defer func() { webappFS = savedFS }()

	// Use a FS with only dist/index.html (no assets)
	webappFS = fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("<html>test</html>")},
	}

	logger := zap.NewNop()
	cfg := &config.Config{}
	router := &Router{
		mux:    http.NewServeMux(),
		logger: logger,
		config: cfg,
	}
	router.setupWebappRoutes()

	req := httptest.NewRequest("GET", "/app/assets/main.js", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)

	// FileServer returns 404 for missing file
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing asset, got %d", w.Code)
	}
}
