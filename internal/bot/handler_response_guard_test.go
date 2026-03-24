package bot

import (
	"strings"
	"testing"
)

func TestLooksLikeWordInfoJSON(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{
			name: "word info json shape",
			in: `{
  "error": false,
  "lemma": "mesa",
  "pos": "noun",
  "examples": [{"example_target":"mesa","gloss_native":"стол"}]
}`,
			want: true,
		},
		{
			name: "non json text",
			in:   "Translated Version:\nMesa.",
			want: false,
		},
		{
			name: "json without required keys",
			in:   `{"status":"ok","message":"hello"}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeWordInfoJSON(tt.in)
			if got != tt.want {
				t.Fatalf("looksLikeWordInfoJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderWordInfoJSONAsMarkdown(t *testing.T) {
	raw := `{
  "error": false,
  "hint": "",
  "input_word": "стол",
  "lemma": "mesa",
  "pos": "noun",
  "transcription": "ˈmesa",
  "definition_native": "стол",
  "examples": [
    {"example_target":"La mesa es grande.","gloss_native":"Стол большой."},
    {"example_target":"Ponlo en la mesa.","gloss_native":"Положи это на стол."}
  ]
}`

	rendered, ok := renderWordInfoJSONAsMarkdown(raw, "es")
	if !ok {
		t.Fatal("expected JSON conversion to markdown to succeed")
	}
	if strings.Contains(rendered, `"lemma"`) || strings.HasPrefix(strings.TrimSpace(rendered), "{") {
		t.Fatalf("expected markdown card, got raw JSON: %q", rendered)
	}
	if !strings.Contains(strings.ToLower(rendered), "mesa") {
		t.Fatalf("expected rendered card to mention lemma, got: %q", rendered)
	}
}

