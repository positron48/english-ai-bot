package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"tgbot-skeleton/internal/database"

	"github.com/jackc/pgx/v5"
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
		sharedPostgres.dsn = normalizeDSNPort(dsn)
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

// sharedSecondContainer holds one Postgres container reused by all SecondPostgresDSN calls.
// Each call creates a fresh database inside this container instead of starting a new container.
var sharedSecondContainer struct {
	once         sync.Once
	host         string // base DSN (pointing to init db) of the shared container
	err          error
	containerRef interface{}
}

// secondDBCounter is used to generate unique database names.
var secondDBCounter atomic.Int64

// SecondPostgresDSN returns a DSN for a fresh isolated Postgres database.
// All calls share one container; each call creates a new database inside it.
// Use when a test needs a connection that will be closed / schema that will be altered
// without affecting the shared test DB.
func SecondPostgresDSN(t *testing.T) string {
	t.Helper()

	// Ensure the shared second container is started.
	sharedSecondContainer.once.Do(func() {
		ctx := context.Background()
		ctr, err := postgres.Run(ctx, "postgres:16-alpine",
			postgres.WithDatabase("english_test_second_init"),
			postgres.WithUsername("english"),
			postgres.WithPassword("english"),
		)
		if err != nil {
			if isDockerUnavailable(err) {
				err = fmt.Errorf("docker unavailable: %w", err)
			}
			sharedSecondContainer.err = err
			return
		}
		dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			_ = ctr.Terminate(ctx)
			sharedSecondContainer.err = err
			return
		}
		sharedSecondContainer.host = normalizeDSNPort(dsn)
		sharedSecondContainer.containerRef = ctr
	})

	if sharedSecondContainer.err != nil {
		if strings.Contains(sharedSecondContainer.err.Error(), "docker unavailable") {
			t.Skipf("Docker unavailable: %v", sharedSecondContainer.err)
		}
		t.Fatalf("second postgres container: %v", sharedSecondContainer.err)
	}

	// Create a unique database for this test inside the shared container.
	n := secondDBCounter.Add(1)
	dbName := fmt.Sprintf("english_test_iso_%d", n)

	// Use a short-lived pgx connection for CREATE DATABASE (no migration, no Ryuk interference).
	// Retry to handle the brief window when the container just started.
	adminDSN := replaceDBName(sharedSecondContainer.host, "english_test_second_init")
	var pgxErr error
	for i := range 30 {
		var conn *pgx.Conn
		conn, pgxErr = pgx.Connect(context.Background(), adminDSN)
		if pgxErr == nil {
			_, pgxErr = conn.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbName))
			_ = conn.Close(context.Background())
			if pgxErr == nil {
				break
			}
		}
		sleep := time.Duration(i+1) * 100 * time.Millisecond
		if sleep > 500*time.Millisecond {
			sleep = 500 * time.Millisecond
		}
		time.Sleep(sleep)
	}
	if pgxErr != nil {
		t.Fatalf("SecondPostgresDSN: create database %s: %v", dbName, pgxErr)
	}

	return replaceDBName(sharedSecondContainer.host, dbName)
}

// replaceDBName swaps the database name in a postgres DSN.
// DSN format: postgres://user:pass@host:port/dbname?params
func replaceDBName(dsn, newDB string) string {
	// Find the last '/' before '?' and replace the db name.
	qIdx := strings.Index(dsn, "?")
	base := dsn
	params := ""
	if qIdx != -1 {
		base = dsn[:qIdx]
		params = dsn[qIdx:]
	}
	slashIdx := strings.LastIndex(base, "/")
	if slashIdx == -1 {
		return dsn
	}
	return base[:slashIdx+1] + newDB + params
}

// normalizeDSNPort strips /tcp and /udp from port in DSN (e.g. port=5432/tcp -> port=5432).
func normalizeDSNPort(dsn string) string {
	return portProtoRegex.ReplaceAllString(dsn, "$1")
}

var portProtoRegex = regexp.MustCompile(`(\d+)/(tcp|udp)`)

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
