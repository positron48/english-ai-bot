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

		// Inject console.log interceptor for Telegram Mini App debugging
		indexHTML := injectConsoleLogger(string(indexData))
		
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

		// Inject console.log interceptor for Telegram Mini App debugging
		indexHTML := injectConsoleLogger(string(indexData))
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexHTML))
	})
}

// injectConsoleLogger injects console.log interceptor into index.html for Telegram Mini App debugging
func injectConsoleLogger(html string) string {
	consoleLoggerScript := `
    <style>
      #console-logger {
        position: fixed;
        bottom: 0;
        left: 0;
        right: 0;
        background: #1e1e1e;
        color: #d4d4d4;
        font-family: 'Courier New', monospace;
        font-size: 11px;
        max-height: 400px;
        overflow-y: auto;
        z-index: 99999;
        border-top: 2px solid #007acc;
        padding: 10px;
        box-shadow: 0 -2px 10px rgba(0,0,0,0.3);
      }
      #console-logger-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;
        padding-bottom: 5px;
        border-bottom: 1px solid #444;
      }
      #console-logger-header strong {
        color: #4ec9b0;
      }
      #console-logger button {
        padding: 5px 10px;
        background: #007acc;
        color: white;
        border: none;
        cursor: pointer;
        border-radius: 3px;
        font-size: 11px;
      }
      #console-logger button:hover {
        background: #005a9e;
      }
      .log-entry {
        margin: 2px 0;
        padding: 2px 5px;
        word-break: break-all;
      }
      .log-entry.log { color: #d4d4d4; }
      .log-entry.info { color: #4ec9b0; }
      .log-entry.warn { color: #dcdcaa; background: rgba(220, 220, 170, 0.1); }
      .log-entry.error { color: #f48771; background: rgba(244, 135, 113, 0.1); }
      .log-time {
        color: #808080;
        margin-right: 8px;
      }
    </style>
    <div id="console-logger">
      <div id="console-logger-header">
        <strong>📋 Console Logs</strong>
        <button onclick="document.getElementById('console-logger').style.display='none'">Hide</button>
      </div>
      <div id="console-logger-content"></div>
    </div>
    <script>
      (function() {
        const loggerDiv = document.getElementById('console-logger-content');
        const maxLogs = 500;
        let logCount = 0;
        
        function formatLog(args) {
          return Array.from(args).map(arg => {
            if (typeof arg === 'object') {
              try {
                return JSON.stringify(arg, null, 2);
              } catch (e) {
                return String(arg);
              }
            }
            return String(arg);
          }).join(' ');
        }
        
        function addLog(level, args) {
          const time = new Date().toISOString().split('T')[1].split('.')[0];
          const logEntry = document.createElement('div');
          logEntry.className = 'log-entry ' + level;
          logEntry.innerHTML = '<span class="log-time">[' + time + ']</span>' + 
            '<span class="log-level">[' + level.toUpperCase() + ']</span> ' + 
            formatLog(args);
          loggerDiv.appendChild(logEntry);
          logCount++;
          
          // Keep only last maxLogs entries
          if (logCount > maxLogs) {
            loggerDiv.removeChild(loggerDiv.firstChild);
            logCount--;
          }
          
          // Auto-scroll to bottom
          loggerDiv.scrollTop = loggerDiv.scrollHeight;
        }
        
        // Intercept console methods
        const originalLog = console.log;
        const originalInfo = console.info;
        const originalWarn = console.warn;
        const originalError = console.error;
        
        console.log = function(...args) {
          originalLog.apply(console, args);
          addLog('log', args);
        };
        
        console.info = function(...args) {
          originalInfo.apply(console, args);
          addLog('info', args);
        };
        
        console.warn = function(...args) {
          originalWarn.apply(console, args);
          addLog('warn', args);
        };
        
        console.error = function(...args) {
          originalError.apply(console, args);
          addLog('error', args);
        };
        
        // Log initialization
        console.log('[Console Logger] Initialized - all console.log calls will be displayed here');
      })();
    </script>
`
	
	// Insert console logger before closing </body> tag
	if idx := strings.LastIndex(html, "</body>"); idx > 0 {
		return html[:idx] + consoleLoggerScript + "\n" + html[idx:]
	}
	
	// If no </body> tag, insert before closing </html>
	if idx := strings.LastIndex(html, "</html>"); idx > 0 {
		return html[:idx] + consoleLoggerScript + "\n" + html[idx:]
	}
	
	// If no closing tags, append at the end
	return html + consoleLoggerScript
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

