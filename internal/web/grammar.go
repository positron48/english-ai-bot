package web

import (
	"encoding/json"
	"net/http"
	"strings"

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

	sections, err := r.grammarService.GetPublishedSections(req.Context(), userID)
	if err != nil {
		r.logger.Error("failed to get grammar categories", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type CategoryResponse struct {
		SectionID         string `json:"section_id"`
		Title             string `json:"title"`
		Level             string `json:"level"`
		Order             int    `json:"order"`
		PublishedChapters int    `json:"published_chapters"`
		PassedChapters    int    `json:"passed_chapters"`
		TotalChapters     int    `json:"total_chapters"`
	}

	categories := make([]CategoryResponse, 0, len(sections))
	for _, section := range sections {
		categories = append(categories, CategoryResponse{
			SectionID:         section.Section.SectionID,
			Title:             section.Title,
			Level:             section.Section.Level,
			Order:             section.Section.Order,
			PublishedChapters: section.PublishedChapters,
			PassedChapters:    section.PassedChapters,
			TotalChapters:     len(section.Section.ChapterIDs),
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
	sectionID := strings.TrimSuffix(path, "/chapters")
	sectionID = strings.Trim(sectionID, "/")

	if sectionID == "" {
		http.Error(w, "section_id required", http.StatusBadRequest)
		return
	}

	chapters, err := r.grammarService.GetPublishedChapters(req.Context(), sectionID, userID)
	if err != nil {
		r.logger.Error("failed to get grammar chapters", zap.String("section_id", sectionID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Section not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type ChapterResponse struct {
		ChapterID      string `json:"chapter_id"`
		Title          string `json:"title"`
		TitleShort     string `json:"title_short,omitempty"`
		Description   string `json:"description,omitempty"`
		Level          string `json:"level,omitempty"`
		Order          int    `json:"order"`
		EstimatedMinutes int `json:"estimated_minutes,omitempty"`
		BestScore      int   `json:"best_score"`
		Passed         bool  `json:"passed"`
		LastAttemptAt  string `json:"last_attempt_at,omitempty"`
	}

	chapterList := make([]ChapterResponse, 0, len(chapters))
	for _, chapter := range chapters {
		resp := ChapterResponse{
			ChapterID:        chapter.Chapter.ID,
			Title:            chapter.Title,
			TitleShort:       chapter.Chapter.TitleShort,
			Description:      chapter.Chapter.Description,
			Level:            chapter.Chapter.Level,
			Order:            chapter.Chapter.Order,
			EstimatedMinutes: chapter.Chapter.EstimatedMinutes,
			BestScore:        chapter.Progress.BestScore,
			Passed:           chapter.Progress.Passed,
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

// handleLearningGrammarChapterOrTest handles both chapter content and test requests
func (r *Router) handleLearningGrammarChapterOrTest(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/grammar/chapters/")
	path = strings.Trim(path, "/")
	
	// Check if it's a test request
	if strings.HasSuffix(path, "/test") {
		r.handleLearningGrammarChapterTest(w, req)
		return
	}
	
	// Otherwise it's a chapter content request
	r.handleLearningGrammarChapter(w, req)
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
		"chapter": content.Chapter,
		"title":   content.Title,
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
		Scope   string                 `json:"scope"`   // "chapter" or "category"
		ScopeID string                 `json:"scope_id"` // chapter_id or section_id
		Answers map[string]interface{} `json:"answers"`
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
