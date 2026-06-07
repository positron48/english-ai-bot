package repository

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestLinglowGrammarSRSBackfillRepository_BackfillTheoryBlockSnapshots(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(99401)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	courseRepo := NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(ctx, lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	insertGrammarTheoryBlockContent(t, conn, "grammar.section.one", "grammar.chapter.one", "block.one", "concept.one")
	if _, err := courseRepo.MapLegacyContentForLearning(ctx, lc); err != nil {
		t.Fatalf("MapLegacyContentForLearning: %v", err)
	}
	nextReview := time.Now().UTC().Add(2 * time.Hour).Truncate(time.Second)
	lastReview := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	if _, err := conn.Exec(`
		INSERT INTO grammar_theory_memory (
			user_id, language, course_id, chapter_id, theory_block_id, concept_id, state,
			review_count, correct_count, wrong_count, lapse_count, correct_streak, wrong_streak,
			ease, interval_days, mastery_score, next_review_at, last_review_at
		)
		VALUES (?, 'es', 'es', 'grammar.chapter.one', 'block.one', 'concept.one', 'review',
			4, 3, 1, 1, 2, 0, 2.7, 6, 72, ?, ?)
	`, user.ID, nextReview, lastReview); err != nil {
		t.Fatalf("insert grammar theory memory: %v", err)
	}

	repo := NewLinglowGrammarSRSBackfillRepository(conn)
	dryRun, err := repo.Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{})
	if err != nil {
		t.Fatalf("dry-run grammar srs backfill: %v", err)
	}
	if dryRun.Missing < 1 || dryRun.Processed != 0 || dryRun.Upserted != 0 {
		t.Fatalf("dry-run summary = %+v, want at least one missing and no writes", dryRun)
	}

	commit, err := repo.Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("commit grammar srs backfill: %v", err)
	}
	if commit.Processed < 1 || commit.Upserted < 1 {
		t.Fatalf("commit summary = %+v, want at least one processed/upserted", commit)
	}

	second, err := repo.Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("second grammar srs backfill: %v", err)
	}
	if second.Processed != 0 || second.Upserted != 0 {
		t.Fatalf("second summary = %+v, want idempotent no-op", second)
	}

	var itemType, sourceID, state string
	var reps, mastery int
	if err := conn.QueryRow(`
		SELECT li.item_type, li.source_id, si.state, si.reps,
			(si.stats_json->'legacy'->>'mastery_score')::int
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		JOIN user_courses uc ON uc.id = si.user_course_id
		WHERE uc.user_id = ? AND li.source_kind = 'grammar_theory_block' AND li.source_id = 'grammar.chapter.one:block.one'
	`, user.ID).Scan(&itemType, &sourceID, &state, &reps, &mastery); err != nil {
		t.Fatalf("query grammar srs item: %v", err)
	}
	if itemType != "grammar_theory_block" || sourceID != "grammar.chapter.one:block.one" || state != "review" || reps != 4 || mastery != 72 {
		t.Fatalf("unexpected grammar srs item: itemType=%s sourceID=%s state=%s reps=%d mastery=%d", itemType, sourceID, state, reps, mastery)
	}
}

func TestLinglowGrammarSRSBackfillRepository_PromotesDueNewStateToLearning(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(99402)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	courseRepo := NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(ctx, lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	insertGrammarTheoryBlockContent(t, conn, "grammar.section.two", "grammar.chapter.two", "block.two", "concept.two")
	if _, err := courseRepo.MapLegacyContentForLearning(ctx, lc); err != nil {
		t.Fatalf("MapLegacyContentForLearning: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_theory_memory (
			user_id, language, course_id, chapter_id, theory_block_id, concept_id, state, next_review_at
		)
		VALUES (?, 'es', 'es', 'grammar.chapter.two', 'block.two', 'concept.two', 'new', CURRENT_TIMESTAMP - INTERVAL '1 hour')
	`, user.ID); err != nil {
		t.Fatalf("insert grammar theory memory: %v", err)
	}

	repo := NewLinglowGrammarSRSBackfillRepository(conn)
	if _, err := repo.Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{Commit: true, Resync: true}); err != nil {
		t.Fatalf("grammar resync: %v", err)
	}

	var state string
	if err := conn.QueryRow(`
		SELECT si.state
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		JOIN user_courses uc ON uc.id = si.user_course_id
		WHERE uc.user_id = ? AND li.source_id = 'grammar.chapter.two:block.two'
	`, user.ID).Scan(&state); err != nil {
		t.Fatalf("query grammar srs state: %v", err)
	}
	if state != "learning" {
		t.Fatalf("due legacy new grammar state = %q, want learning", state)
	}
}

func insertGrammarTheoryBlockContent(t *testing.T, conn backfillTestDB, sectionID, chapterID, blockID, conceptID string) {
	t.Helper()
	if _, err := conn.Exec(`
		INSERT INTO grammar_content_sections (
			bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash
		)
		VALUES ('es', ?, 'Section One', 'A0', 1, '[]', '{}', 'section-hash')
	`, sectionID); err != nil {
		t.Fatalf("insert grammar section: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_content_chapters (
			bundle_id, chapter_id, section_id, title, ui_language, target_language, level,
			sort_order, raw_json, source_hash
		)
		VALUES ('es', ?, ?, 'Chapter One', 'ru', 'es', 'A0',
			1, '{}', 'chapter-hash')
	`, chapterID, sectionID); err != nil {
		t.Fatalf("insert grammar chapter: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_training_content_questions (
			bundle_id, question_id, chapter_id, theory_block_id, concept_id, difficulty, raw_json, source_hash
		)
		VALUES
			('es', ? || '-q-one', ?, ?, ?, 1, '{}', 'q-one-hash'),
			('es', ? || '-q-two', ?, ?, ?, 1, '{}', 'q-two-hash')
	`, chapterID, chapterID, blockID, conceptID, chapterID, chapterID, blockID, conceptID); err != nil {
		t.Fatalf("insert grammar training questions: %v", err)
	}
}
