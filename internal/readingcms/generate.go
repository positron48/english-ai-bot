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

type scriptRunOptions struct {
	course    Course
	level     string
	format    string
	title     string
	withAudio bool
	inputText string
	inputJSON []byte
	origin    string
	jobLog    string
}

// GenerateBatch runs LLM generation for one or more texts into CMS drafts.
func (s *Service) GenerateBatch(ctx context.Context, req GenerateRequest) ([]DraftMeta, error) {
	if req.Count < 1 {
		req.Count = 1
	}
	level, err := normalizeLevel(req.Level)
	if err != nil {
		return nil, err
	}
	course, err := s.paths.Course(req.CourseCode)
	if err != nil {
		return nil, err
	}
	format := normalizeFormat(req.Format)

	var created []DraftMeta
	var errs []string
	for i := 0; i < req.Count; i++ {
		select {
		case <-ctx.Done():
			return created, ctx.Err()
		default:
		}
		title := req.Title
		if req.Count > 1 {
			title = ""
		}
		meta, err := s.runReadingScript(ctx, scriptRunOptions{
			course:    course,
			level:     level,
			format:    format,
			title:     title,
			withAudio: req.WithAudio,
			origin:    OriginLLM,
			jobLog:    "generated via LLM",
		})
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		created = append(created, *meta)
		if i < req.Count-1 {
			time.Sleep(2 * time.Second)
		}
	}
	if len(created) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return created, nil
}

func (s *Service) runReadingScript(ctx context.Context, opts scriptRunOptions) (*DraftMeta, error) {
	workID := fmt.Sprintf("job_%d", time.Now().UnixNano())
	workDir := filepath.Join(s.paths.GenWorkDir(), workID)
	readingDir := filepath.Join(workDir, "reading")
	if err := os.MkdirAll(readingDir, 0o755); err != nil {
		return nil, err
	}
	defer os.RemoveAll(workDir)

	var inputTextPath string
	if strings.TrimSpace(opts.inputText) != "" {
		inputTextPath = filepath.Join(workDir, "input.txt")
		if err := os.WriteFile(inputTextPath, []byte(opts.inputText), 0o644); err != nil {
			return nil, err
		}
	}

	script := filepath.Join(opts.course.GrammarDir, "scripts", "generate-reading-text.py")
	args := []string{
		script,
		"--course-root", opts.course.GrammarDir,
		"--draft-dir", readingDir,
		"--target-lang", opts.course.TargetLang,
		"--level", opts.level,
		"--format", opts.format,
	}
	if strings.TrimSpace(opts.title) != "" {
		args = append(args, "--title", opts.title)
	}
	if inputTextPath != "" {
		args = append(args, "--input-text", inputTextPath)
	}
	if len(opts.inputJSON) > 0 {
		jsonPath := filepath.Join(workDir, "input.json")
		if err := os.WriteFile(jsonPath, opts.inputJSON, 0o644); err != nil {
			return nil, err
		}
		args = append(args, "--input-json", jsonPath)
	}

	cmd := exec.CommandContext(ctx, "python3", args...)
	cmd.Dir = opts.course.GrammarDir
	cmd.Env = os.Environ()
	if opts.withAudio {
		cmd.Env = append(cmd.Env, "READING_TTS_CMD_TEMPLATE="+s.paths.TTSCommandTemplate())
	} else {
		cmd.Env = append(cmd.Env, "READING_TTS_CMD_TEMPLATE=")
	}
	cmd.Env = append(cmd.Env, "READING_ENSURE_LLAMA=0")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if out, err := cmd.Output(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("reading script failed: %s", msg)
	} else if strings.TrimSpace(stderr.String()) != "" {
		_ = out
	}

	textID, doc, err := s.loadGeneratedDraft(readingDir, opts.course)
	if err != nil {
		return nil, err
	}

	stagingRoot := s.paths.StagingDir(textID)
	if err := os.MkdirAll(stagingRoot, 0o755); err != nil {
		return nil, err
	}
	courseAudio := filepath.Join(opts.course.GrammarDir, "assets", "reading", textID)
	stagingAudio := filepath.Join(stagingRoot, "assets", "reading", textID)
	if err := copyDirIfExists(courseAudio, stagingAudio); err != nil {
		return nil, err
	}
	_ = os.RemoveAll(courseAudio)

	total, withAudioCount, audioSt := AudioStats(doc, stagingRoot)
	meta := DraftMeta{
		TextID:            textID,
		CourseCode:        opts.course.Code,
		Title:             doc.Title,
		Level:             opts.level,
		Format:            opts.format,
		TargetLanguage:    opts.course.TargetLang,
		Status:            StatusDraft,
		Origin:            opts.origin,
		AudioStatus:       audioSt,
		CoverStatus:       CoverNone,
		SegmentsTotal:     total,
		SegmentsWithAudio: withAudioCount,
		LastJobLog:        opts.jobLog,
	}
	if err := s.store.SaveDraft(meta, doc); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *Service) loadGeneratedDraft(readingDir string, course Course) (string, *TextDocument, error) {
	indexPath := filepath.Join(readingDir, "index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return "", nil, fmt.Errorf("reading index missing after generation: %w", err)
	}
	var idx readingIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return "", nil, err
	}
	if len(idx.Texts) == 0 {
		return "", nil, fmt.Errorf("no text generated")
	}
	var textID string
	for id := range idx.Texts {
		textID = id
		break
	}
	rel := idx.Texts[textID]
	path := filepath.Join(readingDir, filepath.FromSlash(rel))
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	var doc TextDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", nil, err
	}
	if doc.TargetLanguage == "" {
		doc.TargetLanguage = course.TargetLang
	}
	return textID, &doc, nil
}

func copyDirIfExists(src, dst string) error {
	st, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !st.IsDir() {
		return nil
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
