package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

// handleLearningGrammarCategories returns published grammar categories
// @Summary      Получить опубликованные категории грамматики
// @Description  Возвращает список опубликованных категорий с прогрессом пользователя
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Список категорий"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/categories [get]
func (r *Router) handleLearningGrammarCategories(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sections, err := r.grammarService.GetAllSectionsWithProgress(req.Context(), userID)
	if err != nil {
		r.logger.Error("failed to get grammar categories", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type CategoryResponse struct {
		SectionID          string            `json:"section_id"`
		Title              string            `json:"title"`
		TitleTranslations  map[string]string `json:"title_translations,omitempty"`
		Level              string            `json:"level"`
		Order              int               `json:"order"`
		IsPublished        bool              `json:"is_published"`
		PublishedChapters  int               `json:"published_chapters"`
		PassedChapters     int               `json:"passed_chapters"`
		TotalChapters      int               `json:"total_chapters"`
		ProgressPercentage int               `json:"progress_percentage"`
		CanAccess          bool              `json:"can_access"`
		CategoryTestScore  *int              `json:"category_test_score,omitempty"`
	}

	categories := make([]CategoryResponse, 0, len(sections))
	for _, section := range sections {
		canAccess := false
		if section.IsPublished {
			var errAccess error
			canAccess, errAccess = r.grammarService.CanAccessSection(req.Context(), userID, section.Section.SectionID)
			if errAccess != nil {
				r.logger.Warn("failed to check section access, defaulting to false", zap.String("section_id", section.Section.SectionID), zap.Error(errAccess))
				canAccess = false
			}
		}

		var categoryTestScore *int
		if section.IsPublished {
			bestScore, errScore := r.grammarService.AttemptRepo.GetCategoryTestBestScore(userID, section.Section.SectionID)
			if errScore == nil && bestScore > 0 {
				categoryTestScore = &bestScore
			}
		}

		categories = append(categories, CategoryResponse{
			SectionID:          section.Section.SectionID,
			Title:              section.Title,
			TitleTranslations:  section.Section.TitleTranslations,
			Level:              section.Section.Level,
			Order:              section.Section.Order,
			IsPublished:        section.IsPublished,
			PublishedChapters:  section.PublishedChapters,
			PassedChapters:     section.PassedChapters,
			TotalChapters:      len(section.Section.ChapterIDs),
			ProgressPercentage: section.ProgressPercentage,
			CanAccess:          canAccess,
			CategoryTestScore:  categoryTestScore,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
	})
}

// handleLearningGrammarChapters returns published chapters for a section
// @Summary      Получить опубликованные главы категории
// @Description  Возвращает список опубликованных глав категории с прогрессом пользователя
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        section_id  path  string  true  "ID категории"
// @Success      200  {object}  map[string]interface{}  "Список глав"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Категория не найдена"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/categories/{section_id}/chapters [get]
func (r *Router) handleLearningGrammarChapters(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract section_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/categories/")

	// Check if it's an access check request
	if strings.HasSuffix(path, "/access") {
		r.handleLearningGrammarSectionAccess(w, req)
		return
	}

	// Check if it's a category test request
	if strings.HasSuffix(path, "/test") {
		r.handleLearningGrammarCategoryTest(w, req)
		return
	}

	sectionID := strings.TrimSuffix(path, "/chapters")
	sectionID = strings.Trim(sectionID, "/")

	if sectionID == "" {
		http.Error(w, "section_id required", http.StatusBadRequest)
		return
	}

	chapters, err := r.grammarService.GetPublishedChapters(req.Context(), sectionID, userID)
	if err != nil {
		r.logger.Error("failed to get grammar chapters", zap.String("section_id", sectionID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not published") {
			http.Error(w, "Section not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type ChapterResponse struct {
		ChapterID         string            `json:"chapter_id"`
		Title             string            `json:"title"`
		TitleTranslations map[string]string `json:"title_translations,omitempty"`
		TitleShort        string            `json:"title_short,omitempty"`
		Description       string            `json:"description,omitempty"`
		Level             string            `json:"level,omitempty"`
		Order             int               `json:"order"`
		EstimatedMinutes  int               `json:"estimated_minutes,omitempty"`
		BestScore         int               `json:"best_score"`
		Passed            bool              `json:"passed"`
		LastAttemptAt     string            `json:"last_attempt_at,omitempty"`
		CanAccess         bool              `json:"can_access"`
	}

	chapterList := make([]ChapterResponse, 0, len(chapters))
	for _, chapter := range chapters {
		resp := ChapterResponse{
			ChapterID:         chapter.Chapter.ID,
			Title:             chapter.Title,
			TitleTranslations: chapter.Chapter.TitleTranslations,
			TitleShort:        chapter.Chapter.TitleShort,
			Description:       chapter.Chapter.Description,
			Level:             chapter.Chapter.Level,
			Order:             chapter.Chapter.Order,
			EstimatedMinutes:  chapter.Chapter.EstimatedMinutes,
			BestScore:         chapter.Progress.BestScore,
			Passed:            chapter.Progress.Passed,
			CanAccess:         chapter.CanAccess,
		}
		if !chapter.Progress.LastAttemptAt.IsZero() {
			resp.LastAttemptAt = chapter.Progress.LastAttemptAt.Format("2006-01-02T15:04:05Z")
		}
		chapterList = append(chapterList, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapters": chapterList,
	})
}

// handleLearningGrammarChapterOrTest handles chapter content, test, and access requests
func (r *Router) handleLearningGrammarChapterOrTest(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/chapters/")
	path = strings.Trim(path, "/")

	// Check if it's a "next chapter" request
	if strings.HasSuffix(path, "/next") {
		r.handleLearningGrammarNextChapter(w, req)
		return
	}

	// Check if it's an access check request
	if strings.HasSuffix(path, "/access") {
		r.handleLearningGrammarChapterAccess(w, req)
		return
	}

	// Check if it's a test request
	if strings.HasSuffix(path, "/test") {
		r.handleLearningGrammarChapterTest(w, req)
		return
	}

	// Otherwise it's a chapter content request
	r.handleLearningGrammarChapter(w, req)
}

// handleLearningGrammarNextChapter returns the next chapter id within the same section, in UI order.
// @Summary      Получить следующую главу
// @Description  Возвращает следующую опубликованную главу в той же категории в порядке, как в списке глав.
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        chapter_id  path  string  true  "ID главы"
// @Success      200  {object}  map[string]interface{}  "Следующая глава или признак последней"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Глава/категория не найдена"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/chapters/{chapter_id}/next [get]
func (r *Router) handleLearningGrammarNextChapter(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract chapter_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/chapters/")
	path = strings.Trim(path, "/")
	chapterID := strings.TrimSuffix(path, "/next")
	chapterID = strings.Trim(chapterID, "/")
	if chapterID == "" {
		http.Error(w, "chapter_id required", http.StatusBadRequest)
		return
	}

	nextID, isLast, sectionID, err := r.grammarService.GetNextPublishedChapterID(req.Context(), chapterID)
	if err != nil {
		r.logger.Error("failed to get next chapter", zap.String("chapter_id", chapterID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"section_id":      sectionID,
		"is_last":         isLast,
		"next_chapter_id": nextID,
	})
}

// handleLearningGrammarChapter returns chapter content
// @Summary      Получить контент главы
// @Description  Возвращает контент главы (без ответов для тестов)
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        chapter_id  path  string  true  "ID главы"
// @Success      200  {object}  map[string]interface{}  "Контент главы"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Глава не найдена"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/chapters/{chapter_id} [get]
func (r *Router) handleLearningGrammarChapter(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract chapter_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/chapters/")
	chapterID := strings.Trim(path, "/")

	if chapterID == "" {
		http.Error(w, "chapter_id required", http.StatusBadRequest)
		return
	}

	// For chapter content, we include answers for inline quizzes (quick feedback)
	// but not for chapter tests
	content, err := r.grammarService.GetChapterContent(req.Context(), chapterID, true)
	if err != nil {
		r.logger.Error("failed to get chapter content", zap.String("chapter_id", chapterID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not published") {
			http.Error(w, "Chapter not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapter":            content.Chapter,
		"title":              content.Title,
		"title_translations": content.Chapter.TitleTranslations,
	})
}

// handleLearningGrammarCategoryTest generates a category test
// @Summary      Получить тест по категории
// @Description  Генерирует тест из всех глав категории (минимум 2 вопроса на главу, до 20 всего)
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        section_id  path  string  true  "ID категории"
// @Success      200  {object}  map[string]interface{}  "Вопросы теста"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Категория не найдена"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/categories/{section_id}/test [get]
func (r *Router) handleLearningGrammarCategoryTest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract section_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/categories/")
	sectionID := strings.TrimSuffix(path, "/test")
	sectionID = strings.Trim(sectionID, "/")

	if sectionID == "" {
		http.Error(w, "section_id required", http.StatusBadRequest)
		return
	}

	test, err := r.grammarService.GenerateCategoryTest(req.Context(), sectionID)
	if err != nil {
		r.logger.Error("failed to generate category test", zap.String("section_id", sectionID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Section not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"questions": test.Questions,
		"total":     test.Total,
	})
}

// handleLearningGrammarChapterTest generates a chapter test
// @Summary      Получить тест по главе
// @Description  Генерирует тест из банка вопросов главы (без ответов)
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        chapter_id  path  string  true  "ID главы"
// @Success      200  {object}  map[string]interface{}  "Вопросы теста"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Глава не найдена"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/chapters/{chapter_id}/test [get]
func (r *Router) handleLearningGrammarChapterTest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract chapter_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/chapters/")
	path = strings.Trim(path, "/")

	// Remove /test suffix if present
	chapterID := strings.TrimSuffix(path, "/test")
	chapterID = strings.Trim(chapterID, "/")

	if chapterID == "" {
		http.Error(w, "chapter_id required", http.StatusBadRequest)
		return
	}

	test, err := r.grammarService.GenerateChapterTest(req.Context(), chapterID)
	if err != nil {
		r.logger.Error("failed to generate chapter test", zap.String("chapter_id", chapterID), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"questions": test.Questions,
		"total":     test.Total,
	})
}

// handleLearningGrammarSubmitTest submits test answers
// @Summary      Отправить ответы на тест
// @Description  Проверяет ответы и сохраняет результат теста
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        request  body  object  true  "Данные теста"
// @Success      200  {object}  map[string]interface{}  "Результат теста"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/tests/submit [post]
func (r *Router) handleLearningGrammarSubmitTest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		Scope   string               `json:"scope"`    // "chapter" or "category"
		ScopeID string               `json:"scope_id"` // chapter_id or section_id
		Answers []service.AnswerItem `json:"answers"`  // array of answer objects in test order
	}

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Scope == "" || request.ScopeID == "" {
		http.Error(w, "scope and scope_id required", http.StatusBadRequest)
		return
	}

	result, err := r.grammarService.SubmitTest(req.Context(), userID, request.Scope, request.ScopeID, request.Answers)
	if err != nil {
		r.logger.Error("failed to submit test", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score":   result.Score,
		"passed":  result.Passed,
		"correct": result.Correct,
		"total":   result.Total,
		"results": result.Results,
	})
}

// handleLearningGrammarChapterAccess checks if user can access a chapter
// @Summary      Проверить доступ к главе
// @Description  Проверяет, может ли пользователь получить доступ к главе (предыдущая глава должна быть пройдена)
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        chapter_id  path  string  true  "ID главы"
// @Success      200  {object}  map[string]interface{}  "Результат проверки доступа"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Глава не найдена"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/chapters/{chapter_id}/access [get]
func (r *Router) handleLearningGrammarChapterAccess(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract chapter_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/chapters/")
	path = strings.Trim(path, "/")

	// Remove /access suffix if present
	chapterID := strings.TrimSuffix(path, "/access")
	chapterID = strings.Trim(chapterID, "/")

	if chapterID == "" {
		http.Error(w, "chapter_id required", http.StatusBadRequest)
		return
	}

	canAccess, err := r.grammarService.CanAccessChapter(req.Context(), userID, chapterID)
	if err != nil {
		r.logger.Error("failed to check chapter access", zap.String("chapter_id", chapterID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Chapter not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"can_access": canAccess,
	})
}

// handleLearningGrammarSectionAccess checks if user can access a section
// @Summary      Проверить доступ к категории
// @Description  Проверяет, может ли пользователь получить доступ к категории (все главы предыдущей категории должны быть пройдены)
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        section_id  path  string  true  "ID категории"
// @Success      200  {object}  map[string]interface{}  "Результат проверки доступа"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Категория не найдена"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/categories/{section_id}/access [get]
func (r *Router) handleLearningGrammarSectionAccess(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract section_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/categories/")
	sectionID := strings.TrimSuffix(path, "/access")
	sectionID = strings.Trim(sectionID, "/")

	if sectionID == "" {
		http.Error(w, "section_id required", http.StatusBadRequest)
		return
	}

	canAccess, err := r.grammarService.CanAccessSection(req.Context(), userID, sectionID)
	if err != nil {
		r.logger.Error("failed to check section access", zap.String("section_id", sectionID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Section not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"can_access": canAccess,
	})
}

// handleLearningGrammarStatistics returns overall grammar statistics
// @Summary      Получить статистику грамматики
// @Description  Возвращает подтвержденный уровень грамматики и процент завершения курса
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Статистика грамматики"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/statistics [get]
func (r *Router) handleLearningGrammarStatistics(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := r.grammarService.GetGrammarStatistics(req.Context(), userID)
	if err != nil {
		r.logger.Error("failed to get grammar statistics", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get app settings for placement test button visibility
	appSettingsRepo := repository.NewAppSettingsRepository(r.db, r.logger)
	hidePlacementTestButton, err := appSettingsRepo.GetBoolSetting("hide_placement_test_button")
	if err != nil {
		r.logger.Warn("failed to get app settings for placement test button", zap.Error(err))
		// Default to false (button visible) if error
		hidePlacementTestButton = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"confirmed_level":             stats.ConfirmedLevel,
		"course_completion_pct":       stats.CourseCompletionPct,
		"whole_course_completion_pct": stats.WholeCourseCompletionPct,
		"average_test_score":          stats.AverageTestScore,
		"passed_chapters":             stats.PassedChapters,
		"total_chapters":              stats.TotalChapters,
		"total_chapters_in_course":    stats.TotalChaptersInCourse,
		"hide_placement_test_button":  hidePlacementTestButton,
	})
}

// handleLearningGrammarPlacementTest generates a placement test
// @Summary      Получить placement тест
// @Description  Генерирует тест на определение уровня (25 вопросов из всех опубликованных категорий)
// @Tags         Learning Grammar
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  service.TestQuestions  "Вопросы теста"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Router       /api/learning/grammar/placement-test [get]
func (r *Router) handleLearningGrammarPlacementTest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	test, err := r.grammarService.GeneratePlacementTest(req.Context())
	if err != nil {
		r.logger.Error("failed to generate placement test", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(test)
}

// handleLearningGrammarSubmitPlacementTest submits placement test answers
// @Summary      Отправить ответы placement теста
// @Description  Отправляет ответы на placement тест и определяет уровень пользователя
// @Tags         Learning Grammar
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        request  body  map[string]interface{}  true  "Ответы на вопросы (map question_id -> answer)"
// @Success      200  {object}  service.PlacementTestResult  "Результат теста"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Router       /api/learning/grammar/placement-test/submit [post]
func (r *Router) handleLearningGrammarSubmitPlacementTest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var answers map[string]interface{}
	if err := json.NewDecoder(req.Body).Decode(&answers); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result, err := r.grammarService.SubmitPlacementTest(req.Context(), userID, answers)
	if err != nil {
		r.logger.Error("failed to submit placement test", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
