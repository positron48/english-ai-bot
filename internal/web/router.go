package web

import (
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Router handles web routes
type Router struct {
	mux            *http.ServeMux
	templates      *template.Template
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
	sessionRepo        interface{} // Will be properly typed later
	otpRepo            *repository.WebOTPRepository
	botToken           string
	webTrainingHandler *WebTrainingHandler
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
	r := &Router{
		mux:             http.NewServeMux(),
		logger:          logger,
		config:          cfg,
		db:              db,
		trainingService: trainingService,
		srsService:      srsService,
		optionsService:  optionsService,
		cbService:       cbService,
	}

	// Load templates - parse base first, then others
	tmpl := template.New("base.html")
	tmpl, err := tmpl.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		logger.Fatal("failed to parse templates", zap.Error(err))
	}
	r.templates = tmpl

	// Setup public routes only (protected routes will be added after SetDependencies)
	r.setupPublicRoutes()

	return r
}

// SetDependencies sets additional dependencies
func (r *Router) SetDependencies(
	userRepo interface{},
	wordService interface{},
	aiService interface{},
	bot *tgbotapi.BotAPI,
	sessionRepo interface{},
	botToken string,
) {
	r.userRepo = userRepo
	r.wordService = wordService
	r.aiService = aiService
	r.bot = bot
	r.sessionRepo = sessionRepo
	r.botToken = botToken

	// Create auth middleware
	r.authMiddleware = NewAuthMiddleware(
		userRepo.(*repository.UserRepository),
		sessionRepo.(*repository.WebSessionRepository),
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

// setupPublicRoutes configures public routes (called during initialization)
func (r *Router) setupPublicRoutes() {
	// Static files
	staticFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		r.logger.Fatal("failed to create static sub FS", zap.Error(err))
	}
	r.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Public routes
	r.mux.HandleFunc("/app", r.handleApp)
	r.mux.HandleFunc("/login", r.handleLogin)

	// Auth routes
	r.mux.HandleFunc("/auth/telegram", r.handleAuthTelegram)
	r.mux.HandleFunc("/auth/request_otp", r.handleAuthRequestOTP)
	r.mux.HandleFunc("/auth/otp", r.handleAuthOTP)
	r.mux.HandleFunc("/logout", r.handleLogout)
}

// setupProtectedRoutes configures protected routes (called after SetDependencies)
func (r *Router) setupProtectedRoutes() {
	auth := r.getAuthMiddleware()
	if auth == nil {
		r.logger.Fatal("auth middleware not initialized - call SetDependencies first")
	}

	// Protected user routes (wrapped with auth middleware)
	// Register more specific routes first to avoid prefix matching issues
	r.mux.HandleFunc("/app/dashboard", auth.RequireAuth(r.handleDashboard))
	r.mux.HandleFunc("/app/vocab", auth.RequireAuth(r.handleVocab))
	// Use exact path matching for vocab delete - register specific patterns
	r.mux.HandleFunc("/app/vocab/", auth.RequireAuth(r.handleVocabDelete))
	r.mux.HandleFunc("/app/training/start", auth.RequireAuth(r.handleTrainingStart))
	r.mux.HandleFunc("/app/training/current", auth.RequireAuth(r.handleTrainingCurrent))
	r.mux.HandleFunc("/app/training/reveal", auth.RequireAuth(r.handleTrainingReveal))
	r.mux.HandleFunc("/app/training/answer", auth.RequireAuth(r.handleTrainingAnswer))
	r.mux.HandleFunc("/app/chat", auth.RequireAuth(r.handleChat))

	// Admin routes (wrapped with admin guard)
	adminAuth := auth.RequireAuth
	adminGuard := r.RequireAdmin
	r.mux.HandleFunc("/app/admin", adminAuth(adminGuard(r.handleAdmin)))
	r.mux.HandleFunc("/app/admin/circuit/reset", adminAuth(adminGuard(r.handleAdminCircuitReset)))
	r.mux.HandleFunc("/app/admin/training/", adminAuth(adminGuard(r.handleAdminTraining)))
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Log request for debugging (use Info level to see in production)
	r.logger.Info("handling request",
		zap.String("method", req.Method),
		zap.String("path", req.URL.Path),
		zap.String("remote_addr", req.RemoteAddr),
	)
	r.mux.ServeHTTP(w, req)
}

// renderTemplate renders a template with data
func (r *Router) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		r.logger.Error("failed to render template", zap.String("template", name), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleApp is the main entry point
func (r *Router) handleApp(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	auth := r.getAuthMiddleware()
	if auth == nil {
		// Auth middleware not initialized, redirect to login
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}
	userID, err := auth.GetUserFromSession(req)
	if err != nil || userID == 0 {
		// Not authenticated, redirect to login
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	// User is authenticated, redirect to dashboard
	http.Redirect(w, req, "/app/dashboard", http.StatusFound)
}

// handleLogin shows login page
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("handleLogin called", zap.String("path", req.URL.Path), zap.String("method", req.Method))
	
	// Validate path is exactly /login
	if req.URL.Path != "/login" {
		r.logger.Error("handleLogin called with invalid path", zap.String("path", req.URL.Path))
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.logger.Info("rendering login template")
	r.renderTemplate(w, "login.html", map[string]interface{}{
		"Title": "Login",
	})
}

// handleAuthTelegram handles Telegram WebApp initData authentication
func (r *Router) handleAuthTelegram(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Create session
	if err := auth.CreateSession(w, user.ID); err != nil {
		r.logger.Error("failed to create session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard
	http.Redirect(w, req, "/app/dashboard", http.StatusFound)
}

// handleAuthRequestOTP and handleAuthOTP are implemented in auth_otp.go

func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session from cookie
	cookie, err := req.Cookie("session")
	if err == nil && cookie != nil {
		sessionRepo := r.sessionRepo.(*repository.WebSessionRepository)
		_ = sessionRepo.DeleteSession(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, req, "/login", http.StatusFound)
}

// handleDashboard and handleChat are implemented in dashboard.go
// handleVocab and handleVocabDelete are implemented in vocab.go
// handleTrainingStart, handleTrainingCurrent, handleTrainingReveal, handleTrainingAnswer are implemented in training.go
// handleAdmin, handleAdminCircuitReset, handleAdminTraining are implemented in admin.go

