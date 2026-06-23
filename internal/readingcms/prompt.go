package readingcms

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ReadingPrompt returns the course-specific LLM prompt used by generate-reading-text.py.
func (s *Service) ReadingPrompt(ctx context.Context, req PromptRequest) (string, error) {
	level, err := normalizeLevel(req.Level)
	if err != nil {
		return "", err
	}
	course, err := s.paths.Course(req.CourseCode)
	if err != nil {
		return "", err
	}
	format := normalizeFormat(req.Format)
	kind := strings.ToLower(strings.TrimSpace(req.Kind))
	if kind == "" {
		kind = "generate"
	}
	if kind != "generate" && kind != "transform" {
		return "", fmt.Errorf("kind must be generate or transform")
	}

	workDir, err := os.MkdirTemp(s.paths.GenWorkDir(), "prompt-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)

	script := filepath.Join(course.GrammarDir, "scripts", "generate-reading-text.py")
	args := []string{
		script,
		"--course-root", course.GrammarDir,
		"--print-prompt",
		"--prompt-kind", kind,
		"--target-lang", course.TargetLang,
		"--level", level,
		"--format", format,
	}
	if strings.TrimSpace(req.Title) != "" {
		args = append(args, "--title", req.Title)
	}
	if kind == "transform" {
		source := strings.TrimSpace(req.SourceText)
		if source == "" {
			return "", fmt.Errorf("source_text is required for transform prompt")
		}
		inputPath := filepath.Join(workDir, "source.txt")
		if err := os.WriteFile(inputPath, []byte(source), 0o644); err != nil {
			return "", err
		}
		args = append(args, "--input-text", inputPath)
	}

	cmd := exec.CommandContext(ctx, "python3", args...)
	cmd.Dir = course.GrammarDir
	cmd.Env = append(os.Environ(), "READING_ENSURE_LLAMA=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("prompt script failed: %s", msg)
	}
	return strings.TrimSpace(string(out)), nil
}
