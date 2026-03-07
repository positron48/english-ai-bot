package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"unicode"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

type TTSStatusRepository struct {
	db          *sql.DB
	logger      *zap.Logger
	maxAttempts int
}

// NewTTSStatusRepository creates a TTS status repository. maxAttempts is the default
// for new/updated rows (retries for pronunciation generation); if <= 0, 3 is used.
func NewTTSStatusRepository(db *sql.DB, logger *zap.Logger, maxAttempts int) *TTSStatusRepository {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	return &TTSStatusRepository{db: db, logger: logger, maxAttempts: maxAttempts}
}

func (r *TTSStatusRepository) GetByWord(word string) (*models.TTSGenerationStatus, error) {
	normalized, ok := normalizeTTSWord(word)
	if !ok {
		return nil, nil
	}

	const q = `SELECT word, state, attempt_count, max_attempts, last_error_code, last_error_message,
		last_provider, audio_rel_path, last_attempt_at, created_at, updated_at
		FROM tts_generation_status
		WHERE word = ?`

	var s models.TTSGenerationStatus
	var lastErrorCode, lastErrorMessage, lastProvider, audioRelPath sql.NullString
	var lastAttemptAt sql.NullTime
	err := r.db.QueryRow(q, normalized).Scan(
		&s.Word, &s.State, &s.AttemptCount, &s.MaxAttempts,
		&lastErrorCode, &lastErrorMessage, &lastProvider, &audioRelPath,
		&lastAttemptAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tts status by word: %w", err)
	}
	if lastErrorCode.Valid {
		s.LastErrorCode = &lastErrorCode.String
	}
	if lastErrorMessage.Valid {
		s.LastErrorMessage = &lastErrorMessage.String
	}
	if lastProvider.Valid {
		s.LastProvider = &lastProvider.String
	}
	if audioRelPath.Valid {
		s.AudioRelPath = &audioRelPath.String
	}
	if lastAttemptAt.Valid {
		s.LastAttemptAt = &lastAttemptAt.Time
	}
	return &s, nil
}

func (r *TTSStatusRepository) UpsertPending(word string) error {
	normalized, ok := normalizeTTSWord(word)
	if !ok {
		return nil
	}

	const q = `INSERT INTO tts_generation_status (word, state, attempt_count, max_attempts, created_at, updated_at)
		VALUES (?, ?, 0, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(word) DO UPDATE SET
			state = excluded.state,
			updated_at = CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, normalized, models.TTSStatePending, r.maxAttempts); err != nil {
		return fmt.Errorf("upsert pending tts status: %w", err)
	}
	return nil
}

func (r *TTSStatusRepository) MarkAttempt(word, provider, errorCode, errorMessage string, retryable bool) error {
	normalized, ok := normalizeTTSWord(word)
	if !ok {
		return nil
	}

	state := models.TTSStateFailedRetryable
	if !retryable {
		state = models.TTSStateFailedTerminal
	}

	const upsert = `INSERT INTO tts_generation_status (
			word, state, attempt_count, max_attempts, last_error_code, last_error_message, last_provider, last_attempt_at, created_at, updated_at
		) VALUES (?, ?, 1, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(word) DO UPDATE SET
			state = excluded.state,
			attempt_count = tts_generation_status.attempt_count + 1,
			last_error_code = excluded.last_error_code,
			last_error_message = excluded.last_error_message,
			last_provider = excluded.last_provider,
			last_attempt_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(upsert, normalized, state, r.maxAttempts, nullableString(errorCode), nullableString(errorMessage), nullableString(provider)); err != nil {
		return fmt.Errorf("mark tts attempt: %w", err)
	}

	const capToTerminal = `UPDATE tts_generation_status
		SET state = ?, updated_at = CURRENT_TIMESTAMP
		WHERE word = ? AND attempt_count >= max_attempts`
	if _, err := r.db.Exec(capToTerminal, models.TTSStateFailedTerminal, normalized); err != nil {
		return fmt.Errorf("mark terminal on max attempts: %w", err)
	}
	return nil
}

func (r *TTSStatusRepository) MarkReady(word, provider, relPath string) error {
	normalized, ok := normalizeTTSWord(word)
	if !ok {
		return nil
	}
	const q = `INSERT INTO tts_generation_status (
			word, state, attempt_count, max_attempts, last_provider, audio_rel_path, last_error_code, last_error_message, created_at, updated_at
		) VALUES (?, ?, 0, ?, ?, ?, NULL, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(word) DO UPDATE SET
			state = excluded.state,
			audio_rel_path = excluded.audio_rel_path,
			last_provider = excluded.last_provider,
			last_error_code = NULL,
			last_error_message = NULL,
			updated_at = CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, normalized, models.TTSStateReady, r.maxAttempts, nullableString(provider), nullableString(relPath)); err != nil {
		return fmt.Errorf("mark tts ready: %w", err)
	}
	return nil
}

func (r *TTSStatusRepository) MarkTerminal(word, provider, errorCode, errorMessage string) error {
	normalized, ok := normalizeTTSWord(word)
	if !ok {
		return nil
	}
	const q = `INSERT INTO tts_generation_status (
			word, state, attempt_count, max_attempts, last_provider, last_error_code, last_error_message, last_attempt_at, created_at, updated_at
		) VALUES (?, ?, 0, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(word) DO UPDATE SET
			state = excluded.state,
			last_provider = excluded.last_provider,
			last_error_code = excluded.last_error_code,
			last_error_message = excluded.last_error_message,
			last_attempt_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, normalized, models.TTSStateFailedTerminal, r.maxAttempts, nullableString(provider), nullableString(errorCode), nullableString(errorMessage)); err != nil {
		return fmt.Errorf("mark tts terminal: %w", err)
	}
	return nil
}

func (r *TTSStatusRepository) ResetForForceRegenerate(word string) error {
	normalized, ok := normalizeTTSWord(word)
	if !ok {
		return nil
	}
	const q = `INSERT INTO tts_generation_status (
			word, state, attempt_count, max_attempts, last_error_code, last_error_message, last_provider, audio_rel_path, created_at, updated_at
		) VALUES (?, ?, 0, ?, NULL, NULL, NULL, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(word) DO UPDATE SET
			state = excluded.state,
			attempt_count = 0,
			last_error_code = NULL,
			last_error_message = NULL,
			last_provider = NULL,
			audio_rel_path = NULL,
			last_attempt_at = NULL,
			updated_at = CURRENT_TIMESTAMP`
	if _, err := r.db.Exec(q, normalized, models.TTSStatePending, r.maxAttempts); err != nil {
		return fmt.Errorf("reset tts status for regenerate: %w", err)
	}
	return nil
}

func nullableString(v string) interface{} {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func normalizeTTSWord(raw string) (string, bool) {
	word := strings.TrimSpace(strings.ToLower(raw))
	word = strings.Trim(word, ".,!?;:()[]{}\"`")
	word = strings.Join(strings.Fields(word), " ")
	if word == "" {
		return "", false
	}

	hasLatin := false
	for _, r := range word {
		if unicode.IsLetter(r) {
			if !unicode.Is(unicode.Latin, r) {
				return "", false
			}
			hasLatin = true
			continue
		}
		switch r {
		case ' ', '-', '\'', '’':
			continue
		default:
			return "", false
		}
	}

	if !hasLatin {
		return "", false
	}
	return word, true
}
