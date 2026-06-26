package readingcms

import (
	"strings"
	"sync"
	"time"
)

// Service orchestrates the local Reading CMS pipeline.
type Service struct {
	paths     *Paths
	store     *Store
	coverJobs      sync.Map // key: course:text_id -> *coverProgressState
	coverBatchJobs sync.Map // key: batch_id -> *coverBatchState
}

func NewService(repoRoot string) (*Service, error) {
	paths := NewPaths(repoRoot)
	store := NewStore(paths)
	svc := &Service{paths: paths, store: store}
	if err := store.EnsureDirs(); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *Service) Paths() *Paths { return s.paths }

func (s *Service) ListDrafts(courseCode, level, status, audio, search string) ([]DraftMeta, error) {
	courseCode = strings.ToLower(strings.TrimSpace(courseCode))
	level = strings.ToUpper(strings.TrimSpace(level))
	status = strings.TrimSpace(status)
	audio = strings.TrimSpace(audio)
	search = strings.ToLower(strings.TrimSpace(search))
	return s.store.ListDrafts(func(d DraftMeta) bool {
		if courseCode != "" && d.CourseCode != courseCode {
			return false
		}
		if level != "" && strings.ToUpper(d.Level) != level {
			return false
		}
		if status != "" && d.Status != status {
			return false
		}
		if audio != "" && d.AudioStatus != audio {
			return false
		}
		if search != "" && !strings.Contains(strings.ToLower(d.Title), search) {
			return false
		}
		return true
	})
}

func (s *Service) GetDraft(textID string) (DraftMeta, *TextDocument, error) {
	meta, ok, err := s.store.GetMeta(textID)
	if err != nil {
		return DraftMeta{}, nil, err
	}
	if !ok {
		return DraftMeta{}, nil, errNotFound("draft", textID)
	}
	doc, err := s.store.GetDocument(textID)
	if err != nil {
		return meta, nil, err
	}
	return meta, doc, nil
}

func (s *Service) SaveDocument(textID string, doc *TextDocument) (*DraftMeta, error) {
	meta, ok, err := s.store.GetMeta(textID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errNotFound("draft", textID)
	}
	if doc != nil {
		if strings.TrimSpace(doc.Title) != "" {
			meta.Title = doc.Title
		}
		if strings.TrimSpace(doc.Level) != "" {
			meta.Level = doc.Level
		}
	}
	staging := s.paths.StagingDir(textID)
	total, withAudio, audioSt := AudioStats(doc, staging)
	meta.SegmentsTotal = total
	meta.SegmentsWithAudio = withAudio
	meta.AudioStatus = audioSt
	meta.CoverStatus = CoverStats(doc, staging)
	if doc != nil {
		meta.CoverImagePrompt = doc.CoverImagePrompt
	}
	if err := s.store.SaveDraft(meta, doc); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *Service) Approve(textID string) (*DraftMeta, error) {
	meta, ok, err := s.store.GetMeta(textID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errNotFound("draft", textID)
	}
	if meta.Status == StatusRejected {
		return nil, errInvalid("cannot approve rejected draft")
	}
	meta.Status = StatusApproved
	if meta.AudioStatus == AudioReady {
		meta.Status = StatusAudioReady
	}
	meta.LastJobLog = "approved"
	if err := s.store.SaveDraft(meta, nil); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *Service) Reject(textID string) (*DraftMeta, error) {
	meta, ok, err := s.store.GetMeta(textID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errNotFound("draft", textID)
	}
	meta.Status = StatusRejected
	meta.LastJobLog = "rejected"
	if err := s.store.SaveDraft(meta, nil); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *Service) Courses() []map[string]string {
	return []map[string]string{
		{"code": "en_ru", "title": "English (RU)", "target_lang": "en"},
		{"code": "es_ru", "title": "Spanish (RU)", "target_lang": "es"},
	}
}

type cmsError struct {
	kind string
	msg  string
}

func (e *cmsError) Error() string { return e.msg }

func errNotFound(kind, id string) error {
	return &cmsError{kind: "not_found", msg: kind + " not found: " + id}
}

func errInvalid(msg string) error {
	return &cmsError{kind: "invalid", msg: msg}
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*cmsError); ok {
		return e.kind == "not_found"
	}
	return false
}

// RefreshAudioStats recomputes audio counters for a draft.
func (s *Service) RefreshAudioStats(textID string) (*DraftMeta, error) {
	meta, doc, err := s.GetDraft(textID)
	if err != nil {
		return nil, err
	}
	total, withAudio, audioSt := AudioStats(doc, s.paths.StagingDir(textID))
	meta.SegmentsTotal = total
	meta.SegmentsWithAudio = withAudio
	meta.AudioStatus = audioSt
	meta.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveDraft(meta, nil); err != nil {
		return nil, err
	}
	return &meta, nil
}
