package service

import (
	"errors"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// mockCircuitBreakerRepository is a mock implementation
type mockCircuitBreakerRepository struct {
	state       *models.CircuitBreakerState
	getError    error
	openError   error
	resetError  error
	recordError error
}

func (m *mockCircuitBreakerRepository) GetState() (*models.CircuitBreakerState, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	if m.state == nil {
		return &models.CircuitBreakerState{
			ID:           1,
			IsOpen:       false,
			FailureCount: 0,
		}, nil
	}
	return m.state, nil
}

func (m *mockCircuitBreakerRepository) RecordFailure(errorMessage string) error {
	if m.recordError != nil {
		return m.recordError
	}
	if m.state == nil {
		m.state = &models.CircuitBreakerState{
			ID:           1,
			IsOpen:       false,
			FailureCount: 0,
		}
	}
	m.state.FailureCount++
	m.state.LastFailureMessage = errorMessage
	return nil
}

func (m *mockCircuitBreakerRepository) Open() error {
	if m.openError != nil {
		return m.openError
	}
	if m.state == nil {
		m.state = &models.CircuitBreakerState{}
	}
	m.state.IsOpen = true
	return nil
}

func (m *mockCircuitBreakerRepository) Reset() error {
	if m.resetError != nil {
		return m.resetError
	}
	if m.state == nil {
		m.state = &models.CircuitBreakerState{}
	}
	m.state.IsOpen = false
	m.state.FailureCount = 0
	return nil
}

func TestNewCircuitBreakerService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := repository.NewCircuitBreakerRepository(nil, logger)

	service := NewCircuitBreakerService(repo, 5, logger)
	if service.threshold != 5 {
		t.Errorf("Expected threshold 5, got %d", service.threshold)
	}
}

func TestCircuitBreakerService_IsOpen(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name      string
		state     *models.CircuitBreakerState
		getError  error
		wantOpen  bool
		wantError bool
	}{
		{
			name: "Circuit is closed",
			state: &models.CircuitBreakerState{
				IsOpen: false,
			},
			wantOpen:  false,
			wantError: false,
		},
		{
			name: "Circuit is open",
			state: &models.CircuitBreakerState{
				IsOpen: true,
			},
			wantOpen:  true,
			wantError: false,
		},
		{
			name:      "GetState error",
			getError:  errors.New("database error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCircuitBreakerRepository{
				state:    tt.state,
				getError: tt.getError,
			}

			service := NewCircuitBreakerService(repo, 5, logger)

			gotOpen, err := service.IsOpen()
			if (err != nil) != tt.wantError {
				t.Errorf("IsOpen() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && gotOpen != tt.wantOpen {
				t.Errorf("IsOpen() = %v, want %v", gotOpen, tt.wantOpen)
			}
		})
	}
}

func TestCircuitBreakerService_RecordSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name         string
		initialState *models.CircuitBreakerState
		getError     error
		resetError   error
		wantError    bool
	}{
		{
			name: "Success with no failures",
			initialState: &models.CircuitBreakerState{
				FailureCount: 0,
			},
			wantError: false,
		},
		{
			name: "Success with failures - should reset",
			initialState: &models.CircuitBreakerState{
				FailureCount: 3,
			},
			wantError: false,
		},
		{
			name:      "GetState error",
			getError:  errors.New("database error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCircuitBreakerRepository{
				state:      tt.initialState,
				getError:   tt.getError,
				resetError: tt.resetError,
			}

			service := NewCircuitBreakerService(repo, 5, logger)

			err := service.RecordSuccess()
			if (err != nil) != tt.wantError {
				t.Errorf("RecordSuccess() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCircuitBreakerService_RecordFailure(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name            string
		threshold       int
		initialFailures int
		state           *models.CircuitBreakerState
		recordError     error
		getError        error
		openError       error
		wantError       bool
		shouldOpen      bool
	}{
		{
			name:            "Record failure below threshold",
			threshold:       5,
			initialFailures: 2,
			wantError:       false,
			shouldOpen:      false,
		},
		{
			name:            "Record failure at threshold - should open",
			threshold:       5,
			initialFailures: 4,
			wantError:       false,
			shouldOpen:      true,
		},
		{
			name:            "Record failure above threshold - should open",
			threshold:       5,
			initialFailures: 5,
			wantError:       false,
			shouldOpen:      true,
		},
		{
			name:        "RecordFailure error",
			recordError: errors.New("database error"),
			wantError:   true,
		},
		{
			name:            "Record failure when already open - should not call Open again",
			threshold:       5,
			initialFailures: 10,
			state:           &models.CircuitBreakerState{IsOpen: true, FailureCount: 10},
			wantError:       false,
			shouldOpen:      false,
		},
		{
			name:            "GetState error after RecordFailure",
			threshold:       5,
			initialFailures: 0,
			getError:        errors.New("get state error"),
			wantError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &models.CircuitBreakerState{
				IsOpen:       false,
				FailureCount: tt.initialFailures,
			}
			if tt.state != nil {
				state = tt.state
			}

			repo := &mockCircuitBreakerRepository{
				state:       state,
				recordError: tt.recordError,
				getError:    tt.getError,
				openError:   tt.openError,
			}

			service := NewCircuitBreakerService(repo, tt.threshold, logger)

			err := service.RecordFailure("test error")
			if (err != nil) != tt.wantError {
				t.Errorf("RecordFailure() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError {
				return
			}
			if tt.shouldOpen && !repo.state.IsOpen {
				t.Error("RecordFailure() should have opened circuit")
			}
		})
	}
}

func TestCircuitBreakerService_Reset(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name       string
		resetError error
		wantError  bool
	}{
		{
			name:      "Successful reset",
			wantError: false,
		},
		{
			name:       "Reset error",
			resetError: errors.New("database error"),
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCircuitBreakerRepository{
				resetError: tt.resetError,
			}

			service := NewCircuitBreakerService(repo, 5, logger)

			err := service.Reset()
			if (err != nil) != tt.wantError {
				t.Errorf("Reset() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCircuitBreakerService_GetState(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name      string
		state     *models.CircuitBreakerState
		getError  error
		wantError bool
	}{
		{
			name: "Get state successfully",
			state: &models.CircuitBreakerState{
				ID:                 1,
				IsOpen:             true,
				FailureCount:       5,
				LastFailureMessage: "test error",
			},
			wantError: false,
		},
		{
			name:      "GetState error",
			getError:  errors.New("database error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCircuitBreakerRepository{
				state:    tt.state,
				getError: tt.getError,
			}

			service := NewCircuitBreakerService(repo, 5, logger)

			gotOpen, gotCount, gotMsg, err := service.GetState()
			if (err != nil) != tt.wantError {
				t.Errorf("GetState() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.state != nil {
				if gotOpen != tt.state.IsOpen || gotCount != tt.state.FailureCount || gotMsg != tt.state.LastFailureMessage {
					t.Errorf("GetState() = %v, %v, %q, want %v, %v, %q",
						gotOpen, gotCount, gotMsg, tt.state.IsOpen, tt.state.FailureCount, tt.state.LastFailureMessage)
				}
			}
		})
	}
}
