package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewCircuitBreakerRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)
	if repo == nil {
		t.Error("NewCircuitBreakerRepository() should not return nil")
	}
}

func TestCircuitBreakerRepository_GetState(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	t.Run("Get state when not exists - should initialize", func(t *testing.T) {
		state, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}
		if state == nil {
			t.Fatal("GetState() should not return nil")
		}
		if state.ID != 1 {
			t.Errorf("Expected ID 1, got %d", state.ID)
		}
		if state.IsOpen {
			t.Error("Initial state should not be open")
		}
		if state.FailureCount != 0 {
			t.Errorf("Expected failure count 0, got %d", state.FailureCount)
		}
	})

	t.Run("Get state after initialization", func(t *testing.T) {
		state, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}
		if state == nil {
			t.Fatal("GetState() should not return nil")
		}
	})
}

func TestCircuitBreakerRepository_RecordFailure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	// Initialize state first
	_, err := repo.GetState()
	if err != nil {
		t.Fatalf("Failed to initialize state: %v", err)
	}

	t.Run("Record first failure", func(t *testing.T) {
		err := repo.RecordFailure("test error")
		if err != nil {
			t.Fatalf("RecordFailure() error = %v", err)
		}

		state, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}
		if state.FailureCount != 1 {
			t.Errorf("Expected failure count 1, got %d", state.FailureCount)
		}
		if state.LastFailureMessage != "test error" {
			t.Errorf("Expected error message 'test error', got %q", state.LastFailureMessage)
		}
		if state.LastFailureAt == nil {
			t.Error("LastFailureAt should be set")
		}
	})

	t.Run("Record multiple failures", func(t *testing.T) {
		repo.RecordFailure("error 1")
		repo.RecordFailure("error 2")
		repo.RecordFailure("error 3")

		state, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}
		if state.FailureCount != 4 { // 1 from previous test + 3 new
			t.Errorf("Expected failure count 4, got %d", state.FailureCount)
		}
	})
}

func TestCircuitBreakerRepository_Open(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	// Initialize state first
	_, err := repo.GetState()
	if err != nil {
		t.Fatalf("Failed to initialize state: %v", err)
	}

	t.Run("Open circuit breaker", func(t *testing.T) {
		err := repo.Open()
		if err != nil {
			t.Fatalf("Open() error = %v", err)
		}

		state, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}
		if !state.IsOpen {
			t.Error("Circuit breaker should be open")
		}
	})
}

func TestCircuitBreakerRepository_Reset(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	// Initialize state first
	_, err := repo.GetState()
	if err != nil {
		t.Fatalf("Failed to initialize state: %v", err)
	}

	// Record some failures and open the circuit
	repo.RecordFailure("error 1")
	repo.RecordFailure("error 2")
	repo.Open()

	t.Run("Reset circuit breaker", func(t *testing.T) {
		err := repo.Reset()
		if err != nil {
			t.Fatalf("Reset() error = %v", err)
		}

		state, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState() error = %v", err)
		}
		if state.IsOpen {
			t.Error("Circuit breaker should be closed after reset")
		}
		if state.FailureCount != 0 {
			t.Errorf("Expected failure count 0, got %d", state.FailureCount)
		}
		if state.LastResetAt == nil {
			t.Error("LastResetAt should be set")
		}
	})
}

func TestCircuitBreakerRepository_InitializeState(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	// Delete any existing state
	db.Exec("DELETE FROM circuit_breaker_state")

	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState() should initialize state: %v", err)
	}
	if state == nil {
		t.Fatal("GetState() should return initialized state")
	}
	if state.ID != 1 {
		t.Errorf("Expected ID 1, got %d", state.ID)
	}
}

func TestCircuitBreakerRepository_StateTimestamps(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	// Initialize state
	_, err := repo.GetState()
	if err != nil {
		t.Fatalf("Failed to initialize state: %v", err)
	}

	// Record failure
	err = repo.RecordFailure("test")
	if err != nil {
		t.Fatalf("RecordFailure() error = %v", err)
	}

	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}

	// Check that timestamps are set
	if state.LastFailureAt == nil {
		t.Error("LastFailureAt should be set")
	}
	if state.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	// Reset and check reset timestamp
	err = repo.Reset()
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	state, err = repo.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}

	if state.LastResetAt == nil {
		t.Error("LastResetAt should be set after reset")
	}
}

func TestCircuitBreakerRepository_GetState_AlternativeTimeFormats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	_, _ = repo.GetState()
	_ = repo.RecordFailure("err")

	// Force ISO-style timestamp so GetState hits alternative parse branches
	_, err := db.Exec(`UPDATE circuit_breaker_state SET last_failure_at = '2024-06-15T14:30:00Z'::timestamptz, last_reset_at = '2024-06-14T10:00:00Z'::timestamptz, updated_at = '2024-06-15T15:00:00Z'::timestamptz WHERE id = 1`)
	if err != nil {
		t.Fatalf("UPDATE state: %v", err)
	}

	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if state == nil {
		t.Fatal("GetState() should not return nil")
	}
	if state.LastFailureAt == nil {
		t.Error("LastFailureAt should be parsed from alternative format")
	}
	if state.LastResetAt == nil {
		t.Error("LastResetAt should be parsed from alternative format")
	}
	if state.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be parsed")
	}
}

// TestCircuitBreakerRepository_GetState_StandardTimeFormat covers GetState when DB returns
// timestamps in "2006-01-02 15:04:05" format (first parse succeeds for lastFailureAt/lastResetAt).
func TestCircuitBreakerRepository_GetState_StandardTimeFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTestDB(t)
	repo := NewCircuitBreakerRepository(db, logger)

	_, _ = repo.GetState()
	_ = repo.RecordFailure("err")

	// Some drivers return space-separated datetime; ensure we still parse correctly
	_, err := db.Exec(`UPDATE circuit_breaker_state SET last_failure_at = '2024-06-15 14:30:00', last_reset_at = NULL, updated_at = '2024-06-15 15:00:00' WHERE id = 1`)
	if err != nil {
		t.Skipf("UPDATE with standard format not supported: %v", err)
	}

	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if state == nil {
		t.Fatal("GetState() should not return nil")
	}
	if state.LastFailureAt == nil {
		t.Error("LastFailureAt should be set")
	}
}
