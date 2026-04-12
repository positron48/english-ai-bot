package service

import (
	"strings"
	"testing"

	"tgbot-skeleton/internal/spanishverbs"
)

func TestRuntimeVerbExample_masksForCloze(t *testing.T) {
	es, _ := spanishverbs.GenerateVerbExamplePair(101, "hablar", "indicativo", "presente", "2", "singular", "hablas", "говорить", "", nil)
	masked := MaskClozeVerbSurfaceInQuestion(es, "hablas")
	if !strings.Contains(masked, ClozeBlankPlaceholder) {
		t.Fatalf("expected blank in masked, got %q", masked)
	}
}
