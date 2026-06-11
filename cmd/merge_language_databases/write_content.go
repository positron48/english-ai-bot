package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

func mergeCourseMappings(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "course-mappings"}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		courseCode, ok := sourceCourseCodes[src.Label]
		if !ok {
			summary.Skipped++
			continue
		}
		var sourceCourseID int64
		if err := src.DB.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = $1`, courseCode).Scan(&sourceCourseID); err != nil {
			return nil, fmt.Errorf("source course %s in %s: %w", courseCode, src.Label, err)
		}
		var targetCourseID int64
		if err := targetDB.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = $1`, courseCode).Scan(&targetCourseID); err != nil {
			return nil, fmt.Errorf("target course %s: %w", courseCode, err)
		}
		res, err := targetDB.ExecContext(ctx, `
			INSERT INTO legacy_course_mappings (
				source_app_code, source_db_label, source_course_code, source_course_id,
				target_course_id, mapping_status, metadata_json, created_at, updated_at
			) VALUES ($1, $1, $2, $3, $4, 'mapped', '{}'::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (source_app_code, source_db_label, source_course_code) DO UPDATE SET
				source_course_id = excluded.source_course_id,
				target_course_id = excluded.target_course_id,
				mapping_status = 'mapped',
				updated_at = CURRENT_TIMESTAMP
		`, src.Label, courseCode, strconv.FormatInt(sourceCourseID, 10), targetCourseID)
		if err != nil {
			return nil, err
		}
		if affected, _ := res.RowsAffected(); affected > 0 {
			summary.MappingsCreated++
		} else {
			summary.MappingsExisting++
		}
	}
	return summary, nil
}

func mergeContent(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "content"}
	sharedWords, err := findSharedSourceWords(ctx, sources)
	if err != nil {
		return nil, fmt.Errorf("find shared source words: %w", err)
	}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		courseCode, ok := sourceCourseCodes[src.Label]
		if !ok {
			summary.Skipped++
			continue
		}
		var sourceCourseID, targetCourseID int64
		if err := src.DB.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = $1`, courseCode).Scan(&sourceCourseID); err != nil {
			return nil, fmt.Errorf("source course %s in %s: %w", courseCode, src.Label, err)
		}
		if err := targetDB.QueryRowContext(ctx, `SELECT id FROM courses WHERE code = $1`, courseCode).Scan(&targetCourseID); err != nil {
			return nil, fmt.Errorf("target course %s: %w", courseCode, err)
		}
		// Import source word_cards into unified and build source_wc_id → unified_wc_id map.
		// This ensures that when we copy learning_items with source_kind='word_card', we
		// remap their source_id to the correct unified word_card ID (not the source DB's ID,
		// which may differ when both DBs have the same sequence position but different content).
		wcMap, err := importWordCardsFromSource(ctx, src.DB, targetDB, courseCode, sharedWords, summary)
		if err != nil {
			return nil, fmt.Errorf("import word_cards from %s: %w", src.Label, err)
		}
		if err := copyWordSets(ctx, src.DB, targetDB, courseCode, wcMap, summary); err != nil {
			return nil, fmt.Errorf("copy word sets from %s: %w", src.Label, err)
		}

		moduleMap, err := copyModules(ctx, src.DB, targetDB, sourceCourseID, targetCourseID)
		if err != nil {
			return nil, fmt.Errorf("copy modules %s: %w", src.Label, err)
		}
		if err := copyLearningItems(ctx, src.DB, targetDB, src.Label, sourceCourseID, targetCourseID, moduleMap, wcMap, summary); err != nil {
			return nil, fmt.Errorf("copy learning items %s: %w", src.Label, err)
		}
	}
	return summary, nil
}

func findSharedSourceWords(ctx context.Context, sources []openedSourceDB) (map[string]bool, error) {
	counts := map[string]int{}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		rows, err := src.DB.QueryContext(ctx, `SELECT LOWER(word) FROM word_cards`)
		if err != nil {
			return nil, err
		}
		seen := map[string]struct{}{}
		for rows.Next() {
			var word string
			if err := rows.Scan(&word); err != nil {
				rows.Close()
				return nil, err
			}
			if _, ok := seen[word]; !ok {
				seen[word] = struct{}{}
				counts[word]++
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	shared := map[string]bool{}
	for word, count := range counts {
		if count > 1 {
			shared[word] = true
		}
	}
	return shared, nil
}

func copyModules(ctx context.Context, sourceDB, targetDB *sql.DB, sourceCourseID, targetCourseID int64) (map[int64]int64, error) {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT id, code, module_type, title, source_kind, source_id, sort_order, status, metadata_json
		FROM modules
		WHERE course_id = $1
		ORDER BY id
	`, sourceCourseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	moduleMap := map[int64]int64{}
	for rows.Next() {
		var sourceID int64
		var code, moduleType, title, status string
		var sourceKind, sourceIDText sql.NullString
		var sortOrder int
		var metadataJSON []byte
		if err := rows.Scan(&sourceID, &code, &moduleType, &title, &sourceKind, &sourceIDText, &sortOrder, &status, &metadataJSON); err != nil {
			return nil, err
		}
		var targetID int64
		err := targetDB.QueryRowContext(ctx, `
			INSERT INTO modules (
				course_id, code, module_type, title, source_kind, source_id,
				sort_order, status, metadata_json, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (course_id, code) DO UPDATE SET
				module_type = excluded.module_type,
				title = excluded.title,
				source_kind = excluded.source_kind,
				source_id = excluded.source_id,
				sort_order = excluded.sort_order,
				status = excluded.status,
				metadata_json = excluded.metadata_json,
				updated_at = CURRENT_TIMESTAMP
			RETURNING id
		`, targetCourseID, code, moduleType, title, nullableString(sourceKind), nullableString(sourceIDText), sortOrder, status, string(metadataJSON)).Scan(&targetID)
		if err != nil {
			return nil, err
		}
		moduleMap[sourceID] = targetID
	}
	return moduleMap, rows.Err()
}

func copyLearningItems(
	ctx context.Context,
	sourceDB, targetDB *sql.DB,
	sourceLabel string,
	sourceCourseID, targetCourseID int64,
	moduleMap map[int64]int64,
	wcMap map[int64]int64, // source word_card_id → unified word_card_id; nil = no remapping
	summary *writeSummary,
) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT id, module_id, item_type, source_kind, source_id, title, cefr_level,
		       content_hash, payload_json, status
		FROM learning_items
		WHERE course_id = $1
		ORDER BY id
	`, sourceCourseID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sourceItemID int64
		var moduleID sql.NullInt64
		var itemType, sourceKind, sourceID, status string
		var title, cefrLevel, contentHash sql.NullString
		var payloadJSON []byte
		if err := rows.Scan(
			&sourceItemID, &moduleID, &itemType, &sourceKind, &sourceID,
			&title, &cefrLevel, &contentHash, &payloadJSON, &status,
		); err != nil {
			return err
		}
		summary.ItemsScanned++

		// For word_card learning_items, remap source_id from the source DB's word_card ID
		// to the unified DB's word_card ID (found by word text in importWordCardsFromSource).
		// Without this, both en_ru and es_ru would end up referencing the same ID-space which
		// maps to whichever content was seeded first — producing wrong word_card references.
		if sourceKind == "word_card" && len(wcMap) > 0 {
			if srcWCID, ok := wordCardIDFromString(sourceID); ok {
				if targetWCID, mapped := wcMap[srcWCID]; mapped {
					sourceID = strconv.FormatInt(targetWCID, 10)
				} else {
					summary.Skipped++
					continue // word_card not in unified — skip this item
				}
			}
		}

		var targetModuleID sql.NullInt64
		if moduleID.Valid {
			if mapped, ok := moduleMap[moduleID.Int64]; ok {
				targetModuleID = sql.NullInt64{Int64: mapped, Valid: true}
			}
		}

		var targetItemID int64
		err := targetDB.QueryRowContext(ctx, `
			INSERT INTO learning_items (
				course_id, module_id, item_type, source_kind, source_id, title, cefr_level,
				content_hash, payload_json, status, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			)
			ON CONFLICT (course_id, source_kind, source_id) DO UPDATE SET
				module_id = excluded.module_id,
				item_type = excluded.item_type,
				title = excluded.title,
				cefr_level = excluded.cefr_level,
				content_hash = excluded.content_hash,
				payload_json = excluded.payload_json,
				status = excluded.status,
				updated_at = CURRENT_TIMESTAMP
			RETURNING id
		`, targetCourseID, nullableInt64(targetModuleID), itemType, sourceKind, sourceID,
			nullableString(title), nullableString(cefrLevel), nullableString(contentHash), string(payloadJSON), status,
		).Scan(&targetItemID)
		if err != nil {
			return err
		}

		res, err := targetDB.ExecContext(ctx, `
			INSERT INTO legacy_content_mappings (
				source_app_code, source_db_label, source_table, source_pk,
				source_kind, source_id, target_course_id, target_learning_item_id,
				mapping_status, content_hash, metadata_json, created_at, updated_at
			) VALUES (
				$1, $1, 'learning_items', $2,
				$3, $4, $5, $6,
				'mapped', $7, '{}'::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			)
			ON CONFLICT (source_app_code, source_db_label, source_table, source_pk) DO UPDATE SET
				target_course_id = excluded.target_course_id,
				target_learning_item_id = excluded.target_learning_item_id,
				mapping_status = 'mapped',
				content_hash = excluded.content_hash,
				updated_at = CURRENT_TIMESTAMP
		`, sourceLabel, strconv.FormatInt(sourceItemID, 10), sourceKind, sourceID, targetCourseID, targetItemID, nullableString(contentHash))
		if err != nil {
			return err
		}
		if affected, _ := res.RowsAffected(); affected > 0 {
			summary.MappingsCreated++
		} else {
			summary.MappingsExisting++
		}
		summary.ItemsInserted++
	}
	return rows.Err()
}

func nullableString(v sql.NullString) interface{} {
	if v.Valid {
		return v.String
	}
	return nil
}
