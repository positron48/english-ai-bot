package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// resetWordItems removes word-training content (learning_items where source_kind='word_card',
// their srs_items, and their legacy_content_mappings) for the given source's course.
//
// This is the cleanup phase required before re-importing corrected word content via
// --phase=content. It must be run when the existing word learning_items are wrong
// (e.g. es_ru was accidentally populated with English word_cards instead of Spanish ones).
//
// Exercise_attempts are intentionally NOT deleted: they represent real user activity history
// and the learning_item_id FK is expected to be SET NULL / already nullable.
func resetWordItems(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "reset-word-items"}
	for _, src := range sources {
		courseCode, ok := sourceCourseCodes[src.Label]
		if !ok {
			summary.Skipped++
			continue
		}
		var courseID int64
		if err := targetDB.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = $1`, courseCode).Scan(&courseID); err != nil {
			return nil, fmt.Errorf("resolve course %s: %w", courseCode, err)
		}

		// 1. Delete srs_items that reference word learning_items for this course.
		res, err := targetDB.ExecContext(ctx, `
			DELETE FROM srs_items
			WHERE user_course_id IN (
				SELECT id FROM user_courses WHERE course_id = $1
			)
			AND learning_item_id IN (
				SELECT id FROM learning_items
				WHERE course_id = $1 AND source_kind = 'word_card'
			)
		`, courseID)
		if err != nil {
			return nil, fmt.Errorf("delete srs_items for %s: %w", courseCode, err)
		}
		if affected, _ := res.RowsAffected(); affected > 0 {
			summary.SRSScanned += affected // repurpose field for deleted count
		}

		// 2. Delete legacy_content_mappings for word learning_items of this course.
		_, err = targetDB.ExecContext(ctx, `
			DELETE FROM legacy_content_mappings
			WHERE source_app_code = $1
			  AND source_db_label = $1
			  AND source_table = 'learning_items'
			  AND target_course_id = $2
			  AND source_kind = 'word_card'
		`, src.Label, courseID)
		if err != nil {
			return nil, fmt.Errorf("delete legacy_content_mappings for %s: %w", courseCode, err)
		}

		// 3. Delete the word learning_items themselves (CASCADE removes srs_items if FK is CASCADE;
		//    otherwise they were already deleted above).
		res, err = targetDB.ExecContext(ctx, `
			DELETE FROM learning_items
			WHERE course_id = $1 AND source_kind = 'word_card'
		`, courseID)
		if err != nil {
			return nil, fmt.Errorf("delete word learning_items for %s: %w", courseCode, err)
		}
		if affected, _ := res.RowsAffected(); affected > 0 {
			summary.ItemsScanned += affected // repurpose field for deleted count
		}

		summary.MappingsCreated++ // one course processed
	}
	return summary, nil
}

// importWordCardsFromSource upserts all word_cards from the source DB into the target DB.
// Conflict resolution: ON CONFLICT(word) DO UPDATE to enrich content fields (definition_ru,
// transcription, etc.) but only when the target value is NULL (COALESCE pattern).
// Returns map of source_word_card_id → target_word_card_id (built by joining on word text).
func importWordCardsFromSource(ctx context.Context, sourceDB, targetDB *sql.DB, summary *writeSummary) (map[int64]int64, error) {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT id, word,
		       definition, pos, noun_gender, opposite_gender_word,
		       transcription, definition_ru, examples_json, verb_forms_json, display_en
		FROM word_cards
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sourceWords := map[int64]string{} // source_id → word text (for map build below)
	for rows.Next() {
		var srcID int64
		var word string
		var definition, pos, nounGender, oppositeGenderWord sql.NullString
		var transcription, definitionRU, examplesJSON, verbFormsJSON, displayEN sql.NullString
		if err := rows.Scan(
			&srcID, &word,
			&definition, &pos, &nounGender, &oppositeGenderWord,
			&transcription, &definitionRU, &examplesJSON, &verbFormsJSON, &displayEN,
		); err != nil {
			return nil, err
		}

		// Upsert into unified: keep existing values if not NULL (COALESCE).
		_, err := targetDB.ExecContext(ctx, `
			INSERT INTO word_cards (
				word, definition, pos, noun_gender, opposite_gender_word,
				transcription, definition_ru, examples_json, verb_forms_json, display_en,
				updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10,
				CURRENT_TIMESTAMP
			)
			ON CONFLICT (word) DO UPDATE SET
				definition       = COALESCE(word_cards.definition, excluded.definition),
				pos              = COALESCE(word_cards.pos, excluded.pos),
				noun_gender      = COALESCE(word_cards.noun_gender, excluded.noun_gender),
				opposite_gender_word = COALESCE(word_cards.opposite_gender_word, excluded.opposite_gender_word),
				transcription    = COALESCE(word_cards.transcription, excluded.transcription),
				definition_ru    = COALESCE(word_cards.definition_ru, excluded.definition_ru),
				examples_json    = COALESCE(word_cards.examples_json, excluded.examples_json),
				verb_forms_json  = COALESCE(word_cards.verb_forms_json, excluded.verb_forms_json),
				display_en       = COALESCE(word_cards.display_en, excluded.display_en),
				updated_at       = CURRENT_TIMESTAMP
		`,
			word,
			nullableString(definition), nullableString(pos), nullableString(nounGender), nullableString(oppositeGenderWord),
			nullableString(transcription), nullableString(definitionRU),
			func() interface{} {
				if examplesJSON.Valid && examplesJSON.String != "" {
					return examplesJSON.String
				}
				return nil
			}(),
			func() interface{} {
				if verbFormsJSON.Valid && verbFormsJSON.String != "" {
					return verbFormsJSON.String
				}
				return nil
			}(),
			nullableString(displayEN),
		)
		if err != nil {
			return nil, fmt.Errorf("upsert word_card %q: %w", word, err)
		}
		sourceWords[srcID] = word
		if summary != nil {
			summary.ItemsScanned++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build source_id → target_id map by re-querying unified word_cards by word text.
	wcMap := map[int64]int64{}
	for srcID, word := range sourceWords {
		var targetID int64
		if err := targetDB.QueryRowContext(ctx,
			`SELECT id FROM word_cards WHERE word = $1`, word,
		).Scan(&targetID); err != nil {
			// Word wasn't found after upsert — skip silently.
			continue
		}
		wcMap[srcID] = targetID
		if summary != nil {
			summary.ItemsInserted++
		}
	}
	return wcMap, nil
}

// wordCardIDFromString parses a source_id string to int64.
func wordCardIDFromString(s string) (int64, bool) {
	v, err := strconv.ParseInt(s, 10, 64)
	return v, err == nil
}
