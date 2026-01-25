package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

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
// @Router       /api/dashboard [get]
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
	
	// Get new cards count (exclude orphaned cards - those with non-existent training_cards or word_cards)
	// Excludes words marked as "known" in user_word_knowledge (same as GetNewCards)
	newQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.state = 'new'
		AND NOT EXISTS (
			SELECT 1 FROM user_word_knowledge uwk 
			WHERE uwk.user_id = ? AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
		)`
	var newCount int
	err := r.db.QueryRow(newQuery, userID, userID).Scan(&newCount)
	if err != nil {
		r.logger.Error("failed to get new cards count", zap.Error(err))
		newCount = 0
	}

	// Get due count (cards ready for review, excluding new cards and orphaned cards)
	// Excludes words marked as "known" in user_word_knowledge (same as GetDueCards)
	// Note: GetDueCards doesn't filter by state != 'new', but we do here for clarity
	dueQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.state != 'new' AND (uc.next_due_at IS NULL OR uc.next_due_at <= ?)
		AND NOT EXISTS (
			SELECT 1 FROM user_word_knowledge uwk 
			WHERE uwk.user_id = ? AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
		)`
	var dueCount int
	err = r.db.QueryRow(dueQuery, userID, now, userID).Scan(&dueCount)
	if err != nil {
		r.logger.Error("failed to get due count", zap.Error(err))
		dueCount = 0
	}

	// Get learning cards count (exclude orphaned cards)
	learningQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.state = 'learning'`
	var learningCount int
	err = r.db.QueryRow(learningQuery, userID).Scan(&learningCount)
	if err != nil {
		r.logger.Error("failed to get learning count", zap.Error(err))
		learningCount = 0
	}

	// Get review cards count (exclude orphaned cards)
	reviewQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.state = 'review'`
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

	// Get total cards count (exclude orphaned cards)
	totalQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ?`
	var totalCards int
	err = r.db.QueryRow(totalQuery, userID).Scan(&totalCards)
	if err != nil {
		r.logger.Error("failed to get total cards count", zap.Error(err))
		totalCards = 0
	}

	// Get accuracy (last 30 days)
	monthAgo := now.AddDate(0, 0, -30)
	var totalReviews int
	var correctReviews int
	accuracyQuery := `SELECT 
		COUNT(*) as total,
		COALESCE(SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END), 0) as correct
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

	// Get words added stats by day (last 7 days, exclude orphaned cards)
	wordsAddedStatsQuery := `SELECT 
		DATE(uc.created_at) as day,
		COUNT(*) as words_added
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.created_at >= ?
		GROUP BY DATE(uc.created_at)
		ORDER BY day ASC`
	wordsRows, err := r.db.Query(wordsAddedStatsQuery, userID, weekAgoForDaily)
	var wordsAddedStats []map[string]interface{}
	if err == nil {
		defer wordsRows.Close()
		for wordsRows.Next() {
			var day string
			var wordsAdded int
			if err := wordsRows.Scan(&day, &wordsAdded); err == nil {
				wordsAddedStats = append(wordsAddedStats, map[string]interface{}{
					"day":          day,
					"words_added": wordsAdded,
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
		"accuracy_percent":      accuracyPercent,
		"weekly_stats":         weeklyStats,
		"words_added_stats":     wordsAddedStats,
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
// @Router       /api/chat [post]
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

	// Get AI service interface for both cases
	type AIService interface {
		GenerateResponse(ctx context.Context, text string) (string, error)
	}
	aiService, ok := r.aiService.(AIService)
	if !ok {
		r.logger.Error("AI service does not implement GenerateResponse")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if it's a single word - use word service (DB + AI, saves to DB)
	// BUT: if single word contains Cyrillic, send to AI directly (don't save to DB)
	if wordService.IsSingleWord(message) {
		// Check if word contains Cyrillic characters
		containsCyrillic := false
		for _, r := range message {
			if unicode.Is(unicode.Cyrillic, r) {
				containsCyrillic = true
				break
			}
		}

		if containsCyrillic {
			// Russian word - send to AI directly, don't save to DB
			r.logger.Info("detected single Russian word in chat, sending to AI directly",
				zap.String("word", message),
				zap.Int64("user_id", userID),
			)
			response, err = aiService.GenerateResponse(ctx, message)
		} else {
			// English word - use word service (DB + AI, saves to DB)
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
		}
	} else {
		// Regular message - use AI service directly (no DB saving)
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

// handleSettings handles GET /api/settings - get user settings
// @Summary      Получить настройки пользователя
// @Description  Возвращает текущие настройки пользователя
// @Tags         Settings
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Настройки пользователя"
// @Failure      401  {string}  string  "Неавторизован"
// @Router       /api/settings [get]
func (r *Router) handleSettings(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetUserByID(userID)
	if err != nil {
		r.logger.Error("failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Parse settings
	var settings models.UserSettings
	if user.SettingsJSON != "" {
		if err := json.Unmarshal([]byte(user.SettingsJSON), &settings); err != nil {
			r.logger.Warn("failed to parse user settings", zap.Error(err))
		}
	}

	// Set defaults if not set
	if settings.NotificationFrequency == "" {
		settings.NotificationFrequency = "daily"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"settings": settings,
	})
}

// handleNotificationSettings handles POST /api/settings/notifications - update notification settings
// @Summary      Обновить настройки уведомлений
// @Description  Обновляет периодичность уведомлений пользователя
// @Tags         Settings
// @Accept       application/json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        frequency  body  string  true  "Периодичность: 'daily', 'never', или число дней (например, '3')"
// @Success      200  {object}  map[string]interface{}  "Успешное обновление"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Router       /api/settings/notifications [post]
func (r *Router) handleNotificationSettings(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var requestData struct {
		Frequency string `json:"frequency"`
	}

	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	frequency := strings.TrimSpace(requestData.Frequency)
	if frequency == "" {
		http.Error(w, "frequency is required", http.StatusBadRequest)
		return
	}

	// Validate frequency
	frequencyLower := strings.ToLower(frequency)
	if frequencyLower != "daily" && frequencyLower != "never" {
		// Try to parse as number
		days, err := strconv.Atoi(frequency)
		if err != nil || days < 1 {
			http.Error(w, "frequency must be 'daily', 'never', or a positive number", http.StatusBadRequest)
			return
		}
		frequency = strconv.Itoa(days)
	} else {
		frequency = frequencyLower
	}

	userRepo := r.userRepo.(*repository.UserRepository)
	user, err := userRepo.GetUserByID(userID)
	if err != nil {
		r.logger.Error("failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Parse current settings
	var settings models.UserSettings
	if user.SettingsJSON != "" {
		if err := json.Unmarshal([]byte(user.SettingsJSON), &settings); err != nil {
			r.logger.Warn("failed to parse user settings", zap.Error(err))
		}
	}

	// Update notification frequency
	settings.NotificationFrequency = frequency

	// Save settings
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		r.logger.Error("failed to marshal settings", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		r.logger.Error("failed to update user settings", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Notification settings updated successfully",
		"frequency": frequency,
	})
}

