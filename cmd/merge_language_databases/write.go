package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

var sourceCourseCodes = map[string]string{
	"english": "en_ru",
	"spanish": "es_ru",
}

func runWritePhase(ctx context.Context, phase string, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	switch strings.TrimSpace(strings.ToLower(phase)) {
	case "users":
		return mergeUsers(ctx, sources, targetDB)
	case "user-courses", "user_courses":
		return mergeUserCourses(ctx, sources, targetDB)
	case "course-mappings", "course_mappings":
		return mergeCourseMappings(ctx, sources, targetDB)
	case "content":
		return mergeContent(ctx, sources, targetDB)
	case "attempts":
		return mergeAttempts(ctx, sources, targetDB)
	case "srs":
		return mergeSRS(ctx, sources, targetDB)
	case "legacy-words", "legacy_words":
		return mergeLegacyWords(ctx, sources, targetDB)
	case "reset-word-items", "reset_word_items":
		return resetWordItems(ctx, sources, targetDB)
	default:
		return nil, fmt.Errorf("unsupported write phase %q", phase)
	}
}

type sourceUserRow struct {
	ID                    int64
	TelegramID            sql.NullInt64
	TelegramUsername      string
	Timezone              string
	PreferredTrainingTime string
	SettingsJSON          string
	SubscriptionTier      string
}

func mergeUsers(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "users"}
	for _, src := range sources {
		if src.DB == nil {
			continue
		}
		rows, err := src.DB.QueryContext(ctx, `
			SELECT id, telegram_id,
			       COALESCE(NULLIF(TRIM(telegram_username), ''), ''),
			       COALESCE(timezone, ''),
			       COALESCE(preferred_training_time, ''),
			       COALESCE(settings_json, '{}'),
			       COALESCE(subscription_tier, 'free')
			FROM users
			ORDER BY id
		`)
		if err != nil {
			return nil, fmt.Errorf("list users from %s: %w", src.Label, err)
		}
		for rows.Next() {
			var user sourceUserRow
			if err := rows.Scan(
				&user.ID,
				&user.TelegramID,
				&user.TelegramUsername,
				&user.Timezone,
				&user.PreferredTrainingTime,
				&user.SettingsJSON,
				&user.SubscriptionTier,
			); err != nil {
				rows.Close()
				return nil, err
			}
			summary.UsersScanned++
			if err := upsertLegacyUserMapping(ctx, targetDB, src.Label, user, summary); err != nil {
				rows.Close()
				return nil, err
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return summary, nil
}

func upsertLegacyUserMapping(ctx context.Context, targetDB *sql.DB, sourceLabel string, user sourceUserRow, summary *writeSummary) error {
	sourceUserID := strconv.FormatInt(user.ID, 10)
	var existingTarget sql.NullInt64
	err := targetDB.QueryRowContext(ctx, `
		SELECT target_user_id
		FROM legacy_user_mappings
		WHERE source_app_code = $1
		  AND source_db_label = $1
		  AND source_table = 'users'
		  AND source_user_id = $2
	`, sourceLabel, sourceUserID).Scan(&existingTarget)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && existingTarget.Valid {
		summary.MappingsExisting++
		return nil
	}

	targetUserID, reused, err := resolveTargetUserID(ctx, targetDB, user)
	if err != nil {
		return err
	}
	if reused {
		summary.UsersReused++
	} else {
		summary.UsersInserted++
	}

	res, err := targetDB.ExecContext(ctx, `
		INSERT INTO legacy_user_mappings (
			source_app_code, source_db_label, source_table, source_user_id,
			target_user_id, stable_identity_type, stable_identity_value,
			mapping_status, metadata_json, created_at, updated_at
		) VALUES (
			$1, $1, 'users', $2,
			$3,
			CASE WHEN $4::bigint IS NULL THEN NULL ELSE 'telegram_id' END,
			CASE WHEN $4::bigint IS NULL THEN NULL ELSE $4::text END,
			'mapped', '{}'::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		ON CONFLICT (source_app_code, source_db_label, source_table, source_user_id) DO NOTHING
	`, sourceLabel, sourceUserID, targetUserID, nullableInt64(user.TelegramID))
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected > 0 {
		summary.MappingsCreated++
	}
	return nil
}

func resolveTargetUserID(ctx context.Context, targetDB *sql.DB, user sourceUserRow) (int64, bool, error) {
	if user.TelegramID.Valid {
		var targetID int64
		err := targetDB.QueryRowContext(ctx, `SELECT id FROM users WHERE telegram_id = $1`, user.TelegramID.Int64).Scan(&targetID)
		if err == nil {
			return targetID, true, nil
		}
		if err != sql.ErrNoRows {
			return 0, false, err
		}
	}

	var targetID int64
	err := targetDB.QueryRowContext(ctx, `
		INSERT INTO users (
			telegram_id, telegram_username, timezone, preferred_training_time,
			settings_json, subscription_tier, created_at, updated_at
		) VALUES (
			$1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''),
			$5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING id
	`,
		nullableInt64(user.TelegramID),
		user.TelegramUsername,
		user.Timezone,
		user.PreferredTrainingTime,
		user.SettingsJSON,
		user.SubscriptionTier,
	).Scan(&targetID)
	if err != nil {
		return 0, false, err
	}
	return targetID, false, nil
}

func mergeUserCourses(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "user-courses"}
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
		if src.DB == nil {
			continue
		}
		rows, err := targetDB.QueryContext(ctx, `
			SELECT source_user_id, target_user_id
			FROM legacy_user_mappings
			WHERE source_app_code = $1
			  AND source_db_label = $1
			  AND source_table = 'users'
			  AND mapping_status = 'mapped'
			  AND target_user_id IS NOT NULL
		`, src.Label)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var sourceUserID string
			var targetUserID int64
			if err := rows.Scan(&sourceUserID, &targetUserID); err != nil {
				rows.Close()
				return nil, err
			}
			summary.UsersScanned++
			res, err := targetDB.ExecContext(ctx, `
				INSERT INTO user_courses (user_id, course_id, status, started_at, created_at, updated_at)
				SELECT $1, $2, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
				WHERE NOT EXISTS (
					SELECT 1 FROM user_courses
					WHERE user_id = $1 AND course_id = $2
				)
			`, targetUserID, courseID)
			if err != nil {
				rows.Close()
				return nil, err
			}
			affected, _ := res.RowsAffected()
			if affected > 0 {
				summary.UserCoursesAdded += affected
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return summary, nil
}

func nullableInt64(v sql.NullInt64) interface{} {
	if v.Valid {
		return v.Int64
	}
	return nil
}
