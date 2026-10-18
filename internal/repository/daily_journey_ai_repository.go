package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"tgbot-skeleton/internal/dailyjourney"
)

// JourneyAIReservation is private server state. Input may contain answer keys,
// model prompts and history, and must never be serialized to the HTTP client.
type JourneyAIReservation struct {
	Token  string
	Input  json.RawMessage
	Result *dailyjourney.Journey
}
type PrepareJourneyAI func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step, json.RawMessage) (json.RawMessage, error)

// A crashed worker loses its authority after the lease. Expiring the receipt
// preserves its payload hash/input for a retry while fencing off its late result.
func journeyProcessingBarrier(ctx context.Context, tx *sql.Tx, id string) error {
	if _, err := tx.ExecContext(ctx, `UPDATE daily_journey_actions SET status='retryable',reservation_token=NULL,lease_expires_at=NULL,updated_at=CURRENT_TIMESTAMP WHERE journey_id=? AND status='processing' AND (lease_expires_at IS NULL OR lease_expires_at<=CURRENT_TIMESTAMP)`, id); err != nil {
		return err
	}
	var pending bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM daily_journey_actions WHERE journey_id=? AND status='processing')`, id).Scan(&pending); err != nil {
		return err
	}
	if pending {
		return dailyjourney.ErrProcessing
	}
	return nil
}
func (r *DailyJourneyRepository) ReserveAI(ctx context.Context, user int64, course, id, stepID string, action dailyjourney.Action, lease time.Duration, prepare PrepareJourneyAI) (*JourneyAIReservation, error) {
	hash, err := journeyActionHash(stepID, action)
	if err != nil {
		return nil, err
	}
	if lease < time.Second || lease > 5*time.Minute {
		return nil, fmt.Errorf("invalid daily journey AI lease")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	j, err := loadJourney(ctx, tx, user, course, id, true)
	if err != nil {
		return nil, err
	}
	var oldHash, status string
	var saved, input []byte
	err = tx.QueryRowContext(ctx, `SELECT payload_hash,status,result_json,prepared_json FROM daily_journey_actions WHERE journey_id=? AND client_action_id=?`, id, action.ClientActionID).Scan(&oldHash, &status, &saved, &input)
	exists := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if exists {
		if oldHash != hash {
			return nil, dailyjourney.ErrActionKey
		}
		if status == "completed" {
			var result dailyjourney.Journey
			if err := json.Unmarshal(saved, &result); err != nil {
				return nil, err
			}
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			return &JourneyAIReservation{Result: &result}, nil
		}
	}
	if err := journeyProcessingBarrier(ctx, tx, id); err != nil {
		return nil, err
	}
	if j.Revision != action.ExpectedRevision {
		return nil, dailyjourney.ErrConflict
	}
	step := j.Current()
	if j.Status != "active" || step == nil || step.ID != stepID {
		return nil, dailyjourney.ErrInvalidAction
	}
	prepared, err := prepare(tx, j, step, input)
	if err != nil {
		return nil, err
	}
	if !json.Valid(prepared) || len(prepared) > 2<<20 {
		return nil, fmt.Errorf("invalid prepared daily journey AI input")
	}
	// A claimed retry keeps its exact originally prepared input, after access revalidation.
	if len(input) > 0 {
		prepared = input
	}
	token := uuid.NewString()
	if exists {
		_, err = tx.ExecContext(ctx, `UPDATE daily_journey_actions SET status='processing',reservation_token=?,lease_expires_at=CURRENT_TIMESTAMP+(? * INTERVAL '1 millisecond'),prepared_json=?::jsonb,updated_at=CURRENT_TIMESTAMP WHERE journey_id=? AND client_action_id=?`, token, lease.Milliseconds(), string(prepared), id, action.ClientActionID)
	} else {
		_, err = tx.ExecContext(ctx, `INSERT INTO daily_journey_actions(journey_id,client_action_id,step_id,action_type,payload_hash,status,reservation_token,lease_expires_at,prepared_json) VALUES(?,?,?,?,?,'processing',?,CURRENT_TIMESTAMP+(? * INTERVAL '1 millisecond'),?::jsonb)`, id, action.ClientActionID, stepID, action.Type, hash, token, lease.Milliseconds(), string(prepared))
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &JourneyAIReservation{Token: token, Input: prepared}, nil
}
func (r *DailyJourneyRepository) CompleteAI(ctx context.Context, user int64, course, id, stepID string, action dailyjourney.Action, token string, apply func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step) error) (*dailyjourney.Journey, error) {
	hash, err := journeyActionHash(stepID, action)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, dailyjourney.ErrReservationLost
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	j, err := loadJourney(ctx, tx, user, course, id, true)
	if err != nil {
		return nil, err
	}
	var oldHash, status string
	var currentToken sql.NullString
	var saved []byte
	var live bool
	err = tx.QueryRowContext(ctx, `SELECT payload_hash,status,reservation_token,result_json,COALESCE(lease_expires_at>CURRENT_TIMESTAMP,false) FROM daily_journey_actions WHERE journey_id=? AND client_action_id=?`, id, action.ClientActionID).Scan(&oldHash, &status, &currentToken, &saved, &live)
	if err == sql.ErrNoRows {
		return nil, dailyjourney.ErrReservationLost
	}
	if err != nil {
		return nil, err
	}
	if oldHash != hash {
		return nil, dailyjourney.ErrActionKey
	}
	if status == "completed" {
		var result dailyjourney.Journey
		if err := json.Unmarshal(saved, &result); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &result, nil
	}
	if status != "processing" || !currentToken.Valid || currentToken.String != token || !live {
		return nil, dailyjourney.ErrReservationLost
	}
	if j.Revision != action.ExpectedRevision {
		return nil, dailyjourney.ErrConflict
	}
	step := j.Current()
	if j.Status != "active" || step == nil || step.ID != stepID {
		return nil, dailyjourney.ErrInvalidAction
	}
	if err := apply(tx, j, step); err != nil {
		return nil, err
	}
	if err := persistJourneyAction(ctx, tx, j, step, action, hash, token); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return j, nil
}

// FailAI releases only the matching worker. A failure/late cancellation cannot
// release a newer reservation, and no subject credit has been applied here.
func (r *DailyJourneyRepository) FailAI(ctx context.Context, user int64, course, id, key, token string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE daily_journey_actions a SET status='retryable',reservation_token=NULL,lease_expires_at=NULL,updated_at=CURRENT_TIMESTAMP FROM daily_journeys j JOIN user_courses uc ON uc.id=j.user_course_id JOIN courses c ON c.id=uc.course_id WHERE a.journey_id=j.id AND a.journey_id=? AND a.client_action_id=? AND a.reservation_token=? AND a.status='processing' AND uc.user_id=? AND c.code=?`, id, key, token, user, course)
	return err
}
