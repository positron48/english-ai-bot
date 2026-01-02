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

// webappFS is initialized in webapp_static.go from webappembed package
var webappFS embed.FS

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
	webappRoot, err := fs.Sub(webappFS, "dist")
	if err != nil {
		r.logger.Warn("failed to create webapp filesystem sub", zap.Error(err))
		return
	}

	fileServer := http.FileServer(http.FS(webappRoot))

	// Serve static assets (JS, CSS, images, etc.)
	// Strip /app prefix before serving files
	// This must be registered before the general /app/ handler to ensure assets are served correctly
	r.mux.Handle("/app/assets/", http.StripPrefix("/app", fileServer))

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

		// Inject debug script for Telegram Mini App debugging
		indexHTML := injectDebugScript(string(indexData))
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
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

		// Inject debug script for Telegram Mini App debugging
		indexHTML := injectDebugScript(string(indexData))
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})
}

// injectDebugScript injects debug JavaScript into index.html for Telegram Mini App debugging
func injectDebugScript(html string) string {
	debugScript := `
    <style>
      #debug-info {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        background: #ff4444;
        color: white;
        padding: 10px;
        font-family: monospace;
        font-size: 12px;
        z-index: 99999;
        display: block;
        max-height: 300px;
        overflow-y: auto;
      }
      #debug-info button {
        margin-left: 10px;
        padding: 5px 10px;
        background: white;
        color: #ff4444;
        border: none;
        cursor: pointer;
        border-radius: 3px;
      }
    </style>
    <div id="debug-info"></div>
    <script>
      (function() {
        const debugDiv = document.getElementById('debug-info');
        const messages = [];
        
        function addMessage(msg, type) {
          const time = new Date().toISOString().split('T')[1].split('.')[0];
          messages.push({msg, type, time});
          updateDisplay();
        }
        
        function updateDisplay() {
          if (messages.length === 0) return;
          const html = '<strong>🔍 Debug Info (Telegram Mini App):</strong><br>' + 
            messages.map(m => {
              const icon = m.type === 'error' ? '❌' : m.type === 'warning' ? '⚠️' : m.type === 'success' ? '✅' : '🔄';
              return '[' + m.time + '] ' + icon + ' ' + m.msg;
            }).join('<br>') +
            '<button onclick="document.getElementById(\'debug-info\').style.display=\'none\'">Hide</button>';
          debugDiv.innerHTML = html;
        }
        
        // Check Telegram WebApp
        function checkTelegram() {
          if (typeof window.Telegram === 'undefined') {
            addMessage('Telegram WebApp script NOT loaded', 'error');
          } else {
            addMessage('Telegram WebApp script loaded', 'success');
            const tg = window.Telegram?.WebApp;
            if (tg) {
              addMessage('Telegram.WebApp object available', 'success');
              if (tg.initData) {
                addMessage('initData available (length: ' + tg.initData.length + ')', 'success');
              } else {
                addMessage('initData NOT available', 'warning');
              }
            } else {
              addMessage('Telegram.WebApp object NOT available', 'error');
            }
          }
        }
        
        // Check Vue app mounting
        function checkVueApp() {
          setTimeout(function() {
            const appDiv = document.getElementById('app');
            if (!appDiv) {
              addMessage('app div NOT found in DOM', 'error');
            } else if (appDiv.innerHTML.trim() === '') {
              addMessage('Vue app did not mount - app div is empty', 'error');
            } else {
              addMessage('Vue app mounted successfully', 'success');
            }
          }, 2000);
        }
        
        window.addEventListener('load', function() {
          addMessage('Page loaded, checking components...', 'info');
          setTimeout(checkTelegram, 500);
          checkVueApp();
        });
        
        // Global error handlers
        window.addEventListener('error', function(e) {
          addMessage('JavaScript Error: ' + e.message + ' at ' + (e.filename || 'unknown') + ':' + e.lineno, 'error');
        });
        
        window.addEventListener('unhandledrejection', function(e) {
          addMessage('Unhandled Promise Rejection: ' + (e.reason?.message || e.reason || 'Unknown'), 'error');
        });
        
        addMessage('Debug script initialized', 'info');
      })();
    </script>
`
	
	// Insert debug script before closing </body> tag
	if idx := strings.LastIndex(html, "</body>"); idx > 0 {
		return html[:idx] + debugScript + "\n" + html[idx:]
	}
	
	// If no </body> tag, insert before closing </html>
	if idx := strings.LastIndex(html, "</html>"); idx > 0 {
		return html[:idx] + debugScript + "\n" + html[idx:]
	}
	
	// If no closing tags, append at the end
	return html + debugScript
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

