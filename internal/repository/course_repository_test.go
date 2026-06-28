package repository

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

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

func TestGrammarBundleIDForCourse(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"es_ru", "es"},
		{"en_ru", "en"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := GrammarBundleIDForCourse(tc.code); got != tc.want {
			t.Fatalf("GrammarBundleIDForCourse(%q) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestCourseRepository_ListActiveCourseCodes(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewCourseRepository(conn, zap.NewNop())

	codes, err := repo.ListActiveCourseCodes(context.Background())
	if err != nil {
		t.Fatalf("ListActiveCourseCodes: %v", err)
	}
	if len(codes) < 2 {
		t.Fatalf("expected at least 2 active courses, got %+v", codes)
	}
}

func TestCourseRepository_BackfillUserCourses_Validation(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewCourseRepository(conn, zap.NewNop())

	if _, err := repo.BackfillUserCourses(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty course code")
	}
	if _, err := repo.BackfillUserCourses(context.Background(), "missing_course"); err == nil {
		t.Fatal("expected error for unknown course code")
	}
}

func TestCourseRepository_MapLegacyContent_Validation(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewCourseRepository(conn, zap.NewNop())

	if _, err := repo.MapLegacyContent(context.Background(), "", "es"); err == nil {
		t.Fatal("expected error for empty course code")
	}
	if _, err := repo.MapLegacyContent(context.Background(), "es_ru", ""); err == nil {
		t.Fatal("expected error for empty bundle id")
	}
	if _, err := repo.MapLegacyContent(context.Background(), "missing_course", "es"); err == nil {
		t.Fatal("expected error for unknown course code")
	}
}

func TestCourseRepository_MapLegacyContentForLearning(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	repo := NewCourseRepository(conn, logger)

	if _, err := conn.Exec(`
		INSERT INTO grammar_content_sections (bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash)
		VALUES ('es', 'es.section.learning', 'Grammar', 'A0', 1, '[]', '{}', 'sec-learning')`); err != nil {
		t.Fatalf("insert grammar section: %v", err)
	}

	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	summary, err := repo.MapLegacyContentForLearning(context.Background(), lc)
	if err != nil {
		t.Fatalf("MapLegacyContentForLearning: %v", err)
	}
	if summary.CourseCode != "es_ru" || summary.ModulesCreated < 1 {
		t.Fatalf("summary = %+v", summary)
	}

	second, err := repo.MapLegacyContentForLearning(context.Background(), lc)
	if err != nil {
		t.Fatalf("second MapLegacyContentForLearning: %v", err)
	}
	if second.ModulesCreated != 0 {
		t.Fatalf("expected idempotent mapping, got %+v", second)
	}
}

func TestCourseRepository_MapLegacyContent_GrammarTheoryBlocks(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	repo := NewCourseRepository(conn, logger)

	inserts := []string{
		`INSERT INTO grammar_content_sections (bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash)
		 VALUES ('es', 'es.section.tb', 'Grammar TB', 'A0', 1, '["es.chapter.tb"]', '{}', 'sec-tb')`,
		`INSERT INTO grammar_content_chapters (bundle_id, chapter_id, section_id, title, ui_language, target_language, level, sort_order, raw_json, source_hash)
		 VALUES ('es', 'es.chapter.tb', 'es.section.tb', 'Chapter TB', 'ru', 'es', 'A0', 1, '{}', 'ch-tb')`,
		`INSERT INTO grammar_training_content_questions (bundle_id, chapter_id, theory_block_id, concept_id, question_id, source_hash, raw_json)
		 VALUES ('es', 'es.chapter.tb', 'block-alpha', 'concept-alpha', 'q1', 'hash1', '{}'),
		        ('es', 'es.chapter.tb', 'block-alpha', 'concept-alpha', 'q2', 'hash2', '{}')`,
	}
	for _, q := range inserts {
		if _, err := conn.Exec(q); err != nil {
			t.Fatalf("insert fixture: %v\n%s", err, q)
		}
	}

	summary, err := repo.MapLegacyContent(context.Background(), "es_ru", "es")
	if err != nil {
		t.Fatalf("MapLegacyContent: %v", err)
	}
	if summary.ModulesCreated < 1 || summary.ItemsCreated < 2 {
		t.Fatalf("expected grammar section/chapter/theory block items, got %+v", summary)
	}

	var theoryCount int
	if err := conn.QueryRow(`
		SELECT COUNT(*)
		FROM learning_items li
		JOIN courses c ON c.id = li.course_id
		WHERE c.code = 'es_ru' AND li.source_kind = 'grammar_theory_block'
	`).Scan(&theoryCount); err != nil {
		t.Fatalf("count theory blocks: %v", err)
	}
	if theoryCount != 1 {
		t.Fatalf("theory block items = %d, want 1", theoryCount)
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
		`INSERT INTO word_sets (title, description, is_published, sort_order, course_code)
		 VALUES ('Core words', 'Core', 1, 1, 'es_ru')`,
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
	if summary.ModulesCreated < 1 {
		t.Fatalf("ModulesCreated = %d, want at least 1 (%+v)", summary.ModulesCreated, summary)
	}

	second, err := repo.MapLegacyContent(context.Background(), "es_ru", "es")
	if err != nil {
		t.Fatalf("second MapLegacyContent: %v", err)
	}
	if second.ModulesCreated != 0 || second.ItemsCreated != 0 {
		t.Fatalf("mapping should be idempotent, got %+v", second)
	}
}

func TestCourseRepository_MapLegacyContentMapsCustomUserWords(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(99401)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	if _, err := repo.BackfillUserCourses(context.Background(), "es_ru"); err != nil {
		t.Fatalf("BackfillUserCourses: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_content_sections (bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash)
		VALUES ('es', 'es.section.custom', 'Grammar', 'A0', 1, '[]', '{}', 'sec-custom')`); err != nil {
		t.Fatalf("insert grammar section: %v", err)
	}
	var wordCardID, trainingCardID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition, display_en) VALUES ('custom-libro', 'book', 'libro') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert custom word card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, 'libro', 0, 'книга', 'book') RETURNING id`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training card: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, 'es_ru', 'new', 2.5)`, user.ID, trainingCardID); err != nil {
		t.Fatalf("insert user card: %v", err)
	}

	if _, err := repo.MapLegacyContent(context.Background(), "es_ru", "es"); err != nil {
		t.Fatalf("MapLegacyContent: %v", err)
	}
	var itemCount int
	if err := conn.QueryRow(`
		SELECT COUNT(*)
		FROM learning_items li
		JOIN courses c ON c.id = li.course_id
		WHERE c.code = 'es_ru' AND li.source_kind = 'word_card' AND li.source_id = ?
	`, fmt.Sprintf("%d", wordCardID)).Scan(&itemCount); err != nil {
		t.Fatalf("count custom learning item: %v", err)
	}
	if itemCount != 1 {
		t.Fatalf("custom learning_items count = %d, want 1", itemCount)
	}

	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	wordRepo := NewLinglowWordSRSBackfillRepository(conn)
	summary, err := wordRepo.Backfill(context.Background(), lc, LinglowWordSRSBackfillOptions{})
	if err != nil {
		t.Fatalf("word srs audit: %v", err)
	}
	if summary.UnmappedTotal != 0 {
		t.Fatalf("UnmappedTotal = %d, want 0 after custom mapping", summary.UnmappedTotal)
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

func TestCourseRepository_GetCourseMapEmptyCollections(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	repo := NewCourseRepository(conn, logger)

	courseMap, err := repo.GetCourseMap(context.Background(), "es_ru", 0)
	if err != nil {
		t.Fatalf("GetCourseMap: %v", err)
	}
	if len(courseMap.Districts) == 0 {
		t.Fatalf("expected seeded districts")
	}
	for _, district := range courseMap.Districts {
		if district.Locations == nil {
			t.Fatalf("district %s locations is nil", district.Code)
		}
		for _, location := range district.Locations {
			if location.Modules == nil {
				t.Fatalf("location %s modules is nil", location.Code)
			}
			for _, module := range location.Modules {
				if module.Items == nil {
					t.Fatalf("module %s items is nil", module.Code)
				}
			}
		}
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

func TestCourseRepository_CourseMapForUserResolvesCurrentAndExplicit(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4344)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)

	if _, err := repo.SelectCurrentCourse(context.Background(), user.ID, "en_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}
	currentMap, err := repo.GetCourseMapForUser(context.Background(), user.ID, "es_ru", "")
	if err != nil {
		t.Fatalf("GetCourseMapForUser current: %v", err)
	}
	if currentMap.Course.Code != "en_ru" || currentMap.UserCourse == nil {
		t.Fatalf("current map = %+v", currentMap)
	}

	explicitMap, err := repo.GetCourseMapForUser(context.Background(), user.ID, "es_ru", "es_ru")
	if err != nil {
		t.Fatalf("GetCourseMapForUser explicit: %v", err)
	}
	if explicitMap.Course.Code != "es_ru" || explicitMap.UserCourse == nil {
		t.Fatalf("explicit map = %+v", explicitMap)
	}

	resolved, err := repo.ResolveCurrentCourseCode(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("ResolveCurrentCourseCode: %v", err)
	}
	if resolved != "en_ru" {
		t.Fatalf("explicit read changed current course to %q", resolved)
	}
}

func TestCourseRepository_DailyRouteForUser(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4345)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	current, err := repo.SelectCurrentCourse(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var dueItemID, newItemID int64
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'grammar:route-test', 'grammar', 'Route Grammar', 'grammar_category', 'route-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'grammar_theory_block', 'grammar_theory_block', 'route-due', 'Due Block', 'A0', 'published'
		FROM module
		RETURNING id
	`).Scan(&dueItemID); err != nil {
		t.Fatalf("insert due item: %v", err)
	}
	if err := conn.QueryRow(`
		WITH module AS (
			SELECT m.id, m.course_id, m.district_id, m.location_id
			FROM modules m
			JOIN courses c ON c.id = m.course_id
			WHERE c.code = 'es_ru' AND m.code = 'grammar:route-test'
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'reading_text', 'reading_text', 'route-new', 'New Text', 'A0', 'published'
		FROM module
		RETURNING id
	`).Scan(&newItemID); err != nil {
		t.Fatalf("insert new item: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at, last_review_at, reps)
		VALUES (?, ?, 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour', CURRENT_TIMESTAMP - INTERVAL '1 day', 3)
	`, current.UserCourse.ID, dueItemID); err != nil {
		t.Fatalf("insert srs item: %v", err)
	}

	route, err := repo.GetDailyRouteForUser(context.Background(), user.ID, "es_ru", "", 4)
	if err != nil {
		t.Fatalf("GetDailyRouteForUser: %v", err)
	}
	if route.Course.Code != "es_ru" || route.UserCourse.ID != current.UserCourse.ID {
		t.Fatalf("route course = %+v user_course=%+v", route.Course, route.UserCourse)
	}
	if route.Summary.DueReviewCount != 1 || route.Summary.NewItemCount < 1 {
		t.Fatalf("route summary = %+v", route.Summary)
	}
	if len(route.Review) != 1 || route.Review[0].LearningItemID != dueItemID || route.Review[0].SRSItemID == nil {
		t.Fatalf("route review = %+v", route.Review)
	}
	var foundNew bool
	for _, item := range route.NewItems {
		if item.LearningItemID == newItemID && item.State == "new" {
			foundNew = true
		}
	}
	if !foundNew {
		t.Fatalf("new item %d not found in %+v", newItemID, route.NewItems)
	}
}

func TestCourseRepository_ReviewQueueForUser(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4346)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	current, err := repo.SelectCurrentCourse(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var dueItemID, upcomingItemID int64
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'grammar:review-test', 'grammar', 'Review Grammar', 'grammar_category', 'review-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'grammar_theory_block', 'grammar_theory_block', 'review-due', 'Due Review Block', 'A0', 'published'
		FROM module
		RETURNING id
	`).Scan(&dueItemID); err != nil {
		t.Fatalf("insert due item: %v", err)
	}
	if err := conn.QueryRow(`
		WITH module AS (
			SELECT m.id, m.course_id, m.district_id, m.location_id
			FROM modules m
			JOIN courses c ON c.id = m.course_id
			WHERE c.code = 'es_ru' AND m.code = 'grammar:review-test'
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', 'review-upcoming', 'Upcoming Word', 'A0', 'published'
		FROM module
		RETURNING id
	`).Scan(&upcomingItemID); err != nil {
		t.Fatalf("insert upcoming item: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at, last_review_at, reps)
		VALUES
			(?, ?, 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour', CURRENT_TIMESTAMP - INTERVAL '1 day', 3),
			(?, ?, 'learning', CURRENT_TIMESTAMP + INTERVAL '1 hour', CURRENT_TIMESTAMP, 1)
	`, current.UserCourse.ID, dueItemID, current.UserCourse.ID, upcomingItemID); err != nil {
		t.Fatalf("insert srs items: %v", err)
	}

	queue, err := repo.GetReviewQueueForUser(context.Background(), user.ID, "es_ru", "", 10)
	if err != nil {
		t.Fatalf("GetReviewQueueForUser: %v", err)
	}
	if queue.Course.Code != "es_ru" || queue.UserCourse.ID != current.UserCourse.ID {
		t.Fatalf("queue course = %+v user_course=%+v", queue.Course, queue.UserCourse)
	}
	if queue.Summary.DueCount != 1 || queue.Summary.ReviewCount != 1 || queue.Summary.UpcomingCount != 1 {
		t.Fatalf("queue summary = %+v", queue.Summary)
	}
	if queue.Summary.ReadSource != "canonical" {
		t.Fatalf("queue read source = %q", queue.Summary.ReadSource)
	}
	if queue.Summary.ByType["grammar_theory_block"] != 1 {
		t.Fatalf("queue by type = %+v", queue.Summary.ByType)
	}
	if len(queue.Items) != 1 || queue.Items[0].LearningItemID != dueItemID || queue.Items[0].SRSItemID == nil {
		t.Fatalf("queue items = %+v", queue.Items)
	}
}

func TestCourseRepository_ReviewQueueForUserLegacySRSRead(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(43461)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	current, err := repo.SelectCurrentCourse(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var wordCardID, trainingCardID, wordItemID, grammarItemID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ('legacy-due', 'legacy due') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word card: %v", err)
	}
	if err := conn.QueryRow(`
		INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en)
		VALUES (?, 'legacy-due', 0, 'legacy due', 'legacy due')
		RETURNING id
	`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training card: %v", err)
	}
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'word_set:legacy-review-test', 'word_set', 'Legacy Words', 'word_set', 'legacy-review-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', CAST(? AS text), 'Legacy Due Word', 'A0', 'published'
		FROM module
		RETURNING id
	`, strconv.FormatInt(wordCardID, 10)).Scan(&wordItemID); err != nil {
		t.Fatalf("insert word item: %v", err)
	}
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'grammar:legacy-review-test', 'grammar', 'Legacy Grammar', 'grammar_category', 'legacy-review-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'grammar_theory_block', 'grammar_theory_block', 'legacy.chapter:block1', 'Legacy Grammar Block', 'A0', 'published'
		FROM module
		RETURNING id
	`).Scan(&grammarItemID); err != nil {
		t.Fatalf("insert grammar item: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO user_cards (user_id, training_card_id, direction, state, next_due_at, last_review_at, reps)
		VALUES (?, ?, 'es_ru', 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour', CURRENT_TIMESTAMP - INTERVAL '1 day', 2)
	`, user.ID, trainingCardID); err != nil {
		t.Fatalf("insert user card: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_theory_memory (user_id, language, course_id, chapter_id, theory_block_id, concept_id, state, next_review_at)
		VALUES (?, 'es', 'es', 'legacy.chapter', 'block1', 'legacy.concept', 'learning', CURRENT_TIMESTAMP - INTERVAL '2 hours')
	`, user.ID); err != nil {
		t.Fatalf("insert grammar memory: %v", err)
	}

	queue, err := repo.GetReviewQueueForUserWithSRSRead(context.Background(), user.ID, "es_ru", "", 10, false)
	if err != nil {
		t.Fatalf("GetReviewQueueForUserWithSRSRead legacy: %v", err)
	}
	if queue.UserCourse.ID != current.UserCourse.ID || queue.Summary.ReadSource != "legacy" {
		t.Fatalf("legacy queue course=%+v summary=%+v", queue.UserCourse, queue.Summary)
	}
	if queue.Summary.DueCount != 2 || queue.Summary.ReviewCount != 1 || queue.Summary.LearningCount != 1 {
		t.Fatalf("legacy queue summary = %+v", queue.Summary)
	}
	if queue.Summary.ByType["word"] != 1 || queue.Summary.ByType["grammar_theory_block"] != 1 {
		t.Fatalf("legacy queue by type = %+v", queue.Summary.ByType)
	}
	got := map[int64]bool{}
	for _, item := range queue.Items {
		got[item.LearningItemID] = true
		if item.SRSItemID != nil {
			t.Fatalf("legacy item should not expose canonical srs id: %+v", item)
		}
	}
	if !got[wordItemID] || !got[grammarItemID] {
		t.Fatalf("legacy queue items = %+v", queue.Items)
	}
}

func TestCourseRepository_RecordExerciseAttemptAndProgress(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4347)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	if _, err := repo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var itemID int64
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'grammar:attempt-test', 'grammar', 'Attempt Grammar', 'grammar_category', 'attempt-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'grammar_theory_block', 'grammar_theory_block', 'attempt-item', 'Attempt Block', 'A0', 'published'
		FROM module
		RETURNING id
	`).Scan(&itemID); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	correct := true
	score := 90
	quality := 4
	first, err := repo.RecordExerciseAttempt(context.Background(), ExerciseAttemptInput{
		UserID:          user.ID,
		DefaultCourse:   "es_ru",
		LearningItemID:  itemID,
		Mode:            "grammar",
		ClientAttemptID: "attempt-1",
		IsCorrect:       &correct,
		Score:           &score,
		Quality:         &quality,
		PromptJSON:      `{"kind":"test"}`,
		AnswerJSON:      `{"answer":"x"}`,
		ResultJSON:      `{"ok":true}`,
	})
	if err != nil {
		t.Fatalf("RecordExerciseAttempt first: %v", err)
	}
	second, err := repo.RecordExerciseAttempt(context.Background(), ExerciseAttemptInput{
		UserID:          user.ID,
		DefaultCourse:   "es_ru",
		LearningItemID:  itemID,
		Mode:            "grammar",
		ClientAttemptID: "attempt-1",
		IsCorrect:       &correct,
	})
	if err != nil {
		t.Fatalf("RecordExerciseAttempt duplicate: %v", err)
	}
	if second.ID != first.ID || !second.Duplicate {
		t.Fatalf("duplicate result = %+v first=%+v", second, first)
	}
	var attempts, events int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE client_attempt_id = 'attempt-1'`).Scan(&attempts); err != nil {
		t.Fatalf("count attempts: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM learning_events WHERE exercise_attempt_id = ?`, first.ID).Scan(&events); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if attempts != 1 || events != 1 {
		t.Fatalf("attempts=%d events=%d", attempts, events)
	}

	progress, err := repo.GetProgressForUser(context.Background(), user.ID, "es_ru", "")
	if err != nil {
		t.Fatalf("GetProgressForUser: %v", err)
	}
	if progress.Summary.AttemptedItems != 1 || progress.Summary.AttemptCount != 1 || progress.Summary.CorrectCount != 1 {
		t.Fatalf("progress summary = %+v", progress.Summary)
	}
	if len(progress.ByDistrict) == 0 {
		t.Fatalf("expected district progress rows")
	}
	var districtFound bool
	for _, district := range progress.ByDistrict {
		if district.AttemptedItems == 1 {
			districtFound = true
			if district.Foundation <= 0 || district.Confidence != 100 {
				t.Fatalf("district signals = %+v", district)
			}
		}
	}
	if !districtFound {
		t.Fatalf("expected attempted district progress, got %+v", progress.ByDistrict)
	}
	if len(progress.ByLocation) == 0 {
		t.Fatalf("expected location progress rows")
	}
	var locationFound bool
	for _, location := range progress.ByLocation {
		if location.AttemptedItems == 1 {
			locationFound = true
			if location.Foundation <= 0 || location.Confidence != 100 {
				t.Fatalf("location signals = %+v", location)
			}
		}
	}
	if !locationFound {
		t.Fatalf("expected attempted location progress, got %+v", progress.ByLocation)
	}
}

func TestCourseRepository_RecordExerciseAttemptUpdatesSRS(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4348)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	if _, err := repo.SelectCurrentCourse(context.Background(), user.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var itemID int64
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'grammar:srs-attempt-test', 'grammar', 'SRS Attempt Grammar', 'grammar_category', 'srs-attempt-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'grammar_theory_block', 'grammar_theory_block', 'srs-attempt-item', 'SRS Attempt Block', 'A0', 'published'
		FROM module
		RETURNING id
	`).Scan(&itemID); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	correct := true
	quality := 3
	result, err := repo.RecordExerciseAttempt(context.Background(), ExerciseAttemptInput{
		UserID:          user.ID,
		DefaultCourse:   "es_ru",
		LearningItemID:  itemID,
		Mode:            "grammar",
		ClientAttemptID: "srs-attempt-1",
		IsCorrect:       &correct,
		Quality:         &quality,
		UpdateSRS:       true,
	})
	if err != nil {
		t.Fatalf("RecordExerciseAttempt: %v", err)
	}
	if !result.SRSUpdated || result.SRSItemID == nil {
		t.Fatalf("expected SRS update, got %+v", result)
	}
	var state string
	var dueCount int
	if err := conn.QueryRow(`SELECT state FROM srs_items WHERE id = ?`, *result.SRSItemID).Scan(&state); err != nil {
		t.Fatalf("get srs state: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE id = ? AND srs_item_id = ?`, result.ID, *result.SRSItemID).Scan(&dueCount); err != nil {
		t.Fatalf("count linked attempt: %v", err)
	}
	if state != "learning" || dueCount != 1 {
		t.Fatalf("state=%s linked=%d", state, dueCount)
	}

	duplicate, err := repo.RecordExerciseAttempt(context.Background(), ExerciseAttemptInput{
		UserID:          user.ID,
		DefaultCourse:   "es_ru",
		LearningItemID:  itemID,
		Mode:            "grammar",
		ClientAttemptID: "srs-attempt-1",
		IsCorrect:       &correct,
		UpdateSRS:       true,
	})
	if err != nil {
		t.Fatalf("duplicate RecordExerciseAttempt: %v", err)
	}
	if !duplicate.Duplicate || duplicate.SRSUpdated {
		t.Fatalf("duplicate should not update SRS: %+v", duplicate)
	}
}

func TestCourseRepository_SRSShadowReportForUser(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4349)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	current, err := repo.SelectCurrentCourse(context.Background(), user.ID, "es_ru")
	if err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	var wordCardID, trainingCardID, learningItemID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ('shadow', 'shadow') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word card: %v", err)
	}
	if err := conn.QueryRow(`
		INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word)
		VALUES (?, 'shadow', 0, 'тень', 'shadow', 'noun', 'shadow')
		RETURNING id
	`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training card: %v", err)
	}
	dueAt := time.Now().Add(-time.Hour)
	if _, err := conn.Exec(`
		INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, reps, interval_days, next_due_at)
		VALUES (?, ?, 'ru_en', 'review', 2.5, 4, 7, ?)
	`, user.ID, trainingCardID, dueAt); err != nil {
		t.Fatalf("insert user card: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score) VALUES (?, ?, 80)`, user.ID, wordCardID); err != nil {
		t.Fatalf("insert mastering: %v", err)
	}
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'word:shadow-test', 'word_set', 'Shadow Words', 'word_set', 'shadow-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', CAST(? AS text), 'shadow', 'A0', 'published'
		FROM module
		RETURNING id
	`, fmt.Sprintf("%d", wordCardID)).Scan(&learningItemID); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at, reps, stats_json)
		VALUES (?, ?, 'review', ?, 3, '{"mastery_score":70}'::jsonb)
	`, current.UserCourse.ID, learningItemID, dueAt); err != nil {
		t.Fatalf("insert srs item: %v", err)
	}

	report, err := repo.GetSRSShadowReportForUser(context.Background(), user.ID, "es_ru", "")
	if err != nil {
		t.Fatalf("GetSRSShadowReportForUser: %v", err)
	}
	if report.Due.LegacyDueCount != 1 || report.Due.LinglowDueCount != 1 || report.Due.OverlapCount != 1 {
		t.Fatalf("due shadow = %+v", report.Due)
	}
	if report.ReviewQueue.LegacyDueCount != 1 || report.ReviewQueue.CanonicalDueCount != 1 || report.ReviewQueue.OverlapCount != 1 || !report.ReviewQueue.ReadyForCanonicalRead {
		t.Fatalf("review queue shadow = %+v", report.ReviewQueue)
	}
	if report.ReviewQueue.ByType["word"] != 1 {
		t.Fatalf("review queue by type = %+v", report.ReviewQueue.ByType)
	}
	if report.Mastery.ComparedCount != 1 || report.Mastery.AverageDifference != 10 {
		t.Fatalf("mastery shadow = %+v", report.Mastery)
	}
}

func TestCourseRepository_SRSShadowReportEmpty(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(4350)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	report, err := repo.GetSRSShadowReportForUser(context.Background(), user.ID, "es_ru", "")
	if err != nil {
		t.Fatalf("GetSRSShadowReportForUser empty: %v", err)
	}
	if report.Course.Code != "es_ru" || report.UserCourse.ID == 0 {
		t.Fatalf("empty shadow report = %+v", report)
	}
	if !report.ReviewQueue.ReadyForCanonicalRead {
		t.Fatalf("empty review queue shadow should be ready: %+v", report.ReviewQueue)
	}
}

func TestCourseRepository_SRSReadinessAggregate(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	readyUser, err := userRepo.GetOrCreateUser(4351)
	if err != nil {
		t.Fatalf("create ready user: %v", err)
	}
	notReadyUser, err := userRepo.GetOrCreateUser(4352)
	if err != nil {
		t.Fatalf("create not ready user: %v", err)
	}
	repo := NewCourseRepository(conn, logger)
	readyCourse, err := repo.SelectCurrentCourse(context.Background(), readyUser.ID, "es_ru")
	if err != nil {
		t.Fatalf("SelectCurrentCourse ready: %v", err)
	}
	if _, err := repo.SelectCurrentCourse(context.Background(), notReadyUser.ID, "es_ru"); err != nil {
		t.Fatalf("SelectCurrentCourse not ready: %v", err)
	}

	var wordCardID, trainingCardID, learningItemID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES ('readiness', 'readiness') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word card: %v", err)
	}
	if err := conn.QueryRow(`
		INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en)
		VALUES (?, 'readiness', 0, 'готовность', 'readiness')
		RETURNING id
	`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training card: %v", err)
	}
	if err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
			WHERE c.code = 'es_ru'
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, 'word:readiness-test', 'word_set', 'Readiness Words', 'word_set', 'readiness-test', 1, 'published'
			FROM target
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'word', 'word_card', CAST(? AS text), 'readiness', 'A0', 'published'
		FROM module
		RETURNING id
	`, strconv.FormatInt(wordCardID, 10)).Scan(&learningItemID); err != nil {
		t.Fatalf("insert learning item: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, next_due_at)
		VALUES
			(?, ?, 'es_ru', 'review', 2.5, CURRENT_TIMESTAMP - INTERVAL '1 hour'),
			(?, ?, 'es_ru', 'review', 2.5, CURRENT_TIMESTAMP - INTERVAL '1 hour')
	`, readyUser.ID, trainingCardID, notReadyUser.ID, trainingCardID); err != nil {
		t.Fatalf("insert user cards: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at)
		VALUES (?, ?, 'review', CURRENT_TIMESTAMP - INTERVAL '1 hour')
	`, readyCourse.UserCourse.ID, learningItemID); err != nil {
		t.Fatalf("insert canonical srs: %v", err)
	}

	report, err := repo.GetSRSReadinessAggregate(context.Background(), "es_ru", 10)
	if err != nil {
		t.Fatalf("GetSRSReadinessAggregate: %v", err)
	}
	if report.UserCoursesTotal != 2 || report.ReadyCount != 1 || report.NotReadyCount != 1 || report.ReadyForCanonicalRead {
		t.Fatalf("aggregate readiness = %+v", report)
	}
	if report.LegacyDueTotal != 2 || report.CanonicalDueTotal != 1 || report.LegacyOnlyTotal != 1 || report.CanonicalOnlyTotal != 0 || report.OverlapTotal != 1 {
		t.Fatalf("aggregate totals = %+v", report)
	}
	if len(report.NotReadyUsers) != 1 || report.NotReadyUsers[0].UserID != notReadyUser.ID {
		t.Fatalf("not ready users = %+v", report.NotReadyUsers)
	}
}
