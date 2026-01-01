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

// parseSQLiteTime parses SQLite datetime string
// SQLite stores Go time.Time.String() format: "2006-01-02 15:04:05.999999999 -0700 MST m=+123456.789012345"
// We use substr() in SQL to extract first 19 chars: "2006-01-02 15:04:05"
func parseSQLiteTime(timeStr string) (*time.Time, error) {
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
	
// VocabWord represents a word with statistics
type VocabWord struct {
	WordEN          string     `json:"word_en"`
	TotalCards      int        `json:"total_cards"`
	DueCount        int        `json:"due_count"`
	LastReview      *time.Time `json:"last_review"`
	TotalReps       int        `json:"total_reps"`        // Total number of reviews across all cards
	AddedAt         *time.Time `json:"added_at"`         // Date when first card was added
	MasteryLevel    string     `json:"mastery_level"`     // Calculated mastery level
	ReviewCount     int        `json:"review_count"`     // Total number of review events
}

// handleVocab shows the vocabulary list
// @Summary      Получить список слов
// @Description  Возвращает список всех слов пользователя с статистикой (общее количество карточек, количество готовых к повторению, дата последнего повторения). Поддерживает пагинацию и поиск.
// @Tags         Vocab
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        search  query  string  false  "Поиск по слову"
// @Param        page    query  int     false  "Номер страницы (начиная с 1)"
// @Param        limit   query  int     false  "Количество элементов на странице (по умолчанию 100)"
// @Success      200  {object}  map[string]interface{}  "Список слов с статистикой"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/vocab [get]
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
	page := 1
	limit := 100
	sortBy := "word_en"
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
			"word_en":       "tc.word_en",
			"total_cards":   "total_cards",
			"mastery_level": "mastery_level", // Special handling below
			"total_reps":    "total_reps",
			"review_count":  "review_count",
			"due_count":     "due_count",
			"added_at":      "added_at",
			"last_review":   "last_review",
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
	// Use substr() to extract first 19 chars (YYYY-MM-DD HH:MM:SS) from Go time.String() format
	baseQuery := `SELECT 
		tc.word_en,
		COUNT(DISTINCT uc.id) as total_cards,
		SUM(CASE WHEN uc.next_due_at IS NULL OR uc.next_due_at <= ? THEN 1 ELSE 0 END) as due_count,
		substr(MAX(uc.last_review_at), 1, 19) as last_review,
		SUM(uc.reps) as total_reps,
		substr(MIN(uc.created_at), 1, 19) as added_at,
		COUNT(CASE WHEN uc.state = 'review' THEN 1 END) as review_state_count,
		COUNT(CASE WHEN uc.state = 'learning' THEN 1 END) as learning_state_count,
		COUNT(CASE WHEN uc.state = 'new' THEN 1 END) as new_state_count,
		(SELECT COUNT(*) FROM review_events re 
		 JOIN user_cards uc2 ON re.user_card_id = uc2.id 
		 JOIN training_cards tc2 ON uc2.training_card_id = tc2.id 
		 WHERE tc2.word_en = tc.word_en AND uc2.user_id = ?) as review_count
	FROM user_cards uc
	JOIN training_cards tc ON uc.training_card_id = tc.id
	WHERE uc.user_id = ?`

	args := []interface{}{now, userID, userID}
	if search != "" {
		baseQuery += " AND (tc.word_en LIKE ? OR tc.word_ru LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	baseQuery += " GROUP BY tc.word_en"

	// Get total count for pagination (simpler query without subquery)
	countQuery := `SELECT COUNT(DISTINCT tc.word_en) as total
	FROM user_cards uc
	JOIN training_cards tc ON uc.training_card_id = tc.id
	WHERE uc.user_id = ?`
	countArgs := []interface{}{userID}
	if search != "" {
		countQuery += " AND (tc.word_en LIKE ? OR tc.word_ru LIKE ?)"
		searchPattern := "%" + search + "%"
		countArgs = append(countArgs, searchPattern, searchPattern)
	}
	
	var totalCount int
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to get vocabulary count", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add ordering and pagination
	// Handle special case for mastery_level (calculated field)
	orderByClause := sortBy
	if sortBy == "mastery_level" {
		// For mastery_level, we need to order by the calculated logic
		// Use the same logic as in the SELECT to calculate mastery
		// mastered (1) > learning (2) > new (3)
		orderByClause = `CASE 
			WHEN COUNT(CASE WHEN uc.state = 'review' THEN 1 END) = COUNT(DISTINCT uc.id) AND SUM(uc.reps) > 0 THEN 1
			WHEN COUNT(CASE WHEN uc.state = 'review' THEN 1 END) > 0 OR COUNT(CASE WHEN uc.state = 'learning' THEN 1 END) > 0 THEN 2
			ELSE 3
		END`
	}
	baseQuery += " ORDER BY " + orderByClause + " " + sortOrder + " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		r.logger.Error("failed to get vocabulary", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var words []VocabWord
	for rows.Next() {
		var word VocabWord
		var totalCards, dueCount, totalReps, reviewCount, reviewStateCount, learningStateCount, newStateCount int
		var lastReview, addedAt sql.NullString

		err := rows.Scan(&word.WordEN, &totalCards, &dueCount, &lastReview, &totalReps, &addedAt, 
			&reviewStateCount, &learningStateCount, &newStateCount, &reviewCount)
		if err != nil {
			r.logger.Error("failed to scan word", zap.Error(err))
			continue
		}

		word.TotalCards = totalCards
		word.DueCount = dueCount
		word.TotalReps = totalReps
		word.ReviewCount = reviewCount

		if lastReview.Valid && lastReview.String != "" {
			if t, err := parseSQLiteTime(lastReview.String); err == nil && t != nil {
				word.LastReview = t
			}
		}

		if addedAt.Valid && addedAt.String != "" {
			if t, err := parseSQLiteTime(addedAt.String); err == nil && t != nil {
				word.AddedAt = t
			}
		}

		// Determine mastery level based on state distribution and reps
		if reviewStateCount == totalCards && totalReps > 0 {
			word.MasteryLevel = "mastered"
		} else if reviewStateCount > 0 || learningStateCount > 0 {
			word.MasteryLevel = "learning"
		} else {
			word.MasteryLevel = "new"
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
// @Description  Подтверждение удаления (GET) или удаление слова (POST) из словаря пользователя. Путь: /app/vocab/{word}/confirm_delete или /app/vocab/{word}/delete
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
// @Router       /app/vocab/{word} [get]
func (r *Router) handleVocabDelete(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract word from URL path: /app/vocab/{word}/confirm_delete or /app/vocab/{word}/delete
	path := req.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/app/vocab/"), "/")
	if len(parts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	wordEN := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if wordEN == "" {
		// Invalid path, redirect to vocab list
		http.Redirect(w, req, "/app/vocab", http.StatusFound)
		return
	}

	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)

	if req.Method == http.MethodGet && action == "confirm_delete" {
		// Get word info for confirmation
		query := `SELECT COUNT(*) FROM user_cards uc
				  JOIN training_cards tc ON uc.training_card_id = tc.id
				  WHERE uc.user_id = ? AND tc.word_en = ?`
		var count int
		err := r.db.QueryRow(query, userID, wordEN).Scan(&count)
		if err != nil {
			r.logger.Error("failed to get word count", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// If word not found or empty, return error
		if count == 0 || wordEN == "" {
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
			"word_en": wordEN,
			"count":   count,
		})
		return
	}

	if req.Method == http.MethodPost && action == "delete" {
		// Perform deletion
		rowsAffected, err := userCardRepo.DeleteUserCardsByWordENForUser(userID, wordEN)
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
			"word_en":       wordEN,
			"rows_affected": rowsAffected,
		})
		return
	}

	if req.Method == http.MethodGet && action == "cards" {
		// Get detailed card information for the word
		r.handleVocabWordCards(w, req, userID, wordEN)
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
	ReviewCount     int        `json:"review_count"` // Count of review events
}

// handleVocabWordCards returns detailed information about all cards for a word
// @Summary      Получить детальную информацию о карточках слова
// @Description  Возвращает детальную информацию о всех карточках пользователя для указанного слова
// @Tags         Vocab
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        word  path  string  true  "Английское слово"
// @Success      200  {object}  map[string]interface{}  "Детальная информация о карточках"
// @Failure      400  {string}  string  "Неверный запрос"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Слово не найдено"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /app/vocab/{word}/cards [get]
func (r *Router) handleVocabWordCards(w http.ResponseWriter, req *http.Request, userID int64, wordEN string) {
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
		substr(uc.next_due_at, 1, 19) as next_due_at,
		substr(uc.last_review_at, 1, 19) as last_review_at,
		uc.last_quality,
		substr(uc.created_at, 1, 19) as created_at,
		substr(uc.updated_at, 1, 19) as updated_at,
		tc.word_ru,
		tc.meaning_en,
		COALESCE(tc.example_en, '') as example_en,
		COALESCE(tc.example_ru, '') as example_ru,
		COALESCE(tc.transcription, '') as transcription,
		tc.sense_index,
		(SELECT COUNT(*) FROM review_events re WHERE re.user_card_id = uc.id) as review_count
	FROM user_cards uc
	JOIN training_cards tc ON uc.training_card_id = tc.id
	WHERE uc.user_id = ? AND tc.word_en = ?
	ORDER BY tc.sense_index, uc.direction`

	rows, err := r.db.Query(query, userID, wordEN)
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
			&card.ReviewCount,
		)
		if err != nil {
			r.logger.Error("failed to scan card", zap.Error(err))
			continue
		}

		if t, err := parseSQLiteTime(createdAt); err == nil && t != nil {
			card.CreatedAt = *t
		}
		if t, err := parseSQLiteTime(updatedAt); err == nil && t != nil {
			card.UpdatedAt = *t
		}

		if nextDueAt.Valid && nextDueAt.String != "" {
			if t, err := parseSQLiteTime(nextDueAt.String); err == nil && t != nil {
				card.NextDueAt = t
			}
		}
		if lastReviewAt.Valid && lastReviewAt.String != "" {
			if t, err := parseSQLiteTime(lastReviewAt.String); err == nil && t != nil {
				card.LastReviewAt = t
			}
		}
		if lastQuality.Valid {
			q := int(lastQuality.Int64)
			card.LastQuality = &q
		}

		cards = append(cards, card)
	}

	if len(cards) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Word not found",
		})
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"word_en": wordEN,
		"cards":   cards,
	})
}

