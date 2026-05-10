package web

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"tgbot-skeleton/internal/config"
)

func TestReadingWritableRootDir_GrammarBundleDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "reading"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reading", "index.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleDir: dir}}
	got, err := readingWritableRootDir(cfg)
	if err != nil {
		t.Fatal(err)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != absDir {
		t.Fatalf("got %q want %q", got, absDir)
	}
}

func TestReadingBundleFS_matchesWritableRootInRepo(t *testing.T) {
	t.Parallel()
	repoRoot := findRepoRoot(t)
	if repoRoot == "" {
		t.Skip("not inside english-ai-bot repo")
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "es", GrammarBundleDir: ""}}
	root, err := readingWritableRootDir(cfg)
	if err != nil {
		t.Fatal(err)
	}
	bundleFS, err := readingBundleFS(cfg)
	if err != nil {
		t.Fatal(err)
	}
	data1, err := os.ReadFile(filepath.Join(root, "reading", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	data2, err := fs.ReadFile(bundleFS, "reading/index.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data1) != string(data2) {
		t.Fatal("readingBundleFS and readingWritableRootDir must use the same tree")
	}
}

func TestReadingWritableRootDir_DevRepoFallback(t *testing.T) {
	t.Parallel()
	repoRoot := findRepoRoot(t)
	if repoRoot == "" {
		t.Skip("not inside english-ai-bot repo")
	}
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "es", GrammarBundleDir: ""}}
	got, err := readingWritableRootDir(cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(repoRoot, "internal", "grammarbundle", "es")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReadingWritableRootDir_NoDir(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{Learning: config.LearningConfig{GrammarBundleID: "zz-unknown", GrammarBundleDir: ""}}
	_, err := readingWritableRootDir(cfg)
	if err == nil {
		t.Fatal("expected error")
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for dir != "" && dir != filepath.Dir(dir) {
		if st, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !st.IsDir() {
			return dir
		}
		if st, err := os.Stat(filepath.Join(dir, "internal", "grammarbundle", "es", "reading", "index.json")); err == nil && !st.IsDir() {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}
