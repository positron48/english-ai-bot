package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/testutil"
)

func TestDailyJourneyPostgresAtomicityAndConcurrency(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	user, err := NewUserRepository(db, zap.NewNop()).GetOrCreateUser(808201)
	if err != nil {
		t.Fatal(err)
	}
	courses := NewCourseRepository(db, zap.NewNop())
	if _, err := courses.EnsureUserCourse(ctx, user.ID, "en_ru"); err != nil {
		t.Fatal(err)
	}
	repo := NewDailyJourneyRepository(db)
	now := time.Date(2026, 10, 18, 23, 55, 0, 0, time.FixedZone("test", 3*3600))
	// A subject write uses this existing user's setting as an observable side effect.
	if _, err := db.Exec(`UPDATE users SET settings_json='{}' WHERE id=?`, user.ID); err != nil {
		t.Fatal(err)
	}
	if current, err := repo.Current(ctx, user.ID, "en_ru", now.Format("2006-01-02")); err != nil || current != nil {
		t.Fatalf("GET created state: %v %v", current, err)
	}
	var built atomic.Int32
	build := func(_ *sql.Tx) ([]dailyjourney.Step, error) {
		built.Add(1)
		return dailyjourney.Select([]dailyjourney.Step{{Mode: dailyjourney.Words, Title: "Words", TargetCount: 1}, {Mode: dailyjourney.Reading, Title: "Reading", TargetCount: 1}, {Mode: dailyjourney.Grammar, Title: "Grammar", TargetCount: 1}}, rand.New(rand.NewSource(7))), nil
	}
	var wg sync.WaitGroup
	ids := make(chan string, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			j, err := repo.Start(ctx, user.ID, "en_ru", now, build)
			if err != nil {
				errs <- err
				return
			}
			ids <- j.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	id := ""
	for got := range ids {
		if id != "" && id != got {
			t.Fatal("concurrent start returned different routes")
		}
		id = got
	}
	if built.Load() != 1 {
		t.Fatalf("selection built %d times", built.Load())
	}
	j, err := repo.Get(ctx, user.ID, "en_ru", id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, user.ID+999, "en_ru", id); !errors.Is(err, dailyjourney.ErrNotFound) {
		t.Fatalf("foreign owner: %v", err)
	}
	if _, err := repo.Get(ctx, user.ID, "es_ru", id); !errors.Is(err, dailyjourney.ErrNotFound) {
		t.Fatalf("foreign course: %v", err)
	}
	stepID := j.Steps[0].ID
	action := dailyjourney.Action{ClientActionID: "atomic", ExpectedRevision: 0, Type: "answer", Payload: json.RawMessage(`{"answer":"book"}`)}
	deliberate := errors.New("subject failed after write")
	callback := func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step) error {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET settings_json=jsonb_set(settings_json::jsonb,'{daily_test_count}',to_jsonb(COALESCE((settings_json::jsonb->>'daily_test_count')::int,0)+1))::text WHERE id=?`, user.ID); err != nil {
			return err
		}
		s.Status = "completed"
		s.CompletedCount = 1
		return nil
	}
	_, err = repo.Apply(ctx, user.ID, "en_ru", id, stepID, action, func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step) error {
		if err := callback(tx, j, s); err != nil {
			return err
		}
		return deliberate
	})
	if !errors.Is(err, deliberate) {
		t.Fatalf("rollback error: %v", err)
	}
	var count int
	readCount := func() int {
		t.Helper()
		if err := db.QueryRow(`SELECT COALESCE((settings_json::jsonb->>'daily_test_count')::int,0) FROM users WHERE id=?`, user.ID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		return count
	}
	if readCount() != 0 {
		t.Fatal("subject write escaped rollback")
	}
	fresh, err := repo.Get(ctx, user.ID, "en_ru", id)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Revision != 0 || fresh.Steps[0].Status != "pending" {
		t.Fatal("cursor escaped rollback")
	}
	// Same action from eight tabs awards the subject result exactly once.
	errs = make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.Apply(ctx, user.ID, "en_ru", id, stepID, action, callback)
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if readCount() != 1 {
		t.Fatal("subject credited more than once")
	}
	fresh, err = repo.Get(ctx, user.ID, "en_ru", id)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.Revision != 1 || fresh.Steps[0].Status != "completed" {
		t.Fatalf("bad saved cursor: %+v", fresh)
	}
	changed := action
	changed.Payload = json.RawMessage(`{"answer":"other"}`)
	if _, err := repo.Apply(ctx, user.ID, "en_ru", id, stepID, changed, callback); !errors.Is(err, dailyjourney.ErrActionKey) {
		t.Fatalf("different answer accepted: %v", err)
	}
	stale := action
	stale.ClientActionID = "stale"
	if _, err := repo.Apply(ctx, user.ID, "en_ru", id, fresh.Steps[1].ID, stale, callback); !errors.Is(err, dailyjourney.ErrConflict) {
		t.Fatalf("stale accepted: %v", err)
	}
	// A replay still returns its original receipt after later revisions.
	next := action
	next.ClientActionID = "reading"
	next.ExpectedRevision = 1
	if _, err := repo.Apply(ctx, user.ID, "en_ru", id, fresh.Steps[1].ID, next, callback); err != nil {
		t.Fatal(err)
	}
	replay, err := repo.Apply(ctx, user.ID, "en_ru", id, stepID, action, callback)
	if err != nil {
		t.Fatal(err)
	}
	if replay.Revision != 1 {
		t.Fatalf("receipt not stable: %d", replay.Revision)
	}
	// Midnight does not discard the ongoing itinerary; explicit start archives it.
	tomorrow := now.Add(10 * time.Minute)
	current, err := repo.Current(ctx, user.ID, "en_ru", tomorrow.Format("2006-01-02"))
	if err != nil {
		t.Fatal(err)
	}
	if current == nil || current.ID != id {
		t.Fatal("midnight lost active journey")
	}
	today, err := repo.Start(ctx, user.ID, "en_ru", tomorrow, build)
	if err != nil {
		t.Fatal(err)
	}
	if today.ID == id || today.Date == j.Date {
		t.Fatal("new day reused old journey")
	}
	archived, err := repo.Get(ctx, user.ID, "en_ru", id)
	if err != nil {
		t.Fatal(err)
	}
	if archived.Status != "archived" {
		t.Fatal("old active journey not archived")
	}
	// Recreate the repository to prove state isn't process-local.
	restored, err := NewDailyJourneyRepository(db).Get(ctx, user.ID, "en_ru", today.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID != today.ID {
		t.Fatal("restart lost journey")
	}
}

func TestDailyJourneyLifecycleEventsAtomicAndIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	user, err := NewUserRepository(db, zap.NewNop()).GetOrCreateUser(808209)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewCourseRepository(db, zap.NewNop()).EnsureUserCourse(ctx, user.ID, "en_ru"); err != nil {
		t.Fatal(err)
	}
	repo := NewDailyJourneyRepository(db)
	j, err := repo.Start(ctx, user.ID, "en_ru", time.Now(), func(*sql.Tx) ([]dailyjourney.Step, error) {
		return []dailyjourney.Step{{Mode: dailyjourney.Words, Status: "pending", Required: true}, {Mode: dailyjourney.Reading, Status: "unavailable", Required: true}, {Mode: dailyjourney.Grammar, Status: "pending"}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	count := func(kind string) int {
		t.Helper()
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM daily_journey_events WHERE journey_id=? AND event_type=?`, j.ID, kind).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if count("journey_started") != 1 || count("step_unavailable") != 1 {
		t.Fatal("missing initial lifecycle events")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rolled := *j
	rolled.Steps = append([]dailyjourney.Step(nil), j.Steps...)
	rolled.Steps[0].Status = "completed"
	if err := recordJourneyLifecycle(ctx, tx, &rolled, false); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if count("step_completed") != 0 {
		t.Fatal("event escaped rollback")
	}
	action := dailyjourney.Action{ClientActionID: "finish", Type: "ack_feedback", ExpectedRevision: 0, Payload: json.RawMessage(`{}`)}
	stepID := j.Current().ID
	for i := 0; i < 8; i++ {
		j, err = repo.Apply(ctx, user.ID, "en_ru", j.ID, stepID, action, func(_ *sql.Tx, _ *dailyjourney.Journey, s *dailyjourney.Step) error {
			s.Status = "completed"
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if count("step_completed") != 1 || count("journey_started") != 1 {
		t.Fatal("replay duplicated events")
	}
	j, err = repo.Apply(ctx, user.ID, "en_ru", j.ID, j.Current().ID, dailyjourney.Action{ClientActionID: "skip", Type: "skip", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{}`)}, func(_ *sql.Tx, _ *dailyjourney.Journey, s *dailyjourney.Step) error { s.Status = "skipped"; return nil })
	if err != nil {
		t.Fatal(err)
	}
	if count("step_skipped") != 1 || count("journey_completed") != 1 || count("step_unavailable") != 1 {
		t.Fatal("incorrect final lifecycle events")
	}
}
