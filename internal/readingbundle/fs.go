package readingbundle

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/grammarbundle"
)

// GrammarFilesystemRoot returns an absolute grammar bundle root when the server reads
// grammar/reading from the working tree. Empty string means use the embedded bundle.
func GrammarFilesystemRoot(cfg *config.Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is nil")
	}
	if dir := strings.TrimSpace(cfg.Learning.GrammarBundleDir); dir != "" {
		return filepath.Abs(dir)
	}
	bundleID := strings.TrimSpace(cfg.Learning.GrammarBundleID)
	if bundleID == "" {
		return "", nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", nil
	}
	candidate := filepath.Join(wd, "internal", "grammarbundle", bundleID)
	indexPath := filepath.Join(candidate, "reading", "index.json")
	st, statErr := os.Stat(indexPath)
	if statErr != nil || st.IsDir() {
		return "", nil
	}
	return filepath.Abs(candidate)
}

// BundleFS returns the filesystem (dir or embed) that holds reading/index.json and assets.
func BundleFS(cfg *config.Config) (fs.FS, error) {
	lc := cfg.Learning
	root, err := GrammarFilesystemRoot(cfg)
	if err != nil {
		return nil, err
	}
	if root != "" {
		return os.DirFS(root), nil
	}
	return grammarbundle.BundleFS(lc.GrammarBundleID)
}
