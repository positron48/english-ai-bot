package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/verbtraining"
)

type verbFormsTestDriverCfg struct {
	Columns  []string
	Values   []driver.Value
	CloseErr error
}

var (
	verbFormsTestDriverMu        sync.Mutex
	verbFormsTestDriverState     verbFormsTestDriverCfg
	verbFormsTestDriverResponses []verbFormsTestDriverCfg
	verbFormsTestDriverCall      int
)

func init() {
	sql.Register("verbforms_test_driver", &verbFormsTestDriver{})
}

type verbFormsTestDriver struct{}

func (verbFormsTestDriver) Open(string) (driver.Conn, error) {
	return &verbFormsTestConn{}, nil
}

type verbFormsTestConn struct{}

func (c *verbFormsTestConn) Prepare(string) (driver.Stmt, error) { return &verbFormsTestStmt{}, nil }
func (c *verbFormsTestConn) Close() error                        { return nil }
func (c *verbFormsTestConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }
func (c *verbFormsTestConn) ResetSession(context.Context) error  { return driver.ErrSkip }
func (c *verbFormsTestConn) IsValid() bool                       { return true }

type verbFormsTestStmt struct{}

func (verbFormsTestStmt) Close() error { return nil }
func (verbFormsTestStmt) NumInput() int {
	return -1
}
func (verbFormsTestStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}
func (verbFormsTestStmt) Query([]driver.Value) (driver.Rows, error) {
	verbFormsTestDriverMu.Lock()
	defer verbFormsTestDriverMu.Unlock()
	cfg := verbFormsTestDriverState
	if len(verbFormsTestDriverResponses) > 0 {
		if verbFormsTestDriverCall >= len(verbFormsTestDriverResponses) {
			cfg = verbFormsTestDriverResponses[len(verbFormsTestDriverResponses)-1]
		} else {
			cfg = verbFormsTestDriverResponses[verbFormsTestDriverCall]
		}
		verbFormsTestDriverCall++
	}
	return &verbFormsTestRows{
		columns:  append([]string(nil), cfg.Columns...),
		values:   append([]driver.Value(nil), cfg.Values...),
		closeErr: cfg.CloseErr,
	}, nil
}

type verbFormsTestRows struct {
	columns  []string
	values   []driver.Value
	closeErr error
	done     bool
}

func (r *verbFormsTestRows) Columns() []string { return r.columns }
func (r *verbFormsTestRows) Close() error      { return r.closeErr }
func (r *verbFormsTestRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	if len(r.values) == 0 {
		r.done = true
		return io.EOF
	}
	r.done = true
	for i := range dest {
		if i < len(r.values) {
			dest[i] = r.values[i]
		}
	}
	return nil
}

func setVerbFormsTestDriverCfg(t *testing.T, cfg verbFormsTestDriverCfg) *sql.DB {
	t.Helper()
	verbFormsTestDriverMu.Lock()
	verbFormsTestDriverState = cfg
	verbFormsTestDriverResponses = nil
	verbFormsTestDriverCall = 0
	verbFormsTestDriverMu.Unlock()
	db, err := sql.Open("verbforms_test_driver", "")
	if err != nil {
		t.Fatalf("open test driver db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func newVerbFormsBadScanDB(t *testing.T, cols []string, vals []driver.Value) *sql.DB {
	t.Helper()
	return setVerbFormsTestDriverCfg(t, verbFormsTestDriverCfg{Columns: cols, Values: vals})
}

const verbFormsMoreUserTelegramID = 900001

type verbFormsMoreSeed struct {
	userID         int64
	wordCardID     int64
	verbLemmaID    int64
	formPresenteID int64
	formFuturoID   int64
}

func seedVerbFormsMoreFixtures(t *testing.T, db *sql.DB) verbFormsMoreSeed {
	t.Helper()
	var out verbFormsMoreSeed
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES ($1) RETURNING id`, verbFormsMoreUserTelegramID).Scan(&out.userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ('hablar', 'говорить', 'verb') RETURNING id`).
		Scan(&out.wordCardID); err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	var trainingCardID int64
	if err := db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES ($1, 'hablar', 0, 'говорить', 'to speak', 'verb') RETURNING id`, out.wordCardID).Scan(&trainingCardID); err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'es_ru', 'review')`,
		out.userID, trainingCardID); err != nil {
		t.Fatalf("insert user_card: %v", err)
	}

	repo := NewVerbFormsRepository(db, zap.NewNop())
	var err error
	out.verbLemmaID, err = repo.UpsertVerbLemma("  HABLAR  ", " ES ", "test", "v1", "chk1", `{"ru":{"gloss":"говорить"}}`)
	if err != nil {
		t.Fatalf("UpsertVerbLemma: %v", err)
	}
	out.formPresenteID, err = repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: out.verbLemmaID, Mood: "Indicativo", Tense: "Presente",
		Person: "1", Number: "Singular", SurfaceForm: "HABLO", IsIrregular: true, TagsJSON: `["regular"]`,
	})
	if err != nil {
		t.Fatalf("UpsertVerbForm presente: %v", err)
	}
	out.formFuturoID, err = repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: out.verbLemmaID, Mood: "indicativo", Tense: "futuro_simple",
		Person: "1", Number: "singular", SurfaceForm: "hablaré", IsIrregular: false,
	})
	if err != nil {
		t.Fatalf("UpsertVerbForm futuro: %v", err)
	}
	if err := repo.LinkWordCardToLemma(out.wordCardID, out.verbLemmaID, 1.0, "test"); err != nil {
		t.Fatalf("LinkWordCardToLemma: %v", err)
	}
	return out
}

func TestVerbForms_UpsertLemmaMetadataPreserveOnEmpty(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	id1, err := repo.UpsertVerbLemma("hablar", "es", "src-a", "v1", "c1", `{"ru":{"gloss":"old"}}`)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	id2, err := repo.UpsertVerbLemma("hablar", "es", "src-b", "v2", "c2", "")
	if err != nil {
		t.Fatalf("empty metadata upsert: %v", err)
	}
	if id1 != id2 {
		t.Fatalf("expected same lemma id %d got %d", id1, id2)
	}
	var meta string
	if err := db.QueryRow(`SELECT metadata_json FROM verb_lemmas WHERE id = $1`, id2).Scan(&meta); err != nil {
		t.Fatalf("select metadata: %v", err)
	}
	if meta == "" || meta == "{}" {
		t.Fatalf("empty upsert should preserve metadata, got %q", meta)
	}

	_, err = repo.UpsertVerbLemma("hablar", "es", "src-c", "v3", "c3", `{}`)
	if err != nil {
		t.Fatalf("{} metadata upsert: %v", err)
	}
	if err := db.QueryRow(`SELECT metadata_json FROM verb_lemmas WHERE id = $1`, id2).Scan(&meta); err != nil {
		t.Fatalf("select metadata after {}: %v", err)
	}
	if meta == "" || meta == "{}" {
		t.Fatalf("{} upsert should preserve metadata, got %q", meta)
	}

	_, err = repo.UpsertVerbLemma("hablar", "es", "src-d", "v4", "c4", `{"ru":{"gloss":"new"}}`)
	if err != nil {
		t.Fatalf("non-empty metadata upsert: %v", err)
	}
	if err := db.QueryRow(`SELECT metadata_json FROM verb_lemmas WHERE id = $1`, id2).Scan(&meta); err != nil {
		t.Fatalf("select metadata after replace: %v", err)
	}
	if meta != `{"ru":{"gloss":"new"}}` {
		t.Fatalf("expected replaced metadata, got %q", meta)
	}
	_ = seed
}

func TestVerbForms_ListVerbFormsForLemma(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	if _, err := repo.ListVerbFormsForLemma("  ", "es"); err == nil {
		t.Fatal("expected error for empty lemma")
	}

	rows, err := repo.ListVerbFormsForLemma("hablar", "")
	if err != nil {
		t.Fatalf("ListVerbFormsForLemma default lang: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 forms, got %d", len(rows))
	}
	if rows[0].Mood != "indicativo" || rows[0].SurfaceForm != "hablo" {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}

	missing, err := repo.ListVerbFormsForLemma("noexiste", "es")
	if err != nil {
		t.Fatalf("missing lemma: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("expected empty slice, got %d", len(missing))
	}
	_ = seed
}

func TestVerbForms_ListVerbFormViewRowsForLemma(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	if _, err := repo.ListVerbFormViewRowsForLemma("", "es"); err == nil {
		t.Fatal("expected error for empty lemma")
	}

	rows, err := repo.ListVerbFormViewRowsForLemma("hablar", "")
	if err != nil {
		t.Fatalf("ListVerbFormViewRowsForLemma: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if !rows[0].IsIrregular {
		t.Fatalf("presente form should be irregular: %+v", rows[0])
	}
	if rows[1].IsIrregular {
		t.Fatalf("futuro form should not be irregular: %+v", rows[1])
	}
	_ = seed
}

func TestVerbForms_GetUserVerbForms(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	forms, err := repo.GetUserVerbForms(seed.userID, seed.wordCardID)
	if err != nil {
		t.Fatalf("GetUserVerbForms: %v", err)
	}
	if len(forms) != 2 {
		t.Fatalf("expected 2 forms, got %d", len(forms))
	}
	if forms[0].Lemma != "hablar" {
		t.Fatalf("unexpected lemma: %+v", forms[0])
	}

	other, err := repo.GetUserVerbForms(seed.userID, seed.wordCardID+9999)
	if err != nil {
		t.Fatalf("GetUserVerbForms missing card: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("expected no forms for unknown card, got %d", len(other))
	}
}

func TestVerbForms_LinkWordCardByLemma(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	_ = seedVerbFormsMoreFixtures(t, db)

	var orphanCardID int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ('comer', 'есть', 'verb') RETURNING id`).
		Scan(&orphanCardID); err != nil {
		t.Fatalf("insert orphan card: %v", err)
	}

	linked, err := repo.LinkWordCardByLemma(orphanCardID, "comer", "es", "manual")
	if err != nil {
		t.Fatalf("LinkWordCardByLemma missing lemma: %v", err)
	}
	if linked {
		t.Fatal("expected false when lemma missing")
	}

	lemmaID, err := repo.UpsertVerbLemma("comer", "es", "test", "v1", "c", `{}`)
	if err != nil {
		t.Fatalf("UpsertVerbLemma comer: %v", err)
	}
	linked, err = repo.LinkWordCardByLemma(orphanCardID, "  COMER ", " ES ", "manual")
	if err != nil || !linked {
		t.Fatalf("LinkWordCardByLemma success: linked=%v err=%v", linked, err)
	}
	var gotLemmaID int64
	if err := db.QueryRow(`SELECT verb_lemma_id FROM word_verb_lemmas WHERE word_card_id = $1`, orphanCardID).Scan(&gotLemmaID); err != nil {
		t.Fatalf("select link: %v", err)
	}
	if gotLemmaID != lemmaID {
		t.Fatalf("lemma id %d want %d", gotLemmaID, lemmaID)
	}
}

func TestVerbForms_LinkMissingSpanishVerbLemmasForUser(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	if err := repo.LinkMissingSpanishVerbLemmasForUser(seed.userID); err != nil {
		t.Fatalf("LinkMissingSpanishVerbLemmasForUser user_cards path: %v", err)
	}

	var linked int
	if err := db.QueryRow(`SELECT COUNT(*) FROM word_verb_lemmas WHERE word_card_id = $1`, seed.wordCardID).Scan(&linked); err != nil {
		t.Fatalf("count links: %v", err)
	}
	if linked != 1 {
		t.Fatalf("expected link for hablar card, got %d", linked)
	}

	var userID2 int64
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (900002) RETURNING id`).Scan(&userID2); err != nil {
		t.Fatalf("insert user2: %v", err)
	}
	var cardViaKnown int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ('vivir', 'жить', 'verb') RETURNING id`).
		Scan(&cardViaKnown); err != nil {
		t.Fatalf("insert vivir card: %v", err)
	}
	if _, err := repo.UpsertVerbLemma("vivir", "es", "test", "v1", "c", `{}`); err != nil {
		t.Fatalf("UpsertVerbLemma vivir: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES ($1, $2, 'known')`,
		userID2, cardViaKnown); err != nil {
		t.Fatalf("insert user_word_knowledge: %v", err)
	}
	if err := repo.LinkMissingSpanishVerbLemmasForUser(userID2); err != nil {
		t.Fatalf("LinkMissingSpanishVerbLemmasForUser known path: %v", err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM word_verb_lemmas WHERE word_card_id = $1`, cardViaKnown).Scan(&linked); err != nil {
		t.Fatalf("count vivir link: %v", err)
	}
	if linked != 1 {
		t.Fatalf("expected vivir linked via known vocab, got %d", linked)
	}

	var emptyWordCard int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ('   ', 'x', 'verb') RETURNING id`).
		Scan(&emptyWordCard); err != nil {
		t.Fatalf("insert blank word card: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES ($1, $2, 'known')`,
		userID2, emptyWordCard); err != nil {
		t.Fatalf("insert known blank word: %v", err)
	}
	if err := repo.LinkMissingSpanishVerbLemmasForUser(userID2); err != nil {
		t.Fatalf("LinkMissingSpanishVerbLemmasForUser blank word skip: %v", err)
	}
}

func TestVerbForms_UpsertVerbFormExampleAndExamples(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	id1, err := repo.UpsertVerbFormExample(&models.VerbFormExample{
		VerbFormDictID: seed.formPresenteID,
		ExampleTarget:  "Yo hablo español.",
		GlossNative:    "Я говорю по-испански.",
		Source:         "test",
		QualityScore:   10,
	})
	if err != nil || id1 <= 0 {
		t.Fatalf("insert example: id=%d err=%v", id1, err)
	}
	id2, err := repo.UpsertVerbFormExample(&models.VerbFormExample{
		VerbFormDictID: seed.formPresenteID,
		ExampleTarget:  "Yo hablo español.",
		GlossNative:    "Обновлено.",
		Source:         "test2",
		QualityScore:   99,
	})
	if err != nil || id2 != id1 {
		t.Fatalf("upsert update: id1=%d id2=%d err=%v", id1, id2, err)
	}

	examples, err := repo.GetVerbFormExamples(seed.formPresenteID, 5)
	if err != nil {
		t.Fatalf("GetVerbFormExamples: %v", err)
	}
	if len(examples) != 1 || examples[0].QualityScore != 99 {
		t.Fatalf("unexpected examples: %+v", examples)
	}
}

func TestVerbForms_TrainingCardsQueueAndSRS(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	prompt := map[string]interface{}{
		"type":                models.VerbCardTypeCloze,
		"lemma":               "hablar",
		"example_translation": "Я говорю.",
	}
	promptJSON, _ := json.Marshal(prompt)

	trainingCardID, err := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{
		WordCardID:      seed.wordCardID,
		VerbFormDictID:  seed.formPresenteID,
		CardType:        models.VerbCardTypeCloze,
		PromptJSON:      string(promptJSON),
		AnswerJSON:      `{"surface_form":"hablo"}`,
		DistractorsJSON: `["hablas"]`,
	})
	if err != nil || trainingCardID <= 0 {
		t.Fatalf("UpsertVerbTrainingCard: id=%d err=%v", trainingCardID, err)
	}

	if err := repo.EnsureUserCardsForUserWords(seed.userID, nil); err != nil {
		t.Fatalf("EnsureUserCardsForUserWords empty scopes: %v", err)
	}
	if err := repo.EnsureUserCardsForUserWords(seed.userID, []string{"es.presente.indicativo"}); err != nil {
		t.Fatalf("EnsureUserCardsForUserWords: %v", err)
	}

	uvcID, err := repo.GetOrCreateUserVerbCard(seed.userID, trainingCardID)
	if err != nil || uvcID <= 0 {
		t.Fatalf("GetOrCreateUserVerbCard create: id=%d err=%v", uvcID, err)
	}
	sameID, err := repo.GetOrCreateUserVerbCard(seed.userID, trainingCardID)
	if err != nil || sameID != uvcID {
		t.Fatalf("GetOrCreateUserVerbCard idempotent: %d vs %d err=%v", sameID, uvcID, err)
	}

	missingSRS, err := repo.GetVerbUserCardSRS(uvcID + 99999)
	if err != nil || missingSRS != nil {
		t.Fatalf("GetVerbUserCardSRS missing: %+v err=%v", missingSRS, err)
	}

	srs, err := repo.GetVerbUserCardSRS(uvcID)
	if err != nil || srs == nil {
		t.Fatalf("GetVerbUserCardSRS: %+v err=%v", srs, err)
	}
	srs.State = "review"
	srs.Reps = 1
	nextDue := time.Now().Add(24 * time.Hour)
	if err := repo.UpdateVerbUserCardSRS(srs, nextDue, 4); err != nil {
		t.Fatalf("UpdateVerbUserCardSRS: %v", err)
	}

	sessionID, err := repo.StartVerbSession(seed.userID, 2, `{}`)
	if err != nil || sessionID <= 0 {
		t.Fatalf("StartVerbSession: id=%d err=%v", sessionID, err)
	}
	if err := repo.CreateVerbReviewEvent(sessionID, seed.userID, uvcID, true, 5); err != nil {
		t.Fatalf("CreateVerbReviewEvent correct: %v", err)
	}
	if err := repo.CreateVerbReviewEvent(sessionID, seed.userID, uvcID, false, 1); err != nil {
		t.Fatalf("CreateVerbReviewEvent incorrect: %v", err)
	}
	if err := repo.FinishVerbSession(sessionID, 2); err != nil {
		t.Fatalf("FinishVerbSession: %v", err)
	}
	total, correct, err := repo.GetVerbSessionStats(sessionID)
	if err != nil || total != 2 || correct != 1 {
		t.Fatalf("GetVerbSessionStats total=%d correct=%d err=%v", total, correct, err)
	}

	_, _ = db.Exec(`UPDATE user_verb_cards SET state='new', next_due_at=NULL WHERE id=$1`, uvcID)
	queue, err := repo.GetVerbQueue(seed.userID, time.Now(), 1, -1)
	if err != nil {
		t.Fatalf("GetVerbQueue negative maxNew: %v", err)
	}
	if len(queue) > 1 {
		t.Fatalf("expected maxCards=1 truncation, got %d", len(queue))
	}

	count, err := repo.CountUserVerbClozeCards(seed.userID)
	if err != nil || count != 1 {
		t.Fatalf("CountUserVerbClozeCards: count=%d err=%v", count, err)
	}
}

func TestVerbForms_roundRobinEdgeCases(t *testing.T) {
	if got := roundRobinVerbNewCards(nil, 5); got != nil {
		t.Fatalf("nil pool: %+v", got)
	}
	if got := roundRobinVerbNewCards([]verbNewQueueRow{{card: VerbQueueCard{UserVerbCardID: 1}, wordCardID: 1}}, 0); got != nil {
		t.Fatalf("maxPick 0: %+v", got)
	}
}

func TestVerbForms_LemmaMetadataBatchAndList(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	_ = seedVerbFormsMoreFixtures(t, db)

	empty, err := repo.GetVerbLemmaMetadataJSONBatch(nil)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty batch: %+v err=%v", empty, err)
	}
	blank, err := repo.GetVerbLemmaMetadataJSONBatch([]string{"", "  ", "hablar", "hablar"})
	if err != nil {
		t.Fatalf("dedupe batch: %v", err)
	}
	if len(blank) != 1 {
		t.Fatalf("expected one lemma in batch map, got %d", len(blank))
	}
	if blank["hablar"] == "" {
		t.Fatal("expected metadata for hablar")
	}

	if err := repo.UpdateVerbLemmaMetadataJSON("", `{}`); err == nil {
		t.Fatal("expected error for empty lemma")
	}
	if err := repo.UpdateVerbLemmaMetadataJSON("hablar", `{"ru":{"gloss":"обновлено"}}`); err != nil {
		t.Fatalf("UpdateVerbLemmaMetadataJSON: %v", err)
	}

	lemmas, err := repo.ListSpanishVerbLemmas()
	if err != nil {
		t.Fatalf("ListSpanishVerbLemmas: %v", err)
	}
	if len(lemmas) == 0 {
		t.Fatal("expected at least one Spanish lemma")
	}
}

func TestVerbForms_ListPendingVerbTrainingLemmas(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	got, err := repo.ListPendingVerbTrainingLemmas(0, 0, true)
	if err != nil {
		t.Fatalf("ListPendingVerbTrainingLemmas default limit: %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("expected pending lemma for hablar, got %+v", got)
	}

	all, err := repo.ListPendingVerbTrainingLemmas(5000, seed.wordCardID-1, false)
	if err != nil {
		t.Fatalf("ListPendingVerbTrainingLemmas formsGapOnly=false: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected lemma with formsGapOnly=false")
	}
}

func TestVerbForms_AdminEndpoints(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	prompt := map[string]interface{}{
		"type":                models.VerbCardTypeCloze,
		"example_translation": "Я говорю.",
	}
	promptJSON, _ := json.Marshal(prompt)
	_, err := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{
		WordCardID:     seed.wordCardID,
		VerbFormDictID: seed.formPresenteID,
		CardType:       models.VerbCardTypeCloze,
		PromptJSON:     string(promptJSON),
		AnswerJSON:     `{"surface_form":"hablo"}`,
	})
	if err != nil {
		t.Fatalf("UpsertVerbTrainingCard admin: %v", err)
	}

	adminList, err := repo.ListAdminVerbTrainingLemmas("hab", 0, 0)
	if err != nil {
		t.Fatalf("ListAdminVerbTrainingLemmas: %v", err)
	}
	if len(adminList) != 1 || adminList[0].Lemma != "hablar" {
		t.Fatalf("unexpected admin list: %+v", adminList)
	}
	if adminList[0].RuGloss == "" {
		t.Fatal("expected RuGloss from metadata")
	}

	cards, err := repo.ListAdminVerbTrainingCardsByWordCard(seed.wordCardID)
	if err != nil || len(cards) != 1 {
		t.Fatalf("ListAdminVerbTrainingCardsByWordCard: len=%d err=%v", len(cards), err)
	}
	if len(cards[0].Distractors) != 0 {
		t.Fatalf("expected no distractors json, got %s", cards[0].Distractors)
	}

	if _, err := repo.ListAdminVerbTrainingCardsByWordCard(0); err == nil {
		t.Fatal("expected error for invalid word_card_id")
	}

	lemma, ok, err := repo.AdminVerbLemmaLookup(seed.wordCardID)
	if err != nil || !ok || lemma != "hablar" {
		t.Fatalf("AdminVerbLemmaLookup found: lemma=%q ok=%v err=%v", lemma, ok, err)
	}
	lemma, ok, err = repo.AdminVerbLemmaLookup(0)
	if err != nil || ok || lemma != "" {
		t.Fatalf("AdminVerbLemmaLookup zero id: lemma=%q ok=%v err=%v", lemma, ok, err)
	}
	lemma, ok, err = repo.AdminVerbLemmaLookup(seed.wordCardID + 99999)
	if err != nil || ok {
		t.Fatalf("AdminVerbLemmaLookup missing: lemma=%q ok=%v err=%v", lemma, ok, err)
	}

	emptyPromptCard, err := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{
		WordCardID:     seed.wordCardID,
		VerbFormDictID: seed.formFuturoID,
		CardType:       models.VerbCardTypeCloze,
		PromptJSON:     "  ",
		AnswerJSON:     "",
		DistractorsJSON: `["a","b"]`,
	})
	if err != nil {
		t.Fatalf("UpsertVerbTrainingCard empty prompt: %v", err)
	}
	_ = emptyPromptCard
	allCards, err := repo.ListAdminVerbTrainingCardsByWordCard(seed.wordCardID)
	if err != nil || len(allCards) < 2 {
		t.Fatalf("ListAdminVerbTrainingCardsByWordCard all: len=%d err=%v", len(allCards), err)
	}
	for _, c := range allCards {
		if len(c.Prompt) == 0 || string(c.Prompt) == "null" {
			t.Fatalf("expected normalized prompt, got %s", c.Prompt)
		}
		if len(c.Answer) == 0 {
			t.Fatalf("expected normalized answer, got %s", c.Answer)
		}
	}
}

func TestVerbForms_GetLinkedVerbFormsForUser(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	rows, err := repo.GetLinkedVerbFormsForUser(seed.userID, []string{"es.presente.indicativo"})
	if err != nil {
		t.Fatalf("GetLinkedVerbFormsForUser: %v", err)
	}
	if len(rows) != 1 || rows[0].VerbFormDictID != seed.formPresenteID {
		t.Fatalf("unexpected rows: %+v", rows)
	}
	if nilRows, err := repo.GetLinkedVerbFormsForUser(seed.userID, nil); err != nil || nilRows != nil {
		t.Fatalf("empty scopes: rows=%v err=%v", nilRows, err)
	}
}

func TestVerbForms_ListVerbExampleCatalogTemplatesCached(t *testing.T) {
	ResetVerbExampleCatalogCacheForTests()
	repo, db := setupVerbFormsRepo(t)
	t.Cleanup(func() { ResetVerbExampleCatalogCacheForTests() })

	if _, err := db.Exec(`INSERT INTO verb_example_templates (code, lemma_match, verb_class, mood, tense, es_suffix, ru_pattern, sort_order, active)
		VALUES ('more_tpl', 'hablar', 'ar', 'indicativo', 'presente', ' en casa.', ' дома.', 1, true)`); err != nil {
		t.Fatalf("insert template: %v", err)
	}
	tpl, err := repo.ListVerbExampleCatalogTemplatesCached()
	if err != nil || len(tpl) == 0 {
		t.Fatalf("ListVerbExampleCatalogTemplatesCached: %+v err=%v", tpl, err)
	}
	_, _ = repo.ListVerbExampleCatalogTemplatesCached()
	ResetVerbExampleCatalogCacheForTests()
}

func TestVerbForms_LinkWordCardByLemma_lookupError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())

	mock.ExpectQuery(`SELECT id FROM verb_lemmas WHERE lemma`).
		WithArgs("fallar", "es").
		WillReturnError(sql.ErrConnDone)

	_, err = repo.LinkWordCardByLemma(1, "fallar", "es", "test")
	if err == nil {
		t.Fatal("expected lookup error")
	}
}

func TestVerbForms_LinkMissingSpanishVerbLemmasForUser_queryError(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())

	mock.ExpectQuery(`SELECT DISTINCT wc\.id`).WillReturnError(sql.ErrConnDone)
	if err := repo.LinkMissingSpanishVerbLemmasForUser(900001); err == nil {
		t.Fatal("expected query error")
	}
}

func TestVerbForms_ListPendingVerbTrainingLemmas_mock(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())

	mock.ExpectQuery(`SELECT w\.id, LOWER\(TRIM\(w\.word\)\)`).
		WithArgs(int64(5), models.VerbCardTypeCloze, int64(verbtraining.FullCoverageClozeCardCountV1()), 200).
		WillReturnRows(sqlmock.NewRows([]string{"id", "lemma"}).
			AddRow(int64(0), "").
			AddRow(int64(10), "hablar"))

	got, err := repo.ListPendingVerbTrainingLemmas(-1, 5, true)
	if err != nil {
		t.Fatalf("ListPendingVerbTrainingLemmas: %v", err)
	}
	if len(got) != 1 || got[0].WordCardID != 10 {
		t.Fatalf("expected skipped invalid row, got %+v", got)
	}
}

func TestVerbForms_GetVerbQueue_mockDueAndNew(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())
	now := time.Now()

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(900001), models.VerbCardTypeCloze, 30).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))

	queue, err := repo.GetVerbQueue(900001, now, 5, 1)
	if err != nil {
		t.Fatalf("GetVerbQueue: %v", err)
	}
	if len(queue) != 0 {
		t.Fatalf("expected empty queue, got %d", len(queue))
	}
}

func TestVerbForms_roundRobinMixedLemmas(t *testing.T) {
	pool := []verbNewQueueRow{
		{card: VerbQueueCard{UserVerbCardID: 1}, wordCardID: 10},
		{card: VerbQueueCard{UserVerbCardID: 2}, wordCardID: 10},
		{card: VerbQueueCard{UserVerbCardID: 3}, wordCardID: 10},
		{card: VerbQueueCard{UserVerbCardID: 11}, wordCardID: 20},
		{card: VerbQueueCard{UserVerbCardID: 12}, wordCardID: 20},
		{card: VerbQueueCard{UserVerbCardID: 13}, wordCardID: 20},
	}
	got := roundRobinVerbNewCards(pool, 4)
	if len(got) != 4 {
		t.Fatalf("len=%d", len(got))
	}
	want := []int64{1, 11, 2, 12}
	for i := range want {
		if got[i].UserVerbCardID != want[i] {
			t.Fatalf("i=%d got=%d want=%d", i, got[i].UserVerbCardID, want[i])
		}
	}
}

func TestVerbForms_GetVerbQueue_dueBeforeNewRoundRobin(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())
	now := time.Now()

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
			AddRow(int64(101), int64(10), models.VerbCardTypeCloze, "{}", `{"surface_form":"hablo"}`, "[]"))

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(900001), models.VerbCardTypeCloze, 600).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
			AddRow(int64(101), int64(10), models.VerbCardTypeCloze, "{}", `{"surface_form":"hablo"}`, "[]").
			AddRow(int64(201), int64(20), models.VerbCardTypeCloze, "{}", `{"surface_form":"como"}`, "[]"))

	queue, err := repo.GetVerbQueue(900001, now, 30, 30)
	if err != nil {
		t.Fatalf("GetVerbQueue: %v", err)
	}
	if len(queue) < 2 || queue[0].UserVerbCardID != 101 {
		t.Fatalf("unexpected queue: %+v", queue)
	}
}

func TestVerbForms_GetVerbQueue_newPoolSkipsDuplicateDue(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())
	now := time.Now()

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
			AddRow(int64(500), int64(55), models.VerbCardTypeCloze, "{}", `{"surface_form":"voy"}`, "[]"))

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(900001), models.VerbCardTypeCloze, 60).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
			AddRow(int64(500), int64(55), models.VerbCardTypeCloze, "{}", `{"surface_form":"voy"}`, "[]").
			AddRow(int64(501), int64(55), models.VerbCardTypeCloze, "{}", `{"surface_form":"vas"}`, "[]"))

	queue, err := repo.GetVerbQueue(900001, now, 10, 2)
	if err != nil {
		t.Fatalf("GetVerbQueue: %v", err)
	}
	if len(queue) != 2 || queue[0].UserVerbCardID != 500 || queue[1].UserVerbCardID != 501 {
		t.Fatalf("unexpected queue: %+v", queue)
	}
}

func TestVerbForms_sqlmockErrorPaths(t *testing.T) {
	t.Run("UpsertVerbLemma exec", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_lemmas`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.UpsertVerbLemma("hablar", "es", "s", "v", "c", `{}`); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("UpsertVerbLemma select", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_lemmas`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT id FROM verb_lemmas`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.UpsertVerbLemma("hablar", "es", "s", "v", "c", `{}`); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("UpsertVerbForm exec", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_forms_dict`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.UpsertVerbForm(&models.VerbFormDict{VerbLemmaID: 1, Mood: "indicativo", Tense: "presente", Person: "1", Number: "singular", SurfaceForm: "x"}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("UpsertVerbForm select", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_forms_dict`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT id FROM verb_forms_dict`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.UpsertVerbForm(&models.VerbFormDict{VerbLemmaID: 1, Mood: "indicativo", Tense: "presente", Person: "1", Number: "singular", SurfaceForm: "x"}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("LinkWordCardToLemma", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO word_verb_lemmas`).WillReturnError(sql.ErrConnDone)
		if err := repo.LinkWordCardToLemma(1, 2, 1.0, "test"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("LinkWordCardByLemma link", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT id FROM verb_lemmas`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(9)))
		mock.ExpectExec(`INSERT INTO word_verb_lemmas`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.LinkWordCardByLemma(1, "hablar", "es", "test"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("LinkMissingSpanishVerbLemmasForUser scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT DISTINCT wc\.id`).WillReturnRows(sqlmock.NewRows([]string{"id", "word"}).AddRow("bad", "hablar"))
		if err := repo.LinkMissingSpanishVerbLemmasForUser(900001); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("LinkMissingSpanishVerbLemmasForUser link", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT DISTINCT wc\.id`).WillReturnRows(sqlmock.NewRows([]string{"id", "word"}).AddRow(int64(10), "hablar"))
		mock.ExpectQuery(`SELECT id FROM verb_lemmas`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(9)))
		mock.ExpectExec(`INSERT INTO word_verb_lemmas`).WillReturnError(sql.ErrConnDone)
		if err := repo.LinkMissingSpanishVerbLemmasForUser(900001); err == nil {
			t.Fatal("expected link error")
		}
	})
	t.Run("UpsertVerbFormExample exec", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_form_examples`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.UpsertVerbFormExample(&models.VerbFormExample{VerbFormDictID: 1, ExampleTarget: "x"}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("UpsertVerbFormExample select", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_form_examples`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT id FROM verb_form_examples`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.UpsertVerbFormExample(&models.VerbFormExample{VerbFormDictID: 1, ExampleTarget: "x"}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ListVerbFormsForLemma query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT d\.id, d\.mood`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.ListVerbFormsForLemma("hablar", "es"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ListVerbFormsForLemma scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT d\.id, d\.mood`).WillReturnRows(sqlmock.NewRows([]string{"id", "mood", "tense", "person", "number", "surface_form"}).AddRow("x", "m", "t", "p", "n", "s"))
		if _, err := repo.ListVerbFormsForLemma("hablar", "es"); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("ListVerbFormViewRowsForLemma query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT 0, l\.lemma`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.ListVerbFormViewRowsForLemma("hablar", "es"); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ListVerbFormViewRowsForLemma scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT 0, l\.lemma`).WillReturnRows(sqlmock.NewRows([]string{"word_card_id", "lemma", "mood", "tense", "person", "number", "surface_form", "is_irregular"}).AddRow("x", "l", "m", "t", "p", "n", "s", 0))
		if _, err := repo.ListVerbFormViewRowsForLemma("hablar", "es"); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("GetUserVerbForms query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT w\.id, w\.word, d\.mood`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetUserVerbForms(900001, 1); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetUserVerbForms scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT w\.id, w\.word, d\.mood`).WillReturnRows(sqlmock.NewRows([]string{"id", "word", "mood", "tense", "person", "number", "surface_form", "is_irregular"}).AddRow("x", "w", "m", "t", "p", "n", "s", 0))
		if _, err := repo.GetUserVerbForms(900001, 1); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("GetLinkedVerbFormsForUser query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT w\.id, w\.word, d\.id`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetLinkedVerbFormsForUser(900001, []string{"es.presente.indicativo"}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetLinkedVerbFormsForUser scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT w\.id, w\.word, d\.id`).WillReturnRows(sqlmock.NewRows([]string{"id", "word", "form_id", "mood", "tense", "person", "number", "surface_form"}).AddRow("x", "w", "f", "m", "t", "p", "n", "s"))
		if _, err := repo.GetLinkedVerbFormsForUser(900001, []string{"es.presente.indicativo"}); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("GetVerbFormExamples query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT id, verb_form_dict_id`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetVerbFormExamples(1, 5); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetVerbFormExamples scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT id, verb_form_dict_id`).WillReturnRows(sqlmock.NewRows([]string{"id", "verb_form_dict_id", "example_target", "gloss_native", "source", "quality_score"}).AddRow("x", 1, "e", "g", "s", 1))
		if _, err := repo.GetVerbFormExamples(1, 5); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("GetOrCreateUserVerbCard lookup error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT id FROM user_verb_cards`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetOrCreateUserVerbCard(900001, 1); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetOrCreateUserVerbCard create", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT id FROM user_verb_cards`).WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery(`INSERT INTO user_verb_cards`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(77)))
		id, err := repo.GetOrCreateUserVerbCard(900001, 1)
		if err != nil || id != 77 {
			t.Fatalf("create: id=%d err=%v", id, err)
		}
	})
	t.Run("GetOrCreateUserVerbCard create error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT id FROM user_verb_cards`).WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery(`INSERT INTO user_verb_cards`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetOrCreateUserVerbCard(900001, 1); err == nil {
			t.Fatal("expected create error")
		}
	})
	t.Run("UpsertVerbTrainingCard exec", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_training_cards`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{WordCardID: 1, VerbFormDictID: 2, CardType: models.VerbCardTypeCloze, PromptJSON: `{}`, AnswerJSON: `{}`}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("UpsertVerbTrainingCard select", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_training_cards`).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery(`SELECT id FROM verb_training_cards`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{WordCardID: 1, VerbFormDictID: 2, CardType: models.VerbCardTypeCloze, PromptJSON: `{}`, AnswerJSON: `{}`}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("EnsureUserCardsForUserWords query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT DISTINCT c\.id`).WillReturnError(sql.ErrConnDone)
		if err := repo.EnsureUserCardsForUserWords(900001, []string{"es.presente.indicativo"}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("EnsureUserCardsForUserWords scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT DISTINCT c\.id`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))
		if err := repo.EnsureUserCardsForUserWords(900001, []string{"es.presente.indicativo"}); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("GetVerbQueue due query error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetVerbQueue(900001, time.Now(), 5, 0); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetVerbQueue due scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		now := time.Now()
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).AddRow("x", 1, models.VerbCardTypeCloze, "{}", "{}", "[]"))
		if _, err := repo.GetVerbQueue(900001, now, 5, 0); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("GetVerbQueue new query error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		now := time.Now()
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, 30).
			WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetVerbQueue(900001, now, 5, 1); err == nil {
			t.Fatal("expected new queue error")
		}
	})
	t.Run("GetVerbQueue new scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		now := time.Now()
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, 30).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).AddRow("x", 1, models.VerbCardTypeCloze, "{}", "{}", "[]"))
		if _, err := repo.GetVerbQueue(900001, now, 5, 1); err == nil {
			t.Fatal("expected new scan error")
		}
	})
	t.Run("CountUserVerbClozeCards error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_verb_cards`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.CountUserVerbClozeCards(900001); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetVerbUserCardSRS query error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT id, state, ef`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetVerbUserCardSRS(1); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("UpdateVerbUserCardSRS error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`UPDATE user_verb_cards`).WillReturnError(sql.ErrConnDone)
		if err := repo.UpdateVerbUserCardSRS(&VerbUserCardSRS{ID: 1, State: "new", EF: 2.5}, time.Now(), 3); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("CreateVerbReviewEvent error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`INSERT INTO verb_review_events`).WillReturnError(sql.ErrConnDone)
		if err := repo.CreateVerbReviewEvent(1, 900001, 2, false, 1); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("StartVerbSession error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`INSERT INTO verb_training_sessions`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.StartVerbSession(900001, 1, `{}`); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("FinishVerbSession error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`UPDATE verb_training_sessions`).WillReturnError(sql.ErrConnDone)
		if err := repo.FinishVerbSession(1, 1); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetVerbSessionStats error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrConnDone)
		if _, _, err := repo.GetVerbSessionStats(1); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetVerbLemmaMetadataJSONBatch query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT lemma, COALESCE`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.GetVerbLemmaMetadataJSONBatch([]string{"hablar"}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("GetVerbLemmaMetadataJSONBatch scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"lemma", "metadata_json"}).
			AddRow("hablar", `{}`).
			RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT lemma, COALESCE`).WillReturnRows(rows)
		if _, err := repo.GetVerbLemmaMetadataJSONBatch([]string{"hablar"}); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("UpdateVerbLemmaMetadataJSON exec", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectExec(`UPDATE verb_lemmas SET metadata_json`).WillReturnError(sql.ErrConnDone)
		if err := repo.UpdateVerbLemmaMetadataJSON("hablar", `{}`); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ListSpanishVerbLemmas query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT lemma FROM verb_lemmas`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.ListSpanishVerbLemmas(); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ListSpanishVerbLemmas scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"lemma"}).AddRow("hablar").RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT lemma FROM verb_lemmas`).WillReturnRows(rows)
		if _, err := repo.ListSpanishVerbLemmas(); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("ListPendingVerbTrainingLemmas query error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT w\.id, LOWER\(TRIM\(w\.word\)\)`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.ListPendingVerbTrainingLemmas(10, 0, false); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ListPendingVerbTrainingLemmas scan error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT w\.id, LOWER\(TRIM\(w\.word\)\)`).WillReturnRows(sqlmock.NewRows([]string{"id", "lemma"}).AddRow("x", "hablar"))
		if _, err := repo.ListPendingVerbTrainingLemmas(10, 0, false); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("ListAdminVerbTrainingLemmas query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT w\.id, LOWER\(TRIM\(w\.word\)\)`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.ListAdminVerbTrainingLemmas("", 10, 0); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ListAdminVerbTrainingLemmas scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT w\.id, LOWER\(TRIM\(w\.word\)\)`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "lemma", "cloze_count", "metadata_json"}).AddRow("x", "hablar", 1, `{}`))
		if _, err := repo.ListAdminVerbTrainingLemmas("", 10, 0); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("ListAdminVerbTrainingCardsByWordCard query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT vtc\.id, vtc\.card_type`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.ListAdminVerbTrainingCardsByWordCard(1); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ListAdminVerbTrainingCardsByWordCard scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT vtc\.id, vtc\.card_type`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "card_type", "prompt_json", "answer_json", "distractors_json", "mood", "tense", "person", "number", "surface_form"}).
				AddRow("x", models.VerbCardTypeCloze, "{}", "{}", "[]", "m", "t", "p", "n", "s"))
		if _, err := repo.ListAdminVerbTrainingCardsByWordCard(1); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("AdminVerbLemmaLookup query error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT LOWER\(TRIM\(w\.word\)\)`).WillReturnError(sql.ErrConnDone)
		if _, _, err := repo.AdminVerbLemmaLookup(1); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("listVerbExampleCatalogTemplatesUncached query", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT code, lemma_match`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.listVerbExampleCatalogTemplatesUncached(); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("listVerbExampleCatalogTemplatesUncached scan", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"code", "lemma_match", "verb_class", "mood", "tense", "es_suffix", "ru_pattern"}).
			AddRow("tpl", "hablar", "ar", "indicativo", "presente", ".", ".").
			RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT code, lemma_match`).WillReturnRows(rows)
		if _, err := repo.listVerbExampleCatalogTemplatesUncached(); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("EnsureUserCardsForUserWords create error", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT DISTINCT c\.id`).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(900)))
		mock.ExpectQuery(`SELECT id FROM user_verb_cards`).WillReturnError(sql.ErrConnDone)
		if err := repo.EnsureUserCardsForUserWords(900001, []string{"es.presente.indicativo"}); err == nil {
			t.Fatal("expected create error")
		}
	})
	t.Run("ListVerbFormsForLemma rows err", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"id", "mood", "tense", "person", "number", "surface_form"}).
			AddRow(int64(1), "indicativo", "presente", "1", "singular", "hablo").
			RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT d\.id, d\.mood`).WillReturnRows(rows)
		if _, err := repo.ListVerbFormsForLemma("hablar", "es"); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("LinkMissingSpanishVerbLemmasForUser rows err", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"id", "word"}).AddRow(int64(10), "hablar").RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT DISTINCT wc\.id`).WillReturnRows(rows)
		if err := repo.LinkMissingSpanishVerbLemmasForUser(900001); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("ListVerbExampleCatalogTemplatesCached load error", func(t *testing.T) {
		ResetVerbExampleCatalogCacheForTests()
		t.Cleanup(func() { ResetVerbExampleCatalogCacheForTests() })
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		mock.ExpectQuery(`SELECT code, lemma_match`).WillReturnError(sql.ErrConnDone)
		if _, err := repo.ListVerbExampleCatalogTemplatesCached(); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestVerbForms_GetVerbQueue_limitsAndMaxNewZero(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())
	now := time.Now()

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(900001), models.VerbCardTypeCloze, now, 5000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))

	queue, err := repo.GetVerbQueue(900001, now, 5000, 0)
	if err != nil {
		t.Fatalf("GetVerbQueue maxNew=0: %v", err)
	}
	if len(queue) != 0 {
		t.Fatalf("expected empty queue, got %d", len(queue))
	}

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))

	queue, err = repo.GetVerbQueue(900001, now, -3, -2)
	if err != nil {
		t.Fatalf("GetVerbQueue negative limits: %v", err)
	}
	if len(queue) != 0 {
		t.Fatalf("expected empty queue with negative limits, got %d", len(queue))
	}
}

func TestVerbForms_GetVerbLemmaMetadataJSONBatch_onlyBlank(t *testing.T) {
	repo, _ := setupVerbFormsRepo(t)
	out, err := repo.GetVerbLemmaMetadataJSONBatch([]string{"", "   "})
	if err != nil || len(out) != 0 {
		t.Fatalf("only blank lemmas: %+v err=%v", out, err)
	}
}

func TestVerbForms_ListAdminVerbTrainingLemmas_limitsAndSearch(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsMoreFixtures(t, db)

	capped, err := repo.ListAdminVerbTrainingLemmas("", 500, 0)
	if err != nil {
		t.Fatalf("ListAdminVerbTrainingLemmas cap limit: %v", err)
	}
	if len(capped) == 0 {
		t.Fatal("expected at least one lemma")
	}

	none, err := repo.ListAdminVerbTrainingLemmas("zzz-no-match", 50, 0)
	if err != nil {
		t.Fatalf("ListAdminVerbTrainingLemmas search miss: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("expected no rows for zzz-no-match, got %d", len(none))
	}
	_ = seed
}

func TestVerbForms_ListPendingVerbTrainingLemmas_highLimit(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	_ = seedVerbFormsMoreFixtures(t, db)
	if _, err := repo.ListPendingVerbTrainingLemmas(9999, 0, true); err != nil {
		t.Fatalf("ListPendingVerbTrainingLemmas high limit: %v", err)
	}
}

func TestVerbForms_remainingCoverageGaps(t *testing.T) {
	t.Run("ListVerbFormViewRowsForLemma rows err", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"word_card_id", "lemma", "mood", "tense", "person", "number", "surface_form", "is_irregular"}).
			AddRow(int64(0), "hablar", "indicativo", "presente", "1", "singular", "hablo", int64(0)).
			RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT 0, l\.lemma`).WillReturnRows(rows)
		if _, err := repo.ListVerbFormViewRowsForLemma("hablar", "es"); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("ListPendingVerbTrainingLemmas rows err", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"id", "lemma"}).AddRow(int64(10), "hablar").RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT w\.id, LOWER\(TRIM\(w\.word\)\)`).WillReturnRows(rows)
		if _, err := repo.ListPendingVerbTrainingLemmas(10, 0, false); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("ListAdminVerbTrainingLemmas rows err", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"id", "lemma", "cloze_count", "metadata_json"}).
			AddRow(int64(1), "hablar", int64(0), `{}`).
			RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT w\.id, LOWER\(TRIM\(w\.word\)\)`).WillReturnRows(rows)
		if _, err := repo.ListAdminVerbTrainingLemmas("", 10, 0); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("ListAdminVerbTrainingCardsByWordCard rows err", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		rows := sqlmock.NewRows([]string{"id", "card_type", "prompt_json", "answer_json", "distractors_json", "mood", "tense", "person", "number", "surface_form"}).
			AddRow(int64(1), models.VerbCardTypeCloze, "{}", "{}", "[]", "m", "t", "p", "n", "s").
			RowError(0, sql.ErrConnDone)
		mock.ExpectQuery(`SELECT vtc\.id, vtc\.card_type`).WillReturnRows(rows)
		if _, err := repo.ListAdminVerbTrainingCardsByWordCard(1); err == nil {
			t.Fatal("expected rows error")
		}
	})
	t.Run("GetVerbLemmaMetadataJSONBatch scan type mismatch", func(t *testing.T) {
		db := newVerbFormsBadScanDB(t, []string{"lemma", "metadata_json"}, []driver.Value{struct{}{}, struct{}{}})
		repo := NewVerbFormsRepository(db, zap.NewNop())
		if _, err := repo.GetVerbLemmaMetadataJSONBatch([]string{"hablar"}); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("ListSpanishVerbLemmas scan type mismatch", func(t *testing.T) {
		db := newVerbFormsBadScanDB(t, []string{"lemma"}, []driver.Value{struct{}{}})
		repo := NewVerbFormsRepository(db, zap.NewNop())
		if _, err := repo.ListSpanishVerbLemmas(); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("listVerbExampleCatalogTemplatesUncached scan type mismatch", func(t *testing.T) {
		db := newVerbFormsBadScanDB(t,
			[]string{"code", "lemma_match", "verb_class", "mood", "tense", "es_suffix", "ru_pattern"},
			[]driver.Value{struct{}{}, struct{}{}, struct{}{}, struct{}{}, struct{}{}, struct{}{}, struct{}{}},
		)
		repo := NewVerbFormsRepository(db, zap.NewNop())
		if _, err := repo.listVerbExampleCatalogTemplatesUncached(); err == nil {
			t.Fatal("expected scan error")
		}
	})
	t.Run("GetVerbQueue high maxNew poolCap", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		now := time.Now()
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, 700).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))
		if _, err := repo.GetVerbQueue(900001, now, 10, 700); err != nil {
			t.Fatalf("GetVerbQueue high maxNew: %v", err)
		}
	})
	t.Run("GetVerbQueue truncate to maxCards", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		now := time.Now()
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
				AddRow(int64(101), int64(10), models.VerbCardTypeCloze, "{}", "{}", "[]").
				AddRow(int64(102), int64(11), models.VerbCardTypeCloze, "{}", "{}", "[]"))
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, 30).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
				AddRow(int64(201), int64(20), models.VerbCardTypeCloze, "{}", "{}", "[]"))
		queue, err := repo.GetVerbQueue(900001, now, 2, 1)
		if err != nil {
			t.Fatalf("GetVerbQueue truncate: %v", err)
		}
		if len(queue) != 2 {
			t.Fatalf("expected truncated queue len 2, got %d", len(queue))
		}
	})
	t.Run("GetVerbQueue new scan error closes rows", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		repo := NewVerbFormsRepository(db, zap.NewNop())
		now := time.Now()
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}))
		mock.ExpectQuery(`SELECT uvc\.id, vtc\.word_card_id, vtc\.card_type, vtc\.prompt_json`).
			WithArgs(int64(900001), models.VerbCardTypeCloze, 30).
			WillReturnRows(sqlmock.NewRows([]string{"id", "word_card_id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
				AddRow("bad-id", int64(20), models.VerbCardTypeCloze, "{}", "{}", "[]"))
		if _, err := repo.GetVerbQueue(900001, now, 10, 1); err == nil {
			t.Fatal("expected new queue scan error")
		}
	})
}
