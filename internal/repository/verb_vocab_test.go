package repository

import (
	"testing"
	"tgbot-skeleton/internal/models"
	"time"
)

func TestEnsureUserVerbCardsForWordPreservesProgressAndScope(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsTrainingFixtures(t, db)
	cardID := insertVerbTrainingCard(t, db, seed, seed.formPresenteID, validVerbPromptJSON())
	insertVerbTrainingCard(t, db, seed, seed.formFuturoID, validVerbPromptJSON())
	scopes := []string{"es.presente.indicativo"}
	// An unrelated word must not materialize even the user's existing vocabulary.
	if err := repo.EnsureUserVerbCardsForWord(seed.userID, seed.wordCardID+100000, scopes); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM user_verb_cards WHERE user_id=?`, seed.userID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("unrelated word: %d %v", count, err)
	}
	for i := 0; i < 2; i++ {
		if err := repo.EnsureUserVerbCardsForWord(seed.userID, seed.wordCardID, scopes); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.QueryRow(`SELECT count(*) FROM user_verb_cards WHERE user_id=?`, seed.userID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("scope/idempotency: %d %v", count, err)
	}
	if _, err := db.Exec(`UPDATE user_verb_cards SET state='review',reps=7 WHERE user_id=? AND verb_training_card_id=?`, seed.userID, cardID); err != nil {
		t.Fatal(err)
	}
	if err := repo.EnsureUserVerbCardsForWord(seed.userID, seed.wordCardID, scopes); err != nil {
		t.Fatal(err)
	}
	var state string
	var reps int
	if err := db.QueryRow(`SELECT state,reps FROM user_verb_cards WHERE user_id=? AND verb_training_card_id=?`, seed.userID, cardID).Scan(&state, &reps); err != nil || state != "review" || reps != 7 {
		t.Fatalf("lost SRS: %s %d %v", state, reps, err)
	}
	rows, err := repo.GetLinkedVerbFormsForWord(seed.userID, seed.wordCardID+100000, scopes)
	if err != nil || len(rows) != 0 {
		t.Fatalf("unrelated forms: %v %v", rows, err)
	}
	rows, err = repo.GetLinkedVerbFormsForWord(seed.userID, seed.wordCardID, scopes)
	if err != nil || len(rows) != 1 {
		t.Fatalf("word forms: %v %v", rows, err)
	}
}

func TestUpsertVerbTrainingCardSkipsUnchangedContent(t *testing.T) {
	repo, db := setupVerbFormsRepo(t)
	seed := seedVerbFormsTrainingFixtures(t, db)
	card := &models.VerbTrainingCard{WordCardID: seed.wordCardID, VerbFormDictID: seed.formPresenteID, CardType: models.VerbCardTypeCloze, PromptJSON: validVerbPromptJSON(), AnswerJSON: `{}`, DistractorsJSON: `[]`}
	id, err := repo.UpsertVerbTrainingCard(card)
	if err != nil {
		t.Fatal(err)
	}
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := db.Exec(`UPDATE verb_training_cards SET updated_at=? WHERE id=?`, old, id); err != nil {
		t.Fatal(err)
	}
	again, err := repo.UpsertVerbTrainingCard(card)
	if err != nil || again != id {
		t.Fatalf("upsert: %d %v", again, err)
	}
	var updated time.Time
	if err := db.QueryRow(`SELECT updated_at FROM verb_training_cards WHERE id=?`, id).Scan(&updated); err != nil || !updated.Equal(old) {
		t.Fatalf("unchanged content rewritten: %v %v", updated, err)
	}
	card.AnswerJSON = `{"answer":"hablo"}`
	if _, err := repo.UpsertVerbTrainingCard(card); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT updated_at FROM verb_training_cards WHERE id=?`, id).Scan(&updated); err != nil || updated.Equal(old) {
		t.Fatalf("changed content not updated: %v %v", updated, err)
	}
}
