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
