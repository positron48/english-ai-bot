package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

func startTestPostgres(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	ctr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("english_test"),
		postgres.WithUsername("english"),
		postgres.WithPassword("english"),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "docker") {
			t.Skipf("Docker unavailable: %v", err)
		}
		t.Fatalf("failed to start postgres: %v", err)
	}
	t.Cleanup(func() { _ = ctr.Terminate(ctx) })

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get DSN: %v", err)
	}
	return dsn
}

func TestNewWithConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	dsn := startTestPostgres(t)

	var db *DB
	var err error
	for attempt := 0; attempt < 10; attempt++ {
		db, err = NewWithConfig("postgres", "", dsn, logger)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("NewWithConfig() error = %v", err)
	}
	if db == nil {
		t.Fatal("NewWithConfig() should not return nil")
	}

	conn := db.GetConnection()
	if conn == nil {
		t.Fatal("GetConnection() should not return nil")
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestNewWithConfig_RejectsUnsupportedDriver(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewWithConfig("mysql", "", "", logger)
	if err == nil {
		t.Error("expected error for unsupported driver")
	}
	if err != nil && !strings.Contains(err.Error(), "only postgres") {
		t.Errorf("error should mention only postgres: %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "mysql") {
		t.Errorf("error should mention received driver: %v", err)
	}
}

func TestNewWithConfig_EmptyURL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewWithConfig("postgres", "", "", logger)
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestNewWithConfig_InvalidDSN(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	// Invalid DSN so openPostgresDB fails (open or ping).
	_, err := NewWithConfig("postgres", "", "postgres://invalidhost:5432/db?connect_timeout=1", logger)
	if err == nil {
		t.Error("expected error for invalid DSN")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to open") && !strings.Contains(err.Error(), "failed to ping") {
		t.Errorf("error should mention open or ping failure: %v", err)
	}
}

func TestNewWithConfig_DriverTrimSpace_StillRequiresURL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	// Driver with leading/trailing space is accepted; empty URL still rejected.
	_, err := NewWithConfig("  postgres  ", "", "", logger)
	if err == nil {
		t.Error("expected error for empty URL even with trimmed driver")
	}
	if err != nil && !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("error should mention DATABASE_URL: %v", err)
	}
}

// TestOpenPostgresDB_InvalidDSN covers openPostgresDB error branches (open or ping failure).
func TestOpenPostgresDB_InvalidDSN(t *testing.T) {
	_, err := openPostgresDB("postgres://nonexistent.invalid:5432/db?connect_timeout=1")
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
	msg := err.Error()
	if !strings.Contains(msg, "failed to open") && !strings.Contains(msg, "failed to ping") {
		t.Errorf("error should mention open or ping failure: %v", err)
	}
}

// TestOpenPostgresDB_UnreachableHost covers "failed to ping" branch (connection fails on first use).
func TestOpenPostgresDB_UnreachableHost(t *testing.T) {
	_, err := openPostgresDB("postgres://user:pass@192.0.2.1:5432/db?connect_timeout=1")
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
	if !strings.Contains(err.Error(), "failed to ping") {
		t.Errorf("error should mention failed to ping: %v", err)
	}
}

// TestOpenPostgresDB_EmptyDSN covers "failed to open" branch when driver rejects invalid DSN.
func TestOpenPostgresDB_EmptyDSN(t *testing.T) {
	_, err := openPostgresDB("")
	if err == nil {
		t.Fatal("expected error for empty DSN")
	}
	if !strings.Contains(err.Error(), "failed to open") && !strings.Contains(err.Error(), "failed to ping") {
		t.Errorf("error should mention failed to open or failed to ping: %v", err)
	}
}

// TestOpenPostgresDB_OpenFails covers "failed to open" branch when sql.Open returns error (e.g. unknown driver).
func TestOpenPostgresDB_OpenFails(t *testing.T) {
	orig := openPostgresDBDriverName
	openPostgresDBDriverName = "postgres_compat_nonexistent_driver"
	defer func() { openPostgresDBDriverName = orig }()

	_, err := openPostgresDB("postgres://localhost/db")
	if err == nil {
		t.Fatal("expected error when driver is not registered")
	}
	if !strings.Contains(err.Error(), "failed to open") {
		t.Errorf("error should mention failed to open: %v", err)
	}
}

func TestNewWithConfig_EmptyURL_WhitespaceOnly(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewWithConfig("postgres", "", "   \t  ", logger)
	if err == nil {
		t.Error("expected error for whitespace-only URL")
	}
	if err != nil && !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("error should mention DATABASE_URL: %v", err)
	}
}

func TestNewWithConfig_DriverCaseInsensitive(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	dsn := startTestPostgres(t)

	var db *DB
	var err error
	for attempt := 0; attempt < 10; attempt++ {
		db, err = NewWithConfig("POSTGRES", "", dsn, logger)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("NewWithConfig(POSTGRES) error = %v", err)
	}
	if db == nil {
		t.Fatal("NewWithConfig(POSTGRES) should succeed")
	}
	_ = db.Close()
}
