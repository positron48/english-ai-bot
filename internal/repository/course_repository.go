package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

// CourseRepository handles Linglow v2 course and user-course data.
type CourseRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// UserCourseBackfillSummary describes one idempotent user_courses bootstrap run.
type UserCourseBackfillSummary struct {
	CourseCode   string
	CourseID     int64
	UsersScanned int64
	Existing     int64
	Created      int64
}

func NewCourseRepository(db *sql.DB, logger *zap.Logger) *CourseRepository {
	return &CourseRepository{db: db, logger: logger}
}

// CourseCodeForLearning returns the target canonical course code for a language pair.
func CourseCodeForLearning(lc config.LearningConfig) string {
	target := strings.TrimSpace(strings.ToLower(lc.TargetLang))
	native := strings.TrimSpace(strings.ToLower(lc.NativeLang))
	if target == "" || native == "" {
		return ""
	}
	return target + "_" + native
}

// BackfillUserCoursesForLearning idempotently attaches all existing users to the current course.
func (r *CourseRepository) BackfillUserCoursesForLearning(ctx context.Context, lc config.LearningConfig) (*UserCourseBackfillSummary, error) {
	return r.BackfillUserCourses(ctx, CourseCodeForLearning(lc))
}

// BackfillUserCourses idempotently attaches all existing users to courseCode.
func (r *CourseRepository) BackfillUserCourses(ctx context.Context, courseCode string) (*UserCourseBackfillSummary, error) {
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}

	var courseID int64
	if err := r.db.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = ?`, courseCode).Scan(&courseID); err != nil {
		return nil, fmt.Errorf("get course %q: %w", courseCode, err)
	}

	var usersScanned int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&usersScanned); err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	var existing int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_courses WHERE course_id = ?`, courseID).Scan(&existing); err != nil {
		return nil, fmt.Errorf("count existing user courses: %w", err)
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO user_courses (user_id, course_id, status, started_at, created_at, updated_at)
		SELECT u.id, ?, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM users u
		WHERE NOT EXISTS (
			SELECT 1 FROM user_courses uc
			WHERE uc.user_id = u.id AND uc.course_id = ?
		)
	`, courseID, courseID)
	if err != nil {
		return nil, fmt.Errorf("insert missing user courses: %w", err)
	}

	created, _ := result.RowsAffected()
	return &UserCourseBackfillSummary{
		CourseCode:   courseCode,
		CourseID:     courseID,
		UsersScanned: usersScanned,
		Existing:     existing,
		Created:      created,
	}, nil
}
