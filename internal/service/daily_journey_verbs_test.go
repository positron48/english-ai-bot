package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"
)

func journeyVerbUser(t *testing.T, db *sql.DB, telegram int64) int64 {
	t.Helper()
	user := journeySubjectUser(t, db, telegram)
	if _, err := repository.NewCourseRepository(db, zap.NewNop()).EnsureUserCourse(context.Background(), user, "es_ru"); err != nil {
		t.Fatal(err)
	}
	return user
}
func journeyVerbSeed(t *testing.T, db *sql.DB, user int64, index int, tense string) int64 {
	t.Helper()
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	lemma := fmt.Sprintf("journeyviajar%d", index)
	surface := fmt.Sprintf("viajó%d", index)
	var word int64
	if err := db.QueryRow(`INSERT INTO word_cards(word,definition,pos,course_code) VALUES(?,'{}','verb','es_ru') RETURNING id`, lemma).Scan(&word); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO user_word_knowledge(user_id,word_card_id,status) VALUES(?,?,'known')`, user, word); err != nil {
		t.Fatal(err)
	}
	lemmaID, err := repo.UpsertVerbLemma(lemma, "es", "test", "1", "hash", "{}")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.LinkWordCardToLemma(word, lemmaID, 1, "test"); err != nil {
		t.Fatal(err)
	}
	form, err := repo.UpsertVerbForm(&models.VerbFormDict{VerbLemmaID: lemmaID, Mood: "indicativo", Tense: tense, Person: "3", Number: "singular", SurfaceForm: surface})
	if err != nil {
		t.Fatal(err)
	}
	prompt, _ := json.Marshal(map[string]interface{}{"content_version": 4, "practice_eligible": true, "lemma": lemma, "question": "Ayer ____.", "example_translation": "Вчера он путешествовал.", "person": "3", "number": "singular", "tense": tense, "mood": "indicativo", "rule": map[string]interface{}{"id": fmt.Sprint("test-rule", index), "regular": false}})
	answer, _ := json.Marshal(map[string]string{"surface_form": surface})
	id, err := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{WordCardID: word, VerbFormDictID: form, CardType: models.VerbCardTypeCloze, PromptJSON: string(prompt), AnswerJSON: string(answer), DistractorsJSON: `[]`})
	if err != nil {
		t.Fatal(err)
	}
	return id
}
func journeyVerbAct(t *testing.T, s *DailyJourneyService, user int64, j **dailyjourney.Journey, key, kind string, payload interface{}) {
	t.Helper()
	raw, _ := json.Marshal(payload)
	step := (*j).Current()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	next, err := s.Act(ctx, user, "es_ru", (*j).ID, step.ID, dailyjourney.Action{ClientActionID: key, Type: kind, ExpectedRevision: (*j).Revision, Payload: raw})
	if err != nil {
		t.Fatalf("%s: %v", key, err)
	}
	*j = next
}
func TestDailyJourneyVerbsAtomicIndependentPractice(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := journeyVerbUser(t, db, 808501)
	for i := 0; i < 12; i++ {
		journeyVerbSeed(t, db, user, i, "presente")
	}
	closedID := journeyVerbSeed(t, db, user, 12, "futuro")
	cfg := config.TrainingConfig{SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 20, VerbFormsMaxNew: 20}
	adapter := NewDailyJourneyVerbsAdapter(db, cfg, config.DefaultLearningConfig(), zap.NewNop())
	candidates, err := adapter.Candidates(context.Background(), db, user, "en_ru")
	if err != nil || len(candidates) != 0 {
		t.Fatal("EN verbs offered", err)
	}
	candidates, err = adapter.Candidates(context.Background(), db, user, "es_ru")
	if err != nil || len(candidates) != 1 || candidates[0].TargetCount != 10 {
		t.Fatalf("ready queue %+v %v", candidates, err)
	}
	var ref journeyVerbRef
	_ = json.Unmarshal(candidates[0].ContentRef, &ref)
	for _, id := range ref.Cards {
		if id == closedID {
			t.Fatal("locked tense offered")
		}
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM user_verb_cards WHERE user_id=?`, user).Scan(&n); err != nil || n != 0 {
		t.Fatalf("candidate wrote cards: %d %v", n, err)
	}
	var normalID int64
	if err := db.QueryRow(`INSERT INTO verb_training_sessions(user_id,session_json,planned_count) VALUES(?,'{"independent":true}',30) RETURNING id`, user).Scan(&normalID); err != nil {
		t.Fatal(err)
	}
	oldMax := db.Stats().MaxOpenConnections
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.SetMaxOpenConns(oldMax) })
	svc := NewDailyJourneyService(db, map[dailyjourney.Mode]DailyJourneyAdapter{dailyjourney.VerbForms: adapter})
	j, err := svc.Start(context.Background(), user, "es_ru")
	if err != nil {
		t.Fatal(err)
	}
	journeyVerbAct(t, svc, user, &j, "start", "start", map[string]interface{}{})
	var state repository.VerbPractice
	_ = json.Unmarshal(j.Current().State, &state)
	if len(state.Queue) != 10 || state.ID >= 0 {
		t.Fatal("wrong isolated state")
	}
	view, err := adapter.Exercise(context.Background(), user, j, j.Current())
	if err != nil {
		t.Fatal(err)
	}
	public := view.(map[string]interface{})
	prompt := public["prompt"].(map[string]interface{})
	if prompt["rule"] != nil || prompt["tense"] != nil || public["forms"] != nil {
		t.Fatal("help leaked before recording assistance")
	}
	current := state.Queue[0]
	journeyVerbAct(t, svc, user, &j, "help", "start", map[string]interface{}{"help": true, "index": 0, "card_id": current.UserVerbCardID})
	_ = json.Unmarshal(j.Current().State, &state)
	if !state.Assisted {
		t.Fatal("assistance wasn't saved")
	}
	view, err = adapter.Exercise(context.Background(), user, j, j.Current())
	if err != nil || len(view.(map[string]interface{})["forms"].([]map[string]interface{})) == 0 {
		t.Fatal("missing form help", err)
	}
	var answer map[string]string
	_ = json.Unmarshal([]byte(current.AnswerJSON), &answer)
	payload, _ := json.Marshal(map[string]interface{}{"index": 0, "card_id": current.UserVerbCardID, "answer": answer["surface_form"]})
	action := dailyjourney.Action{ClientActionID: "answer0", ExpectedRevision: j.Revision, Type: "answer", Payload: payload}
	step := j.Current()
	forced := errors.New("after verb SRS failure")
	_, err = svc.repo.Apply(context.Background(), user, "es_ru", j.ID, step.ID, action, func(tx *sql.Tx, j *dailyjourney.Journey, step *dailyjourney.Step) error {
		if err := adapter.Apply(context.Background(), tx, user, j, step, action); err != nil {
			return err
		}
		return forced
	})
	if !errors.Is(err, forced) {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM verb_review_events WHERE user_id=?`, user).Scan(&n); err != nil || n != 0 {
		t.Fatalf("review rollback %d %v", n, err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, e := svc.Act(ctx, user, "es_ru", j.ID, step.ID, action)
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	j, err = svc.Get(context.Background(), user, "es_ru", j.ID)
	if err != nil {
		t.Fatal(err)
	}
	_ = json.Unmarshal(j.Current().State, &state)
	if state.Feedback == nil || !state.Feedback.Assisted || state.Feedback.Outcome != "correct" {
		t.Fatal("assisted feedback lost")
	}
	if err := db.QueryRow(`SELECT reps FROM user_verb_cards WHERE id=?`, current.UserVerbCardID).Scan(&n); err != nil || n != 0 {
		t.Fatal("assistance increased mastery", n, err)
	}
	journeyVerbAct(t, svc, user, &j, "ack0", "ack_feedback", map[string]interface{}{})
	for i := 1; i < 10; i++ {
		_ = json.Unmarshal(j.Current().State, &state)
		card := state.Queue[state.Index]
		journeyVerbAct(t, svc, user, &j, fmt.Sprint("answer", i), "answer", map[string]interface{}{"index": i, "card_id": card.UserVerbCardID, "skip": true})
		journeyVerbAct(t, svc, user, &j, fmt.Sprint("ack", i), "ack_feedback", map[string]interface{}{})
	}
	if j.Status != "completed" {
		t.Fatal(j.Status)
	}
	for _, table := range []string{"verb_review_events", "exercise_attempts", "learning_events"} {
		query := "SELECT COUNT(*) FROM " + table + " WHERE user_course_id IN (SELECT id FROM user_courses WHERE user_id=?)"
		if table == "verb_review_events" {
			query = "SELECT COUNT(*) FROM verb_review_events WHERE user_id=? AND session_id IS NULL"
		}
		if err := db.QueryRow(query, user).Scan(&n); err != nil || n != 10 {
			t.Fatalf("%s %d %v", table, n, err)
		}
	}
	var raw string
	var ended sql.NullTime
	if err := db.QueryRow(`SELECT session_json,ended_at,done_count FROM verb_training_sessions WHERE id=?`, normalID).Scan(&raw, &ended, &n); err != nil || raw != `{"independent":true}` || ended.Valid || n != 0 {
		t.Fatalf("normal session modified %s %+v %d %v", raw, ended, n, err)
	}
}
func TestDailyJourneyVerbsContentAndFeatureGuards(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := journeyVerbUser(t, db, 808502)
	id := journeyVerbSeed(t, db, user, 30, "presente")
	cfg := config.TrainingConfig{SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 10, VerbFormsMaxNew: 10}
	adapter := NewDailyJourneyVerbsAdapter(db, cfg, config.DefaultLearningConfig(), zap.NewNop())
	candidates, err := adapter.Candidates(context.Background(), db, user, "es_ru")
	if err != nil || len(candidates) != 1 {
		t.Fatal(candidates, err)
	}
	if _, err := db.Exec(`UPDATE verb_training_cards SET prompt_json=jsonb_set(prompt_json::jsonb,'{practice_eligible}','false')::text WHERE id=?`, id); err != nil {
		t.Fatal(err)
	}
	if err := adapter.Validate(context.Background(), db, user, "es_ru", &candidates[0]); !errors.Is(err, ErrDailyJourneyContentUnavailable) {
		t.Fatal("withdrawn content accepted", err)
	}
	adapter.cfg.SpanishVerbFormsEnabled = false
	candidates, err = adapter.Candidates(context.Background(), db, user, "es_ru")
	if err != nil || len(candidates) != 0 {
		t.Fatal("disabled verbs offered", err)
	}
}

func TestDailyJourneyVerbsRevokedGrammarScope(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := journeyVerbUser(t, db, 808503)
	journeyVerbSeed(t, db, user, 31, "futuro")
	cfg := config.TrainingConfig{SpanishVerbFormsEnabled: true, VerbFormsMaxCards: 10, VerbFormsMaxNew: 10}
	adapter := NewDailyJourneyVerbsAdapter(db, cfg, config.DefaultLearningConfig(), zap.NewNop())
	candidates, err := adapter.Candidates(context.Background(), db, user, "es_ru")
	if err != nil || len(candidates) != 0 {
		t.Fatal("future unlocked without grammar", candidates, err)
	}
	chapter := "es.grammar.future_conditionals.futuro_simple_forms_core_uses"
	attempts := repository.NewGrammarAttemptRepository(db, zap.NewNop())
	if err := attempts.UpdateProgress(user, chapter, 50, true); err != nil {
		t.Fatal(err)
	}
	candidates, err = adapter.Candidates(context.Background(), db, user, "es_ru")
	if err != nil || len(candidates) != 1 {
		t.Fatal("passed grammar didn't unlock future", candidates, err)
	}
	if _, err := db.Exec(`UPDATE grammar_progress SET passed_at=NULL WHERE user_id=? AND chapter_id=?`, user, chapter); err != nil {
		t.Fatal(err)
	}
	if err := adapter.Validate(context.Background(), db, user, "es_ru", &candidates[0]); !errors.Is(err, ErrDailyJourneyContentUnavailable) {
		t.Fatal("revoked tense allowed", err)
	}
}
