package grammarbundle

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed en es
var rootFS embed.FS

// FS is the default English bundle (sections.json at repository root). Kept for
// backward compatibility: tests and callers that expect the full RU→EN course.
var FS fs.FS

func init() {
	var err error
	FS, err = BundleFS("en")
	if err != nil {
		panic("grammarbundle: embedded en bundle: " + err.Error())
	}
}

// BundleFS returns the embedded grammar bundle for bundleID (e.g. "en", "es").
// Paths inside the returned FS are as before: sections.json, index.json, chapters/…
func BundleFS(bundleID string) (fs.FS, error) {
	id := strings.TrimSpace(strings.ToLower(bundleID))
	if id == "" {
		id = "en"
	}
	sub, err := fs.Sub(rootFS, id)
	if err != nil {
		return nil, fmt.Errorf("grammar bundle %q: %w", id, err)
	}
	return sub, nil
}

// AvailableBundleIDs returns the list of grammar bundle IDs embedded in the binary
// (directory names under the embedded root, e.g. ["en", "es"]).
func AvailableBundleIDs() []string {
	entries, err := fs.ReadDir(rootFS, ".")
	if err != nil {
		return nil
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids
}

// ValidateEmbeddedBundleID checks that bundleID exists under the embedded root (en, es, …).
func ValidateEmbeddedBundleID(bundleID string) error {
	id := strings.TrimSpace(strings.ToLower(bundleID))
	if id == "" {
		return fmt.Errorf("GRAMMAR_BUNDLE_ID is required when not using GRAMMAR_BUNDLE_DIR")
	}
	// fs.Sub does not fail for missing paths; Stat the directory entry on the embedded root.
	if _, err := fs.Stat(rootFS, id); err != nil {
		return fmt.Errorf("embedded grammar bundle %q: %w", id, err)
	}
	return nil
}
