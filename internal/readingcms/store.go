package readingcms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Store struct {
	paths *Paths
	mu    sync.Mutex
}

func NewStore(paths *Paths) *Store {
	return &Store{paths: paths}
}

func (s *Store) EnsureDirs() error {
	for _, dir := range []string{s.paths.DataDir, s.paths.DraftsDir(), s.paths.GenWorkDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

type draftIndex struct {
	Drafts []DraftMeta `json:"drafts"`
}

func (s *Store) loadIndex() (*draftIndex, error) {
	path := s.paths.IndexPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &draftIndex{Drafts: []DraftMeta{}}, nil
		}
		return nil, err
	}
	var idx draftIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	if idx.Drafts == nil {
		idx.Drafts = []DraftMeta{}
	}
	return &idx, nil
}

func (s *Store) saveIndex(idx *draftIndex) error {
	if err := s.EnsureDirs(); err != nil {
		return err
	}
	sort.Slice(idx.Drafts, func(i, j int) bool {
		return idx.Drafts[i].UpdatedAt.After(idx.Drafts[j].UpdatedAt)
	})
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.paths.IndexPath(), append(data, '\n'), 0o644)
}

func (s *Store) ListDrafts(filter func(DraftMeta) bool) ([]DraftMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, err := s.loadIndex()
	if err != nil {
		return nil, err
	}
	out := make([]DraftMeta, 0, len(idx.Drafts))
	for _, d := range idx.Drafts {
		if filter == nil || filter(d) {
			out = append(out, d)
		}
	}
	return out, nil
}

func (s *Store) GetMeta(textID string) (DraftMeta, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, err := s.loadIndex()
	if err != nil {
		return DraftMeta{}, false, err
	}
	for _, d := range idx.Drafts {
		if d.TextID == textID {
			return d, true, nil
		}
	}
	return DraftMeta{}, false, nil
}

func (s *Store) GetDocument(textID string) (*TextDocument, error) {
	path := s.paths.DraftDocPath(textID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc TextDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *Store) SaveDraft(meta DraftMeta, doc *TextDocument) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.EnsureDirs(); err != nil {
		return err
	}
	now := time.Now().UTC()
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = now
	}
	meta.UpdatedAt = now
	if doc != nil {
		if strings.TrimSpace(doc.ID) == "" {
			doc.ID = meta.TextID
		}
		if strings.TrimSpace(doc.Title) == "" {
			doc.Title = meta.Title
		}
		raw, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(s.paths.DraftDocPath(meta.TextID), append(raw, '\n'), 0o644); err != nil {
			return err
		}
	}
	idx, err := s.loadIndex()
	if err != nil {
		return err
	}
	found := false
	for i, d := range idx.Drafts {
		if d.TextID == meta.TextID {
			meta.CreatedAt = d.CreatedAt
			idx.Drafts[i] = meta
			found = true
			break
		}
	}
	if !found {
		idx.Drafts = append(idx.Drafts, meta)
	}
	return s.saveIndex(idx)
}

func (s *Store) UpdateMeta(textID string, fn func(*DraftMeta)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, err := s.loadIndex()
	if err != nil {
		return err
	}
	for i := range idx.Drafts {
		if idx.Drafts[i].TextID == textID {
			fn(&idx.Drafts[i])
			idx.Drafts[i].UpdatedAt = time.Now().UTC()
			return s.saveIndex(idx)
		}
	}
	return fmt.Errorf("draft not found: %s", textID)
}

func (s *Store) DeleteDraft(textID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, err := s.loadIndex()
	if err != nil {
		return err
	}
	filtered := idx.Drafts[:0]
	for _, d := range idx.Drafts {
		if d.TextID != textID {
			filtered = append(filtered, d)
		}
	}
	idx.Drafts = filtered
	_ = os.Remove(s.paths.DraftDocPath(textID))
	_ = os.RemoveAll(s.paths.StagingDir(textID))
	return s.saveIndex(idx)
}

func (s *Store) StagingAssetsDir(textID string) string {
	return filepath.Join(s.paths.StagingDir(textID), "assets", "reading", textID)
}
