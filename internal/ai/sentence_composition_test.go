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

func TestSwapFirstSpanishArticle(t *testing.T) {
	tests := map[string]string{
		"Veo el libro.":       "Veo un libro.",
		"Una amiga lee.":      "La amiga lee.",
		"Busco las llaves.":   "Busco unas llaves.",
		"Quiero unos libros.": "Quiero los libros.",
	}
	for input, want := range tests {
		got, ok := swapFirstSpanishArticle(input)
		if !ok || got != want {
			t.Errorf("swapFirstSpanishArticle(%q) = %q, %v; want %q, true", input, got, ok, want)
		}
	}
	if got, ok := swapFirstSpanishArticle("Bebo agua."); ok || got != "Bebo agua." {
		t.Errorf("unexpected swap without article: %q, %v", got, ok)
	}
}

func TestUsefulArticleExplanation(t *testing.T) {
	if !usefulArticleExplanation("Книгу уже обсуждали, поэтому оба собеседника знают, о какой книге речь — нужен el.") {
		t.Fatal("expected contextual explanation to be useful")
	}
	for _, bad := range []string{"", "Здесь нужен el, а не un.", "Нужен el, потому что книга конкретная."} {
		if usefulArticleExplanation(bad) {
			t.Errorf("expected explanation to be rejected: %q", bad)
		}
	}
}

func TestArticleContextSmokeHelpers(t *testing.T) {
	if !containsSpanishIndefiniteArticle("Compro un libro.") || containsSpanishIndefiniteArticle("Leo el libro.") {
		t.Fatal("indefinite article detection is incorrect")
	}
	if !containsRussianDemonstrative("Впервые вижу эту книгу.") || containsRussianDemonstrative("Впервые вижу книгу.") {
		t.Fatal("Russian demonstrative detection is incorrect")
	}
}

func TestUsableSentenceClarification(t *testing.T) {
	for _, good := range []string{"", "Собеседники уже обсуждали эту книгу."} {
		if !usableSentenceClarification(good) {
			t.Errorf("expected usable context: %q", good)
		}
	}
	for _, bad := range []string{"Речь о конкретной книге.", "Используй el.", "Нужен определённый артикль."} {
		if usableSentenceClarification(bad) {
			t.Errorf("expected rejected context: %q", bad)
		}
	}
}

func TestSentenceExplanationLanguageMatches(t *testing.T) {
	if !sentenceExplanationLanguageMatches("Они едят апельсин.", "Форма глагола должна согласоваться с vosotros.") {
		t.Fatal("expected Russian explanation to match")
	}
	if sentenceExplanationLanguageMatches("Они едят апельсин.", "El verbo debe concordar con el sujeto.") {
		t.Fatal("expected Spanish explanation to be rejected")
	}
	if !sentenceExplanationLanguageMatches("Они едят апельсин.", "") {
		t.Fatal("empty explanation must remain valid")
	}
}

func TestHasSpanishArticle(t *testing.T) {
	for _, sentence := range []string{"Ellos comen una naranja.", "Leo el libro.", "Veo las casas."} {
		if !hasSpanishArticle(sentence) {
			t.Errorf("expected article in %q", sentence)
		}
	}
	if hasSpanishArticle("Bebemos agua.") {
		t.Fatal("unexpected article")
	}
}
