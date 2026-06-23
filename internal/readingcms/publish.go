package readingcms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Publish copies an approved draft into course source files and optional bundle mirror.
func (s *Service) Publish(textID string, syncBundle bool) (*DraftMeta, error) {
	meta, ok, err := s.store.GetMeta(textID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("draft not found: %s", textID)
	}
	if meta.Status != StatusApproved && meta.Status != StatusAudioReady {
		return nil, fmt.Errorf("draft must be approved or audio_ready before publish (status=%s)", meta.Status)
	}
	doc, err := s.store.GetDocument(textID)
	if err != nil {
		return nil, err
	}
	total, withAudio, audioSt := AudioStats(doc, s.paths.StagingDir(textID))
	if audioSt != AudioReady {
		return nil, fmt.Errorf("audio not ready (%d/%d segments)", withAudio, total)
	}

	course, err := s.paths.Course(meta.CourseCode)
	if err != nil {
		return nil, err
	}
	if err := writeTextToCourse(course.GrammarDir, doc); err != nil {
		return nil, err
	}
	if err := copyStagingAudioToCourse(s.paths.StagingDir(textID), course.GrammarDir, textID); err != nil {
		return nil, err
	}
	if syncBundle {
		if err := writeTextToCourse(course.BundleDir, doc); err != nil {
			return nil, err
		}
		if err := copyStagingAudioToCourse(s.paths.StagingDir(textID), course.BundleDir, textID); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	meta.Status = StatusPublished
	meta.PublishedAt = &now
	meta.AudioStatus = audioSt
	meta.SegmentsTotal = total
	meta.SegmentsWithAudio = withAudio
	meta.LastJobLog = "published to " + course.CourseDir
	if err := s.store.SaveDraft(meta, doc); err != nil {
		return nil, err
	}
	return &meta, nil
}

func writeTextToCourse(contentRoot string, doc *TextDocument) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}
	textsDir := filepath.Join(contentRoot, "reading", "texts")
	if err := os.MkdirAll(textsDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(textsDir, doc.ID+".json")
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return err
	}
	idx, err := loadReadingIndex(contentRoot)
	if err != nil {
		return err
	}
	upsertTextInIndex(idx, doc)
	return saveReadingIndex(contentRoot, idx)
}

func copyStagingAudioToCourse(stagingRoot, contentRoot, textID string) error {
	src := filepath.Join(stagingRoot, "assets", "reading", textID)
	dst := filepath.Join(contentRoot, "assets", "reading", textID)
	if err := os.RemoveAll(dst); err != nil && !os.IsNotExist(err) {
		return err
	}
	return copyDirIfExists(src, dst)
}
