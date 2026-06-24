package readingcms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CoverStats reports whether thumb/hero cover files exist in content root.
func CoverStats(doc *TextDocument, contentRoot string) (status string) {
	if doc == nil {
		return CoverNone
	}
	thumb := strings.TrimSpace(doc.CoverThumbRelPath)
	hero := strings.TrimSpace(doc.CoverHeroRelPath)
	if thumb == "" || hero == "" {
		return CoverNone
	}
	for _, rel := range []string{thumb, hero} {
		abs := filepath.Join(contentRoot, filepath.FromSlash(rel))
		st, err := os.Stat(abs)
		if err != nil || st.IsDir() || st.Size() == 0 {
			return CoverNone
		}
	}
	return CoverReady
}

// GenerateCover runs LLM + ComfyUI pipeline into draft staging and updates the document.
func (s *Service) GenerateCover(ctx context.Context, textID string, force bool) (*DraftMeta, error) {
	meta, ok, err := s.store.GetMeta(textID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("draft not found: %s", textID)
	}
	doc, err := s.store.GetDocument(textID)
	if err != nil {
		return nil, err
	}
	stagingRoot := s.paths.StagingDir(textID)
	if !force && CoverStats(doc, stagingRoot) == CoverReady {
		meta.CoverStatus = CoverReady
		meta.CoverImagePrompt = doc.CoverImagePrompt
		meta.LastJobLog = "cover already ready"
		return &meta, nil
	}
	if err := writeStagingCatalogForCover(stagingRoot, doc); err != nil {
		return nil, err
	}
	if err := s.runCoverScript(ctx, stagingRoot, textID, force); err != nil {
		meta.LastJobLog = err.Error()
		_ = s.store.SaveDraft(meta, doc)
		return nil, err
	}
	updated, err := readStagingTextDoc(stagingRoot, textID)
	if err != nil {
		return nil, err
	}
	meta.CoverStatus = CoverStats(updated, stagingRoot)
	meta.CoverImagePrompt = updated.CoverImagePrompt
	meta.LastJobLog = fmt.Sprintf("cover generated (%s)", meta.CoverStatus)
	if err := s.store.SaveDraft(meta, updated); err != nil {
		return nil, err
	}
	return &meta, nil
}

// GenerateCoverBatch generates covers for published texts in a course catalog.
func (s *Service) GenerateCoverBatch(ctx context.Context, req CoverBatchRequest) (int, error) {
	course, err := s.paths.Course(req.CourseCode)
	if err != nil {
		return 0, err
	}
	args := []string{
		filepath.Join(s.paths.RepoRoot, "scripts", "generate-reading-cover.py"),
		"--course-root", course.GrammarDir,
	}
	if req.Force {
		args = append(args, "--force")
	}
	cmd := exec.CommandContext(ctx, "python3", args...)
	cmd.Dir = s.paths.RepoRoot
	cmd.Env = os.Environ()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return 0, fmt.Errorf("%w: %s", err, msg)
		}
		return 0, err
	}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return 0, err
	}
	count := 0
	level := strings.ToUpper(strings.TrimSpace(req.Level))
	for textID := range idx.Texts {
		doc, err := readTextFile(course.GrammarDir, idx, textID)
		if err != nil {
			continue
		}
		if level != "" && strings.ToUpper(doc.Level) != level {
			continue
		}
		if CoverStats(doc, course.GrammarDir) != CoverReady {
			continue
		}
		count++
		_ = textID
	}
	return count, nil
}

func (s *Service) runCoverScript(ctx context.Context, courseDir, textID string, force bool) error {
	args := []string{
		filepath.Join(s.paths.RepoRoot, "scripts", "generate-reading-cover.py"),
		"--course-root", courseDir,
		"--text-id", textID,
	}
	if force {
		args = append(args, "--force")
	}
	cmd := exec.CommandContext(ctx, "python3", args...)
	cmd.Dir = s.paths.RepoRoot
	cmd.Env = os.Environ()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
}

func writeStagingCatalogForCover(stagingRoot string, doc *TextDocument) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}
	textsDir := filepath.Join(stagingRoot, "reading", "texts")
	if err := os.MkdirAll(textsDir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(textsDir, doc.ID+".json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	idx := map[string]interface{}{
		"version":      "1.0.0",
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"categories":   map[string]interface{}{},
		"texts":        map[string]string{doc.ID: "texts/" + doc.ID + ".json"},
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stagingRoot, "reading", "index.json"), append(data, '\n'), 0o644)
}

func readStagingTextDoc(stagingRoot, textID string) (*TextDocument, error) {
	path := filepath.Join(stagingRoot, "reading", "texts", textID+".json")
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
