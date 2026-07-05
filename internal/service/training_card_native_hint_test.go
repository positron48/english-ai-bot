package service

import (
	"strings"
	"testing"
)

func TestNativeLookupHintWord(t *testing.T) {
	if got := NativeLookupHintWord("  капуста  "); got != "капуста" {
		t.Fatalf("got %q", got)
	}
	if got := NativeLookupHintWord("col"); got != "" {
		t.Fatalf("expected empty for Latin, got %q", got)
	}
}

func TestBuildTrainingCardNativeHintInstruction(t *testing.T) {
	block := BuildTrainingCardNativeHintInstruction("капуста", "Russian")
	if block == "" || !ContainsCyrillic(block) {
		t.Fatalf("expected non-empty Cyrillic hint block, got %q", block)
	}
	if !strings.Contains(block, "капуста") {
		t.Fatalf("expected hint to mention lookup word, got %q", block)
	}
	if BuildTrainingCardNativeHintInstruction("col", "Russian") != "" {
		t.Fatal("expected empty block for Latin input")
	}
}
