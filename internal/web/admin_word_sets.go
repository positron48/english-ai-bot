package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// handleAdminWordSetCategories handles CRUD for word set categories
// @Summary      Управление категориями наборов слов
// @Description  CRUD операции для категорий наборов слов (GET список, POST создать, PUT обновить, DELETE удалить)
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id  path  int  false  "ID категории (для PUT/DELETE)"
// @Success      200  {object}  map[string]interface{}  "Список категорий или результат операции"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/admin/word-set-categories [get]
// @Router       /app/admin/word-set-categories [post]
// @Router       /app/admin/word-set-categories/{id} [put]
// @Router       /app/admin/word-set-categories/{id} [delete]
func (r *Router) handleAdminWordSetCategories(w http.ResponseWriter, req *http.Request) {
	categoryRepo := repository.NewWordSetCategoryRepository(r.db, r.logger)

	switch req.Method {
	case http.MethodGet:
		categories, err := categoryRepo.GetAllCategories()
		if err != nil {
			r.logger.Error("failed to get categories", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"categories": categories,
		})

	case http.MethodPost:
		var category models.WordSetCategory
		if err := json.NewDecoder(req.Body).Decode(&category); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if category.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		// Log received data for debugging
		r.logger.Debug("creating category",
			zap.String("name", category.Name),
			zap.Int("sort_order", category.SortOrder),
			zap.Any("parent_id", category.ParentID),
		)

		id, err := categoryRepo.CreateCategory(&category)
		if err != nil {
			r.logger.Error("failed to create category", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
		})

	case http.MethodPut:
		// Extract ID from path
		path := strings.TrimPrefix(req.URL.Path, "/app/admin/word-set-categories/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		var category models.WordSetCategory
		if err := json.NewDecoder(req.Body).Decode(&category); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Log received data for debugging
		r.logger.Debug("updating category",
			zap.Int64("id", id),
			zap.String("name", category.Name),
			zap.Int("sort_order", category.SortOrder),
			zap.Any("parent_id", category.ParentID),
		)

		category.ID = id
		if err := categoryRepo.UpdateCategory(&category); err != nil {
			r.logger.Error("failed to update category", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	case http.MethodDelete:
		// Extract ID from path
		path := strings.TrimPrefix(req.URL.Path, "/app/admin/word-set-categories/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		if err := categoryRepo.DeleteCategory(id); err != nil {
			r.logger.Error("failed to delete category", zap.Error(err))
			if strings.Contains(err.Error(), "cannot delete") {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminWordSets handles CRUD for word sets
// @Summary      Управление наборами слов
// @Description  CRUD операции для наборов слов (GET список, POST создать, PUT обновить, DELETE удалить)
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id  path  int  false  "ID набора (для PUT/DELETE)"
// @Success      200  {object}  map[string]interface{}  "Список наборов или результат операции"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/admin/word-sets [get]
// @Router       /app/admin/word-sets [post]
// @Router       /app/admin/word-sets/{id} [put]
// @Router       /app/admin/word-sets/{id} [delete]
func (r *Router) handleAdminWordSets(w http.ResponseWriter, req *http.Request) {
	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)

	switch req.Method {
	case http.MethodGet:
		// Parse query parameters
		var categoryID *int64
		if catIDStr := req.URL.Query().Get("category_id"); catIDStr != "" {
			if catID, err := strconv.ParseInt(catIDStr, 10, 64); err == nil {
				categoryID = &catID
			}
		}

		limit := 100
		if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
				limit = l
			}
		}

		offset := 0
		if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		// Get all word sets (including unpublished for admin)
		wordSets, err := wordSetRepo.ListWordSets(categoryID, limit, offset, true)
		if err != nil {
			r.logger.Error("failed to list word sets", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"word_sets": wordSets,
		})

	case http.MethodPost:
		var wordSet models.WordSet
		if err := json.NewDecoder(req.Body).Decode(&wordSet); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if wordSet.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		id, err := wordSetRepo.CreateWordSet(&wordSet)
		if err != nil {
			r.logger.Error("failed to create word set", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      id,
		})

	case http.MethodPut:
		// Extract ID from path
		path := strings.TrimPrefix(req.URL.Path, "/app/admin/word-sets/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid word set ID", http.StatusBadRequest)
			return
		}

		// Check if it's items update
		if len(parts) > 1 && parts[1] == "items" {
			var requestData struct {
				Words string `json:"words"` // Comma-separated
			}
			if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			wordSetService := r.getWordSetService()
			if err := wordSetService.ProcessWordSetItems(req.Context(), id, requestData.Words); err != nil {
				r.logger.Error("failed to process word set items", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
			})
			return
		}

		// Regular update
		var wordSet models.WordSet
		if err := json.NewDecoder(req.Body).Decode(&wordSet); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		wordSet.ID = id
		if err := wordSetRepo.UpdateWordSet(&wordSet); err != nil {
			r.logger.Error("failed to update word set", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	case http.MethodDelete:
		// Extract ID from path
		path := strings.TrimPrefix(req.URL.Path, "/app/admin/word-sets/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid word set ID", http.StatusBadRequest)
			return
		}

		if err := wordSetRepo.DeleteWordSet(id); err != nil {
			r.logger.Error("failed to delete word set", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminWordSetDetailOrSets routes to detail or sets handlers
func (r *Router) handleAdminWordSetDetailOrSets(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/app/admin/word-sets/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 1 || parts[0] == "" {
		// No ID, should be handled by handleAdminWordSets
		r.handleAdminWordSets(w, req)
		return
	}
	
	// PUT and DELETE requests should go to handleAdminWordSets
	if req.Method == http.MethodPut || req.Method == http.MethodDelete {
		r.handleAdminWordSets(w, req)
		return
	}
	
	// Check if it's items update
	if len(parts) >= 2 && parts[1] == "items" {
		r.handleAdminWordSets(w, req)
		return
	}
	
	// Default to detail view (GET requests)
	r.handleAdminWordSetDetail(w, req)
}

// handleAdminWordSetDetail returns word set details for admin
// @Summary      Получить детали набора слов (админ)
// @Description  Возвращает детальную информацию о наборе слов для админа
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id  path  int  true  "ID набора слов"
// @Success      200  {object}  map[string]interface{}  "Детали набора"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      404  {string}  string  "Набор не найден"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/admin/word-sets/{id} [get]
func (r *Router) handleAdminWordSetDetail(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(req.URL.Path, "/app/admin/word-sets/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "Invalid word set ID", http.StatusBadRequest)
		return
	}

	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)

	// Get word set
	wordSet, err := wordSetRepo.GetWordSet(id)
	if err != nil {
		r.logger.Error("failed to get word set", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if wordSet == nil {
		http.Error(w, "Word set not found", http.StatusNotFound)
		return
	}

	// Get words in set
	query := `SELECT wc.id, wc.word, 
		COALESCE(
			(SELECT display_word FROM training_cards tc2 WHERE tc2.word_card_id = wc.id AND tc2.display_word IS NOT NULL AND tc2.display_word != '' LIMIT 1),
			wc.display_en,
			wc.word
		) as display_word
		FROM word_set_items wsi
		INNER JOIN word_cards wc ON wsi.word_card_id = wc.id
		WHERE wsi.word_set_id = ?
		ORDER BY wsi.sort_order, wc.word`

	rows, err := r.db.Query(query, id)
	if err != nil {
		r.logger.Error("failed to get word set words", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type WordInfo struct {
		WordCardID  int64  `json:"word_card_id"`
		Word        string `json:"word"`
		DisplayWord string `json:"display_word"`
	}

	var words []WordInfo
	for rows.Next() {
		var word WordInfo
		var displayWord sql.NullString

		if err := rows.Scan(&word.WordCardID, &word.Word, &displayWord); err != nil {
			r.logger.Warn("failed to scan word", zap.Error(err))
			continue
		}

		if displayWord.Valid {
			word.DisplayWord = displayWord.String
		} else {
			word.DisplayWord = word.Word
		}

		words = append(words, word)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"word_set": wordSet,
		"words":    words,
	})
}
