package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// MasteringScoreRecentK is the number of last answers to use for recent accuracy (plan constant K).
const MasteringScoreRecentK = 20

// UserWordPair identifies a user and word for mastering stats.
type UserWordPair struct {
	UserID    int64
	WordCardID int64
}

// WordMasteringStatsRow holds aggregate stats for computing mastering score (from review_events).
type WordMasteringStatsRow struct {
	UserID        int64
	WordCardID    int64
	Total         int64
	Correct       int64
	RecentTotal   int64
	RecentCorrect int64
}

// UserWordMasteringRepository handles user_word_mastering table and batch stats from review_events.
type UserWordMasteringRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserWordMasteringRepository creates a new repository.
func NewUserWordMasteringRepository(db *sql.DB, logger *zap.Logger) *UserWordMasteringRepository {
	return &UserWordMasteringRepository{db: db, logger: logger}
}

// GetWordMasteringStatsBatch returns (total, correct, recent_total, recent_correct) per (user_id, word_card_id)
// in one query with window functions. No per-word queries. K = MasteringScoreRecentK.
// Pairs not present in the result have no review_events (caller should treat as 0,0,0,0).
func (r *UserWordMasteringRepository) GetWordMasteringStatsBatch(pairs []UserWordPair) (map[UserWordPair]WordMasteringStatsRow, error) {
	out := make(map[UserWordPair]WordMasteringStatsRow)
	if len(pairs) == 0 {
		return out, nil
	}

	// Build VALUES (?,?), (?,?), ... for (user_id, word_card_id). Cast to bigint for PostgreSQL.
	var valParts []string
	var args []interface{}
	for i, p := range pairs {
		if i > 0 {
			valParts = append(valParts, ", ")
		}
		valParts = append(valParts, "(?::bigint, ?::bigint)")
		args = append(args, p.UserID, p.WordCardID)
	}
	valuesClause := strings.Join(valParts, "")

	// K for recent window
	args = append(args, MasteringScoreRecentK, MasteringScoreRecentK)

	query := `WITH ev AS (
		SELECT re.user_id, tc.word_card_id, re.is_correct,
			ROW_NUMBER() OVER (PARTITION BY re.user_id, tc.word_card_id ORDER BY re.answered_at DESC NULLS LAST, re.id DESC) AS rn,
			COUNT(*) OVER (PARTITION BY re.user_id, tc.word_card_id) AS total,
			COALESCE(SUM(re.is_correct) OVER (PARTITION BY re.user_id, tc.word_card_id), 0) AS correct
		FROM review_events re
		JOIN user_cards uc ON uc.id = re.user_card_id
		JOIN training_cards tc ON uc.training_card_id = tc.id
		WHERE re.answered_at IS NOT NULL
	),
	pairs(user_id, word_card_id) AS (VALUES ` + valuesClause + `)
	SELECT ev.user_id, ev.word_card_id,
		MAX(ev.total) AS total,
		MAX(ev.correct) AS correct,
		SUM(CASE WHEN ev.rn <= ? THEN 1 ELSE 0 END) AS recent_total,
		SUM(CASE WHEN ev.rn <= ? AND ev.is_correct = 1 THEN 1 ELSE 0 END) AS recent_correct
	FROM ev
	INNER JOIN pairs p ON ev.user_id = p.user_id AND ev.word_card_id = p.word_card_id
	GROUP BY ev.user_id, ev.word_card_id`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get word mastering stats batch: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var row WordMasteringStatsRow
		if err := rows.Scan(&row.UserID, &row.WordCardID, &row.Total, &row.Correct, &row.RecentTotal, &row.RecentCorrect); err != nil {
			return nil, fmt.Errorf("scan word mastering stats: %w", err)
		}
		out[UserWordPair{UserID: row.UserID, WordCardID: row.WordCardID}] = row
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating word mastering stats: %w", err)
	}
	return out, nil
}

// GetWordCardIDsBySessionID returns distinct (user_id, word_card_id) that had review_events in this session.
func (r *UserWordMasteringRepository) GetWordCardIDsBySessionID(sessionID int64) ([]UserWordPair, error) {
	query := `SELECT DISTINCT re.user_id, tc.word_card_id
	FROM review_events re
	JOIN user_cards uc ON re.user_card_id = uc.id
	JOIN training_cards tc ON uc.training_card_id = tc.id
	WHERE re.session_id = ?`
	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session word pairs: %w", err)
	}
	defer rows.Close()
	var pairs []UserWordPair
	for rows.Next() {
		var p UserWordPair
		if err := rows.Scan(&p.UserID, &p.WordCardID); err != nil {
			return nil, fmt.Errorf("scan session word pair: %w", err)
		}
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating session word pairs: %w", err)
	}
	return pairs, nil
}

// GetKnownWordCardIDsForUser returns which of the given word_card_ids are marked known for the user.
func (r *UserWordMasteringRepository) GetKnownWordCardIDsForUser(userID int64, wordCardIDs []int64) (map[int64]bool, error) {
	out := make(map[int64]bool)
	if len(wordCardIDs) == 0 {
		return out, nil
	}
	placeholders := make([]string, len(wordCardIDs))
	args := make([]interface{}, 0, len(wordCardIDs)+1)
	args = append(args, userID)
	for i, id := range wordCardIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query := `SELECT word_card_id FROM user_word_knowledge WHERE user_id = ? AND status = 'known' AND word_card_id IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get known word card ids: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan known word_card_id: %w", err)
		}
		out[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating known word card ids: %w", err)
	}
	return out, nil
}

// GetKnownForPairs returns which of the given (user_id, word_card_id) pairs are marked known. One query.
func (r *UserWordMasteringRepository) GetKnownForPairs(pairs []UserWordPair) (map[UserWordPair]bool, error) {
	out := make(map[UserWordPair]bool)
	if len(pairs) == 0 {
		return out, nil
	}
	var valParts []string
	var args []interface{}
	for i, p := range pairs {
		if i > 0 {
			valParts = append(valParts, ", ")
		}
		valParts = append(valParts, "(?::bigint, ?::bigint)")
		args = append(args, p.UserID, p.WordCardID)
	}
	query := `SELECT user_id, word_card_id FROM user_word_knowledge WHERE status = 'known' AND (user_id, word_card_id) IN (VALUES ` + strings.Join(valParts, "") + `)`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get known for pairs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p UserWordPair
		if err := rows.Scan(&p.UserID, &p.WordCardID); err != nil {
			return nil, fmt.Errorf("scan known pair: %w", err)
		}
		out[p] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating known pairs: %w", err)
	}
	return out, nil
}

// Upsert inserts or updates a single (user_id, word_card_id) with the given score.
func (r *UserWordMasteringRepository) Upsert(userID, wordCardID int64, score int) error {
	_, err := r.db.Exec(`INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, word_card_id) DO UPDATE SET
			mastering_score = EXCLUDED.mastering_score,
			updated_at = CURRENT_TIMESTAMP`, userID, wordCardID, score)
	if err != nil {
		return fmt.Errorf("upsert user_word_mastering: %w", err)
	}
	return nil
}

// UpsertBatch inserts or updates mastering_score for the given entries (batch, no per-row query).
func (r *UserWordMasteringRepository) UpsertBatch(entries []struct {
	UserID     int64
	WordCardID int64
	Score      int
}) error {
	if len(entries) == 0 {
		return nil
	}
	// PostgreSQL: INSERT ... ON CONFLICT (user_id, word_card_id) DO UPDATE
	// Build VALUES (?,?,?,CURRENT_TIMESTAMP), ...
	var valParts []string
	var args []interface{}
	for i, e := range entries {
		if i > 0 {
			valParts = append(valParts, ", ")
		}
		valParts = append(valParts, "(?, ?, ?, CURRENT_TIMESTAMP)")
		args = append(args, e.UserID, e.WordCardID, e.Score)
	}
	query := `INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score, updated_at)
	VALUES ` + strings.Join(valParts, "") + `
	ON CONFLICT (user_id, word_card_id) DO UPDATE SET
		mastering_score = EXCLUDED.mastering_score,
		updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("upsert user_word_mastering batch: %w", err)
	}
	return nil
}

// GetScore returns the stored mastering_score for (user_id, word_card_id). If no row, returns 0, nil.
func (r *UserWordMasteringRepository) GetScore(userID, wordCardID int64) (int, error) {
	var score int
	err := r.db.QueryRow(`SELECT mastering_score FROM user_word_mastering WHERE user_id = ? AND word_card_id = ?`, userID, wordCardID).Scan(&score)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get mastering score: %w", err)
	}
	return score, nil
}
