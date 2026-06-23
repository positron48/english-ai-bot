package readingcms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type readingIndex struct {
	Version     string                     `json:"version"`
	GeneratedAt string                     `json:"generated_at"`
	Categories  map[string]readingCategory `json:"categories"`
	Texts       map[string]string          `json:"texts"`
}

type readingCategory struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Level   string   `json:"level"`
	Order   int      `json:"order"`
	TextIDs []string `json:"text_ids"`
}

func loadReadingIndex(courseDir string) (*readingIndex, error) {
	path := filepath.Join(courseDir, "reading", "index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &readingIndex{
				Version:    "1.0.0",
				Categories: map[string]readingCategory{},
				Texts:      map[string]string{},
			}, nil
		}
		return nil, err
	}
	var idx readingIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	if idx.Categories == nil {
		idx.Categories = map[string]readingCategory{}
	}
	if idx.Texts == nil {
		idx.Texts = map[string]string{}
	}
	return &idx, nil
}

func saveReadingIndex(courseDir string, idx *readingIndex) error {
	path := filepath.Join(courseDir, "reading", "index.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	idx.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func upsertTextInIndex(idx *readingIndex, doc *TextDocument) {
	catID := strings.TrimSpace(doc.CategoryID)
	if catID == "" {
		catID = categoryID(doc.TargetLanguage, doc.Level)
		doc.CategoryID = catID
	}
	cat := idx.Categories[catID]
	if cat.ID == "" {
		cat = readingCategory{
			ID:    catID,
			Title: fmt.Sprintf("%s %s Reading", strings.ToUpper(doc.TargetLanguage), doc.Level),
			Level: doc.Level,
			Order: 0,
		}
	}
	found := false
	for _, id := range cat.TextIDs {
		if id == doc.ID {
			found = true
			break
		}
	}
	if !found {
		cat.TextIDs = append(cat.TextIDs, doc.ID)
	}
	idx.Categories[catID] = cat
	if idx.Texts == nil {
		idx.Texts = map[string]string{}
	}
	idx.Texts[doc.ID] = fmt.Sprintf("texts/%s.json", doc.ID)
}

func removeTextFromIndex(idx *readingIndex, textID string) (relPath string, ok bool) {
	relPath, ok = idx.Texts[textID]
	delete(idx.Texts, textID)
	for catID, cat := range idx.Categories {
		filtered := cat.TextIDs[:0]
		for _, id := range cat.TextIDs {
			if id != textID {
				filtered = append(filtered, id)
			}
		}
		cat.TextIDs = filtered
		if len(cat.TextIDs) == 0 {
			delete(idx.Categories, catID)
		} else {
			idx.Categories[catID] = cat
		}
	}
	return relPath, ok
}

func applyReadingTextDeletion(contentRoot, textID, relPath string) error {
	if strings.Contains(relPath, "..") || strings.HasPrefix(relPath, "/") {
		return fmt.Errorf("invalid reading text path")
	}
	idx, err := loadReadingIndex(contentRoot)
	if err != nil {
		return err
	}
	if relPath == "" {
		relPath = idx.Texts[textID]
	}
	if relPath != "" {
		textFile := filepath.Join(contentRoot, "reading", filepath.FromSlash(relPath))
		_ = os.Remove(textFile)
	}
	removeTextFromIndex(idx, textID)
	if err := saveReadingIndex(contentRoot, idx); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(contentRoot, "assets", "reading", textID))
}
