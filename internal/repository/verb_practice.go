package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"tgbot-skeleton/internal/verbtraining"
	"time"
)

// VerbPractice keeps the current question and its feedback durable across reloads.
// JSON uses the existing session/event columns so older sessions remain readable.
type VerbPractice struct {
	Version   int                    `json:"version"`
	ID        int64                  `json:"id"`
	Queue     []VerbQueueCard        `json:"queue"`
	Index     int                    `json:"index"`
	Assisted  bool                   `json:"assisted"`
	Feedback  *VerbPracticeFeedback  `json:"feedback,omitempty"`
	Results   []VerbPracticeFeedback `json:"results"`
	Retry     bool                   `json:"retry"`
	Completed bool                   `json:"completed"`
}

type VerbPracticeFeedback struct {
	CardID      int64           `json:"card_id"`
	Outcome     string          `json:"outcome"`
	Assisted    bool            `json:"assisted"`
	Chosen      string          `json:"chosen_option"`
	Correct     string          `json:"correct_answer"`
	Sentence    string          `json:"sentence"`
	Translation string          `json:"translation"`
	Lemma       string          `json:"lemma"`
	Tense       string          `json:"tense"`
	Mood        string          `json:"mood"`
	Person      string          `json:"person"`
	Number      string          `json:"number"`
	Rule        json.RawMessage `json:"rule"`
	AccentOnly  bool            `json:"accent_only"`
}

func (r *VerbFormsRepository) LoadVerbPractice(userID int64) (*VerbPractice, error) {
	var raw string
	err := r.db.QueryRow(`SELECT session_json FROM verb_training_sessions WHERE user_id=$1 AND session_json::jsonb->>'version'='2' ORDER BY id DESC LIMIT 1`, userID).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var state VerbPractice
	if err = json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *VerbFormsRepository) CreateVerbPractice(userID int64, state *VerbPractice, scopes []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Serializes starts across tabs without blocking other users.
	if _, err = tx.Exec(`SELECT pg_advisory_xact_lock($1)`, userID); err != nil {
		return err
	}
	var existing string
	lookupErr := tx.QueryRow(`SELECT session_json FROM verb_training_sessions WHERE user_id=$1 AND session_json::jsonb->>'version'='2' AND ended_at IS NULL ORDER BY id DESC LIMIT 1`, userID).Scan(&existing)
	if lookupErr == nil {
		var active VerbPractice
		if json.Unmarshal([]byte(existing), &active) == nil && !active.Completed && active.Index < len(active.Queue) {
			var prompt struct {
				Tense string `json:"tense"`
				Mood  string `json:"mood"`
			}
			_ = json.Unmarshal([]byte(active.Queue[active.Index].PromptJSON), &prompt)
			permitted := false
			for _, scope := range scopes {
				if verbtraining.CanonicalScope(scope) == verbtraining.CanonicalScope("es."+prompt.Tense+"."+prompt.Mood) {
					permitted = true
				}
			}
			if permitted {
				// An already-created concurrent start wins. Callers resume this session.
				*state = active
				return tx.Commit()
			}
			if _, err = tx.Exec(`UPDATE verb_training_sessions SET ended_at=CURRENT_TIMESTAMP WHERE id=$1`, active.ID); err != nil {
				return err
			}
		}
	} else if lookupErr != sql.ErrNoRows {
		return lookupErr
	}
	if err = tx.QueryRow(`INSERT INTO verb_training_sessions(user_id,planned_count,done_count,session_json) VALUES($1,$2,0,'{}') RETURNING id`, userID, len(state.Queue)).Scan(&state.ID); err != nil {
		return err
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE verb_training_sessions SET session_json=$1 WHERE id=$2`, string(raw), state.ID); err != nil {
		return err
	}
	return tx.Commit()
}

// ChangeVerbPractice commits feedback, SRS and position under one row lock.
func (r *VerbFormsRepository) ChangeVerbPractice(ctx context.Context, userID, sessionID int64, change func(*VerbPractice, *sql.Tx) error) (*VerbPractice, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var raw string
	if err = tx.QueryRowContext(ctx, `SELECT session_json FROM verb_training_sessions WHERE id=$1 AND user_id=$2 FOR UPDATE`, sessionID, userID).Scan(&raw); err != nil {
		return nil, err
	}
	var state VerbPractice
	if err = json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, err
	}
	if state.Version != 2 {
		return nil, fmt.Errorf("unsupported session")
	}
	if err = change(&state, tx); err != nil {
		return nil, err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE verb_training_sessions SET session_json=$1,done_count=$2,ended_at=CASE WHEN $3 THEN CURRENT_TIMESTAMP ELSE ended_at END WHERE id=$4`, string(data), len(state.Results), state.Completed, state.ID); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &state, nil
}

func ReadVerbSRSTx(tx *sql.Tx, userID, cardID int64) (*VerbUserCardSRS, error) {
	var c VerbUserCardSRS
	err := tx.QueryRow(`SELECT id,state,ef,reps,interval_days,learning_step,lapse_count FROM user_verb_cards WHERE id=$1 AND user_id=$2 FOR UPDATE`, cardID, userID).Scan(&c.ID, &c.State, &c.EF, &c.Reps, &c.IntervalDays, &c.LearningStep, &c.LapseCount)
	return &c, err
}

func SaveVerbAttemptTx(tx *sql.Tx, userID int64, state *VerbPractice, c *VerbUserCardSRS, next time.Time, quality int, feedback VerbPracticeFeedback) error {
	if !state.Retry {
		if _, err := tx.Exec(`UPDATE user_verb_cards SET state=$1,ef=$2,reps=$3,interval_days=$4,learning_step=$5,lapse_count=$6,next_due_at=$7,last_review_at=CURRENT_TIMESTAMP,last_quality=$8,updated_at=CURRENT_TIMESTAMP WHERE id=$9 AND user_id=$10`, c.State, c.EF, c.Reps, c.IntervalDays, c.LearningStep, c.LapseCount, next, quality, c.ID, userID); err != nil {
			return err
		}
	}
	metrics, err := json.Marshal(map[string]interface{}{"version": 2, "retry": state.Retry, "feedback": feedback})
	if err != nil {
		return err
	}
	correct := 0
	if feedback.Outcome == "correct" && !feedback.Assisted {
		correct = 1
	}
	_, err = tx.Exec(`INSERT INTO verb_review_events(session_id,user_id,user_verb_card_id,answered_at,is_correct,quality,metrics_json) VALUES($1,$2,$3,CURRENT_TIMESTAMP,$4,$5,$6)`, state.ID, userID, feedback.CardID, correct, quality, string(metrics))
	return err
}
