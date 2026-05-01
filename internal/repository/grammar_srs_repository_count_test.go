package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func TestCountDueOrNewTheoryBlocks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewGrammarSRSRepository(db, zap.NewNop())

	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)
	past := now.Add(-time.Hour)

	mock.ExpectQuery(`SELECT theory_block_id, next_review_at FROM grammar_theory_memory`).
		WithArgs(int64(1), "es", "es", "b1", "b2", "b3").
		WillReturnRows(sqlmock.NewRows([]string{"theory_block_id", "next_review_at"}).
			AddRow("b1", past).
			AddRow("b2", future))

	n, err := repo.CountDueOrNewTheoryBlocks(1, "es", "es", []string{"b1", "b2", "b3"}, now)
	if err != nil {
		t.Fatalf("CountDueOrNewTheoryBlocks: %v", err)
	}
	// b1 due, b2 not due, b3 no row -> count 2
	if n != 2 {
		t.Fatalf("got %d want 2", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
