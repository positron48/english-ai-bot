package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"tgbot-skeleton/internal/placement"
	"time"
)

var ErrPlacementNotFound = errors.New("placement session not found")
var ErrPlacementExpired = errors.New("placement session expired")
var ErrPlacementConflict = errors.New("placement session conflict")
var ErrPlacementAnswer = errors.New("invalid placement answer")

type PlacementSession struct {
	ID            string
	UserCourseID  int64
	CourseCode    string
	BankVersion   string
	PolicyVersion string
	Status        string
	ExpiresAt     time.Time
	Snapshot      placement.Snapshot
	Result        *placement.Result
}
type PlacementAttemptRepository struct{ db *sql.DB }

func NewPlacementAttemptRepository(db *sql.DB) *PlacementAttemptRepository {
	return &PlacementAttemptRepository{db}
}

const placementSessionSelect = `SELECT ps.id,ps.user_course_id,c.code,ps.bank_version,ps.policy_version,ps.status,ps.expires_at,ps.snapshot_json,ps.result_json
	FROM placement_sessions ps JOIN user_courses uc ON uc.id=ps.user_course_id JOIN courses c ON c.id=uc.course_id`

type placementScanner interface{ Scan(...interface{}) error }

func scanPlacement(row placementScanner) (*PlacementSession, error) {
	var s PlacementSession
	var snapshot, result []byte
	if err := row.Scan(&s.ID, &s.UserCourseID, &s.CourseCode, &s.BankVersion, &s.PolicyVersion, &s.Status, &s.ExpiresAt, &snapshot, &result); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlacementNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(snapshot, &s.Snapshot); err != nil {
		return nil, err
	}
	if len(result) > 0 {
		if err := json.Unmarshal(result, &s.Result); err != nil {
			return nil, err
		}
	}
	if s.Snapshot.Answers == nil {
		s.Snapshot.Answers = map[string]string{}
	}
	return &s, nil
}
func (r *PlacementAttemptRepository) Get(ctx context.Context, user int64, id string) (*PlacementSession, error) {
	s, err := scanPlacement(r.db.QueryRowContext(ctx, placementSessionSelect+` WHERE ps.id=? AND uc.user_id=?`, id, user))
	if err != nil {
		return nil, err
	}
	if s.Status == "abandoned" || (s.Status == "active" && (time.Now().After(s.ExpiresAt) || s.PolicyVersion != placement.PolicyVersion)) {
		return nil, ErrPlacementExpired
	}
	return s, nil
}

// Start locks the enrollment, so concurrent tabs cannot obtain overlapping new variants.
func (r *PlacementAttemptRepository) Start(ctx context.Context, user, userCourse int64, course, key string, newAttempt bool, build func(map[string]int) (string, placement.Snapshot, error)) (*PlacementSession, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var owned int64
	if err = tx.QueryRowContext(ctx, `SELECT id FROM user_courses WHERE id=? AND user_id=? FOR UPDATE`, userCourse, user).Scan(&owned); err != nil {
		return nil, ErrPlacementNotFound
	}
	previous, err := scanPlacement(tx.QueryRowContext(ctx, placementSessionSelect+` JOIN placement_session_requests pr ON pr.session_id=ps.id WHERE pr.user_course_id=? AND pr.idempotency_key=?`, userCourse, key))
	if err == nil {
		if previous.Status == "abandoned" || (previous.Status == "active" && (time.Now().After(previous.ExpiresAt) || previous.PolicyVersion != placement.PolicyVersion)) {
			return nil, ErrPlacementExpired
		}
		return previous, nil
	}
	if !errors.Is(err, ErrPlacementNotFound) {
		return nil, err
	}
	active, err := scanPlacement(tx.QueryRowContext(ctx, placementSessionSelect+` WHERE ps.user_course_id=? AND ps.status='active' FOR UPDATE OF ps`, userCourse))
	if err != nil && !errors.Is(err, ErrPlacementNotFound) {
		return nil, err
	}
	if active != nil && !newAttempt && time.Now().Before(active.ExpiresAt) && active.PolicyVersion == placement.PolicyVersion {
		if _, err = tx.ExecContext(ctx, `INSERT INTO placement_session_requests(user_course_id,idempotency_key,session_id) VALUES(?,?,?)`, userCourse, key, active.ID); err != nil {
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return active, nil
	}
	if active != nil {
		if _, err = tx.ExecContext(ctx, `UPDATE placement_sessions SET status='abandoned',updated_at=CURRENT_TIMESTAMP WHERE id=? AND status='active'`, active.ID); err != nil {
			return nil, err
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT snapshot_json FROM placement_sessions WHERE user_course_id=? ORDER BY created_at DESC LIMIT 12`, userCourse)
	if err != nil {
		return nil, err
	}
	recent := map[string]int{}
	rank := 0
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			rows.Close()
			return nil, err
		}
		var snap placement.Snapshot
		if err = json.Unmarshal(raw, &snap); err != nil {
			rows.Close()
			return nil, err
		}
		rank++
		for _, q := range snap.Items {
			if recent[q.FamilyID] == 0 {
				recent[q.FamilyID] = rank
			}
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	version, snap, err := build(recent)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}
	idBytes := make([]byte, 16)
	if _, err = rand.Read(idBytes); err != nil {
		return nil, err
	}
	id := hex.EncodeToString(idBytes)
	expires := time.Now().UTC().Add(7 * 24 * time.Hour)
	if _, err = tx.ExecContext(ctx, `INSERT INTO placement_sessions(id,user_course_id,idempotency_key,bank_version,policy_version,snapshot_json,expires_at) VALUES(?,?,?,?,?,?::jsonb,?)`, id, userCourse, key, version, placement.PolicyVersion, string(raw), expires); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO placement_session_requests(user_course_id,idempotency_key,session_id) VALUES(?,?,?)`, userCourse, key, id); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &PlacementSession{ID: id, UserCourseID: userCourse, CourseCode: course, BankVersion: version, PolicyVersion: placement.PolicyVersion, Status: "active", ExpiresAt: expires, Snapshot: snap}, nil
}

// Update serializes answers/finish and commits access in the same transaction as the result.
func (r *PlacementAttemptRepository) Update(ctx context.Context, user int64, id string, fn func(*PlacementSession, *sql.Tx) error) (*PlacementSession, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Same lock order as Start and admin access changes: enrollment, then session.
	// Otherwise a new start could overwrite a concurrent completion, or deadlock
	// against the access table's enrollment foreign key during finish.
	var userCourse int64
	if err = tx.QueryRowContext(ctx, `SELECT ps.user_course_id FROM placement_sessions ps JOIN user_courses uc ON uc.id=ps.user_course_id WHERE ps.id=? AND uc.user_id=?`, id, user).Scan(&userCourse); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlacementNotFound
		}
		return nil, err
	}
	var locked int64
	if err = tx.QueryRowContext(ctx, `SELECT id FROM user_courses WHERE id=? AND user_id=? FOR UPDATE`, userCourse, user).Scan(&locked); err != nil {
		return nil, err
	}
	s, err := scanPlacement(tx.QueryRowContext(ctx, placementSessionSelect+` WHERE ps.id=? AND uc.user_id=? FOR UPDATE OF ps`, id, user))
	if err != nil {
		return nil, err
	}
	if s.Status == "abandoned" {
		return nil, ErrPlacementExpired
	}
	if s.Status == "active" && (time.Now().After(s.ExpiresAt) || s.PolicyVersion != placement.PolicyVersion) {
		return nil, ErrPlacementExpired
	}
	if err = fn(s, tx); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(s.Snapshot)
	if err != nil {
		return nil, err
	}
	var result interface{}
	if s.Result != nil {
		b, e := json.Marshal(s.Result)
		if e != nil {
			return nil, e
		}
		result = string(b)
	}
	if _, err = tx.ExecContext(ctx, `UPDATE placement_sessions SET snapshot_json=?::jsonb,result_json=?::jsonb,status=?,updated_at=CURRENT_TIMESTAMP,
		completed_at=CASE WHEN ?='completed' THEN COALESCE(completed_at,CURRENT_TIMESTAMP) ELSE completed_at END WHERE id=?`, string(raw), result, s.Status, s.Status, id); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return s, nil
}
func (r *PlacementAttemptRepository) History(ctx context.Context, user int64, course string) ([]*PlacementSession, error) {
	rows, err := r.db.QueryContext(ctx, placementSessionSelect+` WHERE uc.user_id=? AND c.code=? AND ps.status='completed' ORDER BY ps.completed_at DESC LIMIT 10`, user, course)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*PlacementSession{}
	for rows.Next() {
		s, err := scanPlacement(rows)
		if err != nil {
			return nil, fmt.Errorf("placement history: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
