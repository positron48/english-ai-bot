package service

import (
	"encoding/json"
	"testing"

	"tgbot-skeleton/internal/repository"
)

func TestShuffleVerbQueue_deterministicSeed(t *testing.T) {
	a := repository.VerbQueueCard{UserVerbCardID: 1, PromptJSON: `{"question":"a"}`}
	b := repository.VerbQueueCard{UserVerbCardID: 2, PromptJSON: `{"question":"b"}`}
	c := repository.VerbQueueCard{UserVerbCardID: 3, PromptJSON: `{"question":"c"}`}
	q1 := []repository.VerbQueueCard{a, b, c}
	q2 := []repository.VerbQueueCard{a, b, c}
	ShuffleVerbQueue(q1, 12345)
	ShuffleVerbQueue(q2, 12345)
	for i := range q1 {
		if q1[i].UserVerbCardID != q2[i].UserVerbCardID {
			t.Fatalf("same seed should give same order, i=%d %v vs %v", i, q1[i], q2[i])
		}
	}
}

func TestSpreadAdjacentDuplicateVerbPromptKeys(t *testing.T) {
	dup := func(q string) repository.VerbQueueCard {
		pj, _ := json.Marshal(map[string]string{"question": q})
		return repository.VerbQueueCard{
			PromptJSON:      string(pj),
			AnswerJSON:      `{"surface_form":"voy"}`,
			DistractorsJSON: "[]",
		}
	}
	a := dup("Yo ... (ir)")
	b := dup("Yo ... (ir)")
	c := dup("Tú ... (ser)")
	queue := []repository.VerbQueueCard{a, b, c}
	out := SpreadAdjacentDuplicateVerbPromptKeys(queue)
	if verbQueuePromptAnswerKey(out[0]) == verbQueuePromptAnswerKey(out[1]) {
		t.Fatalf("adjacent duplicates remain: %#v vs %#v", out[0], out[1])
	}
}
