package service

import "testing"

func TestNormalizeTargetVerbDisplay(t *testing.T) {
	t.Run("English keeps to prefix for verbs", func(t *testing.T) {
		got := normalizeTargetVerbDisplay("en", "verb", "to make")
		if got != "to make" {
			t.Fatalf("got %q, want to make", got)
		}
	})

	t.Run("Spanish verbo strips to prefix", func(t *testing.T) {
		got := normalizeTargetVerbDisplay("es", "verbo", "to hablar")
		if got != "hablar" {
			t.Fatalf("got %q, want hablar", got)
		}
	})

	t.Run("Spanish strips legacy to prefix without POS", func(t *testing.T) {
		got := normalizeTargetVerbDisplay("es", "", "to comer")
		if got != "comer" {
			t.Fatalf("got %q, want comer", got)
		}
	})

	t.Run("Spanish noun keeps word without to", func(t *testing.T) {
		got := normalizeTargetVerbDisplay("es", "noun", "casa")
		if got != "casa" {
			t.Fatalf("got %q, want casa", got)
		}
	})
}
