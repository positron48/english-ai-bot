package readingcms

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreDraftLifecycle(t *testing.T) {
	root := t.TempDir()
	paths := NewPaths(root)
	store := NewStore(paths)
	if err := store.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	meta := DraftMeta{
		TextID:         "cms_en_a2_test",
		CourseCode:     "en_ru",
		Title:          "Lifecycle",
		Level:          "A2",
		Format:         "dialogue",
		TargetLanguage: "en",
		Status:         StatusDraft,
		Origin:         OriginManualText,
		AudioStatus:    AudioNone,
		LastJobLog:     "created in test",
	}
	doc := &TextDocument{
		ID:             meta.TextID,
		CategoryID:     "en_a2",
		Title:          meta.Title,
		Level:          meta.Level,
		TargetLanguage: meta.TargetLanguage,
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{
				{"segment_id": "s1", "text": "Hello"},
			},
		},
	}
	if err := store.SaveDraft(meta, doc); err != nil {
		t.Fatal(err)
	}

	got, ok, err := store.GetMeta(meta.TextID)
	if err != nil || !ok {
		t.Fatalf("meta missing: ok=%v err=%v", ok, err)
	}
	if got.Title != meta.Title {
		t.Fatalf("title=%q", got.Title)
	}

	docPath := paths.DraftDocPath(meta.TextID)
	if _, err := os.Stat(docPath); err != nil {
		t.Fatalf("draft doc missing: %v", err)
	}

	list, err := store.ListDrafts(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("draft count=%d", len(list))
	}

	staging := filepath.Join(paths.StagingDir(meta.TextID), "assets", "reading", meta.TextID, "seg_01.mp3")
	if err := os.MkdirAll(filepath.Dir(staging), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staging, []byte("mp3"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteDraft(meta.TextID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(docPath); !os.IsNotExist(err) {
		t.Fatalf("draft doc still exists: %v", err)
	}
	if _, err := os.Stat(paths.StagingDir(meta.TextID)); !os.IsNotExist(err) {
		t.Fatalf("staging dir still exists: %v", err)
	}
	list, err = store.ListDrafts(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty index, got %d", len(list))
	}
}
