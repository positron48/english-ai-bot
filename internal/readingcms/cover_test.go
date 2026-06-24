package readingcms

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCoverStats(t *testing.T) {
	root := t.TempDir()
	textID := "cover_test_1"
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	heroRel := "assets/reading/" + textID + "/cover_hero.webp"
	doc := &TextDocument{
		ID:                textID,
		CoverThumbRelPath: thumbRel,
		CoverHeroRelPath:  heroRel,
	}
	if got := CoverStats(doc, root); got != CoverNone {
		t.Fatalf("expected none before files, got %s", got)
	}
	thumbAbs := filepath.Join(root, filepath.FromSlash(thumbRel))
	if err := os.MkdirAll(filepath.Dir(thumbAbs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(thumbAbs, []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}
	heroAbs := filepath.Join(root, filepath.FromSlash(heroRel))
	if err := os.WriteFile(heroAbs, []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := CoverStats(doc, root); got != CoverReady {
		t.Fatalf("expected ready, got %s", got)
	}
}

func TestWriteStagingCatalogForCover(t *testing.T) {
	root := t.TempDir()
	doc := &TextDocument{
		ID:             "t1",
		Title:          "Hello",
		Level:          "A2",
		TargetLanguage: "en",
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	if err := writeStagingCatalogForCover(root, doc); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "reading", "texts", "t1.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected staging json: %v", err)
	}
	loaded, err := readStagingTextDoc(root, "t1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Title != "Hello" {
		t.Fatalf("title=%q", loaded.Title)
	}
}
