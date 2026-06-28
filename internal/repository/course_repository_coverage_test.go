package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

const coverageUserBase int64 = 900001

func coverageRepo(t *testing.T) (*CourseRepository, *sql.DB) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	return NewCourseRepository(conn, zap.NewNop()), conn
}

func coverageUser(t *testing.T, conn *sql.DB, telegramID int64) int64 {
	t.Helper()
	userRepo := NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(telegramID)
	if err != nil {
		t.Fatalf("create user %d: %v", telegramID, err)
	}
	return user.ID
}

func coverageInsertGrammarItem(t *testing.T, conn *sql.DB, courseCode, moduleCode, sourceID string) int64 {
	t.Helper()
	var itemID int64
	err := conn.QueryRow(`
		WITH target AS (
			SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
			FROM courses c
			JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
			JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
			WHERE c.code = ?
			LIMIT 1
		), module AS (
			INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
			SELECT course_id, district_id, location_id, ?, 'grammar', 'Cov Grammar', 'grammar_category', ?, 1, 'published'
			FROM target
			ON CONFLICT (course_id, code) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
			RETURNING id, course_id, district_id, location_id
		)
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT course_id, id, district_id, location_id, 'grammar_theory_block', 'grammar_theory_block', ?, 'Cov Block', 'A0', 'published'
		FROM module
		ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, courseCode, moduleCode, moduleCode, sourceID).Scan(&itemID)
	if err != nil {
		t.Fatalf("insert grammar item: %v", err)
	}
	return itemID
}

func TestCourseRepositoryCoverage_CourseCodeForLearning(t *testing.T) {
	tests := []struct {
		name string
		lc   config.LearningConfig
		want string
	}{
		{"es_ru", config.LearningConfig{NativeLang: "ru", TargetLang: "es"}, "es_ru"},
		{"empty target", config.LearningConfig{NativeLang: "ru"}, ""},
		{"empty native", config.LearningConfig{TargetLang: "es"}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CourseCodeForLearning(tc.lc); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestCourseRepositoryCoverage_FillDailyRouteToday(t *testing.T) {
	repo, conn := coverageRepo(t)
	ctx := context.Background()
	userID := coverageUser(t, conn, coverageUserBase)
	current, err := repo.SelectCurrentCourse(ctx, userID, "es_ru")
	if err != nil {
		t.Fatalf("SelectCurrentCourse: %v", err)
	}

	repo.FillDailyRouteToday(ctx, nil, userID)

	day := time.Now().Format("2006-01-02")
	if _, err := conn.Exec(`
		INSERT INTO mode_daily_stats (user_course_id, local_date, mode, attempt_count, correct_count)
		VALUES (?, CAST(? AS date), 'word_training', 3, 2)
		ON CONFLICT (user_course_id, local_date, mode) DO UPDATE SET attempt_count = 3
	`, current.UserCourse.ID, day); err != nil {
		t.Fatalf("mode_daily_stats: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids)
		VALUES ('cov.read', 'Cov Read', 'A0', 1, '["cov.text.unread"]')
		ON CONFLICT (category_id) DO NOTHING
	`); err != nil {
		t.Fatalf("reading category: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
		VALUES ('cov.text.unread', 'cov.read', 'Unread', 'A0', 'es', 'Hola')
		ON CONFLICT (text_id) DO NOTHING
	`); err != nil {
		t.Fatalf("reading text: %v", err)
	}
	eventItemID := coverageInsertGrammarItem(t, conn, "es_ru", "grammar:cov-today", "cov.today.block")
	if _, err := conn.Exec(`
		INSERT INTO learning_events (user_course_id, learning_item_id, event_type, event_time, mode, source_table, source_pk, event_json)
		VALUES
			(?, ?, 'reading_text_completed', CURRENT_TIMESTAMP, 'reading', 'test', 'r1', '{}'),
			(?, ?, 'chat_message_sent', CURRENT_TIMESTAMP, 'chat', 'test', 'c1', '{}')
	`, current.UserCourse.ID, eventItemID, current.UserCourse.ID, eventItemID); err != nil {
		t.Fatalf("events: %v", err)
	}

	route := &DailyRoute{
		Course:     CourseMapCourse{TargetLanguage: "es"},
		UserCourse: current.UserCourse,
		Summary:    DailyRouteSummary{DueReviewCount: 7},
	}
	repo.FillDailyRouteToday(ctx, route, userID)
	if route.Today == nil || route.Today.WordsDone != 3 || route.Today.ReadingSuggestion == nil {
		t.Fatalf("today = %+v", route.Today)
	}

	route2 := &DailyRoute{Course: CourseMapCourse{TargetLanguage: "xx"}, UserCourse: current.UserCourse}
	repo.FillDailyRouteToday(ctx, route2, userID)
	if route2.Today != nil && route2.Today.ReadingSuggestion != nil {
		t.Fatalf("unexpected suggestion for xx language")
	}
}

func TestCourseRepositoryCoverage_SelectEnsureResolve(t *testing.T) {
	repo, conn := coverageRepo(t)
	ctx := context.Background()

	if _, err := repo.SelectCurrentCourse(ctx, 0, "es_ru"); err == nil {
		t.Fatal("user 0")
	}
	if _, err := repo.SelectCurrentCourse(ctx, coverageUserBase+1, ""); err == nil {
		t.Fatal("empty code")
	}
	if _, err := repo.EnsureUserCourse(ctx, 0, "es_ru"); err == nil {
		t.Fatal("ensure user 0")
	}
	if _, err := repo.EnsureUserCourse(ctx, coverageUserBase+1, "ghost"); err == nil {
		t.Fatal("ensure ghost")
	}

	userID := coverageUser(t, conn, coverageUserBase+2)
	if _, err := repo.SelectCurrentCourse(ctx, userID, "en_ru"); err != nil {
		t.Fatalf("select en_ru: %v", err)
	}
	var courseID int64
	if err := conn.QueryRow(`SELECT id FROM courses WHERE code = 'en_ru'`).Scan(&courseID); err != nil {
		t.Fatalf("course id: %v", err)
	}
	if _, err := conn.Exec(`UPDATE user_courses SET status = 'archived' WHERE user_id = ? AND course_id = ?`, userID, courseID); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if _, err := repo.SelectCurrentCourse(ctx, userID, "en_ru"); err != nil {
		t.Fatalf("reactivate: %v", err)
	}

	onlyEN := coverageUser(t, conn, coverageUserBase+10)
	if _, err := repo.SelectCurrentCourse(ctx, onlyEN, "en_ru"); err != nil {
		t.Fatalf("only en: %v", err)
	}
	code, err := repo.ResolveCurrentCourseCode(ctx, onlyEN, "es_ru")
	if err != nil || code != "en_ru" {
		t.Fatalf("resolve enrolled = %q %v", code, err)
	}

	newUser := coverageUser(t, conn, coverageUserBase+11)
	code, err = repo.ResolveCurrentCourseCode(ctx, newUser, "es_ru")
	if err != nil || code != "es_ru" {
		t.Fatalf("new user = %q %v", code, err)
	}

	if _, err := conn.Exec(`UPDATE users SET settings_json = '{"current_course_code":"en_ru"}' WHERE id = ?`, userID); err != nil {
		t.Fatalf("settings: %v", err)
	}
	code, err = repo.ResolveCurrentCourseCode(ctx, userID, "es_ru")
	if err != nil || code != "en_ru" {
		t.Fatalf("settings resolve = %q", code)
	}

	if _, err := repo.ResolveRequestedCourseCode(ctx, userID, "es_ru", "ghost"); err == nil {
		t.Fatal("ghost explicit")
	}

	current, err := repo.GetCurrentCourse(ctx, userID, "es_ru")
	if err != nil || current.Course.Code != "en_ru" {
		t.Fatalf("GetCurrentCourse = %+v %v", current, err)
	}

	courses, err := repo.ListCoursesForUser(ctx, userID, "es_ru")
	if err != nil || len(courses) < 2 {
		t.Fatalf("ListCoursesForUser = %+v %v", courses, err)
	}
}

func TestCourseRepositoryCoverage_BackfillAndMap(t *testing.T) {
	repo, conn := coverageRepo(t)
	ctx := context.Background()
	userRepo := NewUserRepository(conn, zap.NewNop())
	if _, err := userRepo.GetOrCreateUser(coverageUserBase + 30); err != nil {
		t.Fatalf("user: %v", err)
	}
	summary, err := repo.BackfillUserCourses(ctx, "es_ru")
	if err != nil || summary.Created < 1 {
		t.Fatalf("BackfillUserCourses = %+v %v", summary, err)
	}
	fl, err := repo.BackfillUserCoursesForLearning(ctx, config.LearningConfig{NativeLang: "ru", TargetLang: "es"})
	if err != nil || fl.CourseCode != "es_ru" {
		t.Fatalf("BackfillUserCoursesForLearning = %+v %v", fl, err)
	}

	if got := GrammarBundleIDForCourse("es_ru"); got != "es" {
		t.Fatalf("GrammarBundleIDForCourse = %q", got)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_content_sections (bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash)
		VALUES ('es', 'es.section.cov', 'Cov', 'A0', 1, '[]', '{}', 'sec-cov')
	`); err != nil {
		t.Fatalf("section: %v", err)
	}
	if _, err := repo.MapLegacyContentForLearning(ctx, config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}); err != nil {
		t.Fatalf("MapLegacyContentForLearning: %v", err)
	}
}

func TestCourseRepositoryCoverage_CourseMapAndRoutes(t *testing.T) {
	repo, conn := coverageRepo(t)
	ctx := context.Background()
	userID := coverageUser(t, conn, coverageUserBase+40)

	if _, err := repo.GetCourseMap(ctx, "", userID); err == nil {
		t.Fatal("empty course code")
	}
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es"}
	if _, err := repo.GetCourseMapForLearning(ctx, lc, userID); err != nil {
		t.Fatalf("GetCourseMapForLearning: %v", err)
	}
	if _, err := repo.GetCourseMapForUser(ctx, 0, "es_ru", ""); err == nil {
		t.Fatal("user 0 map")
	}

	current, err := repo.SelectCurrentCourse(ctx, userID, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	itemID := coverageInsertGrammarItem(t, conn, "es_ru", "grammar:cov-route", "cov.route.block")
	if _, err := conn.Exec(`
		INSERT INTO srs_items (user_course_id, learning_item_id, state, due_at, reps)
		VALUES (?, ?, 'relearning', CURRENT_TIMESTAMP - INTERVAL '1 hour', 1)
	`, current.UserCourse.ID, itemID); err != nil {
		t.Fatalf("srs: %v", err)
	}

	if _, err := repo.GetDailyRouteForUser(ctx, 0, "es_ru", "", 5); err == nil {
		t.Fatal("daily route user 0")
	}
	route, err := repo.GetDailyRouteForUserWithSRSRead(ctx, userID, "es_ru", "", 0, true)
	if err != nil || len(route.Review) == 0 {
		t.Fatalf("canonical route = %+v %v", route, err)
	}
	legacy, err := repo.GetDailyRouteForUserWithSRSRead(ctx, userID, "es_ru", "", 100, false)
	if err != nil || legacy.Summary.ReadSource != "legacy" {
		t.Fatalf("legacy route = %+v %v", legacy, err)
	}
}

func TestCourseRepositoryCoverage_RecordExerciseAttemptAndSRSHelpers(t *testing.T) {
	repo, conn := coverageRepo(t)
	ctx := context.Background()
	userID := coverageUser(t, conn, coverageUserBase+60)
	current, err := repo.SelectCurrentCourse(ctx, userID, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	itemID := coverageInsertGrammarItem(t, conn, "es_ru", "grammar:cov-at", "cov.at.block")

	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{UserID: 0, Mode: "grammar"}); err == nil {
		t.Fatal("user 0")
	}
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID: userID, DefaultCourse: "es_ru", LearningItemID: itemID + 99999, Mode: "grammar",
	}); err == nil {
		t.Fatal("foreign item")
	}

	correct, wrong := true, false
	q1, q2, q3 := 1, 2, 5
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID: userID, DefaultCourse: "es_ru", LearningItemID: itemID, Mode: "grammar",
		ClientAttemptID: "cov-wrong", IsCorrect: &wrong, UpdateSRS: true,
	}); err != nil {
		t.Fatalf("wrong srs: %v", err)
	}
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID: userID, DefaultCourse: "es_ru", LearningItemID: itemID, Mode: "grammar",
		ClientAttemptID: "cov-q1", IsCorrect: &correct, Quality: &q1, UpdateSRS: true,
	}); err != nil {
		t.Fatalf("q1 srs: %v", err)
	}

	reviewItem := coverageInsertGrammarItem(t, conn, "es_ru", "grammar:cov-review", "cov.review.block")
	var srsID int64
	if err := conn.QueryRow(`
		INSERT INTO srs_items (user_course_id, learning_item_id, state, difficulty, reps, stats_json, due_at)
		VALUES (?, ?, 'review', 2.5, 2, '{"interval_days": 6}'::jsonb, CURRENT_TIMESTAMP)
		RETURNING id
	`, current.UserCourse.ID, reviewItem).Scan(&srsID); err != nil {
		t.Fatalf("review srs: %v", err)
	}
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID: userID, DefaultCourse: "es_ru", LearningItemID: reviewItem, SRSItemID: srsID,
		Mode: "grammar", ClientAttemptID: "cov-review", IsCorrect: &correct, Quality: &q3, UpdateSRS: true,
		PromptJSON: `{"q":1}`, AnswerJSON: `{}`, ResultJSON: `{}`,
	}); err != nil {
		t.Fatalf("review attempt: %v", err)
	}

	srsOnlyItem := coverageInsertGrammarItem(t, conn, "es_ru", "grammar:cov-srsonly", "cov.srsonly.block")
	var srsOnlyID int64
	if err := conn.QueryRow(`
		INSERT INTO srs_items (user_course_id, learning_item_id, state, difficulty, reps, stats_json)
		VALUES (?, ?, 'new', 0, 0, '{}'::jsonb) RETURNING id
	`, current.UserCourse.ID, srsOnlyItem).Scan(&srsOnlyID); err != nil {
		t.Fatalf("srs only: %v", err)
	}
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID: userID, DefaultCourse: "es_ru", SRSItemID: srsOnlyID,
		Mode: "grammar", ClientAttemptID: "cov-srs-only", IsCorrect: &correct, UpdateSRS: true,
	}); err != nil {
		t.Fatalf("srs-only: %v", err)
	}

	dup, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID: userID, DefaultCourse: "es_ru", LearningItemID: reviewItem, SRSItemID: srsID,
		Mode: "grammar", ClientAttemptID: "cov-review", IsCorrect: &correct,
	})
	if err != nil || !dup.Duplicate || dup.LearningItemID == nil || dup.SRSItemID == nil {
		t.Fatalf("dup = %+v %v", dup, err)
	}

	if sm2Quality(1) != 3 || maxInt(2, 5) != 5 || intFromStats(map[string]interface{}{"k": float64(3)}, "k", 0) != 3 {
		t.Fatal("helpers")
	}
	now := time.Now().UTC()
	if applyLinglowSRS(linglowSRSState{State: "review", EF: 2.5, Reps: 3, IntervalDays: 4, Stats: map[string]interface{}{}}, true, &q2, now).IntervalDays < 4 {
		t.Fatal("applyLinglowSRS review")
	}
	for itemType, want := range map[string]string{"word": "word_training", "reading_text": "reading", "x": "x"} {
		if dailyRouteMode(itemType) != want {
			t.Fatalf("mode %q", itemType)
		}
	}
}

func TestCourseRepositoryCoverage_ShadowProgressAggregate(t *testing.T) {
	repo, conn := coverageRepo(t)
	ctx := context.Background()

	if _, err := repo.GetSRSReadinessAggregate(ctx, "", 10); err != ErrCourseNotFound {
		t.Fatalf("empty aggregate err = %v", err)
	}
	if _, err := repo.GetSRSShadowReportForUser(ctx, 0, "es_ru", ""); err == nil {
		t.Fatal("shadow user 0")
	}

	userID := coverageUser(t, conn, coverageUserBase+80)
	current, err := repo.SelectCurrentCourse(ctx, userID, "es_ru")
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	for i, st := range []string{"mastered", "review", "learning", "relearning"} {
		var wcID, liID int64
		word := fmt.Sprintf("cov-m-%d", i)
		if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES (?, 'm') RETURNING id`, word).Scan(&wcID); err != nil {
			t.Fatalf("word: %v", err)
		}
		if err := conn.QueryRow(`
			WITH target AS (
				SELECT c.id AS course_id, d.id AS district_id, l.id AS location_id
				FROM courses c JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
				JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
				WHERE c.code = 'es_ru' LIMIT 1
			), module AS (
				INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
				SELECT course_id, district_id, location_id, ?, 'word_set', 'M', 'word_set', ?, 1, 'published'
				FROM target ON CONFLICT (course_id, code) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
				RETURNING id, course_id, district_id, location_id
			)
			INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, status)
			SELECT course_id, id, district_id, location_id, 'word', 'word_card', CAST(? AS text), ?, 'published'
			FROM module ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
			RETURNING id
		`, "word:cov-m-"+st, "word:cov-m-"+st, strconv.FormatInt(wcID, 10), word).Scan(&liID); err != nil {
			t.Fatalf("item: %v", err)
		}
		if _, err := conn.Exec(`
			INSERT INTO srs_items (user_course_id, learning_item_id, state, reps, stats_json, due_at)
			VALUES (?, ?, ?, 1, '{}'::jsonb, CURRENT_TIMESTAMP)
		`, current.UserCourse.ID, liID, st); err != nil {
			t.Fatalf("srs %s: %v", st, err)
		}
	}
	report, err := repo.GetSRSShadowReportForUser(ctx, userID, "es_ru", "")
	if err != nil || report.Mastery.ComparedCount < 1 {
		t.Fatalf("shadow = %+v %v", report, err)
	}

	masteredItem := coverageInsertGrammarItem(t, conn, "es_ru", "grammar:cov-prog", "cov.prog.block")
	if _, err := conn.Exec(`INSERT INTO srs_items (user_course_id, learning_item_id, state) VALUES (?, ?, 'mastered')`, current.UserCourse.ID, masteredItem); err != nil {
		t.Fatalf("mastered: %v", err)
	}
	ok := true
	if _, err := repo.RecordExerciseAttempt(ctx, ExerciseAttemptInput{
		UserID: userID, DefaultCourse: "es_ru", ExplicitCourse: "es_ru", LearningItemID: masteredItem,
		Mode: "grammar", ClientAttemptID: "cov-prog", IsCorrect: &ok,
	}); err != nil {
		t.Fatalf("attempt: %v", err)
	}
	progress, err := repo.GetProgressForUser(ctx, userID, "en_ru", "es_ru")
	if err != nil || progress.Summary.TotalItems < 1 {
		t.Fatalf("progress = %+v %v", progress, err)
	}

	if _, err := repo.GetSRSReadinessAggregate(ctx, "es_ru", 500); err != nil {
		t.Fatalf("aggregate: %v", err)
	}
}

func TestCourseRepositoryCoverage_SettingsMergeOnSelect(t *testing.T) {
	repo, conn := coverageRepo(t)
	ctx := context.Background()
	userID := coverageUser(t, conn, coverageUserBase+101)
	if _, err := conn.Exec(`UPDATE users SET settings_json = '{"theme":"dark"}' WHERE id = ?`, userID); err != nil {
		t.Fatalf("settings: %v", err)
	}
	if _, err := repo.SelectCurrentCourse(ctx, userID, "es_ru"); err != nil {
		t.Fatalf("select: %v", err)
	}
	var raw string
	if err := conn.QueryRow(`SELECT settings_json FROM users WHERE id = ?`, userID).Scan(&raw); err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(raw, "theme") || !strings.Contains(raw, "current_course_code") {
		t.Fatalf("merged settings: %s", raw)
	}
}
