package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/grammarbundle"

	"go.uber.org/zap"
)

// GrammarContentRepository handles reading grammar course content from embedded filesystem
type GrammarContentRepository struct {
	fs       fs.FS
	db       *sql.DB
	bundleID string
	logger   *zap.Logger
}

// BundleVersionHash returns a stable content hash for the bundle index and section list.
func (r *GrammarContentRepository) BundleVersionHash() (string, error) {
	if r.db != nil {
		var hash string
		err := r.db.QueryRow(`SELECT source_hash FROM grammar_content_bundle_meta WHERE bundle_id = ?`, r.bundleID).Scan(&hash)
		if err != nil {
			return "", fmt.Errorf("db grammar bundle hash %q: %w", r.bundleID, err)
		}
		return hash, nil
	}
	sections, err := fs.ReadFile(r.fs, "sections.json")
	if err != nil {
		return "", fmt.Errorf("failed to read sections.json: %w", err)
	}
	index, err := fs.ReadFile(r.fs, "index.json")
	if err != nil {
		return "", fmt.Errorf("failed to read index.json: %w", err)
	}
	sum := sha256.New()
	_, _ = sum.Write(sections)
	_, _ = sum.Write([]byte{0})
	_, _ = sum.Write(index)
	return hex.EncodeToString(sum.Sum(nil)), nil
}

// NewGrammarContentRepository creates a new grammar content repository using the default embedded English bundle.
func NewGrammarContentRepository(logger *zap.Logger) *GrammarContentRepository {
	return NewGrammarContentRepositoryWithFS(grammarbundle.FS, logger)
}

// NewGrammarContentRepositoryForLearning selects embedded bundle by GRAMMAR_BUNDLE_ID or an on-disk tree via GRAMMAR_BUNDLE_DIR.
func NewGrammarContentRepositoryForLearning(lc config.LearningConfig, logger *zap.Logger) (*GrammarContentRepository, error) {
	fsys, err := grammarFSForLearning(lc)
	if err != nil {
		return nil, err
	}
	return NewGrammarContentRepositoryWithFS(fsys, logger), nil
}

func grammarFSForLearning(lc config.LearningConfig) (fs.FS, error) {
	dir := strings.TrimSpace(lc.GrammarBundleDir)
	if dir != "" {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("grammar bundle dir: %w", err)
		}
		return os.DirFS(abs), nil
	}
	return grammarbundle.BundleFS(lc.GrammarBundleID)
}

// NewGrammarContentRepositoryWithFS creates a grammar content repository with the given fs (for tests or custom bundles).
func NewGrammarContentRepositoryWithFS(filesystem fs.FS, logger *zap.Logger) *GrammarContentRepository {
	return &GrammarContentRepository{
		fs:       filesystem,
		bundleID: "",
		logger:   logger,
	}
}

// NewGrammarContentRepositoryFromDB creates a repository that reads grammar content from PostgreSQL.
func NewGrammarContentRepositoryFromDB(db *sql.DB, bundleID string, logger *zap.Logger) *GrammarContentRepository {
	return &GrammarContentRepository{
		db:       db,
		bundleID: strings.TrimSpace(strings.ToLower(bundleID)),
		logger:   logger,
	}
}

// Section represents a grammar course section (category)
type Section struct {
	SectionID         string            `json:"section_id"`
	Title             string            `json:"title"`
	TitleTranslations map[string]string `json:"title_translations,omitempty"`
	Level             string            `json:"level"`
	Order             int               `json:"order"`
	ChapterIDs        []string          `json:"chapter_ids"`
}

// SectionsData represents the sections configuration
type SectionsData struct {
	Version  string    `json:"version"`
	Sections []Section `json:"sections"`
}

// Chapter represents a grammar chapter
type Chapter struct {
	SchemaVersion      string                 `json:"schema_version"`
	ID                 string                 `json:"id"`
	SectionID          string                 `json:"section_id"`
	Title              string                 `json:"title"`
	TitleTranslations  map[string]string      `json:"title_translations,omitempty"`
	TitleShort         string                 `json:"title_short,omitempty"`
	Description        string                 `json:"description,omitempty"`
	UILanguage         string                 `json:"ui_language"`
	TargetLanguage     string                 `json:"target_language"`
	Level              string                 `json:"level,omitempty"`
	Order              int                    `json:"order"`
	Prerequisites      []string               `json:"prerequisites,omitempty"`
	LearningObjectives []string               `json:"learning_objectives,omitempty"`
	EstimatedMinutes   int                    `json:"estimated_minutes,omitempty"`
	Blocks             []interface{}          `json:"blocks"`
	QuestionBank       map[string]interface{} `json:"question_bank"`
	ChapterTest        map[string]interface{} `json:"chapter_test"`
	Meta               map[string]interface{} `json:"meta,omitempty"`
}

// IndexData represents the bundle index
type IndexData struct {
	Version     string            `json:"version"`
	GeneratedAt string            `json:"generated_at"`
	Chapters    map[string]string `json:"chapters"`
}

// GetSections returns all sections from the bundle
func (r *GrammarContentRepository) GetSections() (*SectionsData, error) {
	if r.db != nil {
		var raw string
		err := r.db.QueryRow(`SELECT sections_json FROM grammar_content_bundle_meta WHERE bundle_id = ?`, r.bundleID).Scan(&raw)
		if err != nil {
			return nil, fmt.Errorf("db grammar sections %q: %w", r.bundleID, err)
		}
		var sections SectionsData
		if err := json.Unmarshal([]byte(raw), &sections); err != nil {
			return nil, fmt.Errorf("parse db sections json: %w", err)
		}
		return &sections, nil
	}
	data, err := fs.ReadFile(r.fs, "sections.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read sections.json: %w", err)
	}

	var sections SectionsData
	if err := json.Unmarshal(data, &sections); err != nil {
		return nil, fmt.Errorf("failed to parse sections.json: %w", err)
	}

	return &sections, nil
}

// GetIndex returns the bundle index
func (r *GrammarContentRepository) GetIndex() (*IndexData, error) {
	if r.db != nil {
		var raw string
		err := r.db.QueryRow(`SELECT index_json FROM grammar_content_bundle_meta WHERE bundle_id = ?`, r.bundleID).Scan(&raw)
		if err != nil {
			return nil, fmt.Errorf("db grammar index %q: %w", r.bundleID, err)
		}
		var index IndexData
		if err := json.Unmarshal([]byte(raw), &index); err != nil {
			return nil, fmt.Errorf("parse db index json: %w", err)
		}
		return &index, nil
	}
	data, err := fs.ReadFile(r.fs, "index.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read index.json: %w", err)
	}

	var index IndexData
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index.json: %w", err)
	}

	return &index, nil
}

// GetChapter returns a chapter by ID
func (r *GrammarContentRepository) GetChapter(chapterID string) (*Chapter, error) {
	data, err := r.GetChapterRawJSON(chapterID)
	if err != nil {
		return nil, err
	}

	var chapter Chapter
	if err := json.Unmarshal(data, &chapter); err != nil {
		return nil, fmt.Errorf("failed to parse chapter JSON: %w", err)
	}

	return &chapter, nil
}

// GetChapterRawJSON returns the original chapter JSON by ID.
func (r *GrammarContentRepository) GetChapterRawJSON(chapterID string) ([]byte, error) {
	if r.db != nil {
		var raw string
		err := r.db.QueryRow(`SELECT raw_json FROM grammar_content_chapters WHERE bundle_id = ? AND chapter_id = ?`, r.bundleID, chapterID).Scan(&raw)
		if err != nil {
			return nil, fmt.Errorf("db chapter not found %s: %w", chapterID, err)
		}
		return []byte(raw), nil
	}
	// Get index to find the filename
	index, err := r.GetIndex()
	if err != nil {
		return nil, err
	}

	filename, exists := index.Chapters[chapterID]
	if !exists {
		return nil, fmt.Errorf("chapter not found: %s", chapterID)
	}

	// Read chapter file
	path := filepath.Join("chapters", filename)
	data, err := fs.ReadFile(r.fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read chapter file %s: %w", filename, err)
	}

	return data, nil
}

// ChapterExists checks if a chapter exists in the bundle
func (r *GrammarContentRepository) ChapterExists(chapterID string) bool {
	index, err := r.GetIndex()
	if err != nil {
		return false
	}
	_, exists := index.Chapters[chapterID]
	return exists
}

// GetAllChapterIDs returns all chapter IDs from the index
func (r *GrammarContentRepository) GetAllChapterIDs() ([]string, error) {
	index, err := r.GetIndex()
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(index.Chapters))
	for id := range index.Chapters {
		ids = append(ids, id)
	}

	return ids, nil
}
