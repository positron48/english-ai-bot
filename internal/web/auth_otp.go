package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/i18n"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// handleAuthRequestOTP handles OTP request
// @Summary      Запрос OTP кода
// @Description  Генерирует и отправляет OTP код пользователю через Telegram бота. Пользователь должен быть найден по username или telegram_id.
// @Tags         Auth
// @Accept       application/x-www-form-urlencoded
// @Produce      application/json
// @Param        username  formData  string  true  "Username или Telegram ID пользователя"
// @Success      200  {object}  map[string]interface{}  "OTP код отправлен"
// @Failure      400  {string}  string  "Неверный запрос (отсутствует username)"
// @Failure      404  {object}  map[string]interface{}  "Пользователь не найден"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /auth/request_otp [post]
func (r *Router) handleAuthRequestOTP(w http.ResponseWriter, req *http.Request) {
	lang := i18n.DetectLanguageFromRequest(req)
	
	if req.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.methodNotAllowed"),
		})
		return
	}

	if err := req.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.invalidFormData"),
		})
		return
	}

	usernameOrID := strings.TrimSpace(req.FormValue("username"))
	if usernameOrID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.usernameRequired"),
		})
		return
	}

	// Find user by username or telegram_id
	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetUserByUsernameOrID(usernameOrID)
	if err != nil {
		r.logger.Error("failed to find user", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.internalError"),
		})
		return
	}

	if user == nil {
		// User not found - return error message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.userNotFound"),
		})
		return
	}

	// Generate OTP
	ttl := time.Duration(r.config.WebApp.OTPTTLSeconds) * time.Second
	code, _, err := r.otpRepo.GenerateOTP(user.ID, ttl)
	if err != nil {
		r.logger.Error("failed to generate OTP", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.internalError"),
		})
		return
	}

	// Send OTP via bot
	if r.bot == nil {
		r.logger.Error("cannot send OTP: Telegram bot not initialized")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.otpServiceUnavailable"),
		})
		return
	}

	msg := tgbotapi.NewMessage(user.TelegramID, fmt.Sprintf("Your login code: %s\n\nValid for %d seconds.", code, r.config.WebApp.OTPTTLSeconds))
	if _, err := r.bot.Send(msg); err != nil {
		r.logger.Error("failed to send OTP", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.failedToSendOTP"),
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": i18n.T(lang, "errors.otpSent"),
		"user_id": user.ID,
	})
}

// handleAuthOTP handles OTP verification
// @Summary      Проверка OTP кода
// @Description  Проверяет OTP код и возвращает JWT токен для пользователя при успешной проверке
// @Tags         Auth
// @Accept       application/x-www-form-urlencoded
// @Produce      application/json
// @Param        user_id  formData  string  true  "ID пользователя"
// @Param        code     formData  string  true  "OTP код (только цифры)"
// @Success      200  {object}  map[string]interface{}  "Успешная аутентификация с JWT токеном"
// @Failure      400  {string}  string  "Неверный запрос (отсутствует user_id или code)"
// @Failure      401  {object}  map[string]interface{}  "Неверный или истекший OTP код"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /auth/otp [post]
func (r *Router) handleAuthOTP(w http.ResponseWriter, req *http.Request) {
	lang := i18n.DetectLanguageFromRequest(req)
	
	if req.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.methodNotAllowed"),
		})
		return
	}

	if err := req.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.invalidFormData"),
		})
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.userIdRequired"),
		})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		r.logger.Warn("OTP validation failed: invalid user_id", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.invalidUserId"),
		})
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
			"error": i18n.T(lang, "errors.invalidToken"),
		})
		return
	}

	r.logger.Info("OTP validated successfully", zap.Int64("user_id", userID), zap.Int64("otp_id", otp.ID))

	// Get user to retrieve telegramID for role determination
	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetUserByID(otp.UserID)
	if err != nil || user == nil {
		r.logger.Error("failed to get user", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.internalError"),
		})
		return
	}

	// Generate JWT token pair
	auth := r.getAuthMiddleware()
	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID, user.TelegramID)
	if err != nil {
		r.logger.Error("failed to generate JWT tokens", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.internalError"),
		})
		return
	}

	// Return success response with JWT tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       i18n.T(lang, "errors.authSuccessful"),
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
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

