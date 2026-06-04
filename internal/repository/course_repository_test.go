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

func TestCourseRepository_MapLegacyContent(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()

	inserts := []string{
		`INSERT INTO grammar_content_sections (bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash)
		 VALUES ('es', 'es.section.a0', 'Grammar A0', 'A0', 1, '["es.chapter.a0"]', '{}', 'sec-hash')`,
		`INSERT INTO grammar_content_chapters (bundle_id, chapter_id, section_id, title, ui_language, target_language, level, sort_order, raw_json, source_hash)
		 VALUES ('es', 'es.chapter.a0', 'es.section.a0', 'Chapter A0', 'ru', 'es', 'A0', 1, '{}', 'ch-hash')`,
		`INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids)
		 VALUES ('read.a0', 'Reading A0', 'A0', 1, '["read.text.a0"]')`,
		`INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
		 VALUES ('read.text.a0', 'read.a0', 'Text A0', 'A0', 'es', 'Hola mundo')`,
		`INSERT INTO speaking_categories (category_id, title, level, sort_order, task_ids)
		 VALUES ('speak.a0', 'Speaking A0', 'A0', 1, '["speak.task.a0"]')`,
		`INSERT INTO speaking_tasks (task_id, category_id, title, level, task_type, target_language, task_json)
		 VALUES ('speak.task.a0', 'speak.a0', 'Task A0', 'A0', 'answer', 'es', '{}')`,
		`INSERT INTO word_cards (word, definition, display_en)
		 VALUES ('hola', 'hello', 'hola')`,
		`INSERT INTO word_sets (title, description, is_published, sort_order)
		 VALUES ('Core words', 'Core', 1, 1)`,
		`INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
		 SELECT ws.id, wc.id, 1 FROM word_sets ws CROSS JOIN word_cards wc WHERE ws.title = 'Core words' AND wc.word = 'hola'`,
	}
	for _, q := range inserts {
		if _, err := conn.Exec(q); err != nil {
			t.Fatalf("insert fixture: %v\n%s", err, q)
		}
	}

	repo := NewCourseRepository(conn, logger)
	summary, err := repo.MapLegacyContent(context.Background(), "es_ru", "es")
	if err != nil {
		t.Fatalf("MapLegacyContent: %v", err)
	}
	if summary.ModulesCreated != 4 {
		t.Fatalf("ModulesCreated = %d, want 4 (%+v)", summary.ModulesCreated, summary)
	}
	if summary.ItemsCreated != 4 {
		t.Fatalf("ItemsCreated = %d, want 4 (%+v)", summary.ItemsCreated, summary)
	}

	second, err := repo.MapLegacyContent(context.Background(), "es_ru", "es")
	if err != nil {
		t.Fatalf("second MapLegacyContent: %v", err)
	}
	if second.ModulesCreated != 0 || second.ItemsCreated != 0 {
		t.Fatalf("mapping should be idempotent, got %+v", second)
	}
}

func TestCourseRepository_GetCourseMap(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4242)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	repo := NewCourseRepository(conn, logger)
	if _, err := repo.BackfillUserCourses(context.Background(), "es_ru"); err != nil {
		t.Fatalf("BackfillUserCourses: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
		SELECT c.id, d.id, l.id, 'grammar_section:test', 'grammar', 'Grammar Test', 'grammar_section', 'test.section', 1, 'published'
		FROM courses c
		JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
		JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
		WHERE c.code = 'es_ru'
	`); err != nil {
		t.Fatalf("insert module: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'grammar_chapter', 'grammar_chapter', 'test.chapter', 'Chapter Test', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'grammar_section:test'
	`); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}

	courseMap, err := repo.GetCourseMap(context.Background(), "es_ru", user.ID)
	if err != nil {
		t.Fatalf("GetCourseMap: %v", err)
	}
	if courseMap.Course.Code != "es_ru" {
		t.Fatalf("course code = %q, want es_ru", courseMap.Course.Code)
	}
	if courseMap.UserCourse == nil {
		t.Fatalf("expected user_course in response")
	}
	if courseMap.Totals.Districts != 6 || courseMap.Totals.Locations != 36 || courseMap.Totals.Modules != 1 || courseMap.Totals.Items != 1 {
		t.Fatalf("unexpected totals: %+v", courseMap.Totals)
	}
	if courseMap.Totals.ByType["grammar_chapter"] != 1 {
		t.Fatalf("grammar_chapter total = %d, want 1", courseMap.Totals.ByType["grammar_chapter"])
	}
	if len(courseMap.Districts) == 0 || len(courseMap.Districts[0].Locations) == 0 {
		t.Fatalf("expected district locations")
	}
}

func TestCourseRepository_CurrentCourseSelection(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4343)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)

	current, err := repo.GetCurrentCourse(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("GetCurrentCourse default: %v", err)
	}
	if current.Course.Code != "es_ru" || current.UserCourse.ID == 0 {
		t.Fatalf("default current = %+v", current)
	}

	selected, err := repo.SelectCurrentCourse(context.Background(), user.ID, "en_ru")
	if err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}
	if selected.Course.Code != "en_ru" || selected.UserCourse.ID == 0 {
		t.Fatalf("selected current = %+v", selected)
	}

	resolved, err := repo.ResolveCurrentCourseCode(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("ResolveCurrentCourseCode: %v", err)
	}
	if resolved != "en_ru" {
		t.Fatalf("resolved = %q, want en_ru", resolved)
	}

	courses, err := repo.ListCoursesForUser(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("ListCoursesForUser: %v", err)
	}
	var foundCurrent bool
	for _, course := range courses {
		if course.Code == "en_ru" && course.IsCurrent && course.UserCourseID != nil {
			foundCurrent = true
		}
	}
	if !foundCurrent {
		t.Fatalf("current selected course not found in %+v", courses)
	}
}
