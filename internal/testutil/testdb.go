package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"tgbot-skeleton/internal/database"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

// sharedPostgres holds one container per test package; reused across tests.
// Each test gets a clean schema via TRUNCATE.
// containerRef keeps the Postgres container alive so it is not GC'd (which would close the DB).
var sharedPostgres struct {
	once         sync.Once
	dsn          string
	db           *database.DB
	err          error
	containerRef interface{} // keep container reference so it is not garbage-collected
}

var truncateTables = []string{
	"grammar_placement_test", "grammar_progress", "grammar_test_attempts",
	"user_access_user_categories", "user_access_category_permissions", "user_access_categories",
	"web_otps", "web_sessions", "review_events", "training_sessions", "user_cards",
	"training_cards", "word_set_items", "word_sets", "word_set_categories",
	"user_word_mastering", "user_word_knowledge", "word_request_history", "word_forms", "word_cards",
	"grammar_published_items", "training_nudges", "users",
	"app_settings", "circuit_breaker_state",
	"tts_generation_status",
}
var circuitBreakerInit = `INSERT INTO circuit_breaker_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING`
var appSettingsInit = `INSERT INTO app_settings (key, value) VALUES ('hide_placement_test_button', 'false') ON CONFLICT (key) DO NOTHING`

func getSharedDB(t *testing.T) *database.DB {
	t.Helper()
	sharedPostgres.once.Do(func() {
		ctx := context.Background()
		ctr, err := postgres.Run(ctx, "postgres:16-alpine",
			postgres.WithDatabase("english_test"),
			postgres.WithUsername("english"),
			postgres.WithPassword("english"),
		)
		if err != nil {
			if isDockerUnavailable(err) {
				err = fmt.Errorf("docker unavailable: %w", err)
			}
			sharedPostgres.err = err
			return
		}
		dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			_ = ctr.Terminate(ctx)
			sharedPostgres.err = err
			return
		}
		logger, _ := zap.NewDevelopment()
		var db *database.DB
		for attempt := 0; attempt < 10; attempt++ {
			db, err = database.NewWithConfig("postgres", "", dsn, logger)
			if err == nil {
				break
			}
			if attempt < 9 {
				time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
			}
		}
		if err != nil {
			sharedPostgres.err = err
			return
		}
		sharedPostgres.dsn = dsn
		sharedPostgres.db = db
		sharedPostgres.containerRef = ctr // keep container alive so connection is not closed by GC
	})
	if sharedPostgres.err != nil {
		if strings.Contains(sharedPostgres.err.Error(), "docker unavailable") {
			t.Skipf("Docker unavailable: %v", sharedPostgres.err)
		}
		t.Fatalf("shared postgres: %v", sharedPostgres.err)
	}
	return sharedPostgres.db
}

func truncateAll(t *testing.T, conn *sql.DB) {
	t.Helper()
	// RESTART IDENTITY resets sequences so the next insert gets id=1
	for _, tbl := range truncateTables {
		_, _ = conn.Exec("TRUNCATE TABLE " + tbl + " RESTART IDENTITY CASCADE")
	}
	_, _ = conn.Exec(circuitBreakerInit)
	_, _ = conn.Exec(appSettingsInit)
}

// setupPostgresDB returns *database.DB with a clean schema (truncated).
func setupPostgresDB(t *testing.T) *database.DB {
	t.Helper()
	db := getSharedDB(t)
	conn := db.GetConnection()
	truncateAll(t, conn)
	return db
}

// GetTestDSN returns the DSN for the shared test Postgres (e.g. for bot tests that need config).
func GetTestDSN(t *testing.T) string {
	t.Helper()
	getSharedDB(t)
	return sharedPostgres.dsn
}

// SecondPostgresDSN starts a separate Postgres container and returns its DSN.
// Use when a test needs a connection that will be closed without affecting the shared test DB
// (e.g. to simulate "database is closed" for error paths). The container is cleaned up by testcontainers.
func SecondPostgresDSN(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	ctr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("english_test_second"),
		postgres.WithUsername("english"),
		postgres.WithPassword("english"),
	)
	if err != nil {
		if isDockerUnavailable(err) {
			t.Skipf("Docker unavailable: %v", err)
		}
		t.Fatalf("second postgres container: %v", err)
	}
	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = ctr.Terminate(ctx)
		t.Fatalf("second postgres DSN: %v", err)
	}
	return dsn
}

// SetupTestDB creates a Postgres database with all tables migrated.
// Uses one shared container per package; each test gets a clean schema via TRUNCATE.
// Do NOT call Close() on the returned connection—it is shared across tests.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return setupPostgresDB(t).GetConnection()
}

// SetupTestDatabase creates a Postgres database and returns the full *database.DB.
// Use when you need database.DB (e.g. for router tests). Call .GetConnection() for *sql.DB.
// Do NOT call Close() on the returned *database.DB—it is shared across tests in the package.
func SetupTestDatabase(t *testing.T) *database.DB {
	t.Helper()
	return setupPostgresDB(t)
}

func isDockerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	checks := []string{
		"cannot connect to the docker daemon",
		"could not connect to docker",
		"docker daemon",
		"connection refused",
		"no such host",
	}
	for _, c := range checks {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}
