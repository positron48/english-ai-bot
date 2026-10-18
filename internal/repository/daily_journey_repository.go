package repository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"tgbot-skeleton/internal/dailyjourney"
)

type DailyJourneyRepository struct{ db *sql.DB }

func NewDailyJourneyRepository(db *sql.DB) *DailyJourneyRepository {
	return &DailyJourneyRepository{db: db}
}

type journeyQuery interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func loadJourney(ctx context.Context, q journeyQuery, userID int64, course, id string, lock bool) (*dailyjourney.Journey, error) {
	query := `SELECT j.id, c.code, j.journey_date::text, j.day_timezone, j.status, j.algorithm_version, j.revision
 FROM daily_journeys j JOIN user_courses uc ON uc.id=j.user_course_id JOIN courses c ON c.id=uc.course_id
 WHERE uc.user_id=? AND c.code=? AND j.id=?`
	if lock {
		query += ` FOR UPDATE OF j`
	}
	j := &dailyjourney.Journey{Steps: []dailyjourney.Step{}}
	err := q.QueryRowContext(ctx, query, userID, course, id).Scan(&j.ID, &j.CourseCode, &j.Date, &j.DayTimezone, &j.Status, &j.AlgorithmVersion, &j.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, dailyjourney.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load daily journey: %w", err)
	}
	rows, err := q.QueryContext(ctx, `SELECT id, mode, title, goal, required, status, content_ref_json, content_version, target_count, completed_count, state_json, launch_json, feedback, result, result_json FROM daily_journey_steps WHERE journey_id=? ORDER BY position`, id)
	if err != nil {
		return nil, fmt.Errorf("load daily journey steps: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var s dailyjourney.Step
		var launch []byte
		if err := rows.Scan(&s.ID, &s.Mode, &s.Title, &s.Goal, &s.Required, &s.Status, &s.ContentRef, &s.ContentVersion, &s.TargetCount, &s.CompletedCount, &s.State, &launch, &s.Feedback, &s.Result, &s.ResultDetails); err != nil {
			return nil, fmt.Errorf("scan daily journey step: %w", err)
		}
		if len(launch) > 0 {
			if err := json.Unmarshal(launch, &s.Launch); err != nil {
				return nil, fmt.Errorf("decode daily journey launch: %w", err)
			}
		}
		j.Steps = append(j.Steps, s)
	}
	return j, rows.Err()
}

// The compatibility driver exposes the legacy Begin API, so transaction modes
// must be set explicitly before the first query rather than via sql.TxOptions.
func (r *DailyJourneyRepository) beginSnapshot(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ, READ ONLY`); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("daily journey read snapshot: %w", err)
	}
	return tx, nil
}

// Get uses a consistent snapshot so a concurrently committed answer cannot mix
// a previous journey revision with a new step cursor.
func (r *DailyJourneyRepository) Get(ctx context.Context, userID int64, course, id string) (*dailyjourney.Journey, error) {
	tx, err := r.beginSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	j, err := loadJourney(ctx, tx, userID, course, id, false)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return j, nil
}

// Current is read-only, including for users who have never started a journey.
// A previous active itinerary survives midnight until an explicit Start today.
func (r *DailyJourneyRepository) Current(ctx context.Context, userID int64, course, date string) (*dailyjourney.Journey, error) {
	tx, err := r.beginSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var id string
	err = tx.QueryRowContext(ctx, `SELECT j.id FROM daily_journeys j JOIN user_courses uc ON uc.id=j.user_course_id JOIN courses c ON c.id=uc.course_id WHERE uc.user_id=? AND c.code=? AND (j.journey_date=?::date OR j.status='active') ORDER BY (j.journey_date=?::date) DESC, j.journey_date DESC LIMIT 1`, userID, course, date, date).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	j, err := loadJourney(ctx, tx, userID, course, id, false)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return j, nil
}

// build must only discover authorized content, never start subject sessions or call AI.
// The user-course lock serializes first starts without changing the day's saved selection.
func (r *DailyJourneyRepository) Start(ctx context.Context, userID int64, course string, now time.Time, build func(*sql.Tx) ([]dailyjourney.Step, error)) (*dailyjourney.Journey, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var ucID int64
	err = tx.QueryRowContext(ctx, `SELECT uc.id FROM user_courses uc JOIN courses c ON c.id=uc.course_id WHERE uc.user_id=? AND c.code=? AND uc.status='active' AND c.status='active' FOR NO KEY UPDATE OF uc`, userID, course).Scan(&ucID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, dailyjourney.ErrAccess
	}
	if err != nil {
		return nil, err
	}
	date := now.Format("2006-01-02")
	var existing string
	err = tx.QueryRowContext(ctx, `SELECT id FROM daily_journeys WHERE user_course_id=? AND journey_date=?::date`, ucID, date).Scan(&existing)
	if err == nil {
		j, err := loadJourney(ctx, tx, userID, course, existing, true)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return j, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	activeRows, err := tx.QueryContext(ctx, `SELECT id FROM daily_journeys WHERE user_course_id=? AND status='active' ORDER BY id FOR UPDATE`, ucID)
	if err != nil {
		return nil, err
	}
	activeIDs := []string{}
	for activeRows.Next() {
		var activeID string
		if err := activeRows.Scan(&activeID); err != nil {
			activeRows.Close()
			return nil, err
		}
		activeIDs = append(activeIDs, activeID)
	}
	if err := activeRows.Err(); err != nil {
		activeRows.Close()
		return nil, err
	}
	if err := activeRows.Close(); err != nil {
		return nil, err
	}
	for _, activeID := range activeIDs {
		if err := journeyProcessingBarrier(ctx, tx, activeID); err != nil {
			return nil, err
		}
	}
	steps, err := build(tx)
	if err != nil {
		return nil, err
	}
	j := &dailyjourney.Journey{ID: uuid.NewString(), CourseCode: course, Date: date, DayTimezone: now.Location().String(), Status: "active", AlgorithmVersion: dailyjourney.AlgorithmVersion, Steps: steps}
	if j.DayTimezone == "Local" {
		j.DayTimezone = now.Format("MST -07:00")
	}
	j.Resolve()
	if _, err := tx.ExecContext(ctx, `UPDATE daily_journeys SET status='archived', revision=revision+1, updated_at=CURRENT_TIMESTAMP WHERE user_course_id=? AND status='active'`, ucID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO daily_journeys (id, user_course_id, journey_date, day_timezone, status, algorithm_version) VALUES (?, ?, ?::date, ?, ?, ?)`, j.ID, ucID, j.Date, j.DayTimezone, j.Status, j.AlgorithmVersion); err != nil {
		return nil, err
	}
	for i := range j.Steps {
		s := &j.Steps[i]
		s.ID = uuid.NewString()
		if len(s.State) == 0 {
			s.State = json.RawMessage(`{}`)
		}
		if len(s.ContentRef) == 0 {
			s.ContentRef = json.RawMessage(`{}`)
		}
		launch, err := json.Marshal(s.Launch)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO daily_journey_steps (id, journey_id, position, mode, required, status, title, goal, content_ref_json, content_version, target_count, state_json, launch_json) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?, ?, ?::jsonb, ?::jsonb)`, s.ID, j.ID, i, s.Mode, s.Required, s.Status, s.Title, s.Goal, string(s.ContentRef), s.ContentVersion, s.TargetCount, string(s.State), string(launch))
		if err != nil {
			return nil, fmt.Errorf("insert daily journey step: %w", err)
		}
	}
	if err := recordJourneyLifecycle(ctx, tx, j, true); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return j, nil
}

// Apply puts the subject write (callback), itinerary cursor and idempotency receipt
// in ONE transaction. The callback must use tx and must never perform network I/O.
func (r *DailyJourneyRepository) Apply(ctx context.Context, userID int64, course, id, stepID string, action dailyjourney.Action, apply func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step) error) (*dailyjourney.Journey, error) {
	hash, err := journeyActionHash(stepID, action)
	if err != nil {
		return nil, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	j, err := loadJourney(ctx, tx, userID, course, id, true)
	if err != nil {
		return nil, err
	}
	var oldHash, receiptStatus string
	var saved []byte
	err = tx.QueryRowContext(ctx, `SELECT payload_hash, result_json,status FROM daily_journey_actions WHERE journey_id=? AND client_action_id=?`, id, action.ClientActionID).Scan(&oldHash, &saved, &receiptStatus)
	if err == nil {
		if hash != oldHash {
			return nil, dailyjourney.ErrActionKey
		}
		if receiptStatus != "completed" {
			if receiptStatus == "retryable" {
				return nil, dailyjourney.ErrConflict
			}
			return nil, dailyjourney.ErrProcessing
		}
		var result dailyjourney.Journey
		if err := json.Unmarshal(saved, &result); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &result, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
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
	if err := apply(tx, j, step); err != nil {
		return nil, err
	}
	if err := persistJourneyAction(ctx, tx, j, step, action, hash, ""); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return j, nil
}

func journeyActionHash(stepID string, action dailyjourney.Action) (string, error) {
	if action.ClientActionID == "" || strings.TrimSpace(action.ClientActionID) != action.ClientActionID || len(action.ClientActionID) > 100 || action.Type == "" || len(action.Type) > 40 || action.ExpectedRevision < 0 || len(action.Payload) > 32<<10 {
		return "", dailyjourney.ErrInvalidAction
	}
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader(action.Payload))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil || payload == nil || !json.Valid(action.Payload) {
		return "", dailyjourney.ErrInvalidAction
	}
	canonical, err := json.Marshal(struct {
		Step     string
		Action   string
		Revision int64
		Payload  any
	}{stepID, action.Type, action.ExpectedRevision, payload})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

func persistJourneyAction(ctx context.Context, tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action, hash, reservationToken string) error {
	j.Resolve()
	j.Revision++
	launch, err := json.Marshal(step.Launch)
	if err != nil {
		return err
	}
	if len(step.ResultDetails) == 0 {
		step.ResultDetails = json.RawMessage(`{}`)
	}

	_, err = tx.ExecContext(ctx, `UPDATE daily_journey_steps SET status=?, completed_count=?, state_json=?::jsonb, feedback=?, result=?, result_json=?::jsonb, target_count=?, launch_json=?::jsonb, skipped_at=CASE WHEN ?='skipped' THEN COALESCE(skipped_at,CURRENT_TIMESTAMP) ELSE skipped_at END, completed_at=CASE WHEN ?='completed' THEN COALESCE(completed_at,CURRENT_TIMESTAMP) ELSE completed_at END WHERE id=? AND journey_id=?`, step.Status, step.CompletedCount, string(step.State), step.Feedback, step.Result, string(step.ResultDetails), step.TargetCount, string(launch), step.Status, step.Status, step.ID, j.ID)
	if err != nil {
		return fmt.Errorf("update daily journey step: %w", err)
	}
	_, err = tx.ExecContext(ctx, `UPDATE daily_journeys SET status=?, revision=?, updated_at=CURRENT_TIMESTAMP, completed_at=CASE WHEN ?='completed' THEN COALESCE(completed_at,CURRENT_TIMESTAMP) ELSE completed_at END WHERE id=?`, j.Status, j.Revision, j.Status, j.ID)
	if err != nil {
		return err
	}
	result, err := json.Marshal(j)
	if err != nil {
		return err
	}

	if reservationToken == "" {
		_, err = tx.ExecContext(ctx, `INSERT INTO daily_journey_actions(journey_id,client_action_id,step_id,action_type,payload_hash,status,result_json) VALUES(?,?,?,?,?,'completed',?::jsonb)`, j.ID, action.ClientActionID, step.ID, action.Type, hash, string(result))
	} else {
		var changed sql.Result
		changed, err = tx.ExecContext(ctx, `UPDATE daily_journey_actions SET status='completed',result_json=?::jsonb,prepared_json=NULL,reservation_token=NULL,lease_expires_at=NULL,updated_at=CURRENT_TIMESTAMP WHERE journey_id=? AND client_action_id=? AND payload_hash=? AND reservation_token=? AND status='processing'`, string(result), j.ID, action.ClientActionID, hash, reservationToken)
		if err == nil {
			n, e := changed.RowsAffected()
			if e != nil {
				return e
			}
			if n != 1 {
				return dailyjourney.ErrReservationLost
			}
		}
	}
	if err != nil {
		return err
	}
	return recordJourneyLifecycle(ctx, tx, j, false)
}

// Lifecycle events share the subject transaction and have no user-authored payload.
// Unique keys make retries and repeated snapshots harmless.
func recordJourneyLifecycle(ctx context.Context, tx *sql.Tx, j *dailyjourney.Journey, started bool) error {
	record := func(kind string, step any) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO daily_journey_events(journey_id,step_id,event_type,revision) VALUES(?,?,?,?) ON CONFLICT DO NOTHING`, j.ID, step, kind, j.Revision)
		return err
	}
	if started {
		if err := record("journey_started", nil); err != nil {
			return err
		}
	}
	for _, step := range j.Steps {
		switch step.Status {
		case "completed", "skipped", "unavailable":
			if err := record("step_"+step.Status, step.ID); err != nil {
				return err
			}
		}
	}
	if j.Status == "completed" || j.Status == "unavailable" {
		return record("journey_"+j.Status, nil)
	}
	return nil
}
