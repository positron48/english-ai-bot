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
}

func TestNewWithConfig_EmptyURL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewWithConfig("postgres", "", "", logger)
	if err == nil {
		t.Error("expected error for empty URL")
	}
}
