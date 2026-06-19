package repository

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestGetStatsForUser_StreakAndWeek(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(778899)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	courseRepo := NewCourseRepository(conn, logger)
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	if _, err := courseRepo.BackfillUserCoursesForLearning(context.Background(), lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	statsRepo := NewLinglowDailyStatsRepository(conn)
	userCourseID, err := statsRepo.ResolveUserCourseID(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("ResolveUserCourseID: %v", err)
	}

	// Active yesterday and the day before; today empty → streak 2, today not active.
	for _, daysAgo := range []int{1, 2} {
		day := time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02")
		if err := statsRepo.Bump(context.Background(), DailyBump{
			UserCourseID: userCourseID, Day: day, Mode: "word_training", Attempts: 5, Correct: 4, ActiveSeconds: 300,
		}); err != nil {
			t.Fatalf("Bump: %v", err)
		}
	}

	stats, err := courseRepo.GetStatsForUser(context.Background(), user.ID, "es_ru", "", "")
	if err != nil {
		t.Fatalf("GetStatsForUser: %v", err)
	}
	if stats.Streak.CurrentDays != 2 {
		t.Fatalf("expected streak 2, got %d", stats.Streak.CurrentDays)
	}
	if stats.Streak.TodayActive {
		t.Fatalf("today must not be active")
	}
	if len(stats.Week) != 7 {
		t.Fatalf("expected 7 week days, got %d", len(stats.Week))
	}
	if stats.Week[6].Status != "today" {
		t.Fatalf("last week day should be today, got %s", stats.Week[6].Status)
	}
	if stats.Week[5].Status != "done" {
		t.Fatalf("yesterday should be done, got %s", stats.Week[5].Status)
	}
	if len(stats.Skills) == 0 || stats.Skills[0].Mode != "word_training" {
		t.Fatalf("expected word_training skill, got %+v", stats.Skills)
	}

	// Today becomes active → streak 3.
	today := time.Now().Format("2006-01-02")
	if err := statsRepo.Bump(context.Background(), DailyBump{UserCourseID: userCourseID, Day: today, Attempts: 1, Correct: 1}); err != nil {
		t.Fatalf("Bump today: %v", err)
	}
	stats, err = courseRepo.GetStatsForUser(context.Background(), user.ID, "es_ru", "", "")
	if err != nil {
		t.Fatalf("GetStatsForUser 2: %v", err)
	}
	if stats.Streak.CurrentDays != 3 || !stats.Streak.TodayActive {
		t.Fatalf("expected streak 3 with active today, got %d / %v", stats.Streak.CurrentDays, stats.Streak.TodayActive)
	}
	if stats.Streak.BestDays < 3 {
		t.Fatalf("best streak should be >= 3, got %d", stats.Streak.BestDays)
	}
	if stats.Today.AttemptCount != 1 {
		t.Fatalf("today attempts: %d", stats.Today.AttemptCount)
	}
}
