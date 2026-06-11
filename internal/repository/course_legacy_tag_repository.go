package repository

import (
	"context"
	"fmt"

	"tgbot-skeleton/internal/config"
)

// HasMultipleWordCourses reports whether this database serves word content for more than one
// course (i.e. training_cards tagged with 2+ distinct course_code values). It lets the runtime
// scope word training by course ONLY on the unified multi-course DB; single-course prod
// instances return false and never apply a course filter (zero behaviour change).
func (r *CourseRepository) HasMultipleWordCourses(ctx context.Context) bool {
	var distinct int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT DISTINCT course_code FROM training_cards WHERE course_code IS NOT NULL
		) t
	`).Scan(&distinct)
	if err != nil {
		return false
	}
	return distinct > 1
}

// LegacyCourseTagSummary reports how many legacy rows were tagged with a course_code.
type LegacyCourseTagSummary struct {
	CourseCode string
	Tagged     map[string]int64 // table -> rows tagged this run
}

// legacyCourseTagTables are the legacy per-user and content tables that the word-training
// engine reads. They gained a nullable course_code column in migration 000022.
var legacyCourseTagTables = []string{
	"word_cards",
	"training_cards",
	"user_cards",
	"review_events",
	"user_word_mastering",
	"user_word_knowledge",
	"word_set_categories",
	"word_sets",
}

// TagLegacyWordTablesForLearning is a zero-touch startup backfill: on a single-course
// instance (English -> en_ru, Spanish -> es_ru) it stamps the instance course_code onto
// legacy word-training rows that are still untagged (course_code IS NULL).
//
// Safe and idempotent:
//   - Only rows with course_code IS NULL are touched, so re-runs are no-ops.
//   - On the unified multi-course DB the legacy tables are populated by the merge tooling
//     with course_code already set per course, so this backfill finds nothing to tag and
//     never mis-labels es_ru rows even when the unified pod runs a transitional
//     LEARNING_APP_CODE=english config.
func (r *CourseRepository) TagLegacyWordTablesForLearning(ctx context.Context, lc config.LearningConfig) (*LegacyCourseTagSummary, error) {
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" {
		return nil, fmt.Errorf("cannot resolve course code from learning config")
	}
	summary := &LegacyCourseTagSummary{CourseCode: courseCode, Tagged: map[string]int64{}}
	for _, table := range legacyCourseTagTables {
		// Table name is from a fixed allow-list, not user input.
		res, err := r.db.ExecContext(ctx, fmt.Sprintf(
			`UPDATE %s SET course_code = ? WHERE course_code IS NULL`, table), courseCode)
		if err != nil {
			return nil, fmt.Errorf("tag %s with course %s: %w", table, courseCode, err)
		}
		if affected, err := res.RowsAffected(); err == nil {
			summary.Tagged[table] = affected
		}
	}
	return summary, nil
}
