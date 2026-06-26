package readingcms

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CoverStats reports cover readiness: ready files, prompt-only, or none.
func CoverStats(doc *TextDocument, contentRoot string) (status string) {
	if doc == nil {
		return CoverNone
	}
	thumb := strings.TrimSpace(doc.CoverThumbRelPath)
	hero := strings.TrimSpace(doc.CoverHeroRelPath)
	if thumb != "" && hero != "" {
		ready := true
		for _, rel := range []string{thumb, hero} {
			abs := filepath.Join(contentRoot, filepath.FromSlash(rel))
			st, err := os.Stat(abs)
			if err != nil || st.IsDir() || st.Size() == 0 {
				ready = false
				break
			}
		}
		if ready {
			return CoverReady
		}
	}
	if strings.TrimSpace(doc.CoverImagePrompt) != "" {
		return CoverPrompt
	}
	return CoverNone
}

// CoverGeneratedAt returns the latest modification time of cover asset files, if any.
func CoverGeneratedAt(doc *TextDocument, contentRoot string) *time.Time {
	if doc == nil {
		return nil
	}
	textID := strings.TrimSpace(doc.ID)
	var best time.Time
	found := false
	for _, rel := range []string{doc.CoverThumbRelPath, doc.CoverHeroRelPath} {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		abs := filepath.Join(contentRoot, filepath.FromSlash(rel))
		st, err := os.Stat(abs)
		if err != nil || st.IsDir() || st.Size() == 0 {
			continue
		}
		mt := st.ModTime().UTC()
		if !found || mt.After(best) {
			best = mt
			found = true
		}
	}
	if !found && textID != "" {
		raw := filepath.Join(contentRoot, "assets", "reading", textID, "cover_raw.png")
		st, err := os.Stat(raw)
		if err == nil && !st.IsDir() && st.Size() > 0 {
			t := st.ModTime().UTC()
			return &t
		}
		return nil
	}
	if !found {
		return nil
	}
	t := best
	return &t
}

// CoverGenerateOpts controls a single cover generation run.
type CoverGenerateOpts struct {
	Force   bool
	Prompt  string
	SkipLLM bool
}

// GenerateCover runs LLM + ComfyUI pipeline into draft staging and updates the document.
func (s *Service) GenerateCover(ctx context.Context, textID string, opts CoverGenerateOpts) (*DraftMeta, error) {
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
	if !opts.SkipLLM && !opts.Force && CoverStats(doc, stagingRoot) == CoverReady {
		meta.CoverStatus = CoverReady
		meta.CoverThumbRelPath = doc.CoverThumbRelPath
		meta.CoverImagePrompt = doc.CoverImagePrompt
		meta.LastJobLog = "cover already ready"
		return &meta, nil
	}
	if err := writeStagingCatalogForCover(stagingRoot, doc); err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(opts.Prompt)
	if opts.SkipLLM && prompt == "" {
		return nil, errInvalid("prompt required when skip_llm")
	}
	force := opts.Force || opts.SkipLLM
	jobLog, err := s.runCoverScript(ctx, meta.CourseCode, stagingRoot, textID, force, prompt)
	meta.LastJobLog = jobLog
	if err != nil {
		_ = s.store.SaveDraft(meta, doc)
		return &meta, err
	}
	updated, err := readStagingTextDoc(stagingRoot, textID)
	if err != nil {
		return nil, err
	}
	meta.CoverStatus = CoverStats(updated, stagingRoot)
	meta.CoverThumbRelPath = updated.CoverThumbRelPath
	meta.CoverImagePrompt = updated.CoverImagePrompt
	if strings.TrimSpace(jobLog) != "" {
		meta.LastJobLog = jobLog
	}
	if err := s.store.SaveDraft(meta, updated); err != nil {
		return nil, err
	}
	return &meta, nil
}

func clearDocumentCover(doc *TextDocument) {
	if doc == nil {
		return
	}
	doc.CoverThumbRelPath = ""
	doc.CoverHeroRelPath = ""
	doc.CoverImagePrompt = ""
}

func removeCoverFiles(contentRoot, textID string, doc *TextDocument) {
	seen := map[string]bool{}
	add := func(abs string) {
		if abs == "" || seen[abs] {
			return
		}
		seen[abs] = true
		_ = os.Remove(abs)
	}
	if doc != nil {
		for _, rel := range []string{doc.CoverThumbRelPath, doc.CoverHeroRelPath} {
			rel = strings.TrimSpace(rel)
			if rel != "" {
				add(filepath.Join(contentRoot, filepath.FromSlash(rel)))
			}
		}
	}
	base := filepath.Join(contentRoot, "assets", "reading", textID)
	for _, name := range []string{"cover_thumb.webp", "cover_hero.webp", "cover_raw.png"} {
		add(filepath.Join(base, name))
	}
}

// DeleteDraftCover removes cover assets and clears cover fields on a draft.
func (s *Service) DeleteDraftCover(textID string) (*DraftMeta, error) {
	meta, doc, err := s.GetDraft(textID)
	if err != nil {
		return nil, err
	}
	staging := s.paths.StagingDir(textID)
	removeCoverFiles(staging, textID, doc)
	clearDocumentCover(doc)
	meta.CoverStatus = CoverNone
	meta.CoverThumbRelPath = ""
	meta.CoverImagePrompt = ""
	meta.LastJobLog = "cover deleted"
	if err := s.store.SaveDraft(meta, doc); err != nil {
		return nil, err
	}
	return &meta, nil
}

// DeletePublishedCover removes cover assets from course catalog and syncs draft if present.
func (s *Service) DeletePublishedCover(courseCode, textID string) (*PublishedItem, error) {
	courseCode = strings.ToLower(strings.TrimSpace(courseCode))
	textID = strings.TrimSpace(textID)
	if courseCode == "" || textID == "" {
		return nil, errInvalid("course_code and text_id required")
	}
	course, err := s.paths.Course(courseCode)
	if err != nil {
		return nil, err
	}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return nil, err
	}
	doc, err := readTextFile(course.GrammarDir, idx, textID)
	if err != nil {
		return nil, fmt.Errorf("published text not found: %s", textID)
	}
	removeCoverFiles(course.GrammarDir, textID, doc)
	clearDocumentCover(doc)
	if err := writeTextToCourse(course.GrammarDir, doc); err != nil {
		return nil, err
	}
	if meta, ok, err := s.store.GetMeta(textID); err == nil && ok {
		staging := s.paths.StagingDir(textID)
		stagingDoc, derr := s.store.GetDocument(textID)
		if derr == nil && stagingDoc != nil {
			removeCoverFiles(staging, textID, stagingDoc)
		}
		clearDocumentCover(doc)
		meta.CoverStatus = CoverNone
		meta.CoverThumbRelPath = ""
		meta.CoverImagePrompt = ""
		meta.LastJobLog = "cover deleted (synced from course)"
		_ = s.store.SaveDraft(meta, doc)
	}
	item := publishedItemFromDoc(course.Code, course.GrammarDir, textID, doc, s.cmsDraftIDs())
	return item, nil
}

func (s *Service) GenerateCoverBatch(_ context.Context, req CoverBatchRequest) (int, error) {
	_, total, err := s.StartCoverBatch(req)
	return total, err
}

// GeneratePublishedCover runs LLM + ComfyUI for one published course text (writes to course dir).
func (s *Service) GeneratePublishedCover(ctx context.Context, courseCode, textID string, opts CoverGenerateOpts) (*PublishedItem, string, error) {
	courseCode = strings.ToLower(strings.TrimSpace(courseCode))
	textID = strings.TrimSpace(textID)
	if courseCode == "" || textID == "" {
		return nil, "", errInvalid("course_code and text_id required")
	}
	course, err := s.paths.Course(courseCode)
	if err != nil {
		return nil, "", err
	}
	idx, err := loadReadingIndex(course.GrammarDir)
	if err != nil {
		return nil, "", err
	}
	doc, err := readTextFile(course.GrammarDir, idx, textID)
	if err != nil {
		return nil, "", fmt.Errorf("published text not found: %s", textID)
	}
	if !opts.SkipLLM && !opts.Force && CoverStats(doc, course.GrammarDir) == CoverReady {
		item := publishedItemFromDoc(course.Code, course.GrammarDir, textID, doc, s.cmsDraftIDs())
		return item, "cover already ready (skipped)", nil
	}
	prompt := strings.TrimSpace(opts.Prompt)
	if opts.SkipLLM && prompt == "" {
		return nil, "", errInvalid("prompt required when skip_llm")
	}
	force := opts.Force || opts.SkipLLM
	jobLog, err := s.runCoverScript(ctx, course.Code, course.GrammarDir, textID, force, prompt)
	if err != nil {
		return nil, jobLog, err
	}
	doc, err = readTextFile(course.GrammarDir, idx, textID)
	if err != nil {
		return nil, jobLog, err
	}
	_ = s.syncPublishedCoverToDraft(textID, doc, course.GrammarDir)
	item := publishedItemFromDoc(course.Code, course.GrammarDir, textID, doc, s.cmsDraftIDs())
	return item, jobLog, nil
}

func (s *Service) cmsDraftIDs() map[string]bool {
	out := map[string]bool{}
	drafts, _ := s.store.ListDrafts(nil)
	for _, d := range drafts {
		out[d.TextID] = true
	}
	return out
}

func publishedItemFromDoc(courseCode, courseDir, textID string, doc *TextDocument, cmsIDs map[string]bool) *PublishedItem {
	title := strings.TrimSpace(doc.Title)
	if title == "" {
		title = textID
	}
	total := CountSegments(doc)
	_, withAudio, audioSt := AudioStats(doc, courseDir)
	coverSt := CoverStats(doc, courseDir)
	return &PublishedItem{
		TextID:            textID,
		CourseCode:        courseCode,
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
		CoverGeneratedAt:  CoverGeneratedAt(doc, courseDir),
		InCMS:             cmsIDs[textID],
	}
}

func (s *Service) syncPublishedCoverToDraft(textID string, doc *TextDocument, courseDir string) error {
	meta, ok, err := s.store.GetMeta(textID)
	if err != nil || !ok {
		return err
	}
	staging := s.paths.StagingDir(textID)
	src := filepath.Join(courseDir, "assets", "reading", textID)
	dst := filepath.Join(staging, "assets", "reading", textID)
	if err := copyDirIfExists(src, dst); err != nil {
		return err
	}
	meta.CoverStatus = CoverStats(doc, staging)
	meta.CoverThumbRelPath = doc.CoverThumbRelPath
	meta.CoverImagePrompt = doc.CoverImagePrompt
	return s.store.SaveDraft(meta, doc)
}

func (s *Service) runCoverScript(ctx context.Context, courseCode, courseDir, textID string, force bool, imagePrompt string, extraLines ...func(string)) (string, error) {
	prog := s.coverProgressRegister(courseCode, textID)
	var runErr error
	defer func() { prog.finish(runErr) }()

	args := []string{
		filepath.Join(s.paths.RepoRoot, "scripts", "generate-reading-cover.py"),
		"--course-root", courseDir,
		"--text-id", textID,
	}
	if p := strings.TrimSpace(imagePrompt); p != "" {
		args = append(args, "--image-prompt", p, "--force")
	} else if force {
		args = append(args, "--force")
	}
	cmd := exec.CommandContext(ctx, "python3", args...)
	cmd.Dir = s.paths.RepoRoot
	if strings.TrimSpace(courseCode) == "" {
		courseCode = CourseCodeFromDir(courseDir)
	}
	cmd.Env = append(s.paths.ScriptEnv(courseCode), "PYTHONUNBUFFERED=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		runErr = err
		log := prog.view().Log
		return log, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		runErr = err
		log := prog.view().Log
		return log, err
	}
	if err := cmd.Start(); err != nil {
		runErr = err
		log := prog.view().Log
		return log, err
	}

	var wg sync.WaitGroup
	stream := func(r io.Reader) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			prog.appendLine(line)
			for _, sink := range extraLines {
				if sink != nil {
					sink(line)
				}
			}
		}
	}
	wg.Add(2)
	go stream(stdout)
	go stream(stderr)
	wg.Wait()

	runErr = cmd.Wait()
	log := prog.view().Log
	if runErr != nil {
		return log, runErr
	}
	return log, nil
}

func (s *Service) runCoverPromptScript(ctx context.Context, courseCode, courseDir, textID string, force bool, extraLines ...func(string)) (string, error) {
	prog := s.coverProgressRegister(courseCode, textID)
	var runErr error
	defer func() { prog.finish(runErr) }()

	args := []string{
		filepath.Join(s.paths.RepoRoot, "scripts", "generate-reading-cover.py"),
		"--course-root", courseDir,
		"--text-id", textID,
		"--prompt-only",
	}
	if force {
		args = append(args, "--force")
	}
	cmd := exec.CommandContext(ctx, "python3", args...)
	cmd.Dir = s.paths.RepoRoot
	if strings.TrimSpace(courseCode) == "" {
		courseCode = CourseCodeFromDir(courseDir)
	}
	cmd.Env = append(s.paths.ScriptEnv(courseCode), "PYTHONUNBUFFERED=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		runErr = err
		return prog.view().Log, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		runErr = err
		return prog.view().Log, err
	}
	if err := cmd.Start(); err != nil {
		runErr = err
		return prog.view().Log, err
	}

	var wg sync.WaitGroup
	stream := func(r io.Reader) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			prog.appendLine(line)
			for _, sink := range extraLines {
				if sink != nil {
					sink(line)
				}
			}
		}
	}
	wg.Add(2)
	go stream(stdout)
	go stream(stderr)
	wg.Wait()

	runErr = cmd.Wait()
	log := prog.view().Log
	if runErr != nil {
		return log, runErr
	}
	return log, nil
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
