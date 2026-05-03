package web

import (
	"strings"
	"testing"
)

func TestHydrateVerbClozePrompt_FillsSentenceWhenQuestionEmpty(t *testing.T) {
	prompt := map[string]interface{}{
		"type":                "cloze_form",
		"lemma":               "hablar",
		"mood":                "indicativo",
		"tense":               "presente",
		"person":              "3",
		"number":              "plural",
		"ru_gloss":            "говорить",
		"example_mode":        "runtime",
		"example_source":      "runtime_templates",
		"allowed_template_ids": []interface{}{},
	}
	hydrateVerbClozePrompt(prompt, "hablan", 42)

	q := promptString(prompt, "question")
	if q == "" {
		t.Fatal("expected non-empty spanish question")
	}
	if !strings.Contains(strings.ToLower(q), "hablan") {
		t.Fatalf("expected spanish sentence to contain surface form, got %q", q)
	}
	if strings.TrimSpace(promptString(prompt, "example_translation")) == "" {
		t.Fatal("expected russian example_translation line")
	}
}

func TestHydrateVerbClozePrompt_KeepsArtifactQuestionUnchanged(t *testing.T) {
	prompt := map[string]interface{}{
		"question":            "Ellos ___ en la plaza.",
		"example_translation":   "Они разговаривают на площади.",
		"lemma":                 "hablar",
		"mood":                  "indicativo",
		"tense":                 "presente",
		"person":                "3",
		"number":                "plural",
	}
	hydrateVerbClozePrompt(prompt, "hablan", 99)
	if promptString(prompt, "question") != "Ellos ___ en la plaza." {
		t.Fatalf("question should stay as in artifact")
	}
	if promptString(prompt, "example_translation") != "Они разговаривают на площади." {
		t.Fatalf("translation should stay as in artifact")
	}
}
