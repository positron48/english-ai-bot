package repository

import (
	"sync"
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestGrammarAnswerAtomicAndIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	logger := zap.NewNop()
	user, err := NewUserRepository(db, logger).GetOrCreateUser(992200)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewGrammarSRSRepository(db, logger)
	_, err = db.Exec(`CREATE FUNCTION reject_grammar_attempt() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected'; END $$`)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DROP FUNCTION reject_grammar_attempt() CASCADE`)
	_, err = db.Exec(`CREATE TRIGGER reject_attempt BEFORE INSERT ON grammar_attempts FOR EACH ROW EXECUTE FUNCTION reject_grammar_attempt()`)
	if err != nil {
		t.Fatal(err)
	}
	record := func() (int64, error) {
		return repo.RecordAnswer(user.ID, "es", "es", "ch", "block", "concept", "q", "a", "a", true, "once", time.Now())
	}
	if _, err = record(); err == nil {
		t.Fatal("must return persistence failure")
	}
	var n int
	if err = db.QueryRow(`SELECT count(*) FROM grammar_theory_memory WHERE user_id = ?`, user.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("failed answer left an SRS update")
	}
	if _, err = db.Exec(`DROP TRIGGER reject_attempt ON grammar_attempts`); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if id, err := record(); err != nil || id == 0 {
				t.Errorf("record: %d %v", id, err)
			}
		}()
	}
	wg.Wait()
	if err = db.QueryRow(`SELECT review_count FROM grammar_theory_memory WHERE user_id = ?`, user.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("duplicate review count=%d", n)
	}
	if err = db.QueryRow(`SELECT count(*) FROM grammar_attempts WHERE user_id = ?`, user.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("duplicate attempts=%d", n)
	}
}
