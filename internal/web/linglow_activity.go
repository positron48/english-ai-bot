package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

const (
	maxActivityPingSeconds  = 120
	maxActivityDailySeconds = 16 * 3600
)

var activityModes = map[string]bool{
	"words": true, "grammar": true, "reading": true,
	"chat": true, "speaking": true, "other": true,
}

// handleLinglowActivity accumulates active study time into daily aggregates.
// @Summary      Записать активное время обучения
// @Description  Принимает heartbeat с количеством активных секунд и копит их в дневной статистике курса
// @Tags         Linglow
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/linglow/activity [post]
func (r *Router) handleLinglowActivity(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		CourseCode string `json:"course_code"`
		ClientDay  string `json:"client_day"`
		Seconds    int    `json:"seconds"`
		Mode       string `json:"mode"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Seconds <= 0 {
		http.Error(w, "seconds must be positive", http.StatusBadRequest)
		return
	}
	if body.Seconds > maxActivityPingSeconds {
		body.Seconds = maxActivityPingSeconds
	}
	mode := strings.TrimSpace(body.Mode)
	if !activityModes[mode] {
		mode = "other"
	}

	courseCode := strings.TrimSpace(body.CourseCode)
	if courseCode == "" {
		courseCode = r.currentCourseCodeForUser(req.Context(), userID)
	}
	statsRepo := repository.NewLinglowDailyStatsRepository(r.db)
	userCourseID, err := statsRepo.ResolveUserCourseID(req.Context(), userID, courseCode)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	day := repository.ValidClientDay(strings.TrimSpace(body.ClientDay))

	var existing int
	_ = r.db.QueryRow(`
		SELECT active_seconds FROM daily_course_stats
		WHERE user_course_id = ? AND local_date = CAST(? AS date)
	`, userCourseID, day).Scan(&existing)
	if existing >= maxActivityDailySeconds {
		writeJSON(w, map[string]interface{}{"accepted": false, "reason": "daily_cap"})
		return
	}

	if err := statsRepo.Bump(req.Context(), repository.DailyBump{
		UserCourseID:  userCourseID,
		Day:           day,
		Mode:          mode,
		ActiveSeconds: body.Seconds,
	}); err != nil {
		r.logger.Error("failed to bump activity", zap.Error(err), zap.Int64("user_id", userID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"accepted": true, "day": day})
}
