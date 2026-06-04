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

func TestLinglowMediaProgressBackfillRepository_BackfillReadingAndSpeaking(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(99601)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	courseRepo := NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(ctx, lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	insertMediaContentFixtures(t, conn)
	if _, err := courseRepo.MapLegacyContentForLearning(ctx, lc); err != nil {
		t.Fatalf("MapLegacyContentForLearning: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	if _, err := conn.Exec(`INSERT INTO reading_text_progress (user_id, chapter_id, read_at) VALUES (?, 'reading.text.one', ?)`, user.ID, now); err != nil {
		t.Fatalf("insert reading progress: %v", err)
	}
	var sessionID, attemptID int64
	if err := conn.QueryRow(`INSERT INTO speaking_sessions (user_id, category_id, status, task_ids, current_task_index) VALUES (?, 'speaking.category.one', 'completed', '["speaking.task.one"]', 1) RETURNING id`, user.ID).Scan(&sessionID); err != nil {
		t.Fatalf("insert speaking session: %v", err)
	}
	if err := conn.QueryRow(`
		INSERT INTO speaking_attempts (
			user_id, session_id, task_id, attempt_no, mode, understood_answer,
			meaning_score, grammar_score, pronunciation_score, fluency_score,
			is_acceptable, audio_quality, feedback_ru, better_version, repeat_task, created_at
		)
		VALUES (?, ?, 'speaking.task.one', 1, 'initial', 'hola',
			80, 70, 90, 60, true, 'ok', 'good', 'hola', '', ?)
		RETURNING id
	`, user.ID, sessionID, now).Scan(&attemptID); err != nil {
		t.Fatalf("insert speaking attempt: %v", err)
	}

	repo := NewLinglowMediaProgressBackfillRepository(conn)
	dryRun, err := repo.Backfill(ctx, lc, LinglowMediaProgressBackfillOptions{})
	if err != nil {
		t.Fatalf("dry-run media progress backfill: %v", err)
	}
	assertMediaSummaryAtLeast(t, dryRun, "reading", 1, 0, 0)
	assertMediaSummaryAtLeast(t, dryRun, "speaking", 1, 0, 0)

	commit, err := repo.Backfill(ctx, lc, LinglowMediaProgressBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("commit media progress backfill: %v", err)
	}
	assertMediaSummaryAtLeast(t, commit, "reading", 0, 1, 1)
	assertMediaSummaryAtLeast(t, commit, "speaking", 0, 1, 1)

	second, err := repo.Backfill(ctx, lc, LinglowMediaProgressBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("second media progress backfill: %v", err)
	}
	assertMediaSummaryAtLeast(t, second, "reading", 0, 0, 0)
	assertMediaSummaryAtLeast(t, second, "speaking", 0, 0, 0)

	var readingEvents, speakingAttempts int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM learning_events WHERE source_table = 'reading_text_progress'`).Scan(&readingEvents); err != nil {
		t.Fatalf("count reading events: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE source_table = 'speaking_attempts' AND source_pk = ?`, fmt.Sprintf("%d", attemptID)).Scan(&speakingAttempts); err != nil {
		t.Fatalf("count speaking attempts: %v", err)
	}
	if readingEvents < 1 || speakingAttempts != 1 {
		t.Fatalf("unexpected media counts: readingEvents=%d speakingAttempts=%d", readingEvents, speakingAttempts)
	}
}

func insertMediaContentFixtures(t *testing.T, conn backfillTestDB) {
	t.Helper()
	if _, err := conn.Exec(`
		INSERT INTO reading_categories (category_id, title, level, sort_order, text_ids)
		VALUES ('reading.category.one', 'Reading Category', 'A0', 1, '["reading.text.one"]')
	`); err != nil {
		t.Fatalf("insert reading category: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO reading_texts (text_id, category_id, title, level, target_language, reading_passage)
		VALUES ('reading.text.one', 'reading.category.one', 'Reading Text', 'A0', 'es', 'Hola')
	`); err != nil {
		t.Fatalf("insert reading text: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO speaking_categories (category_id, title, level, sort_order, task_ids)
		VALUES ('speaking.category.one', 'Speaking Category', 'A0', 1, '["speaking.task.one"]')
	`); err != nil {
		t.Fatalf("insert speaking category: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO speaking_tasks (task_id, category_id, title, level, task_type, target_language, sort_order, task_json)
		VALUES ('speaking.task.one', 'speaking.category.one', 'Speaking Task', 'A0', 'repeat', 'es', 1, '{}')
	`); err != nil {
		t.Fatalf("insert speaking task: %v", err)
	}
}

func assertMediaSummaryAtLeast(t *testing.T, summaries []LinglowMediaProgressBackfillSummary, source string, missingAtLeast, processedAtLeast, insertedAtLeast int64) {
	t.Helper()
	for _, s := range summaries {
		if s.Source == source {
			if s.Missing < missingAtLeast || s.Processed < processedAtLeast || s.Inserted < insertedAtLeast {
				t.Fatalf("summary %s = %+v, want missing>=%d processed>=%d inserted>=%d", source, s, missingAtLeast, processedAtLeast, insertedAtLeast)
			}
			return
		}
	}
	t.Fatalf("summary %s not found in %+v", source, summaries)
}
