package ai

import "testing"

func TestNormalizedSentenceAnswer_IgnoresPresentationOnlyDifferences(t *testing.T) {
	base := NormalizedSentenceAnswer("El gato bebe agua.")
	for _, answer := range []string{"el gato bebe agua", " El gato   bebe agua! ", "¿El gato bebe agua?"} {
		if got := NormalizedSentenceAnswer(answer); got != base {
			t.Errorf("NormalizedSentenceAnswer(%q) = %q, want %q", answer, got, base)
		}
	}
}

func TestNewExactSentenceGrade(t *testing.T) {
	grade := NewExactSentenceGrade("El gato bebe agua")
	if grade.ErrorCount != 0 || grade.Outcome != "star" || grade.CorrectedES != "El gato bebe agua" {
		t.Fatalf("unexpected exact grade: %+v", grade)
	}
}
