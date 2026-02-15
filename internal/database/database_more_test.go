package database

import (
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
