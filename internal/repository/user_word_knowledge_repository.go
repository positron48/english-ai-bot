package repository

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// UserWordKnowledgeRepository handles database operations for user word knowledge
type UserWordKnowledgeRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserWordKnowledgeRepository creates a new user word knowledge repository
func NewUserWordKnowledgeRepository(db *sql.DB, logger *zap.Logger) *UserWordKnowledgeRepository {
	return &UserWordKnowledgeRepository{
		db:     db,
		logger: logger,
	}
}

// MarkKnown marks a word as known for a user
func (r *UserWordKnowledgeRepository) MarkKnown(userID, wordCardID int64) error {
	query := `INSERT OR REPLACE INTO user_word_knowledge (user_id, word_card_id, status, created_at)
			  VALUES (?, ?, 'known', CURRENT_TIMESTAMP)`

	_, err := r.db.Exec(query, userID, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to mark word as known: %w", err)
	}

	return nil
}

// RemoveKnown removes the known status for a word
func (r *UserWordKnowledgeRepository) RemoveKnown(userID, wordCardID int64) error {
	_, err := r.db.Exec(`DELETE FROM user_word_knowledge WHERE user_id = ? AND word_card_id = ?`, userID, wordCardID)
	if err != nil {
		return fmt.Errorf("failed to remove known status: %w", err)
	}

	return nil
}

// IsKnown checks if a word is marked as known
func (r *UserWordKnowledgeRepository) IsKnown(userID, wordCardID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM user_word_knowledge 
		WHERE user_id = ? AND word_card_id = ? AND status = 'known'
	`, userID, wordCardID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check known status: %w", err)
	}

	return count > 0, nil
}

// GetKnownWords retrieves all known words for a user
func (r *UserWordKnowledgeRepository) GetKnownWords(userID int64) ([]int64, error) {
	query := `SELECT word_card_id FROM user_word_knowledge WHERE user_id = ? AND status = 'known'`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get known words: %w", err)
	}
	defer rows.Close()

	var wordCardIDs []int64
	for rows.Next() {
		var wordCardID int64
		if err := rows.Scan(&wordCardID); err != nil {
			r.logger.Warn("failed to scan word card ID", zap.Error(err))
			continue
		}
		wordCardIDs = append(wordCardIDs, wordCardID)
	}

	return wordCardIDs, nil
}
