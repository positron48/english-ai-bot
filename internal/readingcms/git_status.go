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
		"-C", course.GrammarDir,
		"status",
		"--porcelain=v1",
		"-z",
		"--untracked-files=all",
		"--",
	}
	textIDByRelPath := make(map[string]string, len(idx.Texts))
	textIDByAssetDir := make(map[string]string, len(idx.Texts))
	for textID, rel := range idx.Texts {
		repoRel := courseTextGitRelPath(rel)
		if repoRel == "" {
			continue
		}
		repoRel = filepath.ToSlash(repoRel)
		args = append(args, repoRel)
		textIDByRelPath[repoRel] = textID
		if course.CourseDir != "" {
			textIDByRelPath[filepath.ToSlash(filepath.Join(course.CourseDir, filepath.FromSlash(repoRel)))] = textID
		}

		assetDir := courseTextAssetsGitRelPath(textID)
		if assetDir == "" {
			continue
		}
		assetDir = filepath.ToSlash(assetDir)
		args = append(args, assetDir)
		textIDByAssetDir[strings.TrimRight(assetDir, "/")+"/"] = textID
		if course.CourseDir != "" {
			rootRelAssetDir := filepath.ToSlash(filepath.Join(course.CourseDir, filepath.FromSlash(assetDir)))
			textIDByAssetDir[strings.TrimRight(rootRelAssetDir, "/")+"/"] = textID
		}
	}
	if len(textIDByRelPath) == 0 && len(textIDByAssetDir) == 0 {
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
			for assetDir, id := range textIDByAssetDir {
				if strings.HasPrefix(path, assetDir) {
					textID = id
					break
				}
			}
		}
		if textID == "" {
			continue
		}
		if st := classifyGitStatus(xy); st != "" {
			statusByTextID[textID] = mergeGitStatus(statusByTextID[textID], st)
		}
	}
	return statusByTextID
}

func courseTextGitRelPath(indexRelPath string) string {
	rel := strings.TrimSpace(indexRelPath)
	if rel == "" || strings.Contains(rel, "..") || strings.HasPrefix(rel, "/") {
		return ""
	}
	return filepath.Join("reading", filepath.FromSlash(rel))
}

func courseTextAssetsGitRelPath(textID string) string {
	textID = strings.TrimSpace(textID)
	if textID == "" || strings.Contains(textID, "..") || strings.ContainsAny(textID, `/\`) {
		return ""
	}
	return filepath.Join("assets", "reading", textID)
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

func mergeGitStatus(existing, next string) string {
	if gitStatusRank(next) > gitStatusRank(existing) {
		return next
	}
	return existing
}

func gitStatusRank(st string) int {
	switch st {
	case "conflict":
		return 7
	case "untracked", "added":
		return 6
	case "renamed", "copied":
		return 5
	case "modified":
		return 4
	case "deleted":
		return 3
	case "":
		return 0
	default:
		return 1
	}
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
