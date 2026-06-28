package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func resetTTSCircuitBreaker(t *testing.T, conn *sql.DB) {
	t.Helper()
	if _, err := conn.Exec(`DELETE FROM tts_circuit_breaker_state`); err != nil {
		t.Fatalf("delete tts state: %v", err)
	}
}

func TestTTSCircuitBreakerRepository_GetState(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	resetTTSCircuitBreaker(t, conn)
	repo := NewTTSCircuitBreakerRepository(conn, zap.NewNop())

	t.Run("initializes when row missing", func(t *testing.T) {
		if _, err := conn.Exec(`DELETE FROM tts_circuit_breaker_state`); err != nil {
			t.Fatalf("delete state: %v", err)
		}

		state, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState: %v", err)
		}
		if state == nil || state.ID != 1 || state.IsOpen || state.FailureCount != 0 {
			t.Fatalf("state = %+v", state)
		}
	})

	t.Run("reads existing state", func(t *testing.T) {
		state, err := repo.GetState()
		if err != nil {
			t.Fatalf("GetState: %v", err)
		}
		if state == nil {
			t.Fatal("expected state")
		}
	})
}

func TestTTSCircuitBreakerRepository_RecordFailureOpenReset(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	resetTTSCircuitBreaker(t, conn)
	repo := NewTTSCircuitBreakerRepository(conn, zap.NewNop())

	if _, err := repo.GetState(); err != nil {
		t.Fatalf("init state: %v", err)
	}

	if err := repo.RecordFailure("tts outage"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}
	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState after failure: %v", err)
	}
	if state.FailureCount != 1 || state.LastFailureMessage != "tts outage" {
		t.Fatalf("state after failure = %+v", state)
	}

	if err := repo.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	state, err = repo.GetState()
	if err != nil {
		t.Fatalf("GetState after open: %v", err)
	}
	if !state.IsOpen {
		t.Fatal("expected open circuit")
	}

	if err := repo.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	state, err = repo.GetState()
	if err != nil {
		t.Fatalf("GetState after reset: %v", err)
	}
	if state.IsOpen || state.FailureCount != 0 {
		t.Fatalf("state after reset = %+v", state)
	}
}

func TestTTSCircuitBreakerRepository_GetState_Timestamps(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	resetTTSCircuitBreaker(t, conn)
	repo := NewTTSCircuitBreakerRepository(conn, zap.NewNop())

	if _, err := repo.GetState(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := repo.RecordFailure("err"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}

	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state.FailureCount != 1 || state.LastFailureMessage != "err" {
		t.Fatalf("state = %+v", state)
	}
}

func TestTTSCircuitBreakerRepository_MultipleFailures(t *testing.T) {
	conn := testutil.SetupTestDB(t)
	resetTTSCircuitBreaker(t, conn)
	repo := NewTTSCircuitBreakerRepository(conn, zap.NewNop())
	_, _ = repo.GetState()

	for i := 0; i < 3; i++ {
		if err := repo.RecordFailure("err"); err != nil {
			t.Fatalf("RecordFailure %d: %v", i, err)
		}
	}
	state, err := repo.GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state.FailureCount != 3 {
		t.Fatalf("failure count = %d, want 3", state.FailureCount)
	}
}
