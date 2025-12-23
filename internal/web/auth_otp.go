package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// handleAuthRequestOTP handles OTP request
func (r *Router) handleAuthRequestOTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	usernameOrID := strings.TrimSpace(req.FormValue("username"))
	if usernameOrID == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	// Find user by username or telegram_id
	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetUserByUsernameOrID(usernameOrID)
	if err != nil {
		r.logger.Error("failed to find user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// User not found - return error message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "User not found. Please make sure you've started a conversation with the bot first.",
		})
		return
	}

	// Generate OTP
	ttl := time.Duration(r.config.WebApp.OTPTTLSeconds) * time.Second
	code, _, err := r.otpRepo.GenerateOTP(user.ID, ttl)
	if err != nil {
		r.logger.Error("failed to generate OTP", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send OTP via bot
	msg := tgbotapi.NewMessage(user.TelegramID, fmt.Sprintf("Your login code: %s\n\nValid for %d seconds.", code, r.config.WebApp.OTPTTLSeconds))
	if _, err := r.bot.Send(msg); err != nil {
		r.logger.Error("failed to send OTP", zap.Error(err))
		http.Error(w, "Failed to send OTP", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "OTP code sent to your Telegram!",
		"user_id": user.ID,
	})
}

// handleAuthOTP handles OTP verification
func (r *Router) handleAuthOTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userIDStr := req.FormValue("user_id")
	code := strings.TrimSpace(req.FormValue("code"))
	
	// Normalize code - remove all non-digit characters
	code = normalizeOTPCode(code)

	r.logger.Info("OTP validation attempt", 
		zap.String("user_id", userIDStr),
		zap.String("code_length", fmt.Sprintf("%d", len(code))),
		zap.String("code_preview", maskCode(code)))

	if userIDStr == "" || code == "" {
		r.logger.Warn("OTP validation failed: missing user_id or code")
		http.Error(w, "user_id and code are required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		r.logger.Warn("OTP validation failed: invalid user_id", zap.Error(err))
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	// Validate OTP
	otp, err := r.otpRepo.ValidateOTP(userID, code)
	if err != nil {
		r.logger.Warn("OTP validation failed", 
			zap.Int64("user_id", userID),
			zap.String("code_length", fmt.Sprintf("%d", len(code))),
			zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid or expired OTP code. Please try again.",
		})
		return
	}

	r.logger.Info("OTP validated successfully", zap.Int64("user_id", userID), zap.Int64("otp_id", otp.ID))

	// Create session
	auth := r.getAuthMiddleware()
	if err := auth.CreateSession(w, req, otp.UserID); err != nil {
		r.logger.Error("failed to create session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Authentication successful",
	})
}

// normalizeOTPCode removes all non-digit characters from the code
func normalizeOTPCode(code string) string {
	var result strings.Builder
	for _, r := range code {
		if r >= '0' && r <= '9' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// maskCode masks the OTP code for logging (shows first and last character)
func maskCode(code string) string {
	if len(code) == 0 {
		return ""
	}
	if len(code) <= 2 {
		return "**"
	}
	return string(code[0]) + strings.Repeat("*", len(code)-2) + string(code[len(code)-1])
}

