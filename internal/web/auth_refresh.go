package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// handleAuthRefresh handles refresh token request
// @Summary      Обновление токенов
// @Description  Обновляет access токен используя refresh токен. Возвращает новую пару access и refresh токенов. Формат запроса: `{"refresh_token": "ваш_refresh_токен"}`
// @Tags         Auth
// @Accept       application/json
// @Produce      application/json
// @Param        request  body  RefreshTokenRequest  true  "Refresh token request"
// @Success      200  {object}  map[string]interface{}  "Новая пара токенов"
// @Failure      400  {object}  map[string]interface{}  "Неверный запрос (отсутствует refresh_token)"
// @Failure      401  {object}  map[string]interface{}  "Неверный или истекший refresh токен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /auth/refresh [post]
func (r *Router) handleAuthRefresh(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Try to parse JSON body first
	if err := json.NewDecoder(req.Body).Decode(&requestBody); err != nil {
		// Fallback to form data
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}
		requestBody.RefreshToken = strings.TrimSpace(req.FormValue("refresh_token"))
	}

	if requestBody.RefreshToken == "" {
		r.logger.Warn("refresh token is empty")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "refresh_token is required",
		})
		return
	}

	// Validate refresh token
	auth := r.getAuthMiddleware()
	userID, err := auth.jwtService.ValidateRefreshToken(requestBody.RefreshToken)
	if err != nil {
		r.logger.Warn("refresh token validation failed", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	// Get user to retrieve telegramID for role determination
	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		r.logger.Error("failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate new token pair
	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID, user.TelegramID)
	if err != nil {
		r.logger.Error("failed to generate new token pair", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return new token pair
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Tokens refreshed successfully",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
	})
}

