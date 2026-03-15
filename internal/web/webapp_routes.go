package web

import (
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// webappFS is initialized in webapp_static.go from webapp package (embed.FS implements fs.FS)
var webappFS fs.FS

// setupWebappRoutes configures routes for serving the embedded webapp
func (r *Router) setupWebappRoutes() {
	// Check if webappFS is empty (test mode) - in dev mode, Vite serves static files
	// Try to read index.html to check if files are embedded
	_, err := fs.ReadFile(webappFS, "dist/index.html")
	if err != nil {
		r.logger.Info("webapp files not embedded (test/dev mode) - proxying static requests to Vite dev server")
		r.setupDevProxy()
		return
	}

	// Get the embedded filesystem, stripping the dist prefix
	webappRoot, _ := fs.Sub(webappFS, "dist")

	fileServer := http.FileServer(http.FS(webappRoot))

	// Serve static assets (JS, CSS, images, etc.)
	// Strip /app prefix before serving files
	// This must be registered before the general /app/ handler to ensure assets are served correctly
	r.mux.Handle("/app/assets/", http.StripPrefix("/app", fileServer))

	// Serve other static files (favicon, robots.txt, etc.)
	r.mux.HandleFunc("/app/", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		// If it's a request for a real static file, try to serve it.
		// IMPORTANT: we can't treat "any dot in URL" as a file extension,
		// because our SPA route params can contain dots (e.g. sectionId "en.grammar.first_sentences").
		if hasStaticAssetExtension(path) && path != "/app/" {
			// Strip /app prefix before serving
			http.StripPrefix("/app", fileServer).ServeHTTP(w, req)
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
		"/api/",
		"/auth",
		"/swagger",
		"/health",
	}

	for _, apiPath := range apiPaths {
		if strings.HasPrefix(path, apiPath) {
			return true
		}
	}

	return false
}

// hasStaticAssetExtension checks if the request looks like a static asset.
// We intentionally use a whitelist of extensions because SPA routes may contain dots.
func hasStaticAssetExtension(path string) bool {
	if path == "" || strings.HasSuffix(path, "/") {
		return false
	}

	// Only look at the last path segment
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(base))
	if ext == "" || len(ext) <= 1 {
		return false
	}

	switch ext {
	case ".js", ".mjs", ".css", ".map",
		".ico", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp",
		".txt", ".xml", ".webmanifest", ".json",
		".woff", ".woff2", ".ttf", ".eot":
		return true
	default:
		return false
	}
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
