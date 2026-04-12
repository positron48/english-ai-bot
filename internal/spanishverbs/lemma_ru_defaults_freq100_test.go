package spanishverbs

import "testing"

func TestDefaultRuGloss_coversFreq100Verbs(t *testing.T) {
	for _, lem := range Freq100VerbLemmas {
		if g := DefaultRuGloss(lem); g == "" {
			t.Fatalf("missing DefaultRuGloss for freq100 lemma %q", lem)
		}
	}
}
