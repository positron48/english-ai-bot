package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

// handleLearningWordsCategories returns categories for a specific parent
// @Summary      Получить категории по уровню
// @Description  Возвращает категории для указанного родителя (parent_id). Если parent_id не указан или null, возвращает корневые категории
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        parent_id  query  int  false  "ID родительской категории (null для корневых)"
// @Success      200  {object}  map[string]interface{}  "Список категорий"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/words/categories [get]
func (r *Router) handleLearningWordsCategories(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Scope to the course the client selected (explicit course_code), falling back to current.
	courseCode := r.requestedCourseCodeForUser(req, userID)

	// Parse parent_id from query (null means root level)
	var parentID *int64
	if parentIDStr := req.URL.Query().Get("parent_id"); parentIDStr != "" {
		if id, err := strconv.ParseInt(parentIDStr, 10, 64); err == nil {
			parentID = &id
		}
	}

	// Check if we need all categories (for building hierarchy)
	allCategoriesRequested := req.URL.Query().Get("all") == "true"

	categoryRepo := repository.NewWordSetCategoryRepository(r.db, r.logger)
	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)

	// Get categories scoped to the course, then keep only published ones (public API).
	courseCategories, err := categoryRepo.GetAllCategoriesForCourse(courseCode)
	if err != nil {
		r.logger.Error("failed to get categories", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	allCategories := make([]*models.WordSetCategory, 0, len(courseCategories))
	for _, cat := range courseCategories {
		if cat.IsPublished {
			allCategories = append(allCategories, cat)
		}
	}

	// Build parent_id -> children index for subtree progress aggregation.
	childrenByParent := make(map[int64][]*models.WordSetCategory)
	for _, cat := range allCategories {
		if cat.ParentID != nil {
			childrenByParent[*cat.ParentID] = append(childrenByParent[*cat.ParentID], cat)
		}
	}

	// Filter by parent_id
	var filteredCategories []*models.WordSetCategory
	if allCategoriesRequested {
		// Return all categories when all=true
		filteredCategories = allCategories
	} else {
		for _, cat := range allCategories {
			if parentID == nil {
				// Root level - only categories without parent
				if cat.ParentID == nil {
					filteredCategories = append(filteredCategories, cat)
				}
			} else {
				// Specific parent - only direct children
				if cat.ParentID != nil && *cat.ParentID == *parentID {
					filteredCategories = append(filteredCategories, cat)
				}
			}
		}
	}

	// Build response (flat list, no children)
	type CategoryNode struct {
		ID              int64   `json:"id"`
		Name            string  `json:"name"`
		Description     *string `json:"description,omitempty"`
		SortOrder       int     `json:"sort_order"`
		ParentID        *int64  `json:"parent_id,omitempty"`
		LevelCode       *string `json:"level_code,omitempty"`
		TotalWords      int     `json:"total_words"`
		KnownWords      int     `json:"known_words"`
		WordsInVocab    int     `json:"words_in_vocab"`
		UnknownWords    int     `json:"unknown_words"`
		ProgressPercent float64 `json:"progress_percent"`
	}

	result := make([]CategoryNode, 0, len(filteredCategories))
	for _, cat := range filteredCategories {
		node := CategoryNode{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
			SortOrder:   cat.SortOrder,
			ParentID:    cat.ParentID,
			LevelCode:   cat.LevelCode,
		}

		// Aggregate progress across this category and all its descendants.
		subtreeIDs := collectCategorySubtree(cat.ID, childrenByParent)
		total, known, inVocab, err := wordSetRepo.GetCategoriesAggregateProgress(subtreeIDs, userID)
		if err != nil {
			r.logger.Warn("failed to get category progress", zap.Int64("category_id", cat.ID), zap.Error(err))
		} else {
			unknown := total - known - inVocab
			if unknown < 0 {
				unknown = 0
			}
			node.TotalWords = total
			node.KnownWords = known
			node.WordsInVocab = inVocab
			node.UnknownWords = unknown
			if total > 0 {
				node.ProgressPercent = float64(known+inVocab) / float64(total) * 100.0
			}
		}

		result = append(result, node)
	}

	r.logger.Debug("returning categories", zap.Int("count", len(result)), zap.Any("parent_id", parentID))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": result,
	})
}

// collectCategorySubtree returns the given category ID plus all descendant
// category IDs, walking the parent->children index breadth-first.
func collectCategorySubtree(rootID int64, childrenByParent map[int64][]*models.WordSetCategory) []int64 {
	ids := []int64{rootID}
	queue := []int64{rootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range childrenByParent[cur] {
			ids = append(ids, child.ID)
			queue = append(queue, child.ID)
		}
	}
	return ids
}

// handleLearningWordsSets returns word sets with progress
// @Summary      Получить список наборов слов
// @Description  Возвращает список наборов слов с прогрессом пользователя, с возможностью фильтрации
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        category_id  query  int     false  "ID категории"
// @Param        limit        query  int     false  "Лимит (по умолчанию 50)"
// @Param        offset       query  int     false  "Смещение (по умолчанию 0)"
// @Success      200  {object}  map[string]interface{}  "Список наборов с прогрессом"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/words/sets [get]
func (r *Router) handleLearningWordsSets(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)

	// Scope to the course the client selected (explicit course_code), falling back to current.
	courseCode := r.requestedCourseCodeForUser(req, userID)

	// Parse query parameters
	// category_id can be:
	// - not specified: show only sets without category (category_id IS NULL)
	// - specified: show sets for that category
	var categoryID *int64
	var showOnlyWithoutCategory bool
	if catIDStr := req.URL.Query().Get("category_id"); catIDStr != "" {
		if catID, err := strconv.ParseInt(catIDStr, 10, 64); err == nil {
			categoryID = &catID
		}
	} else {
		// If category_id not specified, show only sets without category
		showOnlyWithoutCategory = true
	}

	limit := 50
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := req.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get word sets (only published for users)
	// If showOnlyWithoutCategory is true, we need to filter by category_id IS NULL
	var wordSets []*models.WordSet
	var err error
	if showOnlyWithoutCategory {
		// Query sets without category directly
		query := `SELECT id, category_id, title, description, is_published, sort_order, preferred_pos, created_at, updated_at
				  FROM word_sets WHERE category_id IS NULL AND is_published = 1 AND (? = '' OR course_code = ?)`
		args := []interface{}{courseCode, courseCode}

		query += " ORDER BY sort_order, title LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		rows, err := r.db.Query(query, args...)
		if err != nil {
			r.logger.Error("failed to list word sets without category", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		wordSets = []*models.WordSet{}
		for rows.Next() {
			var ws models.WordSet
			var categoryID sql.NullInt64
			var preferredPOS sql.NullString
			var createdAt, updatedAt string

			if err := rows.Scan(&ws.ID, &categoryID, &ws.Title, &ws.Description, &ws.IsPublished, &ws.SortOrder, &preferredPOS, &createdAt, &updatedAt); err != nil {
				r.logger.Warn("failed to scan word set", zap.Error(err))
				continue
			}

			if preferredPOS.Valid {
				ws.PreferredPOS = &preferredPOS.String
			}

			if createdAt != "" {
				for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
					if t, err := time.Parse(layout, createdAt); err == nil {
						ws.CreatedAt = t
						break
					}
				}
			}
			if updatedAt != "" {
				for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
					if t, err := time.Parse(layout, updatedAt); err == nil {
						ws.UpdatedAt = t
						break
					}
				}
			}

			wordSets = append(wordSets, &ws)
		}
	} else {
		wordSets, err = wordSetRepo.ListWordSetsForCourse(courseCode, categoryID, limit, offset, false)
	}
	if err != nil {
		r.logger.Error("failed to list word sets", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get progress for each set
	setsWithProgress := make([]*models.WordSetWithProgress, 0, len(wordSets))
	for _, ws := range wordSets {
		progress, err := wordSetRepo.GetWordSetProgress(ws.ID, userID)
		if err != nil {
			r.logger.Warn("failed to get progress for word set",
				zap.Int64("word_set_id", ws.ID),
				zap.Error(err),
			)
			// Continue with zero progress
			progress = &models.WordSetWithProgress{
				WordSet:         *ws,
				TotalWords:      0,
				KnownWords:      0,
				WordsInVocab:    0,
				UnknownWords:    0,
				ProgressPercent: 0.0,
			}
		}
		setsWithProgress = append(setsWithProgress, progress)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sets": setsWithProgress,
	})
}

// handleLearningWordsSetDetailOrStudy routes to detail or study handlers
func (r *Router) handleLearningWordsSetDetailOrStudy(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/words/sets/")
	parts := strings.Split(path, "/")

	setIDStr := parts[0]
	if setIDStr == "" {
		http.Error(w, "Set ID required", http.StatusBadRequest)
		return
	}

	// Check if it's a study action
	if len(parts) >= 2 && parts[1] == "study" {
		if len(parts) >= 3 {
			switch parts[2] {
			case "learn":
				r.handleLearningWordsSetStudyLearn(w, req)
				return
			case "know":
				r.handleLearningWordsSetStudyKnow(w, req)
				return
			}
		}
		// GET /app/learning/words/sets/{id}/study
		if req.Method == http.MethodGet {
			r.handleLearningWordsSetStudy(w, req)
			return
		}
	}

	// Default to detail view
	r.handleLearningWordsSetDetail(w, req)
}

// handleLearningWordsSetDetail returns word set details with words
// @Summary      Получить детали набора слов
// @Description  Возвращает детальную информацию о наборе слов, включая список слов со статусом для пользователя
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id  path  int  true  "ID набора слов"
// @Success      200  {object}  map[string]interface{}  "Детали набора с словами"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Набор не найден"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/words/sets/{id} [get]
func (r *Router) handleLearningWordsSetDetail(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract set ID from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/words/sets/")
	parts := strings.Split(path, "/")

	setID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)

	// Get word set
	wordSet, err := wordSetRepo.GetWordSet(setID)
	if err != nil {
		r.logger.Error("failed to get word set", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if wordSet == nil {
		http.Error(w, "Word set not found", http.StatusNotFound)
		return
	}

	// Get progress
	progress, err := wordSetRepo.GetWordSetProgress(setID, userID)
	if err != nil {
		r.logger.Warn("failed to get progress", zap.Error(err))
		progress = &models.WordSetWithProgress{
			WordSet:         *wordSet,
			TotalWords:      0,
			KnownWords:      0,
			WordsInVocab:    0,
			UnknownWords:    0,
			ProgressPercent: 0.0,
		}
	}

	// Get words with status
	words, err := wordSetRepo.GetWordSetWords(setID, userID)
	if err != nil {
		r.logger.Error("failed to get word set words", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	wordRepo := repository.NewWordRepository(r.db, r.logger)
	wordCardIDs := make([]int64, 0, len(words))
	for _, wInfo := range words {
		if wInfo != nil {
			wordCardIDs = append(wordCardIDs, wInfo.WordCardID)
		}
	}
	wordCardsByID, err := wordRepo.GetWordCardsByIDs(wordCardIDs)
	if err != nil {
		r.logger.Warn("failed to batch-load word cards for morphology", zap.Error(err))
		wordCardsByID = map[int64]*models.WordCard{}
	}
	for _, wInfo := range words {
		if wInfo == nil {
			continue
		}
		if wordCard := wordCardsByID[wInfo.WordCardID]; wordCard != nil {
			wInfo.Morph = buildCompactMorphFromWordCard(r.config.Learning.TargetLang, wordCard, nil)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"word_set": progress,
		"words":    words,
	})
}

// handleLearningWordsSetStudy returns training card for a word in the set
// @Summary      Получить словарную карточку для слова из набора
// @Description  Возвращает training card для указанного слова из набора слов
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id           path  int  true  "ID набора слов"
// @Param        word_card_id query int  true  "ID слова (word_card_id)"
// @Success      200  {object}  map[string]interface{}  "Training card для слова"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Слово или карточка не найдены"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/words/sets/{id}/study [get]
func (r *Router) handleLearningWordsSetStudy(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract set ID from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/words/sets/")
	parts := strings.Split(path, "/")

	setID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	// Get word_card_id from query parameter
	wordCardIDStr := req.URL.Query().Get("word_card_id")
	if wordCardIDStr == "" {
		http.Error(w, "word_card_id is required", http.StatusBadRequest)
		return
	}

	wordCardID, err := strconv.ParseInt(wordCardIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid word_card_id", http.StatusBadRequest)
		return
	}

	// Verify word is in the set
	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)
	inSet, err := wordSetRepo.IsWordInSet(setID, wordCardID)
	if err != nil {
		r.logger.Error("failed to check word set membership", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !inSet {
		http.Error(w, "Word not found in set", http.StatusNotFound)
		return
	}

	// Get word set to check preferred_pos (best-effort; nil means no preferred_pos)
	wordSet, _ := wordSetRepo.GetWordSet(setID)

	// Get training cards for the word
	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)
	trainingCards, _ := trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)

	if len(trainingCards) == 0 {
		// Try to ensure training cards exist
		wordSetService := r.getWordSetService()
		if err := wordSetService.EnsureTrainingCardsExist(req.Context(), wordCardID); err != nil {
			r.logger.Warn("failed to ensure training cards",
				zap.Int64("word_card_id", wordCardID),
				zap.Error(err),
			)
		}
		// Try to get again
		trainingCards, _ = trainingCardRepo.GetTrainingCardsByWordCardID(wordCardID)
		if len(trainingCards) == 0 {
			http.Error(w, "Training card not found for word", http.StatusNotFound)
			return
		}
	}

	// Select training card based on preferred_pos if set
	var selectedCard *models.TrainingCard
	if wordSet != nil && wordSet.PreferredPOS != nil && *wordSet.PreferredPOS != "" {
		// Try to find card with matching POS (case-insensitive comparison)
		preferredPOSLower := strings.ToLower(*wordSet.PreferredPOS)
		for _, card := range trainingCards {
			if card.POS != nil && strings.ToLower(*card.POS) == preferredPOSLower {
				selectedCard = card
				break
			}
		}
		// If no matching POS found, fall back to first card (show word anyway)
		if selectedCard == nil && len(trainingCards) > 0 {
			selectedCard = trainingCards[0]
		}
	} else {
		// No preferred_pos, return the first training card (sense_index = 0)
		for _, card := range trainingCards {
			if card.SenseIndex == 0 {
				selectedCard = card
				break
			}
		}
		if selectedCard == nil && len(trainingCards) > 0 {
			selectedCard = trainingCards[0]
		}
	}

	if selectedCard == nil {
		http.Error(w, "Training card not found for word", http.StatusNotFound)
		return
	}

	// Add compact morphology payload from canonical word card.
	wordRepo := repository.NewWordRepository(r.db, r.logger)
	wordCard, _ := wordRepo.GetWordCardByID(wordCardID)
	selectedCard.Morph = buildCompactMorphFromWordCard(r.config.Learning.TargetLang, wordCard, selectedCard.POS)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"training_card": selectedCard,
	})
}

// handleLearningWordsSetStudyLearn adds a word to training
// @Summary      Добавить слово в тренировку
// @Description  Добавляет слово в тренировку (создает user_cards) и убирает статус "known" если был
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id          path  int  true  "ID набора слов"
// @Param        word_card_id  body  int  true  "ID слова (word_card_id)"
// @Success      200  {object}  map[string]interface{}  "Успешное добавление"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/words/sets/{id}/study/learn [post]
func (r *Router) handleLearningWordsSetStudyLearn(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Adding a word to learning changes vocabulary counts (and may add verb cards).
	defer r.BumpUserCache(userID)

	// Extract set ID from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/words/sets/")
	parts := strings.Split(path, "/")

	setID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var requestData struct {
		WordCardID int64 `json:"word_card_id"`
	}

	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify word is in the set
	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)
	inSet, err := wordSetRepo.IsWordInSet(setID, requestData.WordCardID)
	if err != nil {
		r.logger.Error("failed to check word set membership", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !inSet {
		http.Error(w, "Word not found in set", http.StatusBadRequest)
		return
	}

	// Remove known status if exists
	userWordKnowledgeRepo := repository.NewUserWordKnowledgeRepository(r.db, r.logger)
	if err := userWordKnowledgeRepo.RemoveKnown(userID, requestData.WordCardID); err != nil {
		r.logger.Warn("failed to remove known status", zap.Error(err))
	}

	// Ensure training cards + user cards exist (errors are logged but not fatal - cards may be
	// created later); single training-cards fetch instead of two.
	wordSetService := r.getWordSetService()
	if err := wordSetService.EnsureCardsForWord(req.Context(), userID, requestData.WordCardID); err != nil {
		r.logger.Warn("failed to ensure cards",
			zap.Int64("word_card_id", requestData.WordCardID),
			zap.Error(err),
		)
	}

	r.ensureVerbFormUserCardsAfterVocab(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleLearningWordsSetStudyKnow marks a word as known
// @Summary      Отметить слово как известное
// @Description  Отмечает слово как известное и удаляет user_cards, чтобы оно не попадало в тренировки
// @Tags         Learning
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        id          path  int  true  "ID набора слов"
// @Param        word_card_id  body  int  true  "ID слова (word_card_id)"
// @Success      200  {object}  map[string]interface{}  "Успешное действие"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/learning/words/sets/{id}/study/know [post]
func (r *Router) handleLearningWordsSetStudyKnow(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Marking a word known changes vocabulary counts.
	defer r.BumpUserCache(userID)

	// Extract set ID from path
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/words/sets/")
	parts := strings.Split(path, "/")

	setID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid set ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var requestData struct {
		WordCardID int64 `json:"word_card_id"`
	}

	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify word is in the set
	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)
	words, err := wordSetRepo.GetWordSetWords(setID, userID)
	if err != nil {
		r.logger.Error("failed to get word set words", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	found := false
	for _, word := range words {
		if word.WordCardID == requestData.WordCardID {
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Word not found in set", http.StatusBadRequest)
		return
	}

	// Mark as known
	wordSetService := r.getWordSetService()
	if err := wordSetService.MarkKnown(userID, requestData.WordCardID); err != nil {
		r.logger.Error("failed to mark as known", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// getWordSetService returns or creates word set service
func (r *Router) getWordSetService() *service.WordSetService {
	return r.getWordSetServiceForCourse(r.defaultCourseCode())
}

func (r *Router) getWordSetServiceForCourse(courseCode string) *service.WordSetService {
	wordSetRepo := repository.NewWordSetRepository(r.db, r.logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(r.db, r.logger)
	wordRepo := repository.NewWordRepository(r.db, r.logger)
	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)
	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)
	userWordKnowledgeRepo := repository.NewUserWordKnowledgeRepository(r.db, r.logger)
	userWordMasteringRepo := repository.NewUserWordMasteringRepository(r.db, r.logger)

	// Get AI service with type assertion
	var aiService *ai.Service
	if r.aiService != nil {
		if svc, ok := r.aiService.(*ai.Service); ok {
			aiService = svc
		}
	}

	return service.NewWordSetServiceWithMastering(
		wordSetRepo,
		wordSetCategoryRepo,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userWordKnowledgeRepo,
		userWordMasteringRepo,
		aiService,
		learningConfigForCourse(r.config.Learning, courseCode),
		r.config.AI.ModelHigh,
		r.logger,
	)
}

func learningConfigForCourse(fallback config.LearningConfig, courseCode string) config.LearningConfig {
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(courseCode)), "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fallback
	}
	target, native := parts[0], parts[1]
	return config.LearningConfig{
		Pair:            native + "-" + target,
		NativeLang:      native,
		TargetLang:      target,
		AppCode:         map[string]string{"en": "english", "es": "spanish"}[target],
		GrammarBundleID: target,
		ContentSource:   "db",
	}
}
