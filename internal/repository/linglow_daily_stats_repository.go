package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// LinglowDailyStatsRepository maintains daily_course_stats and mode_daily_stats aggregates.
type LinglowDailyStatsRepository struct {
	db *sql.DB
}

func NewLinglowDailyStatsRepository(db *sql.DB) *LinglowDailyStatsRepository {
	return &LinglowDailyStatsRepository{db: db}
}

// DailyBump is a single increment of the per-day aggregates.
type DailyBump struct {
	UserCourseID  int64
	Day           string // local date, YYYY-MM-DD
	Mode          string // optional; also bumps mode_daily_stats when set
	Attempts      int
	Correct       int
	ActiveSeconds int
	ReviewCount   int
	NewCount      int
}

// LocalDayFromTime formats the aggregate day key for an event timestamp.
func LocalDayFromTime(t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	return t.Format("2006-01-02")
}

// ValidClientDay returns clientDay when it is a parseable date within ±1 day
// of the server date, otherwise the server date.
func ValidClientDay(clientDay string) string {
	now := time.Now()
	parsed, err := time.Parse("2006-01-02", clientDay)
	if err != nil {
		return now.Format("2006-01-02")
	}
	diff := parsed.Sub(now.Truncate(24 * time.Hour))
	if diff > 48*time.Hour || diff < -48*time.Hour {
		return now.Format("2006-01-02")
	}
	return clientDay
}

// Bump upserts the daily aggregates for one user course/day.
func (r *LinglowDailyStatsRepository) Bump(ctx context.Context, b DailyBump) error {
	if b.UserCourseID == 0 {
		return fmt.Errorf("user course id is empty")
	}
	if b.Day == "" {
		b.Day = LocalDayFromTime(time.Time{})
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO daily_course_stats (user_course_id, local_date, review_count, new_count, correct_count, attempt_count, active_seconds, updated_at)
		VALUES (?, CAST(? AS date), ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_course_id, local_date) DO UPDATE SET
			review_count   = daily_course_stats.review_count + EXCLUDED.review_count,
			new_count      = daily_course_stats.new_count + EXCLUDED.new_count,
			correct_count  = daily_course_stats.correct_count + EXCLUDED.correct_count,
			attempt_count  = daily_course_stats.attempt_count + EXCLUDED.attempt_count,
			active_seconds = daily_course_stats.active_seconds + EXCLUDED.active_seconds,
			updated_at     = CURRENT_TIMESTAMP
	`, b.UserCourseID, b.Day, b.ReviewCount, b.NewCount, b.Correct, b.Attempts, b.ActiveSeconds); err != nil {
		return fmt.Errorf("bump daily_course_stats: %w", err)
	}
	if b.Mode == "" {
		return nil
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO mode_daily_stats (user_course_id, local_date, mode, attempt_count, correct_count, active_seconds, updated_at)
		VALUES (?, CAST(? AS date), ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_course_id, local_date, mode) DO UPDATE SET
			attempt_count  = mode_daily_stats.attempt_count + EXCLUDED.attempt_count,
			correct_count  = mode_daily_stats.correct_count + EXCLUDED.correct_count,
			active_seconds = mode_daily_stats.active_seconds + EXCLUDED.active_seconds,
			updated_at     = CURRENT_TIMESTAMP
	`, b.UserCourseID, b.Day, b.Mode, b.Attempts, b.Correct, b.ActiveSeconds); err != nil {
		return fmt.Errorf("bump mode_daily_stats: %w", err)
	}
	return nil
}

// ResolveUserCourseID returns the user_courses id for the user on the given course code.
func (r *LinglowDailyStatsRepository) ResolveUserCourseID(ctx context.Context, userID int64, courseCode string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		SELECT uc.id
		FROM user_courses uc
		JOIN courses c ON c.id = uc.course_id
		WHERE uc.user_id = ? AND c.code = ?
	`, userID, courseCode).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("resolve user course: %w", err)
	}
	return id, nil
}

// BackfillSummary reports what the historical aggregation rebuilt.
type BackfillSummary struct {
	CourseDays int
	ModeDays   int
}

// Backfill rebuilds counters from exercise_attempts history and estimates
// historical active_seconds as one minute per distinct active minute.
// Counter columns are overwritten; active_seconds keeps the maximum of the
// existing value and the estimate, so heartbeat data is never reduced.
func (r *LinglowDailyStatsRepository) Backfill(ctx context.Context) (BackfillSummary, error) {
	var summary BackfillSummary
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO daily_course_stats (user_course_id, local_date, review_count, new_count, correct_count, attempt_count, active_seconds, stats_json, updated_at)
		SELECT
			ea.user_course_id,
			CAST(COALESCE(ea.answered_at, ea.started_at) AS date) AS d,
			0, 0,
			COUNT(*) FILTER (WHERE ea.is_correct IS TRUE),
			COUNT(*),
			60 * COUNT(DISTINCT date_trunc('minute', COALESCE(ea.answered_at, ea.started_at))),
			'{"estimated":true}'::jsonb,
			CURRENT_TIMESTAMP
		FROM exercise_attempts ea
		GROUP BY ea.user_course_id, d
		ON CONFLICT (user_course_id, local_date) DO UPDATE SET
			correct_count  = EXCLUDED.correct_count,
			attempt_count  = EXCLUDED.attempt_count,
			active_seconds = GREATEST(daily_course_stats.active_seconds, EXCLUDED.active_seconds),
			updated_at     = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return summary, fmt.Errorf("backfill daily_course_stats: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil {
		summary.CourseDays = int(n)
	}

	res, err = r.db.ExecContext(ctx, `
		INSERT INTO mode_daily_stats (user_course_id, local_date, mode, attempt_count, correct_count, active_seconds, stats_json, updated_at)
		SELECT
			ea.user_course_id,
			CAST(COALESCE(ea.answered_at, ea.started_at) AS date) AS d,
			ea.mode,
			COUNT(*),
			COUNT(*) FILTER (WHERE ea.is_correct IS TRUE),
			60 * COUNT(DISTINCT date_trunc('minute', COALESCE(ea.answered_at, ea.started_at))),
			'{"estimated":true}'::jsonb,
			CURRENT_TIMESTAMP
		FROM exercise_attempts ea
		GROUP BY ea.user_course_id, d, ea.mode
		ON CONFLICT (user_course_id, local_date, mode) DO UPDATE SET
			attempt_count  = EXCLUDED.attempt_count,
			correct_count  = EXCLUDED.correct_count,
			active_seconds = GREATEST(mode_daily_stats.active_seconds, EXCLUDED.active_seconds),
			updated_at     = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return summary, fmt.Errorf("backfill mode_daily_stats: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil {
		summary.ModeDays = int(n)
	}
	return summary, nil
}
