package web

import (
	"context"
	"encoding/json"
	"net/http"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func (r *Router) getUserTierFromDB(ctx context.Context) models.UserTier {
	userID := getUserIDFromContext(ctx)
	if userID == 0 || r.userRepo == nil {
		return models.TierFree
	}
	reader, ok := r.userRepo.(subscriptionTierReader)
	if !ok {
		return models.TierFree
	}
	user, err := reader.GetUserByID(userID)
	if err != nil || user == nil {
		return models.TierFree
	}
	return user.SubscriptionTier
}

type subscriptionTierReader interface {
	GetUserByID(userID int64) (*models.User, error)
}

// userAllowsFeature checks subscription tier feature access.
func (r *Router) userAllowsFeature(ctx context.Context, feature string) bool {
	return models.TierAllowsFeature(r.getUserTierFromDB(ctx), feature)
}

// RequireFeature wraps a handler requiring a subscription tier feature (e.g. "speaking").
func (r *Router) RequireFeature(feature string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			if !r.userAllowsFeature(req.Context(), feature) {
				r.logger.Warn("subscription feature denied",
					zap.String("path", req.URL.Path),
					zap.String("feature", feature),
					zap.Int64("user_id", getUserIDFromContext(req.Context())))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Forbidden",
					"message": "This feature requires a Pro subscription.",
					"feature": feature,
				})
				return
			}
			next(w, req)
		}
	}
}

func (r *Router) speakingModeEnabled() bool {
	return r.config != nil && r.config.Speaking.Enabled
}
