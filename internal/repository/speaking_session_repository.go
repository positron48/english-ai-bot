package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SpeakingSession represents an active or completed speaking practice session.
type SpeakingSession struct {
	ID               int64
	UserID           int64
	CategoryID       string
	Status           string
	TaskIDs          []string
	CurrentTaskIndex int
	StartedAt        time.Time
	CompletedAt      *time.Time
}

// SpeakingAttemptRecord is a stored evaluation attempt.
type SpeakingAttemptRecord struct {
	ID                 int64
	UserID             int64
	SessionID          int64
	TaskID             string
	AttemptNo          int
	Mode               string
	UnderstoodAnswer   string
	MeaningScore       *int
	GrammarScore       *int
	PronunciationScore *int
	FluencyScore       *int
	IsAcceptable       *bool
	AudioQuality       string
	FeedbackRU         string
	BetterVersion      string
	RepeatTask         string
	CreatedAt          time.Time
}

// SpeakingSessionRepository handles speaking sessions and attempts.
type SpeakingSessionRepository struct {
	db *sql.DB
}

func NewSpeakingSessionRepository(db *sql.DB) *SpeakingSessionRepository {
	if db == nil {
		return nil
	}
	return &SpeakingSessionRepository{db: db}
}

func (r *SpeakingSessionRepository) CreateSession(userID int64, categoryID string, taskIDs []string) (*SpeakingSession, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("speaking session repo: nil db")
	}
	raw, err := json.Marshal(taskIDs)
	if err != nil {
		return nil, err
	}
	var id int64
	err = r.db.QueryRow(`
INSERT INTO speaking_sessions (user_id, category_id, status, task_ids, current_task_index)
VALUES (?, ?, 'active', ?, 0)
RETURNING id`, userID, categoryID, string(raw)).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create speaking session: %w", err)
	}
	return r.GetSession(id, userID)
}

func (r *SpeakingSessionRepository) GetSession(sessionID, userID int64) (*SpeakingSession, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("speaking session repo: nil db")
	}
	const q = `
SELECT id, user_id, category_id, status, task_ids, current_task_index, started_at, completed_at
FROM speaking_sessions WHERE id = ? AND user_id = ?`
	var s SpeakingSession
	var taskJSON string
	var completed sql.NullString
	var started string
	err := r.db.QueryRow(q, sessionID, userID).Scan(
		&s.ID, &s.UserID, &s.CategoryID, &s.Status, &taskJSON, &s.CurrentTaskIndex, &started, &completed,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(taskJSON), &s.TaskIDs); err != nil {
		return nil, err
	}
	s.StartedAt, _ = time.Parse("2006-01-02 15:04:05", started)
	if completed.Valid && completed.String != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", completed.String)
		s.CompletedAt = &t
	}
	return &s, nil
}

func (r *SpeakingSessionRepository) CountAttempts(sessionID int64, taskID string) (int, error) {
	var n int
	err := r.db.QueryRow(`
SELECT COUNT(*) FROM speaking_attempts WHERE session_id = ? AND task_id = ?`, sessionID, taskID).Scan(&n)
	return n, err
}

func (r *SpeakingSessionRepository) SaveAttempt(rec *SpeakingAttemptRecord) (int64, error) {
	var id int64
	err := r.db.QueryRow(`
INSERT INTO speaking_attempts (
  user_id, session_id, task_id, attempt_no, mode,
  understood_answer, meaning_score, grammar_score, pronunciation_score, fluency_score,
  is_acceptable, audio_quality, feedback_ru, better_version, repeat_task
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id`,
		rec.UserID, rec.SessionID, rec.TaskID, rec.AttemptNo, rec.Mode,
		rec.UnderstoodAnswer, rec.MeaningScore, rec.GrammarScore, rec.PronunciationScore, rec.FluencyScore,
		rec.IsAcceptable, rec.AudioQuality, rec.FeedbackRU, rec.BetterVersion, rec.RepeatTask,
	).Scan(&id)
	return id, err
}

func (r *SpeakingSessionRepository) AdvanceSession(sessionID, userID int64, nextIndex int, completed bool) error {
	if completed {
		_, err := r.db.Exec(`
UPDATE speaking_sessions SET current_task_index = ?, status = 'completed', completed_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?`, nextIndex, sessionID, userID)
		return err
	}
	_, err := r.db.Exec(`
UPDATE speaking_sessions SET current_task_index = ?
WHERE id = ? AND user_id = ?`, nextIndex, sessionID, userID)
	return err
}

func (r *SpeakingSessionRepository) LastAttemptForTask(sessionID int64, taskID string) (*SpeakingAttemptRecord, error) {
	const q = `
SELECT id, user_id, session_id, task_id, attempt_no, mode,
  COALESCE(understood_answer, ''), meaning_score, grammar_score, pronunciation_score, fluency_score,
  is_acceptable, COALESCE(audio_quality, ''), COALESCE(feedback_ru, ''), COALESCE(better_version, ''), COALESCE(repeat_task, ''),
  created_at
FROM speaking_attempts
WHERE session_id = ? AND task_id = ?
ORDER BY attempt_no DESC, id DESC
LIMIT 1`
	var rec SpeakingAttemptRecord
	var created string
	err := r.db.QueryRow(q, sessionID, taskID).Scan(
		&rec.ID, &rec.UserID, &rec.SessionID, &rec.TaskID, &rec.AttemptNo, &rec.Mode,
		&rec.UnderstoodAnswer, &rec.MeaningScore, &rec.GrammarScore, &rec.PronunciationScore, &rec.FluencyScore,
		&rec.IsAcceptable, &rec.AudioQuality, &rec.FeedbackRU, &rec.BetterVersion, &rec.RepeatTask, &created,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rec.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	return &rec, nil
}
