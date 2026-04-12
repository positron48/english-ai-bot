package models

import (
	"reflect"
	"testing"
)

func TestSpanishVerbScopesThroughIndex(t *testing.T) {
	got := SpanishVerbScopesThroughIndex(1)
	want := []string{"es.presente.indicativo", "es.pretérito.indicativo"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestInferSpanishVerbProgressionIndex(t *testing.T) {
	if g := InferSpanishVerbProgressionIndex([]string{"es.presente.indicativo"}); g != 0 {
		t.Fatalf("got %d", g)
	}
	if g := InferSpanishVerbProgressionIndex([]string{"es.presente.indicativo", "es.pretérito.indicativo", "es.futuro.indicativo"}); g != 1 {
		// not cumulative (skipped imperfecto) → prefix breaks after pretérito
		t.Fatalf("non-cumulative got %d want 1", g)
	}
	full := SpanishVerbScopesThroughIndex(len(SpanishVerbScopeLadder()) - 1)
	if g := InferSpanishVerbProgressionIndex(full); g != len(SpanishVerbScopeLadder())-1 {
		t.Fatalf("full ladder got %d", g)
	}
}
