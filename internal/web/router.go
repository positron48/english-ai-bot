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
	// Parse base.html first to ensure it's available for other templates
	tmpl := template.New("base.html")
	
	// Parse base.html first
	baseContent, err := templatesFS.ReadFile("templates/base.html")
	if err != nil {
		logger.Fatal("failed to read base.html", zap.Error(err))
	}
	tmpl, err = tmpl.Parse(string(baseContent))
	if err != nil {
		logger.Fatal("failed to parse base.html", zap.Error(err))
	}
	
	// Then parse all other templates
	tmpl, err = tmpl.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		logger.Fatal("failed to parse templates", zap.Error(err))
	}
	
	// Log all parsed templates for debugging
	var templateNames []string
	for _, t := range tmpl.Templates() {
		templateNames = append(templateNames, t.Name())
	}
	logger.Info("parsed templates", zap.Strings("templates", templateNames))
	
	r.templates = tmpl

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
}

// SetOTPRepo sets the OTP repository
func (r *Router) SetOTPRepo(otpRepo *repository.WebOTPRepository) {
	r.otpRepo = otpRepo
}

// getAuthMiddleware returns the auth middleware
func (r *Router) getAuthMiddleware() *AuthMiddleware {
	return r.authMiddleware
}

// setupRoutes configures all routes
func (r *Router) setupRoutes() {
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

	// Protected user routes (wrapped with auth middleware)
	auth := r.getAuthMiddleware()
	r.mux.HandleFunc("/app/dashboard", auth.RequireAuth(r.handleDashboard))
	r.mux.HandleFunc("/app/vocab", auth.RequireAuth(r.handleVocab))
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
	r.mux.ServeHTTP(w, req)
}

// renderTemplate renders a template with data
func (r *Router) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// Check if template exists
	tmpl := r.templates.Lookup(name)
	if tmpl == nil {
		// Try with templates/ prefix
		tmpl = r.templates.Lookup("templates/" + name)
		if tmpl != nil {
			name = "templates/" + name
		} else {
			// Log all available template names for debugging
			var availableTemplates []string
			for _, t := range r.templates.Templates() {
				availableTemplates = append(availableTemplates, t.Name())
			}
			r.logger.Error("template not found", 
				zap.String("requested_template", name),
				zap.Strings("available_templates", availableTemplates))
			http.Error(w, "Template not found", http.StatusInternalServerError)
			return
		}
	}
	
	r.logger.Info("rendering template", 
		zap.String("template_name", name), 
		zap.Any("data_keys", getDataKeys(data)),
		zap.String("template_found_name", tmpl.Name()))
	
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		r.logger.Error("failed to render template", zap.String("template", name), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	r.logger.Info("template rendered successfully", zap.String("template_name", name))
}

// getDataKeys extracts keys from data map for logging
func getDataKeys(data interface{}) []string {
	if m, ok := data.(map[string]interface{}); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}
	return []string{"unknown"}
}

// handleApp is the main entry point
func (r *Router) handleApp(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if user is authenticated
	auth := r.getAuthMiddleware()
	if auth != nil {
		userID, err := auth.GetUserFromSession(req)
		if err == nil && userID != 0 {
			// User is authenticated, redirect to dashboard
			http.Redirect(w, req, "/app/dashboard", http.StatusFound)
			return
		}
	}

	// Not authenticated - show app.html which will handle Telegram WebApp initData
	// This allows the JavaScript in app.html to get initData from Telegram and authenticate
	r.renderTemplate(w, "app.html", map[string]interface{}{
		"Title":        "English Bot",
		"ContentBlock": "app-content",
	})
}

// handleLogin shows login page
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Log available templates for debugging
	var availableTemplates []string
	for _, t := range r.templates.Templates() {
		availableTemplates = append(availableTemplates, t.Name())
	}
	r.logger.Info("handleLogin: available templates", zap.Strings("templates", availableTemplates))
	
	// Check if login.html template exists
	loginTmpl := r.templates.Lookup("login.html")
	if loginTmpl == nil {
		r.logger.Error("login.html template not found!")
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	r.logger.Info("handleLogin: login.html template found", zap.String("template_name", loginTmpl.Name()))

	r.renderTemplate(w, "login.html", map[string]interface{}{
		"Title":        "Login",
		"ContentBlock": "login-content",
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

