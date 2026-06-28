package speakingsync

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

func writeSpeakingBundleDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	speakingDir := filepath.Join(dir, "speaking")
	if err := os.MkdirAll(filepath.Join(speakingDir, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}
	index := `{
  "version": "1.0.0",
  "generated_at": "2026-01-01T00:00:00Z",
  "categories": {
    "speak-a": {
      "id": "speak-a",
      "title": "Speaking A",
      "level": "A1",
      "order": 1,
      "task_ids": ["task-one"]
    }
  },
  "tasks": {
    "task-one": "tasks/task-one.json",
    "bad": "../evil.json",
    "missing": "tasks/missing.json"
  }
}`
	if err := os.WriteFile(filepath.Join(speakingDir, "index.json"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	taskJSON := `{
  "schema_version": "1",
  "id": "task-one",
  "category_id": "speak-a",
  "level": "A1",
  "type": "repeat",
  "target_language": "es",
  "title": "Say hello",
  "prompt_ru": "Скажи привет",
  "display_text": "Hola",
  "expected_meaning_ru": "Привет",
  "acceptable_answers": ["hola"],
  "evaluation_notes": "",
  "max_attempts": 3,
  "order": 1
}`
	if err := os.WriteFile(filepath.Join(speakingDir, "tasks", "task-one.json"), []byte(taskJSON), 0o644); err != nil {
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

func TestSyncFromBundle_MissingIndexClearsCatalog(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewSpeakingCatalogRepository(db)
	dir := t.TempDir()
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleDir: dir,
		TargetLang:       "es",
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
	repo := repository.NewSpeakingCatalogRepository(db)
	dir := writeSpeakingBundleDir(t)
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleDir: dir,
		TargetLang:       "es",
	}}
	if err := SyncFromBundle(context.Background(), cfg, repo, zap.NewNop()); err != nil {
		t.Fatalf("sync: %v", err)
	}
	snap, err := repo.LoadSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Categories) != 1 {
		t.Fatalf("categories = %d, want 1", len(snap.Categories))
	}
	task, ok, err := repo.GetTaskPublic("task-one")
	if err != nil || !ok {
		t.Fatalf("task-one: ok=%v err=%v", ok, err)
	}
	if task.Title != "Say hello" || task.TargetLanguage != "es" {
		t.Fatalf("task = %+v", task)
	}
}

func TestSyncFromBundle_InvalidIndexJSON(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewSpeakingCatalogRepository(db)
	dir := t.TempDir()
	speakingDir := filepath.Join(dir, "speaking")
	if err := os.MkdirAll(speakingDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(speakingDir, "index.json"), []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleDir: dir,
		TargetLang:       "es",
	}}
	if err := SyncFromBundle(context.Background(), cfg, repo, zap.NewNop()); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSyncFromBundle_CancelledContext(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := repository.NewSpeakingCatalogRepository(db)
	dir := writeSpeakingBundleDir(t)
	cfg := &config.Config{Learning: config.LearningConfig{
		GrammarBundleDir: dir,
		TargetLang:       "es",
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := SyncFromBundle(ctx, cfg, repo, zap.NewNop()); err == nil {
		t.Fatal("expected context error")
	}
}
