package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/testutil"
)

func journeyAIRepoFixture(t *testing.T, telegram int64) (*sql.DB, *DailyJourneyRepository, int64, *dailyjourney.Journey) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	user, err := NewUserRepository(db, zap.NewNop()).GetOrCreateUser(telegram)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewCourseRepository(db, zap.NewNop()).EnsureUserCourse(context.Background(), user.ID, "en_ru"); err != nil {
		t.Fatal(err)
	}
	repo := NewDailyJourneyRepository(db)
	j, err := repo.Start(context.Background(), user.ID, "en_ru", time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC), func(*sql.Tx) ([]dailyjourney.Step, error) {
		return []dailyjourney.Step{{Mode: dailyjourney.Picture, Status: "active", Title: "Picture", TargetCount: 1, State: json.RawMessage(`{}`)}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return db, repo, user.ID, j
}
func TestDailyJourneyAIReservationConcurrencyAndAtomicCommit(t *testing.T) {
	db, repo, user, j := journeyAIRepoFixture(t, 808601)
	ctx := context.Background()
	step := j.Current()
	action := dailyjourney.Action{ClientActionID: "ai-message", Type: "message", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{"text":"Hello"}`)}
	var prepared atomic.Int32
	prepare := func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step, previous json.RawMessage) (json.RawMessage, error) {
		prepared.Add(1)
		return json.RawMessage(`{"private_prompt":"private input","text":"Hello"}`), nil
	}
	var wg sync.WaitGroup
	claims := make(chan *JourneyAIReservation, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reservation, err := repo.ReserveAI(ctx, user, "en_ru", j.ID, step.ID, action, time.Minute, prepare)
			if err != nil {
				errs <- err
			} else {
				claims <- reservation
			}
		}()
	}
	wg.Wait()
	close(claims)
	close(errs)
	var reservation *JourneyAIReservation
	for claim := range claims {
		if reservation != nil {
			t.Fatal("two live workers")
		}
		reservation = claim
	}
	if reservation == nil || prepared.Load() != 1 {
		t.Fatal("no unique reservation")
	}
	for err := range errs {
		if !errors.Is(err, dailyjourney.ErrProcessing) {
			t.Fatal(err)
		}
	}
	current, err := repo.Get(ctx, user, "en_ru", j.ID)
	if err != nil || current.Revision != j.Revision || current.Current().CompletedCount != 0 {
		t.Fatal("reservation credited progress", current, err)
	}
	other := action
	other.ClientActionID = "skip"
	other.Type = "skip"
	if _, err := repo.Apply(ctx, user, "en_ru", j.ID, step.ID, other, func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step) error {
		t.Fatal("sync callback ran during AI")
		return nil
	}); !errors.Is(err, dailyjourney.ErrProcessing) {
		t.Fatal("sync action bypassed reservation", err)
	}
	changed := action
	changed.Payload = json.RawMessage(`{"text":"Changed"}`)
	if _, err := repo.ReserveAI(ctx, user, "en_ru", j.ID, step.ID, changed, time.Minute, prepare); !errors.Is(err, dailyjourney.ErrActionKey) {
		t.Fatal("changed payload accepted", err)
	}
	if _, err := repo.ReserveAI(ctx, user, "es_ru", j.ID, step.ID, action, time.Minute, prepare); !errors.Is(err, dailyjourney.ErrNotFound) {
		t.Fatal("foreign course read", err)
	}
	forced := errors.New("after subject write")
	_, err = repo.CompleteAI(ctx, user, "en_ru", j.ID, step.ID, action, reservation.Token, func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step) error {
		if _, err := tx.Exec(`UPDATE users SET settings_json='{"ai_applied":true}' WHERE id=?`, user); err != nil {
			return err
		}
		s.CompletedCount = 1
		return forced
	})
	if !errors.Is(err, forced) {
		t.Fatal(err)
	}
	var raw string
	if err := db.QueryRow(`SELECT COALESCE(settings_json,'{}') FROM users WHERE id=?`, user).Scan(&raw); err != nil || raw == `{"ai_applied":true}` {
		t.Fatal("subject did not roll back", raw, err)
	}
	var applied atomic.Int32
	errs = make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.CompleteAI(ctx, user, "en_ru", j.ID, step.ID, action, reservation.Token, func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step) error {
				applied.Add(1)
				if _, err := tx.Exec(`UPDATE users SET settings_json='{"ai_applied":true}' WHERE id=?`, user); err != nil {
					return err
				}
				s.CompletedCount = 1
				return nil
			})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if applied.Load() != 1 {
		t.Fatal("duplicate AI credit", applied.Load())
	}
	replay, err := NewDailyJourneyRepository(db).ReserveAI(ctx, user, "en_ru", j.ID, step.ID, action, time.Minute, prepare)
	if err != nil || replay.Result == nil || replay.Token != "" || replay.Result.Revision != j.Revision+1 {
		t.Fatal("lost durable receipt", replay, err)
	}
	var input []byte
	var token sql.NullString
	if err := db.QueryRow(`SELECT prepared_json,reservation_token FROM daily_journey_actions WHERE journey_id=? AND client_action_id=?`, j.ID, action.ClientActionID).Scan(&input, &token); err != nil || len(input) != 0 || token.Valid {
		t.Fatal("private input kept after completion", err)
	}
}
func TestDailyJourneyAIExpiredWorkerFencingAndRetry(t *testing.T) {
	db, repo, user, j := journeyAIRepoFixture(t, 808602)
	ctx := context.Background()
	action := dailyjourney.Action{ClientActionID: "retry", Type: "message", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{"text":"hello"}`)}
	step := j.Current()
	prepare := func(_ *sql.Tx, _ *dailyjourney.Journey, _ *dailyjourney.Step, previous json.RawMessage) (json.RawMessage, error) {
		if len(previous) > 0 {
			return previous, nil
		}
		return json.RawMessage(`{"snapshot":"original"}`), nil
	}
	first, err := repo.ReserveAI(ctx, user, "en_ru", j.ID, step.ID, action, time.Minute, prepare)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE daily_journey_actions SET lease_expires_at=CURRENT_TIMESTAMP-INTERVAL '1 second' WHERE journey_id=?`, j.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CompleteAI(ctx, user, "en_ru", j.ID, step.ID, action, first.Token, func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step) error {
		t.Fatal("expired worker wrote")
		return nil
	}); !errors.Is(err, dailyjourney.ErrReservationLost) {
		t.Fatal(err)
	}
	second, err := NewDailyJourneyRepository(db).ReserveAI(ctx, user, "en_ru", j.ID, step.ID, action, time.Minute, prepare)
	if err != nil || second.Token == first.Token || string(second.Input) != `{"snapshot": "original"}` && string(second.Input) != `{"snapshot":"original"}` {
		t.Fatal("retry lost input/fencing", second, err)
	}
	if err := repo.FailAI(ctx, user, "en_ru", j.ID, action.ClientActionID, first.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.CompleteAI(ctx, user, "en_ru", j.ID, step.ID, action, first.Token, func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step) error {
		t.Fatal("superseded worker wrote")
		return nil
	}); !errors.Is(err, dailyjourney.ErrReservationLost) {
		t.Fatal(err)
	}
	if _, err := repo.CompleteAI(ctx, user, "en_ru", j.ID, step.ID, action, second.Token, func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step) error { return nil }); err != nil {
		t.Fatal(err)
	}
}
func TestDailyJourneyAIArchiveBarrierAndFailedWorkerRelease(t *testing.T) {
	db, repo, user, j := journeyAIRepoFixture(t, 808603)
	ctx := context.Background()
	action := dailyjourney.Action{ClientActionID: "archive", Type: "message", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{}`)}
	prepare := func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step, json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`{}`), nil
	}
	reservation, err := repo.ReserveAI(ctx, user, "en_ru", j.ID, j.Current().ID, action, time.Minute, prepare)
	if err != nil {
		t.Fatal(err)
	}
	build := func(*sql.Tx) ([]dailyjourney.Step, error) {
		return []dailyjourney.Step{{Mode: dailyjourney.Picture, Status: "active", TargetCount: 1}}, nil
	}
	tomorrow := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	if _, err := repo.Start(ctx, user, "en_ru", tomorrow, build); !errors.Is(err, dailyjourney.ErrProcessing) {
		t.Fatal("active provider was archived", err)
	}
	if err := repo.FailAI(ctx, user, "en_ru", j.ID, action.ClientActionID, reservation.Token); err != nil {
		t.Fatal(err)
	}
	retry, err := repo.ReserveAI(ctx, user, "en_ru", j.ID, j.Current().ID, action, time.Minute, prepare)
	if err != nil {
		t.Fatal("known failure couldn't retry", err)
	}
	if _, err := db.Exec(`UPDATE daily_journey_actions SET lease_expires_at=CURRENT_TIMESTAMP-INTERVAL '1 second' WHERE journey_id=?`, j.ID); err != nil {
		t.Fatal(err)
	}
	next, err := repo.Start(ctx, user, "en_ru", tomorrow, build)
	if err != nil || next.ID == j.ID {
		t.Fatal("expired provider blocked new day", err)
	}
	if _, err := repo.CompleteAI(ctx, user, "en_ru", j.ID, j.Current().ID, action, retry.Token, func(*sql.Tx, *dailyjourney.Journey, *dailyjourney.Step) error {
		t.Fatal("archived route credited")
		return nil
	}); !errors.Is(err, dailyjourney.ErrReservationLost) {
		t.Fatal(err)
	}
}
