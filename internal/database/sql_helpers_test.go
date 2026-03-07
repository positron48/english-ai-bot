package database

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestGetDialect(t *testing.T) {
	t.Run("nil_db_returns_postgres", func(t *testing.T) {
		got := GetDialect(nil)
		if got != DialectPostgres {
			t.Errorf("GetDialect(nil) = %q, want %q", got, DialectPostgres)
		}
	})

	dsn := startTestPostgres(t)
	registerPostgresCompatDriver()

	t.Run("unregistered_db_returns_postgres", func(t *testing.T) {
		conn, err := sql.Open("postgres_compat", dsn)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		defer conn.Close()
		got := GetDialect(conn)
		if got != DialectPostgres {
			t.Errorf("GetDialect(unregistered) = %q, want %q", got, DialectPostgres)
		}
	})

	t.Run("registered_db_returns_stored_dialect", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
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
			t.Fatalf("NewWithConfig: %v", err)
		}
		defer db.Close()
		got := GetDialect(db.GetConnection())
		if got != DialectPostgres {
			t.Errorf("GetDialect(registered) = %q, want %q", got, DialectPostgres)
		}
	})

	t.Run("stored_empty_string_returns_postgres", func(t *testing.T) {
		conn, err := sql.Open("postgres_compat", dsn)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		defer conn.Close()
		connDialects.Store(conn, "")
		got := GetDialect(conn)
		if got != DialectPostgres {
			t.Errorf("GetDialect(empty stored) = %q, want %q", got, DialectPostgres)
		}
	})

	t.Run("stored_custom_dialect_returns_it", func(t *testing.T) {
		conn, err := sql.Open("postgres_compat", dsn)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		defer conn.Close()
		connDialects.Store(conn, "mysql")
		got := GetDialect(conn)
		if got != "mysql" {
			t.Errorf("GetDialect(custom) = %q, want mysql", got)
		}
	})

	t.Run("registerConnDialect_nil_does_not_store", func(t *testing.T) {
		registerConnDialect(nil, "custom")
		got := GetDialect(nil)
		if got != DialectPostgres {
			t.Errorf("GetDialect(nil) after registerConnDialect(nil) = %q, want %q", got, DialectPostgres)
		}
	})

	t.Run("stored_non_string_value_returns_postgres", func(t *testing.T) {
		conn, err := sql.Open("postgres_compat", dsn)
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		defer conn.Close()
		connDialects.Store(conn, 42)
		got := GetDialect(conn)
		if got != DialectPostgres {
			t.Errorf("GetDialect(conn with non-string stored) = %q, want %q", got, DialectPostgres)
		}
	})
}

func TestInsertAndReturnID(t *testing.T) {
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
		t.Fatalf("NewWithConfig: %v", err)
	}
	defer db.Close()
	conn := db.GetConnection()

	ctx := context.Background()
	_, err = conn.ExecContext(ctx, `
		CREATE TEMPORARY TABLE insert_return_id_test (
			id   SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	t.Run("with_returning_appends_nothing", func(t *testing.T) {
		id, err := InsertAndReturnID(conn, `INSERT INTO insert_return_id_test (name) VALUES ($1) RETURNING id`, "alice")
		if err != nil {
			t.Fatalf("InsertAndReturnID: %v", err)
		}
		if id <= 0 {
			t.Errorf("InsertAndReturnID returned id = %d", id)
		}
	})

	t.Run("without_returning_appends_returning_id", func(t *testing.T) {
		id, err := InsertAndReturnID(conn, `INSERT INTO insert_return_id_test (name) VALUES ($1)`, "bob")
		if err != nil {
			t.Fatalf("InsertAndReturnID: %v", err)
		}
		if id <= 0 {
			t.Errorf("InsertAndReturnID returned id = %d", id)
		}
	})

	t.Run("invalid_query_returns_error", func(t *testing.T) {
		_, err := InsertAndReturnID(conn, "INSERT INTO nonexistent_table (x) VALUES (1)")
		if err == nil {
			t.Error("expected error for invalid query")
		}
		if !strings.Contains(err.Error(), "failed to insert") {
			t.Errorf("error message should mention insert: %v", err)
		}
	})
}
