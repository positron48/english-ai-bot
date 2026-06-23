package readingcms

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Course maps Linglow course codes to on-disk course directories and language metadata.
type Course struct {
	Code         string
	Title        string
	CourseDir    string // relative to repo root, e.g. courses/spanish-grammar
	TargetLang   string
	BundleID     string
	GrammarDir   string // absolute path to course dir
	BundleDir    string // absolute path to internal/grammarbundle/<id>
}

// Paths resolves repository-relative locations for the local Reading CMS.
type Paths struct {
	RepoRoot string
	DataDir  string // .local/reading-cms
}

func FindRepoRoot(start string) (string, error) {
	dir := filepath.Clean(start)
	if dir == "" {
		dir, _ = os.Getwd()
	}
	for range 32 {
		if dir == "" || dir == filepath.Dir(dir) {
			break
		}
		goMod := filepath.Join(dir, "go.mod")
		courses := filepath.Join(dir, "courses")
		if st, err := os.Stat(goMod); err == nil && !st.IsDir() {
			if st2, err2 := os.Stat(courses); err2 == nil && st2.IsDir() {
				return dir, nil
			}
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("repo root not found (need go.mod and courses/)")
}

func NewPaths(repoRoot string) *Paths {
	return &Paths{
		RepoRoot: repoRoot,
		DataDir:  filepath.Join(repoRoot, ".local", "reading-cms"),
	}
}

func (p *Paths) DraftsDir() string     { return filepath.Join(p.DataDir, "drafts") }
func (p *Paths) StagingDir(textID string) string {
	return filepath.Join(p.DataDir, "staging", textID)
}
func (p *Paths) IndexPath() string     { return filepath.Join(p.DraftsDir(), "index.json") }
func (p *Paths) DraftDocPath(id string) string {
	return filepath.Join(p.DraftsDir(), id+".json")
}
func (p *Paths) GenWorkDir() string  { return filepath.Join(p.DataDir, "gen-work") }

func (p *Paths) Course(code string) (Course, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	switch code {
	case "en_ru", "":
		return p.course("en_ru", "English (RU)", "courses/english-grammar", "en", "en"), nil
	case "es_ru":
		return p.course("es_ru", "Spanish (RU)", "courses/spanish-grammar", "es", "es"), nil
	default:
		return Course{}, fmt.Errorf("unknown course_code %q", code)
	}
}

func (p *Paths) course(code, title, relDir, targetLang, bundleID string) Course {
	return Course{
		Code:       code,
		Title:      title,
		CourseDir:  relDir,
		TargetLang: targetLang,
		BundleID:   bundleID,
		GrammarDir: filepath.Join(p.RepoRoot, relDir),
		BundleDir:  filepath.Join(p.RepoRoot, "internal", "grammarbundle", bundleID),
	}
}

func (p *Paths) TTSCommandTemplate() string {
	script := filepath.Join(p.RepoRoot, "scripts", "tts-reading-segment.sh")
	return fmt.Sprintf(
		`bash %q --voice-id "{voice_id}" --text "{text}" --output "{output}"`,
		script,
	)
}
