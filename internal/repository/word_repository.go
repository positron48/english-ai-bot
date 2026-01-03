package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// WordRepository handles database operations for word cards
type WordRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewWordRepository creates a new word repository
func NewWordRepository(db *sql.DB, logger *zap.Logger) *WordRepository {
	return &WordRepository{
		db:     db,
		logger: logger,
	}
}

// GetWordCard retrieves a word card by word
func (r *WordRepository) GetWordCard(word string) (*models.WordCard, error) {
	query := `SELECT id, word, definition, created_at, updated_at 
			  FROM word_cards 
			  WHERE LOWER(word) = LOWER(?)`

	var card models.WordCard
	var createdAt, updatedAt string

	err := r.db.QueryRow(query, word).Scan(
		&card.ID,
		&card.Word,
		&card.Definition,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get word card: %w", err)
	}

	card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	card.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &card, nil
}

// SaveWordCard saves a new word card or updates existing one
func (r *WordRepository) SaveWordCard(word, definition string) error {
	query := `INSERT INTO word_cards (word, definition, updated_at) 
			  VALUES (?, ?, CURRENT_TIMESTAMP)
			  ON CONFLICT(word) DO UPDATE SET 
			  	definition = excluded.definition,
			  	updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, word, definition)
	if err != nil {
		return fmt.Errorf("failed to save word card: %w", err)
	}

	r.logger.Debug("word card saved",
		zap.String("word", word),
	)

	return nil
}

// AddWordRequestHistory adds a history entry for a word request
func (r *WordRepository) AddWordRequestHistory(userID int64, word string) error {
	query := `INSERT INTO word_request_history (user_id, word, requested_at) 
			  VALUES (?, ?, CURRENT_TIMESTAMP)`

	_, err := r.db.Exec(query, userID, word)
	if err != nil {
		return fmt.Errorf("failed to add word request history: %w", err)
	}

	r.logger.Debug("word request history added",
		zap.Int64("user_id", userID),
		zap.String("word", word),
	)

	return nil
}

// GetUserIDsByWord gets telegram user IDs who requested a specific word
func (r *WordRepository) GetUserIDsByWord(word string) ([]int64, error) {
	query := `SELECT DISTINCT user_id FROM word_request_history WHERE LOWER(word) = LOWER(?)`
	
	rows, err := r.db.Query(query, word)
	if err != nil {
		return nil, fmt.Errorf("failed to query user IDs: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// MarkWordCardProcessedError marks a word card as processed with an error
func (r *WordRepository) MarkWordCardProcessedError(wordCardID int64, errorText string) error {
	query := `UPDATE word_cards 
			  SET processed_at = CURRENT_TIMESTAMP, 
			      processing_error = ?,
			      updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := r.db.Exec(query, errorText, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to mark word card as processed with error: %w", err)
	}

	r.logger.Debug("word card marked as processed with error",
		zap.Int64("word_card_id", wordCardID),
		zap.String("error", errorText),
	)

	return nil
}

// ResetWordCardProcessed resets the processed status of a word card
func (r *WordRepository) ResetWordCardProcessed(wordCardID int64) error {
	query := `UPDATE word_cards 
			  SET processed_at = NULL, 
			      processing_error = NULL,
			      updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := r.db.Exec(query, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to reset word card processed status: %w", err)
	}

	r.logger.Debug("word card processed status reset",
		zap.Int64("word_card_id", wordCardID),
	)

	return nil
}

// UpdateWordCardDefinition updates the definition of a word card
func (r *WordRepository) UpdateWordCardDefinition(wordCardID int64, definition string) error {
	query := `UPDATE word_cards 
			  SET definition = ?, 
			      updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := r.db.Exec(query, definition, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to update word card definition: %w", err)
	}

	r.logger.Debug("word card definition updated",
		zap.Int64("word_card_id", wordCardID),
	)

	return nil
}

// WordCardAdminItem represents a word card with additional info for admin view
type WordCardAdminItem struct {
	models.WordCard
	HasTrainingCards bool
	RequestingUsers  []int64
}

// ListWordCardsAdmin lists word cards for admin view with optional filters
func (r *WordRepository) ListWordCardsAdmin(filterUserID *int64, onlyWithErrors bool, limit, offset int) ([]*WordCardAdminItem, error) {
	// Use LEFT JOIN with GROUP BY to check for training cards - more reliable than subquery
	query := `SELECT wc.id, wc.word, wc.definition,
			  COALESCE(wc.processed_at, '') as processed_at,
			  COALESCE(wc.processing_error, '') as processing_error,
			  wc.created_at, wc.updated_at,
			  MAX(CASE WHEN tc.id IS NOT NULL THEN 1 ELSE 0 END) as has_training_cards
			  FROM word_cards wc
			  LEFT JOIN training_cards tc ON tc.word_card_id = wc.id`

	args := []interface{}{}
	conditions := []string{}

	// Filter by user if specified - use subquery to avoid duplicates from JOIN
	if filterUserID != nil {
		conditions = append(conditions, "wc.word IN (SELECT DISTINCT word FROM word_request_history WHERE user_id = ?)")
		args = append(args, *filterUserID)
	}

	// Filter by errors if specified
	if onlyWithErrors {
		conditions = append(conditions, "wc.processing_error IS NOT NULL AND wc.processing_error != ''")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY wc.id, wc.word, wc.definition, wc.processed_at, wc.processing_error, wc.created_at, wc.updated_at"
	query += " ORDER BY wc.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list word cards: %w", err)
	}
	defer rows.Close()

	var items []*WordCardAdminItem
	for rows.Next() {
		var item WordCardAdminItem
		var createdAt, updatedAt, processedAtStr, processingErrorStr string
		var hasTrainingCards int

		err := rows.Scan(&item.ID, &item.Word, &item.Definition,
			&processedAtStr, &processingErrorStr,
			&createdAt, &updatedAt, &hasTrainingCards)
		if err != nil {
			return nil, fmt.Errorf("failed to scan word card: %w", err)
		}

		item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		item.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		item.HasTrainingCards = hasTrainingCards > 0

		if processedAtStr != "" {
			processedAt, _ := time.Parse("2006-01-02 15:04:05", processedAtStr)
			item.ProcessedAt = &processedAt
		}
		if processingErrorStr != "" {
			item.ProcessingError = &processingErrorStr
		}

		// Get requesting users for this word
		userIDs, err := r.GetUserIDsByWord(item.Word)
		if err != nil {
			r.logger.Warn("failed to get requesting users", zap.Error(err), zap.String("word", item.Word))
		} else {
			item.RequestingUsers = userIDs
		}

		items = append(items, &item)
	}

	return items, nil
}

// GetWordCardRequestingUsers gets user IDs who requested a specific word card
func (r *WordRepository) GetWordCardRequestingUsers(wordCardID int64) ([]int64, error) {
	// First get the word
	var word string
	err := r.db.QueryRow(`SELECT word FROM word_cards WHERE id = ?`, wordCardID).Scan(&word)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get word: %w", err)
	}

	return r.GetUserIDsByWord(word)
}

// DeleteWordCard deletes a word card by ID (CASCADE will delete related training_cards, user_cards, and word_request_history)
func (r *WordRepository) DeleteWordCard(wordCardID int64) error {
	query := `DELETE FROM word_cards WHERE id = ?`
	result, err := r.db.Exec(query, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to delete word card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("word card not found")
	}

	r.logger.Info("deleted word card",
		zap.Int64("word_card_id", wordCardID),
	)

	return nil
}
