package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

func TestOfflineSyncRollsBackWhenHistoryFails(t *testing.T) {
	r, db, uid, cards, _ := setupTrainingOfflineTest(t)
	id := seedUserCardForOfflineSync(t, db, cards, repository.NewTrainingCardRepository(db, r.logger), uid)
	before, _ := cards.GetUserCard(id)
	_, err := db.Exec(`CREATE FUNCTION audit_reject_review() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'audit injected failure'; END $$`)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DROP FUNCTION audit_reject_review() CASCADE`)
	_, err = db.Exec(`CREATE TRIGGER audit_review_failure BEFORE INSERT ON review_events FOR EACH ROW EXECUTE FUNCTION audit_reject_review()`)
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]interface{}{"attempts": []map[string]interface{}{{"client_attempt_id": "audit-retry", "user_card_id": id, "chosen_option": "offline", "correct_answer": "offline", "options": []string{"offline", "wrong"}, "answer_time_ms": 1000}}}
	raw, _ := json.Marshal(payload)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/training/offline/sync-attempts", bytes.NewReader(raw))
		req = req.WithContext(context.WithValue(req.Context(), userIDKey, uid))
		w := httptest.NewRecorder()
		r.handleTrainingOfflineSyncAttempts(w, req)
		if !bytes.Contains(w.Body.Bytes(), []byte("review_event_create_failed")) {
			t.Fatal(w.Body.String())
		}
		after, _ := cards.GetUserCard(id)
		t.Logf("sync=%s reps_before=%d reps_after=%d interval_after=%d", w.Body.String(), before.Reps, after.Reps, after.IntervalDays)
		if after.Reps != before.Reps || after.IntervalDays != before.IntervalDays {
			t.Fatalf("failed sync changed SRS: %+v", after)
		}
	}
}
func TestOfflineSyncGradesCanonicalAnswer(t *testing.T) {
	for _, mode := range []string{"card", "spell", "type"} {
		t.Run(mode, func(t *testing.T) {
			r, db, uid, cards, _ := setupTrainingOfflineTest(t)
			id := seedUserCardForOfflineSync(t, db, cards, repository.NewTrainingCardRepository(db, r.logger), uid)
			var sessionID int64
			for i, answer := range []string{"totally unrelated", "offline"} {
				event, err := r.syncOfflineWordAttempt(httptest.NewRequest("POST", "/", nil), uid, &sessionID, 2, offlineWordTrainingAttempt{
					ClientAttemptID: mode + answer, UserCardID: id, Mode: mode,
					ChosenOption: answer, AnswerText: answer, CorrectAnswer: "totally unrelated",
				})
				if err != nil {
					t.Fatal(err)
				}
				if event == nil || event.IsCorrect != (i == 1) {
					t.Fatalf("canonical grading of %q: %+v", answer, event)
				}
			}
		})
	}
}

func TestOfflineSyncDuplicateDoesNotGradeTwice(t *testing.T) {
	r, db, uid, cards, _ := setupTrainingOfflineTest(t)
	id := seedUserCardForOfflineSync(t, db, cards, repository.NewTrainingCardRepository(db, r.logger), uid)
	var sid int64
	req := httptest.NewRequest("POST", "/", nil)
	attempt := offlineWordTrainingAttempt{ClientAttemptID: "once", UserCardID: id, ChosenOption: "offline"}
	for i := 0; i < 2; i++ {
		event, err := r.syncOfflineWordAttempt(req, uid, &sid, 1, attempt)
		if err != nil {
			t.Fatal(err)
		}
		if (event == nil) != (i == 1) {
			t.Fatalf("unexpected duplicate result: %+v", event)
		}
	}
	card, _ := cards.GetUserCard(id)
	if card.Reps != 1 {
		t.Fatalf("reps=%d", card.Reps)
	}
}

func TestOfflinePackUsesRequestedLanguage(t *testing.T) {
	r, _, uid, _, _ := setupTrainingOfflineTest(t)
	req := httptest.NewRequest("GET", "/api/training/offline/pack?course_code=es_ru", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, uid))
	w := httptest.NewRecorder()
	r.handleTrainingOfflinePack(w, req)
	var pack map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &pack); err != nil {
		t.Fatal(err)
	}
	if pack["target_lang"] != "es" {
		t.Fatalf("wrong pack language: %s", w.Body.String())
	}
	card := &models.UserCardWithTraining{UserCard: models.UserCard{Direction: models.DirectionRUtoEN}, TrainingCard: models.TrainingCard{WordEN: "libro", WordRU: "книга"}}
	item := r.buildOfflineWordTrainingCard(req, "ru", card, []string{"libro"}, "libro")
	if !strings.Contains(item.Question, "испанский") {
		t.Fatalf("wrong question: %s", item.Question)
	}
}

func TestOfflineSyncConcurrentDuplicate(t *testing.T) {
	r, db, uid, cards, _ := setupTrainingOfflineTest(t)
	id := seedUserCardForOfflineSync(t, db, cards, repository.NewTrainingCardRepository(db, r.logger), uid)
	var wg sync.WaitGroup
	var committed atomic.Int32
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var sessionID int64
			event, err := r.syncOfflineWordAttempt(httptest.NewRequest("POST", "/", nil), uid, &sessionID, 1, offlineWordTrainingAttempt{ClientAttemptID: "concurrent", UserCardID: id, ChosenOption: "offline"})
			if err != nil {
				t.Error(err)
				return
			}
			if event != nil {
				committed.Add(1)
			}
		}()
	}
	wg.Wait()
	card, err := cards.GetUserCard(id)
	if err != nil {
		t.Fatal(err)
	}
	if committed.Load() != 1 || card.Reps != 1 {
		t.Fatalf("committed=%d reps=%d", committed.Load(), card.Reps)
	}
}
