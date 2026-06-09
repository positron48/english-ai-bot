package main

import (
	"context"
	"database/sql"
)

// countMissingWordSRS counts legacy user_cards that map to a learning_item but have no srs_items row.
// Multiple directions per word collapse to one canonical srs_item; raw COUNT(user_cards) is not comparable.
func countMissingWordSRS(ctx context.Context, db *sql.DB) (int64, error) {
	exists, err := tableExists(ctx, db, "user_cards")
	if err != nil || !exists {
		return 0, err
	}
	var count int64
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM user_cards ucards
		JOIN training_cards tc ON tc.id = ucards.training_card_id
		JOIN user_courses ucourse ON ucourse.user_id = ucards.user_id
		JOIN learning_items li ON li.course_id = ucourse.course_id
			AND li.source_kind = 'word_card'
			AND li.source_id = CAST(tc.word_card_id AS TEXT)
		WHERE NOT EXISTS (
			SELECT 1 FROM srs_items si
			WHERE si.user_course_id = ucourse.id AND si.learning_item_id = li.id
		)
	`).Scan(&count)
	return count, err
}

// countMissingGrammarSRS counts grammar_theory_memory rows mapped to learning_items without srs_items.
func countMissingGrammarSRS(ctx context.Context, db *sql.DB) (int64, error) {
	exists, err := tableExists(ctx, db, "grammar_theory_memory")
	if err != nil || !exists {
		return 0, err
	}
	var count int64
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM grammar_theory_memory gtm
		JOIN user_courses ucourse ON ucourse.user_id = gtm.user_id
		JOIN courses c ON c.id = ucourse.course_id
		JOIN learning_items li ON li.course_id = c.id
			AND li.source_kind = 'grammar_theory_block'
			AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
		WHERE NOT EXISTS (
			SELECT 1 FROM srs_items si
			WHERE si.user_course_id = ucourse.id AND si.learning_item_id = li.id
		)
	`).Scan(&count)
	return count, err
}
