package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// UserCardRepository handles database operations for user cards
type UserCardRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserCardRepository creates a new user card repository
func NewUserCardRepository(db *sql.DB, logger *zap.Logger) *UserCardRepository {
	return &UserCardRepository{
		db:     db,
		logger: logger,
	}
}

// CreateUserCard creates a new user card
// If a card with the same user_id, training_card_id and direction already exists, returns its ID
func (r *UserCardRepository) CreateUserCard(card *models.UserCard) (int64, error) {
	// Check if user card already exists
	existing, err := r.GetUserCardByTrainingCard(card.UserID, card.TrainingCardID, card.Direction)
	if err != nil {
		return 0, fmt.Errorf("failed to check existing user card: %w", err)
	}
	if existing != nil {
		r.logger.Debug("user card already exists, returning existing ID",
			zap.Int64("id", existing.ID),
			zap.Int64("user_id", card.UserID),
			zap.Int64("training_card_id", card.TrainingCardID),
			zap.String("direction", string(card.Direction)),
		)
		return existing.ID, nil
	}

	query := `INSERT INTO user_cards (
		user_id, training_card_id, direction, state, ef, reps, interval_days,
		learning_step, lapse_count, next_due_at, last_review_at, last_quality,
		last_options_json, wrong_answers_json, stats_json
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		card.UserID, card.TrainingCardID, card.Direction, card.State, card.EF,
		card.Reps, card.IntervalDays, card.LearningStep, card.LapseCount,
		card.NextDueAt, card.LastReviewAt, card.LastQuality,
		card.LastOptionsJSON, card.WrongAnswersJSON, card.StatsJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user card: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get user card ID: %w", err)
	}

	r.logger.Debug("created user card",
		zap.Int64("id", id),
		zap.Int64("user_id", card.UserID),
		zap.String("direction", string(card.Direction)),
	)

	return id, nil
}

// GetUserCard gets a user card by ID
func (r *UserCardRepository) GetUserCard(id int64) (*models.UserCard, error) {
	query := `SELECT id, user_id, training_card_id, direction, state, ef, reps,
			  interval_days, learning_step, lapse_count, next_due_at, last_review_at,
			  last_quality, COALESCE(last_options_json, ''), 
			  COALESCE(wrong_answers_json, ''), COALESCE(stats_json, ''),
			  created_at, updated_at
			  FROM user_cards WHERE id = ?`

	return r.scanUserCard(r.db.QueryRow(query, id))
}

// GetUserCardByTrainingCard gets a user card by user, training card and direction
func (r *UserCardRepository) GetUserCardByTrainingCard(userID, trainingCardID int64, direction models.CardDirection) (*models.UserCard, error) {
	query := `SELECT id, user_id, training_card_id, direction, state, ef, reps,
			  interval_days, learning_step, lapse_count, next_due_at, last_review_at,
			  last_quality, COALESCE(last_options_json, ''), 
			  COALESCE(wrong_answers_json, ''), COALESCE(stats_json, ''),
			  created_at, updated_at
			  FROM user_cards WHERE user_id = ? AND training_card_id = ? AND direction = ?`

	return r.scanUserCard(r.db.QueryRow(query, userID, trainingCardID, direction))
}

// GetDueCards gets cards that are due for review
func (r *UserCardRepository) GetDueCards(userID int64, now time.Time, limit int) ([]*models.UserCard, error) {
	query := `SELECT id, user_id, training_card_id, direction, state, ef, reps,
			  interval_days, learning_step, lapse_count, next_due_at, last_review_at,
			  last_quality, COALESCE(last_options_json, ''), 
			  COALESCE(wrong_answers_json, ''), COALESCE(stats_json, ''),
			  created_at, updated_at
			  FROM user_cards 
			  WHERE user_id = ? AND (next_due_at IS NULL OR next_due_at <= ?)
			  ORDER BY 
			    CASE WHEN state = 'learning' THEN 0 ELSE 1 END,
			    next_due_at ASC NULLS FIRST
			  LIMIT ?`

	rows, err := r.db.Query(query, userID, now, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get due cards: %w", err)
	}
	defer rows.Close()

	return r.scanUserCards(rows)
}

// GetDueCount gets the count of due cards for a user
func (r *UserCardRepository) GetDueCount(userID int64, now time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM user_cards 
			  WHERE user_id = ? AND (next_due_at IS NULL OR next_due_at <= ?)`

	var count int
	err := r.db.QueryRow(query, userID, now).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get due count: %w", err)
	}
	return count, nil
}

// GetNewCards gets new cards for a user
func (r *UserCardRepository) GetNewCards(userID int64, limit int) ([]*models.UserCard, error) {
	query := `SELECT id, user_id, training_card_id, direction, state, ef, reps,
			  interval_days, learning_step, lapse_count, next_due_at, last_review_at,
			  last_quality, COALESCE(last_options_json, ''), 
			  COALESCE(wrong_answers_json, ''), COALESCE(stats_json, ''),
			  created_at, updated_at
			  FROM user_cards 
			  WHERE user_id = ? AND state = 'new'
			  ORDER BY created_at
			  LIMIT ?`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get new cards: %w", err)
	}
	defer rows.Close()

	return r.scanUserCards(rows)
}

// UpdateUserCard updates a user card's SRS state
func (r *UserCardRepository) UpdateUserCard(card *models.UserCard) error {
	query := `UPDATE user_cards SET
			  state = ?, ef = ?, reps = ?, interval_days = ?, learning_step = ?,
			  lapse_count = ?, next_due_at = ?, last_review_at = ?, last_quality = ?,
			  last_options_json = ?, wrong_answers_json = ?, stats_json = ?,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = ?`

	_, err := r.db.Exec(query,
		card.State, card.EF, card.Reps, card.IntervalDays, card.LearningStep,
		card.LapseCount, card.NextDueAt, card.LastReviewAt, card.LastQuality,
		card.LastOptionsJSON, card.WrongAnswersJSON, card.StatsJSON,
		card.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user card: %w", err)
	}

	r.logger.Debug("updated user card",
		zap.Int64("id", card.ID),
		zap.String("state", string(card.State)),
		zap.Int("interval_days", card.IntervalDays),
	)

	return nil
}

// scanUserCard scans a single user card from a row
func (r *UserCardRepository) scanUserCard(row *sql.Row) (*models.UserCard, error) {
	var card models.UserCard
	var createdAt, updatedAt string
	var nextDueAt, lastReviewAt sql.NullString
	var lastQuality sql.NullInt64

	err := row.Scan(
		&card.ID, &card.UserID, &card.TrainingCardID, &card.Direction, &card.State,
		&card.EF, &card.Reps, &card.IntervalDays, &card.LearningStep, &card.LapseCount,
		&nextDueAt, &lastReviewAt, &lastQuality,
		&card.LastOptionsJSON, &card.WrongAnswersJSON, &card.StatsJSON,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan user card: %w", err)
	}

	card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	card.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	if nextDueAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", nextDueAt.String)
		card.NextDueAt = &t
	}
	if lastReviewAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", lastReviewAt.String)
		card.LastReviewAt = &t
	}
	if lastQuality.Valid {
		q := int(lastQuality.Int64)
		card.LastQuality = &q
	}

	return &card, nil
}

// scanUserCards scans multiple user cards from rows
func (r *UserCardRepository) scanUserCards(rows *sql.Rows) ([]*models.UserCard, error) {
	var cards []*models.UserCard

	for rows.Next() {
		var card models.UserCard
		var createdAt, updatedAt string
		var nextDueAt, lastReviewAt sql.NullString
		var lastQuality sql.NullInt64

		err := rows.Scan(
			&card.ID, &card.UserID, &card.TrainingCardID, &card.Direction, &card.State,
			&card.EF, &card.Reps, &card.IntervalDays, &card.LearningStep, &card.LapseCount,
			&nextDueAt, &lastReviewAt, &lastQuality,
			&card.LastOptionsJSON, &card.WrongAnswersJSON, &card.StatsJSON,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user card: %w", err)
		}

		card.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		card.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		if nextDueAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", nextDueAt.String)
			card.NextDueAt = &t
		}
		if lastReviewAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", lastReviewAt.String)
			card.LastReviewAt = &t
		}
		if lastQuality.Valid {
			q := int(lastQuality.Int64)
			card.LastQuality = &q
		}

		cards = append(cards, &card)
	}

	return cards, nil
}

// DeleteOrphanedUserCards deletes user_cards that reference non-existent training_cards
func (r *UserCardRepository) DeleteOrphanedUserCards() (int64, error) {
	query := `DELETE FROM user_cards 
			  WHERE training_card_id NOT IN (SELECT id FROM training_cards)`
	
	result, err := r.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete orphaned user cards: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		r.logger.Info("deleted orphaned user cards",
			zap.Int64("rows_affected", rowsAffected),
		)
	}

	return rowsAffected, nil
}

// DeleteUserCardsByWordENForUser deletes all user_cards for a specific word and user
// This will cascade delete review_events due to foreign key constraint
// DEPRECATED: Use DeleteUserCardsByWordCardIDForUser instead
func (r *UserCardRepository) DeleteUserCardsByWordENForUser(userID int64, wordEN string) (int64, error) {
	query := `DELETE FROM user_cards 
			  WHERE user_id = ? 
			  AND training_card_id IN (
				  SELECT id FROM training_cards WHERE word_en = ?
			  )`
	
	result, err := r.db.Exec(query, userID, wordEN)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user cards: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("deleted user cards by word",
		zap.Int64("user_id", userID),
		zap.String("word_en", wordEN),
		zap.Int64("rows_affected", rowsAffected),
	)

	return rowsAffected, nil
}

// DeleteUserCardsByWordCardIDForUser deletes all user_cards for a specific word_card_id and user
// This will cascade delete review_events due to foreign key constraint
func (r *UserCardRepository) DeleteUserCardsByWordCardIDForUser(userID int64, wordCardID int64) (int64, error) {
	query := `DELETE FROM user_cards 
			  WHERE user_id = ? 
			  AND training_card_id IN (
				  SELECT id FROM training_cards WHERE word_card_id = ?
			  )`
	
	result, err := r.db.Exec(query, userID, wordCardID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user cards: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("deleted user cards by word_card_id",
		zap.Int64("user_id", userID),
		zap.Int64("word_card_id", wordCardID),
		zap.Int64("rows_affected", rowsAffected),
	)

	return rowsAffected, nil
}

