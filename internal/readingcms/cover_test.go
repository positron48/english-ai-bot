package readingcms

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupCoverService(t *testing.T) (*Service, string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repoRoot, err := FindRepoRoot(wd)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Symlink(filepath.Join(repoRoot, "scripts"), filepath.Join(root, "scripts")); err != nil {
		t.Fatal(err)
	}
	initCourseTree(t, root, "courses/spanish-grammar")
	svc, err := NewService(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("READING_COVER_PROMPT_MOCK", "1")
	return svc, root
}

func seedPublishedDoc(t *testing.T, svc *Service, textID string, doc *TextDocument) {
	t.Helper()
	course, err := svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID == "" {
		doc.ID = textID
	}
	if err := writeTextToCourse(course.GrammarDir, doc); err != nil {
		t.Fatal(err)
	}
}

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

func TestCoverStatsNilAndPartial(t *testing.T) {
	if CoverStats(nil, t.TempDir()) != CoverNone {
		t.Fatal("nil doc")
	}
	root := t.TempDir()
	doc := &TextDocument{
		ID:                "partial",
		CoverThumbRelPath: "assets/reading/partial/cover_thumb.webp",
		CoverHeroRelPath:  "assets/reading/partial/cover_hero.webp",
	}
	if CoverStats(doc, root) != CoverNone {
		t.Fatal("missing files should be none")
	}
}

func TestCoverGeneratedAtRawPNG(t *testing.T) {
	root := t.TempDir()
	textID := "raw_only"
	doc := &TextDocument{ID: textID}
	raw := filepath.Join(root, "assets", "reading", textID, "cover_raw.png")
	if err := os.MkdirAll(filepath.Dir(raw), 0o755); err != nil {
		t.Fatal(err)
	}
	ts := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	if err := os.WriteFile(raw, []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(raw, ts, ts); err != nil {
		t.Fatal(err)
	}
	got := CoverGeneratedAt(doc, root)
	if got == nil || !got.Equal(ts) {
		t.Fatalf("expected raw mtime %v got %v", ts, got)
	}
}

func TestRunCoverPromptScriptMock(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "cover_prompt_mock"
	doc := &TextDocument{
		ID:             textID,
		Title:          "Tapas",
		Level:          "A2",
		TargetLanguage: "es",
		ReadingPassage: map[string]interface{}{
			"segments": []map[string]interface{}{{"text": "Buenos días"}},
		},
	}
	seedPublishedDoc(t, svc, textID, doc)
	course, err := svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}

	log, err := svc.runCoverPromptScript(context.Background(), "es_ru", course.GrammarDir, textID, false)
	if err != nil {
		t.Fatalf("prompt script: %v log=%q", err, log)
	}
	if !strings.Contains(log, "done") && !strings.Contains(log, "skip") {
		t.Fatalf("unexpected log: %q", log)
	}
	updated, err := readTextFile(course.GrammarDir, mustLoadIndex(t, course.GrammarDir), textID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(updated.CoverImagePrompt) == "" {
		t.Fatal("expected cover_image_prompt saved")
	}
}

func mustLoadIndex(t *testing.T, courseDir string) *readingIndex {
	t.Helper()
	idx, err := loadReadingIndex(courseDir)
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

func TestGenerateCoverAlreadyReady(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "cover_ready_skip"
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	heroRel := "assets/reading/" + textID + "/cover_hero.webp"
	meta := DraftMeta{
		TextID:         textID,
		CourseCode:     "es_ru",
		Title:          "Ready",
		Level:          "A2",
		TargetLanguage: "es",
		Status:         StatusDraft,
	}
	doc := &TextDocument{
		ID:                textID,
		Title:             meta.Title,
		Level:             meta.Level,
		TargetLanguage:    meta.TargetLanguage,
		CoverThumbRelPath: thumbRel,
		CoverHeroRelPath:  heroRel,
		CoverImagePrompt:  "existing prompt",
		ReadingPassage:    map[string]interface{}{"segments": []interface{}{}},
	}
	if err := svc.store.SaveDraft(meta, doc); err != nil {
		t.Fatal(err)
	}
	staging := svc.paths.StagingDir(textID)
	for _, rel := range []string{thumbRel, heroRel} {
		abs := filepath.Join(staging, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("webp"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out, err := svc.GenerateCover(context.Background(), textID, CoverGenerateOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if out.LastJobLog != "cover already ready" {
		t.Fatalf("log=%q", out.LastJobLog)
	}
}

func TestGenerateCoverSkipLLMRequiresPrompt(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "cover_skip_llm"
	meta := DraftMeta{
		TextID:         textID,
		CourseCode:     "es_ru",
		Title:          "Skip",
		Level:          "A2",
		TargetLanguage: "es",
		Status:         StatusDraft,
	}
	doc := &TextDocument{
		ID:             textID,
		Title:          meta.Title,
		Level:          meta.Level,
		TargetLanguage: meta.TargetLanguage,
		ReadingPassage: map[string]interface{}{"segments": []interface{}{}},
	}
	if err := svc.store.SaveDraft(meta, doc); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GenerateCover(context.Background(), textID, CoverGenerateOpts{SkipLLM: true}); err == nil {
		t.Fatal("expected error without prompt")
	}
}

func TestDeletePublishedCover(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "pub_cover_del_unit"
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	doc := &TextDocument{
		ID:                textID,
		Title:             "Published cover",
		Level:             "A2",
		TargetLanguage:    "es",
		CoverThumbRelPath: thumbRel,
		CoverImagePrompt:  "prompt",
		ReadingPassage:    map[string]interface{}{"segments": []interface{}{}},
	}
	seedPublishedDoc(t, svc, textID, doc)
	course, err := svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	thumbAbs := filepath.Join(course.GrammarDir, filepath.FromSlash(thumbRel))
	if err := os.MkdirAll(filepath.Dir(thumbAbs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(thumbAbs, []byte("webp"), 0o644); err != nil {
		t.Fatal(err)
	}

	item, err := svc.DeletePublishedCover("es_ru", textID)
	if err != nil {
		t.Fatal(err)
	}
	if item.CoverStatus != CoverNone {
		t.Fatalf("status=%s", item.CoverStatus)
	}
}

func TestGeneratePublishedCoverAlreadyReady(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "pub_cover_ready"
	thumbRel := "assets/reading/" + textID + "/cover_thumb.webp"
	heroRel := "assets/reading/" + textID + "/cover_hero.webp"
	doc := &TextDocument{
		ID:                textID,
		Title:             "Ready pub",
		Level:             "A2",
		TargetLanguage:    "es",
		CoverThumbRelPath: thumbRel,
		CoverHeroRelPath:  heroRel,
		CoverImagePrompt:  "prompt",
		ReadingPassage:    map[string]interface{}{"segments": []interface{}{}},
	}
	seedPublishedDoc(t, svc, textID, doc)
	course, err := svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{thumbRel, heroRel} {
		abs := filepath.Join(course.GrammarDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("webp"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	item, log, err := svc.GeneratePublishedCover(context.Background(), "es_ru", textID, CoverGenerateOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if item.CoverStatus != CoverReady {
		t.Fatalf("status=%s", item.CoverStatus)
	}
	if log == "" {
		t.Fatal("expected skip log")
	}
}

func TestPublishedItemFromDoc(t *testing.T) {
	svc, _ := setupCoverService(t)
	textID := "pub_item_doc"
	doc := &TextDocument{
		ID:               textID,
		Title:            "",
		Level:            "A2",
		TargetLanguage:   "es",
		CoverImagePrompt: "scene",
		ReadingPassage:   map[string]interface{}{"segments": []interface{}{map[string]interface{}{"text": "hola"}}},
	}
	seedPublishedDoc(t, svc, textID, doc)
	course, err := svc.paths.Course("es_ru")
	if err != nil {
		t.Fatal(err)
	}
	item := publishedItemFromDoc("es_ru", course.GrammarDir, textID, doc, map[string]bool{textID: true})
	if item.Title != textID {
		t.Fatalf("title=%q", item.Title)
	}
	if !item.InCMS || item.CoverStatus != CoverPrompt {
		t.Fatalf("item=%+v", item)
	}
}

func TestWriteStagingCatalogForCoverNil(t *testing.T) {
	if err := writeStagingCatalogForCover(t.TempDir(), nil); err == nil {
		t.Fatal("expected error for nil doc")
	}
}

func TestCourseCodeFromDir(t *testing.T) {
	if CourseCodeFromDir("/tmp/courses/spanish-grammar") != "es_ru" {
		t.Fatal("expected es_ru")
	}
	if CourseCodeFromDir("/tmp/courses/english-grammar") != "en_ru" {
		t.Fatal("expected en_ru")
	}
}
