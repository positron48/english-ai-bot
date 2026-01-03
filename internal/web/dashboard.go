package web

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// handleDashboard shows the user dashboard
// @Summary      Получить данные дашборда
// @Description  Возвращает расширенную статистику по карточкам и тренировкам
// @Tags         Dashboard
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Данные дашборда"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Router       /app/dashboard [get]
func (r *Router) handleDashboard(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	now := time.Now()
	
	// Get due count (cards ready for review)
	dueQuery := `SELECT COUNT(*) FROM user_cards WHERE user_id = ? AND (next_due_at IS NULL OR next_due_at <= ?)`
	var dueCount int
	err := r.db.QueryRow(dueQuery, userID, now).Scan(&dueCount)
	if err != nil {
		r.logger.Error("failed to get due count", zap.Error(err))
		dueCount = 0
	}

	// Get new cards count
	newQuery := `SELECT COUNT(*) FROM user_cards WHERE user_id = ? AND state = 'new'`
	var newCount int
	err = r.db.QueryRow(newQuery, userID).Scan(&newCount)
	if err != nil {
		r.logger.Error("failed to get new cards count", zap.Error(err))
		newCount = 0
	}

	// Get learning cards count
	learningQuery := `SELECT COUNT(*) FROM user_cards WHERE user_id = ? AND state = 'learning'`
	var learningCount int
	err = r.db.QueryRow(learningQuery, userID).Scan(&learningCount)
	if err != nil {
		r.logger.Error("failed to get learning count", zap.Error(err))
		learningCount = 0
	}

	// Get review cards count
	reviewQuery := `SELECT COUNT(*) FROM user_cards WHERE user_id = ? AND state = 'review'`
	var reviewCount int
	err = r.db.QueryRow(reviewQuery, userID).Scan(&reviewCount)
	if err != nil {
		r.logger.Error("failed to get review count", zap.Error(err))
		reviewCount = 0
	}

	// Calculate available cards for training
	// Show actual available count, not limited by session size
	// Session will still be limited to MaxCardsPerSession (30), but we show total available
	availableForTraining := dueCount
	if newCount > 0 {
		// Add new cards count (not limited here, session logic will handle the limit)
		availableForTraining += newCount
	}

	// Get total cards count
	totalQuery := `SELECT COUNT(*) FROM user_cards WHERE user_id = ?`
	var totalCards int
	err = r.db.QueryRow(totalQuery, userID).Scan(&totalCards)
	if err != nil {
		r.logger.Error("failed to get total cards count", zap.Error(err))
		totalCards = 0
	}

	// Get today's session count and stats
	today := now.Format("2006-01-02")
	var todaySessions int
	var todayCardsCompleted int
	todayQuery := `SELECT COUNT(*), COALESCE(SUM(done_count), 0) 
				   FROM training_sessions 
				   WHERE user_id = ? AND DATE(started_at) = ?`
	err = r.db.QueryRow(todayQuery, userID, today).Scan(&todaySessions, &todayCardsCompleted)
	if err != nil {
		r.logger.Error("failed to get today stats", zap.Error(err))
		todaySessions = 0
		todayCardsCompleted = 0
	}

	// Get week stats (last 7 days)
	weekAgo := now.AddDate(0, 0, -7)
	var weekSessions int
	var weekCardsCompleted int
	weekQuery := `SELECT COUNT(*), COALESCE(SUM(done_count), 0) 
				  FROM training_sessions 
				  WHERE user_id = ? AND started_at >= ? AND ended_at IS NOT NULL`
	err = r.db.QueryRow(weekQuery, userID, weekAgo).Scan(&weekSessions, &weekCardsCompleted)
	if err != nil {
		r.logger.Error("failed to get week stats", zap.Error(err))
		weekSessions = 0
		weekCardsCompleted = 0
	}

	// Get accuracy (last 30 days)
	monthAgo := now.AddDate(0, 0, -30)
	var totalReviews int
	var correctReviews int
	accuracyQuery := `SELECT 
		COUNT(*) as total,
		SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END) as correct
		FROM review_events 
		WHERE user_id = ? AND answered_at >= ? AND answered_at IS NOT NULL`
	err = r.db.QueryRow(accuracyQuery, userID, monthAgo).Scan(&totalReviews, &correctReviews)
	if err != nil {
		r.logger.Error("failed to get accuracy", zap.Error(err))
		totalReviews = 0
		correctReviews = 0
	}
	
	var accuracyPercent float64
	if totalReviews > 0 {
		accuracyPercent = float64(correctReviews) / float64(totalReviews) * 100
	}

	// Get weekly stats by day (last 7 days) with correct cards count
	weekAgoForDaily := now.AddDate(0, 0, -7)
	weeklyStatsQuery := `SELECT 
		DATE(ts.started_at) as day,
		COALESCE(COUNT(DISTINCT re.id), 0) as cards_completed,
		COALESCE(SUM(CASE WHEN re.is_correct = 1 THEN 1 ELSE 0 END), 0) as cards_correct
		FROM training_sessions ts
		LEFT JOIN review_events re ON re.session_id = ts.id AND re.answered_at IS NOT NULL
		WHERE ts.user_id = ? AND ts.started_at >= ? AND ts.ended_at IS NOT NULL AND ts.done_count > 0
		GROUP BY DATE(ts.started_at)
		ORDER BY day ASC`
	rows, err := r.db.Query(weeklyStatsQuery, userID, weekAgoForDaily)
	var weeklyStats []map[string]interface{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var day string
			var cardsCompleted, cardsCorrect int
			if err := rows.Scan(&day, &cardsCompleted, &cardsCorrect); err == nil {
				weeklyStats = append(weeklyStats, map[string]interface{}{
					"day":             day,
					"cards_completed": cardsCompleted,
					"cards_correct":   cardsCorrect,
				})
			}
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"due_count":             dueCount,
		"new_count":             newCount,
		"learning_count":        learningCount,
		"review_count":          reviewCount,
		"total_cards":           totalCards,
		"available_for_training": availableForTraining,
		"today_sessions":        todaySessions,
		"today_cards_completed": todayCardsCompleted,
		"week_sessions":         weekSessions,
		"week_cards_completed":  weekCardsCompleted,
		"accuracy_percent":      accuracyPercent,
		"weekly_stats":         weeklyStats,
	})
}

// handleChat handles AI chat requests
// @Summary      Отправить сообщение в AI чат
// @Description  Отправляет сообщение в AI чат и получает ответ от AI помощника для изучения языка. Если сообщение - одно слово, оно будет сохранено в БД и привязано к пользователю (как в телеграм-боте).
// @Tags         Chat
// @Accept       application/x-www-form-urlencoded
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        message  formData  string  true  "Текст сообщения для AI"
// @Success      200  {object}  map[string]interface{}  "Ответ от AI"
// @Failure      400  {string}  string  "Неверный запрос (отсутствует message)"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {object}  map[string]interface{}  "Ошибка при обработке сообщения"
// @Router       /app/chat [post]
func (r *Router) handleChat(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	message := req.FormValue("message")
	if message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	var response string
	var err error

	// Get WordService - need to properly type it
	type WordService interface {
		IsSingleWord(text string) bool
		GetWordDefinition(ctx context.Context, userID int64, word string) (string, error)
	}
	wordService, ok := r.wordService.(WordService)
	if !ok {
		r.logger.Error("Word service does not implement required interface")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if it's a single word - use word service (DB + AI, saves to DB)
	if wordService.IsSingleWord(message) {
		r.logger.Info("detected single word in chat",
			zap.String("word", message),
			zap.Int64("user_id", userID),
		)
		// Use word service which will:
		// 1. Check if word exists in DB
		// 2. If not, request from AI
		// 3. Save word to DB
		// 4. Add to word_request_history for this user
		response, err = wordService.GetWordDefinition(ctx, userID, message)
	} else {
		// Regular message - use AI service directly (no DB saving)
		type AIService interface {
			GenerateResponse(ctx context.Context, text string) (string, error)
		}
		aiService, ok := r.aiService.(AIService)
		if !ok {
			r.logger.Error("AI service does not implement GenerateResponse")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		response, err = aiService.GenerateResponse(ctx, message)
	}

	if err != nil {
		r.logger.Error("failed to generate response", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Sorry, an error occurred while processing your message. Please try again.",
		})
		return
	}

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response": response,
	})
}

