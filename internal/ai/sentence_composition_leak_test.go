package ai

import "testing"

func TestLeakedTargetWord(t *testing.T) {
	words := []GenSentenceWord{
		{Lemma: "gato", Translation: "кот"},
		{Lemma: "comer", Translation: "есть"},
		{Lemma: "agua", Translation: "вода"},
		{Lemma: "leer", Translation: "читать"},
	}
	cases := []struct {
		name   string
		prompt string
		want   string // leaked lemma, or "" if clean
	}{
		{"clean", "Кот пьёт воду.", ""},
		{"bare lemma leak", "gato пьёт воду.", "gato"},
		{"inflected noun leak", "У меня есть gatos.", "gato"},
		{"verb conjugated leak", "Он come хлеб.", "comer"},
		{"lemma verbatim verb", "Мне нравится leer книгу.", "leer"},
		{"punctuation-hugged leak", "Это (agua).", "agua"},
		{"clean cyrillic only", "Собака и друг большие.", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := leakedTargetWord(tc.prompt, words); got != tc.want {
				t.Fatalf("leakedTargetWord(%q) = %q, want %q", tc.prompt, got, tc.want)
			}
		})
	}
}
