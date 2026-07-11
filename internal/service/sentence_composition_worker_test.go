package service

import (
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
)

func TestShouldGenerateSentenceSet(t *testing.T) {
	today := "2026-07-08"
	yesterday := "2026-07-07"

	cases := []struct {
		name   string
		latest *models.SentenceSet
		want   bool
	}{
		{name: "no prior set", latest: nil, want: true},
		{name: "ready blocks", latest: &models.SentenceSet{Status: models.SentenceSetReady, GenerationDate: yesterday}, want: false},
		{name: "started blocks", latest: &models.SentenceSet{Status: models.SentenceSetStarted, GenerationDate: yesterday}, want: false},
		{name: "completed today allows", latest: &models.SentenceSet{Status: models.SentenceSetCompleted, GenerationDate: today}, want: true},
		{name: "completed yesterday allows", latest: &models.SentenceSet{Status: models.SentenceSetCompleted, GenerationDate: yesterday}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldGenerateSentenceSet(tc.latest); got != tc.want {
				t.Fatalf("shouldGenerateSentenceSet() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHumanizeScopes(t *testing.T) {
	got := humanizeScopes([]string{"es.presente.indicativo", "es.preterito", "weird"})
	want := []string{"presente (indicativo)", "preterito", "weird"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("scope[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLearningForCourseCode(t *testing.T) {
	lc := learningForCourseCode(config.LearningConfig{TargetLang: "en"}, "es_ru")
	if lc.TargetLang != "es" || lc.NativeLang != "ru" || lc.Pair != "ru-es" {
		t.Fatalf("unexpected learning config: %+v", lc)
	}
	// Malformed code falls back unchanged.
	fb := config.LearningConfig{TargetLang: "en", Pair: "ru-en"}
	if got := learningForCourseCode(fb, "garbage"); got.TargetLang != "en" || got.Pair != "ru-en" {
		t.Fatalf("expected fallback, got %+v", got)
	}
}

func TestSentenceFocusWordCount(t *testing.T) {
	tests := []struct {
		total, sentences, want int
	}{
		{80, 20, 20},
		{40, 20, 13},
		{12, 20, 4},
		{2, 20, 1},
		{0, 20, 0},
	}
	for _, tc := range tests {
		if got := sentenceFocusWordCount(tc.total, tc.sentences); got != tc.want {
			t.Errorf("sentenceFocusWordCount(%d, %d) = %d, want %d", tc.total, tc.sentences, got, tc.want)
		}
	}
}
