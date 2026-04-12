package spanishverbs

import "testing"

func TestBuildVerbTrainingClozeQuestion_indicativoPresente(t *testing.T) {
	got := BuildVerbTrainingClozeQuestion("1", "singular", "haber", "indicativo", "presente")
	want := "Yo ___ (haber, presente)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestClozeSpanishMoodTenseLabel_subjuntivo(t *testing.T) {
	got := ClozeSpanishMoodTenseLabel("subjuntivo", "presente")
	if got != "presente, subjuntivo" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainRussianVerbTrainingTranslation_noGuillemets(t *testing.T) {
	got := PlainRussianVerbTrainingTranslation("2", "singular", "говорить", "indicativo", "presente")
	if got == "" {
		t.Fatal("empty")
	}
	if got != "говоришь" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainRussianVerbTrainingTranslationForLemma_haber_ignoresWrongGloss(t *testing.T) {
	got := PlainRussianVerbTrainingTranslationForLemma("haber", "1", "plural", "существовать", "indicativo", "presente")
	if got != "имеем" {
		t.Fatalf("want Russian anchor for *haber*, got %q", got)
	}
}

func TestPlainRussianVerbTrainingHintLine_includesPronoun(t *testing.T) {
	got := PlainRussianVerbTrainingHintLine("haber", "3", "plural", "иметь", "indicativo", "presente")
	if got != "они имеют" {
		t.Fatalf("got %q", got)
	}
}
