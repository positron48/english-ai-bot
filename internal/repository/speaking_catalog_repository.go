package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// SpeakingCatalogSnapshot mirrors bundle index + task metadata.
type SpeakingCatalogSnapshot struct {
	Version    string
	Categories map[string]*SpeakingCategorySnapshot
}

// SpeakingCategorySnapshot is one speaking category row.
type SpeakingCategorySnapshot struct {
	CategoryID        string
	Title             string
	TitleTranslations map[string]string
	Level             string
	Order             int
	TaskIDs           []string
}

// SpeakingTaskDocument is the public task payload (without evaluation hints for client).
type SpeakingTaskDocument struct {
	ID             string `json:"id"`
	CategoryID     string `json:"category_id"`
	Level          string `json:"level"`
	Type           string `json:"type"`
	TargetLanguage string `json:"target_language"`
	Title          string `json:"title"`
	PromptRU       string `json:"prompt_ru"`
	DisplayText    string `json:"display_text,omitempty"`
	QuestionES     string `json:"question_es,omitempty"`
	MaxAttempts    int    `json:"max_attempts"`
	Order          int    `json:"order"`
}

// SpeakingTaskFull includes evaluator-only fields from stored JSON.
type SpeakingTaskFull struct {
	SpeakingTaskDocument
	ExpectedMeaningRU string   `json:"expected_meaning_ru,omitempty"`
	AcceptableAnswers []string `json:"acceptable_answers,omitempty"`
	EvaluationNotes   string   `json:"evaluation_notes,omitempty"`
}

// SpeakingCatalogRepository persists speaking tasks synced from bundle.
type SpeakingCatalogRepository struct {
	db *sql.DB
}

func NewSpeakingCatalogRepository(db *sql.DB) *SpeakingCatalogRepository {
	if db == nil {
		return nil
	}
	return &SpeakingCatalogRepository{db: db}
}

func (r *SpeakingCatalogRepository) CountCategories() (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	var n int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM speaking_categories`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("speaking_categories count: %w", err)
	}
	return n, nil
}

func (r *SpeakingCatalogRepository) LoadSnapshot() (*SpeakingCatalogSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("speaking catalog repo: nil db")
	}
	rows, err := r.db.Query(`
SELECT category_id, title, COALESCE(title_translations, ''), level, sort_order, task_ids
FROM speaking_categories
ORDER BY sort_order ASC, category_id ASC`)
	if err != nil {
		return nil, fmt.Errorf("speaking_categories query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := &SpeakingCatalogSnapshot{
		Version:    "1.0.0",
		Categories: make(map[string]*SpeakingCategorySnapshot),
	}
	for rows.Next() {
		var catID, title, titleJSON, level, taskIDsJSON string
		var sortOrder int
		if err := rows.Scan(&catID, &title, &titleJSON, &level, &sortOrder, &taskIDsJSON); err != nil {
			return nil, fmt.Errorf("speaking_categories scan: %w", err)
		}
		var titleTrans map[string]string
		if strings.TrimSpace(titleJSON) != "" {
			_ = json.Unmarshal([]byte(titleJSON), &titleTrans)
		}
		if titleTrans == nil {
			titleTrans = map[string]string{}
		}
		var taskIDs []string
		if err := json.Unmarshal([]byte(taskIDsJSON), &taskIDs); err != nil {
			return nil, fmt.Errorf("speaking_categories task_ids %q: %w", catID, err)
		}
		out.Categories[catID] = &SpeakingCategorySnapshot{
			CategoryID:        catID,
			Title:             title,
			TitleTranslations: titleTrans,
			Level:             level,
			Order:             sortOrder,
			TaskIDs:           taskIDs,
		}
	}
	return out, rows.Err()
}

func (r *SpeakingCatalogRepository) GetTaskFull(taskID string) (*SpeakingTaskFull, bool, error) {
	if r == nil || r.db == nil {
		return nil, false, nil
	}
	const q = `SELECT task_json FROM speaking_tasks WHERE task_id = ?`
	var raw string
	err := r.db.QueryRow(q, taskID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("speaking_tasks get: %w", err)
	}
	var full SpeakingTaskFull
	if err := json.Unmarshal([]byte(raw), &full); err != nil {
		return nil, false, fmt.Errorf("speaking task json: %w", err)
	}
	return &full, true, nil
}

func (r *SpeakingCatalogRepository) GetTaskPublic(taskID string) (*SpeakingTaskDocument, bool, error) {
	full, ok, err := r.GetTaskFull(taskID)
	if err != nil || !ok {
		return nil, ok, err
	}
	doc := full.SpeakingTaskDocument
	return &doc, true, nil
}

func (r *SpeakingCatalogRepository) ListTasksByCategory(categoryID string) ([]SpeakingTaskDocument, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("speaking catalog repo: nil db")
	}
	rows, err := r.db.Query(`
SELECT task_json FROM speaking_tasks
WHERE category_id = ?
ORDER BY sort_order ASC, task_id ASC`, categoryID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []SpeakingTaskDocument
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var full SpeakingTaskFull
		if err := json.Unmarshal([]byte(raw), &full); err != nil {
			continue
		}
		out = append(out, full.SpeakingTaskDocument)
	}
	return out, rows.Err()
}

func (r *SpeakingCatalogRepository) ReplaceCatalog(
	version, generatedAt string,
	categories []SpeakingCategoryUpsert,
	tasks []SpeakingTaskUpsert,
) error {
	return r.ReplaceCatalogForTargetLanguage("", version, generatedAt, categories, tasks)
}

func (r *SpeakingCatalogRepository) ReplaceCatalogForTargetLanguage(
	targetLanguage, version, generatedAt string,
	categories []SpeakingCategoryUpsert,
	tasks []SpeakingTaskUpsert,
) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("speaking catalog repo: nil db")
	}
	targetLanguage = strings.TrimSpace(strings.ToLower(targetLanguage))
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if targetLanguage == "" {
		if _, err := tx.Exec(`DELETE FROM speaking_tasks`); err != nil {
			return fmt.Errorf("truncate speaking_tasks: %w", err)
		}
		if _, err := tx.Exec(`DELETE FROM speaking_categories`); err != nil {
			return fmt.Errorf("truncate speaking_categories: %w", err)
		}
	} else {
		if _, err := tx.Exec(`DELETE FROM speaking_tasks WHERE LOWER(target_language) = ?`, targetLanguage); err != nil {
			return fmt.Errorf("delete speaking_tasks for %q: %w", targetLanguage, err)
		}
		if _, err := tx.Exec(`DELETE FROM speaking_categories WHERE LOWER(category_id) LIKE ? ESCAPE '\'`, targetLanguage+`\_%`); err != nil {
			return fmt.Errorf("delete speaking_categories for %q: %w", targetLanguage, err)
		}
	}

	for _, c := range categories {
		titleTrans, err := json.Marshal(c.TitleTranslations)
		if err != nil {
			return err
		}
		taskIDs, err := json.Marshal(c.TaskIDs)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`
INSERT INTO speaking_categories (category_id, title, title_translations, level, sort_order, task_ids)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(category_id) DO UPDATE SET
  title = excluded.title,
  title_translations = excluded.title_translations,
  level = excluded.level,
  sort_order = excluded.sort_order,
  task_ids = excluded.task_ids`,
			c.CategoryID, c.Title, string(titleTrans), c.Level, c.SortOrder, string(taskIDs),
		); err != nil {
			return fmt.Errorf("insert speaking_categories %q: %w", c.CategoryID, err)
		}
	}

	for _, t := range tasks {
		if targetLanguage != "" && strings.TrimSpace(strings.ToLower(t.TargetLanguage)) != targetLanguage {
			return fmt.Errorf("speaking task %q target_language=%q does not match import target %q", t.TaskID, t.TargetLanguage, targetLanguage)
		}
		if _, err := tx.Exec(`
INSERT INTO speaking_tasks (task_id, category_id, title, level, task_type, target_language, sort_order, task_json, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(task_id) DO UPDATE SET
  category_id = excluded.category_id,
  title = excluded.title,
  level = excluded.level,
  task_type = excluded.task_type,
  target_language = excluded.target_language,
  sort_order = excluded.sort_order,
  task_json = excluded.task_json,
  updated_at = CURRENT_TIMESTAMP`,
			t.TaskID, t.CategoryID, t.Title, t.Level, t.TaskType, t.TargetLanguage, t.SortOrder, t.TaskJSON,
		); err != nil {
			return fmt.Errorf("insert speaking_tasks %q: %w", t.TaskID, err)
		}
	}

	_ = version
	_ = generatedAt
	return tx.Commit()
}

type SpeakingCategoryUpsert struct {
	CategoryID        string
	Title             string
	TitleTranslations map[string]string
	Level             string
	SortOrder         int
	TaskIDs           []string
}

type SpeakingTaskUpsert struct {
	TaskID         string
	CategoryID     string
	Title          string
	Level          string
	TaskType       string
	TargetLanguage string
	SortOrder      int
	TaskJSON       string
}
