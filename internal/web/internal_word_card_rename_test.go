package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternalWordCardRename(t *testing.T) {
	r, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	r.internalServiceTokens = map[string]string{"default": "tok"}
	r.setupProtectedRoutes()
	uc, tc := seedWordTrainingReportCard(t, r, userID)
	var id int64
	if err := r.db.QueryRow(`SELECT word_card_id FROM training_cards WHERE id=?`, tc).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`UPDATE word_cards SET course_code='es_ru', verb_forms_json='{"v1":"reportword","v2":"kept"}' WHERE id=?`, id); err != nil {
		t.Fatal(err)
	}
	call := func(token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/internal/word-cards/%d", id), bytes.NewBufferString(body))
		req.Header.Set("X-Service-Token", token)
		w := httptest.NewRecorder()
		r.mux.ServeHTTP(w, req)
		return w
	}
	base := `{"course_code":"es_ru","expected_word":"reportword","new_word":"corrected"`
	for _, c := range []struct {
		token, body string
		code        int
	}{
		{"bad", base + `}`, 401}, {"tok", `{}`, 400}, {"tok", base + `,"oops":1}`, 400}, {"tok", base + `}{}`, 400},
		{"tok", `{"course_code":"en_ru","expected_word":"reportword","new_word":"corrected"}`, 409},
	} {
		if w := call(c.token, c.body); w.Code != c.code {
			t.Fatalf("status %d expected %d: %s", w.Code, c.code, w.Body.String())
		}
	}
	if w := call("tok", base+`}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var word string
	r.db.QueryRow(`SELECT word FROM word_cards WHERE id=?`, id).Scan(&word)
	if word != "reportword" {
		t.Fatal("preview mutated word")
	}
	var duplicate int64
	if err := r.db.QueryRow(`INSERT INTO word_cards(word,definition,course_code) VALUES('corrected','','es_ru') RETURNING id`).Scan(&duplicate); err != nil {
		t.Fatal(err)
	}
	if w := call("tok", base+`,"dry_run":false}`); w.Code != 409 {
		t.Fatal("duplicate not rejected", w.Code, w.Body.String())
	}
	if _, err := r.db.Exec(`DELETE FROM word_cards WHERE id=?`, duplicate); err != nil {
		t.Fatal(err)
	}
	// Force a failure after the parent update and ensure transaction rollback.
	if _, err := r.db.Exec(`ALTER TABLE training_cards ADD CONSTRAINT rename_test_guard CHECK(word_en <> 'corrected')`); err != nil {
		t.Fatal(err)
	}
	w := call("tok", base+`,"dry_run":false}`)
	if _, err := r.db.Exec(`ALTER TABLE training_cards DROP CONSTRAINT rename_test_guard`); err != nil {
		t.Fatal(err)
	}
	if w.Code != 500 {
		t.Fatal("expected rollback failure", w.Code, w.Body.String())
	}
	r.db.QueryRow(`SELECT word FROM word_cards WHERE id=?`, id).Scan(&word)
	if word != "reportword" {
		t.Fatal("rollback failed")
	}
	// The same spelling in another course is allowed; unrelated display forms survive.
	if _, err := r.db.Exec(`INSERT INTO word_cards(word,definition,course_code) VALUES('corrected','','en_ru')`); err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`INSERT INTO word_forms(form,word_card_id) VALUES('reportword',?)`, id); err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`UPDATE training_cards SET display_word='custom display' WHERE id=?`, tc); err != nil {
		t.Fatal(err)
	}
	if w := call("tok", base+`,"transcription":"correct","dry_run":false}`); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var trainingWord, transcription, forms string
	if err := r.db.QueryRow(`SELECT wc.word,tc.word_en,tc.transcription,wc.verb_forms_json FROM word_cards wc JOIN training_cards tc ON tc.word_card_id=wc.id WHERE wc.id=? AND tc.id=?`, id, tc).Scan(&word, &trainingWord, &transcription, &forms); err != nil {
		t.Fatal(err)
	}
	if word != "corrected" || trainingWord != "corrected" || transcription != "correct" {
		t.Fatal(word, trainingWord, transcription)
	}
	var f map[string]string
	if err := json.Unmarshal([]byte(forms), &f); err != nil {
		t.Fatal(err)
	}
	if f["v1"] != "corrected" || f["v2"] != "kept" {
		t.Fatal(forms)
	}
	var aliases int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM word_forms WHERE word_card_id=? AND form='reportword'`, id).Scan(&aliases); err != nil || aliases != 0 {
		t.Fatal("old alias survived", err, aliases)
	}
	var display string
	if err := r.db.QueryRow(`SELECT display_word FROM training_cards WHERE id=?`, tc).Scan(&display); err != nil || display != "custom display" {
		t.Fatal("custom display overwritten", err, display)
	}
	var linked int64
	if err := r.db.QueryRow(`SELECT training_card_id FROM user_cards WHERE id=?`, uc).Scan(&linked); err != nil || linked != tc {
		t.Fatal("progress link changed", err, linked)
	}
	if w := call("tok", base+`,"dry_run":false}`); w.Code != 409 {
		t.Fatal("stale request accepted", w.Code)
	}
}
