package web

import (
	"embed"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// webappFS is defined in webapp_static.go (production) or webapp_static_test.go (tests)
var webappFS embed.FS

// setupWebappRoutes configures routes for serving the embedded webapp
func (r *Router) setupWebappRoutes() {
	// Check if webappFS is empty (test mode) - in dev mode, Vite serves static files
	// Try to read index.html to check if files are embedded
	_, err := fs.ReadFile(webappFS, "webapp/dist/index.html")
	if err != nil {
		r.logger.Info("webapp files not embedded (test/dev mode) - proxying static requests to Vite dev server")
		r.setupDevProxy()
		return
	}

	// Get the embedded filesystem, stripping the webapp/dist prefix
	webappRoot, err := fs.Sub(webappFS, "webapp/dist")
	if err != nil {
		r.logger.Warn("failed to create webapp filesystem sub", zap.Error(err))
		return
	}

	fileServer := http.FileServer(http.FS(webappRoot))

	// Serve static assets (JS, CSS, images, etc.)
	r.mux.HandleFunc("/app/assets/", func(w http.ResponseWriter, req *http.Request) {
		// Only serve if it's a static asset request
		if !isAPIEndpoint(req.URL.Path) {
			fileServer.ServeHTTP(w, req)
		}
	})

	// Serve other static files (favicon, robots.txt, etc.)
	r.mux.HandleFunc("/app/", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		
		// Check if it's an API endpoint - if so, don't serve static files
		if isAPIEndpoint(path) {
			r.handleNotFound(w, req)
			return
		}

		// If it's a request for a file (has extension), try to serve it
		if hasFileExtension(path) && path != "/app/" {
			fileServer.ServeHTTP(w, req)
			return
		}

		// Otherwise, serve index.html for SPA routing
		indexData, err := fs.ReadFile(webappRoot, "index.html")
		if err != nil {
			r.logger.Error("failed to read index.html", zap.Error(err))
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexData)
	})

	// Root /app path - serve index.html
	r.mux.HandleFunc("/app", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/app" {
			r.handleNotFound(w, req)
			return
		}

		indexData, err := fs.ReadFile(webappRoot, "index.html")
		if err != nil {
			r.logger.Error("failed to read index.html", zap.Error(err))
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexData)
	})
}

// isAPIEndpoint checks if the path is a known API endpoint
func isAPIEndpoint(path string) bool {
	apiPaths := []string{
		"/app/dashboard",
		"/app/vocab",
		"/app/training/",
		"/app/chat",
		"/app/admin",
	}

	for _, apiPath := range apiPaths {
		if strings.HasPrefix(path, apiPath) {
			return true
		}
	}

	return false
}

// hasFileExtension checks if the path has a file extension
func hasFileExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext != "" && len(ext) > 1
}

// setupDevProxy sets up handlers for dev mode (redirects to Vite dev server)
// In dev mode, users should access the app via Vite dev server (localhost:5173/app)
// Vite will proxy API requests to backend, avoiding circular proxy issues
func (r *Router) setupDevProxy() {
	viteDevServerURL := r.config.WebApp.ViteDevServerURL
	if viteDevServerURL == "" {
		viteDevServerURL = "http://localhost:5173" // fallback default
	}
	
	viteURL, err := url.Parse(viteDevServerURL)
	if err != nil {
		r.logger.Error("failed to parse Vite dev server URL", 
			zap.String("url", viteDevServerURL), 
			zap.Error(err))
		return
	}
	
	r.logger.Info("dev mode: webapp should be accessed via Vite dev server", 
		zap.String("vite_url", viteDevServerURL),
		zap.String("access_url", viteURL.String()+"/app/"))

	// In dev mode, redirect non-API /app requests to Vite dev server
	// API endpoints are already handled by setupProtectedRoutes
	devHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		// API endpoints should be handled by backend (already registered in setupProtectedRoutes)
		// This handler only catches non-API routes that should go to Vite
		if isAPIEndpoint(path) {
			// This shouldn't happen as API routes are registered first, but just in case
			r.handleNotFound(w, req)
			return
		}

		// Redirect to Vite dev server
		redirectURL := viteURL.String() + path
		if path == "/app" {
			redirectURL = viteURL.String() + "/app/"
		}
		http.Redirect(w, req, redirectURL, http.StatusTemporaryRedirect)
	})

	// Register handlers for /app routes (only for non-API paths)
	// Note: API endpoints are registered in setupProtectedRoutes and take precedence
	r.mux.HandleFunc("/app/", devHandler)
	r.mux.HandleFunc("/app", devHandler)
}

