package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

	sectionsData, err := r.grammarServiceForRequest(req, userID).ContentRepo.GetSections()
	if err != nil {
		r.logger.Error("failed to get grammar categories", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	sectionItems, err := r.grammarServiceForRequest(req, userID).PublishRepo.GetPublishedItemsByType("section")
	if err != nil {
		r.logger.Error("failed to get grammar section publish state", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	chapterItems, err := r.grammarServiceForRequest(req, userID).PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		r.logger.Error("failed to get grammar chapter publish state", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	progressByChapter, err := r.grammarServiceForRequest(req, userID).AttemptRepo.GetAllChapterProgress(userID)
	if err != nil {
		r.logger.Error("failed to get grammar chapter progress", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	categoryScores, err := r.grammarServiceForRequest(req, userID).AttemptRepo.GetCategoryTestBestScores(userID)
	if err != nil {
		r.logger.Warn("failed to get grammar category scores, defaulting section access to locked", zap.Error(err))
		categoryScores = map[string]int{}
	}
	placementResult, _ := r.grammarServiceForRequest(req, userID).AttemptRepo.GetPlacementTestResult(userID)
	levelOrder := map[string]int{"A0": 0, "A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6, "mixed": -1}
	placementOpened := make(map[string]bool)
	placementEffectiveOrder := -1
	if placementResult != nil {
		for _, openedSectionID := range placementResult.OpenedSections {
			placementOpened[openedSectionID] = true
			for i := range sectionsData.Sections {
				if sectionsData.Sections[i].SectionID == openedSectionID {
					if ord, ok := levelOrder[sectionsData.Sections[i].Level]; ok && ord >= 0 && ord > placementEffectiveOrder {
						placementEffectiveOrder = ord
					}
					break
				}
			}
		}
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

	categories := make([]CategoryResponse, 0, len(sectionsData.Sections))
	for i := range sectionsData.Sections {
		section := &sectionsData.Sections[i]
		item, exists := sectionItems[section.SectionID]
		isPublished := exists && item.IsPublished
		title := section.Title
		if isPublished && item.Name != nil && *item.Name != "" {
			title = *item.Name
		}
		publishedChapters := 0
		passedChapters := 0
		totalScore := 0
		for _, chapterID := range section.ChapterIDs {
			chapterItem, exists := chapterItems[chapterID]
			if !exists || !chapterItem.IsPublished {
				continue
			}
			publishedChapters++
			progress := progressByChapter[chapterID]
			if progress != nil {
				if progress.Passed {
					passedChapters++
				}
				totalScore += progress.BestScore
			}
		}
		progressPercentage := 0
		if publishedChapters > 0 {
			progressPercentage = totalScore / publishedChapters
		}
		canAccess := false
		if isPublished {
			if placementOpened[section.SectionID] {
				canAccess = true
			} else if placementEffectiveOrder >= 0 {
				if secOrd, ok := levelOrder[section.Level]; ok && secOrd >= 0 && secOrd <= placementEffectiveOrder {
					canAccess = true
				}
			}
			if !canAccess && i == 0 {
				canAccess = true
			}
			if !canAccess && i > 0 {
				previousSection := &sectionsData.Sections[i-1]
				if categoryScores[previousSection.SectionID] >= 50 {
					canAccess = true
				} else {
					previousPublishedChapters := 0
					previousPassedChapters := 0
					for _, chapterID := range previousSection.ChapterIDs {
						chapterItem, exists := chapterItems[chapterID]
						if !exists || !chapterItem.IsPublished {
							continue
						}
						previousPublishedChapters++
						if progress := progressByChapter[chapterID]; progress != nil && progress.Passed {
							previousPassedChapters++
						}
					}
					canAccess = previousPublishedChapters > 0 && previousPassedChapters == previousPublishedChapters
				}
			}
		}
		var categoryTestScore *int
		if isPublished {
			if bestScore := categoryScores[section.SectionID]; bestScore > 0 {
				score := bestScore
				categoryTestScore = &score
			}
		}

		categories = append(categories, CategoryResponse{
			SectionID:          section.SectionID,
			Title:              title,
			TitleTranslations:  section.TitleTranslations,
			Level:              section.Level,
			Order:              section.Order,
			IsPublished:        isPublished,
			PublishedChapters:  publishedChapters,
			PassedChapters:     passedChapters,
			TotalChapters:      len(section.ChapterIDs),
			ProgressPercentage: progressPercentage,
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

	chapters, err := r.grammarServiceForRequest(req, userID).GetPublishedChapters(req.Context(), sectionID, userID)
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

	nextID, isLast, sectionID, err := r.grammarServiceForRequest(req, userID).GetNextPublishedChapterID(req.Context(), chapterID)
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
	content, err := r.grammarServiceForRequest(req, userID).GetChapterContent(req.Context(), chapterID, true)
	if err != nil {
		r.logger.Error("failed to get chapter content", zap.String("chapter_id", chapterID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not published") {
			http.Error(w, "Chapter not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"chapter":            content.Chapter,
		"title":              content.Title,
		"title_translations": content.Chapter.TitleTranslations,
	}
	if sec, err := r.grammarServiceForRequest(req, userID).GetSectionBySectionID(req.Context(), content.Chapter.SectionID); err == nil && sec != nil {
		resp["section"] = map[string]interface{}{
			"section_id":         sec.SectionID,
			"title":              sec.Title,
			"title_translations": sec.TitleTranslations,
			"level":              sec.Level,
		}
	} else {
		resp["section"] = map[string]interface{}{
			"section_id": content.Chapter.SectionID,
		}
		if err != nil && r.logger != nil {
			r.logger.Debug("grammar chapter: section metadata lookup failed",
				zap.String("section_id", content.Chapter.SectionID),
				zap.Error(err))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
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

	test, err := r.grammarServiceForRequest(req, userID).GenerateCategoryTest(req.Context(), sectionID)
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

	test, err := r.grammarServiceForRequest(req, userID).GenerateChapterTest(req.Context(), chapterID)
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

	result, err := r.grammarServiceForRequest(req, userID).SubmitTest(req.Context(), userID, request.Scope, request.ScopeID, request.Answers)
	if err != nil {
		r.logger.Error("failed to submit test", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	r.recordLinglowGrammarTestAttempt(req, userID, request.Scope, request.ScopeID, "", request.Answers, result)

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

func (r *Router) recordLinglowGrammarTestAttempt(req *http.Request, userID int64, scope, scopeID, clientAttemptID string, answers []service.AnswerItem, result *service.TestResult) {
	if r == nil || r.config == nil || !r.config.Linglow.EventsWriteEnabled || r.linglowEventRepo == nil || result == nil || result.AttemptID == 0 {
		return
	}
	answersJSON, _ := json.Marshal(answers)
	resultsJSON, _ := json.Marshal(result.Results)
	input := repository.GrammarTestEventInput{
		UserID:          userID,
		AttemptID:       result.AttemptID,
		ScopeType:       scope,
		ScopeID:         scopeID,
		Score:           result.Score,
		Passed:          result.Passed,
		TotalQuestions:  result.Total,
		AnswersJSON:     string(answersJSON),
		ResultsJSON:     string(resultsJSON),
		ClientAttemptID: clientAttemptID,
		AnsweredAt:      linglowAnsweredAt(result.AnsweredAt),
	}
	if _, err := r.linglowEventRepo.RecordGrammarTestAttempt(req.Context(), r.config.Learning, input); err != nil {
		r.logger.Warn("failed to dual-write linglow grammar test event",
			zap.Int64("user_id", userID),
			zap.Int64("attempt_id", result.AttemptID),
			zap.String("scope", scope),
			zap.String("scope_id", scopeID),
			zap.Error(err),
		)
	}
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

	canAccess, err := r.grammarServiceForRequest(req, userID).CanAccessChapter(req.Context(), userID, chapterID)
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

	canAccess, err := r.grammarServiceForRequest(req, userID).CanAccessSection(req.Context(), userID, sectionID)
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

// handleLearningGrammarContinueChapter returns the chapter the user should resume studying.
// @Summary      Получить главу для продолжения грамматики
// @Description  Возвращает главу, на которую нужно вести пользователя с home quick access
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Глава для продолжения или null"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/grammar/continue-chapter [get]
func (r *Router) handleLearningGrammarContinueChapter(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	chapter, err := r.grammarServiceForRequest(req, userID).GetContinueChapter(req.Context(), userID)
	if err != nil {
		r.logger.Error("failed to get continue grammar chapter", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	payload := map[string]interface{}{"chapter": nil}
	if chapter != nil {
		payload["chapter"] = map[string]interface{}{
			"chapter_id":         chapter.ChapterID,
			"title":              chapter.Title,
			"title_translations": chapter.TitleTranslations,
			"section_id":         chapter.SectionID,
			"url":                "/learning/grammar/chapter/" + chapter.ChapterID,
		}
	}
	json.NewEncoder(w).Encode(payload)
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

	stats, err := r.grammarServiceForRequest(req, userID).GetGrammarStatistics(req.Context(), userID)
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
// Kept for historical tests; the route is retired and documented in placement_swagger.go.
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

	test, err := r.grammarServiceForRequest(req, userID).GeneratePlacementTest(req.Context())
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
// Kept for historical tests; the route is retired and documented in placement_swagger.go.
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

	result, err := r.grammarServiceForRequest(req, userID).SubmitPlacementTest(req.Context(), userID, answers)
	if err != nil {
		r.logger.Error("failed to submit placement test", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (r *Router) handleLearningGrammarTrainingAvailability(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	availability, err := r.grammarServiceForRequest(req, userID).GetGrammarTrainingAvailability(req.Context(), userID)
	if err != nil {
		r.logger.Error("failed to get grammar training availability", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"grammar_training": availability,
	})
}

type grammarTrainingStartRequest struct {
	Limit int `json:"limit"`
}

func (r *Router) handleLearningGrammarTrainingStart(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body grammarTrainingStartRequest
	_ = json.NewDecoder(req.Body).Decode(&body)
	session, err := r.grammarServiceForRequest(req, userID).StartGrammarSrsSession(req.Context(), userID, body.Limit)
	if err != nil {
		r.logger.Error("failed to start grammar training session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(session)
}

type grammarTrainingAnswerRequest struct {
	QuestionID string      `json:"question_id"`
	Answer     interface{} `json:"answer"`
}

func (r *Router) handleLearningGrammarTrainingAnswer(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body grammarTrainingAnswerRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.QuestionID) == "" {
		http.Error(w, "question_id required", http.StatusBadRequest)
		return
	}
	result, err := r.grammarServiceForRequest(req, userID).SubmitGrammarSrsAnswer(req.Context(), userID, body.QuestionID, body.Answer)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Question not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to submit grammar training answer", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	r.recordLinglowGrammarTrainingAttempt(req, userID, "", result)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (r *Router) recordLinglowGrammarTrainingAttempt(req *http.Request, userID int64, clientAttemptID string, result *service.GrammarSrsAnswerResult) {
	if r == nil || r.config == nil || !r.config.Linglow.EventsWriteEnabled || r.linglowEventRepo == nil || result == nil || result.AttemptID == 0 {
		return
	}
	answerJSON, _ := json.Marshal(result.UserAnswer)
	correctJSON, _ := json.Marshal(result.CorrectAnswer)
	input := repository.GrammarTrainingEventInput{
		UserID:          userID,
		AttemptID:       result.AttemptID,
		ChapterID:       result.ChapterID,
		TheoryBlockID:   result.TheoryBlockID,
		ConceptID:       result.ConceptID,
		QuestionID:      result.QuestionID,
		IsCorrect:       result.Correct,
		AnswerJSON:      string(answerJSON),
		CorrectJSON:     string(correctJSON),
		ClientAttemptID: strings.TrimSpace(firstNonEmpty(clientAttemptID, result.ClientAttemptID)),
		AnsweredAt:      linglowAnsweredAt(result.AnsweredAt),
	}
	if _, err := r.linglowEventRepo.RecordGrammarTrainingAttempt(req.Context(), r.config.Learning, input); err != nil {
		r.logger.Warn("failed to dual-write linglow grammar training event",
			zap.Int64("user_id", userID),
			zap.Int64("attempt_id", result.AttemptID),
			zap.String("question_id", result.QuestionID),
			zap.Error(err),
		)
	}
	r.mirrorLegacyGrammarSRS(req.Context(), userID, result.ChapterID, result.TheoryBlockID)
}

func (r *Router) mirrorLegacyGrammarSRS(ctx context.Context, userID int64, chapterID, theoryBlockID string) {
	if r == nil || r.config == nil || !r.config.Linglow.SRSReadEnabled || r.linglowSRSMirrorRepo == nil {
		return
	}
	if err := r.linglowSRSMirrorRepo.MirrorGrammarTraining(ctx, r.config.Learning, userID, chapterID, theoryBlockID); err != nil {
		r.logger.Warn("failed to mirror legacy grammar srs snapshot",
			zap.Int64("user_id", userID),
			zap.String("chapter_id", chapterID),
			zap.String("theory_block_id", theoryBlockID),
			zap.Error(err),
		)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func linglowAnsweredAt(at time.Time) time.Time {
	if !at.IsZero() {
		return at.UTC()
	}
	return time.Now().UTC()
}
