package ai

import "testing"

func TestPreparePrompt(t *testing.T) {
	raw := `Line1\nLine2 {{pair}} {{native_lang}} {{target_lang}}`
	got := PreparePrompt(raw, "ru", "en", "ru-en")
	want := "Line1\nLine2 ru-en ru en"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderLearningPromptTemplate(t *testing.T) {
	s := RenderLearningPromptTemplate("{{pair}}|{{native_lang}}|{{target_lang}}", "ru", "es", "ru-es")
	if s != "ru-es|ru|es" {
		t.Fatalf("unexpected %q", s)
	}
}
