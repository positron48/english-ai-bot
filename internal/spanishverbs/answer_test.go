package spanishverbs

import "testing"

func TestAcceptedVerbAnswer(t *testing.T) {
	for _, c := range []struct {
		expected, answer, mood, tense string
		want                          bool
	}{
		{"hablara", "hablase", "subjuntivo", "imperfecto", true},
		{"hubiéramos hablado", "hubiésemos hablado", "subjuntivo", "pluscuamperfecto", true},
		{"hablaras", "hablase", "subjuntivo", "imperfecto", false},
		{"habláramos", "hablasemos", "subjuntivo", "imperfecto", false},
		{"barréis", "barreis", "indicativo", "presente", false},
		{"barréis", " BARRÉIS ", "indicativo", "presente", true},
	} {
		if got := AcceptedVerbAnswer(c.expected, c.answer, c.mood, c.tense); got != c.want {
			t.Fatalf("%+v got %v", c, got)
		}
	}
}
