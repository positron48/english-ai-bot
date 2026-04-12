package spanishverbs

import (
	"encoding/json"
	"strings"
)

// RuGlossFromLemmaMetadataJSON extracts Russian gloss for a Spanish verb lemma from verb_lemmas.metadata_json.
// Supported shapes: {"ru":{"gloss":"..."}}, {"ru_gloss":"..."}, legacy {"gloss_ru":"..."}.
func RuGlossFromLemmaMetadataJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return ""
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &top); err != nil || top == nil {
		return ""
	}
	if v, ok := top["ru_gloss"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			return strings.TrimSpace(s)
		}
	}
	if v, ok := top["gloss_ru"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			return strings.TrimSpace(s)
		}
	}
	if v, ok := top["ru"]; ok {
		var ru struct {
			Gloss string `json:"gloss"`
		}
		if json.Unmarshal(v, &ru) == nil {
			return strings.TrimSpace(ru.Gloss)
		}
	}
	return ""
}
