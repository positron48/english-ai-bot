package web

import (
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
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<div class="error">User not found. Please make sure you've started a conversation with the bot first.</div>`)
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

	// Return success message with OTP input form
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<div class="success">OTP code sent to your Telegram! Enter it below:</div>
		<form hx-post="/auth/otp" hx-target="#otp-section" hx-swap="innerHTML">
			<input type="hidden" name="user_id" value="%d">
			<div>
				<label for="otp_code">Enter OTP Code:</label>
				<input type="text" id="otp_code" name="code" required maxlength="6" pattern="[0-9]{6}">
			</div>
			<button type="submit">Verify</button>
		</form>`, user.ID)
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

	if userIDStr == "" || code == "" {
		http.Error(w, "user_id and code are required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	// Validate OTP
	otp, err := r.otpRepo.ValidateOTP(userID, code)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<div class="error">Invalid or expired OTP code. Please try again.</div>`)
		return
	}

	// Create session
	auth := r.getAuthMiddleware()
	if err := auth.CreateSession(w, otp.UserID); err != nil {
		r.logger.Error("failed to create session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard
	w.Header().Set("HX-Redirect", "/app/dashboard")
	w.WriteHeader(http.StatusOK)
}

