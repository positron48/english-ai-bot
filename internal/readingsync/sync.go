package readingsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/readingbundle"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

type bundleIndex struct {
	Version     string                     `json:"version"`
	GeneratedAt string                     `json:"generated_at"`
	Categories  map[string]bundleCategory  `json:"categories"`
	Texts       map[string]string          `json:"texts"`
}

type bundleCategory struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	TitleTranslations map[string]string `json:"title_translations,omitempty"`
	Level             string            `json:"level"`
	Order             int               `json:"order"`
	TextIDs           []string          `json:"text_ids"`
}

type bundleTextFile struct {
	ID                string                 `json:"id"`
	CategoryID        string                 `json:"category_id"`
	Title             string                 `json:"title"`
	TitleTranslations map[string]string      `json:"title_translations,omitempty"`
	Level             string                 `json:"level"`
	TargetLanguage    string                 `json:"target_language"`
	ReadingPassage    map[string]interface{} `json:"reading_passage"`
}

// SyncFromBundle reads reading/index.json and text JSON from the grammar bundle FS and replaces DB rows.
func SyncFromBundle(ctx context.Context, cfg *config.Config, repo *repository.ReadingCatalogRepository, log *zap.Logger) error {
	if cfg == nil || repo == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	bundleFS, err := readingbundle.BundleFS(cfg)
	if err != nil {
		return fmt.Errorf("reading bundle fs: %w", err)
	}
	raw, err := fs.ReadFile(bundleFS, "reading/index.json")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if log != nil {
				log.Info("reading sync: no reading/index.json in bundle, clearing DB catalog")
			}
			return repo.ReplaceCatalog("1.0.0", "", nil, nil)
		}
		return fmt.Errorf("read reading/index.json: %w", err)
	}

	var idx bundleIndex
	if err := json.Unmarshal(raw, &idx); err != nil {
		return fmt.Errorf("parse reading index: %w", err)
	}
	if idx.Categories == nil {
		idx.Categories = map[string]bundleCategory{}
	}
	if idx.Texts == nil {
		idx.Texts = map[string]string{}
	}

	type catKey struct {
		id    string
		order int
	}
	var catOrder []catKey
	for id, c := range idx.Categories {
		catOrder = append(catOrder, catKey{id: id, order: c.Order})
	}
	sort.Slice(catOrder, func(i, j int) bool {
		if catOrder[i].order == catOrder[j].order {
			return catOrder[i].id < catOrder[j].id
		}
		return catOrder[i].order < catOrder[j].order
	})

	cats := make([]repository.ReadingCategoryUpsert, 0, len(idx.Categories))
	for _, ck := range catOrder {
		c := idx.Categories[ck.id]
		title := strings.TrimSpace(c.Title)
		if title == "" {
			title = ck.id
		}
		catID := strings.TrimSpace(c.ID)
		if catID == "" {
			catID = ck.id
		}
		textIDs := append([]string(nil), c.TextIDs...)
		cats = append(cats, repository.ReadingCategoryUpsert{
			CategoryID:        catID,
			Title:             title,
			TitleTranslations: c.TitleTranslations,
			Level:             c.Level,
			SortOrder:         c.Order,
			TextIDs:           textIDs,
		})
	}

	textUpserts := make([]repository.ReadingTextUpsert, 0, len(idx.Texts))
	textToCategory := make(map[string]string, len(idx.Texts))
	for mapKey, c := range idx.Categories {
		effectiveCat := strings.TrimSpace(c.ID)
		if effectiveCat == "" {
			effectiveCat = mapKey
		}
		for _, tid := range c.TextIDs {
			textToCategory[tid] = effectiveCat
		}
	}
	for textID, rel := range idx.Texts {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
			if log != nil {
				log.Warn("reading sync: skip invalid text path", zap.String("text_id", textID), zap.String("rel", rel))
			}
			continue
		}
		data, err := fs.ReadFile(bundleFS, filepath.Join("reading", filepath.Clean(rel)))
		if err != nil {
			if log != nil {
				log.Warn("reading sync: skip missing text file", zap.String("text_id", textID), zap.Error(err))
			}
			continue
		}
		var doc bundleTextFile
		if err := json.Unmarshal(data, &doc); err != nil {
			if log != nil {
				log.Warn("reading sync: skip invalid text json", zap.String("text_id", textID), zap.Error(err))
			}
			continue
		}
		passageJSON, err := json.Marshal(doc.ReadingPassage)
		if err != nil {
			return err
		}
		tid := strings.TrimSpace(doc.ID)
		if tid == "" {
			tid = textID
		}
		title := strings.TrimSpace(doc.Title)
		if title == "" {
			title = tid
		}
		catID := strings.TrimSpace(doc.CategoryID)
		if catID == "" {
			catID = textToCategory[tid]
		}
		if catID == "" {
			if log != nil {
				log.Warn("reading sync: skip text without category", zap.String("text_id", tid))
			}
			continue
		}
		textUpserts = append(textUpserts, repository.ReadingTextUpsert{
			TextID:               tid,
			CategoryID:           catID,
			Title:                title,
			TitleTranslations:    doc.TitleTranslations,
			Level:                doc.Level,
			TargetLanguage:       doc.TargetLanguage,
			ReadingPassageJSON:   string(passageJSON),
		})
	}

	if err := repo.ReplaceCatalog(idx.Version, idx.GeneratedAt, cats, textUpserts); err != nil {
		return err
	}
	if log != nil {
		log.Info("reading catalog synced from bundle",
			zap.Int("categories", len(cats)),
			zap.Int("texts", len(textUpserts)),
		)
	}
	return nil
}
