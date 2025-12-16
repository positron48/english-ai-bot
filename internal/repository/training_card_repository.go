package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// TrainingCardRepository handles database operations for training cards
type TrainingCardRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewTrainingCardRepository creates a new training card repository
func NewTrainingCardRepository(db *sql.DB, logger *zap.Logger) *TrainingCardRepository {
	return &TrainingCardRepository{
		db:     db,
		logger: logger,
	}
}

// CreateTrainingCard creates a new training card
func (r *TrainingCardRepository) CreateTrainingCard(card *models.TrainingCard) (int64, error) {
	query := `INSERT INTO training_cards (
		word_card_id, word_en, transcription, sense_index,
		meaning_ru, meaning_en, example_en, example_ru,
		distractors_ru, distractors_en, hint
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		card.WordCardID, card.WordEN, card.Transcription, card.SenseIndex,
		card.MeaningRU, card.MeaningEN, card.ExampleEN, card.ExampleRU,
		card.DistractorsRU, card.DistractorsEN, card.Hint,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create training card: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get training card ID: %w", err)
	}

	r.logger.Debug("created training card",
		zap.Int64("id", id),
		zap.String("word", card.WordEN),
		zap.Int("sense_index", card.SenseIndex),
	)

	return id, nil
}

// GetTrainingCard gets a training card by ID
func (r *TrainingCardRepository) GetTrainingCard(id int64) (*models.TrainingCard, error) {
	query := `SELECT id, word_card_id, word_en, COALESCE(transcription, ''), sense_index,
			  meaning_ru, meaning_en, COALESCE(example_en, ''), COALESCE(example_ru, ''),
			  COALESCE(distractors_ru, ''), COALESCE(distractors_en, ''), COALESCE(hint, ''),
			  created_at
			  FROM training_cards WHERE id = ?`

	var card models.TrainingCard
	var createdAt string

	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.WordCardID, &card.WordEN, &card.Transcription, &card.SenseIndex,
		&card.MeaningRU, &card.MeaningEN, &card.ExampleEN, &card.ExampleRU,
		&card.DistractorsRU, &card.DistractorsEN, &card.Hint,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get training card: %w", err)
	}

	card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)

	return &card, nil
}

// GetTrainingCardsByWordCardID gets all training cards for a word card
func (r *TrainingCardRepository) GetTrainingCardsByWordCardID(wordCardID int64) ([]*models.TrainingCard, error) {
	query := `SELECT id, word_card_id, word_en, COALESCE(transcription, ''), sense_index,
			  meaning_ru, meaning_en, COALESCE(example_en, ''), COALESCE(example_ru, ''),
			  COALESCE(distractors_ru, ''), COALESCE(distractors_en, ''), COALESCE(hint, ''),
			  created_at
			  FROM training_cards WHERE word_card_id = ? ORDER BY sense_index`

	rows, err := r.db.Query(query, wordCardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get training cards: %w", err)
	}
	defer rows.Close()

	var cards []*models.TrainingCard
	for rows.Next() {
		var card models.TrainingCard
		var createdAt string

		err := rows.Scan(
			&card.ID, &card.WordCardID, &card.WordEN, &card.Transcription, &card.SenseIndex,
			&card.MeaningRU, &card.MeaningEN, &card.ExampleEN, &card.ExampleRU,
			&card.DistractorsRU, &card.DistractorsEN, &card.Hint,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan training card: %w", err)
		}

		card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		cards = append(cards, &card)
	}

	return cards, nil
}

// GetWordCardsWithoutTrainingCards gets word cards that don't have training cards yet
func (r *TrainingCardRepository) GetWordCardsWithoutTrainingCards(limit int) ([]*models.WordCard, error) {
	query := `SELECT wc.id, wc.word, wc.definition, wc.created_at, wc.updated_at
			  FROM word_cards wc
			  LEFT JOIN training_cards tc ON wc.id = tc.word_card_id
			  WHERE tc.id IS NULL
			  ORDER BY wc.created_at
			  LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get word cards without training cards: %w", err)
	}
	defer rows.Close()

	var cards []*models.WordCard
	for rows.Next() {
		var card models.WordCard
		var createdAt, updatedAt string

		err := rows.Scan(&card.ID, &card.Word, &card.Definition, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan word card: %w", err)
		}

		card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		card.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		cards = append(cards, &card)
	}

	return cards, nil
}

// HasTrainingCards checks if a word card has training cards
func (r *TrainingCardRepository) HasTrainingCards(wordCardID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM training_cards WHERE word_card_id = ?`
	var count int
	err := r.db.QueryRow(query, wordCardID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check training cards: %w", err)
	}
	return count > 0, nil
}

// DeleteTrainingCardsByWordEN deletes all training cards for a word by word_en
func (r *TrainingCardRepository) DeleteTrainingCardsByWordEN(wordEN string) (int64, error) {
	query := `DELETE FROM training_cards WHERE word_en = ?`
	result, err := r.db.Exec(query, wordEN)
	if err != nil {
		return 0, fmt.Errorf("failed to delete training cards: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("deleted training cards",
		zap.String("word_en", wordEN),
		zap.Int64("rows_affected", rowsAffected),
	)

	return rowsAffected, nil
}

