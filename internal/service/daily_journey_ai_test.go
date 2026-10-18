package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/testutil"
)

type journeyAITestAdapter struct {
	*journeyTestAdapter
	db                   *sql.DB
	calls                atomic.Int32
	prepared             atomic.Int32
	runErr, errorOnApply error
	entered, release     chan struct{}
	lastInput            json.RawMessage
}

func (a *journeyAITestAdapter) RequiresAI(action dailyjourney.Action) bool {
	return action.Type == "message"
}
func (a *journeyAITestAdapter) PrepareAI(_ context.Context, _ *sql.Tx, _ int64, _ *dailyjourney.Journey, _ *dailyjourney.Step, _ dailyjourney.Action) (json.RawMessage, error) {
	a.prepared.Add(1)
	return json.RawMessage(`{"prompt":"private prompt","text":"Hello"}`), nil
}
func (a *journeyAITestAdapter) RunAI(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	a.calls.Add(1)
	a.lastInput = input
	// A pool with one connection would deadlock if the reservation transaction
	// were still open during provider I/O.
	var n int
	if err := a.db.QueryRowContext(ctx, `SELECT 1`).Scan(&n); err != nil {
		return nil, err
	}
	if a.entered != nil {
		a.entered <- struct{}{}
		select {
		case <-a.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if a.runErr != nil {
		return nil, a.runErr
	}
	return json.RawMessage(`{"reply":"Hi"}`), nil
}
func (a *journeyAITestAdapter) ApplyAI(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action, result json.RawMessage) error {
	if _, err := tx.ExecContext(ctx, `UPDATE users SET settings_json='{"ai_applied":true}' WHERE id=?`, user); err != nil {
		return err
	}
	step.CompletedCount = 1
	step.Feedback = new(string)
	*step.Feedback = "Hi"
	return a.errorOnApply
}
func journeyAIFake(db *sql.DB) *journeyAITestAdapter {
	return &journeyAITestAdapter{journeyTestAdapter: &journeyTestAdapter{step: dailyjourney.Step{Mode: dailyjourney.Words, Title: "AI test", TargetCount: 1, ContentVersion: "v1", ContentRef: json.RawMessage(`{}`), Launch: &dailyjourney.Launch{RouteName: "Training"}}}, db: db}
}
func TestDailyJourneyAIServiceRetriesAndNoTransactionDuringProvider(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := journeySubjectUser(t, db, 808611)
	adapter := journeyAIFake(db)
	svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Words: adapter})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	j, err := svc.Start(ctx, user, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	journeySubjectAct(t, svc, user, &j, "start", "start", map[string]interface{}{})
	oldMax := db.Stats().MaxOpenConnections
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.SetMaxOpenConns(oldMax) })
	stepID := j.Current().ID
	action := dailyjourney.Action{ClientActionID: "message", Type: "message", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{"text":"Hello"}`)}
	providerErr := errors.New("private provider detail")
	adapter.runErr = providerErr
	if _, err := svc.Act(ctx, user, "en_ru", j.ID, stepID, action); !errors.Is(err, dailyjourney.ErrAIProvider) || !errors.Is(err, providerErr) {
		t.Fatal("provider cause lost", err)
	}
	var status string
	if err := db.QueryRow(`SELECT status FROM daily_journey_actions WHERE journey_id=? AND client_action_id='message'`, j.ID).Scan(&status); err != nil || status != "retryable" {
		t.Fatal("provider failure stranded lease", status, err)
	}
	adapter.runErr = nil
	adapter.errorOnApply = errors.New("subject write failure")
	if _, err := svc.Act(ctx, user, "en_ru", j.ID, stepID, action); !errors.Is(err, adapter.errorOnApply) {
		t.Fatal(err)
	}
	var raw string
	if err := db.QueryRow(`SELECT settings_json FROM users WHERE id=?`, user).Scan(&raw); err != nil || raw == `{"ai_applied":true}` {
		t.Fatal("subject rollback failed", raw, err)
	}
	if err := db.QueryRow(`SELECT status FROM daily_journey_actions WHERE journey_id=? AND client_action_id='message'`, j.ID).Scan(&status); err != nil || status != "retryable" {
		t.Fatal("apply failure stranded lease", status, err)
	}
	adapter.errorOnApply = nil
	restored := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Words: adapter})
	result, err := restored.Act(ctx, user, "en_ru", j.ID, stepID, action)
	if err != nil || result.Revision != j.Revision+1 || result.Current().CompletedCount != 1 {
		t.Fatal(result, err)
	}
	replay, err := restored.Act(ctx, user, "en_ru", j.ID, stepID, action)
	if err != nil || replay.Revision != result.Revision || adapter.calls.Load() != 3 || adapter.prepared.Load() != 1 {
		t.Fatal("provider re-executed completed receipt", adapter.calls.Load(), adapter.prepared.Load(), err)
	}
	serialized, _ := json.Marshal(replay)
	if strings.Contains(string(serialized), "private prompt") || strings.Contains(string(serialized), "reservation_token") {
		t.Fatal("private AI input leaked")
	}

}
func TestDailyJourneyAIServiceProcessingDoesNotBlockReads(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := journeySubjectUser(t, db, 808612)
	adapter := journeyAIFake(db)
	adapter.entered = make(chan struct{}, 1)
	adapter.release = make(chan struct{})
	svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Words: adapter})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	j, err := svc.Start(ctx, user, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	journeySubjectAct(t, svc, user, &j, "start", "start", map[string]interface{}{})
	oldMax := db.Stats().MaxOpenConnections
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.SetMaxOpenConns(oldMax) })
	action := dailyjourney.Action{ClientActionID: "inflight", Type: "message", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{}`)}
	done := make(chan error, 1)
	go func() { _, err := svc.Act(ctx, user, "en_ru", j.ID, j.Current().ID, action); done <- err }()
	select {
	case <-adapter.entered:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if _, err := svc.Get(ctx, user, "en_ru", j.ID); err != nil {
		t.Fatal("GET blocked behind provider", err)
	}
	if _, err := svc.Act(ctx, user, "en_ru", j.ID, j.Current().ID, action); !errors.Is(err, dailyjourney.ErrProcessing) {
		t.Fatal("second worker ran", err)
	}
	// Access is revalidated after the provider; a revoked item earns no credit.
	adapter.validateErr = ErrDailyJourneyContentUnavailable
	close(adapter.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	current, err := svc.Get(ctx, user, "en_ru", j.ID)
	if err != nil || current.Steps[0].Status != "unavailable" || current.Steps[0].CompletedCount != 0 {
		t.Fatal("revoked content credited", current, err)
	}
	if adapter.calls.Load() != 1 {
		t.Fatal("duplicate provider call")
	}
}
