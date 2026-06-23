package readingcms

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func saveTestDraft(t *testing.T, svc *Service, meta DraftMeta, segments int) *TextDocument {
	t.Helper()
	textID := meta.TextID
	if textID == "" {
		textID = newTextID(meta.TargetLanguage, meta.Level)
		meta.TextID = textID
	}
	segList := make([]map[string]interface{}, 0, segments)
	for i := 0; i < segments; i++ {
		speakerID := "speaker_a"
		if i%2 == 1 {
			speakerID = "speaker_b"
		}
		text := "Hello world."
		if i == 1 {
			text = "Second line here."
		}
		segList = append(segList, map[string]interface{}{
			"segment_id":          fmt.Sprintf("s%d", i+1),
			"speaker_id":          speakerID,
			"voice_id":            defaultVoiceID(meta.TargetLanguage, speakerID),
			"text":                text,
			"text_translation_ru": "",
			"audio_rel_path":      fmt.Sprintf("assets/reading/%s/seg_%02d_%s.mp3", textID, i+1, speakerID),
			"tokens":              Tokenize(text),
		})
	}
	doc := &TextDocument{
		ID:             textID,
		CategoryID:     categoryID(meta.TargetLanguage, meta.Level),
		Title:          meta.Title,
		Level:          meta.Level,
		TargetLanguage: meta.TargetLanguage,
		ReadingPassage: map[string]interface{}{
			"title":                   meta.Title,
			"level":                   meta.Level,
			"target_language":         meta.TargetLanguage,
			"estimated_minutes":       6,
			"segments":                segList,
			"vocab_focus":             []interface{}{},
			"comprehension_questions": []interface{}{},
		},
	}
	total, withAudio, audioSt := AudioStats(doc, svc.paths.StagingDir(textID))
	meta.SegmentsTotal = total
	meta.SegmentsWithAudio = withAudio
	meta.AudioStatus = audioSt
	if err := svc.store.SaveDraft(meta, doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestImportPlainTextAndPublish(t *testing.T) {
	root := t.TempDir()
	courseDir := filepath.Join(root, "courses", "english-grammar")
	if err := os.MkdirAll(filepath.Join(courseDir, "reading", "texts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(courseDir, "reading", "index.json"), []byte(`{"version":"1.0.0","categories":{},"texts":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	svc, err := NewService(root)
	if err != nil {
		t.Fatal(err)
	}
	meta := DraftMeta{
		CourseCode:     "en_ru",
		Title:          "Test story",
		Level:          "A2",
		Format:         "dialogue",
		TargetLanguage: "en",
		Status:         StatusApproved,
		Origin:         OriginManualText,
	}
	doc := saveTestDraft(t, svc, meta, 2)
	meta.TextID = doc.ID
	if CountSegments(doc) != 2 {
		t.Fatalf("segments=%d want 2", CountSegments(doc))
	}

	approved, err := svc.Approve(meta.TextID)
	if err != nil {
		t.Fatal(err)
	}
	if approved.Status != StatusApproved {
		t.Fatalf("status=%s", approved.Status)
	}

	staging := svc.paths.StagingDir(meta.TextID)
	segs := passageSegments(doc.ReadingPassage)
	for i, seg := range segs {
		rel := seg["audio_rel_path"].(string)
		abs := filepath.Join(staging, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("fake"), 0o644); err != nil {
			t.Fatal(err)
		}
		_ = i
	}
	doc2, err := svc.store.GetDocument(meta.TextID)
	if err != nil {
		t.Fatal(err)
	}
	_, _, audioSt := AudioStats(doc2, staging)
	if audioSt != AudioReady {
		t.Fatalf("audio=%s", audioSt)
	}
	meta.Status = StatusAudioReady
	if err := svc.store.SaveDraft(meta, doc2); err != nil {
		t.Fatal(err)
	}

	published, err := svc.Publish(meta.TextID, false)
	if err != nil {
		t.Fatal(err)
	}
	if published.Status != StatusPublished {
		t.Fatalf("status=%s", published.Status)
	}
	path := filepath.Join(courseDir, "reading", "texts", meta.TextID+".json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("published file missing: %v", err)
	}

	if err := svc.DeletePublished("en_ru", meta.TextID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected deleted published file, err=%v", err)
	}
}
