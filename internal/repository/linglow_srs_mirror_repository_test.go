package repository

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestLinglowSRSMirrorRepository_MirrorWordReviewUpdatesCanonicalSRS(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	ctx := context.Background()
	logger := zap.NewNop()
	userRepo := NewUserRepository(conn, logger)
	user, err := userRepo.GetOrCreateUser(99305)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	lc := config.LearningConfig{NativeLang: "ru", TargetLang: "es", GrammarBundleID: "es"}
	courseRepo := NewCourseRepository(conn, logger)
	if _, err := courseRepo.BackfillUserCoursesForLearning(ctx, lc); err != nil {
		t.Fatalf("BackfillUserCoursesForLearning: %v", err)
	}
	userCardID := insertBackfillWordFixtures(t, conn, user.ID)
	nextDue := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	if _, err := conn.Exec(`
		UPDATE user_cards
		SET state = 'review', reps = 5, next_due_at = ?
		WHERE id = ?
	`, nextDue, userCardID); err != nil {
		t.Fatalf("update user card: %v", err)
	}

	wordRepo := NewLinglowWordSRSBackfillRepository(conn)
	if _, err := wordRepo.Backfill(ctx, lc, LinglowWordSRSBackfillOptions{Commit: true}); err != nil {
		t.Fatalf("initial word srs backfill: %v", err)
	}
	futureDue := time.Now().UTC().Add(72 * time.Hour).Truncate(time.Second)
	if _, err := conn.Exec(`UPDATE srs_items SET state = 'learning', reps = 1, due_at = ? WHERE user_course_id IN (SELECT id FROM user_courses WHERE user_id = ?)`, futureDue, user.ID); err != nil {
		t.Fatalf("stale canonical srs: %v", err)
	}

	mirror := NewLinglowSRSMirrorRepository(conn)
	if err := mirror.MirrorWordReview(ctx, lc, user.ID, userCardID); err != nil {
		t.Fatalf("MirrorWordReview: %v", err)
	}

	var state string
	var reps int
	var due time.Time
	if err := conn.QueryRow(`
		SELECT si.state, si.reps, si.due_at
		FROM srs_items si
		JOIN user_courses uc ON uc.id = si.user_course_id
		WHERE uc.user_id = ?
	`, user.ID).Scan(&state, &reps, &due); err != nil {
		t.Fatalf("query mirrored srs: %v", err)
	}
	if state != "review" || reps != 5 || !due.Equal(nextDue) {
		t.Fatalf("mirrored srs = state %s reps %d due %s, want review/5/%s", state, reps, due, nextDue)
	}
}
