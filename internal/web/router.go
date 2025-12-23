package web

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

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

// setupRoutes configures all routes
func (r *Router) setupRoutes() {
	// Auth routes
	r.mux.HandleFunc("/auth/telegram", r.handleAuthTelegram)
	r.mux.HandleFunc("/auth/telegram_unsafe", r.handleAuthTelegramUnsafe)
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

// handleAuthTelegram handles Telegram WebApp initData authentication
func (r *Router) handleAuthTelegram(w http.ResponseWriter, req *http.Request) {
	r.logger.Info("handleAuthTelegram called", zap.String("method", req.Method))
	
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	initData := req.FormValue("initData")
	r.logger.Info("received initData", 
		zap.String("initData_length", fmt.Sprintf("%d", len(initData))),
		zap.String("initData_preview", maskInitData(initData)))
	
	if initData == "" {
		r.logger.Warn("initData is empty")
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
	
	r.logger.Info("initData validated successfully", zap.Int64("telegram_id", telegramID))

	// Get or create user
	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		r.logger.Error("failed to get/create user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
	if err := auth.CreateSession(w, req, user.ID); err != nil {
		r.logger.Error("failed to create session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true, "message": "Authentication successful"}`)
}

// handleAuthTelegramUnsafe handles Telegram WebApp authentication using initDataUnsafe (less secure, fallback)
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

	// Create session
	auth := r.getAuthMiddleware()
	if err := auth.CreateSession(w, req, user.ID); err != nil {
		r.logger.Error("failed to create session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true, "message": "Authentication successful"}`)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true, "message": "Logged out successfully"}`)
}

// maskInitData masks the initData for logging (shows first 20 and last 10 characters)
func maskInitData(initData string) string {
	if len(initData) == 0 {
		return ""
	}
	if len(initData) <= 30 {
		return strings.Repeat("*", len(initData))
	}
	return initData[:20] + "..." + initData[len(initData)-10:]
}

// handleDashboard and handleChat are implemented in dashboard.go
// handleVocab and handleVocabDelete are implemented in vocab.go
// handleTrainingStart, handleTrainingCurrent, handleTrainingReveal, handleTrainingAnswer are implemented in training.go
// handleAdmin, handleAdminCircuitReset, handleAdminTraining are implemented in admin.go

