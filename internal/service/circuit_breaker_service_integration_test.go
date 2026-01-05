package service

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupCircuitBreakerServiceTestDB(t *testing.T) (*sql.DB, *repository.CircuitBreakerRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS circuit_breaker_state (
		id INTEGER PRIMARY KEY,
		is_open INTEGER NOT NULL DEFAULT 0,
		failure_count INTEGER NOT NULL DEFAULT 0,
		last_failure_message TEXT,
		last_failure_at TEXT,
		opened_at TEXT,
		last_reset_at TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = db.Exec(createTable)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	cbRepo := repository.NewCircuitBreakerRepository(db, logger)

	return db, cbRepo
}

func TestCircuitBreakerService_IsOpen_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, cbRepo := setupCircuitBreakerServiceTestDB(t)
	defer db.Close()

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
	db, cbRepo := setupCircuitBreakerServiceTestDB(t)
	defer db.Close()

	service := NewCircuitBreakerService(cbRepo, 5, logger)

	// Record success
	err := service.RecordSuccess()
	if err != nil {
		t.Fatalf("RecordSuccess() error = %v", err)
	}
}

func TestCircuitBreakerService_RecordFailure_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, cbRepo := setupCircuitBreakerServiceTestDB(t)
	defer db.Close()

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
	db, cbRepo := setupCircuitBreakerServiceTestDB(t)
	defer db.Close()

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
	db, cbRepo := setupCircuitBreakerServiceTestDB(t)
	defer db.Close()

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
