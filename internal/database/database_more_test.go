package database

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewWithConfig_More(t *testing.T) {
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
	defer db.Close()

	if db == nil {
		t.Error("Expected non-nil database")
	}

	if db.GetConnection() == nil {
		t.Error("Expected non-nil connection")
	}
}

func TestNewWithConfig_MultipleInstances(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	dsn := startTestPostgres(t)

	var db1, db2 *DB
	var err error
	for attempt := 0; attempt < 10; attempt++ {
		db1, err = NewWithConfig("postgres", "", dsn, logger)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("NewWithConfig() db1 error = %v", err)
	}
	defer db1.Close()

	db2, err = NewWithConfig("postgres", "", dsn, logger)
	if err != nil {
		t.Fatalf("NewWithConfig() db2 error = %v", err)
	}
	defer db2.Close()

	if db1.GetConnection() == db2.GetConnection() {
		t.Error("Expected different connection instances")
	}
}

func TestDatabase_Close(t *testing.T) {
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

	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestDatabase_CoreTables_Exist(t *testing.T) {
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
	defer db.Close()

	tables := []string{
		"users",
		"word_cards",
		"training_cards",
		"user_cards",
		"training_sessions",
		"review_events",
		"circuit_breaker_state",
	}

	conn := db.GetConnection()
	for _, table := range tables {
		var count int
		err := conn.QueryRow(
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=$1",
			table,
		).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
			continue
		}
		if count == 0 {
			t.Errorf("Table %s does not exist", table)
		}
	}
}

func TestDatabase_LinglowCourseArchitectureSeeded(t *testing.T) {
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
	defer db.Close()

	conn := db.GetConnection()
	expectedTables := []string{
		"courses",
		"user_courses",
		"districts",
		"locations",
		"theme_lines",
		"modules",
		"learning_objectives",
		"learning_items",
		"srs_items",
		"exercise_attempts",
		"learning_events",
		"daily_course_stats",
		"mode_daily_stats",
		"district_progress",
		"learning_item_stats",
		"content_performance_stats",
	}
	for _, table := range expectedTables {
		var count int
		err := conn.QueryRow(
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=$1",
			table,
		).Scan(&count)
		if err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if count == 0 {
			t.Fatalf("expected table %s to exist", table)
		}
	}

	checkCount := func(query string, want int) {
		t.Helper()
		var got int
		if err := conn.QueryRow(query).Scan(&got); err != nil {
			t.Fatalf("count query failed: %v", err)
		}
		if got != want {
			t.Fatalf("count query got %d, want %d: %s", got, want, query)
		}
	}
	checkCount(`SELECT COUNT(*) FROM courses WHERE code IN ('en_ru', 'es_ru')`, 2)
	checkCount(`SELECT COUNT(*) FROM districts`, 12)
	checkCount(`SELECT COUNT(*) FROM locations`, 72)
	checkCount(`SELECT COUNT(*) FROM theme_lines`, 8)
}

// TestMigratePostgres_ErrorWhenConnClosed covers migratePostgres error branch when Exec fails (e.g. closed connection).
func TestMigratePostgres_ErrorWhenConnClosed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	dsn := startTestPostgres(t)

	var conn *sql.DB
	var err error
	for attempt := 0; attempt < 10; attempt++ {
		conn, err = openPostgresDB(dsn)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("openPostgresDB: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("conn.Close: %v", err)
	}

	db := &DB{conn: conn, logger: logger}
	err = db.migratePostgres()
	if err == nil {
		t.Fatal("migratePostgres() expected error when conn is closed")
	}
	if !strings.Contains(err.Error(), "failed to execute postgres migration") {
		t.Errorf("error should mention failed to execute postgres migration: %v", err)
	}
}

// TestNewWithConfig_MigrateFails covers the branch where openPostgresDB succeeds but migratePostgres fails (conn.Close + return error).
func TestNewWithConfig_MigrateFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	dsn := startTestPostgres(t)

	var conn *sql.DB
	var err error
	for attempt := 0; attempt < 10; attempt++ {
		conn, err = openPostgresDB(dsn)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("openPostgresDB: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("conn.Close: %v", err)
	}

	orig := openPostgresDBFunc
	openPostgresDBFunc = func(string) (*sql.DB, error) { return conn, nil }
	defer func() { openPostgresDBFunc = orig }()

	_, err = NewWithConfig("postgres", "", dsn, logger)
	if err == nil {
		t.Fatal("NewWithConfig expected error when migration fails")
	}
	if !strings.Contains(err.Error(), "failed to migrate database") {
		t.Errorf("error should mention failed to migrate database: %v", err)
	}
}
