package repository

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupVerbFormsRepo(t *testing.T) (*VerbFormsRepository, *sql.DB) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	return NewVerbFormsRepository(db, zap.NewNop()), db
}

type verbFormsSeed struct {
	userID         int64
	wordCardID     int64
	verbLemmaID    int64
	formPresenteID int64
	formFuturoID   int64
}

func seedVerbFormsTrainingFixtures(t *testing.T, db *sql.DB) verbFormsSeed {
	t.Helper()
	var out verbFormsSeed
	if err := db.QueryRow(`INSERT INTO users (telegram_id) VALUES (77001) RETURNING id`).Scan(&out.userID); err != nil {
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
	out.verbLemmaID, err = repo.UpsertVerbLemma("hablar", "es", "test", "v1", "chk", `{"ru":{"gloss":"говорить"}}`)
	if err != nil {
		t.Fatalf("UpsertVerbLemma: %v", err)
	}
	out.formPresenteID, err = repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: out.verbLemmaID, Mood: "indicativo", Tense: "presente",
		Person: "1", Number: "singular", SurfaceForm: "hablo",
	})
	if err != nil {
		t.Fatalf("UpsertVerbForm presente: %v", err)
	}
	out.formFuturoID, err = repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: out.verbLemmaID, Mood: "indicativo", Tense: "futuro_simple",
		Person: "1", Number: "singular", SurfaceForm: "hablaré",
	})
	if err != nil {
		t.Fatalf("UpsertVerbForm futuro: %v", err)
	}
	if err := repo.LinkWordCardToLemma(out.wordCardID, out.verbLemmaID, 1.0, "test"); err != nil {
		t.Fatalf("LinkWordCardToLemma: %v", err)
	}
	return out
}

func insertVerbTrainingCard(t *testing.T, db *sql.DB, seed verbFormsSeed, formID int64, promptJSON string) int64 {
	t.Helper()
	repo := NewVerbFormsRepository(db, zap.NewNop())
	id, err := repo.UpsertVerbTrainingCard(&models.VerbTrainingCard{
		WordCardID:      seed.wordCardID,
		VerbFormDictID:  formID,
		CardType:        models.VerbCardTypeCloze,
		PromptJSON:      promptJSON,
		AnswerJSON:      `{"surface_form":"hablo"}`,
		DistractorsJSON: `["hablas","habla","hablamos"]`,
	})
	if err != nil {
		t.Fatalf("UpsertVerbTrainingCard: %v", err)
	}
	return id
}

func validVerbPromptJSON() string {
	prompt := map[string]interface{}{
		"type":                models.VerbCardTypeCloze,
		"lemma":               "hablar",
		"mood":                "indicativo",
		"tense":               "presente",
		"question":            "Yo ___ español.",
		"example_translation": "Я говорю по-испански.",
	}
	raw, _ := json.Marshal(prompt)
	return string(raw)
}

func TestUpsertVerbFormExample_CRUD(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsTrainingFixtures(t, db)

	id1, err := repo.UpsertVerbFormExample(&models.VerbFormExample{
		VerbFormDictID: seed.formPresenteID,
		ExampleTarget:  "Yo hablo español.",
		GlossNative:    "Я говорю по-испански.",
		Source:         "test",
		QualityScore:   10,
	})
	if err != nil {
		t.Fatalf("UpsertVerbFormExample insert: %v", err)
	}
	if id1 <= 0 {
		t.Fatalf("expected positive id, got %d", id1)
	}

	id2, err := repo.UpsertVerbFormExample(&models.VerbFormExample{
		VerbFormDictID: seed.formPresenteID,
		ExampleTarget:  "Yo hablo español.",
		GlossNative:    "Я говорю испанский.",
		Source:         "test-updated",
		QualityScore:   20,
	})
	if err != nil {
		t.Fatalf("UpsertVerbFormExample update: %v", err)
	}
	if id2 != id1 {
		t.Fatalf("upsert on conflict should keep id: got %d want %d", id2, id1)
	}

	_, err = repo.UpsertVerbFormExample(&models.VerbFormExample{
		VerbFormDictID: seed.formPresenteID,
		ExampleTarget:  "Hablo con mi amigo.",
		GlossNative:    "Я говорю с другом.",
		Source:         "test",
		QualityScore:   5,
	})
	if err != nil {
		t.Fatalf("UpsertVerbFormExample second row: %v", err)
	}

	examples, err := repo.GetVerbFormExamples(seed.formPresenteID, 10)
	if err != nil {
		t.Fatalf("GetVerbFormExamples: %v", err)
	}
	if len(examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(examples))
	}
	if examples[0].QualityScore != 20 || examples[0].GlossNative != "Я говорю испанский." {
		t.Fatalf("expected highest quality first, got %+v", examples[0])
	}
}

func TestGetLinkedVerbFormsForUser_unlockLadderScopes(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsTrainingFixtures(t, db)

	presente, err := repo.GetLinkedVerbFormsForUser(seed.userID, []string{"es.presente.indicativo"})
	if err != nil {
		t.Fatalf("GetLinkedVerbFormsForUser presente: %v", err)
	}
	if len(presente) != 1 || presente[0].VerbFormDictID != seed.formPresenteID {
		t.Fatalf("presente scope: got %+v", presente)
	}

	futuro, err := repo.GetLinkedVerbFormsForUser(seed.userID, []string{"es.futuro_simple.indicativo"})
	if err != nil {
		t.Fatalf("GetLinkedVerbFormsForUser futuro: %v", err)
	}
	if len(futuro) != 1 || futuro[0].VerbFormDictID != seed.formFuturoID {
		t.Fatalf("futuro scope: got %+v", futuro)
	}

	both, err := repo.GetLinkedVerbFormsForUser(seed.userID, []string{
		"es.presente.indicativo", "es.futuro_simple.indicativo",
	})
	if err != nil {
		t.Fatalf("GetLinkedVerbFormsForUser both: %v", err)
	}
	if len(both) != 2 {
		t.Fatalf("expected 2 forms for both scopes, got %d", len(both))
	}

	if rows, err := repo.GetLinkedVerbFormsForUser(seed.userID, nil); err != nil || rows != nil {
		t.Fatalf("empty scopes should return nil,nil; got %v err=%v", rows, err)
	}
}

func TestEnsureUserCardsForUserWords_respectsUnlockScope(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsTrainingFixtures(t, db)

	presenteCardID := insertVerbTrainingCard(t, db, seed, seed.formPresenteID, validVerbPromptJSON())
	insertVerbTrainingCard(t, db, seed, seed.formFuturoID, validVerbPromptJSON())

	if err := repo.EnsureUserCardsForUserWords(seed.userID, []string{"es.presente.indicativo"}); err != nil {
		t.Fatalf("EnsureUserCardsForUserWords: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM user_verb_cards uvc
		INNER JOIN verb_training_cards vtc ON vtc.id = uvc.verb_training_card_id
		WHERE uvc.user_id = ?`, seed.userID).Scan(&count); err != nil {
		t.Fatalf("count user_verb_cards: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 user_verb_card for presente scope, got %d", count)
	}

	uvcID, err := repo.GetOrCreateUserVerbCard(seed.userID, presenteCardID)
	if err != nil || uvcID <= 0 {
		t.Fatalf("GetOrCreateUserVerbCard existing: id=%d err=%v", uvcID, err)
	}
	uvcID2, err := repo.GetOrCreateUserVerbCard(seed.userID, presenteCardID)
	if err != nil || uvcID2 != uvcID {
		t.Fatalf("GetOrCreateUserVerbCard should be idempotent: %d vs %d err=%v", uvcID2, uvcID, err)
	}
}

func TestGetVerbQueue_dueBeforeNewAndRequiresExampleTranslation(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsTrainingFixtures(t, db)

	validCardID := insertVerbTrainingCard(t, db, seed, seed.formPresenteID, validVerbPromptJSON())
	badPrompt, _ := json.Marshal(map[string]interface{}{
		"type": models.VerbCardTypeCloze, "lemma": "hablar", "question": "Yo ___.",
	})
	insertVerbTrainingCard(t, db, seed, seed.formFuturoID, string(badPrompt))

	validUVC, err := repo.GetOrCreateUserVerbCard(seed.userID, validCardID)
	if err != nil {
		t.Fatalf("GetOrCreateUserVerbCard valid: %v", err)
	}
	_, _ = db.Exec(`UPDATE user_verb_cards SET state='review', next_due_at=$1 WHERE id=$2`,
		time.Now().Add(-time.Hour), validUVC)

	now := time.Now()
	queue, err := repo.GetVerbQueue(seed.userID, now, 10, 5)
	if err != nil {
		t.Fatalf("GetVerbQueue: %v", err)
	}
	if len(queue) != 1 {
		t.Fatalf("expected only card with example_translation, got %d: %+v", len(queue), queue)
	}
	if queue[0].UserVerbCardID != validUVC {
		t.Fatalf("expected due card %d, got %d", validUVC, queue[0].UserVerbCardID)
	}
}

func TestGetVerbQueue_newCardsRoundRobin(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsTrainingFixtures(t, db)

	var wordCard2 int64
	if err := db.QueryRow(`INSERT INTO word_cards (word, definition, pos) VALUES ('comer', 'есть', 'verb') RETURNING id`).
		Scan(&wordCard2); err != nil {
		t.Fatalf("insert second word_card: %v", err)
	}
	var tc2 int64
	if err := db.QueryRow(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos)
		VALUES ($1, 'comer', 0, 'есть', 'to eat', 'verb') RETURNING id`, wordCard2).Scan(&tc2); err != nil {
		t.Fatalf("insert training_card 2: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state) VALUES ($1, $2, 'es_ru', 'review')`,
		seed.userID, tc2); err != nil {
		t.Fatalf("insert user_card 2: %v", err)
	}
	lemma2ID, err := repo.UpsertVerbLemma("comer", "es", "test", "v1", "chk", `{}`)
	if err != nil {
		t.Fatalf("UpsertVerbLemma comer: %v", err)
	}
	form2ID, err := repo.UpsertVerbForm(&models.VerbFormDict{
		VerbLemmaID: lemma2ID, Mood: "indicativo", Tense: "presente",
		Person: "1", Number: "singular", SurfaceForm: "como",
	})
	if err != nil {
		t.Fatalf("UpsertVerbForm comer: %v", err)
	}
	if err := repo.LinkWordCardToLemma(wordCard2, lemma2ID, 1.0, "test"); err != nil {
		t.Fatalf("link comer: %v", err)
	}

	seed2 := seed
	seed2.wordCardID = wordCard2
	c1 := insertVerbTrainingCard(t, db, seed, seed.formPresenteID, validVerbPromptJSON())
	c2 := insertVerbTrainingCard(t, db, seed2, form2ID, validVerbPromptJSON())
	for _, cardID := range []int64{c1, c2} {
		uvc, err := repo.GetOrCreateUserVerbCard(seed.userID, cardID)
		if err != nil {
			t.Fatalf("GetOrCreateUserVerbCard: %v", err)
		}
		_, _ = db.Exec(`UPDATE user_verb_cards SET state='new', next_due_at=NULL WHERE id=$1`, uvc)
	}

	queue, err := repo.GetVerbQueue(seed.userID, time.Now(), 10, 2)
	if err != nil {
		t.Fatalf("GetVerbQueue: %v", err)
	}
	if len(queue) != 2 {
		t.Fatalf("expected 2 new cards, got %d", len(queue))
	}
	if queue[0].WordCardID == queue[1].WordCardID {
		t.Fatalf("round-robin should mix lemmas, got same word_card_id twice: %+v", queue)
	}
}

func TestVerbSessionAndSRSFlow(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsTrainingFixtures(t, db)
	trainingCardID := insertVerbTrainingCard(t, db, seed, seed.formPresenteID, validVerbPromptJSON())
	uvcID, err := repo.GetOrCreateUserVerbCard(seed.userID, trainingCardID)
	if err != nil {
		t.Fatalf("GetOrCreateUserVerbCard: %v", err)
	}

	sessionID, err := repo.StartVerbSession(seed.userID, 1, `{"scopes":["es.presente.indicativo"]}`)
	if err != nil || sessionID <= 0 {
		t.Fatalf("StartVerbSession: id=%d err=%v", sessionID, err)
	}

	srs, err := repo.GetVerbUserCardSRS(uvcID)
	if err != nil || srs == nil || srs.State != "new" {
		t.Fatalf("GetVerbUserCardSRS: %+v err=%v", srs, err)
	}
	srs.Reps = 1
	srs.State = "review"
	srs.IntervalDays = 1
	nextDue := time.Now().Add(24 * time.Hour)
	if err := repo.UpdateVerbUserCardSRS(srs, nextDue, 5); err != nil {
		t.Fatalf("UpdateVerbUserCardSRS: %v", err)
	}
	if err := repo.CreateVerbReviewEvent(sessionID, seed.userID, uvcID, true, 5); err != nil {
		t.Fatalf("CreateVerbReviewEvent: %v", err)
	}
	if err := repo.FinishVerbSession(sessionID, 1); err != nil {
		t.Fatalf("FinishVerbSession: %v", err)
	}
	total, correct, err := repo.GetVerbSessionStats(sessionID)
	if err != nil {
		t.Fatalf("GetVerbSessionStats: %v", err)
	}
	if total != 1 || correct != 1 {
		t.Fatalf("stats total=%d correct=%d", total, correct)
	}
}

func TestListVerbExampleCatalogTemplatesCached(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	t.Cleanup(func() { ResetVerbExampleCatalogCacheForTests() })

	if _, err := db.Exec(`INSERT INTO verb_example_templates (code, lemma_match, verb_class, mood, tense, es_suffix, ru_pattern, sort_order, active)
		VALUES ('test_tpl', 'hablar', 'ar', 'indicativo', 'presente', ' en casa.', ' дома.', 1, true)`); err != nil {
		t.Fatalf("insert template: %v", err)
	}

	tpl1, err := repo.ListVerbExampleCatalogTemplatesCached()
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if len(tpl1) == 0 || tpl1[0].ID != "test_tpl" {
		t.Fatalf("unexpected templates: %+v", tpl1)
	}

	tpl2, err := repo.ListVerbExampleCatalogTemplatesCached()
	if err != nil {
		t.Fatalf("cached load: %v", err)
	}
	if len(tpl2) != len(tpl1) {
		t.Fatalf("cache should return same count")
	}

	ResetVerbExampleCatalogCacheForTests()
	if _, err := db.Exec(`UPDATE verb_example_templates SET code='test_tpl_v2' WHERE code='test_tpl'`); err != nil {
		t.Fatalf("update template: %v", err)
	}
	tpl3, err := repo.ListVerbExampleCatalogTemplatesCached()
	if err != nil {
		t.Fatalf("reload after cache reset: %v", err)
	}
	if len(tpl3) == 0 || tpl3[0].ID != "test_tpl_v2" {
		t.Fatalf("expected refreshed template, got %+v", tpl3)
	}
}
