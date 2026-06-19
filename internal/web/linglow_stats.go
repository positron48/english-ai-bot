package web

import (
	"errors"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// handleLinglowStats returns streak, weekly rhythm, monthly metrics and per-mode skills.
// @Summary      Сводная статистика обучения
// @Description  Стрик, ритм недели, метрики месяца (минуты/слова/тексты/диалоги), навыки по режимам, любимый район
// @Tags         Linglow
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/linglow/stats [get]
func (r *Router) handleLinglowStats(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	explicitCourseCode := strings.TrimSpace(req.URL.Query().Get("course_code"))
	month := strings.TrimSpace(req.URL.Query().Get("month"))

	stats, err := r.courseRepo.GetStatsForUser(req.Context(), userID, r.defaultCourseCode(), explicitCourseCode, month)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to get linglow stats", zap.Error(err), zap.Int64("user_id", userID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}
