package models

import "testing"

func TestNormalizeReportCategory(t *testing.T) {
	if got := NormalizeReportCategory("word_training", "bad_audio"); got != "bad_audio" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeReportCategory("word_training", "invalid"); got != "other" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeReportCategory("grammar_training", ""); got != "other" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeReportCategory("reading_text", "bad_audio"); got != "bad_audio" {
		t.Fatalf("got %q", got)
	}
}
