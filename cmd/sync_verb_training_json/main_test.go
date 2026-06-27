package main

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func beginMockTx(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *sql.Tx) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock, tx
}

func TestLookupWordCardIDForVerbLemma_ScopesToCourse(t *testing.T) {
	_, mock, tx := beginMockTx(t)
	mock.ExpectQuery(`SELECT id FROM word_cards WHERE LOWER\(TRIM\(word\)\) = \? AND LOWER\(COALESCE\(course_code, ''\)\) = \?`).
		WithArgs("algo", "es_ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1002)))
	mock.ExpectRollback()

	got, err := lookupWordCardIDForVerbLemma(tx, "algo", "es_ru")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if got != 1002 {
		t.Fatalf("got word_card_id=%d, want 1002", got)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLookupWordCardIDForVerbLemma_DoesNotFallbackToOtherCourse(t *testing.T) {
	_, mock, tx := beginMockTx(t)
	mock.ExpectQuery(`SELECT id FROM word_cards WHERE LOWER\(TRIM\(word\)\) = \? AND LOWER\(COALESCE\(course_code, ''\)\) = \?`).
		WithArgs("algo", "es_ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT tc\.word_card_id\s+FROM training_cards tc\s+JOIN word_cards wc ON wc\.id = tc\.word_card_id`).
		WithArgs("algo", "es_ru").
		WillReturnRows(sqlmock.NewRows([]string{"word_card_id"}))
	mock.ExpectQuery(`SELECT id\s+FROM word_cards wc\s+WHERE LOWER\(TRIM\(wc\.word\)\) = \?`).
		WithArgs("algo", "algo", "es_ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT tc\.word_card_id\s+FROM training_cards tc\s+JOIN word_cards wc ON wc\.id = tc\.word_card_id`).
		WithArgs("algo", "algo", "es_ru").
		WillReturnRows(sqlmock.NewRows([]string{"word_card_id"}))
	mock.ExpectRollback()

	_, err := lookupWordCardIDForVerbLemma(tx, "algo", "es_ru")
	if !errors.Is(err, ErrLemmaNoWordCard) {
		t.Fatalf("got err=%v, want ErrLemmaNoWordCard", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLookupWordCardIDForVerbLemma_AllowsLegacyUntaggedWhenNoOtherCourse(t *testing.T) {
	_, mock, tx := beginMockTx(t)
	mock.ExpectQuery(`SELECT id FROM word_cards WHERE LOWER\(TRIM\(word\)\) = \? AND LOWER\(COALESCE\(course_code, ''\)\) = \?`).
		WithArgs("hablar", "es_ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(`SELECT tc\.word_card_id\s+FROM training_cards tc\s+JOIN word_cards wc ON wc\.id = tc\.word_card_id`).
		WithArgs("hablar", "es_ru").
		WillReturnRows(sqlmock.NewRows([]string{"word_card_id"}))
	mock.ExpectQuery(`SELECT id\s+FROM word_cards wc\s+WHERE LOWER\(TRIM\(wc\.word\)\) = \?`).
		WithArgs("hablar", "hablar", "es_ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1020)))
	mock.ExpectRollback()

	got, err := lookupWordCardIDForVerbLemma(tx, "hablar", "es_ru")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if got != 1020 {
		t.Fatalf("got word_card_id=%d, want 1020", got)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
