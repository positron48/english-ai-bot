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
// @Description  Возвращает количество карточек, готовых к повторению (due_count)
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

	// Calculate available cards for training (same logic as training service)
	// MaxCardsPerSession = 30, MaxNewPerSession = 5
	maxCardsPerSession := 30
	maxNewPerSession := 5
	
	availableForTraining := dueCount
	remainingSlots := maxCardsPerSession - dueCount
	if remainingSlots > 0 {
		maxNew := maxNewPerSession
		if remainingSlots < maxNew {
			maxNew = remainingSlots
		}
		if newCount > maxNew {
			availableForTraining += maxNew
		} else {
			availableForTraining += newCount
		}
	}
	// Cap at maxCardsPerSession
	if availableForTraining > maxCardsPerSession {
		availableForTraining = maxCardsPerSession
	}

	// Get total cards count
	totalQuery := `SELECT COUNT(*) FROM user_cards WHERE user_id = ?`
	var totalCards int
	err = r.db.QueryRow(totalQuery, userID).Scan(&totalCards)
	if err != nil {
		r.logger.Error("failed to get total cards count", zap.Error(err))
		totalCards = 0
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"due_count":            dueCount,
		"total_cards":          totalCards,
		"available_for_training": availableForTraining,
	})
}

// handleChat handles AI chat requests
// @Summary      Отправить сообщение в AI чат
// @Description  Отправляет сообщение в AI чат и получает ответ от AI помощника для изучения языка
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

	// Get AI service - need to properly type it
	// For now, access via interface assertion
	type AIService interface {
		GenerateResponse(ctx context.Context, text string) (string, error)
	}
	aiService, ok := r.aiService.(AIService)
	if !ok {
		r.logger.Error("AI service does not implement GenerateResponse")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate response
	ctx := req.Context()
	response, err := aiService.GenerateResponse(ctx, message)
	if err != nil {
		r.logger.Error("failed to generate AI response", zap.Error(err))
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

