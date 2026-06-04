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
	insertGrammarTheoryBlockContent(t, conn)
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
	assertGrammarSRSSummary(t, dryRun, 1, 1, 0, 1, 0, 0, 0)

	commit, err := repo.Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("commit grammar srs backfill: %v", err)
	}
	assertGrammarSRSSummary(t, commit, 1, 1, 1, 0, 1, 1, 0)

	second, err := repo.Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("second grammar srs backfill: %v", err)
	}
	assertGrammarSRSSummary(t, second, 1, 1, 1, 0, 0, 0, 0)

	var itemType, sourceID, state string
	var reps, mastery int
	if err := conn.QueryRow(`
		SELECT li.item_type, li.source_id, si.state, si.reps,
			(si.stats_json->'legacy'->>'mastery_score')::int
		FROM srs_items si
		JOIN learning_items li ON li.id = si.learning_item_id
		JOIN user_courses uc ON uc.id = si.user_course_id
		WHERE uc.user_id = ? AND li.source_kind = 'grammar_theory_block'
	`, user.ID).Scan(&itemType, &sourceID, &state, &reps, &mastery); err != nil {
		t.Fatalf("query grammar srs item: %v", err)
	}
	if itemType != "grammar_theory_block" || sourceID != "grammar.chapter.one:block.one" || state != "review" || reps != 4 || mastery != 72 {
		t.Fatalf("unexpected grammar srs item: itemType=%s sourceID=%s state=%s reps=%d mastery=%d", itemType, sourceID, state, reps, mastery)
	}
}

func insertGrammarTheoryBlockContent(t *testing.T, conn backfillTestDB) {
	t.Helper()
	if _, err := conn.Exec(`
		INSERT INTO grammar_content_sections (
			bundle_id, section_id, title, level, sort_order, chapter_ids_json, raw_json, source_hash
		)
		VALUES ('es', 'grammar.section.one', 'Section One', 'A0', 1, '[]', '{}', 'section-hash')
	`); err != nil {
		t.Fatalf("insert grammar section: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_content_chapters (
			bundle_id, chapter_id, section_id, title, ui_language, target_language, level,
			sort_order, raw_json, source_hash
		)
		VALUES ('es', 'grammar.chapter.one', 'grammar.section.one', 'Chapter One', 'ru', 'es', 'A0',
			1, '{}', 'chapter-hash')
	`); err != nil {
		t.Fatalf("insert grammar chapter: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_training_content_questions (
			bundle_id, question_id, chapter_id, theory_block_id, concept_id, difficulty, raw_json, source_hash
		)
		VALUES
			('es', 'q-one', 'grammar.chapter.one', 'block.one', 'concept.one', 1, '{}', 'q-one-hash'),
			('es', 'q-two', 'grammar.chapter.one', 'block.one', 'concept.one', 1, '{}', 'q-two-hash')
	`); err != nil {
		t.Fatalf("insert grammar training questions: %v", err)
	}
}

func assertGrammarSRSSummary(t *testing.T, s *LinglowGrammarSRSBackfillSummary, legacy, mapped, srs, missing, processed, upserted, unmapped int64) {
	t.Helper()
	if s.LegacyTotal != legacy || s.MappedTotal != mapped || s.SRSTotal != srs || s.Missing != missing || s.Processed != processed || s.Upserted != upserted || s.UnmappedTotal != unmapped {
		t.Fatalf("summary = %+v, want legacy=%d mapped=%d srs=%d missing=%d processed=%d upserted=%d unmapped=%d", s, legacy, mapped, srs, missing, processed, upserted, unmapped)
	}
}
