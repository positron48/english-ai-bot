package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

func mergeSRS(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "srs"}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		maps, err := loadTargetMaps(ctx, targetDB, src.Label)
		if err != nil {
			return nil, fmt.Errorf("load target maps for %s: %w", src.Label, err)
		}
		srsKeyToTargetID, err := copySRSFromSource(ctx, src.DB, src.Label, targetDB, maps, summary)
		if err != nil {
			return nil, fmt.Errorf("copy srs from %s: %w", src.Label, err)
		}
		if err := linkAttemptSRSItems(ctx, src.DB, src.Label, targetDB, maps, srsKeyToTargetID, summary); err != nil {
			return nil, fmt.Errorf("link attempt srs from %s: %w", src.Label, err)
		}
	}
	return summary, nil
}

func srsTargetKey(userCourseID, learningItemID int64) string {
	return strconv.FormatInt(userCourseID, 10) + "\x00" + strconv.FormatInt(learningItemID, 10)
}

func copySRSFromSource(ctx context.Context, sourceDB *sql.DB, sourceLabel string, targetDB *sql.DB, maps *targetMaps, summary *writeSummary) (map[string]int64, error) {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT si.id, uc.user_id, c.code, si.learning_item_id,
		       si.state, si.stability, si.difficulty, si.due_at, si.last_review_at,
		       si.reps, si.lapse_count, si.stats_json::text, si.created_at, si.updated_at
		FROM srs_items si
		JOIN user_courses uc ON uc.id = si.user_course_id
		JOIN courses c ON c.id = uc.course_id
		ORDER BY si.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	srsKeyToTargetID := map[string]int64{}
	for rows.Next() {
		var sourceID, userID, sourceLearningItemID int64
		var courseCode, state, statsJSON string
		var stability, difficulty sql.NullFloat64
		var dueAt, lastReviewAt, createdAt, updatedAt sql.NullTime
		var reps, lapseCount int
		if err := rows.Scan(
			&sourceID, &userID, &courseCode, &sourceLearningItemID,
			&state, &stability, &difficulty, &dueAt, &lastReviewAt,
			&reps, &lapseCount, &statsJSON, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		summary.SRSScanned++

		targetUC, ok := maps.userCourseByUserAndCourse[userCourseKey(strconv.FormatInt(userID, 10), courseCode)]
		if !ok {
			summary.Skipped++
			continue
		}
		targetLearningItem, ok := maps.learningItemBySourcePK[strconv.FormatInt(sourceLearningItemID, 10)]
		if !ok {
			summary.Skipped++
			continue
		}

		var targetID int64
		err := targetDB.QueryRowContext(ctx, `
			INSERT INTO srs_items (
				user_course_id, learning_item_id, state, stability, difficulty,
				due_at, last_review_at, reps, lapse_count, stats_json, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10::jsonb, COALESCE($11, CURRENT_TIMESTAMP), COALESCE($12, CURRENT_TIMESTAMP)
			)
			ON CONFLICT (user_course_id, learning_item_id) DO UPDATE SET
				state = excluded.state,
				stability = excluded.stability,
				difficulty = excluded.difficulty,
				due_at = excluded.due_at,
				last_review_at = excluded.last_review_at,
				reps = excluded.reps,
				lapse_count = excluded.lapse_count,
				stats_json = excluded.stats_json,
				updated_at = excluded.updated_at
			RETURNING id
		`,
			targetUC, targetLearningItem, state, nullableFloat(stability), nullableFloat(difficulty),
			nullableTime(dueAt), nullableTime(lastReviewAt), reps, lapseCount, statsJSON, nullableTime(createdAt), nullableTime(updatedAt),
		).Scan(&targetID)
		if err != nil {
			return nil, fmt.Errorf("upsert srs source_id=%d: %w", sourceID, err)
		}
		summary.SRSInserted++
		srsKeyToTargetID[srsTargetKey(targetUC, targetLearningItem)] = targetID
	}
	return srsKeyToTargetID, rows.Err()
}

func linkAttemptSRSItems(ctx context.Context, sourceDB *sql.DB, sourceLabel string, targetDB *sql.DB, maps *targetMaps, srsKeyToTargetID map[string]int64, summary *writeSummary) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT ea.id, ea.srs_item_id, uc.user_id, c.code, si.learning_item_id
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		JOIN srs_items si ON si.id = ea.srs_item_id
		WHERE ea.srs_item_id IS NOT NULL
		ORDER BY ea.id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sourceAttemptID, sourceSRSID, userID, learningItemID int64
		var courseCode string
		if err := rows.Scan(&sourceAttemptID, &sourceSRSID, &userID, &courseCode, &learningItemID); err != nil {
			return err
		}

		var targetAttemptID sql.NullInt64
		err := targetDB.QueryRowContext(ctx, `
			SELECT target_exercise_attempt_id
			FROM legacy_attempt_mappings
			WHERE source_app_code = $1
			  AND source_db_label = $1
			  AND source_table = 'exercise_attempts'
			  AND source_pk = $2
		`, sourceLabel, strconv.FormatInt(sourceAttemptID, 10)).Scan(&targetAttemptID)
		if err != nil || !targetAttemptID.Valid {
			continue
		}

		targetUC, ok := maps.userCourseByUserAndCourse[userCourseKey(strconv.FormatInt(userID, 10), courseCode)]
		if !ok {
			continue
		}
		targetLearningItem, ok := maps.learningItemBySourcePK[strconv.FormatInt(learningItemID, 10)]
		if !ok {
			continue
		}
		targetSRSID, ok := srsKeyToTargetID[srsTargetKey(targetUC, targetLearningItem)]
		if !ok {
			continue
		}

		res, err := targetDB.ExecContext(ctx, `
			UPDATE exercise_attempts
			SET srs_item_id = $1
			WHERE id = $2 AND (srs_item_id IS NULL OR srs_item_id = $1)
		`, targetSRSID, targetAttemptID.Int64)
		if err != nil {
			return err
		}
		if affected, _ := res.RowsAffected(); affected > 0 {
			summary.SRSLinksUpdated++
		}
	}
	return rows.Err()
}

func nullableFloat(v sql.NullFloat64) interface{} {
	if v.Valid {
		return v.Float64
	}
	return nil
}
