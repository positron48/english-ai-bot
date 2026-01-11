package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/models"
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
		
		// First, find word_card by word (from word_cards table)
		wordRepo := repository.NewWordRepository(r.db, r.logger)
		wordCard, err := wordRepo.GetWordCard(wordEN)
		if err != nil {
			r.logger.Error("failed to get word card", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		
		var cards []*models.TrainingCard
		if wordCard != nil {
			// Get training cards by word_card_id (more reliable than word_en)
			cards, err = trainingCardRepo.GetTrainingCardsByWordCardID(wordCard.ID)
			if err != nil {
				r.logger.Error("failed to get training cards", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		} else {
			// Word card not found, return empty cards
			cards = []*models.TrainingCard{}
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

// handleAdminTrainingCard handles individual training card operations (edit/delete)
// @Summary      Управление отдельной тренировочной карточкой
// @Description  Редактирование (PUT) или удаление (DELETE) отдельной тренировочной карточки по ID
// @Tags         Admin
// @Accept       json,application/x-www-form-urlencoded
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id  path  int64  true  "ID тренировочной карточки"
// @Param        word_ru  formData  string  false  "Русский перевод слова"
// @Param        meaning_en  formData  string  false  "Английское значение"
// @Param        example_en  formData  string  false  "Пример на английском"
// @Param        example_ru  formData  string  false  "Пример на русском"
// @Param        transcription  formData  string  false  "Транскрипция"
// @Param        distractors_ru  formData  string  false  "Отвлекающие варианты (RU, JSON array)"
// @Param        distractors_en  formData  string  false  "Отвлекающие варианты (EN, JSON array)"
// @Param        hint  formData  string  false  "Подсказка"
// @Success      200  {object}  map[string]interface{}  "Успешная операция"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен (требуются права администратора)"
// @Failure      404  {string}  string  "Карточка не найдена"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/admin/training/card/{id} [put]
// @Router       /app/admin/training/card/{id} [delete]
func (r *Router) handleAdminTrainingCard(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/app/admin/training/card/"), "/")
	
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Card ID is required", http.StatusBadRequest)
		return
	}

	var cardID int64
	if _, err := fmt.Sscanf(parts[0], "%d", &cardID); err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)

	if req.Method == http.MethodDelete {
		// Delete training card
		err := trainingCardRepo.DeleteTrainingCard(cardID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "Training card not found", http.StatusNotFound)
				return
			}
			r.logger.Error("failed to delete training card", zap.Error(err), zap.Int64("card_id", cardID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Training card deleted successfully",
			"card_id": cardID,
		})
		return
	}

	if req.Method == http.MethodPut {
		// Update training card
		// Support both JSON and form data
		contentType := req.Header.Get("Content-Type")
		var wordEN, wordRU, meaningEN, exampleEN, exampleRU, transcription, distractorsRU, distractorsEN, hint, pos, displayWord string
		posProvided := false // Track if pos was provided in the request

		if strings.Contains(contentType, "application/json") {
			// Parse JSON body
			var updateData map[string]interface{}
			if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
				http.Error(w, "Invalid JSON data", http.StatusBadRequest)
				return
			}

			if val, ok := updateData["word_en"].(string); ok {
				wordEN = val
			}
			if val, ok := updateData["pos"].(string); ok {
				pos = val
				posProvided = true
			}
			if val, ok := updateData["display_word"].(string); ok {
				displayWord = val
			}
			if val, ok := updateData["word_ru"].(string); ok {
				wordRU = val
			}
			if val, ok := updateData["meaning_en"].(string); ok {
				meaningEN = val
			}
			if val, ok := updateData["example_en"].(string); ok {
				exampleEN = val
			}
			if val, ok := updateData["example_ru"].(string); ok {
				exampleRU = val
			}
			if val, ok := updateData["transcription"].(string); ok {
				transcription = val
			}
			if val, ok := updateData["distractors_ru"].(string); ok {
				distractorsRU = val
			}
			if val, ok := updateData["distractors_en"].(string); ok {
				distractorsEN = val
			}
			if val, ok := updateData["hint"].(string); ok {
				hint = val
			}
		} else {
			// Parse form data (application/x-www-form-urlencoded or multipart/form-data)
			if err := req.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			wordEN = req.FormValue("word_en")
			if req.Form.Has("pos") {
				pos = req.FormValue("pos")
				posProvided = true
			}
			displayWord = req.FormValue("display_word")
			wordRU = req.FormValue("word_ru")
			meaningEN = req.FormValue("meaning_en")
			exampleEN = req.FormValue("example_en")
			exampleRU = req.FormValue("example_ru")
			transcription = req.FormValue("transcription")
			distractorsRU = req.FormValue("distractors_ru")
			distractorsEN = req.FormValue("distractors_en")
			hint = req.FormValue("hint")
		}

		// Get existing card
		card, err := trainingCardRepo.GetTrainingCard(cardID)
		if err != nil {
			r.logger.Error("failed to get training card", zap.Error(err), zap.Int64("card_id", cardID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if card == nil {
			http.Error(w, "Training card not found", http.StatusNotFound)
			return
		}

		// Update fields (always update, even if empty)
		if wordEN != "" {
			card.WordEN = wordEN
		}
		// Update POS only if it was provided in the request
		if posProvided {
			if pos != "" {
				card.POS = &pos
			} else {
				card.POS = nil
			}
		}
		if displayWord != "" {
			card.DisplayWord = &displayWord
		} else {
			card.DisplayWord = nil
		}
		card.WordRU = wordRU
		card.MeaningEN = meaningEN
		card.ExampleEN = exampleEN
		card.ExampleRU = exampleRU
		card.Transcription = transcription
		card.DistractorsRU = distractorsRU
		card.DistractorsEN = distractorsEN
		card.Hint = hint

		err = trainingCardRepo.UpdateTrainingCard(card)
		if err != nil {
			r.logger.Error("failed to update training card", zap.Error(err), zap.Int64("card_id", cardID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Training card updated successfully",
			"card_id": cardID,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleAdminWords handles word cards listing for admin
// @Summary      Получить список слов
// @Description  Возвращает список слов с возможностью фильтрации по пользователю и ошибкам
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        user_id      query     int64   false  "Фильтр по ID пользователя"
// @Param        only_errors  query     bool    false  "Показать только слова с ошибками"
// @Param        limit        query     int     false  "Лимит записей (по умолчанию 50)"
// @Param        offset       query     int     false  "Смещение (по умолчанию 0)"
// @Success      200  {object}  map[string]interface{}  "Список слов"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Router       /app/admin/words [get]
func (r *Router) handleAdminWords(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wordRepo := repository.NewWordRepository(r.db, r.logger)

	// Parse query parameters
	var filterUserID *int64
	if userIDStr := req.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filterUserID = &userID
		}
	}

	onlyWithErrors := req.URL.Query().Get("only_errors") == "1" || req.URL.Query().Get("only_errors") == "true"
	searchQuery := req.URL.Query().Get("search")

	limit := 50
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	words, err := wordRepo.ListWordCardsAdmin(filterUserID, onlyWithErrors, searchQuery, limit, offset)
	if err != nil {
		r.logger.Error("failed to list word cards", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get total count for pagination
	total, err := wordRepo.CountWordCardsAdmin(filterUserID, onlyWithErrors, searchQuery)
	if err != nil {
		r.logger.Error("failed to count word cards", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalPages := (total + limit - 1) / limit // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}
	currentPage := (offset / limit) + 1

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"words": words,
		"pagination": map[string]interface{}{
			"page":        currentPage,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// handleAdminWord handles individual word card operations (edit/reset/delete)
// @Summary      Управление отдельным словом
// @Description  Редактирование определения (PUT), сброс ошибки обработки (POST /reset) или удаление слова (DELETE)
// @Tags         Admin
// @Accept       json,application/x-www-form-urlencoded
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id          path      int64   true   "ID слова"
// @Param        definition  formData  string  false  "Новое определение"
// @Success      200  {object}  map[string]interface{}  "Успешная операция"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      404  {string}  string  "Слово не найдено"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/admin/words/{id} [put]
// @Router       /app/admin/words/{id} [delete]
// @Router       /app/admin/words/{id}/reset [post]
func (r *Router) handleAdminWord(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/app/admin/words/"), "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Word ID is required", http.StatusBadRequest)
		return
	}

	var wordCardID int64
	if _, err := fmt.Sscanf(parts[0], "%d", &wordCardID); err != nil {
		http.Error(w, "Invalid word ID", http.StatusBadRequest)
		return
	}

	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	wordRepo := repository.NewWordRepository(r.db, r.logger)

	if req.Method == http.MethodPut {
		// Update word card with all fields
		contentType := req.Header.Get("Content-Type")
		var word, definition, pos, transcription, definitionRU, examplesJSON, verbFormsJSON, displayEN string

		if strings.Contains(contentType, "application/json") {
			var updateData map[string]interface{}
			if err := json.NewDecoder(req.Body).Decode(&updateData); err != nil {
				http.Error(w, "Invalid JSON data", http.StatusBadRequest)
				return
			}
			if val, ok := updateData["word"].(string); ok {
				word = val
			}
			if val, ok := updateData["definition"].(string); ok {
				definition = val
			}
			if val, ok := updateData["pos"].(string); ok {
				pos = val
			}
			if val, ok := updateData["transcription"].(string); ok {
				transcription = val
			}
			if val, ok := updateData["definition_ru"].(string); ok {
				definitionRU = val
			}
			if val, ok := updateData["examples_json"].(string); ok {
				examplesJSON = val
			}
			if val, ok := updateData["verb_forms_json"].(string); ok {
				verbFormsJSON = val
			}
			if val, ok := updateData["display_en"].(string); ok {
				displayEN = val
			}
		} else {
			if err := req.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}
			word = req.FormValue("word")
			definition = req.FormValue("definition")
			pos = req.FormValue("pos")
			transcription = req.FormValue("transcription")
			definitionRU = req.FormValue("definition_ru")
			examplesJSON = req.FormValue("examples_json")
			verbFormsJSON = req.FormValue("verb_forms_json")
			displayEN = req.FormValue("display_en")
		}

		// Get existing card
		existingCard, err := wordRepo.GetWordCardByID(wordCardID)
		if err != nil {
			r.logger.Error("failed to get word card", zap.Error(err), zap.Int64("word_card_id", wordCardID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if existingCard == nil {
			http.Error(w, "Word card not found", http.StatusNotFound)
			return
		}

		// Update fields (use provided values or keep existing)
		if word == "" {
			word = existingCard.Word
		}
		card := &models.WordCard{
			ID:            wordCardID,
			Word:          word,
			Definition:    definition,
			POS:           &pos,
			Transcription: &transcription,
			DefinitionRU: &definitionRU,
			ExamplesJSON: &examplesJSON,
			VerbFormsJSON: &verbFormsJSON,
			DisplayEN:     &displayEN,
		}

		// Set to nil if empty strings
		if pos == "" {
			card.POS = nil
		}
		if transcription == "" {
			card.Transcription = nil
		}
		if definitionRU == "" {
			card.DefinitionRU = nil
		}
		if examplesJSON == "" {
			card.ExamplesJSON = nil
		}
		if verbFormsJSON == "" {
			card.VerbFormsJSON = nil
		}
		if displayEN == "" {
			card.DisplayEN = nil
		}

		err = wordRepo.UpdateWordCard(card)
		if err != nil {
			r.logger.Error("failed to update word card", zap.Error(err), zap.Int64("word_card_id", wordCardID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Word card updated successfully",
			"word_card_id": wordCardID,
		})
		return
	}

	if req.Method == http.MethodPost && action == "reset" {
		// Reset processed status
		err := wordRepo.ResetWordCardProcessed(wordCardID)
		if err != nil {
			r.logger.Error("failed to reset word card processed status", zap.Error(err), zap.Int64("word_card_id", wordCardID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Word card processed status reset successfully",
			"word_card_id": wordCardID,
		})
		return
	}

	if req.Method == http.MethodDelete {
		// Delete word card (CASCADE will delete related training_cards, user_cards, and word_request_history)
		err := wordRepo.DeleteWordCard(wordCardID)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "Word card not found", http.StatusNotFound)
				return
			}
			r.logger.Error("failed to delete word card", zap.Error(err), zap.Int64("word_card_id", wordCardID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Word card and all related data deleted successfully",
			"word_card_id": wordCardID,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleAdminUsers returns list of users for admin filtering
// @Summary      Получить список пользователей
// @Description  Возвращает список всех пользователей для фильтрации в админке
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Список пользователей"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Router       /app/admin/users [get]
func (r *Router) handleAdminUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userRepo := r.userRepo.(*repository.UserRepository)
	users, err := userRepo.GetAllUsers()
	if err != nil {
		r.logger.Error("failed to get all users", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return simplified user list (id, telegram_id, telegram_username)
	userList := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userList = append(userList, map[string]interface{}{
			"id":              user.ID,
			"telegram_id":     user.TelegramID,
			"telegram_username": user.TelegramUsername,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": userList,
	})
}

// handleAdminOrphanedCards handles orphaned training cards management
// @Summary      Получить список подвешенных тренировочных карточек
// @Description  Возвращает список training_cards, которые ссылаются на несуществующие word_cards
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        limit   query     int     false  "Лимит записей"  default(50)
// @Param        offset  query     int     false  "Смещение"       default(0)
// @Success      200     {object}  map[string]interface{}  "Список подвешенных карточек"
// @Failure      401     {string}  string  "Неавторизован"
// @Failure      403     {string}  string  "Доступ запрещен"
// @Router       /app/admin/orphaned-cards [get]
func (r *Router) handleAdminOrphanedCards(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)

		limit := 50
		if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		offset := 0
		if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		cards, err := trainingCardRepo.ListOrphanedTrainingCards(limit, offset)
		if err != nil {
			r.logger.Error("failed to list orphaned training cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		total, err := trainingCardRepo.CountOrphanedTrainingCards()
		if err != nil {
			r.logger.Error("failed to count orphaned training cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		totalPages := (total + limit - 1) / limit
		if totalPages == 0 {
			totalPages = 1
		}
		currentPage := (offset / limit) + 1

		// Convert to JSON-friendly format
		cardsList := make([]map[string]interface{}, 0, len(cards))
		for _, card := range cards {
			cardMap := map[string]interface{}{
				"id":                card.TrainingCardID,
				"word_card_id":      card.WordCardID,
				"word_en":           card.WordEN,
				"transcription":     card.Transcription,
				"sense_index":       card.SenseIndex,
				"word_ru":           card.WordRU,
				"meaning_en":        card.MeaningEN,
				"example_en":         card.ExampleEN,
				"example_ru":         card.ExampleRU,
				"pos":               card.POS,
				"display_word":      card.DisplayWord,
				"created_at":        card.CreatedAt.Format("2006-01-02 15:04:05"),
				"user_cards_count":  card.UserCardsCount,
			}
			cardsList = append(cardsList, cardMap)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"cards": cardsList,
			"pagination": map[string]interface{}{
				"page":        currentPage,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleAdminOrphanedCard handles individual orphaned card operations (delete)
// @Summary      Удалить подвешенную тренировочную карточку
// @Description  Удаляет training_card по ID (каскадно удаляет связанные user_cards и review_events)
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id   path      int     true  "ID тренировочной карточки"
// @Success      200  {object}  map[string]interface{}  "Успешное удаление"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      404  {string}  string  "Карточка не найдена"
// @Router       /app/admin/orphaned-cards/{id} [delete]
func (r *Router) handleAdminOrphanedCard(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(req.URL.Path, "/app/admin/orphaned-cards/")
	trainingCardID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "Invalid training card ID", http.StatusBadRequest)
		return
	}

	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)
	err = trainingCardRepo.DeleteTrainingCard(trainingCardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Training card not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to delete training card", zap.Error(err), zap.Int64("training_card_id", trainingCardID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Training card and all related data deleted successfully",
		"training_card_id": trainingCardID,
	})
}

// handleAdminOrphanedUserCards handles orphaned user cards management
// @Summary      Получить список подвешенных user карточек
// @Description  Возвращает список user_cards, которые ссылаются на несуществующие training_cards
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        limit   query     int     false  "Лимит записей (по умолчанию 50)"
// @Param        offset  query     int     false  "Смещение для пагинации (по умолчанию 0)"
// @Success      200     {object}  map[string]interface{}  "Список подвешенных user карточек"
// @Failure      401     {string}  string  "Неавторизован"
// @Failure      403     {string}  string  "Доступ запрещен"
// @Router       /app/admin/orphaned-user-cards [get]
func (r *Router) handleAdminOrphanedUserCards(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		userCardRepo := repository.NewUserCardRepository(r.db, r.logger)

		limit := 50
		if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		offset := 0
		if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		cards, err := userCardRepo.ListOrphanedUserCards(limit, offset)
		if err != nil {
			r.logger.Error("failed to list orphaned user cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		total, err := userCardRepo.CountOrphanedUserCards()
		if err != nil {
			r.logger.Error("failed to count orphaned user cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		totalPages := (total + limit - 1) / limit
		if totalPages == 0 {
			totalPages = 1
		}
		currentPage := (offset / limit) + 1

		// Convert to JSON-friendly format
		cardsList := make([]map[string]interface{}, 0, len(cards))
		for _, card := range cards {
			cardMap := map[string]interface{}{
				"user_card_id":        card.UserCardID,
				"user_id":             card.UserID,
				"telegram_id":         card.TelegramID,
				"telegram_username":   card.TelegramUsername,
				"training_card_id":    card.TrainingCardID,
				"direction":           card.Direction,
				"state":               card.State,
				"reps":                card.Reps,
				"created_at":          card.CreatedAt.Format("2006-01-02 15:04:05"),
				"review_events_count": card.ReviewEventsCount,
			}
			cardsList = append(cardsList, cardMap)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"cards": cardsList,
			"pagination": map[string]interface{}{
				"page":        currentPage,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleAdminOrphanedUserCard handles individual orphaned user card operations (delete)
// @Summary      Удалить подвешенную user карточку
// @Description  Удаляет user_card по ID (каскадно удаляет связанные review_events)
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id   path      int     true  "ID user карточки"
// @Success      200  {object}  map[string]interface{}  "Успешное удаление"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      404  {string}  string  "Карточка не найдена"
// @Router       /app/admin/orphaned-user-cards/{id} [delete]
func (r *Router) handleAdminOrphanedUserCard(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(req.URL.Path, "/app/admin/orphaned-user-cards/")
	userCardID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user card ID", http.StatusBadRequest)
		return
	}

	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)
	err = userCardRepo.DeleteUserCard(userCardID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User card not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to delete user card", zap.Error(err), zap.Int64("user_card_id", userCardID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "User card and all related data deleted successfully",
		"user_card_id": userCardID,
	})
}

