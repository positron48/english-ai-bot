package speakingsync

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
	Version     string                    `json:"version"`
	GeneratedAt string                    `json:"generated_at"`
	Categories  map[string]bundleCategory `json:"categories"`
	Tasks       map[string]string         `json:"tasks"`
}

type bundleCategory struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	TitleTranslations map[string]string `json:"title_translations,omitempty"`
	Level             string            `json:"level"`
	Order             int               `json:"order"`
	TaskIDs           []string          `json:"task_ids"`
}

type bundleTaskFile struct {
	SchemaVersion     string   `json:"schema_version"`
	ID                string   `json:"id"`
	CategoryID        string   `json:"category_id"`
	Level             string   `json:"level"`
	Type              string   `json:"type"`
	TargetLanguage    string   `json:"target_language"`
	Title             string   `json:"title"`
	PromptRU          string   `json:"prompt_ru"`
	DisplayText       string   `json:"display_text"`
	ExpectedMeaningRU string   `json:"expected_meaning_ru"`
	AcceptableAnswers []string `json:"acceptable_answers"`
	EvaluationNotes   string   `json:"evaluation_notes"`
	MaxAttempts       int      `json:"max_attempts"`
	Order             int      `json:"order"`
}

// SyncFromBundle reads speaking/index.json and task JSON from grammar bundle FS.
func SyncFromBundle(ctx context.Context, cfg *config.Config, repo *repository.SpeakingCatalogRepository, log *zap.Logger) error {
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
		return fmt.Errorf("speaking bundle fs: %w", err)
	}
	raw, err := fs.ReadFile(bundleFS, "speaking/index.json")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if log != nil {
				log.Info("speaking sync: no speaking/index.json in bundle, clearing DB catalog")
			}
			return repo.ReplaceCatalog("1.0.0", "", nil, nil)
		}
		return fmt.Errorf("read speaking/index.json: %w", err)
	}

	var idx bundleIndex
	if err := json.Unmarshal(raw, &idx); err != nil {
		return fmt.Errorf("parse speaking index: %w", err)
	}
	if idx.Categories == nil {
		idx.Categories = map[string]bundleCategory{}
	}
	if idx.Tasks == nil {
		idx.Tasks = map[string]string{}
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

	cats := make([]repository.SpeakingCategoryUpsert, 0, len(idx.Categories))
	for _, ck := range catOrder {
		c := idx.Categories[ck.id]
		catID := strings.TrimSpace(c.ID)
		if catID == "" {
			catID = ck.id
		}
		title := strings.TrimSpace(c.Title)
		if title == "" {
			title = catID
		}
		cats = append(cats, repository.SpeakingCategoryUpsert{
			CategoryID:        catID,
			Title:             title,
			TitleTranslations: c.TitleTranslations,
			Level:             c.Level,
			SortOrder:         c.Order,
			TaskIDs:           append([]string(nil), c.TaskIDs...),
		})
	}

	taskToCategory := make(map[string]string)
	for mapKey, c := range idx.Categories {
		effectiveCat := strings.TrimSpace(c.ID)
		if effectiveCat == "" {
			effectiveCat = mapKey
		}
		for _, tid := range c.TaskIDs {
			taskToCategory[tid] = effectiveCat
		}
	}

	taskUpserts := make([]repository.SpeakingTaskUpsert, 0, len(idx.Tasks))
	for taskID, rel := range idx.Tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
			continue
		}
		data, err := fs.ReadFile(bundleFS, filepath.Join("speaking", filepath.Clean(rel)))
		if err != nil {
			if log != nil {
				log.Warn("speaking sync: skip missing task file", zap.String("task_id", taskID), zap.Error(err))
			}
			continue
		}
		var doc bundleTaskFile
		if err := json.Unmarshal(data, &doc); err != nil {
			if log != nil {
				log.Warn("speaking sync: skip invalid task json", zap.String("task_id", taskID), zap.Error(err))
			}
			continue
		}
		tid := strings.TrimSpace(doc.ID)
		if tid == "" {
			tid = taskID
		}
		catID := strings.TrimSpace(doc.CategoryID)
		if catID == "" {
			catID = taskToCategory[tid]
		}
		if catID == "" {
			continue
		}
		title := strings.TrimSpace(doc.Title)
		if title == "" {
			title = tid
		}
		taskUpserts = append(taskUpserts, repository.SpeakingTaskUpsert{
			TaskID:         tid,
			CategoryID:     catID,
			Title:          title,
			Level:          doc.Level,
			TaskType:       doc.Type,
			TargetLanguage: doc.TargetLanguage,
			SortOrder:      doc.Order,
			TaskJSON:       string(data),
		})
	}

	if err := repo.ReplaceCatalog(idx.Version, idx.GeneratedAt, cats, taskUpserts); err != nil {
		return err
	}
	if log != nil {
		log.Info("speaking catalog synced from bundle",
			zap.Int("categories", len(cats)),
			zap.Int("tasks", len(taskUpserts)),
		)
	}
	return nil
}
