package readingcms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListPublished scans course reading catalog on disk.
func (s *Service) ListPublished(courseCode, levelFilter, search, coverFilter string) ([]PublishedItem, error) {
	course, err := s.paths.Course(courseCode)
	if err != nil {
		return nil, err
	}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return nil, err
	}
	cmsIDs := map[string]bool{}
	drafts, _ := s.store.ListDrafts(nil)
	for _, d := range drafts {
		cmsIDs[d.TextID] = true
	}
	levelFilter = strings.ToUpper(strings.TrimSpace(levelFilter))
	search = strings.ToLower(strings.TrimSpace(search))
	coverFilter = strings.TrimSpace(coverFilter)

	var out []PublishedItem
	seen := map[string]bool{}
	for textID := range idx.Texts {
		if seen[textID] {
			continue
		}
		seen[textID] = true
		doc, err := readTextFile(course.GrammarDir, idx, textID)
		if err != nil {
			continue
		}
		title := strings.TrimSpace(doc.Title)
		if title == "" {
			title = textID
		}
		if levelFilter != "" && strings.ToUpper(doc.Level) != levelFilter {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(title), search) {
			continue
		}
		coverSt := CoverStats(doc, course.GrammarDir)
		if coverFilter != "" && coverSt != coverFilter {
			continue
		}
		segs := CountSegments(doc)
		_, withAudio, audioSt := AudioStats(doc, course.GrammarDir)
		out = append(out, PublishedItem{
			TextID:            textID,
			CourseCode:        course.Code,
			Title:             title,
			Level:             doc.Level,
			TargetLanguage:    doc.TargetLanguage,
			CategoryID:        doc.CategoryID,
			SegmentsCount:     segs,
			SegmentsWithAudio: withAudio,
			AudioStatus:       audioSt,
			AudioReady:        audioSt == AudioReady,
			CoverStatus:       coverSt,
			CoverThumbRelPath: doc.CoverThumbRelPath,
			CoverHeroRelPath:  doc.CoverHeroRelPath,
			CoverImagePrompt:  doc.CoverImagePrompt,
			CoverGeneratedAt:  CoverGeneratedAt(doc, course.GrammarDir),
			InCMS:             cmsIDs[textID],
		})
	}
	return out, nil
}

// GetPublishedDocument loads one course catalog text with its JSON document.
func (s *Service) GetPublishedDocument(courseCode, textID string) (*PublishedItem, *TextDocument, error) {
	courseCode = strings.ToLower(strings.TrimSpace(courseCode))
	textID = strings.TrimSpace(textID)
	if courseCode == "" || textID == "" {
		return nil, nil, errInvalid("course_code and text_id required")
	}
	course, err := s.paths.Course(courseCode)
	if err != nil {
		return nil, nil, err
	}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return nil, nil, err
	}
	doc, err := readTextFile(course.GrammarDir, idx, textID)
	if err != nil {
		return nil, nil, fmt.Errorf("published text not found: %s", textID)
	}
	cmsIDs := s.cmsDraftIDs()
	title := strings.TrimSpace(doc.Title)
	if title == "" {
		title = textID
	}
	total := CountSegments(doc)
	_, withAudio, audioSt := AudioStats(doc, course.GrammarDir)
	coverSt := CoverStats(doc, course.GrammarDir)
	item := &PublishedItem{
		TextID:            textID,
		CourseCode:        course.Code,
		Title:             title,
		Level:             doc.Level,
		TargetLanguage:    doc.TargetLanguage,
		CategoryID:        doc.CategoryID,
		SegmentsCount:     total,
		SegmentsWithAudio: withAudio,
		AudioStatus:       audioSt,
		AudioReady:        audioSt == AudioReady,
		CoverStatus:       coverSt,
		CoverThumbRelPath: doc.CoverThumbRelPath,
		CoverHeroRelPath:  doc.CoverHeroRelPath,
		CoverImagePrompt:  doc.CoverImagePrompt,
		CoverGeneratedAt:  CoverGeneratedAt(doc, course.GrammarDir),
		InCMS:             cmsIDs[textID],
	}
	return item, doc, nil
}

func readTextFile(courseDir string, idx *readingIndex, textID string) (*TextDocument, error) {
	rel, ok := idx.Texts[textID]
	if !ok {
		return nil, os.ErrNotExist
	}
	path := filepath.Join(courseDir, "reading", filepath.FromSlash(rel))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc TextDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.ID == "" {
		doc.ID = textID
	}
	return &doc, nil
}

func segmentsHaveAudio(courseDir string, doc *TextDocument) bool {
	if doc == nil || doc.ReadingPassage == nil {
		return false
	}
	segs := passageSegments(doc.ReadingPassage)
	if len(segs) == 0 {
		return false
	}
	for _, seg := range segs {
		rel, _ := seg["audio_rel_path"].(string)
		rel = strings.TrimSpace(rel)
		if rel == "" {
			return false
		}
		abs := filepath.Join(courseDir, filepath.FromSlash(rel))
		st, err := os.Stat(abs)
		if err != nil || st.IsDir() || st.Size() == 0 {
			return false
		}
	}
	return true
}
