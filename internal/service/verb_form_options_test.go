package service

import (
	"strings"
	"testing"
)

func TestBuildVerbFormMultipleChoiceOptions(t *testing.T) {
	opts := BuildVerbFormMultipleChoiceOptions("hablo", "hablar", 42)
	if len(opts) != VerbChoiceOptionCount {
		t.Fatalf("want %d options, got %d %#v", VerbChoiceOptionCount, len(opts), opts)
	}
	var hasHablo bool
	for _, o := range opts {
		if strings.EqualFold(o, "hablo") {
			hasHablo = true
		}
	}
	if !hasHablo {
		t.Fatalf("correct form missing: %#v", opts)
	}
	seen := map[string]bool{}
	for _, o := range opts {
		k := strings.ToLower(strings.TrimSpace(o))
		if seen[k] {
			t.Fatalf("duplicate option %q in %#v", o, opts)
		}
		seen[k] = true
	}
}

func TestBuildVerbFormMultipleChoiceOptions_noDigitSuffixArtifacts(t *testing.T) {
	opts := BuildVerbFormMultipleChoiceOptions("somos", "ser", 99)
	if len(opts) != VerbChoiceOptionCount {
		t.Fatalf("want %d options, got %d", VerbChoiceOptionCount, len(opts))
	}
	for _, o := range opts {
		if strings.HasSuffix(o, "3") && strings.HasPrefix(o, "somos") && o != "somos" {
			t.Fatalf("unexpected digit-suffix distractor %q in %#v", o, opts)
		}
	}
}

func TestMaskClozeVerbSurfaceInQuestion(t *testing.T) {
	got := MaskClozeVerbSurfaceInQuestion("Nosotros somos cada día.", "somos")
	if strings.Contains(strings.ToLower(got), "somos") {
		t.Fatalf("answer should be masked, got %q", got)
	}
	if !strings.Contains(got, ClozeBlankPlaceholder) {
		t.Fatalf("expected blank placeholder in %q", got)
	}
	got2 := MaskClozeVerbSurfaceInQuestion("Nosotros SOMOS cada día.", "somos")
	if strings.Contains(strings.ToLower(got2), "somos") {
		t.Fatalf("case-insensitive mask failed: %q", got2)
	}
}

func TestParseStringJSONArray(t *testing.T) {
	if ParseStringJSONArray("") != nil {
		t.Fatal("empty")
	}
	if got := ParseStringJSONArray(`["a","b"]`); len(got) != 2 || got[0] != "a" {
		t.Fatalf("%#v", got)
	}
}

func TestCapVerbMultipleChoiceOptions(t *testing.T) {
	opts := []string{"hablo", "habla", "hablas", "hablan", "hablamos", "habláis"}
	capped := CapVerbMultipleChoiceOptions("hablo", opts, 7)
	if len(capped) != VerbChoiceOptionCount {
		t.Fatalf("want %d options, got %d %#v", VerbChoiceOptionCount, len(capped), capped)
	}
	var hasCorrect bool
	for _, o := range capped {
		if strings.EqualFold(o, "hablo") {
			hasCorrect = true
		}
	}
	if !hasCorrect {
		t.Fatalf("correct answer missing: %#v", capped)
	}
}

func TestCapVerbMultipleChoiceOptions_preservesCanonicalCorrect(t *testing.T) {
	opts := []string{"HABLO", "habla", "hablas", "hablan", "hablamos"}
	capped := CapVerbMultipleChoiceOptions("hablo", opts, 11)
	if !strings.EqualFold(capped[0], "HABLO") && !containsFold(capped, "HABLO") {
		t.Fatalf("expected canonical correct preserved: %#v", capped)
	}
}

func containsFold(ss []string, want string) bool {
	for _, s := range ss {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}

func TestBuildVerbFormMultipleChoiceOptions_oSuffix(t *testing.T) {
	opts := BuildVerbFormMultipleChoiceOptions("hablo", "hablar", 1)
	if len(opts) != VerbChoiceOptionCount {
		t.Fatalf("got %d options", len(opts))
	}
	if !containsFold(opts, "habla") {
		t.Fatalf("expected -o->-a distractor in %#v", opts)
	}
}

func TestBuildVerbFormMultipleChoiceOptions_shortSurface(t *testing.T) {
	opts := BuildVerbFormMultipleChoiceOptions("a", "ir", 5)
	if len(opts) != VerbChoiceOptionCount {
		t.Fatalf("pathological input should still return %d options, got %d", VerbChoiceOptionCount, len(opts))
	}
}

func TestMaskClozeVerbSurfaceInQuestion_emptyInputs(t *testing.T) {
	if got := MaskClozeVerbSurfaceInQuestion("", "somos"); got != "" {
		t.Fatalf("empty question: got %q", got)
	}
	if got := MaskClozeVerbSurfaceInQuestion("Nosotros somos.", ""); got != "Nosotros somos." {
		t.Fatalf("empty surface: got %q", got)
	}
}

func TestParseStringJSONArray_invalidAndNull(t *testing.T) {
	if ParseStringJSONArray("null") != nil {
		t.Fatal("null should return nil")
	}
	if ParseStringJSONArray("{") != nil {
		t.Fatal("invalid JSON should return nil")
	}
}

func TestNormalizeVerbFormat_StripsEnglishToForSpanishTarget(t *testing.T) {
	svc := NewOptionsService(nil, nil, "es")
	got := svc.normalizeVerbFormat("to hablar", "verb", "ru_en")
	if got != "hablar" {
		t.Fatalf("got %q, want hablar", got)
	}

	enSvc := NewOptionsService(nil, nil, "en")
	if got := enSvc.normalizeVerbFormat("speak", "verb", "ru_en"); got != "to speak" {
		t.Fatalf("English target should keep infinitive marker, got %q", got)
	}
}
