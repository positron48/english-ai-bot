//go:build integration

package testkit

import (
	"context"
	"strings"
	"testing"
)

// TestStartPostgres_StartsContainer covers StartPostgres with testcontainers.
// Skips if Docker is unavailable.
func TestStartPostgres_StartsContainer(t *testing.T) {
	pc, cleanup := StartPostgres(t)
	defer cleanup()
	if pc == nil {
		t.Fatal("StartPostgres: expected non-nil container")
	}
	dsn := pc.DSN()
	if dsn == "" {
		t.Error("StartPostgres: DSN should be non-empty")
	}
	if !strings.HasPrefix(dsn, "postgres://") {
		t.Errorf("StartPostgres: DSN should start with postgres://, got %q", dsn)
	}
	// cleanup is callable and should not panic
	cleanup()
	// Terminate is safe to call again on already-terminated container
	ctx := context.Background()
	if err := pc.Terminate(ctx); err != nil {
		t.Logf("Terminate after cleanup: %v (may be ok)", err)
	}
}
