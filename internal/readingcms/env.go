package readingcms

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// ScriptEnv returns process environment for Python pipeline scripts.
// Loads repo .env and course overlay (.env.es / .env.en); later files override earlier keys.
func (p *Paths) ScriptEnv(courseCode string) []string {
	env := environMap(os.Environ())
	for _, rel := range p.envFilesForCourse(courseCode) {
		mergeEnvFile(env, filepath.Join(p.RepoRoot, rel))
	}
	return mapToEnviron(env)
}

func (p *Paths) envFilesForCourse(courseCode string) []string {
	files := []string{".env"}
	switch strings.ToLower(strings.TrimSpace(courseCode)) {
	case "es_ru":
		files = append(files, ".env.es")
	case "en_ru", "":
		files = append(files, ".env.en")
	}
	return files
}

func mergeEnvFile(env map[string]string, path string) {
	values, err := godotenv.Read(path)
	if err != nil {
		return
	}
	for k, v := range values {
		env[k] = v
	}
}

func environMap(environ []string) map[string]string {
	m := make(map[string]string, len(environ))
	for _, kv := range environ {
		i := strings.IndexByte(kv, '=')
		if i <= 0 {
			continue
		}
		m[kv[:i]] = kv[i+1:]
	}
	return m
}

func mapToEnviron(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}

// CourseCodeFromDir infers course_code from an absolute course grammar directory path.
func CourseCodeFromDir(courseDir string) string {
	lower := strings.ToLower(filepath.ToSlash(courseDir))
	if strings.Contains(lower, "spanish-grammar") {
		return "es_ru"
	}
	return "en_ru"
}
