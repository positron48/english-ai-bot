package readingsync

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func writeReadingBundleDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	readingDir := filepath.Join(dir, "reading")
	if err := os.MkdirAll(filepath.Join(readingDir, "texts"), 0o755); err != nil {
		t.Fatal(err)
	}
	index := `{
  "version": "2.0.0",
  "generated_at": "2026-01-01T00:00:00Z",
  "categories": {
    "cat-a": {
      "id": "cat-a",
      "title": "Category A",
      "level": "A1",
      "order": 2,
      "text_ids": ["text-one"]
    },
    "cat-b": {
      "title": "",
      "level": "A0",
      "order": 1,
      "text_ids": ["text-two"]
    }
  },
  "texts": {
    "text-one": "texts/text-one.json",
    "text-two": "texts/text-two.json",
    "bad-path": "../evil.json",
    "missing": "texts/missing.json"
  }
}`
	if err := os.WriteFile(filepath.Join(readingDir, "index.json"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	textOne := `{
  "id": "text-one",
  "category_id": "cat-a",
  "title": "Hello",
  "level": "A1",
  "target_language": "en",
  "reading_passage": {"segments": [{"text": "Hi"}]}
}`
	textTwo := `{
  "category_id": "cat-b",
  "title": "",
  "level": "A0",
  "target_language": "en",
  "reading_passage": {"segments": []}
}`
	if err := os.WriteFile(filepath.Join(readingDir, "texts", "text-one.json"), []byte(textOne), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(readingDir, "texts", "text-two.json"), []byte(textTwo), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(readingDir, "texts", "invalid.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestSyncFromBundle_NilSafe(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if err := SyncFromBundle(ctx, nil, nil, nil); err != nil {
		t.Fatalf("nil cfg/repo: %v", err)
	}
	cfg := &config.Config{Learning: config.DefaultLearningConfig()}
	if err := SyncFromBundle(ctx, cfg, nil, nil); err != nil {
		t.Fatalf("nil repo: %v", err)
	}
}

func TestSyncFromBundle_CancelledContext(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewReadingCatalogRepository(db)
	dir := writeReadingBundleDir(t)
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleDir: dir,
		TargetLang:       "en",
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := SyncFromBundle(ctx, cfg, repo, zap.NewNop()); err == nil {
		t.Fatal("expected context error")
	}
}

func TestSyncFromBundle_MissingIndexClearsCatalog(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewReadingCatalogRepository(db)
	dir := t.TempDir()
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleDir: dir,
		TargetLang:       "en",
	}}
	if err := SyncFromBundle(context.Background(), cfg, repo, zap.NewNop()); err != nil {
		t.Fatalf("sync missing index: %v", err)
	}
	n, err := repo.CountCategories()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected empty catalog, got %d categories", n)
	}
}

func TestSyncFromBundle_ReplacesCatalogFromBundle(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewReadingCatalogRepository(db)
	dir := writeReadingBundleDir(t)
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleDir: dir,
		TargetLang:       "en",
	}}
	if err := SyncFromBundle(context.Background(), cfg, repo, zap.NewNop()); err != nil {
		t.Fatalf("sync: %v", err)
	}
	snap, err := repo.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Categories) != 2 {
		t.Fatalf("categories = %d, want 2", len(snap.Categories))
	}
	if snap.Categories["cat-b"].Order != 1 || snap.Categories["cat-a"].Order != 2 {
		t.Fatalf("sort order wrong: %+v", snap.Categories)
	}
	doc, ok, err := repo.GetTextDocument("text-one")
	if err != nil || !ok {
		t.Fatalf("text-one: ok=%v err=%v", ok, err)
	}
	if doc.Title != "Hello" || doc.TargetLanguage != "en" {
		t.Fatalf("doc = %+v", doc)
	}
	doc2, ok, err := repo.GetTextDocument("text-two")
	if err != nil || !ok {
		t.Fatalf("text-two: ok=%v err=%v", ok, err)
	}
	if doc2.CategoryID != "cat-b" || doc2.Title != "text-two" {
		t.Fatalf("doc2 = %+v", doc2)
	}
}

func TestSyncFromBundle_InvalidIndexJSON(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewReadingCatalogRepository(db)
	dir := t.TempDir()
	readingDir := filepath.Join(dir, "reading")
	if err := os.MkdirAll(readingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(readingDir, "index.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleDir: dir,
		TargetLang:       "en",
	}}
	if err := SyncFromBundle(context.Background(), cfg, repo, zap.NewNop()); err == nil {
		t.Fatal("expected parse error")
	}
}
