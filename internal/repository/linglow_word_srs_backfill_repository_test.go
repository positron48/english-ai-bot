package repository

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestLinglowWordSRSBackfillRepository_BackfillDryRunCommitAndIdempotency(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(99301)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	courseRepo := NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(ctx, lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	userCardID := insertBackfillWordFixtures(t, conn, user.ID)
	nextDue := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	lastReview := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)
	if _, err := conn.Exec(`
		UPDATE user_cards
		SET state = 'review', ef = 2.8, reps = 3, interval_days = 5, learning_step = 2,
			lapse_count = 1, next_due_at = ?, last_review_at = ?, last_quality = 4,
			stats_json = '{"total":3,"correct":2}'
		WHERE id = ?
	`, nextDue, lastReview, userCardID); err != nil {
		t.Fatalf("update user card srs state: %v", err)
	}

	repo := NewLinglowWordSRSBackfillRepository(conn)
	dryRun, err := repo.Backfill(ctx, lc, LinglowWordSRSBackfillOptions{})
	if err != nil {
		t.Fatalf("dry-run word srs backfill: %v", err)
	}
	assertWordSRSSummary(t, dryRun, 1, 1, 0, 1, 0, 0, 0)

	commit, err := repo.Backfill(ctx, lc, LinglowWordSRSBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("commit word srs backfill: %v", err)
	}
	assertWordSRSSummary(t, commit, 1, 1, 1, 0, 1, 1, 0)

	second, err := repo.Backfill(ctx, lc, LinglowWordSRSBackfillOptions{Commit: true})
	if err != nil {
		t.Fatalf("second word srs backfill: %v", err)
	}
	assertWordSRSSummary(t, second, 1, 1, 1, 0, 0, 0, 0)

	var state string
	var reps, lapseCount int
	var statsTotal int
	if err := conn.QueryRow(`
		SELECT si.state, si.reps, si.lapse_count, (si.stats_json->'legacy'->'stats_json'->>'total')::int
		FROM srs_items si
		JOIN user_courses uc ON uc.id = si.user_course_id
		JOIN learning_items li ON li.id = si.learning_item_id
		WHERE uc.user_id = ? AND li.source_kind = 'word_card'
	`, user.ID).Scan(&state, &reps, &lapseCount, &statsTotal); err != nil {
		t.Fatalf("query srs item: %v", err)
	}
	if state != "review" || reps != 3 || lapseCount != 1 || statsTotal != 3 {
		t.Fatalf("unexpected srs snapshot: state=%s reps=%d lapse=%d statsTotal=%d", state, reps, lapseCount, statsTotal)
	}
}

func assertWordSRSSummary(t *testing.T, s *LinglowWordSRSBackfillSummary, legacy, mapped, srs, missing, processed, upserted, unmapped int64) {
	t.Helper()
	if s.LegacyTotal != legacy || s.MappedTotal != mapped || s.SRSTotal != srs || s.Missing != missing || s.Processed != processed || s.Upserted != upserted || s.UnmappedTotal != unmapped {
		t.Fatalf("summary = %+v, want legacy=%d mapped=%d srs=%d missing=%d processed=%d upserted=%d unmapped=%d", s, legacy, mapped, srs, missing, processed, upserted, unmapped)
	}
}
