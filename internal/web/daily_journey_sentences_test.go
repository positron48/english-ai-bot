package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"
)

func sentenceJourneyFixture(t *testing.T, course string, count int) (*dailyJourneySentencesAdapter, int64, int64, []models.SentenceItem) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	user, err := repository.NewUserRepository(db, logger).GetOrCreateUser(809601)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE users SET subscription_tier='pro' WHERE id=?`, user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.NewCourseRepository(db, logger).EnsureUserCourse(context.Background(), user.ID, course); err != nil {
		t.Fatal(err)
	}
	r := NewRouter(logger, &config.Config{SentenceComposition: config.SentenceCompositionConfig{Enabled: true}}, db, nil, nil, nil, nil)
	a := &dailyJourneySentencesAdapter{router: r, run: func(context.Context, string, string, string, string, string, ...string) (*ai.SentenceGrade, error) {
		return ai.NewExactSentenceGrade("Private reference"), nil
	}}
	items := []models.SentenceItem{}
	for i := 0; i < count; i++ {
		items = append(items, models.SentenceItem{Position: i + 1, PromptRU: fmt.Sprintf("Предложение %d", i), ReferenceES: "Private reference"})
	}
	setID, err := r.sentenceRepo().CreateSet(&models.SentenceSet{UserID: user.ID, CourseCode: course, GenerationDate: time.Now().Format("2006-01-02"), Scopes: []string{}}, items, nil)
	if err != nil {
		t.Fatal(err)
	}
	items, err = r.sentenceRepo().GetItems(setID)
	if err != nil {
		t.Fatal(err)
	}
	return a, user.ID, setID, items
}

func TestDailyJourneySentencesCapResumeAndSingleShot(t *testing.T) {
	for _, course := range []string{"es_ru", "en_ru"} {
		t.Run(course, func(t *testing.T) {
			a, user, setID, items := sentenceJourneyFixture(t, course, 13)
			db := a.router.db
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			maxConnections := db.Stats().MaxOpenConnections
			db.SetMaxOpenConns(1)
			defer db.SetMaxOpenConns(maxConnections)
			calls := 0
			fail := true
			a.run = func(ctx context.Context, code, prompt, clarification, reference, input string, _ ...string) (*ai.SentenceGrade, error) {
				calls++
				if code != course {
					t.Fatalf("wrong grading course %s", code)
				}
				var n int
				if err := db.QueryRowContext(ctx, `SELECT 1`).Scan(&n); err != nil {
					return nil, err
				}
				if fail {
					return nil, errors.New("provider temporarily down")
				}
				return ai.NewExactSentenceGrade(reference), nil
			}
			svc := service.NewDailyJourneyService(db, map[dailyjourney.Mode]service.DailyJourneyAdapter{dailyjourney.Sentences: a})
			j, err := svc.Start(ctx, user, course)
			if err != nil {
				t.Fatal(err)
			}
			if j.Current() == nil || j.Current().Mode != dailyjourney.Sentences || j.Current().TargetCount != 10 {
				t.Fatalf("missing/cap: %+v", j)
			}
			set, _ := a.router.sentenceRepo().GetSetByID(setID)
			if set.Status != models.SentenceSetReady {
				t.Fatal("candidate selection consumed the set")
			}
			act := func(kind string, payload any) dailyjourney.Action {
				raw, _ := json.Marshal(payload)
				return dailyjourney.Action{ClientActionID: fmt.Sprintf("%s-%d", kind, j.Revision), Type: kind, ExpectedRevision: j.Revision, Payload: raw}
			}
			start := act("start", map[string]any{})
			j, err = svc.Act(ctx, user, course, j.ID, j.Current().ID, start)
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(j)
			if strings.Contains(string(raw), "Private reference") {
				t.Fatal("reference leaked before answer")
			}
			stepID := j.Current().ID
			bad := act("answer", map[string]any{"item_id": items[12].ID, "user_input": "answer"})
			if _, err := svc.Act(ctx, user, course, j.ID, stepID, bad); !errors.Is(err, dailyjourney.ErrInvalidAction) {
				t.Fatalf("unbound item accepted: %v", err)
			}
			for i := 0; i < 10; i++ {
				text := "Private reference"
				if i == 0 {
					text = "A different answer"
				}
				action := act("answer", map[string]any{"item_id": items[i].ID, "user_input": text})
				if i == 0 {
					if _, err := svc.Act(ctx, user, course, j.ID, stepID, action); !errors.Is(err, dailyjourney.ErrAIProvider) {
						t.Fatalf("provider failure: %v", err)
					}
					item, _, _ := a.router.sentenceRepo().GetItemWithSet(items[i].ID)
					if item.AttemptedAt != nil {
						t.Fatal("failed grading consumed item")
					}
					fail = false
				}
				j, err = svc.Act(ctx, user, course, j.ID, stepID, action)
				if err != nil {
					t.Fatal(err)
				}
				if j.Current() == nil || j.Current().Status != "active" || j.Current().CompletedCount != i+1 {
					t.Fatal("feedback skipped/progress wrong")
				}
				replay, err := svc.Act(ctx, user, course, j.ID, stepID, action)
				if err != nil || replay.Revision != j.Revision {
					t.Fatalf("replay failed: %v", err)
				}
				restored, err := svc.Exercise(ctx, user, course, j.ID, stepID)
				if err != nil {
					t.Fatal(err)
				}
				public, _ := json.Marshal(restored.Data)
				if !strings.Contains(string(public), `"grading"`) {
					t.Fatal("feedback lost on refresh")
				}
				j, err = svc.Act(ctx, user, course, j.ID, stepID, act("ack_feedback", map[string]any{}))
				if err != nil {
					t.Fatal(err)
				}
			}
			if j.Status != "completed" || calls != 2 {
				t.Fatalf("completion/provider calls: %s %d", j.Status, calls)
			}
			set, _ = a.router.sentenceRepo().GetSetByID(setID)
			if set.Status != models.SentenceSetStarted || set.StarCount != 10 {
				t.Fatalf("set accounting: %+v", set)
			}
			var n int
			if err := db.QueryRow(`SELECT SUM(attempt_count) FROM mode_daily_stats WHERE mode='sentences'`).Scan(&n); err != nil || n != 10 {
				t.Fatalf("duplicate/missing stats: %d %v", n, err)
			}
			candidates, err := a.Candidates(ctx, db, user, course)
			if err != nil || len(candidates) != 1 || candidates[0].TargetCount != 3 {
				t.Fatalf("remaining set: %+v %v", candidates, err)
			}
			if _, err := db.Exec(`UPDATE users SET subscription_tier='free' WHERE id=?`, user); err != nil {
				t.Fatal(err)
			}
			candidates, err = a.Candidates(ctx, db, user, course)
			if err != nil || len(candidates) != 0 {
				t.Fatal("free user received sentence candidate", err)
			}
		})
	}
}

func TestDailyJourneySentencesAtomicRollbackAndSharedAnswer(t *testing.T) {
	a, user, setID, items := sentenceJourneyFixture(t, "es_ru", 1)
	ctx := context.Background()
	db := a.router.db
	svc := service.NewDailyJourneyService(db, map[dailyjourney.Mode]service.DailyJourneyAdapter{dailyjourney.Sentences: a})
	j, err := svc.Start(ctx, user, "es_ru")
	if err != nil {
		t.Fatal(err)
	}
	id := j.Current().ID
	j, err = svc.Act(ctx, user, "es_ru", j.ID, id, dailyjourney.Action{ClientActionID: "start", Type: "start", ExpectedRevision: j.Revision, Payload: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"item_id": items[0].ID, "user_input": "Private reference"})
	action := dailyjourney.Action{ClientActionID: "answer", Type: "answer", ExpectedRevision: j.Revision, Payload: raw}
	repo := repository.NewDailyJourneyRepository(db)
	reservation, err := repo.ReserveAI(ctx, user, "es_ru", j.ID, id, action, time.Minute, func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step, _ json.RawMessage) (json.RawMessage, error) {
		return a.PrepareAI(ctx, tx, user, j, s, action)
	})
	if err != nil {
		t.Fatal(err)
	}
	grade, err := a.RunAI(ctx, reservation.Input)
	if err != nil {
		t.Fatal(err)
	}
	forced := errors.New("rollback")
	_, err = repo.CompleteAI(ctx, user, "es_ru", j.ID, id, action, reservation.Token, func(tx *sql.Tx, j *dailyjourney.Journey, s *dailyjourney.Step) error {
		if err := a.ApplyAI(ctx, tx, user, j, s, action, grade); err != nil {
			return err
		}
		return forced
	})
	if !errors.Is(err, forced) {
		t.Fatal(err)
	}
	if err := repo.FailAI(ctx, user, "es_ru", j.ID, action.ClientActionID, reservation.Token); err != nil {
		t.Fatal(err)
	}
	item, _, _ := a.router.sentenceRepo().GetItemWithSet(items[0].ID)
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM mode_daily_stats`).Scan(&n); err != nil || n != 0 || item.AttemptedAt != nil {
		t.Fatal("attempt/stats survived rollback", n, err)
	}
	// A direct submission made while the journey was open is authoritative.
	if _, err := a.router.sentenceRepo().RecordAttempt(items[0].ID, "Direct answer", 0, "star", `{"outcome":"star","error_count":0,"corrected_es":"Direct answer"}`); err != nil {
		t.Fatal(err)
	}
	j, err = svc.Act(ctx, user, "es_ru", j.ID, id, action)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(j.Exercise.Data), "Direct answer") {
		t.Fatal("direct attempt overwritten")
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM mode_daily_stats`).Scan(&n); err != nil || n != 0 {
		t.Fatal("shared attempt credited twice", err)
	}
	set, _ := a.router.sentenceRepo().GetSetByID(setID)
	if set.StarCount != 1 {
		t.Fatal("duplicate counter")
	}
}
