package main

import (
	"context"
	"database/sql"
	"fmt"
)

// mergeGrammarProgress copies grammar_progress, grammar_test_attempts, and
// grammar_placement_test from each source prod DB into the unified target.
//
// This phase must run after --phase=users (so legacy_user_mappings exist).
// Idempotent: grammar_progress uses ON CONFLICT DO UPDATE (best score wins),
// grammar_test_attempts deduplicates by client_attempt_id when available,
// otherwise by (user_id, scope_type, scope_id, started_at).
func mergeGrammarProgress(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "grammar-progress"}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		userMap, err := buildUserMap(ctx, targetDB, src.Label)
		if err != nil {
			return nil, fmt.Errorf("build user map for %s: %w", src.Label, err)
		}

		if err := copyGrammarProgress(ctx, src.DB, targetDB, userMap, summary); err != nil {
			return nil, fmt.Errorf("copy grammar_progress from %s: %w", src.Label, err)
		}
		if err := copyGrammarTestAttempts(ctx, src.DB, targetDB, userMap, summary); err != nil {
			return nil, fmt.Errorf("copy grammar_test_attempts from %s: %w", src.Label, err)
		}
		if err := copyGrammarPlacementTest(ctx, src.DB, targetDB, userMap, summary); err != nil {
			return nil, fmt.Errorf("copy grammar_placement_test from %s: %w", src.Label, err)
		}
	}
	return summary, nil
}

func copyGrammarProgress(ctx context.Context, sourceDB, targetDB *sql.DB, userMap map[int64]int64, summary *writeSummary) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT user_id, chapter_id, best_score, passed_at, last_attempt_at
		FROM grammar_progress
	`)
	if err != nil {
		return nil // table may not exist on older source
	}
	defer rows.Close()

	for rows.Next() {
		var srcUserID int64
		var chapterID string
		var bestScore int
		var passedAt, lastAttemptAt sql.NullTime
		if err := rows.Scan(&srcUserID, &chapterID, &bestScore, &passedAt, &lastAttemptAt); err != nil {
			return err
		}
		targetUserID, ok := userMap[srcUserID]
		if !ok {
			summary.Skipped++
			continue
		}
		res, err := targetDB.ExecContext(ctx, `
			INSERT INTO grammar_progress (user_id, chapter_id, best_score, passed_at, last_attempt_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id, chapter_id) DO UPDATE SET
				best_score     = GREATEST(grammar_progress.best_score, excluded.best_score),
				passed_at      = COALESCE(grammar_progress.passed_at, excluded.passed_at),
				last_attempt_at = GREATEST(grammar_progress.last_attempt_at, excluded.last_attempt_at)
		`, targetUserID, chapterID, bestScore, nullableTime(passedAt), nullableTime(lastAttemptAt))
		if err != nil {
			return fmt.Errorf("upsert grammar_progress user=%d chapter=%s: %w", targetUserID, chapterID, err)
		}
		if affected, _ := res.RowsAffected(); affected > 0 {
			summary.ItemsInserted++
		} else {
			summary.MappingsExisting++
		}
	}
	return rows.Err()
}

func copyGrammarTestAttempts(ctx context.Context, sourceDB, targetDB *sql.DB, userMap map[int64]int64, summary *writeSummary) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT user_id, scope_type, scope_id, started_at, finished_at,
		       score, passed, total_questions, answers_json, results_json,
		       course_version, client_attempt_id
		FROM grammar_test_attempts
		ORDER BY id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var srcUserID int64
		var scopeType, scopeID string
		var score, passed, totalQ int
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		var answersJSON, resultsJSON, courseVersion, clientAttemptID sql.NullString
		if err := rows.Scan(
			&srcUserID, &scopeType, &scopeID, &startedAt, &finishedAt,
			&score, &passed, &totalQ, &answersJSON, &resultsJSON,
			&courseVersion, &clientAttemptID,
		); err != nil {
			return err
		}
		summary.AttemptsScanned++
		targetUserID, ok := userMap[srcUserID]
		if !ok {
			summary.Skipped++
			continue
		}

		// Dedup by client_attempt_id if present, otherwise by (user_id, scope_type, scope_id, started_at).
		if clientAttemptID.Valid && clientAttemptID.String != "" {
			_, err = targetDB.ExecContext(ctx, `
				INSERT INTO grammar_test_attempts (
					user_id, scope_type, scope_id, started_at, finished_at,
					score, passed, total_questions, answers_json, results_json,
					course_version, client_attempt_id
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
				ON CONFLICT (user_id, client_attempt_id) WHERE client_attempt_id IS NOT NULL DO NOTHING
			`, targetUserID, scopeType, scopeID, nullableTime(startedAt), nullableTime(finishedAt),
				score, passed, totalQ,
				nullableString(answersJSON), nullableString(resultsJSON),
				nullableString(courseVersion), clientAttemptID.String)
		} else {
			// No client_attempt_id — deduplicate on natural key.
			_, err = targetDB.ExecContext(ctx, `
				INSERT INTO grammar_test_attempts (
					user_id, scope_type, scope_id, started_at, finished_at,
					score, passed, total_questions, answers_json, results_json,
					course_version
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
				ON CONFLICT DO NOTHING
			`, targetUserID, scopeType, scopeID, nullableTime(startedAt), nullableTime(finishedAt),
				score, passed, totalQ,
				nullableString(answersJSON), nullableString(resultsJSON),
				nullableString(courseVersion))
		}
		if err != nil {
			return fmt.Errorf("insert grammar_test_attempt user=%d scope=%s/%s: %w", targetUserID, scopeType, scopeID, err)
		}
		summary.AttemptsInserted++
	}
	return rows.Err()
}

func copyGrammarPlacementTest(ctx context.Context, sourceDB, targetDB *sql.DB, userMap map[int64]int64, summary *writeSummary) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT user_id, score, total_questions, opened_sections_json, completed_at
		FROM grammar_placement_test
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var srcUserID int64
		var score, totalQ int
		var openedSections string
		var completedAt sql.NullTime
		if err := rows.Scan(&srcUserID, &score, &totalQ, &openedSections, &completedAt); err != nil {
			return err
		}
		targetUserID, ok := userMap[srcUserID]
		if !ok {
			summary.Skipped++
			continue
		}
		_, err := targetDB.ExecContext(ctx, `
			INSERT INTO grammar_placement_test (user_id, score, total_questions, opened_sections_json, completed_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id) DO UPDATE SET
				score                = GREATEST(grammar_placement_test.score, excluded.score),
				total_questions      = excluded.total_questions,
				opened_sections_json = excluded.opened_sections_json,
				completed_at         = COALESCE(grammar_placement_test.completed_at, excluded.completed_at)
		`, targetUserID, score, totalQ, openedSections, nullableTime(completedAt))
		if err != nil {
			return fmt.Errorf("upsert grammar_placement_test user=%d: %w", targetUserID, err)
		}
		summary.EventsInserted++
	}
	return rows.Err()
}
