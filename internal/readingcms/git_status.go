package readingcms

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
)

func (s *Service) gitStatusesForCourseTexts(course Course, idx *readingIndex) map[string]string {
	if idx == nil || len(idx.Texts) == 0 {
		return nil
	}
	args := []string{
		"-C", s.paths.RepoRoot,
		"status",
		"--porcelain=v1",
		"-z",
		"--untracked-files=all",
		"--",
	}
	textIDByRelPath := make(map[string]string, len(idx.Texts))
	for textID, rel := range idx.Texts {
		repoRel := courseTextRepoRelPath(course, rel)
		if repoRel == "" {
			continue
		}
		repoRel = filepath.ToSlash(repoRel)
		args = append(args, repoRel)
		textIDByRelPath[repoRel] = textID
	}
	if len(textIDByRelPath) == 0 {
		return nil
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil
	}
	statusByTextID := map[string]string{}
	for _, entry := range bytes.Split(out, []byte{0}) {
		if len(entry) < 4 {
			continue
		}
		xy := string(entry[:2])
		path := filepath.ToSlash(string(entry[3:]))
		if i := strings.LastIndex(path, " -> "); i >= 0 {
			path = path[i+4:]
		}
		textID := textIDByRelPath[path]
		if textID == "" {
			continue
		}
		if st := classifyGitStatus(xy); st != "" {
			statusByTextID[textID] = st
		}
	}
	return statusByTextID
}

func courseTextRepoRelPath(course Course, indexRelPath string) string {
	rel := strings.TrimSpace(indexRelPath)
	if rel == "" || strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
		return ""
	}
	return filepath.Join(course.CourseDir, "reading", filepath.FromSlash(rel))
}

func classifyGitStatus(xy string) string {
	if xy == "??" {
		return "untracked"
	}
	if strings.Contains(xy, "A") {
		return "added"
	}
	if strings.Contains(xy, "R") {
		return "renamed"
	}
	if strings.Contains(xy, "C") {
		return "copied"
	}
	if strings.Contains(xy, "M") {
		return "modified"
	}
	if strings.Contains(xy, "D") {
		return "deleted"
	}
	if strings.Contains(xy, "U") {
		return "conflict"
	}
	return ""
}

func isNewUncommittedGitStatus(st string) bool {
	return st == "untracked" || st == "added"
}

func matchesGitFilter(st, filter string) bool {
	switch strings.ToLower(strings.TrimSpace(filter)) {
	case "", "all":
		return true
	case "committed":
		return st == ""
	case "uncommitted":
		return st != ""
	case "new":
		return isNewUncommittedGitStatus(st)
	case "changed":
		return st != "" && !isNewUncommittedGitStatus(st)
	default:
		return true
	}
}
