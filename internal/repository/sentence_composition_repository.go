package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// SentenceCompositionRepository persists daily sentence-composition sets, items and
// per-word participation counters.
type SentenceCompositionRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewSentenceCompositionRepository(db *sql.DB, logger *zap.Logger) *SentenceCompositionRepository {
	return &SentenceCompositionRepository{db: db, logger: logger}
}

// SelectCandidateWords returns well-learned words for the course (review state with stored
// mastering_score >= minMastery, excluding "known" words), ordered by least participation in
// this feature so far (used_count ASC, last_used_on ASC NULLS FIRST), then random.
func (r *SentenceCompositionRepository) SelectCandidateWords(userID int64, courseCode string, minMastery, limit int) ([]models.SentenceWordCandidate, error) {
	if minMastery < 0 {
		minMastery = 0
	}
	query := `
		SELECT tc.word_card_id,
		       COALESCE(MAX(wc.word), '') AS lemma,
		       MAX(tc.word_ru) AS word_ru,
		       COALESCE(MAX(swu.used_count), 0) AS used_count
		FROM user_cards uc
		JOIN training_cards tc ON uc.training_card_id = tc.id
		JOIN word_cards wc ON tc.word_card_id = wc.id
		LEFT JOIN user_word_mastering uwm ON uwm.user_id = uc.user_id AND uwm.word_card_id = tc.word_card_id
		LEFT JOIN sentence_word_usage swu ON swu.user_id = uc.user_id AND swu.word_card_id = tc.word_card_id AND swu.course_code = ?
		WHERE uc.user_id = ?
		  AND uc.state = 'review'
		  AND (? = '' OR uc.course_code = ?)
		  AND COALESCE(uwm.mastering_score, 0) >= ?
		  AND NOT EXISTS (
		      SELECT 1 FROM user_word_knowledge uwk
		      WHERE uwk.user_id = uc.user_id AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
		  )
		GROUP BY tc.word_card_id
		ORDER BY used_count ASC, MAX(COALESCE(swu.last_used_on, '1970-01-01')) ASC, RANDOM()
		LIMIT ?`
	rows, err := r.db.Query(query, courseCode, userID, courseCode, courseCode, minMastery, limit)
	if err != nil {
		return nil, fmt.Errorf("select candidate words: %w", err)
	}
	defer rows.Close()
	var out []models.SentenceWordCandidate
	for rows.Next() {
		var c models.SentenceWordCandidate
		var lemma, wordRU sql.NullString
		if err := rows.Scan(&c.WordCardID, &lemma, &wordRU, &c.UsedCount); err != nil {
			return nil, fmt.Errorf("scan candidate word: %w", err)
		}
		c.Lemma = lemma.String
		c.Translation = wordRU.String
		if c.Lemma == "" {
			continue
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// LatestSet returns the user's most recent set for the course, or nil if none exist.
// Used both for the regeneration guard and for "today's available training".
func (r *SentenceCompositionRepository) LatestSet(userID int64, courseCode string) (*models.SentenceSet, error) {
	row := r.db.QueryRow(`
		SELECT id, user_id, course_code, CAST(generation_date AS text), COALESCE(scopes_json::text, '[]'),
		       status, started_at, completed_at, star_count, passed_count, failed_count, created_at
		FROM sentence_sets
		WHERE user_id = ? AND course_code = ?
		ORDER BY generation_date DESC, id DESC
		LIMIT 1`, userID, courseCode)
	return scanSentenceSet(row)
}

// CreateSet inserts a new set with its items and bumps per-word participation counters,
// all in one transaction. usedWordIDs is the set of word_card_ids that participated.
func (r *SentenceCompositionRepository) CreateSet(set *models.SentenceSet, items []models.SentenceItem, usedWordIDs []int64) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	scopesJSON, _ := json.Marshal(set.Scopes)
	var setID int64
	if err := tx.QueryRow(`
		INSERT INTO sentence_sets (user_id, course_code, generation_date, scopes_json, status)
		VALUES (?, ?, CAST(? AS date), CAST(? AS jsonb), 'ready')
		RETURNING id`,
		set.UserID, set.CourseCode, set.GenerationDate, string(scopesJSON)).Scan(&setID); err != nil {
		return 0, fmt.Errorf("insert set: %w", err)
	}

	for _, it := range items {
		idsJSON, _ := json.Marshal(it.WordCardIDs)
		if _, err := tx.Exec(`
			INSERT INTO sentence_items (set_id, position, prompt_ru, reference_es, word_card_ids)
			VALUES (?, ?, ?, ?, CAST(? AS jsonb))`,
			setID, it.Position, it.PromptRU, it.ReferenceES, string(idsJSON)); err != nil {
			return 0, fmt.Errorf("insert item: %w", err)
		}
	}

	today := set.GenerationDate
	for _, wid := range usedWordIDs {
		if _, err := tx.Exec(`
			INSERT INTO sentence_word_usage (user_id, word_card_id, course_code, used_count, last_used_on)
			VALUES (?, ?, ?, 1, CAST(? AS date))
			ON CONFLICT (user_id, word_card_id, course_code) DO UPDATE SET
				used_count = sentence_word_usage.used_count + 1,
				last_used_on = EXCLUDED.last_used_on`,
			set.UserID, wid, set.CourseCode, today); err != nil {
			return 0, fmt.Errorf("bump word usage: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return setID, nil
}

// GetSetByID loads a set by id.
func (r *SentenceCompositionRepository) GetSetByID(setID int64) (*models.SentenceSet, error) {
	row := r.db.QueryRow(`
		SELECT id, user_id, course_code, CAST(generation_date AS text), COALESCE(scopes_json::text, '[]'),
		       status, started_at, completed_at, star_count, passed_count, failed_count, created_at
		FROM sentence_sets WHERE id = ?`, setID)
	return scanSentenceSet(row)
}

// GetItems returns all items of a set ordered by position.
func (r *SentenceCompositionRepository) GetItems(setID int64) ([]models.SentenceItem, error) {
	rows, err := r.db.Query(`
		SELECT id, set_id, position, prompt_ru, reference_es, COALESCE(word_card_ids::text, '[]'),
		       attempted_at, user_input, error_count, outcome, grading_json::text
		FROM sentence_items WHERE set_id = ? ORDER BY position ASC`, setID)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()
	var out []models.SentenceItem
	for rows.Next() {
		it, err := scanSentenceItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	return out, rows.Err()
}

// GetItemWithSet loads a single item plus its parent set (for ownership and grading context).
func (r *SentenceCompositionRepository) GetItemWithSet(itemID int64) (*models.SentenceItem, *models.SentenceSet, error) {
	itemRow := r.db.QueryRow(`
		SELECT id, set_id, position, prompt_ru, reference_es, COALESCE(word_card_ids::text, '[]'),
		       attempted_at, user_input, error_count, outcome, grading_json::text
		FROM sentence_items WHERE id = ?`, itemID)
	it, err := scanSentenceItem(itemRow)
	if err != nil {
		return nil, nil, err
	}
	set, err := r.GetSetByID(it.SetID)
	if err != nil {
		return nil, nil, err
	}
	return it, set, nil
}

// MarkStarted flips a 'ready' set to 'started' and stamps started_at (the consumption marker).
// No-op when already started/completed.
func (r *SentenceCompositionRepository) MarkStarted(setID int64) error {
	_, err := r.db.Exec(`
		UPDATE sentence_sets
		SET status = 'started', started_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'ready'`, setID)
	if err != nil {
		return fmt.Errorf("mark started: %w", err)
	}
	return nil
}

// RecordAttempt persists a single-shot grading result for an item and rolls the outcome up
// into the set counters, completing the set when every item has been attempted. It returns
// the updated set. The update is idempotent-guarded: an already-attempted item is rejected.
func (r *SentenceCompositionRepository) RecordAttempt(itemID int64, userInput string, errorCount int, outcome, gradingJSON string) (*models.SentenceSet, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var setID int64
	res, err := tx.Exec(`
		UPDATE sentence_items
		SET attempted_at = CURRENT_TIMESTAMP, user_input = ?, error_count = ?, outcome = ?, grading_json = CAST(? AS jsonb)
		WHERE id = ? AND attempted_at IS NULL`,
		userInput, errorCount, outcome, gradingJSON, itemID)
	if err != nil {
		return nil, fmt.Errorf("update item: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, fmt.Errorf("item %d already attempted or missing", itemID)
	}
	if err := tx.QueryRow(`SELECT set_id FROM sentence_items WHERE id = ?`, itemID).Scan(&setID); err != nil {
		return nil, fmt.Errorf("lookup set id: %w", err)
	}

	// Recompute set counters and completion from item rows (authoritative, avoids drift).
	if _, err := tx.Exec(`
		UPDATE sentence_sets s SET
			star_count   = agg.stars,
			passed_count = agg.passed,
			failed_count = agg.failed,
			completed_at = CASE WHEN agg.attempted = agg.total THEN CURRENT_TIMESTAMP ELSE s.completed_at END,
			status       = CASE WHEN agg.attempted = agg.total THEN 'completed'
			                    WHEN agg.attempted > 0 THEN 'started' ELSE s.status END
		FROM (
			SELECT
				COUNT(*) AS total,
				COUNT(attempted_at) AS attempted,
				COUNT(*) FILTER (WHERE outcome = 'star') AS stars,
				COUNT(*) FILTER (WHERE outcome = 'passed') AS passed,
				COUNT(*) FILTER (WHERE outcome = 'failed') AS failed
			FROM sentence_items WHERE set_id = ?
		) agg
		WHERE s.id = ?`, setID, setID); err != nil {
		return nil, fmt.Errorf("update set counters: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return r.GetSetByID(setID)
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanSentenceSet(row rowScanner) (*models.SentenceSet, error) {
	var s models.SentenceSet
	var scopesJSON string
	var startedAt, completedAt sql.NullTime
	err := row.Scan(&s.ID, &s.UserID, &s.CourseCode, &s.GenerationDate, &scopesJSON,
		&s.Status, &startedAt, &completedAt, &s.StarCount, &s.PassedCount, &s.FailedCount, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan set: %w", err)
	}
	_ = json.Unmarshal([]byte(scopesJSON), &s.Scopes)
	if startedAt.Valid {
		s.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		s.CompletedAt = &completedAt.Time
	}
	return &s, nil
}

func scanSentenceItem(row rowScanner) (*models.SentenceItem, error) {
	var it models.SentenceItem
	var idsJSON string
	var attemptedAt sql.NullTime
	var userInput, outcome, gradingJSON sql.NullString
	var errorCount sql.NullInt64
	err := row.Scan(&it.ID, &it.SetID, &it.Position, &it.PromptRU, &it.ReferenceES, &idsJSON,
		&attemptedAt, &userInput, &errorCount, &outcome, &gradingJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan item: %w", err)
	}
	_ = json.Unmarshal([]byte(idsJSON), &it.WordCardIDs)
	if attemptedAt.Valid {
		it.AttemptedAt = &attemptedAt.Time
	}
	if userInput.Valid {
		it.UserInput = &userInput.String
	}
	if errorCount.Valid {
		n := int(errorCount.Int64)
		it.ErrorCount = &n
	}
	if outcome.Valid {
		it.Outcome = &outcome.String
	}
	if gradingJSON.Valid {
		it.GradingJSON = &gradingJSON.String
	}
	return &it, nil
}

// UserSetOverview is a per-user summary of sentence-composition activity for the admin list.
type UserSetOverview struct {
	UserID           int64  `json:"user_id"`
	TelegramID       int64  `json:"telegram_id"`
	TelegramUsername string `json:"telegram_username"`
	SubscriptionTier string `json:"subscription_tier"`
	SetCount         int    `json:"set_count"`
	LastGenerationOn string `json:"last_generation_on"`
	TotalStars       int    `json:"total_stars"`
	TotalPassed      int    `json:"total_passed"`
	TotalFailed      int    `json:"total_failed"`
}

// ListUserOverviews returns one row per user that has at least one sentence set, ordered by
// most recent activity. Used by the admin "results by users" screen.
func (r *SentenceCompositionRepository) ListUserOverviews() ([]UserSetOverview, error) {
	rows, err := r.db.Query(`
		SELECT s.user_id,
		       COALESCE(u.telegram_id, 0),
		       COALESCE(u.telegram_username, ''),
		       COALESCE(u.subscription_tier, 'free'),
		       COUNT(*) AS set_count,
		       CAST(MAX(s.generation_date) AS text) AS last_generation_on,
		       COALESCE(SUM(s.star_count), 0),
		       COALESCE(SUM(s.passed_count), 0),
		       COALESCE(SUM(s.failed_count), 0)
		FROM sentence_sets s
		LEFT JOIN users u ON u.id = s.user_id
		GROUP BY s.user_id, u.telegram_id, u.telegram_username, u.subscription_tier
		ORDER BY last_generation_on DESC, s.user_id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list user overviews: %w", err)
	}
	defer rows.Close()
	var out []UserSetOverview
	for rows.Next() {
		var o UserSetOverview
		if err := rows.Scan(&o.UserID, &o.TelegramID, &o.TelegramUsername, &o.SubscriptionTier,
			&o.SetCount, &o.LastGenerationOn, &o.TotalStars, &o.TotalPassed, &o.TotalFailed); err != nil {
			return nil, fmt.Errorf("scan user overview: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ListSetsByUser returns a user's sets across all courses, most recent first (capped by limit).
func (r *SentenceCompositionRepository) ListSetsByUser(userID int64, limit int) ([]*models.SentenceSet, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`
		SELECT id, user_id, course_code, CAST(generation_date AS text), COALESCE(scopes_json::text, '[]'),
		       status, started_at, completed_at, star_count, passed_count, failed_count, created_at
		FROM sentence_sets
		WHERE user_id = ?
		ORDER BY generation_date DESC, id DESC
		LIMIT ?`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list sets by user: %w", err)
	}
	defer rows.Close()
	var out []*models.SentenceSet
	for rows.Next() {
		s, err := scanSentenceSet(rows)
		if err != nil {
			return nil, err
		}
		if s != nil {
			out = append(out, s)
		}
	}
	return out, rows.Err()
}

// ParticipationCount returns the stored participation counter for one word (testing/inspection).
func (r *SentenceCompositionRepository) ParticipationCount(userID, wordCardID int64, courseCode string) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COALESCE(used_count, 0) FROM sentence_word_usage WHERE user_id = ? AND word_card_id = ? AND course_code = ?`,
		userID, wordCardID, courseCode).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return n, err
}
