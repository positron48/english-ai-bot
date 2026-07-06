package web

import (
	"database/sql"
	"fmt"
)

// vocabSummaryBreakdown is a mutually exclusive per-word partition of the learner vocabulary.
// Each word is assigned exactly one bucket so new+learning+review+mastered == total.
type vocabSummaryBreakdown struct {
	Total    int
	New      int
	Learning int
	Review   int
	Mastered int
}

// userCardStateRank orders SRS states from weakest (0) to strongest (2).
func userCardStateRank(state string) int {
	switch state {
	case "new":
		return 0
	case "learning":
		return 1
	case "review":
		return 2
	default:
		return 1
	}
}

// vocabWordSummaryBucket assigns a word to exactly one summary bucket.
// minCardRank is the weakest rank among all user_cards for the word.
// When the word is marked known and every card is at least review, it counts as mastered.
func vocabWordSummaryBucket(minCardRank int, isKnown bool) int {
	if minCardRank < 0 {
		minCardRank = 0
	}
	if minCardRank > 2 {
		minCardRank = 2
	}
	if isKnown && minCardRank >= 2 {
		return 3
	}
	return minCardRank
}

func (r *Router) queryVocabSummary(userID int64, courseCode string) (vocabSummaryBreakdown, error) {
	courseFilter := ""
	args := []interface{}{userID}
	if courseCode != "" {
		courseFilter = ` AND LOWER(wc.course_code) = LOWER(?)`
		args = append(args, courseCode)
	}

	knownArgs := []interface{}{userID}
	if courseCode != "" {
		knownArgs = append(knownArgs, courseCode)
	}

	query := `
WITH per_word AS (
	SELECT
		tc.word_card_id,
		MIN(CASE uc.state
			WHEN 'new' THEN 0
			WHEN 'learning' THEN 1
			WHEN 'review' THEN 2
			ELSE 1
		END) AS min_card_rank,
		MAX(CASE WHEN uwk.word_card_id IS NOT NULL THEN 1 ELSE 0 END) AS is_known
	FROM user_cards uc
	JOIN training_cards tc ON uc.training_card_id = tc.id
	JOIN word_cards wc ON wc.id = tc.word_card_id
	LEFT JOIN user_word_knowledge uwk
		ON uwk.user_id = uc.user_id
		AND uwk.word_card_id = tc.word_card_id
		AND uwk.status = 'known'
	WHERE uc.user_id = ?` + courseFilter + `
	GROUP BY tc.word_card_id

	UNION ALL

	SELECT
		uwk.word_card_id,
		2 AS min_card_rank,
		1 AS is_known
	FROM user_word_knowledge uwk
	JOIN word_cards wc ON wc.id = uwk.word_card_id
	WHERE uwk.user_id = ? AND uwk.status = 'known'` + courseFilter + `
		AND NOT EXISTS (
			SELECT 1
			FROM user_cards uc
			JOIN training_cards tc ON uc.training_card_id = tc.id
			WHERE uc.user_id = uwk.user_id AND tc.word_card_id = uwk.word_card_id
		)
),
classified AS (
	SELECT
		word_card_id,
		CASE
			WHEN is_known = 1 AND min_card_rank >= 2 THEN 3
			ELSE min_card_rank
		END AS bucket
	FROM per_word
)
SELECT
	COUNT(*) AS total,
	COALESCE(SUM(CASE WHEN bucket = 0 THEN 1 ELSE 0 END), 0) AS new_count,
	COALESCE(SUM(CASE WHEN bucket = 1 THEN 1 ELSE 0 END), 0) AS learning_count,
	COALESCE(SUM(CASE WHEN bucket = 2 THEN 1 ELSE 0 END), 0) AS review_count,
	COALESCE(SUM(CASE WHEN bucket = 3 THEN 1 ELSE 0 END), 0) AS mastered_count
FROM classified`

	args = append(args, knownArgs...)

	var out vocabSummaryBreakdown
	err := r.db.QueryRow(query, args...).Scan(
		&out.Total,
		&out.New,
		&out.Learning,
		&out.Review,
		&out.Mastered,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return vocabSummaryBreakdown{}, nil
		}
		return vocabSummaryBreakdown{}, fmt.Errorf("query vocab summary: %w", err)
	}
	return out, nil
}
