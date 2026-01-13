package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// @title           English Bot API
// @version         1.0
// @description     API для English Bot - системы изучения английского языка с использованием SRS (Spaced Repetition System)
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8184
// @BasePath  /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer JWT token authentication. В Swagger UI введите только токен (без "Bearer "), префикс будет добавлен автоматически.

// @tag.name Auth
// @tag.description Методы аутентификации (Telegram WebApp и OTP)

// @tag.name Dashboard
// @tag.description Главная панель пользователя

// @tag.name Vocab
// @tag.description Управление словарным запасом

// @tag.name Training
// @tag.description Система тренировок и обучения

// @tag.name Chat
// @tag.description AI чат для помощи в изучении языка

// @tag.name Admin
// @tag.description Административные функции

// Router handles web routes
type Router struct {
	mux            *http.ServeMux
	logger         *zap.Logger
	config         *config.Config
	db             *sql.DB
	userRepo       interface{} // Will be properly typed later
	trainingService *service.TrainingService
	srsService      *service.SRSService
	optionsService  *service.OptionsService
	wordService     interface{} // Will be properly typed later
	cbService       *service.CircuitBreakerService
	aiService          interface{} // Will be properly typed later
	bot                *tgbotapi.BotAPI
	authMiddleware     *AuthMiddleware
	otpRepo            *repository.WebOTPRepository
	botToken           string
	webTrainingHandler *WebTrainingHandler
	rateLimiter        *RateLimiter
}

// NewRouter creates a new web router
func NewRouter(
	logger *zap.Logger,
	cfg *config.Config,
	db *sql.DB,
	trainingService *service.TrainingService,
	srsService *service.SRSService,
	optionsService *service.OptionsService,
	cbService *service.CircuitBreakerService,
) *Router {
	// Initialize rate limiter (cleanup every 5 minutes, TTL 1 hour)
	rateLimiter := NewRateLimiter(5*time.Minute, 1*time.Hour)

	r := &Router{
		mux:             http.NewServeMux(),
		logger:          logger,
		config:          cfg,
		db:              db,
		trainingService: trainingService,
		srsService:      srsService,
		optionsService:  optionsService,
		cbService:       cbService,
		rateLimiter:     rateLimiter,
	}

	// Setup routes
	r.setupRoutes()

	return r
}

// SetDependencies sets additional dependencies
func (r *Router) SetDependencies(
	userRepo interface{},
	wordService interface{},
	aiService interface{},
	bot *tgbotapi.BotAPI,
	botToken string,
) {
	r.userRepo = userRepo
	r.wordService = wordService
	r.aiService = aiService
	r.bot = bot
	r.botToken = botToken

	// Create JWT service
	jwtService, err := NewJWTService(r.config, r.logger)
	if err != nil {
		r.logger.Fatal("failed to create JWT service", zap.Error(err))
	}

	// Create auth middleware
	r.authMiddleware = NewAuthMiddleware(
		userRepo.(*repository.UserRepository),
		jwtService,
		r.logger,
		r.config,
		botToken,
	)
	
	// Setup protected routes now that auth middleware is initialized
	r.setupProtectedRoutes()
}

// SetOTPRepo sets the OTP repository
func (r *Router) SetOTPRepo(otpRepo *repository.WebOTPRepository) {
	r.otpRepo = otpRepo
}

// getAuthMiddleware returns the auth middleware
func (r *Router) getAuthMiddleware() *AuthMiddleware {
	return r.authMiddleware
}

// getRateLimitPolicy creates a rate limit policy from config values
func (r *Router) getRateLimitPolicy(requestsPerWindow, burstMultiplier int) RateLimitPolicy {
	// Use defaults if config values are 0 or negative
	if requestsPerWindow <= 0 {
		requestsPerWindow = 60 // Safe default
	}
	windowMinutes := r.config.WebApp.RateLimitWindowMinutes
	if windowMinutes <= 0 {
		windowMinutes = 1
	}
	if burstMultiplier <= 0 {
		burstMultiplier = 2
	}
	burst := requestsPerWindow * burstMultiplier
	if burst < requestsPerWindow {
		burst = requestsPerWindow
	}
	return RateLimitPolicy{
		RequestsPerWindow: requestsPerWindow,
		WindowDuration:    time.Duration(windowMinutes) * time.Minute,
		BurstSize:         burst,
	}
}

// setupRoutes configures all routes
func (r *Router) setupRoutes() {
	// Swagger documentation with custom UI that auto-adds "Bearer " prefix
	r.mux.HandleFunc("/swagger/", r.swaggerHandler)
	r.mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, req, "docs/swagger/swagger.json")
	})

	// Auth routes with rate limiting
	// POST /auth/telegram - moderate limit per IP
	telegramPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthTelegramPerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	telegramMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, telegramPolicy, KeyFuncIP)
	r.mux.HandleFunc("/auth/telegram", telegramMiddleware.Wrap(r.handleAuthTelegram))

	// POST /auth/telegram_unsafe - strict limit per IP, stricter per IP+user_id
	telegramUnsafePolicyIP := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthTelegramUnsafePerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	telegramUnsafePolicyIPUser := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthTelegramUnsafePerIPUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	telegramUnsafeMiddlewareIP := NewRateLimitMiddleware(r.rateLimiter, r.logger, telegramUnsafePolicyIP, KeyFuncIP)
	telegramUnsafeMiddlewareIPUser := NewRateLimitMiddleware(r.rateLimiter, r.logger, telegramUnsafePolicyIPUser, KeyFuncIPAndUserID)
	r.mux.HandleFunc("/auth/telegram_unsafe", telegramUnsafeMiddlewareIP.Wrap(telegramUnsafeMiddlewareIPUser.Wrap(r.handleAuthTelegramUnsafe)))

	// POST /auth/request_otp - strict limit per IP, stricter per IP+username
	requestOTPPolicyIP := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthRequestOTPPerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	requestOTPPolicyIPUser := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthRequestOTPPerIPUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	requestOTPMiddlewareIP := NewRateLimitMiddleware(r.rateLimiter, r.logger, requestOTPPolicyIP, KeyFuncIP)
	requestOTPMiddlewareIPUser := NewRateLimitMiddleware(r.rateLimiter, r.logger, requestOTPPolicyIPUser, KeyFuncIPAndUsername)
	r.mux.HandleFunc("/auth/request_otp", requestOTPMiddlewareIP.Wrap(requestOTPMiddlewareIPUser.Wrap(r.handleAuthRequestOTP)))

	// POST /auth/otp - moderate limit per IP, stricter per IP+user_id
	otpPolicyIP := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthOTPPerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	otpPolicyIPUser := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthOTPPerIPUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	otpMiddlewareIP := NewRateLimitMiddleware(r.rateLimiter, r.logger, otpPolicyIP, KeyFuncIP)
	otpMiddlewareIPUser := NewRateLimitMiddleware(r.rateLimiter, r.logger, otpPolicyIPUser, KeyFuncIPAndUserID)
	r.mux.HandleFunc("/auth/otp", otpMiddlewareIP.Wrap(otpMiddlewareIPUser.Wrap(r.handleAuthOTP)))

	// POST /auth/refresh - moderate limit per IP
	refreshPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAuthRefreshPerIP,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	refreshMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, refreshPolicy, KeyFuncIP)
	r.mux.HandleFunc("/auth/refresh", refreshMiddleware.Wrap(r.handleAuthRefresh))
	
	// Health check
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok"}`)
	})

	// Webapp static files (must be registered after API routes to avoid conflicts)
	r.setupWebappRoutes()
}

// setupProtectedRoutes configures protected routes (called after SetDependencies)
func (r *Router) setupProtectedRoutes() {
	auth := r.getAuthMiddleware()
	if auth == nil {
		r.logger.Fatal("auth middleware not initialized - call SetDependencies first")
	}

	// Rate limit policies for /app/* routes
	appAPIPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAppAPIPerUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	appChatPolicy := r.getRateLimitPolicy(
		r.config.WebApp.RateLimitAppChatPerUser,
		r.config.WebApp.RateLimitBurstMultiplier,
	)
	appAPIMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, appAPIPolicy, KeyFuncIPAndUserIDFromContext)
	appChatMiddleware := NewRateLimitMiddleware(r.rateLimiter, r.logger, appChatPolicy, KeyFuncIPAndUserIDFromContext)

	// Protected user routes (wrapped with auth middleware and rate limiting)
	r.mux.HandleFunc("/app/dashboard", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleDashboard)))
	r.mux.HandleFunc("/app/vocab", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVocab)))
	r.mux.HandleFunc("/app/vocab/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleVocabDelete)))
	
	// Learning word sets routes
	r.mux.HandleFunc("/app/learning/words/categories", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningWordsCategories)))
	r.mux.HandleFunc("/app/learning/words/sets", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningWordsSets)))
	r.mux.HandleFunc("/app/learning/words/sets/", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleLearningWordsSetDetailOrStudy)))
	
	r.mux.HandleFunc("/app/training/start", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingStart)))
	r.mux.HandleFunc("/app/training/current", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingCurrent)))
	r.mux.HandleFunc("/app/training/reveal", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingReveal)))
	r.mux.HandleFunc("/app/training/answer", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingAnswer)))
	r.mux.HandleFunc("/app/training/upcoming", appAPIMiddleware.Wrap(auth.RequireAuth(r.handleTrainingUpcoming)))
	r.mux.HandleFunc("/app/chat", appChatMiddleware.Wrap(auth.RequireAuth(r.handleChat)))

	// Admin routes (wrapped with admin guard and rate limiting)
	adminAuth := auth.RequireAuth
	adminGuard := r.RequireAdmin
	r.mux.HandleFunc("/app/admin", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdmin))))
	r.mux.HandleFunc("/app/admin/circuit/reset", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminCircuitReset))))
	r.mux.HandleFunc("/app/admin/training/", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminTraining))))
	r.mux.HandleFunc("/app/admin/training/card/", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminTrainingCard))))
	r.mux.HandleFunc("/app/admin/words", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminWords))))
	r.mux.HandleFunc("/app/admin/words/", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminWord))))
	r.mux.HandleFunc("/app/admin/users", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminUsers))))
	r.mux.HandleFunc("/app/admin/db-schema", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleDBSchema))))
	r.mux.HandleFunc("/app/admin/orphaned-cards", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminOrphanedCards))))
	r.mux.HandleFunc("/app/admin/orphaned-cards/", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminOrphanedCard))))
	r.mux.HandleFunc("/app/admin/orphaned-user-cards", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminOrphanedUserCards))))
	r.mux.HandleFunc("/app/admin/orphaned-user-cards/", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminOrphanedUserCard))))
	r.mux.HandleFunc("/app/admin/prompt-tester/default-prompts", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminPromptTesterDefaultPrompts))))
	r.mux.HandleFunc("/app/admin/prompt-tester/run", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminPromptTesterRun))))
	
	// Word sets admin routes
	r.mux.HandleFunc("/app/admin/word-set-categories", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminWordSetCategories))))
	r.mux.HandleFunc("/app/admin/word-set-categories/", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminWordSetCategories))))
	r.mux.HandleFunc("/app/admin/word-sets", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminWordSets))))
	r.mux.HandleFunc("/app/admin/word-sets/", appAPIMiddleware.Wrap(adminAuth(adminGuard(r.handleAdminWordSetDetailOrSets))))
}

// corsMiddleware adds CORS headers to allow Swagger UI to make requests
func (r *Router) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Set CORS headers
		origin := req.Header.Get("Origin")
		
		// Allow requests from localhost origins (for Swagger UI)
		allowedOrigins := []string{
			"http://localhost:8184",
			"http://127.0.0.1:8184",
			"http://localhost:8080",
			"http://127.0.0.1:8080",
		}
		
		if origin != "" {
			// Check if origin is in allowed list
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
			// If origin matches localhost pattern, allow it
			if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			// Allow Telegram Web App origins (web.telegram.org, etc.)
			if strings.Contains(origin, "telegram.org") || strings.Contains(origin, "t.me") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			// Allow same-origin requests (when app is served from same domain)
			if strings.HasPrefix(origin, "https://") || strings.HasPrefix(origin, "http://") {
				// For production, allow requests from the same domain
				// This handles cases where webapp is served from the same domain as API
				if w.Header().Get("Access-Control-Allow-Origin") == "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}
		} else {
			// For requests without Origin header (same-origin), allow all
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		
		// Only set credentials if we have a specific origin (not *)
		if origin != "" && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")) {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, req)
	}
}

// swaggerResponseWriter wraps http.ResponseWriter to inject JavaScript for auto Bearer prefix
type swaggerResponseWriter struct {
	http.ResponseWriter
	buf         []byte
	headerWritten bool
	statusCode    int
}

func (w *swaggerResponseWriter) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.buf = append(w.buf, b...)
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

func (w *swaggerResponseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.statusCode = statusCode
	w.headerWritten = true
	
	// Process buffered content only if it's HTML
	if len(w.buf) > 0 {
		contentType := w.Header().Get("Content-Type")
		content := string(w.buf)
		
		// Only modify HTML content, not JSON, CSS, JS, etc.
		isHTML := strings.Contains(contentType, "text/html") || 
			strings.Contains(content, "<html") || 
			strings.Contains(content, "<!DOCTYPE") ||
			strings.Contains(content, "swagger-ui-bundle.js")
		
		if isHTML {
			// Inject JavaScript before closing body tag
			js := `<script>
(function() {
    'use strict';
    
    // Function to add Bearer prefix if missing
    function ensureBearerPrefix(value) {
        if (!value) return value;
        const trimmed = value.trim();
        if (trimmed && !trimmed.startsWith('Bearer ')) {
            return 'Bearer ' + trimmed;
        }
        return trimmed;
    }
    
    // Function to setup input field watcher
    function setupInputWatcher(input) {
        if (input.dataset.bearerWatcherAdded) return;
        input.dataset.bearerWatcherAdded = 'true';
        
        // Watch for input changes
        input.addEventListener('input', function() {
            const value = this.value;
            if (value && !value.startsWith('Bearer ') && value.length > 10) {
                // Don't auto-add while typing, wait for blur or authorize
            }
        });
        
        // Add Bearer on blur
        input.addEventListener('blur', function() {
            const value = this.value;
            if (value && !value.startsWith('Bearer ')) {
                this.value = ensureBearerPrefix(value);
            }
        });
        
        // Add Bearer on Enter key
        input.addEventListener('keydown', function(e) {
            if (e.key === 'Enter') {
                const value = this.value;
                if (value && !value.startsWith('Bearer ')) {
                    this.value = ensureBearerPrefix(value);
                }
            }
        });
    }
    
    // Function to intercept Swagger UI authorize
    function interceptSwaggerAuthorize() {
        // Try multiple ways to access Swagger UI system
        let system = null;
        if (window.ui && window.ui.getSystem) {
            system = window.ui.getSystem();
        } else if (window.SwaggerUIBundle) {
            // Try to get system from SwaggerUIBundle
            const swaggerUI = document.querySelector('#swagger-ui');
            if (swaggerUI && swaggerUI.swaggerUISystem) {
                system = swaggerUI.swaggerUISystem;
            }
        }
        
        if (system && system.authActions) {
            const originalAuthorize = system.authActions.authorize;
            if (originalAuthorize) {
                system.authActions.authorize = function(payload) {
                    // Process all auth schemes
                    if (payload) {
                        Object.keys(payload).forEach(function(key) {
                            if (payload[key] && payload[key].value) {
                                const value = payload[key].value;
                                if (typeof value === 'string' && value.trim() && !value.trim().startsWith('Bearer ')) {
                                    payload[key].value = ensureBearerPrefix(value);
                                }
                            }
                        });
                    }
                    return originalAuthorize.call(this, payload);
                };
                return true;
            }
        }
        return false;
    }
    
    // Store refresh token for automatic token refresh
    let refreshToken = null;
    
    // Function to get refresh token from localStorage or Swagger UI
    function getRefreshToken() {
        // Try to get from localStorage first
        const stored = localStorage.getItem('refresh_token');
        if (stored) return stored;
        
        // Try to get from Swagger UI auth state
        try {
            const system = window.ui ? window.ui.getSystem() : null;
            if (system && system.getState) {
                const state = system.getState();
                if (state && state.auth && state.auth.authorized) {
                    // Check if refresh token is stored somewhere in auth state
                    // This is a fallback - refresh token should be stored separately
                }
            }
        } catch (e) {
            console.warn('Could not get refresh token from Swagger UI state', e);
        }
        
        return null;
    }
    
    // Function to refresh access token using refresh token
    async function refreshAccessToken() {
        const refresh = getRefreshToken();
        if (!refresh) {
            console.warn('[Swagger] No refresh token available for automatic token refresh');
            return null;
        }
        
        try {
            const response = await fetch('/auth/refresh', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ refresh_token: refresh })
            });
            
            if (response.ok) {
                const data = await response.json();
                if (data.access_token && data.refresh_token) {
                    // Update tokens
                    refreshToken = data.refresh_token;
                    localStorage.setItem('refresh_token', data.refresh_token);
                    
                    // Update Swagger UI authorization
                    const system = window.ui ? window.ui.getSystem() : null;
                    if (system && system.authActions) {
                        system.authActions.authorize({
                            ApiKeyAuth: {
                                name: 'ApiKeyAuth',
                                value: 'Bearer ' + data.access_token
                            }
                        });
                    }
                    
                    console.log('[Swagger] Access token refreshed automatically');
                    return data.access_token;
                }
            } else {
                console.warn('[Swagger] Failed to refresh token:', response.status);
                // Clear invalid refresh token
                localStorage.removeItem('refresh_token');
                refreshToken = null;
            }
        } catch (error) {
            console.error('[Swagger] Error refreshing token:', error);
        }
        
        return null;
    }
    
    // Function to intercept fetch/XHR requests and add Bearer if needed, and handle 401 errors
    function interceptRequests() {
        // Intercept fetch - only for API requests, not for Swagger UI assets
        const originalFetch = window.fetch;
        window.fetch = function(...args) {
            const url = args[0];
            // Don't intercept requests to Swagger UI assets or doc.json
            if (typeof url === 'string' && (url.includes('/swagger/') || url.includes('swagger-ui') || url.includes('doc.json'))) {
                return originalFetch.apply(this, args);
            }
            
            // Add Bearer prefix if needed
            if (args[1] && args[1].headers) {
                const authHeader = args[1].headers.get ? args[1].headers.get('Authorization') : args[1].headers['Authorization'];
                if (authHeader && !authHeader.startsWith('Bearer ')) {
                    if (args[1].headers instanceof Headers) {
                        args[1].headers.set('Authorization', ensureBearerPrefix(authHeader));
                    } else {
                        args[1].headers['Authorization'] = ensureBearerPrefix(authHeader);
                    }
                }
            }
            
            // Make the request
            return originalFetch.apply(this, args).then(async (response) => {
                // If 401 Unauthorized, try to refresh token and retry once
                if (response.status === 401 && !url.includes('/auth/')) {
                    const newAccessToken = await refreshAccessToken();
                    if (newAccessToken) {
                        // Retry the request with new token
                        const newHeaders = args[1] ? { ...args[1].headers } : {};
                        if (newHeaders instanceof Headers) {
                            newHeaders.set('Authorization', 'Bearer ' + newAccessToken);
                        } else {
                            newHeaders['Authorization'] = 'Bearer ' + newAccessToken;
                        }
                        const retryArgs = [args[0], { ...args[1], headers: newHeaders }];
                        return originalFetch.apply(this, retryArgs);
                    }
                }
                return response;
            });
        };
        
        // Intercept XMLHttpRequest - only for API requests
        const originalSetRequestHeader = XMLHttpRequest.prototype.setRequestHeader;
        const originalSend = XMLHttpRequest.prototype.send;
        
        XMLHttpRequest.prototype.setRequestHeader = function(header, value) {
            // Don't intercept if it's a Swagger UI asset request
            const url = this.responseURL || this._url || '';
            if (url.includes('/swagger/') || url.includes('swagger-ui') || url.includes('doc.json')) {
                return originalSetRequestHeader.call(this, header, value);
            }
            
            if (header.toLowerCase() === 'authorization' && value && !value.startsWith('Bearer ')) {
                value = ensureBearerPrefix(value);
            }
            return originalSetRequestHeader.call(this, header, value);
        };
        
        // Also intercept send to handle 401 responses
        XMLHttpRequest.prototype.send = function(...args) {
            const xhr = this;
            const originalOnReadyStateChange = xhr.onreadystatechange;
            
            xhr.onreadystatechange = async function() {
                if (xhr.readyState === 4 && xhr.status === 401 && !xhr.responseURL.includes('/auth/')) {
                    const newAccessToken = await refreshAccessToken();
                    if (newAccessToken) {
                        // Create new request with refreshed token
                        const newXhr = new XMLHttpRequest();
                        newXhr.open(xhr._method || 'GET', xhr.responseURL);
                        newXhr.setRequestHeader('Authorization', 'Bearer ' + newAccessToken);
                        // Copy other headers
                        for (let header of xhr.getAllResponseHeaders().split('\r\n')) {
                            const parts = header.split(': ');
                            if (parts.length === 2 && parts[0].toLowerCase() !== 'authorization') {
                                newXhr.setRequestHeader(parts[0], parts[1]);
                            }
                        }
                        newXhr.send(args[0]);
                        // Replace xhr properties
                        Object.defineProperty(xhr, 'status', { value: newXhr.status, writable: true });
                        Object.defineProperty(xhr, 'responseText', { value: newXhr.responseText, writable: true });
                        Object.defineProperty(xhr, 'response', { value: newXhr.response, writable: true });
                    }
                }
                if (originalOnReadyStateChange) {
                    originalOnReadyStateChange.apply(this, arguments);
                }
            };
            
            return originalSend.apply(this, args);
        };
    }
    
    // Function to save refresh token when user authorizes
    function setupRefreshTokenSaving() {
        // Monitor Swagger UI authorize actions
        const system = window.ui ? window.ui.getSystem() : null;
        if (system && system.authActions) {
            const originalAuthorize = system.authActions.authorize;
            system.authActions.authorize = function(payload) {
                // Check if this is setting a token
                if (payload && payload.ApiKeyAuth && payload.ApiKeyAuth.value) {
                    const token = payload.ApiKeyAuth.value;
                    // If token starts with Bearer, extract it
                    const actualToken = token.startsWith('Bearer ') ? token.substring(7) : token;
                    // Try to decode JWT to check if it's an access token
                    try {
                        const parts = actualToken.split('.');
                        if (parts.length === 3) {
                            const payload = JSON.parse(atob(parts[1]));
                            // Check if it has exp claim (access token) vs longer exp (refresh token)
                            // For now, we'll rely on user to save refresh token manually or via response
                        }
                    } catch (e) {
                        // Not a valid JWT, ignore
                    }
                }
                return originalAuthorize.call(this, payload);
            };
        }
        
        // Also intercept responses from auth endpoints to save refresh token
        const originalFetch = window.fetch;
        window.fetch = function(...args) {
            const url = args[0];
            if (typeof url === 'string' && (url.includes('/auth/otp') || url.includes('/auth/telegram'))) {
                return originalFetch.apply(this, args).then(async (response) => {
                    if (response.ok) {
                        const data = await response.json();
                        if (data.refresh_token) {
                            refreshToken = data.refresh_token;
                            localStorage.setItem('refresh_token', data.refresh_token);
                            console.log('[Swagger] Refresh token saved for automatic token refresh');
                        }
                    }
                    return response;
                });
            }
            return originalFetch.apply(this, args);
        };
    }
    
    // Wait for Swagger UI to load
    function init() {
        // Load refresh token from localStorage
        refreshToken = getRefreshToken();
        if (refreshToken) {
            console.log('[Swagger] Refresh token loaded from storage');
        }
        
        // Setup request interceptors immediately
        interceptRequests();
        
        // Setup refresh token saving
        setupRefreshTokenSaving();
        
        // Try to intercept Swagger authorize
        if (interceptSwaggerAuthorize()) {
            console.log('[Swagger] Bearer prefix auto-addition enabled');
        }
        
        // Setup MutationObserver to watch for new input fields
        const observer = new MutationObserver(function(mutations) {
            // Look for input fields
            const inputs = document.querySelectorAll('input[type="password"], input[type="text"][name*="token" i], input[placeholder*="token" i], input[placeholder*="Token" i], input[placeholder*="authorization" i], input[placeholder*="Authorization" i]');
            inputs.forEach(setupInputWatcher);
            
            // Also look for authorize buttons and intercept clicks
            const authButtons = document.querySelectorAll('button.authorize, .authorize-btn, [class*="authorize"]');
            authButtons.forEach(function(btn) {
                if (!btn.dataset.bearerClickAdded) {
                    btn.dataset.bearerClickAdded = 'true';
                    btn.addEventListener('click', function() {
                        setTimeout(function() {
                            const inputs = document.querySelectorAll('input[type="password"], input[type="text"]');
                            inputs.forEach(function(input) {
                                const value = input.value;
                                if (value && !value.startsWith('Bearer ')) {
                                    input.value = ensureBearerPrefix(value);
                                }
                            });
                        }, 100);
                    });
                }
            });
        });
        
        observer.observe(document.body, {
            childList: true,
            subtree: true,
            attributes: false
        });
        
        // Initial setup
        setTimeout(function() {
            const inputs = document.querySelectorAll('input[type="password"], input[type="text"]');
            inputs.forEach(setupInputWatcher);
        }, 1000);
    }
    
    // Start initialization
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
    
    // Also try after a delay to catch late-loading Swagger UI
    setTimeout(init, 2000);
})();
</script>`
			// Insert before </body> tag
			if idx := strings.LastIndex(content, "</body>"); idx > 0 {
				content = content[:idx] + js + "\n" + content[idx:]
				w.buf = []byte(content)
			} else {
				// If no </body> tag, append before </html> or at the end
				if idx := strings.LastIndex(content, "</html>"); idx > 0 {
					content = content[:idx] + js + "\n" + content[idx:]
					w.buf = []byte(content)
				} else {
					// Append at the end
					w.buf = append(w.buf, []byte(js)...)
				}
			}
		}
	}
	
	// Write header and buffered content
	w.ResponseWriter.WriteHeader(w.statusCode)
	if len(w.buf) > 0 {
		w.ResponseWriter.Write(w.buf)
		w.buf = nil
	}
}

func (w *swaggerResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// swaggerHandler serves Swagger UI with custom JavaScript to auto-add "Bearer " prefix
func (r *Router) swaggerHandler(w http.ResponseWriter, req *http.Request) {
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)
	
	// For HTML pages only (not static files), wrap response writer to inject JavaScript
	path := req.URL.Path
	isHTMLPage := (path == "/swagger/" || path == "/swagger/index.html" || path == "/swagger") && 
		!strings.Contains(path, ".") && 
		!strings.HasSuffix(path, ".js") && 
		!strings.HasSuffix(path, ".css") && 
		!strings.HasSuffix(path, ".json") &&
		!strings.HasSuffix(path, ".png") &&
		!strings.HasSuffix(path, ".ico")
	
	if isHTMLPage {
		wrapped := &swaggerResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		swaggerHandler.ServeHTTP(wrapped, req)
		// Flush any remaining buffer if header wasn't written
		if !wrapped.headerWritten && len(wrapped.buf) > 0 {
			wrapped.WriteHeader(http.StatusOK)
		}
		return
	}
	
	swaggerHandler.ServeHTTP(w, req)
}

// handleNotFound handles 404 errors
func (r *Router) handleNotFound(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error": "Not Found", "message": "The requested resource was not found", "path": "%s"}`, req.URL.Path)
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply CORS middleware to all requests
	r.corsMiddleware(func(w http.ResponseWriter, req *http.Request) {
		r.mux.ServeHTTP(w, req)
	})(w, req)
}

// handleAuthTelegram handles Telegram WebApp initData authentication
// @Summary      Аутентификация через Telegram WebApp
// @Description  Аутентификация пользователя через Telegram WebApp initData. Возвращает пару JWT токенов (access и refresh) для авторизованного пользователя.
// @Tags         Auth
// @Accept       application/x-www-form-urlencoded
// @Produce      application/json
// @Param        initData  formData  string  true  "Telegram WebApp initData (подписанные данные от Telegram)"
// @Success      200  {object}  map[string]interface{}  "Успешная аутентификация с JWT токенами"
// @Failure      400  {string}  string  "Неверный запрос (отсутствует initData)"
// @Failure      401  {string}  string  "Неавторизован (неверный initData)"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /auth/telegram [post]
func (r *Router) handleAuthTelegram(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := req.ParseForm(); err != nil {
		r.logger.Error("failed to parse form", zap.Error(err))
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	initData := req.FormValue("initData")
	if initData == "" {
		http.Error(w, "initData is required", http.StatusBadRequest)
		return
	}

	// Validate initData and get user ID
	auth := r.getAuthMiddleware()
	telegramID, err := auth.ValidateTelegramInitData(initData)
	if err != nil {
		r.logger.Warn("failed to validate initData", zap.Error(err))
		http.Error(w, "Invalid initData", http.StatusUnauthorized)
		return
	}

	// Get or create user
	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		r.logger.Error("failed to get/create user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate JWT token pair
	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID)
	if err != nil {
		r.logger.Error("failed to generate JWT tokens", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response with JWT tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Authentication successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// handleAuthTelegramUnsafe handles Telegram WebApp authentication using initDataUnsafe (less secure, fallback)
// @Summary      Аутентификация через Telegram WebApp (небезопасный метод)
// @Description  Аутентификация пользователя через user_id напрямую (менее безопасный метод, используется как fallback). Возвращает пару JWT токенов (access и refresh) для авторизованного пользователя.
// @Tags         Auth
// @Accept       application/x-www-form-urlencoded
// @Produce      application/json
// @Param        user_id  formData  string  true  "Telegram User ID"
// @Success      200  {object}  map[string]interface{}  "Успешная аутентификация с JWT токенами"
// @Failure      400  {string}  string  "Неверный запрос (отсутствует user_id)"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /auth/telegram_unsafe [post]
func (r *Router) handleAuthTelegramUnsafe(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("handleAuthTelegramUnsafe called", 
		zap.String("method", req.Method),
		zap.String("path", req.URL.Path),
		zap.String("remote_addr", req.RemoteAddr))
	
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID directly from initDataUnsafe (less secure, but works when initData is not available)
	userIDStr := req.FormValue("user_id")
	r.logger.Info("received user_id", zap.String("user_id", userIDStr))
	
	if userIDStr == "" {
		r.logger.Warn("user_id is empty")
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	telegramID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		r.logger.Warn("invalid user_id", zap.String("user_id", userIDStr), zap.Error(err))
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	r.logger.Info("authenticating via initDataUnsafe", zap.Int64("telegram_id", telegramID))

	// Get or create user
	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		r.logger.Error("failed to get/create user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate JWT token pair
	auth := r.getAuthMiddleware()
	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID)
	if err != nil {
		r.logger.Error("failed to generate JWT tokens", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response with JWT tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Authentication successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

// handleAuthRequestOTP and handleAuthOTP are implemented in auth_otp.go

// handleDashboard and handleChat are implemented in dashboard.go
// handleVocab and handleVocabDelete are implemented in vocab.go
// handleTrainingStart, handleTrainingCurrent, handleTrainingReveal, handleTrainingAnswer are implemented in training.go
// handleAdmin, handleAdminCircuitReset, handleAdminTraining are implemented in admin.go

