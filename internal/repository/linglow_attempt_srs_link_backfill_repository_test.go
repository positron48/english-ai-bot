package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestLinglowAttemptSRSLinkBackfillRepository_BackfillWordAndGrammarLinks(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(99501)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	courseRepo := NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(ctx, lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	userCardID := insertBackfillWordFixtures(t, conn, user.ID)
	insertGrammarTheoryBlockContent(t, conn, "grammar.section.link", "grammar.chapter.link", "block.link", "concept.link")
	if _, err := courseRepo.MapLegacyContentForLearning(ctx, lc); err != nil {
		t.Fatalf("MapLegacyContentForLearning: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	var reviewEventID, grammarAttemptID int64
	if err := conn.QueryRow(`
		INSERT INTO review_events (
			user_id, user_card_id, client_attempt_id, direction, answered_at, is_correct, quality,
			options_json, chosen_option
		)
		VALUES (?, ?, 'word-link-client', 'es_ru', ?, 1, 5, '["дом"]', 'дом')
		RETURNING id
	`, user.ID, userCardID, now).Scan(&reviewEventID); err != nil {
		t.Fatalf("insert review event: %v", err)
	}
	if err := conn.QueryRow(`
		INSERT INTO grammar_attempts (
			user_id, language, course_id, chapter_id, theory_block_id, concept_id, question_id,
			is_correct, answer_payload_json, correct_payload_json, client_attempt_id, answered_at
		)
		VALUES (?, 'es', 'es', 'grammar.chapter.link', 'block.link', 'concept.link', 'q-link',
			true, '{}', '{}', 'grammar-link-client', ?)
		RETURNING id
	`, user.ID, now).Scan(&grammarAttemptID); err != nil {
		t.Fatalf("insert grammar attempt: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO grammar_theory_memory (
			user_id, language, course_id, chapter_id, theory_block_id, concept_id, state,
			review_count, correct_count, wrong_count, lapse_count, ease, interval_days, mastery_score, next_review_at
		)
		VALUES (?, 'es', 'es', 'grammar.chapter.link', 'block.link', 'concept.link', 'review',
			1, 1, 0, 0, 2.5, 1, 50, ?)
	`, user.ID, now); err != nil {
		t.Fatalf("insert grammar memory: %v", err)
	}
	if _, err := NewLinglowWordSRSBackfillRepository(conn).Backfill(ctx, lc, LinglowWordSRSBackfillOptions{Commit: true}); err != nil {
		t.Fatalf("word srs backfill: %v", err)
	}
	if _, err := NewLinglowGrammarSRSBackfillRepository(conn).Backfill(ctx, lc, LinglowGrammarSRSBackfillOptions{Commit: true}); err != nil {
		t.Fatalf("grammar srs backfill: %v", err)
	}

	eventRepo := NewLinglowEventRepository(conn)
	if _, err := eventRepo.RecordWordReviewEvent(ctx, lc, WordReviewEventInput{UserID: user.ID, ReviewEventID: reviewEventID, UserCardID: userCardID, Direction: "es_ru", IsCorrect: true, Quality: 5, OptionsJSON: `["дом"]`, ChosenOption: "дом", AnsweredAt: now}); err != nil {
		t.Fatalf("RecordWordReviewEvent: %v", err)
	}
	if _, err := eventRepo.RecordGrammarTrainingAttempt(ctx, lc, GrammarTrainingEventInput{UserID: user.ID, AttemptID: grammarAttemptID, ChapterID: "grammar.chapter.link", TheoryBlockID: "block.link", ConceptID: "concept.link", QuestionID: "q-link", IsCorrect: true, AnswerJSON: `{}`, CorrectJSON: `{}`, AnsweredAt: now}); err != nil {
		t.Fatalf("RecordGrammarTrainingAttempt: %v", err)
	}

	var linkedBefore int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE srs_item_id IS NOT NULL`).Scan(&linkedBefore); err != nil {
		t.Fatalf("count linked before: %v", err)
	}
	if linkedBefore != 2 {
		t.Fatalf("runtime writer linked %d attempts, want 2", linkedBefore)
	}

	if _, err := conn.Exec(`UPDATE exercise_attempts SET srs_item_id = NULL WHERE mode IN ('word_training', 'grammar_training')`); err != nil {
		t.Fatalf("clear srs links: %v", err)
	}
	repo := NewLinglowAttemptSRSLinkBackfillRepository(conn)
	dryRun, err := repo.Backfill(ctx, lc, LinglowAttemptSRSLinkBackfillOptions{})
	if err != nil {
		t.Fatalf("dry-run attempt srs links: %v", err)
	}
	assertAttemptLinkSummary(t, dryRun, "word-training", 1, 0, 0)
	assertAttemptLinkSummary(t, dryRun, "grammar-training", 1, 0, 0)

	commit, err := repo.Backfill(ctx, lc, LinglowAttemptSRSLinkBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("commit attempt srs links: %v", err)
	}
	assertAttemptLinkSummary(t, commit, "word-training", 0, 1, 0)
	assertAttemptLinkSummary(t, commit, "grammar-training", 0, 1, 0)

	var grammarItemType string
	if err := conn.QueryRow(`
		SELECT li.item_type
		FROM exercise_attempts ea
		JOIN learning_items li ON li.id = ea.learning_item_id
		WHERE ea.source_table = 'grammar_attempts' AND ea.source_pk = ?
	`, fmt.Sprintf("%d", grammarAttemptID)).Scan(&grammarItemType); err != nil {
		t.Fatalf("query grammar attempt item type: %v", err)
	}
	if grammarItemType != "grammar_theory_block" {
		t.Fatalf("grammar attempt item type = %s, want grammar_theory_block", grammarItemType)
	}
}

func assertAttemptLinkSummary(t *testing.T, summaries []LinglowAttemptSRSLinkBackfillSummary, source string, missing, updated, unmapped int64) {
	t.Helper()
	for _, s := range summaries {
		if s.Source == source {
			if s.Missing != missing || s.Updated != updated || s.UnmappedTotal != unmapped {
				t.Fatalf("summary %s = %+v, want missing=%d updated=%d unmapped=%d", source, s, missing, updated, unmapped)
			}
			return
		}
	}
	t.Fatalf("summary %s not found in %+v", source, summaries)
}
