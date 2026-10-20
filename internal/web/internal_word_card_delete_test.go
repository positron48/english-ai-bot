package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/repository"
)

func TestInternalWordCardDelete(t *testing.T) {
	r, _, userID, cleanup := setupGrammarTest(t)
	defer cleanup()
	r.internalServiceTokens = map[string]string{"default": "tok"}
	r.setupProtectedRoutes()
	uc, tc := seedWordTrainingReportCard(t, r, userID)
	var wordID int64
	if err := r.db.QueryRow(`SELECT word_card_id FROM training_cards WHERE id=?`, tc).Scan(&wordID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`UPDATE word_cards SET course_code='es_ru' WHERE id=?`, wordID); err != nil {
		t.Fatal(err)
	}
	reports := repository.NewContentReportRepository(r.db, r.logger)
	reportID, err := reports.Create(repository.CreateContentReportInput{UserID: userID, SourceType: "word_training", Word: "reportword", WordCardID: &wordID, TrainingCardID: &tc, UserCardID: &uc, ReportCategory: "other", Payload: map[string]interface{}{"word": "reportword"}})
	if err != nil {
		t.Fatal(err)
	}
	// Same spelling in another course must survive the deletion.
	if _, err := r.db.Exec(`INSERT INTO word_cards(word,definition,course_code) VALUES ('reportword','keep','en_ru')`); err != nil {
		t.Fatal(err)
	}
	call := func(path, token, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodDelete, path, bytes.NewBufferString(body))
		req.Header.Set("X-Service-Token", token)
		w := httptest.NewRecorder()
		r.mux.ServeHTTP(w, req)
		return w
	}
	path := fmt.Sprintf("/api/internal/word-cards/%d", wordID)
	body := `{"course_code":"es_ru","expected_word":"reportword"}`
	for _, tc := range []struct {
		name, path, token, body string
		status                  int
	}{
		{"unauthorized", path, "wrong", body, 401},
		{"invalid id", "/api/internal/word-cards/1/extra", "tok", body, 400},
		{"missing identity", path, "tok", `{"dry_run":false}`, 400},
		{"unknown field", path, "tok", `{"course_code":"es_ru","expected_word":"reportword","dryrun":false}`, 400},
		{"multiple objects", path, "tok", body + body, 400},
		{"wrong course", path, "tok", `{"course_code":"en_ru","expected_word":"reportword","dry_run":false}`, 409},
		{"wrong spelling", path, "tok", `{"course_code":"es_ru","expected_word":"changed","dry_run":false}`, 409},
		{"missing word", "/api/internal/word-cards/999999999", "tok", body, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := call(tc.path, tc.token, tc.body)
			if w.Code != tc.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
		})
	}
	w := call(path, "tok", body)
	if w.Code != 200 {
		t.Fatalf("preview %d %s", w.Code, w.Body.String())
	}
	var result repository.WordCardDeletion
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.DryRun || result.Deleted || result.TrainingCards != 1 || result.UserCards != 1 {
		t.Fatalf("unexpected preview: %+v", result)
	}
	// Verify the exact production write shape with EXPLAIN on the test database.
	rows, err := r.db.Query(`EXPLAIN DELETE FROM word_cards WHERE id = ? AND course_code = ? AND word = ?`, wordID, "es_ru", "reportword")
	if err != nil {
		t.Fatal(err)
	}
	rows.Close()
	// A downstream failure must leave both parent and dependent cards intact.
	if _, err := r.db.Exec(`CREATE TABLE complaint_delete_guard (word_card_id BIGINT REFERENCES word_cards(id) ON DELETE RESTRICT)`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.db.Exec(`DROP TABLE IF EXISTS complaint_delete_guard`) })
	if _, err := r.db.Exec(`INSERT INTO complaint_delete_guard VALUES (?)`, wordID); err != nil {
		t.Fatal(err)
	}
	deleting := `{"course_code":"es_ru","expected_word":"reportword","dry_run":false}`
	w = call(path, "tok", deleting)
	if w.Code != 500 {
		t.Fatalf("expected FK failure, got %d %s", w.Code, w.Body.String())
	}
	var remaining int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM user_cards WHERE id=?`, uc).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatalf("failed deletion changed user cards: count=%d err=%v", remaining, err)
	}
	if _, err := r.db.Exec(`DROP TABLE complaint_delete_guard`); err != nil {
		t.Fatal(err)
	}
	w = call(path, "tok", deleting)
	if w.Code != 200 {
		t.Fatalf("delete %d %s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Deleted || result.DryRun {
		t.Fatalf("not deleted: %+v", result)
	}
	for _, query := range []string{`SELECT COUNT(*) FROM word_cards WHERE id=?`, `SELECT COUNT(*) FROM training_cards WHERE word_card_id=?`, `SELECT COUNT(*) FROM user_cards WHERE training_card_id IN (SELECT id FROM training_cards WHERE word_card_id=?)`} {
		if err := r.db.QueryRow(query, wordID).Scan(&remaining); err != nil || remaining != 0 {
			t.Fatalf("remaining rows: count=%d err=%v", remaining, err)
		}
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM user_cards WHERE id=?`, uc).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("user card survived: count=%d err=%v", remaining, err)
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM word_cards WHERE word='reportword' AND course_code='en_ru'`).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatalf("other course changed: count=%d err=%v", remaining, err)
	}
	report, err := reports.GetByID(reportID)
	if err != nil || report == nil {
		t.Fatalf("report lost: %v", err)
	}
	if report.WordCardID != nil || report.TrainingCardID != nil || report.UserCardID != nil || report.PayloadJSON != `{"word":"reportword"}` {
		t.Fatalf("report history not retained: %+v", report)
	}
	w = call(path, "tok", deleting)
	if w.Code != 404 {
		t.Fatalf("repeat delete %d", w.Code)
	}
}
