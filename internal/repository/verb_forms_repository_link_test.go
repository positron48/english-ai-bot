package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"tgbot-skeleton/internal/models"
)

func TestLinkMissingSpanishVerbLemmasForUser_EmptyVocab(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerbFormsRepository(db, zap.NewNop())
	mock.ExpectQuery(`SELECT DISTINCT wc\.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "word"}),
	)

	if err := repo.LinkMissingSpanishVerbLemmasForUser(5); err != nil {
		t.Fatalf("LinkMissingSpanishVerbLemmasForUser: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLinkMissingSpanishVerbLemmasForUser_LinksWhenLemmaExists(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerbFormsRepository(db, zap.NewNop())

	mock.ExpectQuery(`SELECT DISTINCT wc\.id`).WillReturnRows(
		sqlmock.NewRows([]string{"id", "word"}).AddRow(int64(10), "hablar"),
	)
	mock.ExpectQuery(`SELECT id FROM verb_lemmas WHERE lemma`).
		WithArgs("hablar", "es").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))
	mock.ExpectExec(`INSERT INTO word_verb_lemmas`).
		WithArgs(int64(10), int64(99), 1.0, "auto_user_vocab").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.LinkMissingSpanishVerbLemmasForUser(7); err != nil {
		t.Fatalf("LinkMissingSpanishVerbLemmasForUser: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// EnsureUserCardsForUserWords must bind placeholders as: card_type, each scope, user_id, user_id
// (WHERE c.card_type = ? comes before IN (...)). Wrong order silently returns no rows → no user_verb_cards.
func TestListPendingVerbTrainingLemmas_JoinsVerbLemmasAndTrainingCardPOS(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerbFormsRepository(db, zap.NewNop())

	mock.ExpectQuery(`SELECT w\.id, LOWER\(TRIM\(w\.word\)\)`).
		WithArgs(int64(0), 50).
		WillReturnRows(sqlmock.NewRows([]string{"id", "lemma"}).
			AddRow(int64(100), "hablar"))

	got, err := repo.ListPendingVerbTrainingLemmas(50, 0)
	if err != nil {
		t.Fatalf("ListPendingVerbTrainingLemmas: %v", err)
	}
	if len(got) != 1 || got[0].WordCardID != 100 || got[0].Lemma != "hablar" {
		t.Fatalf("got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureUserCardsForUserWords_QueryArgOrder(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewVerbFormsRepository(db, zap.NewNop())

	mock.ExpectQuery(`SELECT DISTINCT c\.id`).
		WithArgs(models.VerbCardTypeCloze, "es.presente.indicativo", int64(42), int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(900)))
	mock.ExpectQuery(`SELECT id FROM user_verb_cards WHERE user_id=\? AND verb_training_card_id=\?`).
		WithArgs(int64(42), int64(900)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	if err := repo.EnsureUserCardsForUserWords(42, []string{"es.presente.indicativo"}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
