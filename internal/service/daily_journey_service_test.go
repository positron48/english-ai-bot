package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"
)

type journeyTestAdapter struct {
	step                dailyjourney.Step
	candidates, applied int
	validateErr         error
}

func (a *journeyTestAdapter) Candidates(context.Context, database.ContextDBTX, int64, string) ([]dailyjourney.Step, error) {
	a.candidates++
	return []dailyjourney.Step{a.step}, nil
}
func (a *journeyTestAdapter) Validate(context.Context, database.ContextDBTX, int64, string, *dailyjourney.Step) error {
	return a.validateErr
}
func (a *journeyTestAdapter) Apply(_ context.Context, _ *sql.Tx, _ int64, _ *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) error {
	a.applied++
	if action.Type == "answer" {
		step.CompletedCount = step.TargetCount
		step.Status = "completed"
		result := "subject attempt finished"
		step.Result = &result
	}
	return nil
}
func TestDailyJourneyLaunchValidation(t *testing.T) {
	for _, tc := range []struct {
		mode   dailyjourney.Mode
		name   string
		params map[string]string
		want   bool
	}{
		{dailyjourney.Words, "Training", nil, true},
		{dailyjourney.Words, "Chat", nil, false},
		{dailyjourney.Reading, "ReadingText", map[string]string{"textId": "text-a"}, true},
		{dailyjourney.Reading, "ReadingText", nil, false},
		{dailyjourney.Reading, "ReadingText", map[string]string{"textId": "text-a", "unexpected": "extra"}, false},
		{dailyjourney.Grammar, "GrammarCategoryTest", map[string]string{"sectionId": "a0"}, true},
		{dailyjourney.Conversation, "PlaceChat", map[string]string{"districtCode": "a0", "scenarioCode": "cafe"}, true},
		{dailyjourney.Conversation, "PlaceChat", map[string]string{"scenarioCode": "cafe"}, false},
	} {
		if got := validJourneyLaunch(dailyjourney.Step{Mode: tc.mode, Launch: &dailyjourney.Launch{RouteName: tc.name, Params: tc.params}}); got != tc.want {
			t.Fatalf("mode=%s name=%s got=%v", tc.mode, tc.name, got)
		}
	}
}
func TestDailyJourneyServiceGuardsAndResume(t *testing.T) {
	db := testutil.SetupTestDB(t)
	ctx := context.Background()
	user, err := repository.NewUserRepository(db, zap.NewNop()).GetOrCreateUser(808202)
	if err != nil {
		t.Fatal(err)
	}
	words := &journeyTestAdapter{step: dailyjourney.Step{Mode: dailyjourney.Words, Title: "Words", TargetCount: 1, ContentVersion: "v1", ContentRef: json.RawMessage(`{"queue":"pinned"}`), Launch: &dailyjourney.Launch{RouteName: "Training"}}}
	reading := &journeyTestAdapter{step: dailyjourney.Step{Mode: dailyjourney.Reading, Title: "Reading", TargetCount: 1, ContentVersion: "v1", ContentRef: json.RawMessage(`{"text_id":"read-a"}`), Launch: &dailyjourney.Launch{RouteName: "ReadingText", Params: map[string]string{"textId": "read-a"}}}}
	s := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.Words: words, dailyjourney.Reading: reading})
	if s.Ready() {
		t.Fatal("partial adapter registry exposed to production")
	}
	j, err := s.Start(ctx, user.ID, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	restored, err := s.Start(ctx, user.ID, "en_ru")
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID != j.ID || words.candidates != 1 || words.applied != 0 {
		t.Fatal("resume reselected content/started exercise")
	}
	action := func(key, kind string) error {
		current := j.Current()
		if current == nil {
			t.Fatal("no current step")
		}
		result, err := s.Act(ctx, user.ID, "en_ru", j.ID, current.ID, dailyjourney.Action{ClientActionID: key, Type: kind, ExpectedRevision: j.Revision, Payload: json.RawMessage(`{}`)})
		if err == nil {
			j = result
		}
		return err
	}
	if err := action("skip-mandatory", "skip"); !errors.Is(err, dailyjourney.ErrInvalidAction) {
		t.Fatalf("mandatory skipped: %v", err)
	}
	transient := errors.New("temporary DB/provider failure")
	words.validateErr = transient
	if err := action("start-word", "start"); !errors.Is(err, transient) {
		t.Fatalf("transient swallowed: %v", err)
	}
	if j.Current().Status != "pending" || words.applied != 0 {
		t.Fatal("failure changed cursor")
	}
	words.validateErr = nil
	if err := action("start-word", "start"); err != nil {
		t.Fatal(err)
	}
	if j.Current().Status != "active" {
		t.Fatal("step not started")
	}
	if err := action("answer-word", "answer"); err != nil {
		t.Fatal(err)
	}
	reading.validateErr = ErrDailyJourneyContentUnavailable
	if err := action("start-reading", "start"); err != nil {
		t.Fatal(err)
	}
	if j.Steps[1].Status != "unavailable" || reading.applied != 0 || j.Status != "completed" {
		t.Fatalf("withdrawn content: %+v", j)
	}
	if j.Steps[1].Result == nil {
		t.Fatal("missing unavailable explanation")
	}
}
