package web

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// handleLinglowDistrictExtras returns the "new discovery" recommendation and
// today's tasks for one district.
// @Summary      Открытия и задачи дня района
// @Tags         Linglow
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/linglow/district-extras [get]
func (r *Router) handleLinglowDistrictExtras(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	districtCode := strings.TrimSpace(req.URL.Query().Get("district_code"))
	if districtCode == "" {
		http.Error(w, "district_code is required", http.StatusBadRequest)
		return
	}
	courseCode := strings.TrimSpace(req.URL.Query().Get("course_code"))
	if courseCode == "" {
		courseCode = r.currentCourseCodeForUser(req.Context(), userID)
	}

	ctx := req.Context()
	var courseID, userCourseID int64
	var targetLang, levelCode string
	if err := r.db.QueryRowContext(ctx, `
		SELECT c.id, uc.id, c.target_lang, d.level_code
		FROM courses c
		JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?
		JOIN districts d ON d.course_id = c.id AND d.code = ?
		WHERE c.code = ?
	`, userID, districtCode, courseCode).Scan(&courseID, &userCourseID, &targetLang, &levelCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "District not found", http.StatusNotFound)
			return
		}
		r.logger.Error("district extras lookup", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Discovery: first unread reading text at the district level.
	var discovery map[string]interface{}
	var textID, title, categoryID string
	err := r.db.QueryRowContext(ctx, `
		SELECT rt.text_id, rt.title, rt.category_id
		FROM reading_texts rt
		WHERE rt.level = ? AND rt.target_language = ?
			AND NOT EXISTS (
				SELECT 1 FROM reading_text_progress rtp
				WHERE rtp.user_id = ? AND rtp.chapter_id = rt.text_id
			)
		ORDER BY rt.category_id, rt.text_id
		LIMIT 1
	`, levelCode, targetLang, userID).Scan(&textID, &title, &categoryID)
	if err == nil {
		discovery = map[string]interface{}{
			"kind":        "reading_text",
			"text_id":     textID,
			"title":       title,
			"category_id": categoryID,
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		r.logger.Warn("district discovery lookup failed", zap.Error(err))
	}

	day := repository.LocalDayFromTime(time.Time{})

	// Task: review due words from this district.
	var dueWords, doneWords int
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT CASE WHEN si.state IN ('learning','review','relearning') AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP) THEN si.id END),
			COUNT(DISTINCT CASE WHEN ea.answered_at >= CAST(? AS date) THEN ea.learning_item_id END)
		FROM learning_items li
		LEFT JOIN srs_items si ON si.learning_item_id = li.id AND si.user_course_id = ?
		LEFT JOIN exercise_attempts ea ON ea.learning_item_id = li.id AND ea.user_course_id = ?
		JOIN districts d ON d.id = li.district_id
		WHERE d.code = ? AND d.course_id = ? AND li.item_type = 'word' AND li.status = 'published'
	`, day, userCourseID, userCourseID, districtCode, courseID).Scan(&dueWords, &doneWords); err != nil {
		r.logger.Warn("district word task lookup failed", zap.Error(err))
	}

	// Tasks done today course-wide: reading and chat.
	var readToday, chatToday int
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE event_type = 'reading_text_completed'),
			COUNT(*) FILTER (WHERE event_type = 'chat_message_sent')
		FROM learning_events
		WHERE user_course_id = ? AND event_time >= CAST(? AS date)
	`, userCourseID, day).Scan(&readToday, &chatToday); err != nil {
		r.logger.Warn("district events lookup failed", zap.Error(err))
	}

	tasks := []map[string]interface{}{}
	if dueWords > 0 {
		target := dueWords
		if target > 10 {
			target = 10
		}
		done := doneWords
		if done > target {
			done = target
		}
		tasks = append(tasks, map[string]interface{}{"kind": "review_words", "target": target, "done": done})
	}
	if discovery != nil || readToday > 0 {
		done := 0
		if readToday > 0 {
			done = 1
		}
		tasks = append(tasks, map[string]interface{}{"kind": "read_text", "target": 1, "done": done})
	}
	chatDone := 0
	if chatToday > 0 {
		chatDone = 1
	}
	tasks = append(tasks, map[string]interface{}{"kind": "chat", "target": 1, "done": chatDone})

	writeJSON(w, map[string]interface{}{
		"discovery": discovery,
		"tasks":     tasks,
	})
}
