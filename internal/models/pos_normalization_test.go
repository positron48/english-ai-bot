package models

import "testing"

func TestIsVerbPOS_auxiliarySpanishTag(t *testing.T) {
	for _, raw := range []string{"aux", "AUX", "Aux"} {
		if !IsVerbPOS(raw) {
			t.Fatalf("IsVerbPOS(%q) = false, want true", raw)
		}
	}
	if !IsVerbPOS("verb") || !IsVerbPOS("VERB") {
		t.Fatal("IsVerbPOS should still accept verb")
	}
	if IsVerbPOS("noun") || IsVerbPOS("") {
		t.Fatal("IsVerbPOS should reject non-verb-like POS")
	}
}
