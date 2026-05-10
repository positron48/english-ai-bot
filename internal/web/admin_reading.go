package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

type adminReadingTextItem struct {
	TextID         string `json:"text_id"`
	CategoryID     string `json:"category_id"`
	Title          string `json:"title"`
	Level          string `json:"level"`
	TargetLanguage string `json:"target_language"`
	SegmentsCount  int    `json:"segments_count"`
}

func (r *Router) handleAdminReadingTexts(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleAdminReadingTextsList(w, req)
		return
	case http.MethodDelete:
		r.handleAdminReadingTextDelete(w, req)
		return
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (r *Router) handleAdminReadingTextsList(w http.ResponseWriter, req *http.Request) {
	idx, err := r.readReadingIndex()
	if err != nil {
		r.logger.Error("admin reading: failed to load index", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	search := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("search")))
	levelFilter := strings.ToUpper(strings.TrimSpace(req.URL.Query().Get("level")))
	langFilter := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("target_lang")))

	items := make([]adminReadingTextItem, 0, len(idx.Texts))
	for textID := range idx.Texts {
		doc, err := r.readReadingText(idx, textID)
		if err != nil {
			r.logger.Warn("admin reading: failed to read text doc", zap.String("text_id", textID), zap.Error(err))
			continue
		}
		title := strings.TrimSpace(doc.Title)
		if title == "" {
			title = textID
		}
		if search != "" && !strings.Contains(strings.ToLower(title), search) {
			continue
		}
		if levelFilter != "" && strings.ToUpper(strings.TrimSpace(doc.Level)) != levelFilter {
			continue
		}
		if langFilter != "" && strings.ToLower(strings.TrimSpace(doc.TargetLanguage)) != langFilter {
			continue
		}

		segmentsCount := 0
		if segs, ok := doc.ReadingPassage["segments"].([]interface{}); ok {
			segmentsCount = len(segs)
		}

		items = append(items, adminReadingTextItem{
			TextID:         doc.ID,
			CategoryID:     doc.CategoryID,
			Title:          title,
			Level:          doc.Level,
			TargetLanguage: doc.TargetLanguage,
			SegmentsCount:  segmentsCount,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Level == items[j].Level {
			return items[i].Title < items[j].Title
		}
		return items[i].Level < items[j].Level
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"texts": items,
		"total": len(items),
	})
}

func (r *Router) handleAdminReadingTextDelete(w http.ResponseWriter, req *http.Request) {
	textID := strings.Trim(strings.TrimPrefix(req.URL.Path, "/api/admin/reading/texts/"), "/")
	if textID == "" {
		http.Error(w, "text_id required", http.StatusBadRequest)
		return
	}

	rootDir, err := readingWritableRootDir(r.config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	readingDir := filepath.Join(rootDir, "reading")
	indexPath := filepath.Join(readingDir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		http.Error(w, "reading index not found", http.StatusNotFound)
		return
	}
	var idx readingIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		http.Error(w, "invalid reading index", http.StatusInternalServerError)
		return
	}

	relPath, ok := idx.Texts[textID]
	if !ok {
		http.Error(w, "reading text not found", http.StatusNotFound)
		return
	}
	if strings.Contains(relPath, "..") || strings.HasPrefix(relPath, "/") {
		http.Error(w, "invalid reading text path", http.StatusBadRequest)
		return
	}

	textFilePath := filepath.Join(readingDir, filepath.Clean(relPath))
	_ = os.Remove(textFilePath)

	delete(idx.Texts, textID)
	for catID, cat := range idx.Categories {
		filtered := make([]string, 0, len(cat.TextIDs))
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
	idx.GeneratedAt = nowRFC3339UTC()
	updated, _ := json.MarshalIndent(idx, "", "  ")
	if err := os.WriteFile(indexPath, append(updated, '\n'), 0o644); err != nil {
		r.logger.Error("admin reading: failed to write index", zap.Error(err))
		http.Error(w, "failed to update reading index", http.StatusInternalServerError)
		return
	}

	audioDir := filepath.Join(rootDir, "assets", "reading", textID)
	_ = os.RemoveAll(audioDir)

	if r.db != nil {
		if _, err := r.db.Exec(`DELETE FROM reading_text_progress WHERE chapter_id = $1`, textID); err != nil {
			r.logger.Warn("admin reading: failed to cleanup reading progress", zap.String("text_id", textID), zap.Error(err))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"text_id": textID,
	})
}

func readingWritableRootDir(cfg *config.Config) (string, error) {
	root, err := grammarBundleFilesystemRoot(cfg)
	if err != nil {
		return "", err
	}
	if root == "" {
		return "", fmt.Errorf(
			"reading admin needs a filesystem grammar bundle root: set GRAMMAR_BUNDLE_DIR, or run the server from the repo root (internal/grammarbundle/%s)",
			strings.TrimSpace(cfg.Learning.GrammarBundleID),
		)
	}
	return root, nil
}

func nowRFC3339UTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

