package spanishverbs

import (
	"encoding/json"
	"strings"
)

// MergeRuGlossIntoLemmaMetadataJSON merges ru.gloss (and gloss_source) into existing verb_lemmas.metadata_json.
func MergeRuGlossIntoLemmaMetadataJSON(existingJSON, ruGloss, source string) (string, error) {
	ruGloss = strings.TrimSpace(ruGloss)
	if ruGloss == "" {
		return existingJSON, nil
	}
	var m map[string]interface{}
	if strings.TrimSpace(existingJSON) != "" && existingJSON != "{}" {
		_ = json.Unmarshal([]byte(existingJSON), &m)
	}
	if m == nil {
		m = map[string]interface{}{}
	}
	var ru map[string]interface{}
	if raw, ok := m["ru"]; ok {
		if sub, ok2 := raw.(map[string]interface{}); ok2 {
			ru = sub
		}
	}
	if ru == nil {
		ru = map[string]interface{}{}
	}
	ru["gloss"] = ruGloss
	if strings.TrimSpace(source) != "" {
		ru["gloss_source"] = strings.TrimSpace(source)
	}
	m["ru"] = ru
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MergeVerbClassIntoLemmaMetadataJSON sets top-level verb_class (and optional class_source).
func MergeVerbClassIntoLemmaMetadataJSON(existingJSON, verbClass, source string) (string, error) {
	verbClass = strings.TrimSpace(strings.ToLower(verbClass))
	if verbClass == "" {
		return existingJSON, nil
	}
	var m map[string]interface{}
	if strings.TrimSpace(existingJSON) != "" && existingJSON != "{}" {
		_ = json.Unmarshal([]byte(existingJSON), &m)
	}
	if m == nil {
		m = map[string]interface{}{}
	}
	m["verb_class"] = verbClass
	if strings.TrimSpace(source) != "" {
		m["verb_class_source"] = strings.TrimSpace(source)
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MergeAllowedTemplateIDsIntoLemmaMetadataJSON sets top-level allowed_template_ids (string slice).
func MergeAllowedTemplateIDsIntoLemmaMetadataJSON(existingJSON string, ids []string, source string) (string, error) {
	if len(ids) == 0 {
		return existingJSON, nil
	}
	var m map[string]interface{}
	if strings.TrimSpace(existingJSON) != "" && existingJSON != "{}" {
		_ = json.Unmarshal([]byte(existingJSON), &m)
	}
	if m == nil {
		m = map[string]interface{}{}
	}
	uniq := make([]string, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		k := strings.ToLower(id)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		return existingJSON, nil
	}
	m["allowed_template_ids"] = uniq
	if strings.TrimSpace(source) != "" {
		m["allowed_template_ids_source"] = strings.TrimSpace(source)
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
