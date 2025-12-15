package repository

import (
	"database/sql"
	"fmt"
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
