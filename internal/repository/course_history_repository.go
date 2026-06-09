package repository

import (
	"context"
	"fmt"
	"time"
)

// HistoryDayStat is one day of training activity.
type HistoryDayStat struct {
	Day            string `json:"day"`
	CardsCompleted int    `json:"cards_completed"`
	CardsCorrect   int    `json:"cards_correct"`
}

// HistoryWordsAddedStat is one day of words added.
type HistoryWordsAddedStat struct {
	Day        string `json:"day"`
	WordsAdded int    `json:"words_added"`
}

// HistoryModeStat aggregates attempts by exercise mode over the window.
type HistoryModeStat struct {
	Mode         string `json:"mode"`
	AttemptCount int    `json:"attempt_count"`
	CorrectCount int    `json:"correct_count"`
}

// LinglowHistory is the course-scoped training history from canonical tables.
type LinglowHistory struct {
	Course          CourseMapCourse         `json:"course"`
	UserCourse      CourseMapUserCourse     `json:"user_course"`
	AccuracyPercent float64                 `json:"accuracy_percent"`
	TotalAttempts   int                     `json:"total_attempts"`
	CorrectAttempts int                     `json:"correct_attempts"`
	WeeklyStats     []HistoryDayStat        `json:"weekly_stats"`
	WordsAddedStats []HistoryWordsAddedStat `json:"words_added_stats"`
	ByMode          []HistoryModeStat       `json:"by_mode"`
	Generated       string                  `json:"generated_at"`
}

// GetHistoryForUser returns course-scoped training history (weekly activity, words added,
// accuracy, per-mode breakdown) from exercise_attempts / srs_items in the Linglow v2 schema.
func (r *CourseRepository) GetHistoryForUser(ctx context.Context, userID int64, defaultCourseCode, explicitCourseCode string, days int) (*LinglowHistory, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	if days <= 0 {
		days = 7
	}
	courseCode, err := r.ResolveRequestedCourseCode(ctx, userID, defaultCourseCode, explicitCourseCode)
	if err != nil {
		return nil, err
	}
	userCourse, err := r.EnsureUserCourse(ctx, userID, courseCode)
	if err != nil {
		return nil, err
	}
	courseMap, err := r.GetCourseMap(ctx, courseCode, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	windowStart := now.AddDate(0, 0, -days)
	monthAgo := now.AddDate(0, 0, -30)

	hist := &LinglowHistory{
		Course:     courseMap.Course,
		UserCourse: *userCourse,
		Generated:  now.Format("2006-01-02T15:04:05Z"),
	}
	if courseMap.UserCourse != nil {
		hist.UserCourse = *courseMap.UserCourse
	}

	// Accuracy over the last 30 days.
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0)
		FROM exercise_attempts
		WHERE user_course_id = ? AND answered_at IS NOT NULL AND answered_at >= ?
	`, userCourse.ID, monthAgo).Scan(&hist.TotalAttempts, &hist.CorrectAttempts); err != nil {
		return nil, fmt.Errorf("history accuracy: %w", err)
	}
	if hist.TotalAttempts > 0 {
		hist.AccuracyPercent = float64(hist.CorrectAttempts) / float64(hist.TotalAttempts) * 100
	}

	// Daily activity over the window.
	weekly, err := r.queryHistoryWeekly(ctx, userCourse.ID, windowStart)
	if err != nil {
		return nil, fmt.Errorf("history weekly: %w", err)
	}
	hist.WeeklyStats = weekly

	// Words added per day (new word SRS items created) over the window.
	wordsAdded, err := r.queryHistoryWordsAdded(ctx, userCourse.ID, windowStart)
	if err != nil {
		return nil, fmt.Errorf("history words added: %w", err)
	}
	hist.WordsAddedStats = wordsAdded

	// Per-mode breakdown over the window.
	byMode, err := r.queryHistoryByMode(ctx, userCourse.ID, windowStart)
	if err != nil {
		return nil, fmt.Errorf("history by mode: %w", err)
	}
	hist.ByMode = byMode

	return hist, nil
}

func (r *CourseRepository) queryHistoryWeekly(ctx context.Context, userCourseID int64, since time.Time) ([]HistoryDayStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT CAST(DATE(answered_at) AS TEXT) AS day,
		       COUNT(*) AS cards_completed,
		       COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0) AS cards_correct
		FROM exercise_attempts
		WHERE user_course_id = ? AND answered_at IS NOT NULL AND answered_at >= ?
		GROUP BY DATE(answered_at)
		ORDER BY day ASC
	`, userCourseID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HistoryDayStat
	for rows.Next() {
		var s HistoryDayStat
		if err := rows.Scan(&s.Day, &s.CardsCompleted, &s.CardsCorrect); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *CourseRepository) queryHistoryWordsAdded(ctx context.Context, userCourseID int64, since time.Time) ([]HistoryWordsAddedStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT CAST(DATE(si.created_at) AS TEXT) AS day,
		       COUNT(*) AS words_added
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		WHERE si.user_course_id = ? AND li.source_kind = 'word_card' AND si.created_at >= ?
		GROUP BY DATE(si.created_at)
		ORDER BY day ASC
	`, userCourseID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HistoryWordsAddedStat
	for rows.Next() {
		var s HistoryWordsAddedStat
		if err := rows.Scan(&s.Day, &s.WordsAdded); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *CourseRepository) queryHistoryByMode(ctx context.Context, userCourseID int64, since time.Time) ([]HistoryModeStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT mode,
		       COUNT(*) AS attempt_count,
		       COALESCE(SUM(CASE WHEN is_correct THEN 1 ELSE 0 END), 0) AS correct_count
		FROM exercise_attempts
		WHERE user_course_id = ? AND answered_at IS NOT NULL AND answered_at >= ?
		GROUP BY mode
		ORDER BY attempt_count DESC
	`, userCourseID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HistoryModeStat
	for rows.Next() {
		var s HistoryModeStat
		if err := rows.Scan(&s.Mode, &s.AttemptCount, &s.CorrectCount); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
