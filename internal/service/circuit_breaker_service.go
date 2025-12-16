package service

import (
	"fmt"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// CircuitBreakerService manages circuit breaker for worker protection
type CircuitBreakerService struct {
	cbRepo    *repository.CircuitBreakerRepository
	threshold int
	logger    *zap.Logger
}

// NewCircuitBreakerService creates a new circuit breaker service
func NewCircuitBreakerService(
	cbRepo *repository.CircuitBreakerRepository,
	threshold int,
	logger *zap.Logger,
) *CircuitBreakerService {
	return &CircuitBreakerService{
		cbRepo:    cbRepo,
		threshold: threshold,
		logger:    logger,
	}
}

// IsOpen checks if the circuit breaker is open
func (s *CircuitBreakerService) IsOpen() (bool, error) {
	state, err := s.cbRepo.GetState()
	if err != nil {
		return false, fmt.Errorf("failed to get circuit breaker state: %w", err)
	}
	return state.IsOpen, nil
}

// RecordSuccess records a successful operation and resets failure count
func (s *CircuitBreakerService) RecordSuccess() error {
	state, err := s.cbRepo.GetState()
	if err != nil {
		return fmt.Errorf("failed to get circuit breaker state: %w", err)
	}

	// If there were failures, reset them
	if state.FailureCount > 0 {
		s.logger.Info("resetting circuit breaker failure count after success")
		return s.cbRepo.Reset()
	}

	return nil
}

// RecordFailure records a failure and potentially opens the circuit
func (s *CircuitBreakerService) RecordFailure(errorMessage string) error {
	// Record the failure
	if err := s.cbRepo.RecordFailure(errorMessage); err != nil {
		return fmt.Errorf("failed to record failure: %w", err)
	}

	// Check if we should open the circuit
	state, err := s.cbRepo.GetState()
	if err != nil {
		return fmt.Errorf("failed to get circuit breaker state: %w", err)
	}

	if state.FailureCount >= s.threshold && !state.IsOpen {
		s.logger.Warn("opening circuit breaker after threshold reached",
			zap.Int("failure_count", state.FailureCount),
			zap.Int("threshold", s.threshold),
		)
		return s.cbRepo.Open()
	}

	return nil
}

// Reset manually resets the circuit breaker
func (s *CircuitBreakerService) Reset() error {
	s.logger.Info("manually resetting circuit breaker")
	return s.cbRepo.Reset()
}

// GetState gets the current state
func (s *CircuitBreakerService) GetState() (bool, int, string, error) {
	state, err := s.cbRepo.GetState()
	if err != nil {
		return false, 0, "", err
	}
	return state.IsOpen, state.FailureCount, state.LastFailureMessage, nil
}

