package wordsetimport

import (
	"testing"

	"tgbot-skeleton/internal/config"
)

func TestAssertImportProfile(t *testing.T) {
	t.Parallel()

	cfgEN := &config.Config{
		Learning: config.LearningConfig{
			TargetLang: "en",
			AppCode:    "english",
		},
	}
	if err := assertImportProfile(cfgEN, "en", "resources/wordsets/english_word_freq.csv"); err != nil {
		t.Fatalf("expected english profile to pass, got error: %v", err)
	}
	if err := assertImportProfile(cfgEN, "en", "resources/wordsets/spanish_word_freq.csv"); err == nil {
		t.Fatal("expected english profile to reject spanish path")
	}

	cfgES := &config.Config{
		Learning: config.LearningConfig{
			TargetLang: "es",
			AppCode:    "spanish",
		},
	}
	if err := assertImportProfile(cfgES, "es", "resources/wordsets/spanish_word_freq.csv"); err != nil {
		t.Fatalf("expected spanish profile to pass, got error: %v", err)
	}
	if err := assertImportProfile(cfgES, "es", "resources/wordsets/english_word_freq.csv"); err == nil {
		t.Fatal("expected spanish profile to reject english path")
	}
}

func TestParseRankRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		title string
		ok    bool
		s     int
		e     int
	}{
		{title: "Core Verbs — Top 50 (Ranks 1–50)", ok: true, s: 1, e: 50},
		{title: "Verbos esenciales — Top 50 (rangos 101-150)", ok: true, s: 101, e: 150},
		{title: "No ranks here", ok: false},
	}
	for _, tc := range tests {
		s, e, ok := parseRankRange(tc.title)
		if ok != tc.ok {
			t.Fatalf("title=%q: expected ok=%v, got %v", tc.title, tc.ok, ok)
		}
		if ok && (s != tc.s || e != tc.e) {
			t.Fatalf("title=%q: expected range=%d-%d, got %d-%d", tc.title, tc.s, tc.e, s, e)
		}
	}
}
