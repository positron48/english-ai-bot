package spanishverbs

import (
	"encoding/json"
	"strings"
)

// VerbClassFromLemmaMetadataJSON returns top-level verb_class from verb_lemmas.metadata_json.
func VerbClassFromLemmaMetadataJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return ""
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &top); err != nil || top == nil {
		return ""
	}
	if v, ok := top["verb_class"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			return strings.TrimSpace(strings.ToLower(s))
		}
	}
	return ""
}

// AllowedTemplateIDsFromLemmaMetadataJSON returns allowed_template_ids (string codes) if set.
func AllowedTemplateIDsFromLemmaMetadataJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return nil
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &top); err != nil || top == nil {
		return nil
	}
	rawIDs, ok := top["allowed_template_ids"]
	if !ok {
		return nil
	}
	var arr []string
	if json.Unmarshal(rawIDs, &arr) == nil {
		out := make([]string, 0, len(arr))
		for _, s := range arr {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
