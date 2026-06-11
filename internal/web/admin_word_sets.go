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

func adminCourseCode(req *http.Request, bodyCourseCode string) string {
	courseCode := strings.TrimSpace(strings.ToLower(bodyCourseCode))
	if courseCode == "" {
		courseCode = strings.TrimSpace(strings.ToLower(req.URL.Query().Get("course_code")))
	}
	return courseCode
}

func (r *Router) validateAdminCourseCode(courseCode string) error {
	if courseCode == "" {
		return nil
	}
	var exists int
	err := r.db.QueryRow(`SELECT 1 FROM courses WHERE code = ? AND status = 'active'`, courseCode).Scan(&exists)
	if err == sql.ErrNoRows {
		return repository.ErrCourseNotFound
	}
	return err
}

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
// @Router       /api/admin/word-set-categories [get]
// @Router       /api/admin/word-set-categories [post]
// @Router       /api/admin/word-set-categories/{id} [put]
// @Router       /api/admin/word-set-categories/{id} [delete]
func (r *Router) handleAdminWordSetCategories(w http.ResponseWriter, req *http.Request) {
	categoryRepo := repository.NewWordSetCategoryRepository(r.db, r.logger)

	switch req.Method {
	case http.MethodGet:
		courseCode := adminCourseCode(req, "")
		if err := r.validateAdminCourseCode(courseCode); err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		categories, err := categoryRepo.GetAllCategoriesForCourse(courseCode)
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
		// Check edit permission for POST
		ctx := req.Context()
		if !r.checkPermissionInHandler(ctx, PermissionWordSetsEdit) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Forbidden",
				"message": "You don't have permission to create word set categories.",
			})
			return
		}

		var category models.WordSetCategory
		if err := json.NewDecoder(req.Body).Decode(&category); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if category.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		category.CourseCode = adminCourseCode(req, category.CourseCode)
		if err := r.validateAdminCourseCode(category.CourseCode); err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		if category.ParentID != nil && category.CourseCode != "" {
			parent, err := categoryRepo.GetCategoryForCourse(*category.ParentID, category.CourseCode)
			if err != nil || parent == nil || parent.CourseCode != category.CourseCode {
				http.Error(w, "Parent category must belong to the selected course", http.StatusBadRequest)
				return
			}
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
		// Check edit permission for PUT
		ctx := req.Context()
		if !r.checkPermissionInHandler(ctx, PermissionWordSetsEdit) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Forbidden",
				"message": "You don't have permission to edit word set categories.",
			})
			return
		}

		// Extract ID from path
		path := strings.TrimPrefix(req.URL.Path, "/api/admin/word-set-categories/")
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
		category.CourseCode = adminCourseCode(req, category.CourseCode)
		if err := r.validateAdminCourseCode(category.CourseCode); err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		if category.ParentID != nil && category.CourseCode != "" {
			parent, err := categoryRepo.GetCategoryForCourse(*category.ParentID, category.CourseCode)
			if err != nil || parent == nil || parent.CourseCode != category.CourseCode {
				http.Error(w, "Parent category must belong to the selected course", http.StatusBadRequest)
				return
			}
		}
		var updateErr error
		if category.CourseCode == "" {
			updateErr = categoryRepo.UpdateCategory(&category)
		} else {
			updateErr = categoryRepo.UpdateCategoryForCourse(&category)
		}
		if updateErr != nil {
			r.logger.Error("failed to update category", zap.Error(updateErr))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	case http.MethodDelete:
		// Check edit permission for DELETE
		ctx := req.Context()
		if !r.checkPermissionInHandler(ctx, PermissionWordSetsEdit) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Forbidden",
				"message": "You don't have permission to delete word set categories.",
			})
			return
		}

		// Extract ID from path
		path := strings.TrimPrefix(req.URL.Path, "/api/admin/word-set-categories/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		courseCode := adminCourseCode(req, "")
		if err := r.validateAdminCourseCode(courseCode); err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		if err := categoryRepo.DeleteCategoryForCourse(id, courseCode); err != nil {
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
// @Router       /api/admin/word-sets [get]
// @Router       /api/admin/word-sets [post]
// @Router       /api/admin/word-sets/{id} [put]
// @Router       /api/admin/word-sets/{id} [delete]
func (r *Router) handleAdminWordSets(w http.ResponseWriter, req *http.Request) {
	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)

	switch req.Method {
	case http.MethodGet:
		courseCode := adminCourseCode(req, "")
		if err := r.validateAdminCourseCode(courseCode); err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
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
		wordSets, err := wordSetRepo.ListWordSetsForCourse(courseCode, categoryID, limit, offset, true)
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
		// Check edit permission for POST
		ctx := req.Context()
		if !r.checkPermissionInHandler(ctx, PermissionWordSetsEdit) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Forbidden",
				"message": "You don't have permission to create word sets.",
			})
			return
		}

		var wordSet models.WordSet
		if err := json.NewDecoder(req.Body).Decode(&wordSet); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if wordSet.Title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}
		wordSet.CourseCode = adminCourseCode(req, wordSet.CourseCode)
		if err := r.validateAdminCourseCode(wordSet.CourseCode); err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		if wordSet.CategoryID != nil && wordSet.CourseCode != "" {
			category, err := repository.NewWordSetCategoryRepository(r.db, r.logger).GetCategoryForCourse(*wordSet.CategoryID, wordSet.CourseCode)
			if err != nil || category == nil || category.CourseCode != wordSet.CourseCode {
				http.Error(w, "Category must belong to the selected course", http.StatusBadRequest)
				return
			}
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
		// Check edit permission for PUT
		ctx := req.Context()
		if !r.checkPermissionInHandler(ctx, PermissionWordSetsEdit) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Forbidden",
				"message": "You don't have permission to edit word sets.",
			})
			return
		}

		// Extract ID from path
		path := strings.TrimPrefix(req.URL.Path, "/api/admin/word-sets/")
		parts := strings.Split(path, "/")

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

			courseCode := adminCourseCode(req, "")
			if err := r.validateAdminCourseCode(courseCode); err != nil {
				http.Error(w, "Course not found", http.StatusNotFound)
				return
			}
			if courseCode != "" && courseCode != r.defaultCourseCode() {
				http.Error(w, "Word generation for the selected course is not available in this runtime", http.StatusConflict)
				return
			}
			wordSetService := r.getWordSetService()
			if err := wordSetService.ProcessWordSetItemsForCourse(req.Context(), id, courseCode, requestData.Words); err != nil {
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
		wordSet.CourseCode = adminCourseCode(req, wordSet.CourseCode)
		if err := r.validateAdminCourseCode(wordSet.CourseCode); err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		if wordSet.CategoryID != nil && wordSet.CourseCode != "" {
			category, err := repository.NewWordSetCategoryRepository(r.db, r.logger).GetCategoryForCourse(*wordSet.CategoryID, wordSet.CourseCode)
			if err != nil || category == nil || category.CourseCode != wordSet.CourseCode {
				http.Error(w, "Category must belong to the selected course", http.StatusBadRequest)
				return
			}
		}
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
		// Check edit permission for DELETE
		ctx := req.Context()
		if !r.checkPermissionInHandler(ctx, PermissionWordSetsEdit) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Forbidden",
				"message": "You don't have permission to delete word sets.",
			})
			return
		}

		// Extract ID from path
		path := strings.TrimPrefix(req.URL.Path, "/api/admin/word-sets/")
		id, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			http.Error(w, "Invalid word set ID", http.StatusBadRequest)
			return
		}

		courseCode := adminCourseCode(req, "")
		if err := r.validateAdminCourseCode(courseCode); err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		if err := wordSetRepo.DeleteWordSetForCourse(id, courseCode); err != nil {
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
	path := strings.TrimPrefix(req.URL.Path, "/api/admin/word-sets/")
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
// @Router       /api/admin/word-sets/{id} [get]
func (r *Router) handleAdminWordSetDetail(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(req.URL.Path, "/api/admin/word-sets/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "Invalid word set ID", http.StatusBadRequest)
		return
	}

	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)

	// Get word set
	courseCode := adminCourseCode(req, "")
	if err := r.validateAdminCourseCode(courseCode); err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	wordSet, err := wordSetRepo.GetWordSetForCourse(id, courseCode)
	if err != nil {
		r.logger.Error("failed to get word set", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if wordSet == nil {
		http.Error(w, "Word set not found", http.StatusNotFound)
		return
	}

	// Get words in set (using repository method to respect preferred_pos filtering)
	// For admin, we use userID = 0 (no user-specific filtering needed)
	wordsWithStatus, err := wordSetRepo.GetWordSetWords(id, 0)
	if err != nil {
		r.logger.Error("failed to get word set words", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type WordInfo struct {
		WordCardID  int64  `json:"word_card_id"`
		Word        string `json:"word"`
		DisplayWord string `json:"display_word"`
	}

	words := make([]WordInfo, 0, len(wordsWithStatus))
	for _, w := range wordsWithStatus {
		words = append(words, WordInfo{
			WordCardID:  w.WordCardID,
			Word:        w.Word,
			DisplayWord: w.DisplayWord,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"word_set": wordSet,
		"words":    words,
	})
}
