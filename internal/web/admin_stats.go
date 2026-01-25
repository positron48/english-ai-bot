package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// handleAdminStats returns observability statistics for admin panel
// @Summary      Получить статистику системы
// @Description  Возвращает метрики по пользователям, активности, тренировкам и карточкам
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        days  query     int     false  "Количество дней для daily timeseries (по умолчанию 30)"
// @Success      200  {object}  map[string]interface{}  "Статистика системы"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Router       /api/admin/stats [get]
func (r *Router) handleAdminStats(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse days parameter (default 30)
	days := 30
	if daysStr := req.URL.Query().Get("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 && parsedDays <= 365 {
			days = parsedDays
		}
	}

	now := time.Now()
	day24hAgo := now.AddDate(0, 0, -1)
	day7dAgo := now.AddDate(0, 0, -7)
	day30dAgo := now.AddDate(0, 0, -30)
	dayNDaysAgo := now.AddDate(0, 0, -days)

	// Helper function to calculate accuracy
	calculateAccuracy := func(total, correct int) float64 {
		if total > 0 {
			return float64(correct) / float64(total) * 100
		}
		return 0
	}

	// Helper function to count active users (users with training_sessions or review_events in window)
	countActiveUsers := func(since time.Time) (int, error) {
		query := `SELECT COUNT(DISTINCT user_id) FROM (
			SELECT user_id FROM training_sessions WHERE started_at >= ?
			UNION
			SELECT user_id FROM review_events WHERE answered_at >= ? AND answered_at IS NOT NULL
		)`
		var count int
		err := r.db.QueryRow(query, since, since).Scan(&count)
		return count, err
	}

	// Windows: 24h, 7d, 30d metrics
	windows := make(map[string]map[string]interface{})

	// Users metrics
	var usersTotal int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&usersTotal)
	if err != nil {
		r.logger.Error("failed to get total users", zap.Error(err))
		usersTotal = 0
	}

	var usersNew24h, usersNew7d, usersNew30d int
	r.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= ?", day24hAgo).Scan(&usersNew24h)
	r.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= ?", day7dAgo).Scan(&usersNew7d)
	r.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= ?", day30dAgo).Scan(&usersNew30d)

	usersActive24h, _ := countActiveUsers(day24hAgo)
	usersActive7d, _ := countActiveUsers(day7dAgo)
	usersActive30d, _ := countActiveUsers(day30dAgo)

	// Sessions metrics
	var sessionsStarted24h, sessionsStarted7d, sessionsStarted30d int
	r.db.QueryRow("SELECT COUNT(*) FROM training_sessions WHERE started_at >= ?", day24hAgo).Scan(&sessionsStarted24h)
	r.db.QueryRow("SELECT COUNT(*) FROM training_sessions WHERE started_at >= ?", day7dAgo).Scan(&sessionsStarted7d)
	r.db.QueryRow("SELECT COUNT(*) FROM training_sessions WHERE started_at >= ?", day30dAgo).Scan(&sessionsStarted30d)

	var sessionsCompleted24h, sessionsCompleted7d, sessionsCompleted30d int
	completedQuery := `SELECT COUNT(*) FROM training_sessions 
		WHERE started_at >= ? AND ended_at IS NOT NULL AND done_count > 0`
	r.db.QueryRow(completedQuery, day24hAgo).Scan(&sessionsCompleted24h)
	r.db.QueryRow(completedQuery, day7dAgo).Scan(&sessionsCompleted7d)
	r.db.QueryRow(completedQuery, day30dAgo).Scan(&sessionsCompleted30d)

	// Reviews metrics
	var reviewsAnswered24h, reviewsAnswered7d, reviewsAnswered30d int
	var reviewsCorrect24h, reviewsCorrect7d, reviewsCorrect30d int
	reviewsQuery24h := `SELECT 
		COUNT(*) as total,
		COALESCE(SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END), 0) as correct
		FROM review_events WHERE answered_at >= ? AND answered_at IS NOT NULL`
	r.db.QueryRow(reviewsQuery24h, day24hAgo).Scan(&reviewsAnswered24h, &reviewsCorrect24h)
	r.db.QueryRow(reviewsQuery24h, day7dAgo).Scan(&reviewsAnswered7d, &reviewsCorrect7d)
	r.db.QueryRow(reviewsQuery24h, day30dAgo).Scan(&reviewsAnswered30d, &reviewsCorrect30d)

	// Cards added metrics (exclude orphaned cards)
	cardsAddedQuery := `SELECT COUNT(*) FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.created_at >= ?`
	var cardsAdded24h, cardsAdded7d, cardsAdded30d int
	r.db.QueryRow(cardsAddedQuery, day24hAgo).Scan(&cardsAdded24h)
	r.db.QueryRow(cardsAddedQuery, day7dAgo).Scan(&cardsAdded7d)
	r.db.QueryRow(cardsAddedQuery, day30dAgo).Scan(&cardsAdded30d)

	windows["24h"] = map[string]interface{}{
		"users_new":          usersNew24h,
		"users_active":       usersActive24h,
		"sessions_started":   sessionsStarted24h,
		"sessions_completed": sessionsCompleted24h,
		"reviews_answered":   reviewsAnswered24h,
		"reviews_correct":    reviewsCorrect24h,
		"accuracy_percent":   calculateAccuracy(reviewsAnswered24h, reviewsCorrect24h),
		"cards_added":        cardsAdded24h,
	}

	windows["7d"] = map[string]interface{}{
		"users_new":          usersNew7d,
		"users_active":       usersActive7d,
		"sessions_started":   sessionsStarted7d,
		"sessions_completed": sessionsCompleted7d,
		"reviews_answered":   reviewsAnswered7d,
		"reviews_correct":    reviewsCorrect7d,
		"accuracy_percent":   calculateAccuracy(reviewsAnswered7d, reviewsCorrect7d),
		"cards_added":        cardsAdded7d,
	}

	windows["30d"] = map[string]interface{}{
		"users_new":          usersNew30d,
		"users_active":       usersActive30d,
		"sessions_started":   sessionsStarted30d,
		"sessions_completed": sessionsCompleted30d,
		"reviews_answered":   reviewsAnswered30d,
		"reviews_correct":    reviewsCorrect30d,
		"accuracy_percent":   calculateAccuracy(reviewsAnswered30d, reviewsCorrect30d),
		"cards_added":        cardsAdded30d,
	}

	// Cards state (global)
	var userCardsTotal, userCardsNew, userCardsLearning, userCardsReview int
	cardsStateQuery := `SELECT 
		COUNT(*) as total,
		SUM(CASE WHEN uc.state = 'new' THEN 1 ELSE 0 END) as new_count,
		SUM(CASE WHEN uc.state = 'learning' THEN 1 ELSE 0 END) as learning_count,
		SUM(CASE WHEN uc.state = 'review' THEN 1 ELSE 0 END) as review_count
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id`
	err = r.db.QueryRow(cardsStateQuery).Scan(&userCardsTotal, &userCardsNew, &userCardsLearning, &userCardsReview)
	if err != nil {
		r.logger.Error("failed to get cards state", zap.Error(err))
	}

	var dueNowTotal int
	dueQuery := `SELECT COUNT(*) FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.state != 'new' AND (uc.next_due_at IS NULL OR uc.next_due_at <= ?)`
	r.db.QueryRow(dueQuery, now).Scan(&dueNowTotal)

	cardsState := map[string]interface{}{
		"total":    userCardsTotal,
		"new":      userCardsNew,
		"learning": userCardsLearning,
		"review":   userCardsReview,
		"due_now":  dueNowTotal,
	}

	// Daily timeseries for charts
	dailyStats := make([]map[string]interface{}, 0)
	
	// Generate all days in range
	dayMap := make(map[string]map[string]interface{})
	for i := 0; i < days; i++ {
		day := dayNDaysAgo.AddDate(0, 0, i)
		dayStr := day.Format("2006-01-02")
		dayMap[dayStr] = map[string]interface{}{
			"day":              dayStr,
			"active_users":     0,
			"sessions_started": 0,
			"reviews_answered": 0,
			"cards_added":      0,
			"accuracy_percent": 0.0,
		}
	}

	// Active users by day
	activeUsersQuery := `SELECT DATE(activity_date) as day, COUNT(DISTINCT user_id) as count FROM (
		SELECT DATE(started_at) as activity_date, user_id FROM training_sessions WHERE started_at >= ?
		UNION
		SELECT DATE(answered_at) as activity_date, user_id FROM review_events WHERE answered_at >= ? AND answered_at IS NOT NULL
	) GROUP BY DATE(activity_date)`
	activeUsersRows, err := r.db.Query(activeUsersQuery, dayNDaysAgo, dayNDaysAgo)
	if err == nil {
		defer activeUsersRows.Close()
		for activeUsersRows.Next() {
			var day string
			var count int
			if err := activeUsersRows.Scan(&day, &count); err == nil {
				if dayData, ok := dayMap[day]; ok {
					dayData["active_users"] = count
				}
			}
		}
	} else {
		r.logger.Warn("failed to get active users daily stats", zap.Error(err))
	}

	// Sessions started by day
	sessionsQuery := `SELECT DATE(started_at) as day, COUNT(*) as count
		FROM training_sessions WHERE started_at >= ?
		GROUP BY DATE(started_at)`
	sessionsRows, err := r.db.Query(sessionsQuery, dayNDaysAgo)
	if err == nil {
		defer sessionsRows.Close()
		for sessionsRows.Next() {
			var day string
			var count int
			if err := sessionsRows.Scan(&day, &count); err == nil {
				if dayData, ok := dayMap[day]; ok {
					dayData["sessions_started"] = count
				}
			}
		}
	} else {
		r.logger.Warn("failed to get sessions daily stats", zap.Error(err))
	}

	// Reviews answered by day with accuracy
	reviewsQuery := `SELECT 
		DATE(answered_at) as day,
		COUNT(*) as total,
		COALESCE(SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END), 0) as correct
		FROM review_events 
		WHERE answered_at >= ? AND answered_at IS NOT NULL
		GROUP BY DATE(answered_at)`
	reviewsRows, err := r.db.Query(reviewsQuery, dayNDaysAgo)
	if err == nil {
		defer reviewsRows.Close()
		for reviewsRows.Next() {
			var day string
			var total, correct int
			if err := reviewsRows.Scan(&day, &total, &correct); err == nil {
				if dayData, ok := dayMap[day]; ok {
					dayData["reviews_answered"] = total
					dayData["accuracy_percent"] = calculateAccuracy(total, correct)
				}
			}
		}
	} else {
		r.logger.Warn("failed to get reviews daily stats", zap.Error(err))
	}

	// Cards added by day
	cardsAddedDailyQuery := `SELECT DATE(uc.created_at) as day, COUNT(*) as count
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.created_at >= ?
		GROUP BY DATE(uc.created_at)`
	cardsAddedRows, err := r.db.Query(cardsAddedDailyQuery, dayNDaysAgo)
	if err == nil {
		defer cardsAddedRows.Close()
		for cardsAddedRows.Next() {
			var day string
			var count int
			if err := cardsAddedRows.Scan(&day, &count); err == nil {
				if dayData, ok := dayMap[day]; ok {
					dayData["cards_added"] = count
				}
			}
		}
	} else {
		r.logger.Warn("failed to get cards added daily stats", zap.Error(err))
	}

	// Convert map to sorted slice
	for i := 0; i < days; i++ {
		day := dayNDaysAgo.AddDate(0, 0, i)
		dayStr := day.Format("2006-01-02")
		if dayData, ok := dayMap[dayStr]; ok {
			dailyStats = append(dailyStats, dayData)
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users_total": usersTotal,
		"windows":     windows,
		"cards_state": cardsState,
		"daily":       dailyStats,
	})
}
