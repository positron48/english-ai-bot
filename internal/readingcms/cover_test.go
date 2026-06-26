package readingcms

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestCoverStatsPromptOnly(t *testing.T) {
	root := t.TempDir()
	doc := &TextDocument{
		ID:               "cover_prompt_1",
		CoverImagePrompt: "A quiet plaza at sunset, watercolor style",
	}
	if got := CoverStats(doc, root); got != CoverPrompt {
		t.Fatalf("expected prompt, got %s", got)
	}
}

func TestCoverGeneratedAt(t *testing.T) {
	root := t.TempDir()
	textID := "cover_time_1"
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	heroRel := "assets/reading/" + textID + "/cover_hero.webp"
	doc := &TextDocument{
		ID:                textID,
		CoverThumbRelPath: thumbRel,
		CoverHeroRelPath:  heroRel,
	}
	if got := CoverGeneratedAt(doc, root); got != nil {
		t.Fatal("expected nil before files")
	}
	thumbAbs := filepath.Join(root, filepath.FromSlash(thumbRel))
	if err := os.MkdirAll(filepath.Dir(thumbAbs), 0o755); err != nil {
		t.Fatal(err)
	}
	thumbTime := time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC)
	if err := os.WriteFile(thumbAbs, []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(thumbAbs, thumbTime, thumbTime); err != nil {
		t.Fatal(err)
	}
	heroAbs := filepath.Join(root, filepath.FromSlash(heroRel))
	heroTime := time.Date(2026, 6, 23, 9, 49, 0, 0, time.UTC)
	if err := os.WriteFile(heroAbs, []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(heroAbs, heroTime, heroTime); err != nil {
		t.Fatal(err)
	}
	got := CoverGeneratedAt(doc, root)
	if got == nil {
		t.Fatal("expected time")
	}
	if !got.Equal(heroTime) {
		t.Fatalf("expected latest hero mtime %v, got %v", heroTime, got)
	}
}

func TestDeleteDraftCover(t *testing.T) {
	root := t.TempDir()
	svc, err := NewService(root)
	if err != nil {
		t.Fatal(err)
	}
	textID := "cover_del_draft"
	staging := svc.paths.StagingDir(textID)
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	heroRel := "assets/reading/" + textID + "/cover_hero.webp"
	doc := &TextDocument{
		ID:                textID,
		Title:             "Del me",
		Level:             "A0",
		TargetLanguage:    "es",
		CoverThumbRelPath: thumbRel,
		CoverHeroRelPath:  heroRel,
		CoverImagePrompt:  "prompt",
		ReadingPassage:    map[string]interface{}{"segments": []interface{}{}},
	}
	meta := DraftMeta{
		TextID:     textID,
		CourseCode: "es_ru",
		Title:      "Del me",
		Level:      "A0",
		Status:     StatusDraft,
	}
	if err := svc.store.SaveDraft(meta, doc); err != nil {
		t.Fatal(err)
	}
	thumbAbs := filepath.Join(staging, filepath.FromSlash(thumbRel))
	if err := os.MkdirAll(filepath.Dir(thumbAbs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(thumbAbs, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	heroAbs := filepath.Join(staging, filepath.FromSlash(heroRel))
	if err := os.WriteFile(heroAbs, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := svc.DeleteDraftCover(textID)
	if err != nil {
		t.Fatal(err)
	}
	if out.CoverStatus != CoverNone {
		t.Fatalf("status=%s", out.CoverStatus)
	}
	if _, err := os.Stat(thumbAbs); !os.IsNotExist(err) {
		t.Fatalf("thumb should be removed: %v", err)
	}
	_, doc2, err := svc.GetDraft(textID)
	if err != nil {
		t.Fatal(err)
	}
	if doc2.CoverThumbRelPath != "" || doc2.CoverImagePrompt != "" {
		t.Fatalf("cover fields not cleared: %+v", doc2)
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
