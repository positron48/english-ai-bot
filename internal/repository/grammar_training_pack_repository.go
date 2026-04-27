package repository

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/grammartrainingpack"

	"go.uber.org/zap"
)

type GrammarTrainingPackRepository struct {
	fs     fs.FS
	logger *zap.Logger
}

// GrammarTrainingPackIndex is a parsed training pack index.
// v1: chapters → one file per chapter (string or one-element list).
// v2: blocks → "chapterID::theoryBlockID" → path under chapters/ (per-theory file).
type GrammarTrainingPackIndex struct {
	Version     string            `json:"version"`
	Language    string            `json:"language"`
	CourseID    string            `json:"course_id"`
	GeneratedAt string            `json:"generated_at"`
	// Chapters is a legacy view: at most one file per chapter (for backward compatibility).
	Chapters    map[string]string `json:"chapters"`
	blockFiles  map[string]string
	chapterFiles map[string][]string
}

type GrammarTrainingPackChapter struct {
	ChapterID string                   `json:"chapter_id"`
	Questions []map[string]interface{} `json:"questions"`
}

func NewGrammarTrainingPackRepository(logger *zap.Logger) *GrammarTrainingPackRepository {
	return NewGrammarTrainingPackRepositoryWithFS(grammartrainingpack.FS, logger)
}

func NewGrammarTrainingPackRepositoryForLearning(lc config.LearningConfig, logger *zap.Logger) (*GrammarTrainingPackRepository, error) {
	fsys, err := grammartrainingpack.PackFS(lc.GrammarBundleID)
	if err != nil {
		return nil, err
	}
	return NewGrammarTrainingPackRepositoryWithFS(fsys, logger), nil
}

func NewGrammarTrainingPackRepositoryWithFS(filesystem fs.FS, logger *zap.Logger) *GrammarTrainingPackRepository {
	return &GrammarTrainingPackRepository{
		fs:     filesystem,
		logger: logger,
	}
}

func parseTrainingPackChapterFiles(raw json.RawMessage) (map[string][]string, error) {
	out := make(map[string][]string)
	if len(raw) == 0 || string(raw) == "null" {
		return out, nil
	}
	// v1: { "ch1": "file.json" }
	var oneFile map[string]string
	if err := json.Unmarshal(raw, &oneFile); err == nil {
		for k, v := range oneFile {
			if strings.TrimSpace(v) == "" {
				continue
			}
			out[k] = []string{v}
		}
		return out, nil
	}
	// v1.5: { "ch1": ["a.json", "b.json"] }
	var many map[string][]string
	if err := json.Unmarshal(raw, &many); err == nil {
		for k, files := range many {
			keep := make([]string, 0, len(files))
			for _, f := range files {
				if strings.TrimSpace(f) != "" {
					keep = append(keep, f)
				}
			}
			if len(keep) > 0 {
				out[k] = keep
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("chapters: expected object with string or string[] values")
}

func parseTrainingPackIndex(data []byte) (*GrammarTrainingPackIndex, error) {
	var top struct {
		Version     string            `json:"version"`
		Language    string            `json:"language"`
		CourseID    string            `json:"course_id"`
		GeneratedAt string            `json:"generated_at"`
		Chapters    json.RawMessage   `json:"chapters"`
		Blocks      map[string]string `json:"blocks"`
	}
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, err
	}
	cfl, err := parseTrainingPackChapterFiles(top.Chapters)
	if err != nil {
		return nil, err
	}
	if top.Blocks == nil {
		top.Blocks = map[string]string{}
	}
	legacy := make(map[string]string)
	for k, files := range cfl {
		if len(files) == 0 {
			continue
		}
		legacy[k] = files[0]
	}
	return &GrammarTrainingPackIndex{
		Version:      top.Version,
		Language:     top.Language,
		CourseID:     top.CourseID,
		GeneratedAt:  top.GeneratedAt,
		Chapters:     legacy,
		blockFiles:   top.Blocks,
		chapterFiles: cfl,
	}, nil
}

func (r *GrammarTrainingPackRepository) GetIndex() (*GrammarTrainingPackIndex, error) {
	data, err := fs.ReadFile(r.fs, "index.json")
	if err != nil {
		return nil, fmt.Errorf("read grammar training index: %w", err)
	}
	idx, err := parseTrainingPackIndex(data)
	if err != nil {
		return nil, fmt.Errorf("parse grammar training index: %w", err)
	}
	if idx.Chapters == nil {
		idx.Chapters = map[string]string{}
	}
	return idx, nil
}

// collectTrainingPackFilePaths returns sorted unique relative paths (under training pack root) to JSON question files.
func collectTrainingPackFilePaths(idx *GrammarTrainingPackIndex) []string {
	seen := make(map[string]struct{})
	if len(idx.blockFiles) > 0 {
		for _, p := range idx.blockFiles {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			seen[p] = struct{}{}
		}
	} else {
		for _, files := range idx.chapterFiles {
			for _, p := range files {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				seen[p] = struct{}{}
			}
		}
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// assignStableQuestionIDs makes each question id unique across the pack (per-file ids are often q1, q2, …).
func assignStableQuestionIDs(questions []map[string]interface{}) {
	for _, q := range questions {
		if q == nil {
			continue
		}
		ch, _ := q["chapter_id"].(string)
		tb, _ := q["theory_block_id"].(string)
		if strings.TrimSpace(ch) == "" || strings.TrimSpace(tb) == "" {
			continue
		}
		rid, has := q["id"]
		if !has {
			continue
		}
		q["id"] = fmt.Sprintf("%s::%s::%v", ch, tb, rid)
	}
}

func (r *GrammarTrainingPackRepository) readQuestionsFile(rel string) ([]map[string]interface{}, error) {
	rel = strings.TrimLeft(rel, "/")
	chapterPath := filepath.Join("chapters", rel)
	data, err := fs.ReadFile(r.fs, chapterPath)
	if err != nil {
		return nil, fmt.Errorf("read training file %s: %w", chapterPath, err)
	}
	var chapter GrammarTrainingPackChapter
	if err := json.Unmarshal(data, &chapter); err != nil {
		return nil, fmt.Errorf("parse training file %s: %w", chapterPath, err)
	}
	if chapter.Questions == nil {
		return nil, nil
	}
	return chapter.Questions, nil
}

func (r *GrammarTrainingPackRepository) HasAnyQuestions() (bool, int, error) {
	qs, err := r.GetAllQuestions()
	if err != nil {
		return false, 0, err
	}
	return len(qs) > 0, len(qs), nil
}

func (r *GrammarTrainingPackRepository) GetAllQuestions() ([]map[string]interface{}, error) {
	idx, err := r.GetIndex()
	if err != nil {
		return nil, err
	}
	paths := collectTrainingPackFilePaths(idx)
	out := make([]map[string]interface{}, 0, len(paths)*8)
	for _, p := range paths {
		qs, err := r.readQuestionsFile(p)
		if err != nil {
			if r.logger != nil {
				r.logger.Warn("skip invalid training file", zap.String("path", p), zap.Error(err))
			}
			continue
		}
		out = append(out, qs...)
	}
	assignStableQuestionIDs(out)
	return out, nil
}

// GetChapterQuestions returns all training questions for one grammar chapter.
func (r *GrammarTrainingPackRepository) GetChapterQuestions(chapterID string) ([]map[string]interface{}, error) {
	idx, err := r.GetIndex()
	if err != nil {
		return nil, err
	}
	var paths []string
	if len(idx.blockFiles) > 0 {
		prefix := chapterID + "::"
		for k, rel := range idx.blockFiles {
			if strings.HasPrefix(k, prefix) {
				if rel = strings.TrimSpace(rel); rel != "" {
					paths = append(paths, rel)
				}
			}
		}
	} else {
		paths = append(paths, idx.chapterFiles[chapterID]...)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("chapter %s not found in training index", chapterID)
	}
	seen := make(map[string]struct{})
	uniq := make([]string, 0, len(paths))
	for _, p := range paths {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		uniq = append(uniq, p)
	}
	sort.Strings(uniq)

	out := make([]map[string]interface{}, 0, len(uniq)*8)
	for _, p := range uniq {
		qs, err := r.readQuestionsFile(p)
		if err != nil {
			if r.logger != nil {
				r.logger.Warn("skip invalid training file in chapter", zap.String("chapter_id", chapterID), zap.String("path", p), zap.Error(err))
			}
			continue
		}
		out = append(out, qs...)
	}
	assignStableQuestionIDs(out)
	return out, nil
}

func (r *GrammarTrainingPackRepository) QuestionsByTheoryBlock() (map[string][]map[string]interface{}, error) {
	all, err := r.GetAllQuestions()
	if err != nil {
		return nil, err
	}
	out := make(map[string][]map[string]interface{})
	for _, q := range all {
		theoryBlockID, _ := q["theory_block_id"].(string)
		if strings.TrimSpace(theoryBlockID) == "" {
			continue
		}
		out[theoryBlockID] = append(out[theoryBlockID], q)
	}
	return out, nil
}
