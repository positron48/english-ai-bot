package readingcms

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PublishedSyncRequest imports course catalog texts into CMS drafts (mirror for editing/covers).
type PublishedSyncRequest struct {
	CourseCode string `json:"course_code"`
	Level      string `json:"level,omitempty"`
	Search     string `json:"search,omitempty"`
	Cover      string `json:"cover,omitempty"` // none|ready|"" (all)
	Force      bool   `json:"force"`           // re-import existing CMS entries
}

// PublishedSyncResult reports import stats.
type PublishedSyncResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Updated  int `json:"updated"`
	Total    int `json:"total"`
}

// SyncPublishedToCMS copies published course texts into local CMS storage.
func (s *Service) SyncPublishedToCMS(req PublishedSyncRequest) (*PublishedSyncResult, error) {
	course, err := s.paths.Course(req.CourseCode)
	if err != nil {
		return nil, err
	}
	items, err := s.ListPublished(req.CourseCode, req.Level, req.Search, req.Cover)
	if err != nil {
		return nil, err
	}
	coverFilter := strings.TrimSpace(req.Cover)
	if coverFilter != "" {
		filtered := items[:0]
		for _, item := range items {
			if item.CoverStatus == coverFilter {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	res := &PublishedSyncResult{Total: len(items)}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		existing, ok, err := s.store.GetMeta(item.TextID)
		if err != nil {
			return nil, err
		}
		if ok && !req.Force {
			res.Skipped++
			continue
		}

		doc, err := readTextFile(course.GrammarDir, idx, item.TextID)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", item.TextID, err)
		}
		staging := s.paths.StagingDir(item.TextID)
		srcAssets := course.GrammarDir
		if err := copyCourseAssetsToStaging(srcAssets, staging, item.TextID); err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		meta := DraftMeta{
			TextID:            item.TextID,
			CourseCode:        course.Code,
			Title:             item.Title,
			Level:             item.Level,
			TargetLanguage:    item.TargetLanguage,
			Status:            StatusPublished,
			Origin:            OriginCourseImport,
			AudioStatus:       item.AudioStatus,
			CoverStatus:       item.CoverStatus,
			CoverThumbRelPath: doc.CoverThumbRelPath,
			CoverImagePrompt:  doc.CoverImagePrompt,
			SegmentsTotal:     item.SegmentsCount,
			SegmentsWithAudio: item.SegmentsWithAudio,
			LastJobLog:        "imported from " + course.CourseDir,
			UpdatedAt:         now,
		}
		if ok {
			meta.CreatedAt = existing.CreatedAt
			if existing.PublishedAt != nil {
				meta.PublishedAt = existing.PublishedAt
			}
			res.Updated++
		} else {
			meta.CreatedAt = now
			res.Imported++
		}
		if meta.PublishedAt == nil {
			meta.PublishedAt = &now
		}
		if err := s.store.SaveDraft(meta, doc); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func copyCourseAssetsToStaging(courseDir, stagingRoot, textID string) error {
	src := filepath.Join(courseDir, "assets", "reading", textID)
	dst := filepath.Join(stagingRoot, "assets", "reading", textID)
	if err := os.RemoveAll(dst); err != nil && !os.IsNotExist(err) {
		return err
	}
	return copyDirIfExists(src, dst)
}
