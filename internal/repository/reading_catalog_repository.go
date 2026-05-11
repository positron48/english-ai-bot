package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ReadingCatalogSnapshot is the reading catalog loaded from DB (mirrors bundle index + text rows).
type ReadingCatalogSnapshot struct {
	Version     string
	GeneratedAt string
	Categories  map[string]*ReadingCategorySnapshot
}

// ReadingCategorySnapshot is one category row for API handlers.
type ReadingCategorySnapshot struct {
	CategoryID        string
	Title             string
	TitleTranslations map[string]string
	Level             string
	Order             int
	TextIDs           []string
}

// ReadingTextDocument is one reading text document (same shape as bundle JSON file).
type ReadingTextDocument struct {
	ID                string                 `json:"id"`
	CategoryID        string                 `json:"category_id"`
	Title             string                 `json:"title"`
	TitleTranslations map[string]string      `json:"title_translations,omitempty"`
	Level             string                 `json:"level"`
	TargetLanguage    string                 `json:"target_language"`
	ReadingPassage    map[string]interface{} `json:"reading_passage"`
}

// ReadingCatalogRepository persists reading texts/categories synced from the grammar bundle.
type ReadingCatalogRepository struct {
	db *sql.DB
}

func NewReadingCatalogRepository(db *sql.DB) *ReadingCatalogRepository {
	if db == nil {
		return nil
	}
	return &ReadingCatalogRepository{db: db}
}

// CountCategories returns how many categories exist (0 means DB catalog not populated).
func (r *ReadingCatalogRepository) CountCategories() (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	var n int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM reading_categories`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("reading_categories count: %w", err)
	}
	return n, nil
}

// LoadSnapshot loads categories (ordered) for building the public reading index.
func (r *ReadingCatalogRepository) LoadSnapshot() (*ReadingCatalogSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("reading catalog repo: nil db")
	}
	rows, err := r.db.Query(`
SELECT category_id, title, COALESCE(title_translations, ''), level, sort_order, text_ids
FROM reading_categories
ORDER BY sort_order ASC, category_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("reading_categories query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := &ReadingCatalogSnapshot{
		Version:     "1.0.0",
		GeneratedAt: "",
		Categories:  make(map[string]*ReadingCategorySnapshot),
	}
	for rows.Next() {
		var catID, title, titleJSON, level, textIDsJSON string
		var sortOrder int
		if err := rows.Scan(&catID, &title, &titleJSON, &level, &sortOrder, &textIDsJSON); err != nil {
			return nil, fmt.Errorf("reading_categories scan: %w", err)
		}
		var titleTrans map[string]string
		if strings.TrimSpace(titleJSON) != "" {
			_ = json.Unmarshal([]byte(titleJSON), &titleTrans)
		}
		if titleTrans == nil {
			titleTrans = map[string]string{}
		}
		var textIDs []string
		if err := json.Unmarshal([]byte(textIDsJSON), &textIDs); err != nil {
			return nil, fmt.Errorf("reading_categories text_ids %q: %w", catID, err)
		}
		out.Categories[catID] = &ReadingCategorySnapshot{
			CategoryID:        catID,
			Title:             title,
			TitleTranslations: titleTrans,
			Level:             level,
			Order:             sortOrder,
			TextIDs:           textIDs,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTextDocument returns a text by id when present in DB.
func (r *ReadingCatalogRepository) GetTextDocument(textID string) (*ReadingTextDocument, bool, error) {
	if r == nil || r.db == nil {
		return nil, false, nil
	}
	const q = `
SELECT text_id, category_id, title, COALESCE(title_translations, ''), level, target_language, reading_passage
FROM reading_texts
WHERE text_id = ?`
	var tid, catID, title, titleJSON, level, targetLang, passageJSON string
	err := r.db.QueryRow(q, textID).Scan(&tid, &catID, &title, &titleJSON, &level, &targetLang, &passageJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("reading_texts get: %w", err)
	}
	var titleTrans map[string]string
	if strings.TrimSpace(titleJSON) != "" {
		_ = json.Unmarshal([]byte(titleJSON), &titleTrans)
	}
	if titleTrans == nil {
		titleTrans = map[string]string{}
	}
	var passage map[string]interface{}
	if err := json.Unmarshal([]byte(passageJSON), &passage); err != nil {
		return nil, false, fmt.Errorf("reading_passage json: %w", err)
	}
	doc := &ReadingTextDocument{
		ID:                tid,
		CategoryID:        catID,
		Title:             title,
		TitleTranslations: titleTrans,
		Level:             level,
		TargetLanguage:    targetLang,
		ReadingPassage:    passage,
	}
	return doc, true, nil
}

// AllTextIDs returns every text_id (for bootstrap token scan).
func (r *ReadingCatalogRepository) AllTextIDs() ([]string, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	rows, err := r.db.Query(`SELECT text_id FROM reading_texts ORDER BY text_id`)
	if err != nil {
		return nil, fmt.Errorf("reading_texts list ids: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// ReplaceCatalog replaces the entire reading catalog from bundle-derived rows (transactional).
func (r *ReadingCatalogRepository) ReplaceCatalog(
	version, generatedAt string,
	categories []ReadingCategoryUpsert,
	texts []ReadingTextUpsert,
) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("reading catalog repo: nil db")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM reading_texts`); err != nil {
		return fmt.Errorf("truncate reading_texts: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM reading_categories`); err != nil {
		return fmt.Errorf("truncate reading_categories: %w", err)
	}

	for _, c := range categories {
		titleTrans, err := json.Marshal(c.TitleTranslations)
		if err != nil {
			return err
		}
		textIDs, err := json.Marshal(c.TextIDs)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
INSERT INTO reading_categories (category_id, title, title_translations, level, sort_order, text_ids)
VALUES (?, ?, ?, ?, ?, ?)`,
			c.CategoryID,
			c.Title,
			string(titleTrans),
			c.Level,
			c.SortOrder,
			string(textIDs),
		); err != nil {
			return fmt.Errorf("insert reading_categories %q: %w", c.CategoryID, err)
		}
	}

	for _, t := range texts {
		titleTrans, err := json.Marshal(t.TitleTranslations)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
INSERT INTO reading_texts (text_id, category_id, title, title_translations, level, target_language, reading_passage, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			t.TextID,
			t.CategoryID,
			t.Title,
			string(titleTrans),
			t.Level,
			t.TargetLanguage,
			t.ReadingPassageJSON,
		); err != nil {
			return fmt.Errorf("insert reading_texts %q: %w", t.TextID, err)
		}
	}

	if _, err := tx.Exec(`
DELETE FROM reading_text_progress p
WHERE NOT EXISTS (SELECT 1 FROM reading_texts t WHERE t.text_id = p.chapter_id)`); err != nil {
		return fmt.Errorf("cleanup orphan reading_text_progress: %w", err)
	}

	_ = version
	_ = generatedAt

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// ReadingCategoryUpsert is one category row for ReplaceCatalog.
type ReadingCategoryUpsert struct {
	CategoryID        string
	Title             string
	TitleTranslations map[string]string
	Level             string
	SortOrder         int
	TextIDs           []string
}

// ReadingTextUpsert is one text row for ReplaceCatalog.
type ReadingTextUpsert struct {
	TextID               string
	CategoryID           string
	Title                string
	TitleTranslations    map[string]string
	Level                string
	TargetLanguage       string
	ReadingPassageJSON   string
}
