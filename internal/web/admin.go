package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// RequireAdmin wraps a handler to require admin access
func (r *Router) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID := getUserIDFromContext(req.Context())
		if userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get user's telegram_id
		userRepo := r.userRepo.(*repository.UserRepository)
		user, err := userRepo.GetUserByID(userID)
		if err != nil || user == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Check if user is admin
		if user.TelegramID != int64(r.config.Admin.TelegramID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next(w, req)
	}
}

// handleAdmin shows the admin panel
// @Summary      Получить данные админ-панели
// @Description  Возвращает состояние circuit breaker и информацию об администраторе (только для администраторов)
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Данные админ-панели"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен (требуются права администратора)"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Router       /app/admin [get]
func (r *Router) handleAdmin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get circuit breaker status directly from DB (same approach as dashboard sessions)
	// Query directly to get dates as strings, avoiding time.Time parsing issues
	query := `SELECT id, is_open, failure_count, 
			  COALESCE(last_failure_at, '') as last_failure_at,
			  COALESCE(last_failure_message, '') as last_failure_message,
			  COALESCE(last_reset_at, '') as last_reset_at
			  FROM circuit_breaker_state WHERE id = 1`
	
	var cbID int64
	var isOpen bool
	var failureCount int
	var lastFailureAt, lastFailureMessage, lastResetAt string
	
	err := r.db.QueryRow(query).Scan(&cbID, &isOpen, &failureCount, &lastFailureAt, &lastFailureMessage, &lastResetAt)
	
	var cbResponse map[string]interface{}
	if err == sql.ErrNoRows {
		// Initialize if not exists
		initQuery := `INSERT OR IGNORE INTO circuit_breaker_state (id) VALUES (1)`
		_, initErr := r.db.Exec(initQuery)
		if initErr != nil {
			r.logger.Error("failed to initialize circuit breaker state", zap.Error(initErr))
		}
		// Retry query after initialization
		err = r.db.QueryRow(query).Scan(&cbID, &isOpen, &failureCount, &lastFailureAt, &lastFailureMessage, &lastResetAt)
	}
	
	if err == nil {
		state := "closed"
		if isOpen {
			state = "open"
		}
		cbResponse = map[string]interface{}{
			"state":             state,
			"is_open":           isOpen,
			"failures":          failureCount,
			"last_failure":       nil,
			"last_failure_at":   nil,
			"last_reset_at":     nil,
		}
		// Return dates as strings directly from SQL (same format as dashboard sessions)
		if lastFailureAt != "" {
			cbResponse["last_failure_at"] = lastFailureAt
		}
		if lastResetAt != "" {
			cbResponse["last_reset_at"] = lastResetAt
		}
		if lastFailureMessage != "" {
			cbResponse["last_failure"] = lastFailureMessage
		}
	} else {
		r.logger.Error("failed to get circuit breaker state", zap.Error(err))
		// Default state if error
		cbResponse = map[string]interface{}{
			"state":   "closed",
			"is_open": false,
			"failures": 0,
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"circuit_breaker": cbResponse,
		"admin_id":        r.config.Admin.TelegramID,
	})
}

// handleAdminCircuitReset resets the circuit breaker
// @Summary      Сбросить circuit breaker
// @Description  Сбрасывает состояние circuit breaker (только для администраторов)
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Успешный сброс"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен (требуются права администратора)"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/admin/circuit/reset [post]
func (r *Router) handleAdminCircuitReset(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.cbService.Reset(); err != nil {
		r.logger.Error("failed to reset circuit breaker", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Circuit breaker reset successfully",
	})
}

// handleAdminTraining handles training card management
// @Summary      Управление тренировочными карточками
// @Description  Получение (GET) или удаление (POST) тренировочных карточек по слову. Путь: /app/admin/training/{word} или /app/admin/training/{word}/delete или /app/admin/training/delete_all
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        word    path      string  false  "Английское слово"
// @Param        action  path      string  false  "Действие: delete или delete_all"
// @Param        word    query     string  false  "Английское слово (для GET запроса)"
// @Success      200  {object}  map[string]interface{}  "Данные карточек или результат удаления"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен (требуются права администратора)"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/admin/training/{word} [get]
func (r *Router) handleAdminTraining(w http.ResponseWriter, req *http.Request) {
	// Extract action and word from path: /app/admin/training/{word}/{action}
	path := req.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/app/admin/training/"), "/")
	
	if len(parts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	wordEN := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)

	if req.Method == http.MethodGet && wordEN != "" && action == "" {
		// Get training data for word
		// Extract word from query parameter if path doesn't have it
		if wordEN == "" {
			wordEN = req.URL.Query().Get("word")
		}
		if wordEN == "" {
			http.Error(w, "word is required", http.StatusBadRequest)
			return
		}
		
		cards, err := trainingCardRepo.GetTrainingCardsByWordEN(wordEN)
		if err != nil {
			r.logger.Error("failed to get training cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"word_en": wordEN,
			"cards":   cards,
		})
		return
	}

	if req.Method == http.MethodPost && action == "delete" {
		// Delete training cards by word
		// Get word from form or path
		if wordEN == "" {
			wordEN = req.FormValue("word")
		}
		if wordEN == "" {
			http.Error(w, "word is required", http.StatusBadRequest)
			return
		}
		
		rowsAffected, err := trainingCardRepo.DeleteTrainingCardsByWordEN(wordEN)
		if err != nil {
			r.logger.Error("failed to delete training cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"word_en":       wordEN,
			"rows_affected": rowsAffected,
		})
		return
	}

	if req.Method == http.MethodPost && wordEN == "delete_all" {
		// Delete all training cards
		rowsAffected, err := trainingCardRepo.DeleteAllTrainingCards()
		if err != nil {
			r.logger.Error("failed to delete all training cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"rows_affected": rowsAffected,
		})
		return
	}

	http.Error(w, "Invalid request", http.StatusBadRequest)
}

