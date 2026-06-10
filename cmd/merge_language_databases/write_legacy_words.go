package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// mergeLegacyWords copies the word-training legacy tables (training_cards, user_cards,
// user_word_mastering, user_word_knowledge) from each source prod DB into the unified target,
// course-scoping every row with the course_code that corresponds to the source (en_ru / es_ru).
//
// review_events are intentionally NOT copied: the training engine does not read them to
// determine what cards to show; dashboard history falls back to canonical exercise_attempts.
//
// Idempotent: re-running the phase is safe — training_cards are looked up by
// (word_card_id, sense_index, course_code), user_cards by (user_id, training_card_id,
// direction), mastering/knowledge by (user_id, word_card_id) before inserting.
func mergeLegacyWords(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "legacy-words"}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		courseCode, ok := sourceCourseCodes[src.Label]
		if !ok {
			summary.Skipped++
			continue
		}

		// Build word_card_id map: source wc id → target wc id (join on word text).
		wcMap, err := buildWordCardMap(ctx, src.DB, targetDB)
		if err != nil {
			return nil, fmt.Errorf("build word_card map for %s: %w", src.Label, err)
		}

		// Build user_id map from legacy_user_mappings.
		userMap, err := buildUserMap(ctx, targetDB, src.Label)
		if err != nil {
			return nil, fmt.Errorf("build user map for %s: %w", src.Label, err)
		}

		// Copy training_cards; build source→target id map.
		tcMap, err := copyTrainingCards(ctx, src.DB, targetDB, courseCode, wcMap, summary)
		if err != nil {
			return nil, fmt.Errorf("copy training_cards from %s: %w", src.Label, err)
		}

		if err := copyUserCards(ctx, src.DB, targetDB, courseCode, userMap, tcMap, summary); err != nil {
			return nil, fmt.Errorf("copy user_cards from %s: %w", src.Label, err)
		}

		if err := copyUserWordMastering(ctx, src.DB, targetDB, courseCode, userMap, wcMap, summary); err != nil {
			return nil, fmt.Errorf("copy user_word_mastering from %s: %w", src.Label, err)
		}

		if err := copyUserWordKnowledge(ctx, src.DB, targetDB, courseCode, userMap, wcMap, summary); err != nil {
			return nil, fmt.Errorf("copy user_word_knowledge from %s: %w", src.Label, err)
		}
	}
	return summary, nil
}

// buildWordCardMap returns source_word_card_id → target_word_card_id by joining on word text.
// word_cards.word is UNIQUE in the target (unified) DB, so the join is 1:1.
func buildWordCardMap(ctx context.Context, sourceDB, targetDB *sql.DB) (map[int64]int64, error) {
	// Load target word → id.
	rows, err := targetDB.QueryContext(ctx, `SELECT id, word FROM word_cards`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	targetWordToID := map[string]int64{}
	for rows.Next() {
		var id int64
		var word string
		if err := rows.Scan(&id, &word); err != nil {
			return nil, err
		}
		targetWordToID[word] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Map source ids.
	srows, err := sourceDB.QueryContext(ctx, `SELECT id, word FROM word_cards`)
	if err != nil {
		return nil, err
	}
	defer srows.Close()
	m := map[int64]int64{}
	for srows.Next() {
		var id int64
		var word string
		if err := srows.Scan(&id, &word); err != nil {
			return nil, err
		}
		if targetID, ok := targetWordToID[word]; ok {
			m[id] = targetID
		}
		// Words with no match in target are silently skipped; training_cards that reference
		// them will also be skipped in copyTrainingCards.
	}
	return m, srows.Err()
}

// buildUserMap returns source_user_id (as int64) → target_user_id from legacy_user_mappings.
func buildUserMap(ctx context.Context, targetDB *sql.DB, sourceLabel string) (map[int64]int64, error) {
	rows, err := targetDB.QueryContext(ctx, `
		SELECT source_user_id, target_user_id
		FROM legacy_user_mappings
		WHERE source_app_code = $1
		  AND source_db_label = $1
		  AND source_table = 'users'
		  AND mapping_status = 'mapped'
		  AND target_user_id IS NOT NULL
	`, sourceLabel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[int64]int64{}
	for rows.Next() {
		var sourceIDStr string
		var targetID int64
		if err := rows.Scan(&sourceIDStr, &targetID); err != nil {
			return nil, err
		}
		if srcID, err := strconv.ParseInt(sourceIDStr, 10, 64); err == nil {
			m[srcID] = targetID
		}
	}
	return m, rows.Err()
}

// copyTrainingCards copies training_cards from source to target with course_code tag.
// Returns source_tc_id → target_tc_id map for user_cards wiring.
func copyTrainingCards(
	ctx context.Context,
	sourceDB, targetDB *sql.DB,
	courseCode string,
	wcMap map[int64]int64,
	summary *writeSummary,
) (map[int64]int64, error) {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT id, word_card_id, word_en, transcription, sense_index,
		       word_ru, meaning_en, example_en, example_ru,
		       distractors_ru, distractors_en, hint, pos, display_word
		FROM training_cards
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tcMap := map[int64]int64{}
	for rows.Next() {
		var srcID, srcWCID int64
		var wordEN, wordRU, meaningEN string
		var senseIndex int
		var transcription, exampleEN, exampleRU, distractorsRU, distractorsEN, hint, pos, displayWord sql.NullString
		if err := rows.Scan(
			&srcID, &srcWCID, &wordEN, &transcription, &senseIndex,
			&wordRU, &meaningEN, &exampleEN, &exampleRU,
			&distractorsRU, &distractorsEN, &hint, &pos, &displayWord,
		); err != nil {
			return nil, err
		}
		summary.ItemsScanned++

		targetWCID, ok := wcMap[srcWCID]
		if !ok {
			summary.Skipped++
			continue
		}

		// Look up existing training_card for (word_card_id, sense_index, course_code).
		var targetID int64
		lookupErr := targetDB.QueryRowContext(ctx, `
			SELECT id FROM training_cards
			WHERE word_card_id = $1 AND sense_index = $2 AND course_code = $3
			LIMIT 1
		`, targetWCID, senseIndex, courseCode).Scan(&targetID)

		if lookupErr == nil {
			// Already exists.
			tcMap[srcID] = targetID
			summary.MappingsExisting++
			continue
		}
		if lookupErr != sql.ErrNoRows {
			return nil, fmt.Errorf("lookup training_card wc=%d si=%d: %w", targetWCID, senseIndex, lookupErr)
		}

		// Insert new.
		if err := targetDB.QueryRowContext(ctx, `
			INSERT INTO training_cards (
				word_card_id, word_en, transcription, sense_index,
				word_ru, meaning_en, example_en, example_ru,
				distractors_ru, distractors_en, hint, pos, display_word, course_code
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7, $8,
				$9, $10, $11, $12, $13, $14
			)
			RETURNING id
		`,
			targetWCID, wordEN, nullableString(transcription), senseIndex,
			wordRU, meaningEN, nullableString(exampleEN), nullableString(exampleRU),
			nullableString(distractorsRU), nullableString(distractorsEN),
			nullableString(hint), nullableString(pos),
			func() interface{} {
				if displayWord.Valid {
					return displayWord.String
				}
				return wordEN // mirror CreateTrainingCard logic: fall back to word_en
			}(),
			courseCode,
		).Scan(&targetID); err != nil {
			return nil, fmt.Errorf("insert training_card wc=%d si=%d: %w", targetWCID, senseIndex, err)
		}
		tcMap[srcID] = targetID
		summary.ItemsInserted++
	}
	return tcMap, rows.Err()
}

// copyUserCards copies user_cards from source to target with user+training_card remapping.
// Existing cards (same user_id, training_card_id, direction) are updated with source SRS state.
func copyUserCards(
	ctx context.Context,
	sourceDB, targetDB *sql.DB,
	courseCode string,
	userMap, tcMap map[int64]int64,
	summary *writeSummary,
) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT id, user_id, training_card_id, direction, state,
		       ef, reps, interval_days, learning_step, lapse_count,
		       next_due_at, last_review_at, last_quality,
		       last_options_json, wrong_answers_json, stats_json,
		       created_at
		FROM user_cards
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var srcID, srcUserID, srcTCID int64
		var direction, state string
		var ef sql.NullFloat64
		var reps, intervalDays, learningStep, lapseCount sql.NullInt64
		var nextDueAt, lastReviewAt sql.NullTime
		var lastQuality sql.NullInt64
		var lastOptionsJSON, wrongAnswersJSON, statsJSON sql.NullString
		var createdAt sql.NullTime

		if err := rows.Scan(
			&srcID, &srcUserID, &srcTCID, &direction, &state,
			&ef, &reps, &intervalDays, &learningStep, &lapseCount,
			&nextDueAt, &lastReviewAt, &lastQuality,
			&lastOptionsJSON, &wrongAnswersJSON, &statsJSON,
			&createdAt,
		); err != nil {
			return err
		}
		summary.AttemptsScanned++

		targetUserID, ok := userMap[srcUserID]
		if !ok {
			summary.Skipped++
			continue
		}
		targetTCID, ok := tcMap[srcTCID]
		if !ok {
			summary.Skipped++
			continue
		}

		// Upsert: insert or update SRS state.
		_, err := targetDB.ExecContext(ctx, `
			INSERT INTO user_cards (
				user_id, training_card_id, direction, state,
				ef, reps, interval_days, learning_step, lapse_count,
				next_due_at, last_review_at, last_quality,
				last_options_json, wrong_answers_json, stats_json,
				course_code, created_at
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7, $8, $9,
				$10, $11, $12,
				COALESCE($13, '{}'), COALESCE($14, '[]'), COALESCE($15, '{}'),
				$16, COALESCE($17, CURRENT_TIMESTAMP)
			)
			ON CONFLICT (user_id, training_card_id, direction) DO UPDATE SET
				state          = excluded.state,
				ef             = excluded.ef,
				reps           = excluded.reps,
				interval_days  = excluded.interval_days,
				learning_step  = excluded.learning_step,
				lapse_count    = excluded.lapse_count,
				next_due_at    = excluded.next_due_at,
				last_review_at = excluded.last_review_at,
				last_quality   = excluded.last_quality,
				last_options_json  = excluded.last_options_json,
				wrong_answers_json = excluded.wrong_answers_json,
				stats_json     = excluded.stats_json,
				course_code    = excluded.course_code
		`,
			targetUserID, targetTCID, direction, state,
			nullableFloat(ef), nullableInt64(reps), nullableInt64(intervalDays),
			nullableInt64(learningStep), nullableInt64(lapseCount),
			nullableTime(nextDueAt), nullableTime(lastReviewAt), nullableInt64(lastQuality),
			nullableString(lastOptionsJSON), nullableString(wrongAnswersJSON), nullableString(statsJSON),
			courseCode, nullableTime(createdAt),
		)
		if err != nil {
			return fmt.Errorf("upsert user_card user=%d tc=%d dir=%s: %w", targetUserID, targetTCID, direction, err)
		}
		summary.AttemptsInserted++
	}
	return rows.Err()
}

// copyUserWordMastering copies user_word_mastering with user+word_card remapping.
func copyUserWordMastering(
	ctx context.Context,
	sourceDB, targetDB *sql.DB,
	courseCode string,
	userMap, wcMap map[int64]int64,
	summary *writeSummary,
) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT user_id, word_card_id, mastering_score FROM user_word_mastering
	`)
	if err != nil {
		// Table might not exist on older source DBs.
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var srcUserID, srcWCID int64
		var score sql.NullInt64
		if err := rows.Scan(&srcUserID, &srcWCID, &score); err != nil {
			return err
		}
		targetUserID, ok := userMap[srcUserID]
		if !ok {
			continue
		}
		targetWCID, ok := wcMap[srcWCID]
		if !ok {
			continue
		}
		_, err := targetDB.ExecContext(ctx, `
			INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score, course_code)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, word_card_id) DO UPDATE SET
				mastering_score = excluded.mastering_score,
				course_code     = excluded.course_code
		`, targetUserID, targetWCID, nullableInt64(score), courseCode)
		if err != nil {
			return fmt.Errorf("upsert user_word_mastering user=%d wc=%d: %w", targetUserID, targetWCID, err)
		}
		summary.EventsInserted++
	}
	return rows.Err()
}

// copyUserWordKnowledge copies user_word_knowledge with user+word_card remapping.
func copyUserWordKnowledge(
	ctx context.Context,
	sourceDB, targetDB *sql.DB,
	courseCode string,
	userMap, wcMap map[int64]int64,
	summary *writeSummary,
) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT user_id, word_card_id, status FROM user_word_knowledge
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var srcUserID, srcWCID int64
		var status string
		if err := rows.Scan(&srcUserID, &srcWCID, &status); err != nil {
			return err
		}
		targetUserID, ok := userMap[srcUserID]
		if !ok {
			continue
		}
		targetWCID, ok := wcMap[srcWCID]
		if !ok {
			continue
		}
		_, err := targetDB.ExecContext(ctx, `
			INSERT INTO user_word_knowledge (user_id, word_card_id, status, course_code)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, word_card_id) DO UPDATE SET
				status      = excluded.status,
				course_code = excluded.course_code
		`, targetUserID, targetWCID, status, courseCode)
		if err != nil {
			return fmt.Errorf("upsert user_word_knowledge user=%d wc=%d: %w", targetUserID, targetWCID, err)
		}
		summary.EventsInserted++
	}
	return rows.Err()
}
