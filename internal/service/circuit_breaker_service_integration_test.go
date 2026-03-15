package service

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/testutil"
)

func setupCircuitBreakerServiceTestDB(t *testing.T) (*sql.DB, *repository.CircuitBreakerRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	cbRepo := repository.NewCircuitBreakerRepository(db, logger)

	return db, cbRepo
}

func TestCircuitBreakerService_IsOpen_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, cbRepo := setupCircuitBreakerServiceTestDB(t)

	service := NewCircuitBreakerService(cbRepo, 5, logger)

	// Initially should be closed
	isOpen, err := service.IsOpen()
	if err != nil {
		t.Fatalf("IsOpen() error = %v", err)
	}
	if isOpen {
		t.Error("Circuit breaker should be closed initially")
	}
}

func TestCircuitBreakerService_RecordSuccess_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, cbRepo := setupCircuitBreakerServiceTestDB(t)

	service := NewCircuitBreakerService(cbRepo, 5, logger)

	// Record success (no prior failures — FailureCount stays 0, no reset)
	err := service.RecordSuccess()
	if err != nil {
		t.Fatalf("RecordSuccess() error = %v", err)
	}
}

// TestCircuitBreakerService_RecordSuccess_AfterFailures_Integration covers RecordSuccess when FailureCount > 0 (reset branch).
func TestCircuitBreakerService_RecordSuccess_AfterFailures_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, cbRepo := setupCircuitBreakerServiceTestDB(t)

	service := NewCircuitBreakerService(cbRepo, 5, logger)

	// Record some failures so FailureCount > 0
	for i := 0; i < 2; i++ {
		if err := service.RecordFailure("test failure"); err != nil {
			t.Fatalf("RecordFailure() error = %v", err)
		}
	}
	isOpen, fc, _, err := service.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if isOpen || fc != 2 {
		t.Fatalf("expected closed with 2 failures, got isOpen=%v failureCount=%d", isOpen, fc)
	}

	// RecordSuccess should reset failure count
	err = service.RecordSuccess()
	if err != nil {
		t.Fatalf("RecordSuccess() after failures error = %v", err)
	}
	_, fc2, _, err := service.GetState()
	if err != nil {
		t.Fatalf("GetState() after success error = %v", err)
	}
	if fc2 != 0 {
		t.Errorf("expected failure count 0 after RecordSuccess, got %d", fc2)
	}
}

func TestCircuitBreakerService_RecordFailure_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, cbRepo := setupCircuitBreakerServiceTestDB(t)

	service := NewCircuitBreakerService(cbRepo, 3, logger) // Lower threshold for testing

	// Record failures up to threshold
	for i := 0; i < 3; i++ {
		err := service.RecordFailure("test error")
		if err != nil {
			t.Fatalf("RecordFailure() error = %v", err)
		}
	}

	// Record one more failure to trigger opening
	err := service.RecordFailure("test error")
	if err != nil {
		t.Fatalf("RecordFailure() error = %v", err)
	}

	// Circuit should be open now
	isOpen, err := service.IsOpen()
	if err != nil {
		t.Fatalf("IsOpen() error = %v", err)
	}
	if !isOpen {
		t.Error("Circuit breaker should be open after threshold failures")
	}
}

func TestCircuitBreakerService_Reset_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, cbRepo := setupCircuitBreakerServiceTestDB(t)

	service := NewCircuitBreakerService(cbRepo, 3, logger)

	// Open the circuit
	for i := 0; i < 3; i++ {
		service.RecordFailure("test error")
	}

	// Reset
	err := service.Reset()
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	// Circuit should be closed
	isOpen, err := service.IsOpen()
	if err != nil {
		t.Fatalf("IsOpen() error = %v", err)
	}
	if isOpen {
		t.Error("Circuit breaker should be closed after reset")
	}
}

func TestCircuitBreakerService_GetState_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, cbRepo := setupCircuitBreakerServiceTestDB(t)

	service := NewCircuitBreakerService(cbRepo, 5, logger)

	// Get initial state
	isOpen, failureCount, lastMessage, err := service.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	// Verify we can get state
	_ = isOpen
	_ = failureCount
	_ = lastMessage
}
