package repository

import (
	"testing"
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
