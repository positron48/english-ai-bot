package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestWordTargetForLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		level  string
		total  int
		target int
	}{
		{"A0", 100, 40},
		{"A0", 20, 20},
		{"A1", 200, 80},
		{"A2", 300, 120},
		{"B1", 400, 160},
		{"C1", 1000, 350},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.level, func(t *testing.T) {
			t.Parallel()
			if got := wordTargetForLevel(tc.level, tc.total); got != tc.target {
				t.Fatalf("wordTargetForLevel(%s, %d) = %d, want %d", tc.level, tc.total, got, tc.target)
			}
		})
	}
}

func TestWeightedMasteryPercent_FreeIgnoresProMetrics(t *testing.T) {
	t.Parallel()
	metrics := map[string]CourseMasteryMetric{
		"grammar":      {Percent: 100, Target: 10, Included: true},
		"words":        {Percent: 0, Target: 40, Included: true},
		"reading":      {Percent: 0, Target: 3, Included: true},
		"conversation": {Percent: 0, Target: 3, Included: false},
		"picture":      {Percent: 0, Target: 3, Included: false},
	}
	got := weightedMasteryPercent(metrics, false, true, true)
	want := 45.0
	if got != want {
		t.Fatalf("free weighted = %v, want %v", got, want)
	}
}

func TestWeightedMasteryPercent_ProIncludesConversationPicture(t *testing.T) {
	t.Parallel()
	metrics := map[string]CourseMasteryMetric{
		"grammar":      {Percent: 100, Target: 10, Included: true},
		"words":        {Percent: 100, Target: 40, Included: true},
		"reading":      {Percent: 100, Target: 3, Included: true},
		"conversation": {Percent: 100, Target: 3, Included: true},
		"picture":      {Percent: 100, Target: 3, Included: true},
	}
	got := weightedMasteryPercent(metrics, true, true, true)
	if got < 99.9 || got > 100.1 {
		t.Fatalf("pro weighted = %v, want ~100", got)
	}
}

func TestBuildCourseMastery_GrammarUnlock(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewCourseRepository(conn, zap.NewNop())
	ctx := context.Background()
	userRepo := NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(88001)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	courseID, userCourseID := insertMasteryCourseFixture(t, conn, user.ID)

	for _, q := range []string{
		`INSERT INTO grammar_content_sections (bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash)
		 VALUES ('es', 'es.section.a0', 'A0 Grammar', 'A0', 1, '["es.chapter.a0"]', '{}', 'sec-a0')`,
		`INSERT INTO grammar_content_chapters (bundle_id, chapter_id, section_id, title, ui_language, target_language, level, sort_order, raw_json, source_hash)
		 VALUES ('es', 'es.chapter.a0', 'es.section.a0', 'Chapter A0', 'ru', 'es', 'A0', 1, '{}', 'ch-a0')`,
		`INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at) VALUES ('section', 'es.section.a0', 1, CURRENT_TIMESTAMP)`,
		`INSERT INTO grammar_published_items (item_type, item_id, is_published, updated_at) VALUES ('chapter', 'es.chapter.a0', 1, CURRENT_TIMESTAMP)`,
	} {
		if _, err := conn.Exec(q); err != nil {
			t.Fatalf("fixture: %v\n%s", err, q)
		}
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_progress (user_id, chapter_id, best_score, passed_at, last_attempt_at)
		VALUES (?, 'es.chapter.a0', 80, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, user.ID); err != nil {
		t.Fatalf("progress: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_test_attempts (user_id, scope_type, scope_id, started_at, finished_at, score, passed, total_questions, answers_json, results_json)
		VALUES (?, 'category', 'es.section.a0', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 60, 1, 5, '[]', '[]')`, user.ID); err != nil {
		t.Fatalf("category attempt: %v", err)
	}

	mastery, err := repo.BuildCourseMastery(ctx, user.ID, courseID, userCourseID, "es_ru", "es", models.TierFree)
	if err != nil {
		t.Fatalf("BuildCourseMastery: %v", err)
	}
	lv := MasteryLevelByCode(mastery, "A0")
	if lv == nil {
		t.Fatal("missing A0 level")
	}
	if !lv.CanOpenNext {
		t.Fatalf("expected grammar unlock for A0, metrics=%+v", lv.Metrics)
	}
	if lv.Metrics["grammar"].Done != 2 || lv.Metrics["grammar"].Target != 2 {
		t.Fatalf("grammar metric = %+v", lv.Metrics["grammar"])
	}
}

func TestBuildCourseMastery_WordUnlockWithMasteringScore(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewCourseRepository(conn, zap.NewNop())
	ctx := context.Background()
	userRepo := NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(88002)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	courseID, userCourseID := insertMasteryCourseFixture(t, conn, user.ID)

	var rootCatID int64
	if err := conn.QueryRow(`
		INSERT INTO word_set_categories (parent_id, name, description, level_code, is_published, sort_order, course_code)
		VALUES (NULL, 'Mastery A0', 'A0', 'A0', 1, 99, 'es_ru') RETURNING id`).Scan(&rootCatID); err != nil {
		t.Fatalf("root category: %v", err)
	}
	var wordSetID int64
	if err := conn.QueryRow(`
		INSERT INTO word_sets (category_id, title, description, is_published, sort_order, course_code)
		VALUES (?, 'Mastery Set', 'set', 1, 1, 'es_ru') RETURNING id`, rootCatID).Scan(&wordSetID); err != nil {
		t.Fatalf("set: %v", err)
	}
	for i := 1; i <= 50; i++ {
		var cardID int64
		word := fmt.Sprintf("lemma%d", i)
		if err := conn.QueryRow(`INSERT INTO word_cards (word, definition) VALUES (?, 'def') RETURNING id`, word).Scan(&cardID); err != nil {
			t.Fatalf("card: %v", err)
		}
		if _, err := conn.Exec(`INSERT INTO word_set_items (word_set_id, word_card_id, sort_order) VALUES (?, ?, ?)`, wordSetID, cardID, i); err != nil {
			t.Fatalf("item: %v", err)
		}
		if i <= 40 {
			if _, err := conn.Exec(`INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score, course_code) VALUES (?, ?, 85, 'es_ru')`, user.ID, cardID); err != nil {
				t.Fatalf("mastering: %v", err)
			}
		}
	}

	mastery, err := repo.BuildCourseMastery(ctx, user.ID, courseID, userCourseID, "es_ru", "es", models.TierFree)
	if err != nil {
		t.Fatalf("BuildCourseMastery: %v", err)
	}
	lv := MasteryLevelByCode(mastery, "A0")
	if lv == nil {
		t.Fatal("missing A0")
	}
	if !lv.CanOpenNext {
		t.Fatalf("expected word unlock, words=%+v", lv.Metrics["words"])
	}
	if lv.Metrics["words"].Done < 40 {
		t.Fatalf("mastered words = %d", lv.Metrics["words"].Done)
	}
}

func TestBuildCourseMastery_ReadingTargetCapped(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewCourseRepository(conn, zap.NewNop())
	ctx := context.Background()
	userRepo := NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(88003)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	courseID, userCourseID := insertMasteryCourseFixture(t, conn, user.ID)

	if _, err := conn.Exec(`INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids) VALUES ('mastery.read.a0', 'A0 Read', 'A0', 1, '[]')`); err != nil {
		t.Fatalf("category: %v", err)
	}
	for i := 0; i < 10; i++ {
		textID := fmt.Sprintf("mastery.read.a0.%d", i)
		if _, err := conn.Exec(`
			INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
			VALUES (?, 'mastery.read.a0', ?, 'A0', 'es', 'texto')`, textID, textID); err != nil {
			t.Fatalf("text: %v", err)
		}
		if i < 3 {
			if _, err := conn.Exec(`INSERT INTO reading_text_progress (user_id, chapter_id) VALUES (?, ?)`, user.ID, textID); err != nil {
				t.Fatalf("progress: %v", err)
			}
		}
	}

	mastery, err := repo.BuildCourseMastery(ctx, user.ID, courseID, userCourseID, "es_ru", "es", models.TierFree)
	if err != nil {
		t.Fatalf("BuildCourseMastery: %v", err)
	}
	read := MasteryLevelByCode(mastery, "A0").Metrics["reading"]
	if read.Target != 3 {
		t.Fatalf("reading target = %d, want 3", read.Target)
	}
	if read.Percent != 100 {
		t.Fatalf("reading percent = %v, want 100", read.Percent)
	}
}

func TestBuildCourseMastery_ProWeightsIncludeConversation(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	repo := NewCourseRepository(conn, zap.NewNop())
	ctx := context.Background()
	userRepo := NewUserRepository(conn, zap.NewNop())
	user, err := userRepo.GetOrCreateUser(88004)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	courseID, userCourseID := insertMasteryCourseFixture(t, conn, user.ID)

	var districtID int64
	if err := conn.QueryRow(`SELECT id FROM districts WHERE course_id = ? AND code = 'a0_spark_gate'`, courseID).Scan(&districtID); err != nil {
		t.Fatalf("district: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO conversation_scenarios (course_id, district_id, code, place_type, cefr_level, title, npc_name, npc_persona, scene_setup, is_quest, max_turns, token_budget, status, sort_order)
		VALUES (?, ?, 'cov.scenario', 'cafe', 'A0', 'Cafe', 'Ana', 'friendly', 'scene', true, 10, 1000, 'active', 1)`, courseID, districtID); err != nil {
		t.Fatalf("scenario: %v", err)
	}

	freeMastery, err := repo.BuildCourseMastery(ctx, user.ID, courseID, userCourseID, "es_ru", "es", models.TierFree)
	if err != nil {
		t.Fatalf("free mastery: %v", err)
	}
	freeConv := MasteryLevelByCode(freeMastery, "A0").Metrics["conversation"]
	if freeConv.Included {
		t.Fatal("free user should not include conversation in weights")
	}

	proMastery, err := repo.BuildCourseMastery(ctx, user.ID, courseID, userCourseID, "es_ru", "es", models.TierPro)
	if err != nil {
		t.Fatalf("pro mastery: %v", err)
	}
	proConv := MasteryLevelByCode(proMastery, "A0").Metrics["conversation"]
	if !proConv.Included {
		t.Fatal("pro user should include conversation when content exists")
	}
}

func insertMasteryCourseFixture(t *testing.T, conn *sql.DB, userID int64) (courseID, userCourseID int64) {
	t.Helper()
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	repo := NewCourseRepository(conn, zap.NewNop())
	if _, err := repo.BackfillUserCoursesForLearning(context.Background(), lc); err != nil {
		t.Fatalf("backfill user courses: %v", err)
	}
	if err := conn.QueryRow(`SELECT id FROM courses WHERE code = 'es_ru'`).Scan(&courseID); err != nil {
		t.Fatalf("course id: %v", err)
	}
	if err := conn.QueryRow(`SELECT id FROM user_courses WHERE user_id = ? AND course_id = ?`, userID, courseID).Scan(&userCourseID); err != nil {
		t.Fatalf("user course id: %v", err)
	}
	return courseID, userCourseID
}
