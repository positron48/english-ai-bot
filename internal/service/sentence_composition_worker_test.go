package service

import (
	"testing"

	"tgbot-skeleton/internal/config"
)

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
