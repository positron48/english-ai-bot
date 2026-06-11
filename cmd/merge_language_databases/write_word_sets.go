package main

import (
	"context"
	"database/sql"
	"fmt"
)

type sourceWordSetCategory struct {
	ID          int64
	ParentID    sql.NullInt64
	Name        string
	Description sql.NullString
	IsPublished int64
	SortOrder   int
}

func copyWordSets(
	ctx context.Context,
	sourceDB, targetDB *sql.DB,
	courseCode string,
	wordCardMap map[int64]int64,
	summary *writeSummary,
) error {
	categoryMap, err := copyWordSetCategories(ctx, sourceDB, targetDB, courseCode, summary)
	if err != nil {
		return err
	}

	rows, err := sourceDB.QueryContext(ctx, `
		SELECT id, category_id, title, description, is_published, sort_order, preferred_pos
		FROM word_sets
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sourceID int64
		var sourceCategoryID sql.NullInt64
		var title string
		var description, preferredPOS sql.NullString
		var isPublished int64
		var sortOrder int
		if err := rows.Scan(
			&sourceID, &sourceCategoryID, &title, &description,
			&isPublished, &sortOrder, &preferredPOS,
		); err != nil {
			return err
		}

		var targetCategoryID sql.NullInt64
		if sourceCategoryID.Valid {
			mapped, ok := categoryMap[sourceCategoryID.Int64]
			if !ok {
				return fmt.Errorf("word set %d category %d was not mapped", sourceID, sourceCategoryID.Int64)
			}
			targetCategoryID = sql.NullInt64{Int64: mapped, Valid: true}
		}

		var targetID int64
		err := targetDB.QueryRowContext(ctx, `
			SELECT id
			FROM word_sets
			WHERE course_code = $1
			  AND (($2::bigint IS NULL AND category_id IS NULL) OR category_id = $2)
			  AND title = $3
			ORDER BY id
			LIMIT 1
		`, courseCode, nullableInt64(targetCategoryID), title).Scan(&targetID)
		if err == sql.ErrNoRows {
			err = targetDB.QueryRowContext(ctx, `
				INSERT INTO word_sets (
					course_code, category_id, title, description, is_published,
					sort_order, preferred_pos, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				RETURNING id
			`, courseCode, nullableInt64(targetCategoryID), title, nullableString(description),
				isPublished, sortOrder, nullableString(preferredPOS)).Scan(&targetID)
		}
		if err != nil {
			return err
		}
		if _, err := targetDB.ExecContext(ctx, `
			UPDATE word_sets
			SET description = $2,
			    is_published = $3,
			    sort_order = $4,
			    preferred_pos = $5,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`, targetID, nullableString(description), isPublished, sortOrder, nullableString(preferredPOS)); err != nil {
			return err
		}
		if err := copyWordSetItems(ctx, sourceDB, targetDB, sourceID, targetID, wordCardMap, summary); err != nil {
			return err
		}
	}
	return rows.Err()
}

func copyWordSetCategories(
	ctx context.Context,
	sourceDB, targetDB *sql.DB,
	courseCode string,
	summary *writeSummary,
) (map[int64]int64, error) {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT id, parent_id, name, description, is_published, sort_order
		FROM word_set_categories
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := map[int64]sourceWordSetCategory{}
	for rows.Next() {
		var category sourceWordSetCategory
		if err := rows.Scan(
			&category.ID, &category.ParentID, &category.Name, &category.Description,
			&category.IsPublished, &category.SortOrder,
		); err != nil {
			return nil, err
		}
		categories[category.ID] = category
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	mapped := map[int64]int64{}
	visiting := map[int64]bool{}
	var ensureCategory func(int64) (int64, error)
	ensureCategory = func(sourceID int64) (int64, error) {
		if targetID, ok := mapped[sourceID]; ok {
			return targetID, nil
		}
		category, ok := categories[sourceID]
		if !ok {
			return 0, fmt.Errorf("word set category %d not found", sourceID)
		}
		if visiting[sourceID] {
			return 0, fmt.Errorf("word set category cycle at %d", sourceID)
		}
		visiting[sourceID] = true
		defer delete(visiting, sourceID)

		var targetParentID sql.NullInt64
		if category.ParentID.Valid {
			parentID, err := ensureCategory(category.ParentID.Int64)
			if err != nil {
				return 0, err
			}
			targetParentID = sql.NullInt64{Int64: parentID, Valid: true}
		}

		var targetID int64
		err := targetDB.QueryRowContext(ctx, `
			SELECT id
			FROM word_set_categories
			WHERE course_code = $1
			  AND (($2::bigint IS NULL AND parent_id IS NULL) OR parent_id = $2)
			  AND name = $3
			ORDER BY id
			LIMIT 1
		`, courseCode, nullableInt64(targetParentID), category.Name).Scan(&targetID)
		if err == sql.ErrNoRows {
			err = targetDB.QueryRowContext(ctx, `
				INSERT INTO word_set_categories (
					course_code, parent_id, name, description, is_published,
					sort_order, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				RETURNING id
			`, courseCode, nullableInt64(targetParentID), category.Name,
				nullableString(category.Description), category.IsPublished, category.SortOrder).Scan(&targetID)
		}
		if err != nil {
			return 0, err
		}
		if _, err := targetDB.ExecContext(ctx, `
			UPDATE word_set_categories
			SET description = $2,
			    is_published = $3,
			    sort_order = $4,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`, targetID, nullableString(category.Description), category.IsPublished, category.SortOrder); err != nil {
			return 0, err
		}
		mapped[sourceID] = targetID
		summary.ItemsScanned++
		return targetID, nil
	}

	for sourceID := range categories {
		if _, err := ensureCategory(sourceID); err != nil {
			return nil, err
		}
	}
	return mapped, nil
}

func copyWordSetItems(
	ctx context.Context,
	sourceDB, targetDB *sql.DB,
	sourceSetID, targetSetID int64,
	wordCardMap map[int64]int64,
	summary *writeSummary,
) error {
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT word_card_id, sort_order
		FROM word_set_items
		WHERE word_set_id = $1
		ORDER BY sort_order, id
	`, sourceSetID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var sourceWordCardID int64
		var sortOrder int
		if err := rows.Scan(&sourceWordCardID, &sortOrder); err != nil {
			return err
		}
		summary.ItemsScanned++
		targetWordCardID, ok := wordCardMap[sourceWordCardID]
		if !ok {
			summary.Skipped++
			continue
		}
		if _, err := targetDB.ExecContext(ctx, `
			INSERT INTO word_set_items (word_set_id, word_card_id, sort_order)
			VALUES ($1, $2, $3)
			ON CONFLICT (word_set_id, word_card_id) DO UPDATE SET sort_order = excluded.sort_order
		`, targetSetID, targetWordCardID, sortOrder); err != nil {
			return err
		}
		summary.ItemsInserted++
	}
	return rows.Err()
}
