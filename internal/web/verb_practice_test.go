package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

func TestVerbPractice_DurableFeedbackHelpRetryAndOwnership(t *testing.T) {
	r, db, user := setupVerbFormsHandlerTest(t)
	word := seedVerbFormsHandlerVocab(t, db, user)
	// A core lemma is available even with no personal vocabulary.
	if _, err := db.Exec(`DELETE FROM user_cards WHERE user_id=$1`, user); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	var lemma int64
	if err := db.QueryRow(`SELECT verb_lemma_id FROM word_verb_lemmas WHERE word_card_id=$1`, word).Scan(&lemma); err != nil {
		t.Fatal(err)
	}
	for _, form := range []models.VerbFormDict{{VerbLemmaID: lemma, Mood: "indicativo", Tense: "presente", Person: "2", Number: "singular", SurfaceForm: "hablas"}, {VerbLemmaID: lemma, Mood: "indicativo", Tense: "pretérito", Person: "1", Number: "singular", SurfaceForm: "hablé"}} {
		if _, err := repo.UpsertVerbForm(&form); err != nil {
			t.Fatal(err)
		}
	}
	// Even a previously-created future/past card must stay out of the queue.
	if err := r.newVerbTrainingServiceForUser(context.Background(), user).EnsureVerbFormUserCards(user, []string{"es.pretérito.indicativo"}); err != nil {
		t.Fatal(err)
	}
	call := func(action string, body map[string]interface{}, uid int64) (int, map[string]interface{}) {
		t.Helper()
		raw, _ := json.Marshal(body)
		method := "POST"
		if action == "current" {
			method = "GET"
		}
		req := verbFormsUserContext(httptest.NewRequest(method, "/api/verb-training/v2/"+action, bytes.NewReader(raw)), uid)
		rr := httptest.NewRecorder()
		r.handleVerbPractice(rr, req)
		result := map[string]interface{}{}
		_ = json.Unmarshal(rr.Body.Bytes(), &result)
		return rr.Code, result
	}
	type response struct {
		code  int
		state map[string]interface{}
	}
	starts := make(chan response, 2)
	for i := 0; i < 2; i++ {
		go func() { code, state := call("start", map[string]interface{}{}, user); starts <- response{code, state} }()
	}
	first, second := <-starts, <-starts
	code, state := first.code, first.state
	if second.code != 200 || first.state["session_id"] != second.state["session_id"] {
		t.Fatalf("concurrent starts: %+v %+v", first, second)
	}
	if code != 200 {
		t.Fatalf("start %d %v", code, state)
	}
	if state["total_cards"] != float64(2) {
		t.Fatalf("only present core cards expected: %v", state)
	}
	prompt := state["prompt"].(map[string]interface{})
	if prompt["expected_form"] != nil || prompt["example_translation"] != nil || prompt["rule"] != nil {
		t.Fatalf("answer leaked: %v", prompt)
	}
	sessionID := state["session_id"]
	for index := 0; index < 2; index++ {
		body := map[string]interface{}{"session_id": sessionID, "card_id": state["card_id"]}
		if code, _ = call("answer", body, user+999); code == 200 {
			t.Fatal("accepted another user's session")
		}
		if index == 0 {
			code, state = call("help", body, user)
			if code != 200 || state["assisted"] != true {
				t.Fatalf("help: %v", state)
			}
		}
		body["skip"] = true
		answers := make(chan response, 2)
		for i := 0; i < 2; i++ {
			go func() { code, state := call("answer", body, user); answers <- response{code, state} }()
		}
		first, second = <-answers, <-answers
		code, state = first.code, first.state
		if second.code != 200 || len(second.state["results"].([]interface{})) != index+1 {
			t.Fatal("concurrent duplicate answer changed results")
		}
		if code != 200 || state["feedback"] == nil {
			t.Fatalf("answer: %d %v", code, state)
		}
		code, again := call("answer", body, user)
		if code != 200 || len(again["results"].([]interface{})) != index+1 {
			t.Fatal("duplicate answer counted twice")
		}
		_, restored := call("current", nil, user)
		if restored["feedback"] == nil {
			t.Fatal("feedback not restored")
		}
		code, state = call("advance", body, user)
		if code != 200 {
			t.Fatalf("advance %v", state)
		}
		_, again = call("advance", body, user)
		if again["card_index"] != state["card_index"] {
			t.Fatal("duplicate advance moved twice")
		}
	}
	if state["completed"] != true {
		t.Fatal("not completed")
	}
	var reviews int
	if err := db.QueryRow(`SELECT COUNT(*) FROM verb_review_events WHERE user_id=$1`, user).Scan(&reviews); err != nil || reviews != 2 {
		t.Fatalf("reviews=%d err=%v", reviews, err)
	}
	code, state = call("repeat", map[string]interface{}{"session_id": sessionID}, user)
	if code != 200 || state["retry"] != true {
		t.Fatalf("repeat %v", state)
	}
	var before, after int
	id := state["card_id"]
	_ = db.QueryRow(`SELECT lapse_count FROM user_verb_cards WHERE id=$1`, id).Scan(&before)
	code, _ = call("answer", map[string]interface{}{"session_id": state["session_id"], "card_id": id, "skip": true}, user)
	_ = db.QueryRow(`SELECT lapse_count FROM user_verb_cards WHERE id=$1`, id).Scan(&after)
	if code != 200 || before != after {
		t.Fatalf("retry changed SRS: %d -> %d", before, after)
	}
}

func TestVerbPractice_RuleMasteryRequiresDifferentVerbs(t *testing.T) {
	r, db, user := setupVerbFormsHandlerTest(t)
	seedVerbFormsHandlerVocab(t, db, user)
	if err := r.newVerbTrainingServiceForUser(context.Background(), user).EnsureVerbFormUserCards(user, []string{"es.presente.indicativo"}); err != nil {
		t.Fatal(err)
	}
	var cardID int64
	if err := db.QueryRow(`SELECT id FROM user_verb_cards WHERE user_id=$1 LIMIT 1`, user).Scan(&cardID); err != nil {
		t.Fatal(err)
	}
	repo := repository.NewVerbFormsRepository(db, zap.NewNop())
	insert := func(lemma, outcome string, retry bool) {
		raw, _ := json.Marshal(map[string]interface{}{"version": 2, "retry": retry, "feedback": map[string]interface{}{"lemma": lemma, "outcome": outcome, "assisted": false, "rule": map[string]interface{}{"id": "regular.ar.2s", "regular": true}}})
		if _, err := db.Exec(`INSERT INTO verb_review_events(user_id,user_verb_card_id,answered_at,is_correct,quality,metrics_json) VALUES($1,$2,CURRENT_TIMESTAMP,1,5,$3)`, user, cardID, string(raw)); err != nil {
			t.Fatal(err)
		}
	}
	insert("hablar", "correct", false)
	insert("hablar", "correct", false)
	insert("hablar", "correct", false)
	rules, err := repo.MasteredVerbRules(user)
	if err != nil || rules["regular.ar.2s"] {
		t.Fatal("one verb is not transfer", err, rules)
	}
	insert("tomar", "correct", true)
	insert("trabajar", "correct", true)
	rules, err = repo.MasteredVerbRules(user)
	if err != nil || rules["regular.ar.2s"] {
		t.Fatal("immediate retry must not prove transfer", err, rules)
	}
	insert("tomar", "correct", false)
	insert("trabajar", "correct", false)
	rules, err = repo.MasteredVerbRules(user)
	if err != nil || !rules["regular.ar.2s"] {
		t.Fatal("three verbs should prove transfer", err, rules)
	}
	insert("hablar", "incorrect", false)
	rules, err = repo.MasteredVerbRules(user)
	if err != nil || rules["regular.ar.2s"] {
		t.Fatal("new mistake must reopen practice", err, rules)
	}
}
