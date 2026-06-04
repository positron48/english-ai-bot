package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

type backfillTestDB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func TestLinglowEventBackfillRepository_BackfillDryRunCommitAndIdempotency(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(99201)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	courseRepo := NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(ctx, lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	insertBackfillLearningItems(t, conn)

	var grammarTestID, grammarAttemptID, reviewEventID int64
	now := time.Now().UTC().Truncate(time.Second)
	if err := conn.QueryRow(`
		INSERT INTO grammar_test_attempts (
			user_id, scope_type, scope_id, started_at, finished_at, score, passed, total_questions,
			answers_json, results_json, client_attempt_id
		)
		VALUES (?, 'chapter', 'backfill.chapter', ?, ?, 100, 1, 2, '[]', '[]', 'grammar-test-client')
		RETURNING id
	`, user.ID, now, now).Scan(&grammarTestID); err != nil {
		t.Fatalf("insert grammar test attempt: %v", err)
	}
	if err := conn.QueryRow(`
		INSERT INTO grammar_attempts (
			user_id, language, course_id, chapter_id, theory_block_id, concept_id, question_id,
			is_correct, answer_payload_json, correct_payload_json, client_attempt_id, answered_at
		)
		VALUES (?, 'es', 'es', 'backfill.chapter', 'block-1', 'concept-1', 'question-1',
			true, '{"answer":"a"}', '{"answer":"a"}', 'grammar-training-client', ?)
		RETURNING id
	`, user.ID, now).Scan(&grammarAttemptID); err != nil {
		t.Fatalf("insert grammar training attempt: %v", err)
	}

	userCardID := insertBackfillWordFixtures(t, conn, user.ID)
	if err := conn.QueryRow(`
		INSERT INTO review_events (
			user_id, user_card_id, client_attempt_id, direction, answered_at, is_correct, quality,
			options_json, chosen_option, metrics_json, srs_before_json, srs_after_json
		)
		VALUES (?, ?, 'word-review-client', 'es_ru', ?, 1, 5,
			'["дом","кот"]', 'дом', '{"answer_time_ms":1000}', '{"state":"new"}', '{"state":"learning"}')
		RETURNING id
	`, user.ID, userCardID, now).Scan(&reviewEventID); err != nil {
		t.Fatalf("insert review event: %v", err)
	}

	repo := NewLinglowEventBackfillRepository(conn)
	dryRun, err := repo.Backfill(ctx, lc, LinglowEventBackfillOptions{Source: LinglowBackfillSourceAll})
	if err != nil {
		t.Fatalf("dry-run backfill: %v", err)
	}
	assertBackfillSummary(t, dryRun, LinglowBackfillSourceGrammarTests, 1, 0, 1, 0, 0)
	assertBackfillSummary(t, dryRun, LinglowBackfillSourceGrammarTraining, 1, 0, 1, 0, 0)
	assertBackfillSummary(t, dryRun, LinglowBackfillSourceWordReviews, 1, 0, 1, 0, 0)

	commit, err := repo.Backfill(ctx, lc, LinglowEventBackfillOptions{Source: LinglowBackfillSourceAll, Commit: true})
	if err != nil {
		t.Fatalf("commit backfill: %v", err)
	}
	assertBackfillSummary(t, commit, LinglowBackfillSourceGrammarTests, 1, 1, 0, 1, 1)
	assertBackfillSummary(t, commit, LinglowBackfillSourceGrammarTraining, 1, 1, 0, 1, 1)
	assertBackfillSummary(t, commit, LinglowBackfillSourceWordReviews, 1, 1, 0, 1, 1)

	secondCommit, err := repo.Backfill(ctx, lc, LinglowEventBackfillOptions{Source: LinglowBackfillSourceAll, Commit: true})
	if err != nil {
		t.Fatalf("second commit backfill: %v", err)
	}
	assertBackfillSummary(t, secondCommit, LinglowBackfillSourceGrammarTests, 1, 1, 0, 0, 0)
	assertBackfillSummary(t, secondCommit, LinglowBackfillSourceGrammarTraining, 1, 1, 0, 0, 0)
	assertBackfillSummary(t, secondCommit, LinglowBackfillSourceWordReviews, 1, 1, 0, 0, 0)

	assertExerciseAttemptExists(t, conn, "grammar_test_attempts", grammarTestID)
	assertExerciseAttemptExists(t, conn, "grammar_attempts", grammarAttemptID)
	assertExerciseAttemptExists(t, conn, "review_events", reviewEventID)
}

func insertBackfillLearningItems(t *testing.T, conn backfillTestDB) {
	t.Helper()
	if _, err := conn.Exec(`
		INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
		SELECT c.id, d.id, l.id, 'grammar_section:backfill-events', 'grammar', 'Backfill Events', 'grammar_section', 'backfill.section', 1, 'published'
		FROM courses c
		JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
		JOIN locations l ON l.district_id = d.id AND l.location_type = 'grammar'
		WHERE c.code = 'es_ru'
	`); err != nil {
		t.Fatalf("insert grammar module: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'grammar_chapter', 'grammar_chapter', 'backfill.chapter', 'Backfill Chapter', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'grammar_section:backfill-events'
	`); err != nil {
		t.Fatalf("insert grammar learning item: %v", err)
	}
}

func insertBackfillWordFixtures(t *testing.T, conn backfillTestDB, userID int64) int64 {
	t.Helper()
	var wordCardID, trainingCardID, userCardID int64
	if err := conn.QueryRow(`INSERT INTO word_cards (word, definition, display_en) VALUES ('backfill-casa', 'house', 'casa') RETURNING id`).Scan(&wordCardID); err != nil {
		t.Fatalf("insert word card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en) VALUES (?, 'casa', 0, 'дом', 'house') RETURNING id`, wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training card: %v", err)
	}
	if err := conn.QueryRow(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef) VALUES (?, ?, 'es_ru', 'new', 2.5) RETURNING id`, userID, trainingCardID).Scan(&userCardID); err != nil {
		t.Fatalf("insert user card: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO modules (course_id, district_id, location_id, code, module_type, title, source_kind, source_id, sort_order, status)
		SELECT c.id, d.id, l.id, 'word_set:backfill-events', 'word_set', 'Backfill Words', 'word_set', 'backfill-word-set', 1, 'published'
		FROM courses c
		JOIN districts d ON d.course_id = c.id AND d.level_code = 'A0'
		JOIN locations l ON l.district_id = d.id AND l.location_type = 'word_market'
		WHERE c.code = 'es_ru'
	`); err != nil {
		t.Fatalf("insert word module: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO learning_items (course_id, module_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
		SELECT m.course_id, m.id, m.district_id, m.location_id, 'word', 'word_card', ?, 'casa', 'A0', 'published'
		FROM modules m
		JOIN courses c ON c.id = m.course_id
		WHERE c.code = 'es_ru' AND m.code = 'word_set:backfill-events'
	`, fmt.Sprintf("%d", wordCardID)); err != nil {
		t.Fatalf("insert word learning item: %v", err)
	}
	return userCardID
}

func assertBackfillSummary(t *testing.T, summaries []LinglowEventBackfillSummary, source string, legacy, mirrored, missing, processed, inserted int64) {
	t.Helper()
	for _, s := range summaries {
		if s.Source == source {
			if s.LegacyTotal != legacy || s.MirroredTotal != mirrored || s.Missing != missing || s.Processed != processed || s.Inserted != inserted || s.Failed != 0 {
				t.Fatalf("summary %s = %+v, want legacy=%d mirrored=%d missing=%d processed=%d inserted=%d failed=0", source, s, legacy, mirrored, missing, processed, inserted)
			}
			return
		}
	}
	t.Fatalf("summary %s not found in %+v", source, summaries)
}

func assertExerciseAttemptExists(t *testing.T, conn backfillTestDB, sourceTable string, sourcePK int64) {
	t.Helper()
	var exercises, events int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE source_table = ? AND source_pk = ?`, sourceTable, fmt.Sprintf("%d", sourcePK)).Scan(&exercises); err != nil {
		t.Fatalf("count exercise attempt %s/%d: %v", sourceTable, sourcePK, err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM learning_events WHERE source_table = ? AND source_pk = ?`, sourceTable, fmt.Sprintf("%d", sourcePK)).Scan(&events); err != nil {
		t.Fatalf("count learning event %s/%d: %v", sourceTable, sourcePK, err)
	}
	if exercises != 1 || events != 1 {
		t.Fatalf("source %s/%d counts: exercises=%d events=%d, want 1/1", sourceTable, sourcePK, exercises, events)
	}
}
