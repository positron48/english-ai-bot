package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// NudgeRepository handles database operations for training nudges
type NudgeRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewNudgeRepository creates a new nudge repository
func NewNudgeRepository(db *sql.DB, logger *zap.Logger) *NudgeRepository {
	return &NudgeRepository{
		db:     db,
		logger: logger,
	}
}

// CreateNudge creates a new training nudge
func (r *NudgeRepository) CreateNudge(nudge *models.TrainingNudge) (int64, error) {
	query := `INSERT INTO training_nudges (
		user_id, local_date, due_count_at_send, message_id
	) VALUES (?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		nudge.UserID, nudge.LocalDate, nudge.DueCountAtSend, nudge.MessageID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create nudge: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get nudge ID: %w", err)
	}

	r.logger.Debug("created training nudge",
		zap.Int64("nudge_id", id),
		zap.Int64("user_id", nudge.UserID),
		zap.String("local_date", nudge.LocalDate),
	)

	return id, nil
}

// HasNudgeToday checks if a nudge was sent to user today
func (r *NudgeRepository) HasNudgeToday(userID int64, localDate string) (bool, error) {
	query := `SELECT COUNT(*) FROM training_nudges 
			  WHERE user_id = ? AND local_date = ?`

	var count int
	err := r.db.QueryRow(query, userID, localDate).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check nudge: %w", err)
	}

	return count > 0, nil
}

// ConsumeNudge marks a nudge as consumed (user started training)
func (r *NudgeRepository) ConsumeNudge(userID int64, localDate string) error {
	query := `UPDATE training_nudges SET consumed_at = CURRENT_TIMESTAMP 
			  WHERE user_id = ? AND local_date = ? AND consumed_at IS NULL`

	_, err := r.db.Exec(query, userID, localDate)
	if err != nil {
		return fmt.Errorf("failed to consume nudge: %w", err)
	}

	r.logger.Debug("consumed training nudge",
		zap.Int64("user_id", userID),
		zap.String("local_date", localDate),
	)

	return nil
}

// GetUnconsumedNudge gets the unconsumed nudge for a user on a date
func (r *NudgeRepository) GetUnconsumedNudge(userID int64, localDate string) (*models.TrainingNudge, error) {
	query := `SELECT id, user_id, local_date, sent_at, consumed_at, 
			  due_count_at_send, message_id
			  FROM training_nudges 
			  WHERE user_id = ? AND local_date = ? AND consumed_at IS NULL`

	var nudge models.TrainingNudge
	var sentAt string
	var consumedAt sql.NullString
	var messageID sql.NullInt64

	err := r.db.QueryRow(query, userID, localDate).Scan(
		&nudge.ID, &nudge.UserID, &nudge.LocalDate, &sentAt,
		&consumedAt, &nudge.DueCountAtSend, &messageID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get nudge: %w", err)
	}

	nudge.SentAt, _ = time.Parse("2006-01-02 15:04:05", sentAt)
	if consumedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", consumedAt.String)
		nudge.ConsumedAt = &t
	}
	if messageID.Valid {
		msgID := int(messageID.Int64)
		nudge.MessageID = &msgID
	}

	return &nudge, nil
}

