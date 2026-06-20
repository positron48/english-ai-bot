package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// TTSCircuitBreakerRepository handles database operations for the pronunciation/TTS
// circuit breaker state. It is kept separate from CircuitBreakerRepository (used for
// AI word-card generation) so a billing/outage issue with one provider doesn't pause
// the other pipeline.
type TTSCircuitBreakerRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewTTSCircuitBreakerRepository creates a new TTS circuit breaker repository.
func NewTTSCircuitBreakerRepository(db *sql.DB, logger *zap.Logger) *TTSCircuitBreakerRepository {
	return &TTSCircuitBreakerRepository{
		db:     db,
		logger: logger,
	}
}

// GetState gets the circuit breaker state.
func (r *TTSCircuitBreakerRepository) GetState() (*models.CircuitBreakerState, error) {
	query := `SELECT id, is_open, failure_count,
			  COALESCE(last_failure_at::text, ''), COALESCE(last_failure_message, ''),
			  COALESCE(last_reset_at::text, ''), COALESCE(updated_at::text, '')
			  FROM tts_circuit_breaker_state WHERE id = 1`

	var state models.CircuitBreakerState
	var lastFailureAt, lastResetAt, updatedAt sql.NullString

	err := r.db.QueryRow(query).Scan(
		&state.ID, &state.IsOpen, &state.FailureCount,
		&lastFailureAt, &state.LastFailureMessage, &lastResetAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return r.initializeState()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tts circuit breaker state: %w", err)
	}

	if lastFailureAt.Valid && lastFailureAt.String != "" {
		if t, perr := time.Parse("2006-01-02 15:04:05", lastFailureAt.String); perr == nil {
			state.LastFailureAt = &t
		}
	}
	if lastResetAt.Valid && lastResetAt.String != "" {
		if t, perr := time.Parse("2006-01-02 15:04:05", lastResetAt.String); perr == nil {
			state.LastResetAt = &t
		}
	}
	if updatedAt.Valid && updatedAt.String != "" {
		if t, perr := time.Parse("2006-01-02 15:04:05", updatedAt.String); perr == nil {
			state.UpdatedAt = t
		}
	}

	return &state, nil
}

// RecordFailure records a provider-level failure.
func (r *TTSCircuitBreakerRepository) RecordFailure(errorMessage string) error {
	query := `UPDATE tts_circuit_breaker_state SET
			  failure_count = failure_count + 1,
			  last_failure_at = CURRENT_TIMESTAMP,
			  last_failure_message = ?,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = 1`

	_, err := r.db.Exec(query, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to record tts circuit breaker failure: %w", err)
	}

	r.logger.Debug("recorded tts circuit breaker failure", zap.String("error", errorMessage))
	return nil
}

// Open opens the circuit breaker.
func (r *TTSCircuitBreakerRepository) Open() error {
	query := `UPDATE tts_circuit_breaker_state SET
			  is_open = 1,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = 1`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to open tts circuit breaker: %w", err)
	}

	r.logger.Warn("tts circuit breaker opened")
	return nil
}

// Reset resets the circuit breaker.
func (r *TTSCircuitBreakerRepository) Reset() error {
	query := `UPDATE tts_circuit_breaker_state SET
			  is_open = 0,
			  failure_count = 0,
			  last_reset_at = CURRENT_TIMESTAMP,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = 1`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to reset tts circuit breaker: %w", err)
	}

	r.logger.Info("tts circuit breaker reset")
	return nil
}

func (r *TTSCircuitBreakerRepository) initializeState() (*models.CircuitBreakerState, error) {
	query := `INSERT INTO tts_circuit_breaker_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING`
	if _, err := r.db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to initialize tts circuit breaker state: %w", err)
	}

	return r.GetState()
}
