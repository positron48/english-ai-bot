package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

const mergeAttemptsBatchSize = 500

type targetMaps struct {
	userCourseByUserAndCourse map[string]int64
	learningItemBySourcePK    map[string]int64
}

func mergeAttempts(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "attempts"}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		maps, err := loadTargetMaps(ctx, targetDB, src.Label)
		if err != nil {
			return nil, fmt.Errorf("load target maps for %s: %w", src.Label, err)
		}
		if err := copyAttemptsFromSource(ctx, src.DB, src.Label, targetDB, maps, summary); err != nil {
			return nil, fmt.Errorf("copy attempts from %s: %w", src.Label, err)
		}
	}
	return summary, nil
}

func loadTargetMaps(ctx context.Context, targetDB *sql.DB, sourceLabel string) (*targetMaps, error) {
	m := &targetMaps{
		userCourseByUserAndCourse: map[string]int64{},
		learningItemBySourcePK:    map[string]int64{},
	}
	rows, err := targetDB.QueryContext(ctx, `
		SELECT lum.source_user_id, lcm.source_course_code, tuc.id
		FROM legacy_user_mappings lum
		JOIN legacy_course_mappings lcm
		  ON lcm.source_app_code = lum.source_app_code
		 AND lcm.source_db_label = lum.source_db_label
		JOIN user_courses tuc
		  ON tuc.user_id = lum.target_user_id
		 AND tuc.course_id = lcm.target_course_id
		WHERE lum.source_app_code = $1
		  AND lum.source_db_label = $1
		  AND lum.mapping_status = 'mapped'
		  AND lum.target_user_id IS NOT NULL
	`, sourceLabel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sourceUserID, courseCode string
		var targetUC int64
		if err := rows.Scan(&sourceUserID, &courseCode, &targetUC); err != nil {
			return nil, err
		}
		m.userCourseByUserAndCourse[userCourseKey(sourceUserID, courseCode)] = targetUC
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = targetDB.QueryContext(ctx, `
		SELECT source_pk, target_learning_item_id
		FROM legacy_content_mappings
		WHERE source_app_code = $1
		  AND source_db_label = $1
		  AND source_table = 'learning_items'
		  AND mapping_status = 'mapped'
		  AND target_learning_item_id IS NOT NULL
	`, sourceLabel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sourcePK string
		var targetItemID int64
		if err := rows.Scan(&sourcePK, &targetItemID); err != nil {
			return nil, err
		}
		m.learningItemBySourcePK[sourcePK] = targetItemID
	}
	return m, rows.Err()
}

func userCourseKey(sourceUserID, courseCode string) string {
	return sourceUserID + "\x00" + courseCode
}

func copyAttemptsFromSource(ctx context.Context, sourceDB *sql.DB, sourceLabel string, targetDB *sql.DB, maps *targetMaps, summary *writeSummary) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT ea.id, ea.user_course_id, uc.user_id, c.code,
		       ea.learning_item_id, ea.mode, ea.client_attempt_id,
		       ea.started_at, ea.answered_at, ea.is_correct, ea.score, ea.quality,
		       ea.prompt_json::text, ea.answer_json::text, ea.result_json::text,
		       ea.source_table, ea.source_pk, ea.created_at
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		ORDER BY ea.id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	batch := make([]sourceAttemptRow, 0, mergeAttemptsBatchSize)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := insertAttemptBatch(ctx, sourceDB, sourceLabel, targetDB, maps, batch, summary); err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}

	for rows.Next() {
		var row sourceAttemptRow
		var learningItemID sql.NullInt64
		var clientAttemptID, sourceTable, sourcePK sql.NullString
		var answeredAt sql.NullTime
		var isCorrect sql.NullBool
		var score, quality sql.NullInt64
		if err := rows.Scan(
			&row.ID, &row.UserCourseID, &row.UserID, &row.CourseCode,
			&learningItemID, &row.Mode, &clientAttemptID,
			&row.StartedAt, &answeredAt, &isCorrect, &score, &quality,
			&row.PromptJSON, &row.AnswerJSON, &row.ResultJSON,
			&sourceTable, &sourcePK, &row.CreatedAt,
		); err != nil {
			return err
		}
		row.LearningItemID = learningItemID
		row.ClientAttemptID = clientAttemptID
		row.AnsweredAt = answeredAt
		row.IsCorrect = isCorrect
		row.Score = score
		row.Quality = quality
		row.SourceTable = sourceTable
		row.SourcePK = sourcePK
		summary.AttemptsScanned++
		batch = append(batch, row)
		if len(batch) >= mergeAttemptsBatchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return flush()
}

type sourceAttemptRow struct {
	ID              int64
	UserCourseID    int64
	UserID          int64
	CourseCode      string
	LearningItemID  sql.NullInt64
	Mode            string
	ClientAttemptID sql.NullString
	StartedAt       sql.NullTime
	AnsweredAt      sql.NullTime
	IsCorrect       sql.NullBool
	Score           sql.NullInt64
	Quality         sql.NullInt64
	PromptJSON      string
	AnswerJSON      string
	ResultJSON      string
	SourceTable     sql.NullString
	SourcePK        sql.NullString
	CreatedAt       sql.NullTime
}

func insertAttemptBatch(ctx context.Context, sourceDB *sql.DB, sourceLabel string, targetDB *sql.DB, maps *targetMaps, batch []sourceAttemptRow, summary *writeSummary) error {
	tx, err := targetDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, row := range batch {
		sourcePK := strconv.FormatInt(row.ID, 10)
		var existingTarget sql.NullInt64
		err := tx.QueryRowContext(ctx, `
			SELECT target_exercise_attempt_id
			FROM legacy_attempt_mappings
			WHERE source_app_code = $1
			  AND source_db_label = $1
			  AND source_table = 'exercise_attempts'
			  AND source_pk = $2
		`, sourceLabel, sourcePK).Scan(&existingTarget)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if err == nil && existingTarget.Valid {
			summary.MappingsExisting++
			continue
		}

		targetUC, ok := maps.userCourseByUserAndCourse[userCourseKey(strconv.FormatInt(row.UserID, 10), row.CourseCode)]
		if !ok {
			summary.Skipped++
			continue
		}
		var targetLearningItem interface{}
		if row.LearningItemID.Valid {
			if mapped, ok := maps.learningItemBySourcePK[strconv.FormatInt(row.LearningItemID.Int64, 10)]; ok {
				targetLearningItem = mapped
			}
		}

		var targetAttemptID int64
		err = tx.QueryRowContext(ctx, `
			INSERT INTO exercise_attempts (
				user_course_id, learning_item_id, srs_item_id, mode, client_attempt_id,
				started_at, answered_at, is_correct, score, quality,
				prompt_json, answer_json, result_json, source_table, source_pk, created_at
			) VALUES (
				$1, $2, NULL, $3, $4,
				$5, $6, $7, $8, $9,
				$10::jsonb, $11::jsonb, $12::jsonb, $13, $14, COALESCE($15, CURRENT_TIMESTAMP)
			)
			RETURNING id
		`,
			targetUC, targetLearningItem, row.Mode, nullableString(row.ClientAttemptID),
			nullableTime(row.StartedAt), nullableTime(row.AnsweredAt), nullableBool(row.IsCorrect), nullableInt64(row.Score), nullableInt64(row.Quality),
			row.PromptJSON, row.AnswerJSON, row.ResultJSON, nullableString(row.SourceTable), nullableString(row.SourcePK), nullableTime(row.CreatedAt),
		).Scan(&targetAttemptID)
		if err != nil {
			if row.ClientAttemptID.Valid {
				if lookupErr := tx.QueryRowContext(ctx, `
					SELECT id FROM exercise_attempts
					WHERE user_course_id = $1 AND client_attempt_id = $2
				`, targetUC, row.ClientAttemptID.String).Scan(&targetAttemptID); lookupErr != nil {
					return fmt.Errorf("insert attempt id=%d: %w", row.ID, err)
				}
				summary.MappingsExisting++
			} else {
				return fmt.Errorf("insert attempt id=%d: %w", row.ID, err)
			}
		} else {
			summary.AttemptsInserted++
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO legacy_attempt_mappings (
				source_app_code, source_db_label, source_table, source_pk,
				source_user_id, source_course_code, target_user_course_id,
				target_exercise_attempt_id, mapping_status, metadata_json, created_at, updated_at
			) VALUES (
				$1, $1, 'exercise_attempts', $2,
				$3, $4, $5,
				$6, 'mapped', '{}'::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			)
			ON CONFLICT (source_app_code, source_db_label, source_table, source_pk) DO UPDATE SET
				target_user_course_id = excluded.target_user_course_id,
				target_exercise_attempt_id = excluded.target_exercise_attempt_id,
				mapping_status = 'mapped',
				updated_at = CURRENT_TIMESTAMP
		`, sourceLabel, sourcePK, strconv.FormatInt(row.UserID, 10), row.CourseCode, targetUC, targetAttemptID)
		if err != nil {
			return err
		}
		summary.MappingsCreated++

		n, err := copyLearningEventsForAttempt(ctx, sourceDB, tx, sourceLabel, row.ID, targetUC, targetLearningItem, targetAttemptID)
		if err != nil {
			return err
		}
		summary.EventsInserted += n
	}
	return tx.Commit()
}

func copyLearningEventsForAttempt(
	ctx context.Context,
	sourceDB *sql.DB,
	tx *sql.Tx,
	sourceLabel string,
	sourceAttemptID, targetUC int64,
	targetLearningItem interface{},
	targetAttemptID int64,
) (int64, error) {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT id, learning_item_id, event_type, event_time, mode,
		       source_table, source_pk, event_json::text, created_at
		FROM learning_events
		WHERE exercise_attempt_id = $1
		ORDER BY id
	`, sourceAttemptID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var inserted int64
	for rows.Next() {
		var sourceEventID int64
		var learningItemID sql.NullInt64
		var eventType, mode string
		var sourceTable, sourcePK sql.NullString
		var eventJSON string
		var eventTime, createdAt sql.NullTime
		if err := rows.Scan(&sourceEventID, &learningItemID, &eventType, &eventTime, &mode, &sourceTable, &sourcePK, &eventJSON, &createdAt); err != nil {
			return inserted, err
		}

		var existing sql.NullInt64
		err := tx.QueryRowContext(ctx, `
			SELECT target_learning_event_id
			FROM legacy_attempt_mappings
			WHERE source_app_code = $1
			  AND source_db_label = $1
			  AND source_table = 'learning_events'
			  AND source_pk = $2
		`, sourceLabel, strconv.FormatInt(sourceEventID, 10)).Scan(&existing)
		if err != nil && err != sql.ErrNoRows {
			return inserted, err
		}
		if err == nil && existing.Valid {
			continue
		}

		itemID := targetLearningItem
		var targetEventID int64
		err = tx.QueryRowContext(ctx, `
			INSERT INTO learning_events (
				user_course_id, learning_item_id, exercise_attempt_id,
				event_type, event_time, mode, source_table, source_pk, event_json, created_at
			) VALUES (
				$1, $2, $3,
				$4, $5, $6, $7, $8, $9::jsonb, COALESCE($10, CURRENT_TIMESTAMP)
			)
			RETURNING id
		`,
			targetUC, itemID, targetAttemptID,
			eventType, nullableTime(eventTime), mode, nullableString(sourceTable), nullableString(sourcePK), eventJSON, nullableTime(createdAt),
		).Scan(&targetEventID)
		if err != nil {
			return inserted, err
		}
		inserted++

		_, err = tx.ExecContext(ctx, `
			INSERT INTO legacy_attempt_mappings (
				source_app_code, source_db_label, source_table, source_pk,
				target_user_course_id, target_learning_event_id,
				mapping_status, metadata_json, created_at, updated_at
			) VALUES (
				$1, $1, 'learning_events', $2,
				$3, $4,
				'mapped', '{}'::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			)
			ON CONFLICT (source_app_code, source_db_label, source_table, source_pk) DO UPDATE SET
				target_learning_event_id = excluded.target_learning_event_id,
				mapping_status = 'mapped',
				updated_at = CURRENT_TIMESTAMP
		`, sourceLabel, strconv.FormatInt(sourceEventID, 10), targetUC, targetEventID)
		if err != nil {
			return inserted, err
		}
	}
	return inserted, rows.Err()
}

func nullableTime(v sql.NullTime) interface{} {
	if v.Valid {
		return v.Time
	}
	return nil
}

func nullableBool(v sql.NullBool) interface{} {
	if v.Valid {
		return v.Bool
	}
	return nil
}
