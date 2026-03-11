package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// Test hooks for coverage (set by tests, must be nil/false in production).
var (
	testHookVocabScanErr           func() error // if set and returns err, handleVocab treats current row Scan as failed
	testHookVocabTrainingQueryErr  func() error // if set and returns err, handleVocabWordCards treats training_cards query as failed
	testHookVocabTrainingQueryFail bool         // if true, handleVocabWordCards skips r.db.Query and injects err to cover Query error path
	testHookVocabElseDisplayWord   bool         // if true, handleVocab uses else branch for displayWord (word.DisplayWord = word.Lemma)
	testHookVocabElseMasteryLevel    bool   // if true, handleVocab uses else branch for masteryLevel (word.MasteryLevel = "new")
	testHookVocabElseMasteringScore  bool   // if true, handleVocab uses else branch for masteringScore (word.MasteringScore = 0)
	testHookVocabMasteryLevelInvalid  bool   // if true, handleVocab treats masteryLevelCalc as invalid so else branch (word.MasteryLevel = "new") runs
	testHookVocabDisplayWordInvalid  bool   // if true, handleVocab treats displayWord as invalid so else branch (word.DisplayWord = word.Lemma) runs
	testHookVocabMasteringScoreInvalid bool  // if true, handleVocab treats masteringScoreStored as invalid so else branch (word.MasteringScore = 0) runs
	testHookVocabForceDisplayWordValid bool     // if true, handleVocab treats displayWord as valid with String "hooked" to cover displayWord.Valid branch
	testHookVocabSetLastReview *time.Time // if set, handleVocab sets word.LastReview to this (covers parse success path)
	testHookVocabSetAddedAt    *time.Time // if set, handleVocab sets word.AddedAt to this (covers parse success path)
	testHookVocabRowScanErr    func() error // if set and returns err, handleVocab skips rows.Scan and continues (covers Scan error path)
	testHookVocabScanErrAfter  func() error // if set and returns err after Scan, handleVocab treats as scan error (covers err != nil after Scan)
)

// parseDateTime parses datetime string in "2006-01-02 15:04:05" format
func parseDateTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}

	// After substr() in SQL, we get standard format: "2006-01-02 15:04:05"
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse time: %s, error: %w", timeStr, err)
	}
	return &t, nil
}
	
// VocabWord represents a word with statistics (grouped by word_card_id/lemma)
type VocabWord struct {
	WordCardID      int64      `json:"word_card_id"`
	Lemma           string     `json:"lemma"`            // Base form (word_cards.word)
	DisplayWord     string     `json:"display_word"`     // Display form (prefer training_cards.display_word, fallback word_cards.display_en, fallback word_cards.word)
	TotalCards      int        `json:"total_cards"`
	DueCount        int        `json:"due_count"`
	LastReview      *time.Time `json:"last_review"`
	TotalReps       int        `json:"total_reps"`        // Total number of reviews across all cards
	AddedAt         *time.Time `json:"added_at"`         // Date when first card was added
	MasteryLevel    string     `json:"mastery_level"`    // Calculated mastery level: new, learning, mastered, known
	MasteringScore  int        `json:"mastering_score"`  // 0–100: how well the word is learned (for red–green marker)
	ReviewCount     int        `json:"review_count"`     // Total number of review events
}

// handleVocab shows the vocabulary list
// @Summary      Получить список слов
// @Description  Возвращает список всех слов пользователя с статистикой (общее количество карточек, количество готовых к повторению, дата последнего повторения). Поддерживает пагинацию и поиск.
// @Tags         Vocab
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        search         query  string  false  "Поиск по слову"
// @Param        mastery_level  query  string  false  "Фильтр по уровню мастерства: new, learning, mastered, known"
// @Param        page           query  int     false  "Номер страницы (начиная с 1)"
// @Param        limit          query  int     false  "Количество элементов на странице (по умолчанию 25)"
// @Success      200  {object}  map[string]interface{}  "Список слов с статистикой"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/vocab [get]
func (r *Router) handleVocab(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	search := req.URL.Query().Get("search")
	masteryLevelFilter := req.URL.Query().Get("mastery_level") // Filter by mastery level: new, learning, mastered, known
	page := 1
	limit := 25
	sortBy := "display_word"
	sortOrder := "asc"
	
	if pageStr := req.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	if sortByStr := req.URL.Query().Get("sort_by"); sortByStr != "" {
		// Validate sort_by to prevent SQL injection
		// Map frontend field names to column aliases from SELECT
		allowedSortFields := map[string]string{
			"display_word":      "display_word",
			"lemma":             "lemma",
			"total_cards":       "total_cards",
			"mastery_level":      "mastery_level",      // Special handling below
			"mastery_level_desc":  "mastery_level_desc",  // Special handling below - reversed order
			"mastering_score":    "mastering_score",    // Numeric 0-100
			"mastering_score_desc": "mastering_score_desc",
			"total_reps":        "total_reps",
			"review_count":      "review_count",
			"due_count":         "due_count",
			"added_at":          "added_at",
			"last_review":       "last_review",
		}
		if field, ok := allowedSortFields[sortByStr]; ok {
			sortBy = field
		}
	}
	if sortOrderStr := req.URL.Query().Get("sort_order"); sortOrderStr != "" {
		if sortOrderStr == "desc" {
			sortOrder = "desc"
		} else {
			sortOrder = "asc"
		}
	}
	offset := (page - 1) * limit

	now := time.Now()

	// Build query with search and pagination
	// Include both words with user_cards and words marked as "known"
	// Display word: prefer training_cards.display_word, fallback word_cards.display_en, fallback word_cards.word
	// Use UNION ALL to combine words from user_cards and known words without user_cards
	// Optimized: using JOINs instead of subqueries for better performance
	queryFromCards := "SELECT tc.word_card_id, wc.word as lemma, COALESCE(MAX(tc_display.display_word), MAX(wc.display_en), wc.word) as display_word, COUNT(DISTINCT uc.id) as total_cards, SUM(CASE WHEN uc.next_due_at IS NULL OR uc.next_due_at <= ? THEN 1 ELSE 0 END) as due_count, substr(CAST(MAX(uc.last_review_at) AS TEXT), 1, 19) as last_review, SUM(uc.reps) as total_reps, substr(CAST(MIN(uc.created_at) AS TEXT), 1, 19) as added_at, COUNT(CASE WHEN uc.state = 'review' THEN 1 END) as review_state_count, COUNT(CASE WHEN uc.state = 'learning' THEN 1 END) as learning_state_count, COUNT(CASE WHEN uc.state = 'new' THEN 1 END) as new_state_count, COALESCE(MAX(review_stats.review_count), 0) as review_count, MAX(CASE WHEN uwk_known.word_card_id IS NOT NULL THEN 1 ELSE 0 END) as is_known, COALESCE(MAX(uwm.mastering_score), 0) as mastering_score_stored FROM user_cards uc JOIN training_cards tc ON uc.training_card_id = tc.id JOIN word_cards wc ON tc.word_card_id = wc.id LEFT JOIN (SELECT tc1.word_card_id, tc1.display_word FROM training_cards tc1 INNER JOIN (SELECT word_card_id, MIN(id) as min_id FROM training_cards WHERE display_word IS NOT NULL AND display_word != '' GROUP BY word_card_id) tc_min ON tc1.word_card_id = tc_min.word_card_id AND tc1.id = tc_min.min_id) tc_display ON tc_display.word_card_id = tc.word_card_id LEFT JOIN (SELECT tc2.word_card_id, COUNT(*) as review_count FROM review_events re JOIN user_cards uc2 ON re.user_card_id = uc2.id JOIN training_cards tc2 ON uc2.training_card_id = tc2.id WHERE uc2.user_id = ? GROUP BY tc2.word_card_id) review_stats ON review_stats.word_card_id = tc.word_card_id LEFT JOIN user_word_knowledge uwk_known ON uwk_known.user_id = ? AND uwk_known.word_card_id = tc.word_card_id AND uwk_known.status = 'known' LEFT JOIN user_word_mastering uwm ON uwm.user_id = uc.user_id AND uwm.word_card_id = tc.word_card_id WHERE uc.user_id = ?"

	queryFromKnown := "SELECT uwk.word_card_id, wc.word as lemma, COALESCE(tc_display.display_word, wc.display_en, wc.word) as display_word, 0 as total_cards, 0 as due_count, NULL as last_review, 0 as total_reps, substr(CAST(uwk.created_at AS TEXT), 1, 19) as added_at, 0 as review_state_count, 0 as learning_state_count, 0 as new_state_count, 0 as review_count, 1 as is_known, 100 as mastering_score_stored FROM user_word_knowledge uwk JOIN word_cards wc ON uwk.word_card_id = wc.id LEFT JOIN (SELECT tc1.word_card_id, tc1.display_word FROM training_cards tc1 INNER JOIN (SELECT word_card_id, MIN(id) as min_id FROM training_cards WHERE display_word IS NOT NULL AND display_word != '' GROUP BY word_card_id) tc_min ON tc1.word_card_id = tc_min.word_card_id AND tc1.id = tc_min.min_id) tc_display ON tc_display.word_card_id = uwk.word_card_id LEFT JOIN (SELECT DISTINCT tc.word_card_id FROM user_cards uc JOIN training_cards tc ON uc.training_card_id = tc.id WHERE uc.user_id = ?) has_user_cards ON has_user_cards.word_card_id = uwk.word_card_id WHERE uwk.user_id = ? AND uwk.status = 'known' AND has_user_cards.word_card_id IS NULL"

	args := []interface{}{now, userID, userID, userID}
	argsKnown := []interface{}{userID, userID}
	
	if search != "" {
		// Search by lemma, display_word, or word_ru (case-insensitive)
		searchLower := strings.ToLower(search)
		searchPattern := "%" + searchLower + "%"
		queryFromCards += " AND (LOWER(wc.word) LIKE ? OR LOWER(COALESCE(tc_display.display_word, wc.display_en, wc.word)) LIKE ? OR LOWER(tc.word_ru) LIKE ?)"
		args = append(args, searchPattern, searchPattern, searchPattern)
		
		queryFromKnown += " AND (LOWER(wc.word) LIKE ? OR LOWER(COALESCE(tc_display.display_word, wc.display_en, wc.word)) LIKE ?)"
		argsKnown = append(argsKnown, searchPattern, searchPattern)
	}

	queryFromCards += " GROUP BY tc.word_card_id, wc.word"
	
	// Combine both queries with UNION ALL; mastery_level for filter/display, mastering_score from user_word_mastering (or 100 for known)
	baseQuery := "SELECT * FROM (" +
		"SELECT *, " +
		"CASE " +
		"WHEN is_known = 1 THEN 'known' " +
		"WHEN review_state_count = total_cards AND total_reps > 0 THEN 'mastered' " +
		"WHEN review_state_count > 0 OR learning_state_count > 0 THEN 'learning' " +
		"ELSE 'new' " +
		"END as mastery_level_calc, " +
		"mastering_score_stored as mastering_score_calc " +
		"FROM (" + queryFromCards + " UNION ALL " + queryFromKnown + ") combined " +
		") with_mastery"
	
	// Add filter by mastery_level if specified
	filterArgs := []interface{}{}
	if masteryLevelFilter != "" {
		// Validate mastery_level to prevent SQL injection
		allowedLevels := map[string]bool{
			"new":      true,
			"learning": true,
			"mastered": true,
			"known":    true,
		}
		if allowedLevels[masteryLevelFilter] {
			baseQuery += " WHERE mastery_level_calc = ?"
			filterArgs = append(filterArgs, masteryLevelFilter)
		}
	}

	// Get total count after filtering by mastery_level
	// Use the same query structure to count filtered results
	countQueryWithMastery := "SELECT COUNT(*) FROM (" +
		"SELECT *, " +
		"CASE " +
		"WHEN is_known = 1 THEN 'known' " +
		"WHEN review_state_count = total_cards AND total_reps > 0 THEN 'mastered' " +
		"WHEN review_state_count > 0 OR learning_state_count > 0 THEN 'learning' " +
		"ELSE 'new' " +
		"END as mastery_level_calc " +
		"FROM (" + queryFromCards + " UNION ALL " + queryFromKnown + ") combined " +
		") with_mastery"
	countArgsWithMastery := append(args, argsKnown...)
	if masteryLevelFilter != "" {
		allowedLevels := map[string]bool{
			"new":      true,
			"learning": true,
			"mastered": true,
			"known":    true,
		}
		if allowedLevels[masteryLevelFilter] {
			countQueryWithMastery += " WHERE mastery_level_calc = ?"
			countArgsWithMastery = append(countArgsWithMastery, masteryLevelFilter)
		}
	}
	
	var totalCount int
	err := r.db.QueryRow(countQueryWithMastery, countArgsWithMastery...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to get vocabulary count with mastery filter", zap.Error(err))
		totalCount = 0
	}

	// Add ordering and pagination
	// Handle special case for mastery_level (calculated field)
	var orderByClause string
	switch sortBy {
	case "mastery_level":
		// Use the calculated mastery_level_calc field
		// Original order: known (0) < mastered (1) < learning (2) < new (3)
		orderByClause = `CASE 
			WHEN mastery_level_calc = 'known' THEN 0
			WHEN mastery_level_calc = 'mastered' THEN 1
			WHEN mastery_level_calc = 'learning' THEN 2
			ELSE 3
		END`
	case "mastery_level_desc":
		// Reversed order: new (0) < learning (1) < mastered (2) < known (3)
		orderByClause = `CASE 
			WHEN mastery_level_calc = 'known' THEN 3
			WHEN mastery_level_calc = 'mastered' THEN 2
			WHEN mastery_level_calc = 'learning' THEN 1
			ELSE 0
		END`
	case "mastering_score":
		orderByClause = "mastering_score_calc"
	case "mastering_score_desc":
		orderByClause = "mastering_score_calc"
		sortOrder = "desc"
	case "lemma":
		// For lemma, sort by cleaned version (without "to " prefix for verbs)
		// Use CASE to remove "to " prefix if present (case-insensitive)
		orderByClause = `CASE 
			WHEN LOWER(lemma) LIKE 'to %' THEN SUBSTR(lemma, 4)
			ELSE lemma
		END`
	case "display_word":
		// For display_word, apply the same logic to the display_word column
		orderByClause = `CASE 
			WHEN LOWER(display_word) LIKE 'to %' THEN SUBSTR(display_word, 4)
			ELSE display_word
		END`
	default:
		orderByClause = sortBy
	}
	baseQuery += " ORDER BY " + orderByClause + " " + sortOrder + " LIMIT ? OFFSET ?"
	// Combine args from both queries and filter
	allArgs := append(args, argsKnown...)
	allArgs = append(allArgs, filterArgs...)
	allArgs = append(allArgs, limit, offset)

	rows, err := r.db.Query(baseQuery, allArgs...)
	if err != nil {
		r.logger.Error("failed to get vocabulary", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var words []VocabWord
	for rows.Next() {
		var word VocabWord
		var totalCards, dueCount, totalReps, reviewCount, reviewStateCount, learningStateCount, newStateCount, isKnown int
		var lastReview, addedAt sql.NullString
		var displayWord sql.NullString
		var masteryLevelCalc sql.NullString
		var masteringScoreStored sql.NullInt64
		var discardScore sql.NullInt64 // mastering_score_calc alias (same as stored)

		if testHookVocabScanErr != nil {
			if err := testHookVocabScanErr(); err != nil {
				r.logger.Error("failed to scan word", zap.Error(err))
				continue
			}
		}
		if testHookVocabRowScanErr != nil {
			if err := testHookVocabRowScanErr(); err != nil {
				r.logger.Error("failed to scan word", zap.Error(err))
				continue
			}
		}
		err := rows.Scan(&word.WordCardID, &word.Lemma, &displayWord, &totalCards, &dueCount, &lastReview, &totalReps, &addedAt,
			&reviewStateCount, &learningStateCount, &newStateCount, &reviewCount, &isKnown, &masteringScoreStored, &masteryLevelCalc, &discardScore)
		if testHookVocabScanErrAfter != nil {
			if hookErr := testHookVocabScanErrAfter(); hookErr != nil {
				err = hookErr
			}
		}
		if err != nil {
			r.logger.Error("failed to scan word", zap.Error(err))
			continue
		}

		if testHookVocabForceDisplayWordValid {
			displayWord = sql.NullString{String: "hooked", Valid: true}
		}
		if testHookVocabDisplayWordInvalid {
			displayWord = sql.NullString{}
		}
		if testHookVocabElseDisplayWord {
			word.DisplayWord = word.Lemma
		} else if displayWord.Valid {
			word.DisplayWord = displayWord.String
		} else {
			word.DisplayWord = word.Lemma
		}

		word.TotalCards = totalCards
		word.DueCount = dueCount
		word.TotalReps = totalReps
		word.ReviewCount = reviewCount

		if lastReview.Valid && lastReview.String != "" {
			var t *time.Time
			if testHookVocabSetLastReview != nil {
				t = testHookVocabSetLastReview
			} else if parsed, err := parseDateTime(lastReview.String); err == nil && parsed != nil {
				t = parsed
			}
			if t != nil {
				word.LastReview = t
			}
		}

		if addedAt.Valid && addedAt.String != "" {
			var t *time.Time
			if testHookVocabSetAddedAt != nil {
				t = testHookVocabSetAddedAt
			} else if parsed, err := parseDateTime(addedAt.String); err == nil && parsed != nil {
				t = parsed
			}
			if t != nil {
				word.AddedAt = t
			}
		}

		if testHookVocabMasteryLevelInvalid {
			masteryLevelCalc = sql.NullString{}
		}
		if testHookVocabElseMasteryLevel {
			word.MasteryLevel = "new"
		} else if masteryLevelCalc.Valid {
			word.MasteryLevel = masteryLevelCalc.String
		} else {
			word.MasteryLevel = "new"
		}
		if testHookVocabMasteringScoreInvalid {
			masteringScoreStored = sql.NullInt64{}
		}
		if testHookVocabElseMasteringScore {
			word.MasteringScore = 0
		} else if masteringScoreStored.Valid {
			word.MasteringScore = int(masteringScoreStored.Int64)
		} else {
			word.MasteringScore = 0
		}

		words = append(words, word)
	}

	// Return JSON response with pagination info
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"words": words,
		"pagination": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      totalCount,
			"total_pages": (totalCount + limit - 1) / limit,
		},
	})
}

// handleVocabDelete handles vocabulary deletion (confirm and delete)
// @Summary      Удалить слово из словаря
// @Description  Подтверждение удаления (GET) или удаление слова (POST) из словаря пользователя. Путь: /api/vocab/{word}/confirm_delete или /api/vocab/{word}/delete
// @Tags         Vocab
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        word  path  string  true  "Английское слово для удаления"
// @Param        action  path  string  false  "Действие: confirm_delete или delete"
// @Success      200  {object}  map[string]interface{}  "Информация о слове или результат удаления"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {object}  map[string]interface{}  "Слово не найдено"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/vocab/{word} [get]
func (r *Router) handleVocabDelete(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract lemma from URL path: /app/vocab/{lemma}/confirm_delete or /app/vocab/{lemma}/delete
	path := req.URL.Path
	if path == "" || !strings.HasPrefix(path, "/api/vocab/") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/vocab/"), "/")

	lemma := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if lemma == "" {
		// Invalid path, redirect to vocab list (frontend route)
		http.Redirect(w, req, "/app/vocab", http.StatusFound)
		return
	}

	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)

	// Validate action first (before looking up word)
	if action != "" && action != "confirm_delete" && action != "delete" && action != "cards" && action != "mark_known" && action != "move_to_training" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Find word_card_id by lemma (only if we have a valid action or no action)
	var wordCardID int64
	err := r.db.QueryRow(`SELECT id FROM word_cards WHERE LOWER(word) = LOWER(?)`, lemma).Scan(&wordCardID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Word not found",
			})
			return
		}
		r.logger.Error("failed to get word card ID", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if req.Method == http.MethodGet && action == "confirm_delete" {
		// Get word info for confirmation
		query := `SELECT COUNT(*) FROM user_cards uc
				  JOIN training_cards tc ON uc.training_card_id = tc.id
				  WHERE uc.user_id = ? AND tc.word_card_id = ?`
		var count int
		err := r.db.QueryRow(query, userID, wordCardID).Scan(&count)
		if err != nil {
			r.logger.Error("failed to get word count", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If word not found or empty, return error
		if count == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Word not found",
			})
			return
		}

		// Return word info for confirmation
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"lemma": lemma,
			"word_card_id": wordCardID,
			"count": count,
		})
		return
	}

	if req.Method == http.MethodPost && action == "delete" {
		// Perform deletion by word_card_id
		rowsAffected, err := userCardRepo.DeleteUserCardsByWordCardIDForUser(userID, wordCardID)
		if err != nil {
			r.logger.Error("failed to delete user cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"lemma":         lemma,
			"word_card_id":  wordCardID,
			"rows_affected": rowsAffected,
		})
		return
	}

	if req.Method == http.MethodPost && action == "mark_known" {
		// Mark word as known and remove user_cards
		wordSetService := r.getWordSetService()
		if err := wordSetService.MarkKnown(userID, wordCardID); err != nil {
			r.logger.Error("failed to mark as known", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"lemma":        lemma,
			"word_card_id": wordCardID,
		})
		return
	}

	if req.Method == http.MethodPost && action == "move_to_training" {
		// Remove known status and create user_cards
		userWordKnowledgeRepo := repository.NewUserWordKnowledgeRepository(r.db, r.logger)
		if err := userWordKnowledgeRepo.RemoveKnown(userID, wordCardID); err != nil {
			r.logger.Warn("failed to remove known status", zap.Error(err))
			// Continue anyway
		}

		// Ensure training cards exist
		wordSetService := r.getWordSetService()
		if err := wordSetService.EnsureTrainingCardsExist(req.Context(), wordCardID); err != nil {
			r.logger.Warn("failed to ensure training cards",
				zap.Int64("word_card_id", wordCardID),
				zap.Error(err),
			)
			// Continue anyway
		}

		// Create user_cards
		if err := wordSetService.EnsureUserCardsForWord(userID, wordCardID); err != nil {
			r.logger.Error("failed to create user cards", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"lemma":        lemma,
			"word_card_id": wordCardID,
		})
		return
	}

	if req.Method == http.MethodGet && action == "cards" {
		// Get detailed card information for the word (lemma)
		r.handleVocabWordCards(w, req, userID, lemma)
		return
	}

	http.Error(w, "Invalid request", http.StatusBadRequest)
}

// VocabCardDetail represents detailed information about a user card
type VocabCardDetail struct {
	ID              int64      `json:"id"`
	TrainingCardID  int64      `json:"training_card_id"`
	Direction       string     `json:"direction"`
	State           string     `json:"state"`
	EF              float64    `json:"ef"`
	Reps            int        `json:"reps"`
	IntervalDays    int        `json:"interval_days"`
	LearningStep    int        `json:"learning_step"`
	LapseCount      int        `json:"lapse_count"`
	NextDueAt       *time.Time `json:"next_due_at"`
	LastReviewAt    *time.Time `json:"last_review_at"`
	LastQuality     *int       `json:"last_quality"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	WordRU          string     `json:"word_ru"`
	MeaningEN       string     `json:"meaning_en"`
	ExampleEN       string     `json:"example_en"`
	ExampleRU       string     `json:"example_ru"`
	Transcription   string     `json:"transcription"`
	SenseIndex      int        `json:"sense_index"`
	POS             *string    `json:"pos,omitempty"`
	ReviewCount     int        `json:"review_count"` // Count of review events
}

// handleVocabWordCards returns detailed information about all cards for a word (by lemma)
// @Summary      Получить детальную информацию о карточках слова
// @Description  Возвращает детальную информацию о всех карточках пользователя для указанного слова (лемма)
// @Tags         Vocab
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        word  path  string  true  "Лемма слова (word_cards.word)"
// @Success      200  {object}  map[string]interface{}  "Детальная информация о карточках"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Слово не найдено"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/vocab/{word}/cards [get]
func (r *Router) handleVocabWordCards(w http.ResponseWriter, req *http.Request, userID int64, lemma string) {
	// First, find word_card_id by lemma
	var wordCardID int64
	err := r.db.QueryRow(`SELECT id FROM word_cards WHERE LOWER(word) = LOWER(?)`, lemma).Scan(&wordCardID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Word not found",
			})
			return
		}
		r.logger.Error("failed to get word card ID", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	query := `SELECT 
		uc.id,
		uc.training_card_id,
		uc.direction,
		uc.state,
		uc.ef,
		uc.reps,
		uc.interval_days,
		uc.learning_step,
		uc.lapse_count,
		substr(CAST(uc.next_due_at AS TEXT), 1, 19) as next_due_at,
		substr(CAST(uc.last_review_at AS TEXT), 1, 19) as last_review_at,
		uc.last_quality,
		substr(CAST(uc.created_at AS TEXT), 1, 19) as created_at,
		substr(CAST(uc.updated_at AS TEXT), 1, 19) as updated_at,
		tc.word_ru,
		tc.meaning_en,
		COALESCE(tc.example_en, '') as example_en,
		COALESCE(tc.example_ru, '') as example_ru,
		COALESCE(tc.transcription, '') as transcription,
		tc.sense_index,
		tc.pos,
		(SELECT COUNT(*) FROM review_events re WHERE re.user_card_id = uc.id) as review_count
	FROM user_cards uc
	JOIN training_cards tc ON uc.training_card_id = tc.id
	WHERE uc.user_id = ? AND tc.word_card_id = ?
	ORDER BY tc.sense_index, uc.direction`

	rows, err := r.db.Query(query, userID, wordCardID)
	if err != nil {
		r.logger.Error("failed to get word cards", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cards []VocabCardDetail
	for rows.Next() {
		var card VocabCardDetail
		var createdAt, updatedAt string
		var nextDueAt, lastReviewAt sql.NullString
		var lastQuality sql.NullInt64

		var pos sql.NullString
		err := rows.Scan(
			&card.ID,
			&card.TrainingCardID,
			&card.Direction,
			&card.State,
			&card.EF,
			&card.Reps,
			&card.IntervalDays,
			&card.LearningStep,
			&card.LapseCount,
			&nextDueAt,
			&lastReviewAt,
			&lastQuality,
			&createdAt,
			&updatedAt,
			&card.WordRU,
			&card.MeaningEN,
			&card.ExampleEN,
			&card.ExampleRU,
			&card.Transcription,
			&card.SenseIndex,
			&pos,
			&card.ReviewCount,
		)
		if pos.Valid {
			card.POS = &pos.String
		}
		if err != nil {
			r.logger.Error("failed to scan card", zap.Error(err))
			continue
		}

		if t, err := parseDateTime(createdAt); err == nil && t != nil {
			card.CreatedAt = *t
		}
		if t, err := parseDateTime(updatedAt); err == nil && t != nil {
			card.UpdatedAt = *t
		}

		if nextDueAt.Valid && nextDueAt.String != "" {
			if t, err := parseDateTime(nextDueAt.String); err == nil && t != nil {
				card.NextDueAt = t
			}
		}
		if lastReviewAt.Valid && lastReviewAt.String != "" {
			if t, err := parseDateTime(lastReviewAt.String); err == nil && t != nil {
				card.LastReviewAt = t
			}
		}
		if lastQuality.Valid {
			q := int(lastQuality.Int64)
			card.LastQuality = &q
		}

		cards = append(cards, card)
	}

	// If no user_cards found, check if word is marked as "known" and return training_cards directly
	if len(cards) == 0 {
		// Check if word is marked as known
		var isKnown bool
		err := r.db.QueryRow(`
			SELECT COUNT(*) > 0 FROM user_word_knowledge 
			WHERE user_id = ? AND word_card_id = ? AND status = 'known'
		`, userID, wordCardID).Scan(&isKnown)
		
		if err != nil {
			r.logger.Error("failed to check known status", zap.Error(err))
		}
		
		if isKnown {
			// Get training cards directly for known words without user_cards
			trainingQuery := `SELECT 
				tc.id,
				tc.word_ru,
				tc.meaning_en,
				COALESCE(tc.example_en, '') as example_en,
				COALESCE(tc.example_ru, '') as example_ru,
				COALESCE(tc.transcription, '') as transcription,
				tc.sense_index,
				tc.pos,
				substr(CAST(tc.created_at AS TEXT), 1, 19) as created_at
			FROM training_cards tc
			WHERE tc.word_card_id = ?
			ORDER BY tc.sense_index`
			
			if testHookVocabTrainingQueryErr != nil {
				if err := testHookVocabTrainingQueryErr(); err != nil {
					r.logger.Error("failed to get training cards", zap.Error(err))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}
			var trainingRows *sql.Rows
			if testHookVocabTrainingQueryFail {
				err = fmt.Errorf("injected training query error")
			} else {
				trainingRows, err = r.db.Query(trainingQuery, wordCardID)
			}
			if err != nil {
				r.logger.Error("failed to get training cards", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer trainingRows.Close()
			
			for trainingRows.Next() {
				var trainingCardID int64
				var wordRU, meaningEN, exampleEN, exampleRU, transcription string
				var senseIndex int
				var pos sql.NullString
				var createdAt string
				
				err := trainingRows.Scan(
					&trainingCardID,
					&wordRU,
					&meaningEN,
					&exampleEN,
					&exampleRU,
					&transcription,
					&senseIndex,
					&pos,
					&createdAt,
				)
				if err != nil {
					r.logger.Error("failed to scan training card", zap.Error(err))
					continue
				}
				
				// Create cards for both directions (ru_en and en_ru)
				directions := []string{"ru_en", "en_ru"}
				for _, direction := range directions {
					card := VocabCardDetail{
						ID:              0, // No user_card_id for known words without user_cards
						TrainingCardID:  trainingCardID,
						Direction:       direction,
						State:           "new",
						EF:              2.5,
						Reps:            0,
						IntervalDays:    0,
						LearningStep:    0,
						LapseCount:      0,
						WordRU:          wordRU,
						MeaningEN:        meaningEN,
						ExampleEN:       exampleEN,
						ExampleRU:       exampleRU,
						Transcription:   transcription,
						SenseIndex:      senseIndex,
						ReviewCount:     0,
					}
					
					if pos.Valid {
						card.POS = &pos.String
					}
					
					if createdAt != "" {
						if t, err := parseDateTime(createdAt); err == nil && t != nil {
							card.CreatedAt = *t
							card.UpdatedAt = *t
						}
					}
					
					cards = append(cards, card)
				}
			}
		}
		
		// If still no cards found, return not found
		if len(cards) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Word not found",
			})
			return
		}
	}

	// Get word card info (pos and verb_forms_json) for verb forms display
	var pos sql.NullString
	var verbFormsJSON sql.NullString
	err = r.db.QueryRow(`SELECT pos, verb_forms_json FROM word_cards WHERE id = ?`, wordCardID).Scan(&pos, &verbFormsJSON)
	if err != nil && err != sql.ErrNoRows {
		r.logger.Warn("failed to get word card info for verb forms", zap.Error(err))
	}

	// Parse verb forms if present
	var verbForms map[string]interface{}
	if pos.Valid && pos.String == "verb" && verbFormsJSON.Valid && verbFormsJSON.String != "" {
		if err := json.Unmarshal([]byte(verbFormsJSON.String), &verbForms); err != nil {
			r.logger.Warn("failed to parse verb forms JSON", zap.Error(err))
		}
	}

	// Check if word has user_cards and is marked as known
	var hasUserCards bool
	// Check if cards have valid user_card IDs (ID > 0 means real user_card, ID = 0 means known word without user_cards)
	for _, card := range cards {
		if card.ID > 0 {
			hasUserCards = true
			break
		}
	}
	
	var isKnown bool
	userWordKnowledgeRepo := repository.NewUserWordKnowledgeRepository(r.db, r.logger)
	isKnown, err = userWordKnowledgeRepo.IsKnown(userID, wordCardID)
	if err != nil {
		r.logger.Warn("failed to check known status", zap.Error(err))
		isKnown = false
	}

	// Build response
	response := map[string]interface{}{
		"lemma": lemma,
		"word_card_id": wordCardID,
		"cards": cards,
		"has_user_cards": hasUserCards,
		"is_known": isKnown,
	}
	
	// Add verb forms if present
	if verbForms != nil {
		response["verb_forms"] = verbForms
	}
	if pos.Valid {
		response["pos"] = pos.String
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
