package web

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (r *Router) ensureVerbFormUserCardsAfterWord(req *http.Request, userID, wordCardID int64) {
	if !r.verbFormsEnabledForUser(req.Context(), userID) {
		return
	}
	started := time.Now()
	vs := r.newVerbTrainingServiceForUser(req.Context(), userID)
	err := vs.EnsureVerbFormUserCardsForWord(userID, wordCardID, r.getUserVerbScopes(req.Context(), userID))
	r.logger.Debug("prepared verb cards for vocabulary word", zap.Int64("word_card_id", wordCardID), zap.Duration("duration", time.Since(started)))
	if err != nil {
		r.logger.Warn("ensure verb cards for vocabulary word", zap.Int64("user_id", userID), zap.Int64("word_card_id", wordCardID), zap.Error(err))
	}
}
