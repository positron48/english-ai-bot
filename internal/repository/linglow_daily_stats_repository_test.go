package repository

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestLinglowDailyStatsRepository_BumpAndBackfill(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(556677)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := NewCourseRepository(conn, logger)
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	if _, err := courseRepo.BackfillUserCoursesForLearning(context.Background(), lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}

	repo := NewLinglowDailyStatsRepository(conn)
	userCourseID, err := repo.ResolveUserCourseID(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("ResolveUserCourseID: %v", err)
	}

	day := LocalDayFromTime(time.Now())
	for i := 0; i < 3; i++ {
		if err := repo.Bump(context.Background(), DailyBump{
			UserCourseID: userCourseID,
			Day:          day,
			Mode:         "word_training",
			Attempts:     1,
			Correct:      1,
		}); err != nil {
			t.Fatalf("Bump: %v", err)
		}
	}
	if err := repo.Bump(context.Background(), DailyBump{
		UserCourseID:  userCourseID,
		Day:           day,
		ActiveSeconds: 90,
	}); err != nil {
		t.Fatalf("Bump seconds: %v", err)
	}

	var attempts, correct, seconds int
	if err := conn.QueryRow(`
		SELECT attempt_count, correct_count, active_seconds
		FROM daily_course_stats WHERE user_course_id = ? AND local_date = CAST(? AS date)
	`, userCourseID, day).Scan(&attempts, &correct, &seconds); err != nil {
		t.Fatalf("read daily_course_stats: %v", err)
	}
	if attempts != 3 || correct != 3 || seconds != 90 {
		t.Fatalf("unexpected daily stats: attempts=%d correct=%d seconds=%d", attempts, correct, seconds)
	}

	var modeAttempts int
	if err := conn.QueryRow(`
		SELECT attempt_count FROM mode_daily_stats
		WHERE user_course_id = ? AND local_date = CAST(? AS date) AND mode = 'word_training'
	`, userCourseID, day).Scan(&modeAttempts); err != nil {
		t.Fatalf("read mode_daily_stats: %v", err)
	}
	if modeAttempts != 3 {
		t.Fatalf("unexpected mode attempts: %d", modeAttempts)
	}

	// Backfill should not reduce heartbeat seconds and should overwrite counters from attempts.
	if _, err := conn.Exec(`
		INSERT INTO exercise_attempts (user_course_id, mode, started_at, answered_at, is_correct)
		VALUES (?, 'word_training', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, true)
	`, userCourseID); err != nil {
		t.Fatalf("insert attempt: %v", err)
	}
	if _, err := repo.Backfill(context.Background()); err != nil {
		t.Fatalf("Backfill: %v", err)
	}
	if err := conn.QueryRow(`
		SELECT attempt_count, active_seconds FROM daily_course_stats
		WHERE user_course_id = ? AND local_date = CAST(? AS date)
	`, userCourseID, day).Scan(&attempts, &seconds); err != nil {
		t.Fatalf("read after backfill: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("backfill should overwrite attempt_count from attempts, got %d", attempts)
	}
	if seconds < 90 {
		t.Fatalf("backfill must not reduce active_seconds, got %d", seconds)
	}
}

func TestValidClientDay(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	if got := ValidClientDay("garbage"); got != today {
		t.Fatalf("expected server day for garbage, got %s", got)
	}
	if got := ValidClientDay("1999-01-01"); got != today {
		t.Fatalf("expected server day for far past, got %s", got)
	}
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if got := ValidClientDay(yesterday); got != yesterday {
		t.Fatalf("expected yesterday accepted, got %s", got)
	}
}
