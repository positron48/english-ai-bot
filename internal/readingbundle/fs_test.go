package readingbundle

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"tgbot-skeleton/internal/config"
)

func TestGrammarFilesystemRoot_NilConfig(t *testing.T) {
	t.Parallel()
	if _, err := GrammarFilesystemRoot(nil); err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestGrammarFilesystemRoot_ExplicitDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "reading"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reading", "index.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleDir: dir}}
	got, err := GrammarFilesystemRoot(cfg)
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGrammarFilesystemRoot_NoIndexInRepo(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "nonexistent-bundle-id-xyz"}}
	got, err := GrammarFilesystemRoot(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("expected empty root, got %q", got)
	}
}

func TestBundleFS_DirOverride(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "reading"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reading", "index.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleDir: dir}}
	bundleFS, err := BundleFS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := fs.ReadFile(bundleFS, "reading/index.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "{}" {
		t.Fatalf("index = %q", raw)
	}
}

func TestBundleFS_EmbeddedFallback(t *testing.T) {
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "en"}}
	bundleFS, err := BundleFS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ReadFile(bundleFS, "reading/index.json"); err != nil {
		t.Fatalf("embedded reading index: %v", err)
	}
}
