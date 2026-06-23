package readingcms

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func normalizeLevel(level string) (string, error) {
	lv := strings.ToUpper(strings.TrimSpace(level))
	switch lv {
	case "A0", "A1", "A2", "B1", "B2", "C1":
		return lv, nil
	default:
		return "", fmt.Errorf("level must be one of A0..C1")
	}
}

func normalizeFormat(format string) string {
	f := strings.ToLower(strings.TrimSpace(format))
	if f == "narrative" {
		return "narrative"
	}
	return "dialogue"
}

func newTextID(targetLang, level string) string {
	return fmt.Sprintf("cms_%s_%s_%d", targetLang, strings.ToLower(level), time.Now().Unix())
}

func categoryID(targetLang, level string) string {
	return fmt.Sprintf("%s_%s", targetLang, strings.ToLower(level))
}

func passageSegments(passage map[string]interface{}) []map[string]interface{} {
	if passage == nil {
		return nil
	}
	raw := passage["segments"]
	switch v := raw.(type) {
	case []map[string]interface{}:
		return v
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			seg, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			out = append(out, seg)
		}
		return out
	default:
		return nil
	}
}

func CountSegments(doc *TextDocument) int {
	if doc == nil {
		return 0
	}
	return len(passageSegments(doc.ReadingPassage))
}

func AudioStats(doc *TextDocument, stagingRoot string) (total, withAudio int, status string) {
	segs := passageSegments(doc.ReadingPassage)
	total = len(segs)
	if total == 0 {
		return 0, 0, AudioNone
	}
	for _, seg := range segs {
		rel, _ := seg["audio_rel_path"].(string)
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		abs := filepath.Join(stagingRoot, filepath.FromSlash(rel))
		if st, err := os.Stat(abs); err == nil && !st.IsDir() && st.Size() > 0 {
			withAudio++
		}
	}
	switch {
	case withAudio == 0:
		status = AudioNone
	case withAudio < total:
		status = AudioPartial
	default:
		status = AudioReady
	}
	return total, withAudio, status
}
