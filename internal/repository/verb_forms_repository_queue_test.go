package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"tgbot-skeleton/internal/models"
)

func TestRoundRobinVerbNewCards_mixedLemmas(t *testing.T) {
	// Two lemmas, three cards each; maxPick 4 → one from each verb first, then one more from first.
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
	// Order of lemmas follows first appearance in pool: 10 then 20.
	want := []int64{1, 11, 2, 12}
	for i := range want {
		if got[i].UserVerbCardID != want[i] {
			t.Fatalf("i=%d got id=%d want %d", i, got[i].UserVerbCardID, want[i])
		}
	}
}

func TestRoundRobinVerbNewCards_singleLemma(t *testing.T) {
	pool := []verbNewQueueRow{
		{card: VerbQueueCard{UserVerbCardID: 1}, wordCardID: 10},
		{card: VerbQueueCard{UserVerbCardID: 2}, wordCardID: 10},
	}
	got := roundRobinVerbNewCards(pool, 10)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
}

func TestGetVerbQueue_IncludesDueWhenMaxCardsEqualsMaxNew(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())

	now := time.Now()
	mock.ExpectQuery(`SELECT uvc\.id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(7), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
			AddRow(int64(101), models.VerbCardTypeCloze, "{}", `{"surface_form":"hablo"}`, "[]"))

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(7), models.VerbCardTypeCloze, 600).
		WillReturnRows(sqlmock.NewRows([]string{"id", "card_type", "prompt_json", "answer_json", "distractors_json", "word_card_id"}).
			AddRow(int64(101), models.VerbCardTypeCloze, "{}", `{"surface_form":"hablo"}`, "[]", int64(10)).
			AddRow(int64(201), models.VerbCardTypeCloze, "{}", `{"surface_form":"como"}`, "[]", int64(20)))

	queue, err := repo.GetVerbQueue(7, now, 30, 30)
	if err != nil {
		t.Fatalf("GetVerbQueue error: %v", err)
	}
	if len(queue) < 2 {
		t.Fatalf("expected at least due+new cards, got %d", len(queue))
	}
	if queue[0].UserVerbCardID != 101 {
		t.Fatalf("expected due card first in queue, got %d", queue[0].UserVerbCardID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetVerbQueue_NewPoolSkipsAlreadyIncludedDue(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewVerbFormsRepository(db, zap.NewNop())

	now := time.Now()
	mock.ExpectQuery(`SELECT uvc\.id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(9), models.VerbCardTypeCloze, now, models.MaxDuePoolSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "card_type", "prompt_json", "answer_json", "distractors_json"}).
			AddRow(int64(500), models.VerbCardTypeCloze, "{}", `{"surface_form":"voy"}`, "[]"))

	mock.ExpectQuery(`SELECT uvc\.id, vtc\.card_type, vtc\.prompt_json`).
		WithArgs(int64(9), models.VerbCardTypeCloze, 60).
		WillReturnRows(sqlmock.NewRows([]string{"id", "card_type", "prompt_json", "answer_json", "distractors_json", "word_card_id"}).
			AddRow(int64(500), models.VerbCardTypeCloze, "{}", `{"surface_form":"voy"}`, "[]", int64(55)).
			AddRow(int64(501), models.VerbCardTypeCloze, "{}", `{"surface_form":"vas"}`, "[]", int64(55)))

	queue, err := repo.GetVerbQueue(9, now, 10, 2)
	if err != nil {
		t.Fatalf("GetVerbQueue error: %v", err)
	}
	if len(queue) != 2 {
		t.Fatalf("expected 2 cards in queue, got %d", len(queue))
	}
	if queue[0].UserVerbCardID != 500 || queue[1].UserVerbCardID != 501 {
		t.Fatalf("unexpected queue order/content: %+v", queue)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
