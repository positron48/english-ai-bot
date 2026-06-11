package main

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"strings"
)

func mergeTTS(ctx context.Context, sources []openedSourceDB, targetDB *sql.DB) (*writeSummary, error) {
	summary := &writeSummary{Phase: "tts"}
	for _, src := range sources {
		courseCode, ok := sourceCourseCodes[src.Label]
		if !ok || src.DB == nil {
			summary.Skipped++
			continue
		}
		rows, err := src.DB.QueryContext(ctx, `
			SELECT word, state, attempt_count, max_attempts, last_error_code,
			       last_error_message, last_provider, audio_rel_path,
			       last_attempt_at, created_at, updated_at
			FROM tts_generation_status
			ORDER BY word`)
		if err != nil {
			return nil, fmt.Errorf("list tts statuses from %s: %w", src.Label, err)
		}
		for rows.Next() {
			var word, state string
			var attempts, maxAttempts int
			var errorCode, errorMessage, provider, audioPath sql.NullString
			var lastAttempt, createdAt, updatedAt sql.NullTime
			if err := rows.Scan(&word, &state, &attempts, &maxAttempts, &errorCode,
				&errorMessage, &provider, &audioPath, &lastAttempt, &createdAt, &updatedAt); err != nil {
				rows.Close()
				return nil, err
			}
			summary.TTSScanned++
			var scopedPath interface{}
			if audioPath.Valid && strings.TrimSpace(audioPath.String) != "" {
				scopedPath = path.Join(courseCode, strings.TrimLeft(audioPath.String, "/"))
			}
			_, err := targetDB.ExecContext(ctx, `
				INSERT INTO tts_generation_status (
					course_code, word, state, attempt_count, max_attempts,
					last_error_code, last_error_message, last_provider, audio_rel_path,
					last_attempt_at, created_at, updated_at
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,COALESCE($11,CURRENT_TIMESTAMP),COALESCE($12,CURRENT_TIMESTAMP))
				ON CONFLICT (course_code, word) DO UPDATE SET
					state=excluded.state, attempt_count=excluded.attempt_count,
					max_attempts=excluded.max_attempts, last_error_code=excluded.last_error_code,
					last_error_message=excluded.last_error_message, last_provider=excluded.last_provider,
					audio_rel_path=excluded.audio_rel_path, last_attempt_at=excluded.last_attempt_at,
					updated_at=excluded.updated_at`,
				courseCode, word, state, attempts, maxAttempts, nullableStringValue(errorCode),
				nullableStringValue(errorMessage), nullableStringValue(provider), scopedPath,
				nullableTime(lastAttempt), nullableTime(createdAt), nullableTime(updatedAt))
			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("upsert tts %s/%s: %w", courseCode, word, err)
			}
			summary.TTSUpserted++
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return summary, nil
}

func nullableStringValue(v sql.NullString) interface{} {
	if v.Valid {
		return v.String
	}
	return nil
}
