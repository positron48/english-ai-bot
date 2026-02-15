package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// CircuitBreakerRepository handles database operations for circuit breaker state
type CircuitBreakerRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewCircuitBreakerRepository creates a new circuit breaker repository
func NewCircuitBreakerRepository(db *sql.DB, logger *zap.Logger) *CircuitBreakerRepository {
	return &CircuitBreakerRepository{
		db:     db,
		logger: logger,
	}
}

// GetState gets the circuit breaker state
func (r *CircuitBreakerRepository) GetState() (*models.CircuitBreakerState, error) {
	query := `SELECT id, is_open, failure_count, last_failure_at, 
			  COALESCE(last_failure_message, ''), last_reset_at, updated_at
			  FROM circuit_breaker_state WHERE id = 1`

	var state models.CircuitBreakerState
	var lastFailureAt, lastResetAt, updatedAt sql.NullString

	err := r.db.QueryRow(query).Scan(
		&state.ID, &state.IsOpen, &state.FailureCount,
		&lastFailureAt, &state.LastFailureMessage, &lastResetAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		// Initialize if not exists
		return r.initializeState()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get circuit breaker state: %w", err)
	}

	if lastFailureAt.Valid {
		t, err := time.Parse("2006-01-02 15:04:05", lastFailureAt.String)
		if err != nil {
			// Try alternative formats
			if t2, err2 := time.Parse("2006-01-02T15:04:05Z", lastFailureAt.String); err2 == nil {
				t = t2
			} else if t3, err3 := time.Parse("2006-01-02T15:04:05", lastFailureAt.String); err3 == nil {
				t = t3
			}
		}
		state.LastFailureAt = &t
	}
	if lastResetAt.Valid {
		t, err := time.Parse("2006-01-02 15:04:05", lastResetAt.String)
		if err != nil {
			// Try alternative formats
			if t2, err2 := time.Parse("2006-01-02T15:04:05Z", lastResetAt.String); err2 == nil {
				t = t2
			} else if t3, err3 := time.Parse("2006-01-02T15:04:05", lastResetAt.String); err3 == nil {
				t = t3
			}
		}
		state.LastResetAt = &t
	}
	if updatedAt.Valid {
		var err error
		state.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt.String)
		if err != nil {
			state.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAt.String)
		}
		if err != nil {
			if t2, err2 := time.Parse("2006-01-02T15:04:05Z", updatedAt.String); err2 == nil {
				state.UpdatedAt = t2
			} else if t3, err3 := time.Parse("2006-01-02T15:04:05", updatedAt.String); err3 == nil {
				state.UpdatedAt = t3
			}
		}
	}

	return &state, nil
}

// RecordFailure records a failure
func (r *CircuitBreakerRepository) RecordFailure(errorMessage string) error {
	query := `UPDATE circuit_breaker_state SET
			  failure_count = failure_count + 1,
			  last_failure_at = CURRENT_TIMESTAMP,
			  last_failure_message = ?,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = 1`

	_, err := r.db.Exec(query, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to record failure: %w", err)
	}

	r.logger.Debug("recorded circuit breaker failure", zap.String("error", errorMessage))
	return nil
}

// Open opens the circuit breaker
func (r *CircuitBreakerRepository) Open() error {
	query := `UPDATE circuit_breaker_state SET
			  is_open = 1,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = 1`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to open circuit breaker: %w", err)
	}

	r.logger.Warn("circuit breaker opened")
	return nil
}

// Reset resets the circuit breaker
func (r *CircuitBreakerRepository) Reset() error {
	query := `UPDATE circuit_breaker_state SET
			  is_open = 0,
			  failure_count = 0,
			  last_reset_at = CURRENT_TIMESTAMP,
			  updated_at = CURRENT_TIMESTAMP
			  WHERE id = 1`

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to reset circuit breaker: %w", err)
	}

	r.logger.Info("circuit breaker reset")
	return nil
}

// initializeState initializes the circuit breaker state if not exists
func (r *CircuitBreakerRepository) initializeState() (*models.CircuitBreakerState, error) {
	query := `INSERT INTO circuit_breaker_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING`
	_, err := r.db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize circuit breaker state: %w", err)
	}

	return r.GetState()
}

