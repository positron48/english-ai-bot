package repository

import (
	"context"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestCourseCodeForLearning(t *testing.T) {
	got := CourseCodeForLearning(config.LearningConfig{NativeLang: "ru", TargetLang: "es"})
	if got != "es_ru" {
		t.Fatalf("CourseCodeForLearning() = %q, want es_ru", got)
	}
}

func TestCourseRepository_BackfillUserCoursesForLearning(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	for _, telegramID := range []int64{1001, 1002, 1003} {
		if _, err := userRepo.GetOrCreateUser(telegramID); err != nil {
			t.Fatalf("create user %d: %v", telegramID, err)
		}
	}

	repo := NewCourseRepository(conn, logger)
	summary, err := repo.BackfillUserCoursesForLearning(context.Background(), config.LearningConfig{NativeLang: "ru", TargetLang: "es"})
	if err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	if summary.CourseCode != "es_ru" {
		t.Fatalf("course code = %q, want es_ru", summary.CourseCode)
	}
	if summary.UsersScanned != 3 || summary.Existing != 0 || summary.Created != 3 {
		t.Fatalf("unexpected first summary: %+v", summary)
	}

	second, err := repo.BackfillUserCoursesForLearning(context.Background(), config.LearningConfig{NativeLang: "ru", TargetLang: "es"})
	if err != nil {
		t.Fatalf("second BackfillUserCoursesForLearning: %v", err)
	}
	if second.UsersScanned != 3 || second.Existing != 3 || second.Created != 0 {
		t.Fatalf("unexpected idempotent summary: %+v", second)
	}

	var count int
	if err := conn.QueryRow(`
		SELECT COUNT(*)
		FROM user_courses uc
		JOIN courses c ON c.id = uc.course_id
		WHERE c.code = 'es_ru'
	`).Scan(&count); err != nil {
		t.Fatalf("count user_courses: %v", err)
	}
	if count != 3 {
		t.Fatalf("user_courses count = %d, want 3", count)
	}
}
