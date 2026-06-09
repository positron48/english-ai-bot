package main

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
)

// auditTelegram checks cross-app telegram overlap and per-DB duplicate telegram_id.
func auditTelegram(ctx context.Context, dbs []openedSourceDB, targetDB *sql.DB) ([]telegramMultiCourseUser, []identityConflict, error) {
	var enDB, esDB *sql.DB
	for _, src := range dbs {
		switch src.Label {
		case "english":
			enDB = src.DB
		case "spanish":
			esDB = src.DB
		}
	}

	var conflicts []identityConflict
	for _, src := range dbs {
		if src.DB == nil {
			continue
		}
		dupes, err := duplicateTelegramIDs(ctx, src.Label, src.DB)
		if err != nil {
			return nil, nil, err
		}
		conflicts = append(conflicts, dupes...)
	}

	var multi []telegramMultiCourseUser
	if enDB != nil && esDB != nil {
		enByTelegram, err := telegramUserIDs(ctx, enDB)
		if err != nil {
			return nil, nil, err
		}
		rows, err := esDB.QueryContext(ctx, `SELECT id, telegram_id FROM users WHERE telegram_id IS NOT NULL`)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var userID, telegramID int64
			if err := rows.Scan(&userID, &telegramID); err != nil {
				return nil, nil, err
			}
			if enUserID, ok := enByTelegram[telegramID]; ok {
				multi = append(multi, telegramMultiCourseUser{
					TelegramID: telegramID,
					SourceLabels: []string{"english", "spanish"},
					SourceUserIDs: sortedCopy([]string{
						fmt.Sprintf("english:%d", enUserID),
						fmt.Sprintf("spanish:%d", userID),
					}),
				})
			}
		}
		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
		sort.Slice(multi, func(i, j int) bool { return multi[i].TelegramID < multi[j].TelegramID })
	}

	if targetDB != nil {
		targetConflicts, err := targetTelegramConflicts(ctx, dbs, targetDB)
		if err != nil {
			return nil, nil, err
		}
		conflicts = append(conflicts, targetConflicts...)
	}
	if conflicts == nil {
		conflicts = []identityConflict{}
	}
	if multi == nil {
		multi = []telegramMultiCourseUser{}
	}
	return multi, conflicts, nil
}

func telegramUserIDs(ctx context.Context, db *sql.DB) (map[int64]int64, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, telegram_id FROM users WHERE telegram_id IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]int64{}
	for rows.Next() {
		var userID, telegramID int64
		if err := rows.Scan(&userID, &telegramID); err != nil {
			return nil, err
		}
		out[telegramID] = userID
	}
	return out, rows.Err()
}

func duplicateTelegramIDs(ctx context.Context, label string, db *sql.DB) ([]identityConflict, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT u.telegram_id, u.id
		FROM users u
		WHERE u.telegram_id IS NOT NULL
		  AND u.telegram_id IN (
			SELECT telegram_id
			FROM users
			WHERE telegram_id IS NOT NULL
			GROUP BY telegram_id
			HAVING COUNT(*) > 1
		  )
		ORDER BY u.telegram_id, u.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byTelegram := map[int64]*identityConflict{}
	var order []int64
	for rows.Next() {
		var telegramID, userID int64
		if err := rows.Scan(&telegramID, &userID); err != nil {
			return nil, err
		}
		conflict := byTelegram[telegramID]
		if conflict == nil {
			conflict = &identityConflict{
				TelegramID:   telegramID,
				SourceLabels: []string{label},
			}
			byTelegram[telegramID] = conflict
			order = append(order, telegramID)
		}
		conflict.SourceUserIDs = append(conflict.SourceUserIDs, fmt.Sprintf("%s:%d", label, userID))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]identityConflict, 0, len(order))
	for _, telegramID := range order {
		out = append(out, *byTelegram[telegramID])
	}
	return out, nil
}

func targetTelegramConflicts(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) ([]identityConflict, error) {
	seen := map[int64]*identityConflict{}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		rows, err := src.DB.QueryContext(ctx, `SELECT id, telegram_id FROM users WHERE telegram_id IS NOT NULL`)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var userID, telegramID int64
			if err := rows.Scan(&userID, &telegramID); err != nil {
				rows.Close()
				return nil, err
			}
			entry := seen[telegramID]
			if entry == nil {
				entry = &identityConflict{TelegramID: telegramID}
				seen[telegramID] = entry
			}
			entry.SourceLabels = append(entry.SourceLabels, src.Label)
			entry.SourceUserIDs = append(entry.SourceUserIDs, fmt.Sprintf("%s:%d", src.Label, userID))
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	rows, err := targetDB.QueryContext(ctx, `SELECT id, telegram_id FROM users WHERE telegram_id IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID, telegramID int64
		if err := rows.Scan(&userID, &telegramID); err != nil {
			return nil, err
		}
		if entry := seen[telegramID]; entry != nil {
			entry.TargetUserIDs = append(entry.TargetUserIDs, fmt.Sprintf("target:%d", userID))
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	_, conflicts := classifyTelegramEntries(seen)
	return conflicts, nil
}
