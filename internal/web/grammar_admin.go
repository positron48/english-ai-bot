package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// handleAdminGrammarCategories returns all grammar categories for admin
// @Summary      Получить все категории грамматики (админ)
// @Description  Возвращает полный список категорий с информацией о публикации
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Список категорий"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/admin/grammar/categories [get]
func (r *Router) handleAdminGrammarCategories(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sectionsData, err := r.grammarService.ContentRepo.GetSections()
	if err != nil {
		r.logger.Error("failed to get sections", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	publishedItems, err := r.grammarService.PublishRepo.GetPublishedItemsByType("section")
	if err != nil {
		r.logger.Error("failed to get published items", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type CategoryAdminResponse struct {
		SectionID         string  `json:"section_id"`
		Title             string  `json:"title"`
		Level             string  `json:"level"`
		Order             int     `json:"order"`
		IsPublished       bool    `json:"is_published"`
		CustomName        *string `json:"custom_name,omitempty"`
		TotalChapters     int     `json:"total_chapters"`
		AvailableChapters int     `json:"available_chapters"`
		PublishedChapters int     `json:"published_chapters"`
	}

	categories := make([]CategoryAdminResponse, 0, len(sectionsData.Sections))
	for _, section := range sectionsData.Sections {
		item, exists := publishedItems[section.SectionID]
		isPublished := exists && item.IsPublished

		// Count chapters
		availableChapters := 0
		publishedChapters := 0
		for _, chapterID := range section.ChapterIDs {
			if r.grammarService.ContentRepo.ChapterExists(chapterID) {
				availableChapters++
				chapterItem, _ := r.grammarService.PublishRepo.GetPublishedItem("chapter", chapterID)
				if chapterItem.IsPublished {
					publishedChapters++
				}
			}
		}

		resp := CategoryAdminResponse{
			SectionID:         section.SectionID,
			Title:             section.Title,
			Level:             section.Level,
			Order:             section.Order,
			IsPublished:       isPublished,
			TotalChapters:     len(section.ChapterIDs),
			AvailableChapters: availableChapters,
			PublishedChapters: publishedChapters,
		}

		if exists && item.Name != nil {
			resp.CustomName = item.Name
		}

		categories = append(categories, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
	})
}

// handleAdminGrammarCategoryPublish sets published status for a category
// @Summary      Установить статус публикации категории
// @Description  Включает/выключает публикацию категории, опционально каскадно для глав
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        section_id  path  string  true  "ID категории"
// @Param        request     body  object  true  "Данные публикации"
// @Success      200  {object}  map[string]interface{}  "Результат"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/admin/grammar/categories/{section_id}/publish [post]
func (r *Router) handleAdminGrammarCategoryPublish(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract section_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/admin/grammar/categories/")
	sectionID := strings.TrimSuffix(path, "/publish")
	sectionID = strings.Trim(sectionID, "/")

	if sectionID == "" {
		http.Error(w, "section_id required", http.StatusBadRequest)
		return
	}

	var request struct {
		IsPublished bool `json:"is_published"`
		Cascade     bool `json:"cascade,omitempty"`
	}

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set category published status
	if err := r.grammarService.PublishRepo.SetPublished("section", sectionID, request.IsPublished, &userID); err != nil {
		r.logger.Error("failed to set category published", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Cascade to chapters if requested
	if request.Cascade {
		sectionsData, err := r.grammarService.ContentRepo.GetSections()
		if err == nil {
			for _, section := range sectionsData.Sections {
				if section.SectionID == sectionID {
					chapterIDs := make([]string, 0, len(section.ChapterIDs))
					for _, chapterID := range section.ChapterIDs {
						if r.grammarService.ContentRepo.ChapterExists(chapterID) {
							chapterIDs = append(chapterIDs, chapterID)
						}
					}
					if err := r.grammarService.PublishRepo.BulkSetPublished("chapter", chapterIDs, request.IsPublished, &userID); err != nil {
						r.logger.Error("failed to cascade publish to chapters", zap.Error(err))
					}
					break
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleAdminGrammarChapters returns chapters for admin
// @Summary      Получить главы категории (админ)
// @Description  Возвращает список глав категории с информацией о публикации
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        section_id  query  string  true  "ID категории"
// @Success      200  {object}  map[string]interface{}  "Список глав"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/admin/grammar/chapters [get]
func (r *Router) handleAdminGrammarChapters(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sectionID := req.URL.Query().Get("section_id")
	if sectionID == "" {
		http.Error(w, "section_id required", http.StatusBadRequest)
		return
	}

	sectionsData, err := r.grammarService.ContentRepo.GetSections()
	if err != nil {
		r.logger.Error("failed to get sections", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var section *repository.Section
	for i := range sectionsData.Sections {
		if sectionsData.Sections[i].SectionID == sectionID {
			section = &sectionsData.Sections[i]
			break
		}
	}

	if section == nil {
		http.Error(w, "Section not found", http.StatusNotFound)
		return
	}

	publishedItems, err := r.grammarService.PublishRepo.GetPublishedItemsByType("chapter")
	if err != nil {
		r.logger.Error("failed to get published items", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type ChapterAdminResponse struct {
		ChapterID   string  `json:"chapter_id"`
		Title       string  `json:"title"`
		IsPublished bool    `json:"is_published"`
		CustomName  *string `json:"custom_name,omitempty"`
		FileExists  bool    `json:"file_exists"`
	}

	chapters := make([]ChapterAdminResponse, 0, len(section.ChapterIDs))
	for _, chapterID := range section.ChapterIDs {
		fileExists := r.grammarService.ContentRepo.ChapterExists(chapterID)
		item, exists := publishedItems[chapterID]
		isPublished := exists && item.IsPublished

		// Try to get chapter title
		title := chapterID
		if fileExists {
			if chapter, err := r.grammarService.ContentRepo.GetChapter(chapterID); err == nil {
				title = chapter.Title
			}
		}

		resp := ChapterAdminResponse{
			ChapterID:   chapterID,
			Title:       title,
			IsPublished: isPublished,
			FileExists:  fileExists,
		}

		if exists && item.Name != nil {
			resp.CustomName = item.Name
		}

		chapters = append(chapters, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chapters": chapters,
	})
}

// handleAdminGrammarChapterPublish sets published status for a chapter
// @Summary      Установить статус публикации главы
// @Description  Включает/выключает публикацию главы
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        chapter_id  path  string  true  "ID главы"
// @Param        request     body  object  true  "Данные публикации"
// @Success      200  {object}  map[string]interface{}  "Результат"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/admin/grammar/chapters/{chapter_id}/publish [post]
func (r *Router) handleAdminGrammarChapterPublish(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract chapter_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/admin/grammar/chapters/")
	chapterID := strings.TrimSuffix(path, "/publish")
	chapterID = strings.Trim(chapterID, "/")

	if chapterID == "" {
		http.Error(w, "chapter_id required", http.StatusBadRequest)
		return
	}

	var request struct {
		IsPublished bool `json:"is_published"`
	}

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := r.grammarService.PublishRepo.SetPublished("chapter", chapterID, request.IsPublished, &userID); err != nil {
		r.logger.Error("failed to set chapter published", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleAdminGrammarItemRename sets custom name for an item
// @Summary      Переименовать категорию/главу
// @Description  Устанавливает кастомное имя для категории или главы
// @Tags         Admin
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        item_type  path  string  true  "Тип элемента (section/chapter)"
// @Param        item_id    path  string  true  "ID элемента"
// @Param        request    body  object  true  "Данные переименования"
// @Success      200  {object}  map[string]interface{}  "Результат"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      403  {string}  string  "Доступ запрещен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/admin/grammar/items/{item_type}/{item_id}/rename [post]
func (r *Router) handleAdminGrammarItemRename(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract item_type and item_id from path
	path := strings.TrimPrefix(req.URL.Path, "/api/admin/grammar/items/")
	path = strings.TrimSuffix(path, "/rename")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	itemType := parts[0]
	itemID := parts[1]

	if itemType != "section" && itemType != "chapter" {
		http.Error(w, "Invalid item_type, must be 'section' or 'chapter'", http.StatusBadRequest)
		return
	}

	var request struct {
		Name *string `json:"name"` // null to clear custom name
	}

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := r.grammarService.PublishRepo.SetName(itemType, itemID, request.Name, &userID); err != nil {
		r.logger.Error("failed to set name", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
