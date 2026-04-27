package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type GrammarSRSRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

type GrammarTheoryMemory struct {
	ID           int64
	UserID       int64
	Language     string
	CourseID     string
	ChapterID    string
	TheoryBlockID string
	ConceptID    string
	State        string
	ReviewCount  int
	CorrectCount int
	WrongCount   int
	LapseCount   int
	CorrectStreak int
	WrongStreak  int
	Ease         float64
	IntervalDays int
	MasteryScore int
	NextReviewAt time.Time
	LastReviewAt *time.Time
}

func NewGrammarSRSRepository(db *sql.DB, logger *zap.Logger) *GrammarSRSRepository {
	return &GrammarSRSRepository{db: db, logger: logger}
}

func (r *GrammarSRSRepository) EnsureTheoryMemory(userID int64, language, courseID, chapterID, theoryBlockID, conceptID string) error {
	q := `INSERT INTO grammar_theory_memory (user_id, language, course_id, chapter_id, theory_block_id, concept_id, state, next_review_at)
	      VALUES (?, ?, ?, ?, ?, ?, 'new', CURRENT_TIMESTAMP)
	      ON CONFLICT(user_id, language, course_id, theory_block_id) DO NOTHING`
	if _, err := r.db.Exec(q, userID, language, courseID, chapterID, theoryBlockID, conceptID); err != nil {
		return fmt.Errorf("ensure grammar_theory_memory: %w", err)
	}
	return nil
}

func (r *GrammarSRSRepository) ListDueMemories(userID int64, language, courseID string, now time.Time, limit int) ([]*GrammarTheoryMemory, error) {
	if limit <= 0 {
		limit = 20
	}
	q := `SELECT id, user_id, language, course_id, chapter_id, theory_block_id, COALESCE(concept_id,''), state,
	             review_count, correct_count, wrong_count, lapse_count, correct_streak, wrong_streak,
	             ease, interval_days, mastery_score, next_review_at, last_review_at
	      FROM grammar_theory_memory
	      WHERE user_id=? AND language=? AND course_id=? AND next_review_at <= ?
	      ORDER BY next_review_at ASC
	      LIMIT ?`
	rows, err := r.db.Query(q, userID, language, courseID, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list due grammar memories: %w", err)
	}
	defer rows.Close()
	out := make([]*GrammarTheoryMemory, 0, limit)
	for rows.Next() {
		m := &GrammarTheoryMemory{}
		var last sql.NullTime
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Language, &m.CourseID, &m.ChapterID, &m.TheoryBlockID, &m.ConceptID, &m.State,
			&m.ReviewCount, &m.CorrectCount, &m.WrongCount, &m.LapseCount, &m.CorrectStreak, &m.WrongStreak,
			&m.Ease, &m.IntervalDays, &m.MasteryScore, &m.NextReviewAt, &last,
		); err != nil {
			return nil, fmt.Errorf("scan due grammar memory: %w", err)
		}
		if last.Valid {
			t := last.Time
			m.LastReviewAt = &t
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *GrammarSRSRepository) UpdateAfterAnswer(memory *GrammarTheoryMemory, isCorrect bool) error {
	now := time.Now()
	var next time.Time
	var state string
	mastery := memory.MasteryScore
	reviewCount := memory.ReviewCount + 1
	correctCount := memory.CorrectCount
	wrongCount := memory.WrongCount
	lapseCount := memory.LapseCount
	correctStreak := memory.CorrectStreak
	wrongStreak := memory.WrongStreak
	interval := memory.IntervalDays

	if isCorrect {
		correctCount++
		correctStreak++
		wrongStreak = 0
		if interval <= 0 {
			interval = 1
		} else {
			interval = interval * 2
			if interval > 30 {
				interval = 30
			}
		}
		next = now.Add(time.Duration(interval) * 24 * time.Hour)
		mastery += 5
		if mastery > 100 {
			mastery = 100
		}
		state = "review"
	} else {
		wrongCount++
		lapseCount++
		wrongStreak++
		correctStreak = 0
		interval = 0
		next = now.Add(15 * time.Minute)
		mastery -= 8
		if mastery < 0 {
			mastery = 0
		}
		state = "relearning"
	}

	q := `UPDATE grammar_theory_memory
	      SET state=?, review_count=?, correct_count=?, wrong_count=?, lapse_count=?,
	          correct_streak=?, wrong_streak=?, interval_days=?, mastery_score=?,
	          next_review_at=?, last_review_at=?, updated_at=CURRENT_TIMESTAMP
	      WHERE id=?`
	_, err := r.db.Exec(q, state, reviewCount, correctCount, wrongCount, lapseCount, correctStreak, wrongStreak, interval, mastery, next, now, memory.ID)
	if err != nil {
		return fmt.Errorf("update grammar memory after answer: %w", err)
	}
	return nil
}

func (r *GrammarSRSRepository) SaveAttempt(
	userID int64,
	language, courseID, chapterID, theoryBlockID, conceptID, questionID string,
	answerPayload interface{},
	correctPayload interface{},
	isCorrect bool,
) error {
	answerJSON, _ := json.Marshal(answerPayload)
	correctJSON, _ := json.Marshal(correctPayload)
	q := `INSERT INTO grammar_attempts
	      (user_id, language, course_id, chapter_id, theory_block_id, concept_id, question_id, question_source, is_correct, answer_payload_json, correct_payload_json, answered_at)
	      VALUES (?, ?, ?, ?, ?, ?, ?, 'training_pack', ?, ?, ?, CURRENT_TIMESTAMP)`
	if _, err := r.db.Exec(q, userID, language, courseID, chapterID, theoryBlockID, conceptID, questionID, isCorrect, string(answerJSON), string(correctJSON)); err != nil {
		return fmt.Errorf("insert grammar attempt: %w", err)
	}
	return nil
}

