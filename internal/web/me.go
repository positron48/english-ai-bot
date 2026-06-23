package web

import (
	"net/http"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// handleMe returns the authenticated user's profile basics.
// @Summary      Профиль текущего пользователя
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/me [get]
func (r *Router) handleMe(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userRepo, ok := r.userRepo.(*repository.UserRepository)
	if !ok || userRepo == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user, err := userRepo.GetUserByID(userID)
	if err != nil {
		r.logger.Error("failed to get user for /api/me", zap.Error(err), zap.Int64("user_id", userID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	tier := models.ParseUserTier(string(user.SubscriptionTier))
	writeJSON(w, map[string]interface{}{
		"id":                user.ID,
		"telegram_id":       user.TelegramID,
		"telegram_username": user.TelegramUsername,
		"created_at":        user.CreatedAt,
		"subscription_tier": string(tier),
		"features":          models.UserFeaturesForTier(tier),
	})
}
