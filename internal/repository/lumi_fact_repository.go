package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// LumiFactRepository manages rotating "Lumi knows" facts.
type LumiFactRepository struct {
	db *sql.DB
}

func NewLumiFactRepository(db *sql.DB) *LumiFactRepository {
	return &LumiFactRepository{db: db}
}

// LumiFact is one fact row.
type LumiFact struct {
	ID          int64  `json:"id"`
	CourseCode  string `json:"course_code"`
	Context     string `json:"context"`
	Locale      string `json:"locale"`
	Body        string `json:"body"`
	Status      string `json:"status"`
	LastShownOn string `json:"last_shown_on,omitempty"`
	ShownCount  int    `json:"shown_count"`
	CreatedAt   string `json:"created_at,omitempty"`
}

var lumiFactContexts = map[string]bool{
	"general": true, "grammar": true, "reading": true,
	"practice": true, "progress": true, "city": true,
}

// NormalizeLumiContext maps unknown contexts to "general".
func NormalizeLumiContext(context string) string {
	context = strings.TrimSpace(strings.ToLower(context))
	if lumiFactContexts[context] {
		return context
	}
	return "general"
}

// GetDailyFact returns the fact of the day for (course, context, locale),
// rotating to the least recently shown fact at the first request of a day.
// Fallback order: exact context → general; course → any course ('').
func (r *LumiFactRepository) GetDailyFact(ctx context.Context, courseCode, factContext, locale string) (*LumiFact, error) {
	factContext = NormalizeLumiContext(factContext)
	today := time.Now().Format("2006-01-02")

	tryScopes := [][2]string{{courseCode, factContext}}
	if factContext != "general" {
		tryScopes = append(tryScopes, [2]string{courseCode, "general"})
	}
	tryScopes = append(tryScopes, [2]string{"", factContext})
	if factContext != "general" {
		tryScopes = append(tryScopes, [2]string{"", "general"})
	}

	for _, scope := range tryScopes {
		fact, err := r.dailyFactForScope(ctx, scope[0], scope[1], locale, today)
		if err != nil {
			return nil, err
		}
		if fact != nil {
			return fact, nil
		}
	}
	return nil, nil
}

func (r *LumiFactRepository) dailyFactForScope(ctx context.Context, courseCode, factContext, locale, today string) (*LumiFact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin lumi fact pick: %w", err)
	}
	defer tx.Rollback()

	scan := func(row *sql.Row) (*LumiFact, error) {
		var f LumiFact
		if err := row.Scan(&f.ID, &f.CourseCode, &f.Context, &f.Locale, &f.Body); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return &f, nil
	}

	// Already chosen today?
	fact, err := scan(tx.QueryRowContext(ctx, `
		SELECT id, course_code, context, locale, body
		FROM lumi_facts
		WHERE course_code = ? AND context = ? AND locale = ? AND status = 'active'
			AND last_shown_on = CAST(? AS date)
		ORDER BY id LIMIT 1
	`, courseCode, factContext, locale, today))
	if err != nil {
		return nil, fmt.Errorf("lumi fact today lookup: %w", err)
	}
	if fact != nil {
		return fact, tx.Commit()
	}

	// Rotate: least recently shown first.
	fact, err = scan(tx.QueryRowContext(ctx, `
		SELECT id, course_code, context, locale, body
		FROM lumi_facts
		WHERE course_code = ? AND context = ? AND locale = ? AND status = 'active'
		ORDER BY last_shown_on ASC NULLS FIRST, id
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, courseCode, factContext, locale))
	if err != nil {
		return nil, fmt.Errorf("lumi fact rotate lookup: %w", err)
	}
	if fact == nil {
		return nil, tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE lumi_facts
		SET last_shown_on = CAST(? AS date), shown_count = shown_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, today, fact.ID); err != nil {
		return nil, fmt.Errorf("lumi fact rotate update: %w", err)
	}
	return fact, tx.Commit()
}

// LumiFactFilter filters the admin list.
type LumiFactFilter struct {
	CourseCode string
	Context    string
	Locale     string
	Status     string
	Limit      int
	Offset     int
}

// List returns facts for the admin table.
func (r *LumiFactRepository) List(ctx context.Context, f LumiFactFilter) ([]LumiFact, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if f.CourseCode != "" {
		where = append(where, "course_code = ?")
		args = append(args, f.CourseCode)
	}
	if f.Context != "" {
		where = append(where, "context = ?")
		args = append(args, f.Context)
	}
	if f.Locale != "" {
		where = append(where, "locale = ?")
		args = append(args, f.Locale)
	}
	if f.Status != "" {
		where = append(where, "status = ?")
		args = append(args, f.Status)
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM lumi_facts WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count lumi facts: %w", err)
	}

	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	query := `
		SELECT id, course_code, context, locale, body, status,
			COALESCE(CAST(last_shown_on AS text), ''), shown_count, CAST(created_at AS text)
		FROM lumi_facts WHERE ` + cond + `
		ORDER BY id DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, append(args, f.Limit, f.Offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list lumi facts: %w", err)
	}
	defer rows.Close()
	out := []LumiFact{}
	for rows.Next() {
		var fact LumiFact
		if err := rows.Scan(&fact.ID, &fact.CourseCode, &fact.Context, &fact.Locale, &fact.Body, &fact.Status, &fact.LastShownOn, &fact.ShownCount, &fact.CreatedAt); err != nil {
			return nil, 0, err
		}
		if len(fact.LastShownOn) > 10 {
			fact.LastShownOn = fact.LastShownOn[:10]
		}
		out = append(out, fact)
	}
	return out, total, rows.Err()
}

// BulkInsert adds many facts at once (admin textarea form).
func (r *LumiFactRepository) BulkInsert(ctx context.Context, courseCode, factContext, locale string, bodies []string, createdBy int64) (int, error) {
	factContext = NormalizeLumiContext(factContext)
	if locale == "" {
		locale = "ru"
	}
	inserted := 0
	for _, body := range bodies {
		body = strings.TrimSpace(body)
		if body == "" {
			continue
		}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO lumi_facts (course_code, context, locale, body, created_by)
			VALUES (?, ?, ?, ?, ?)
		`, courseCode, factContext, locale, body, nullableID(createdBy)); err != nil {
			return inserted, fmt.Errorf("insert lumi fact: %w", err)
		}
		inserted++
	}
	return inserted, nil
}

// Update edits body/status/context/course of one fact.
func (r *LumiFactRepository) Update(ctx context.Context, fact LumiFact) error {
	if fact.ID == 0 {
		return fmt.Errorf("fact id is required")
	}
	if strings.TrimSpace(fact.Body) == "" {
		return fmt.Errorf("fact body is required")
	}
	status := fact.Status
	if status != "archived" {
		status = "active"
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE lumi_facts
		SET course_code = ?, context = ?, locale = ?, body = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, fact.CourseCode, NormalizeLumiContext(fact.Context), fact.Locale, strings.TrimSpace(fact.Body), status, fact.ID)
	if err != nil {
		return fmt.Errorf("update lumi fact: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func nullableID(id int64) interface{} {
	if id == 0 {
		return nil
	}
	return id
}
