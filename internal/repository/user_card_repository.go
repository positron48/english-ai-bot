package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/database"
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

	id, err := database.InsertAndReturnID(r.db, query,
		card.UserID, card.TrainingCardID, card.Direction, card.State, card.EF,
		card.Reps, card.IntervalDays, card.LearningStep, card.LapseCount,
		card.NextDueAt, card.LastReviewAt, card.LastQuality,
		card.LastOptionsJSON, card.WrongAnswersJSON, card.StatsJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create user card: %w", err)
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
// Excludes words marked as "known" in user_word_knowledge
func (r *UserCardRepository) GetDueCards(userID int64, now time.Time, limit int) ([]*models.UserCard, error) {
	query := `SELECT uc.id, uc.user_id, uc.training_card_id, uc.direction, uc.state, uc.ef, uc.reps,
			  uc.interval_days, uc.learning_step, uc.lapse_count, uc.next_due_at, uc.last_review_at,
			  uc.last_quality, COALESCE(uc.last_options_json, ''), 
			  COALESCE(uc.wrong_answers_json, ''), COALESCE(uc.stats_json, ''),
			  uc.created_at, uc.updated_at
			  FROM user_cards uc
			  INNER JOIN training_cards tc ON uc.training_card_id = tc.id
			  WHERE uc.user_id = ? 
			    AND (uc.next_due_at IS NULL OR uc.next_due_at <= ?)
			    AND NOT EXISTS (
			      SELECT 1 FROM user_word_knowledge uwk 
			      WHERE uwk.user_id = ? AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
			    )
			  ORDER BY 
			    CASE WHEN uc.state = 'learning' THEN 0 ELSE 1 END,
			    uc.next_due_at ASC NULLS FIRST
			  LIMIT ?`

	rows, err := r.db.Query(query, userID, now, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get due cards: %w", err)
	}
	defer rows.Close()

	return r.scanUserCards(rows)
}

// GetDueCount gets the count of due cards for a user
// Excludes words marked as "known" in user_word_knowledge
func (r *UserCardRepository) GetDueCount(userID int64, now time.Time) (int, error) {
	query := `SELECT COUNT(*) 
			  FROM user_cards uc
			  INNER JOIN training_cards tc ON uc.training_card_id = tc.id
			  WHERE uc.user_id = ? 
			    AND (uc.next_due_at IS NULL OR uc.next_due_at <= ?)
			    AND NOT EXISTS (
			      SELECT 1 FROM user_word_knowledge uwk 
			      WHERE uwk.user_id = ? AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
			    )`

	var count int
	err := r.db.QueryRow(query, userID, now, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get due count: %w", err)
	}
	return count, nil
}

// GetNewCards gets new cards for a user
// Excludes words marked as "known" in user_word_knowledge
func (r *UserCardRepository) GetNewCards(userID int64, limit int) ([]*models.UserCard, error) {
	query := `SELECT uc.id, uc.user_id, uc.training_card_id, uc.direction, uc.state, uc.ef, uc.reps,
			  uc.interval_days, uc.learning_step, uc.lapse_count, uc.next_due_at, uc.last_review_at,
			  uc.last_quality, COALESCE(uc.last_options_json, ''), 
			  COALESCE(uc.wrong_answers_json, ''), COALESCE(uc.stats_json, ''),
			  uc.created_at, uc.updated_at
			  FROM user_cards uc
			  INNER JOIN training_cards tc ON uc.training_card_id = tc.id
			  WHERE uc.user_id = ? 
			    AND uc.state = 'new'
			    AND NOT EXISTS (
			      SELECT 1 FROM user_word_knowledge uwk 
			      WHERE uwk.user_id = ? AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
			    )
			  ORDER BY uc.created_at
			  LIMIT ?`

	rows, err := r.db.Query(query, userID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get new cards: %w", err)
	}
	defer rows.Close()

	return r.scanUserCards(rows)
}

// WordMasteringStats holds aggregate stats for one word (word_card_id) for computing mastering score 0-100
type WordMasteringStats struct {
	TotalCards        int
	ReviewStateCount  int
	LearningStateCount int
	NewStateCount     int
	TotalReps         int
	IsKnown           bool
}

// GetWordMasteringStats returns aggregate stats for the given word so the service can compute mastering score (0-100)
func (r *UserCardRepository) GetWordMasteringStats(userID, wordCardID int64) (*WordMasteringStats, error) {
	query := `SELECT
		COUNT(DISTINCT uc.id) AS total_cards,
		COALESCE(SUM(CASE WHEN uc.state = 'review' THEN 1 ELSE 0 END), 0) AS review_state_count,
		COALESCE(SUM(CASE WHEN uc.state = 'learning' THEN 1 ELSE 0 END), 0) AS learning_state_count,
		COALESCE(SUM(CASE WHEN uc.state = 'new' THEN 1 ELSE 0 END), 0) AS new_state_count,
		COALESCE(SUM(uc.reps), 0) AS total_reps,
		CASE WHEN MAX(uwk.word_card_id) IS NOT NULL THEN 1 ELSE 0 END AS is_known
	FROM user_cards uc
	JOIN training_cards tc ON uc.training_card_id = tc.id
	LEFT JOIN user_word_knowledge uwk ON uwk.user_id = uc.user_id AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
	WHERE uc.user_id = ? AND tc.word_card_id = ?`
	var s WordMasteringStats
	var isKnown int
	err := r.db.QueryRow(query, userID, wordCardID).Scan(
		&s.TotalCards, &s.ReviewStateCount, &s.LearningStateCount, &s.NewStateCount, &s.TotalReps, &isKnown)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get word mastering stats: %w", err)
	}
	s.IsKnown = isKnown == 1
	return &s, nil
}

// SpellEligibleWord is a word the user has in review state (suitable for "compose word" spell challenge)
type SpellEligibleWord struct {
	WordCardID  int64
	DisplayWord string
	WordRU      string
}

// GetWordsEligibleForSpell returns words where the user has at least one card in 'review' state (one row per word_card_id)
func (r *UserCardRepository) GetWordsEligibleForSpell(userID int64, limit int) ([]*SpellEligibleWord, error) {
	return r.GetWordsEligibleForSpellByMastery(userID, 50, limit)
}

// GetWordsEligibleForSpellByMastery returns words with stored mastering_score >= minScore (0-100).
func (r *UserCardRepository) GetWordsEligibleForSpellByMastery(userID int64, minScore int, limit int) ([]*SpellEligibleWord, error) {
	if minScore < 0 {
		minScore = 0
	}
	if minScore > 100 {
		minScore = 100
	}
	query := `SELECT tc.word_card_id,
		COALESCE(MAX(CASE WHEN tc.display_word IS NOT NULL AND tc.display_word != '' THEN tc.display_word END), MAX(wc.word)) AS display_word,
		MAX(tc.word_ru) AS word_ru
		FROM user_cards uc
		JOIN training_cards tc ON uc.training_card_id = tc.id
		JOIN word_cards wc ON tc.word_card_id = wc.id
		LEFT JOIN user_word_mastering uwm ON uwm.user_id = uc.user_id AND uwm.word_card_id = tc.word_card_id
		WHERE uc.user_id = ? AND COALESCE(uwm.mastering_score, 0) >= ?
		GROUP BY tc.word_card_id
		LIMIT ?`
	rows, err := r.db.Query(query, userID, minScore, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get words eligible for spell: %w", err)
	}
	defer rows.Close()
	var out []*SpellEligibleWord
	for rows.Next() {
		var w SpellEligibleWord
		var displayWord, wordRU sql.NullString
		if err := rows.Scan(&w.WordCardID, &displayWord, &wordRU); err != nil {
			return nil, fmt.Errorf("scan spell eligible word: %w", err)
		}
		if displayWord.Valid {
			w.DisplayWord = displayWord.String
		}
		if wordRU.Valid {
			w.WordRU = wordRU.String
		}
		out = append(out, &w)
	}
	return out, nil
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
// Also deletes user_cards whose training_cards reference non-existent word_cards
func (r *UserCardRepository) DeleteOrphanedUserCards() (int64, error) {
	// Delete user_cards that reference non-existent training_cards
	query1 := `DELETE FROM user_cards 
			   WHERE training_card_id NOT IN (SELECT id FROM training_cards)`
	
	result1, err := r.db.Exec(query1)
	if err != nil {
		return 0, fmt.Errorf("failed to delete orphaned user cards: %w", err)
	}

	rowsAffected1, err := result1.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Delete user_cards whose training_cards reference non-existent word_cards
	// Use subquery for portable DELETE
	query2 := `DELETE FROM user_cards 
			   WHERE training_card_id IN (
				   SELECT tc.id 
				   FROM training_cards tc
				   LEFT JOIN word_cards wc ON tc.word_card_id = wc.id
				   WHERE wc.id IS NULL
			   )`
	
	result2, err := r.db.Exec(query2)
	if err != nil {
		return 0, fmt.Errorf("failed to delete user cards with orphaned training cards: %w", err)
	}

	rowsAffected2, err := result2.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	totalAffected := rowsAffected1 + rowsAffected2

	if totalAffected > 0 {
		r.logger.Info("deleted orphaned user cards",
			zap.Int64("type1_deleted", rowsAffected1),
			zap.Int64("type2_deleted", rowsAffected2),
			zap.Int64("total_deleted", totalAffected),
		)
	}

	return totalAffected, nil
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

// OrphanedUserCardInfo represents information about an orphaned user card
type OrphanedUserCardInfo struct {
	UserCardID      int64
	UserID          int64
	TelegramID      int64
	TelegramUsername *string
	TrainingCardID  int64
	Direction       string
	State           string
	Reps            int
	CreatedAt       time.Time
	ReviewEventsCount int64
}

// ListOrphanedUserCards lists user_cards that reference non-existent training_cards
func (r *UserCardRepository) ListOrphanedUserCards(limit, offset int) ([]*OrphanedUserCardInfo, error) {
	query := `SELECT 
		uc.id as user_card_id,
		uc.user_id,
		u.telegram_id,
		u.telegram_username,
		uc.training_card_id,
		uc.direction,
		uc.state,
		uc.reps,
		uc.created_at,
		COALESCE(COUNT(re.id), 0) as review_events_count
	FROM user_cards uc
	LEFT JOIN users u ON uc.user_id = u.id
	LEFT JOIN review_events re ON re.user_card_id = uc.id
	WHERE uc.training_card_id NOT IN (SELECT id FROM training_cards)
	GROUP BY uc.id, uc.user_id, u.telegram_id, u.telegram_username, uc.training_card_id, uc.direction, uc.state, uc.reps, uc.created_at
	ORDER BY uc.created_at DESC
	LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orphaned user cards: %w", err)
	}
	defer rows.Close()

	var items []*OrphanedUserCardInfo
	for rows.Next() {
		var item OrphanedUserCardInfo
		var createdAt string
		var telegramUsername sql.NullString

		err := rows.Scan(
			&item.UserCardID,
			&item.UserID,
			&item.TelegramID,
			&telegramUsername,
			&item.TrainingCardID,
			&item.Direction,
			&item.State,
			&item.Reps,
			&createdAt,
			&item.ReviewEventsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan orphaned user card: %w", err)
		}

		if telegramUsername.Valid {
			item.TelegramUsername = &telegramUsername.String
		}

		item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		items = append(items, &item)
	}

	return items, nil
}

// CountOrphanedUserCards counts total orphaned user cards
func (r *UserCardRepository) CountOrphanedUserCards() (int, error) {
	query := `SELECT COUNT(*) 
	FROM user_cards 
	WHERE training_card_id NOT IN (SELECT id FROM training_cards)`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count orphaned user cards: %w", err)
	}

	return count, nil
}

// ListUserCardsWithOrphanedTrainingCards lists user_cards whose training_cards reference non-existent word_cards
// These cards pass INNER JOIN but cannot be used in training because word_card is missing
func (r *UserCardRepository) ListUserCardsWithOrphanedTrainingCards(limit, offset int) ([]*OrphanedUserCardInfo, error) {
	query := `SELECT 
		uc.id as user_card_id,
		uc.user_id,
		u.telegram_id,
		u.telegram_username,
		uc.training_card_id,
		uc.direction,
		uc.state,
		uc.reps,
		uc.created_at,
		COALESCE(COUNT(re.id), 0) as review_events_count
	FROM user_cards uc
	INNER JOIN training_cards tc ON uc.training_card_id = tc.id
	LEFT JOIN word_cards wc ON tc.word_card_id = wc.id
	LEFT JOIN users u ON uc.user_id = u.id
	LEFT JOIN review_events re ON re.user_card_id = uc.id
	WHERE wc.id IS NULL
	GROUP BY uc.id, uc.user_id, u.telegram_id, u.telegram_username, uc.training_card_id, uc.direction, uc.state, uc.reps, uc.created_at
	ORDER BY uc.created_at DESC
	LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user cards with orphaned training cards: %w", err)
	}
	defer rows.Close()

	var items []*OrphanedUserCardInfo
	for rows.Next() {
		var item OrphanedUserCardInfo
		var createdAt string
		var telegramUsername sql.NullString

		err := rows.Scan(
			&item.UserCardID,
			&item.UserID,
			&item.TelegramID,
			&telegramUsername,
			&item.TrainingCardID,
			&item.Direction,
			&item.State,
			&item.Reps,
			&createdAt,
			&item.ReviewEventsCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user card with orphaned training card: %w", err)
		}

		if telegramUsername.Valid {
			item.TelegramUsername = &telegramUsername.String
		}

		item.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		items = append(items, &item)
	}

	return items, nil
}

// CountUserCardsWithOrphanedTrainingCards counts user_cards whose training_cards reference non-existent word_cards
func (r *UserCardRepository) CountUserCardsWithOrphanedTrainingCards() (int, error) {
	query := `SELECT COUNT(*) 
	FROM user_cards uc
	INNER JOIN training_cards tc ON uc.training_card_id = tc.id
	LEFT JOIN word_cards wc ON tc.word_card_id = wc.id
	WHERE wc.id IS NULL`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user cards with orphaned training cards: %w", err)
	}
	return count, nil
}

// DeleteUserCard deletes a specific user card by ID
// This will cascade delete review_events due to foreign key constraint
func (r *UserCardRepository) DeleteUserCard(userCardID int64) error {
	query := `DELETE FROM user_cards WHERE id = ?`
	
	result, err := r.db.Exec(query, userCardID)
	if err != nil {
		return fmt.Errorf("failed to delete user card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user card not found")
	}

	r.logger.Info("deleted user card",
		zap.Int64("user_card_id", userCardID),
		zap.Int64("rows_affected", rowsAffected),
	)

	return nil
}

// GetUserIDsByWordCardID gets all distinct user IDs who have user_cards for a specific word_card_id
func (r *UserCardRepository) GetUserIDsByWordCardID(wordCardID int64) ([]int64, error) {
	query := `SELECT DISTINCT uc.user_id 
			  FROM user_cards uc
			  INNER JOIN training_cards tc ON uc.training_card_id = tc.id
			  WHERE tc.word_card_id = ?`

	rows, err := r.db.Query(query, wordCardID)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return userIDs, nil
}

// GetUpcomingCardsByDate gets the count of cards that will become due in the next 7 days
// Returns a map where key is date string (YYYY-MM-DD) and value is the count of cards
// Excludes words marked as "known" in user_word_knowledge
func (r *UserCardRepository) GetUpcomingCardsByDate(userID int64, startDate time.Time) (map[string]int, error) {
	// Calculate end date (7 days from start, at end of day)
	endDate := startDate.AddDate(0, 0, 7)
	
	r.logger.Debug("getting upcoming cards by date",
		zap.Int64("user_id", userID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)
	
	// Get all cards with next_due_at in the range, then group by date in Go
	query := `SELECT uc.next_due_at
	FROM user_cards uc
	INNER JOIN training_cards tc ON uc.training_card_id = tc.id
	WHERE uc.user_id = ? 
		AND uc.next_due_at IS NOT NULL
		AND uc.next_due_at > ?
		AND uc.next_due_at <= ?
		AND NOT EXISTS (
			SELECT 1 FROM user_word_knowledge uwk 
			WHERE uwk.user_id = ? AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
		)`

	rows, err := r.db.Query(query, userID, startDate, endDate, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming cards by date: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	
	// Initialize all 7 days with 0
	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		result[dateStr] = 0
	}

	// Group cards by date
	cardCount := 0
	for rows.Next() {
		var nextDueAtStr sql.NullString
		if err := rows.Scan(&nextDueAtStr); err != nil {
			return nil, fmt.Errorf("failed to scan upcoming cards: %w", err)
		}
		if nextDueAtStr.Valid {
			cardCount++
			// Try multiple date formats (Postgres TIMESTAMPTZ can vary)
			var dueTime time.Time
			var err error
			
			// Try RFC3339 format first (ISO 8601 with timezone)
			dueTime, err = time.Parse(time.RFC3339Nano, nextDueAtStr.String)
			if err != nil {
				// Try RFC3339 without nanoseconds
				dueTime, err = time.Parse(time.RFC3339, nextDueAtStr.String)
			}
			if err != nil {
				// Try standard format
				dueTime, err = time.Parse("2006-01-02 15:04:05", nextDueAtStr.String)
			}
			if err != nil {
				// Try format with timezone offset
				dueTime, err = time.Parse("2006-01-02 15:04:05-07:00", nextDueAtStr.String)
			}
			if err != nil {
				r.logger.Warn("failed to parse next_due_at", zap.String("date", nextDueAtStr.String), zap.Error(err))
				continue
			}
			// Convert to same timezone as startDate
			dueTime = dueTime.In(startDate.Location())
			// Get date string (YYYY-MM-DD)
			dateStr := dueTime.Format("2006-01-02")
			// Increment count for this date
			if _, exists := result[dateStr]; exists {
				result[dateStr]++
			} else {
				// If date is outside our 7-day range, still count it but log
				r.logger.Debug("card date outside 7-day range", zap.String("date", dateStr))
			}
		}
	}
	
	r.logger.Debug("processed upcoming cards",
		zap.Int("total_cards_found", cardCount),
		zap.Any("result", result),
	)

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	r.logger.Debug("upcoming cards by date result",
		zap.Any("result", result),
	)

	return result, nil
}
