package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// WordSetRepository handles database operations for word sets
type WordSetRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewWordSetRepository creates a new word set repository
func NewWordSetRepository(db *sql.DB, logger *zap.Logger) *WordSetRepository {
	return &WordSetRepository{
		db:     db,
		logger: logger,
	}
}

// CreateWordSet creates a new word set
func (r *WordSetRepository) CreateWordSet(wordSet *models.WordSet) (int64, error) {
	query := `INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	var categoryID interface{}
	if wordSet.CategoryID != nil {
		categoryID = *wordSet.CategoryID
	}

	var description interface{}
	if wordSet.Description != nil {
		description = *wordSet.Description
	}

	var preferredPOS interface{}
	if wordSet.PreferredPOS != nil {
		preferredPOS = *wordSet.PreferredPOS
	}

	isPublished := 0
	if wordSet.IsPublished {
		isPublished = 1
	}

	result, err := r.db.Exec(query, categoryID, wordSet.Title, description, isPublished, wordSet.SortOrder, preferredPOS)
	if err != nil {
		return 0, fmt.Errorf("failed to create word set: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get word set ID: %w", err)
	}

	return id, nil
}

// GetWordSet retrieves a word set by ID
func (r *WordSetRepository) GetWordSet(id int64) (*models.WordSet, error) {
	query := `SELECT id, category_id, title, description, is_published, sort_order, preferred_pos, created_at, updated_at
			  FROM word_sets WHERE id = ?`

	var wordSet models.WordSet
	var createdAt, updatedAt string
	var categoryID sql.NullInt64
	var description sql.NullString
	var preferredPOS sql.NullString
	var isPublished int

	err := r.db.QueryRow(query, id).Scan(
		&wordSet.ID,
		&categoryID,
		&wordSet.Title,
		&description,
		&isPublished,
		&wordSet.SortOrder,
		&preferredPOS,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get word set: %w", err)
	}

	if categoryID.Valid {
		cid := int64(categoryID.Int64)
		wordSet.CategoryID = &cid
	}
	if description.Valid {
		wordSet.Description = &description.String
	}
	if preferredPOS.Valid {
		wordSet.PreferredPOS = &preferredPOS.String
	}
	wordSet.IsPublished = isPublished == 1

	wordSet.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	wordSet.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &wordSet, nil
}

// ListWordSets retrieves word sets with optional filters
// If includeUnpublished is true, returns all sets including unpublished (for admin)
func (r *WordSetRepository) ListWordSets(categoryID *int64, limit, offset int, includeUnpublished ...bool) ([]*models.WordSet, error) {
	includeUnpub := false
	if len(includeUnpublished) > 0 {
		includeUnpub = includeUnpublished[0]
	}

	query := `SELECT id, category_id, title, description, is_published, sort_order, preferred_pos, created_at, updated_at
			  FROM word_sets WHERE 1=1`
	args := []interface{}{}

	// categoryID can be:
	// - nil: no filter (show all categories)
	// - pointer to int64: filter by specific category_id
	// Note: To filter by category_id IS NULL, use a special value or modify this function
	if categoryID != nil {
		query += " AND category_id = ?"
		args = append(args, *categoryID)
	}

	if !includeUnpub {
		query += " AND is_published = 1"
	}

	// Sort by sort_order within category, then by title
	query += " ORDER BY sort_order, title LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list word sets: %w", err)
	}
	defer rows.Close()

	var wordSets []*models.WordSet
	for rows.Next() {
		var wordSet models.WordSet
		var createdAt, updatedAt string
		var categoryID sql.NullInt64
		var description sql.NullString
		var preferredPOS sql.NullString
		var isPublished int

		err := rows.Scan(
			&wordSet.ID,
			&categoryID,
			&wordSet.Title,
			&description,
			&isPublished,
			&wordSet.SortOrder,
			&preferredPOS,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			r.logger.Warn("failed to scan word set", zap.Error(err))
			continue
		}

		if categoryID.Valid {
			cid := int64(categoryID.Int64)
			wordSet.CategoryID = &cid
		}
		if description.Valid {
			wordSet.Description = &description.String
		}
		if preferredPOS.Valid {
			wordSet.PreferredPOS = &preferredPOS.String
		}
		wordSet.IsPublished = isPublished == 1

		wordSet.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		wordSet.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		wordSets = append(wordSets, &wordSet)
	}

	return wordSets, nil
}

// UpdateWordSet updates a word set
func (r *WordSetRepository) UpdateWordSet(wordSet *models.WordSet) error {
	query := `UPDATE word_sets 
			  SET category_id = ?, title = ?, description = ?, is_published = ?, sort_order = ?, preferred_pos = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	var categoryID interface{}
	if wordSet.CategoryID != nil {
		categoryID = *wordSet.CategoryID
	}

	var description interface{}
	if wordSet.Description != nil {
		description = *wordSet.Description
	}

	var preferredPOS interface{}
	if wordSet.PreferredPOS != nil {
		preferredPOS = *wordSet.PreferredPOS
	}

	isPublished := 0
	if wordSet.IsPublished {
		isPublished = 1
	}

	_, err := r.db.Exec(query, categoryID, wordSet.Title, description, isPublished, wordSet.SortOrder, preferredPOS, wordSet.ID)
	if err != nil {
		return fmt.Errorf("failed to update word set: %w", err)
	}

	return nil
}

// DeleteWordSet deletes a word set (cascade will delete items)
func (r *WordSetRepository) DeleteWordSet(id int64) error {
	_, err := r.db.Exec(`DELETE FROM word_sets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete word set: %w", err)
	}

	return nil
}

// GetWordSetProgress calculates progress for a word set for a user
// All words are counted regardless of preferred_pos
func (r *WordSetRepository) GetWordSetProgress(wordSetID, userID int64) (*models.WordSetWithProgress, error) {
	// Get word set
	wordSet, err := r.GetWordSet(wordSetID)
	if err != nil {
		return nil, err
	}
	if wordSet == nil {
		return nil, fmt.Errorf("word set not found")
	}

	// Count total words (all words, no filtering)
	var totalWords int
	err = r.db.QueryRow(`SELECT COUNT(*) FROM word_set_items WHERE word_set_id = ?`, wordSetID).Scan(&totalWords)
	if err != nil {
		return nil, fmt.Errorf("failed to count total words: %w", err)
	}

	// Count known words (in user_word_knowledge)
	var knownWords int
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT wsi.word_card_id)
		FROM word_set_items wsi
		INNER JOIN user_word_knowledge uwk ON wsi.word_card_id = uwk.word_card_id
		WHERE wsi.word_set_id = ? AND uwk.user_id = ? AND uwk.status = 'known'
	`, wordSetID, userID).Scan(&knownWords)
	if err != nil {
		return nil, fmt.Errorf("failed to count known words: %w", err)
	}

	// Count words in vocab (has user_cards but not known)
	var wordsInVocab int
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT wsi.word_card_id)
		FROM word_set_items wsi
		INNER JOIN training_cards tc ON wsi.word_card_id = tc.word_card_id
		INNER JOIN user_cards uc ON tc.id = uc.training_card_id
		WHERE wsi.word_set_id = ? 
		  AND uc.user_id = ?
		  AND NOT EXISTS (
		    SELECT 1 FROM user_word_knowledge uwk 
		    WHERE uwk.user_id = ? AND uwk.word_card_id = wsi.word_card_id AND uwk.status = 'known'
		  )
	`, wordSetID, userID, userID).Scan(&wordsInVocab)
	if err != nil {
		return nil, fmt.Errorf("failed to count words in vocab: %w", err)
	}

	// Unknown words = total - known - in_vocab
	unknownWords := totalWords - knownWords - wordsInVocab
	if unknownWords < 0 {
		unknownWords = 0
	}

	progressPercent := 0.0
	if totalWords > 0 {
		progressPercent = float64(knownWords+wordsInVocab) / float64(totalWords) * 100.0
	}

	return &models.WordSetWithProgress{
		WordSet:        *wordSet,
		TotalWords:     totalWords,
		KnownWords:     knownWords,
		WordsInVocab:   wordsInVocab,
		UnknownWords:   unknownWords,
		ProgressPercent: progressPercent,
	}, nil
}

// GetWordSetWords retrieves words in a set with their status for a user
// If the word set has preferred_pos set, data from matching training cards is included
// All words are returned regardless of whether they have matching training cards
func (r *WordSetRepository) GetWordSetWords(wordSetID, userID int64) ([]*models.WordSetWordInfo, error) {
	// First, get the word set to check preferred_pos
	wordSet, err := r.GetWordSet(wordSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get word set: %w", err)
	}
	if wordSet == nil {
		return nil, fmt.Errorf("word set not found")
	}

	// Build query - always return all words, but join with training cards if preferred_pos is set
	query := `
		SELECT 
			wc.id,
			wc.word,
			CASE
				WHEN EXISTS (SELECT 1 FROM user_word_knowledge uwk WHERE uwk.user_id = ? AND uwk.word_card_id = wc.id AND uwk.status = 'known') THEN 'known'
				WHEN EXISTS (SELECT 1 FROM user_cards uc INNER JOIN training_cards tc ON uc.training_card_id = tc.id WHERE uc.user_id = ? AND tc.word_card_id = wc.id) THEN 'in_vocab'
				ELSE 'unknown'
			END as status`

	args := []interface{}{userID, userID}

	// If preferred_pos is set, join with training cards to get data from matching card
	// Use subqueries instead of LEFT JOIN to avoid duplicates when multiple training_cards
	// exist for the same word_card_id with the same POS
	// Use case-insensitive comparison for POS (LOWER() for SQLite compatibility)
	if wordSet.PreferredPOS != nil && *wordSet.PreferredPOS != "" {
		query += `,
			COALESCE(
				(SELECT display_word FROM training_cards tc_pref WHERE tc_pref.word_card_id = wc.id AND LOWER(COALESCE(tc_pref.pos, '')) = LOWER(?) AND tc_pref.display_word IS NOT NULL AND tc_pref.display_word != '' LIMIT 1),
				(SELECT word_en FROM training_cards tc_pref WHERE tc_pref.word_card_id = wc.id AND LOWER(COALESCE(tc_pref.pos, '')) = LOWER(?) LIMIT 1),
				(SELECT display_word FROM training_cards tc2 WHERE tc2.word_card_id = wc.id AND tc2.display_word IS NOT NULL AND tc2.display_word != '' LIMIT 1),
				wc.display_en,
				wc.word
			) as display_word_pref,
			(SELECT transcription FROM training_cards tc_pref WHERE tc_pref.word_card_id = wc.id AND LOWER(COALESCE(tc_pref.pos, '')) = LOWER(?) LIMIT 1) as transcription_pref,
			(SELECT word_ru FROM training_cards tc_pref WHERE tc_pref.word_card_id = wc.id AND LOWER(COALESCE(tc_pref.pos, '')) = LOWER(?) LIMIT 1) as word_ru_pref,
			(SELECT meaning_en FROM training_cards tc_pref WHERE tc_pref.word_card_id = wc.id AND LOWER(COALESCE(tc_pref.pos, '')) = LOWER(?) LIMIT 1) as meaning_en_pref,
			(SELECT example_en FROM training_cards tc_pref WHERE tc_pref.word_card_id = wc.id AND LOWER(COALESCE(tc_pref.pos, '')) = LOWER(?) LIMIT 1) as example_en_pref,
			(SELECT example_ru FROM training_cards tc_pref WHERE tc_pref.word_card_id = wc.id AND LOWER(COALESCE(tc_pref.pos, '')) = LOWER(?) LIMIT 1) as example_ru_pref`
		query += `
		FROM word_set_items wsi
		INNER JOIN word_cards wc ON wsi.word_card_id = wc.id`
		// Add preferred_pos parameter multiple times (once for each subquery)
		// There are 7 subqueries that use preferred_pos: lines 343, 344, 349, 350, 351, 352, 353
		for i := 0; i < 7; i++ {
			args = append(args, *wordSet.PreferredPOS)
		}
	} else {
		query += `,
			COALESCE(
				(SELECT display_word FROM training_cards tc2 WHERE tc2.word_card_id = wc.id AND tc2.display_word IS NOT NULL AND tc2.display_word != '' LIMIT 1),
				wc.display_en,
				wc.word
			) as display_word_pref,
			NULL as transcription_pref,
			NULL as word_ru_pref,
			NULL as meaning_en_pref,
			NULL as example_en_pref,
			NULL as example_ru_pref`
		query += `
		FROM word_set_items wsi
		INNER JOIN word_cards wc ON wsi.word_card_id = wc.id`
	}

	query += `
		WHERE wsi.word_set_id = ?`

	args = append(args, wordSetID)

	query += `
		ORDER BY wsi.sort_order, wc.word`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get word set words: %w", err)
	}
	defer rows.Close()

	var words []*models.WordSetWordInfo
	for rows.Next() {
		var word models.WordSetWordInfo
		var displayWord sql.NullString
		var transcription sql.NullString
		var wordRU sql.NullString
		var meaningEN sql.NullString
		var exampleEN sql.NullString
		var exampleRU sql.NullString

		err := rows.Scan(
			&word.WordCardID,
			&word.Word,
			&word.Status,
			&displayWord,
			&transcription,
			&wordRU,
			&meaningEN,
			&exampleEN,
			&exampleRU,
		)
		if err != nil {
			r.logger.Warn("failed to scan word", zap.Error(err))
			continue
		}

		// Set display_word
		if displayWord.Valid && displayWord.String != "" {
			word.DisplayWord = displayWord.String
		} else {
			word.DisplayWord = word.Word
		}

		// Set training card data if available
		if transcription.Valid && transcription.String != "" {
			word.Transcription = &transcription.String
		}
		if wordRU.Valid && wordRU.String != "" {
			word.WordRU = &wordRU.String
		}
		if meaningEN.Valid && meaningEN.String != "" {
			word.MeaningEN = &meaningEN.String
		}
		if exampleEN.Valid && exampleEN.String != "" {
			word.ExampleEN = &exampleEN.String
		}
		if exampleRU.Valid && exampleRU.String != "" {
			word.ExampleRU = &exampleRU.String
		}

		words = append(words, &word)
	}

	return words, nil
}

// SetWordSetItems replaces all items in a word set
func (r *WordSetRepository) SetWordSetItems(wordSetID int64, wordCardIDs []int64) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing items
	_, err = tx.Exec(`DELETE FROM word_set_items WHERE word_set_id = ?`, wordSetID)
	if err != nil {
		return fmt.Errorf("failed to delete existing items: %w", err)
	}

	// Insert new items
	if len(wordCardIDs) > 0 {
		placeholders := strings.Repeat("(?, ?, ?),", len(wordCardIDs))
		placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

		query := fmt.Sprintf(`INSERT INTO word_set_items (word_set_id, word_card_id, sort_order) VALUES %s`, placeholders)
		args := make([]interface{}, 0, len(wordCardIDs)*3)
		for i, wordCardID := range wordCardIDs {
			args = append(args, wordSetID, wordCardID, i)
		}

		_, err = tx.Exec(query, args...)
		if err != nil {
			return fmt.Errorf("failed to insert items: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
