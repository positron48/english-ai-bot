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

// ContentMappingSummary describes one idempotent legacy-content-to-Linglow mapping run.
type ContentMappingSummary struct {
	CourseCode     string
	CourseID       int64
	ModulesCreated int64
	ItemsCreated   int64
	ModulesTotal   int64
	ItemsTotal     int64
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

// MapLegacyContentForLearning idempotently maps existing DB-first/bundle-derived content into Linglow v2 modules and learning_items.
func (r *CourseRepository) MapLegacyContentForLearning(ctx context.Context, lc config.LearningConfig) (*ContentMappingSummary, error) {
	return r.MapLegacyContent(ctx, CourseCodeForLearning(lc), strings.TrimSpace(strings.ToLower(lc.GrammarBundleID)))
}

// MapLegacyContent idempotently maps existing legacy content tables into Linglow v2 content tables.
func (r *CourseRepository) MapLegacyContent(ctx context.Context, courseCode, bundleID string) (*ContentMappingSummary, error) {
	courseCode = strings.TrimSpace(strings.ToLower(courseCode))
	bundleID = strings.TrimSpace(strings.ToLower(bundleID))
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}
	if bundleID == "" {
		return nil, fmt.Errorf("bundle id is empty")
	}

	var courseID int64
	if err := r.db.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = ?`, courseCode).Scan(&courseID); err != nil {
		return nil, fmt.Errorf("get course %q: %w", courseCode, err)
	}
	var modulesBefore, itemsBefore int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM modules WHERE course_id = ?`, courseID).Scan(&modulesBefore); err != nil {
		return nil, fmt.Errorf("count modules before: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM learning_items WHERE course_id = ?`, courseID).Scan(&itemsBefore); err != nil {
		return nil, fmt.Errorf("count learning items before: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin content mapping: %w", err)
	}
	defer tx.Rollback()

	statements := []struct {
		query      string
		usesBundle bool
	}{
		{mapGrammarSectionModulesSQL, true},
		{mapGrammarChapterItemsSQL, true},
		{mapReadingCategoryModulesSQL, false},
		{mapReadingTextItemsSQL, false},
		{mapSpeakingCategoryModulesSQL, false},
		{mapSpeakingTaskItemsSQL, false},
		{mapWordSetModulesSQL, false},
		{mapWordCardItemsSQL, false},
	}
	for _, stmt := range statements {
		args := []interface{}{courseID}
		if stmt.usesBundle {
			args = []interface{}{bundleID, courseID}
		}
		if _, err := tx.ExecContext(ctx, stmt.query, args...); err != nil {
			return nil, fmt.Errorf("map legacy content: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit content mapping: %w", err)
	}

	var modulesAfter, itemsAfter int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM modules WHERE course_id = ?`, courseID).Scan(&modulesAfter); err != nil {
		return nil, fmt.Errorf("count modules after: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM learning_items WHERE course_id = ?`, courseID).Scan(&itemsAfter); err != nil {
		return nil, fmt.Errorf("count learning items after: %w", err)
	}
	return &ContentMappingSummary{
		CourseCode:     courseCode,
		CourseID:       courseID,
		ModulesCreated: modulesAfter - modulesBefore,
		ItemsCreated:   itemsAfter - itemsBefore,
		ModulesTotal:   modulesAfter,
		ItemsTotal:     itemsAfter,
	}, nil
}

const levelDistrictJoinSQL = `
JOIN districts d ON d.course_id = c.id
    AND d.level_code = CASE
        WHEN upper(coalesce(src.level, '')) IN ('A0', 'A1', 'A2', 'B1', 'B2', 'C1') THEN upper(src.level)
        ELSE 'A0'
    END`

var mapGrammarSectionModulesSQL = `
WITH src AS (
    SELECT section_id AS source_id, title, level, sort_order
    FROM grammar_content_sections
    WHERE bundle_id = ?
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
)
INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status, updated_at)
SELECT course_id, district_id, location_id, 'grammar_section:' || source_id, 'grammar', title, 'grammar_section', source_id, sort_order, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, code) DO UPDATE SET
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    source_kind = excluded.source_kind,
    source_id = excluded.source_id,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapGrammarChapterItemsSQL = `
WITH src AS (
    SELECT chapter_id AS source_id, section_id, title, level, source_hash, sort_order
    FROM grammar_content_chapters
    WHERE bundle_id = ?
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'grammar_section' AND m.source_id = src.section_id
)
INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, content_hash, status, updated_at)
SELECT course_id, module_id, district_id, location_id, 'grammar_chapter', 'grammar_chapter', source_id, title, level, source_hash, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapReadingCategoryModulesSQL = `
WITH src AS (
    SELECT category_id AS source_id, title, level, sort_order
    FROM reading_categories
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'reading'
)
INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status, updated_at)
SELECT course_id, district_id, location_id, 'reading_category:' || source_id, 'reading', title, 'reading_category', source_id, sort_order, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, code) DO UPDATE SET
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    source_kind = excluded.source_kind,
    source_id = excluded.source_id,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapReadingTextItemsSQL = `
WITH src AS (
    SELECT text_id AS source_id, category_id, title, level, md5(reading_passage) AS source_hash
    FROM reading_texts
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'reading'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'reading_category' AND m.source_id = src.category_id
)
INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, content_hash, status, updated_at)
SELECT course_id, module_id, district_id, location_id, 'reading_text', 'reading_text', source_id, title, level, source_hash, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapSpeakingCategoryModulesSQL = `
WITH src AS (
    SELECT category_id AS source_id, title, level, sort_order
    FROM speaking_categories
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'conversation'
)
INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status, updated_at)
SELECT course_id, district_id, location_id, 'speaking_category:' || source_id, 'speaking', title, 'speaking_category', source_id, sort_order, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, code) DO UPDATE SET
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    source_kind = excluded.source_kind,
    source_id = excluded.source_id,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapSpeakingTaskItemsSQL = `
WITH src AS (
    SELECT task_id AS source_id, category_id, title, level, md5(task_json) AS source_hash
    FROM speaking_tasks
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'conversation'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'speaking_category' AND m.source_id = src.category_id
)
INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, content_hash, status, updated_at)
SELECT course_id, module_id, district_id, location_id, 'speaking_task', 'speaking_task', source_id, title, level, source_hash, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapWordSetModulesSQL = `
WITH src AS (
    SELECT id::text AS source_id, title, sort_order, 'A0' AS level
    FROM word_sets
    WHERE is_published = 1
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
)
INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status, updated_at)
SELECT course_id, district_id, location_id, 'word_set:' || source_id, 'word_set', title, 'word_set', source_id, sort_order, 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, code) DO UPDATE SET
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    source_kind = excluded.source_kind,
    source_id = excluded.source_id,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`

var mapWordCardItemsSQL = `
WITH src AS (
    SELECT DISTINCT ON (wc.id)
        wc.id::text AS source_id,
        wsi.word_set_id::text AS module_source_id,
        coalesce(nullif(wc.display_en, ''), wc.word) AS title,
        coalesce(wc.updated_at::text, wc.created_at::text, '') AS source_hash,
        'A0' AS level
    FROM word_cards wc
    JOIN word_set_items wsi ON wsi.word_card_id = wc.id
    JOIN word_sets ws ON ws.id = wsi.word_set_id AND ws.is_published = 1
    ORDER BY wc.id, wsi.sort_order
), target AS (
    SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id, m.id AS module_id, src.*
    FROM src
    JOIN courses c ON c.id = ?
    ` + levelDistrictJoinSQL + `
    JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
    LEFT JOIN modules m ON m.course_id = c.id AND m.source_kind = 'word_set' AND m.source_id = src.module_source_id
)
INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, content_hash, status, updated_at)
SELECT course_id, module_id, district_id, location_id, 'word', 'word_card', source_id, title, level, md5(source_hash), 'published', CURRENT_TIMESTAMP
FROM target
ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
    module_id = excluded.module_id,
    district_id = excluded.district_id,
    location_id = excluded.location_id,
    title = excluded.title,
    cefr_level = excluded.cefr_level,
    content_hash = excluded.content_hash,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP`
