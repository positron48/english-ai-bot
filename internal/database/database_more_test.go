package database

import (
	"testing"

	"go.uber.org/zap"
)

func TestNew_InMemory_More(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Error("Expected non-nil database")
	}

	if db.GetConnection() == nil {
		t.Error("Expected non-nil connection")
	}
}

func TestNew_MultipleInstances(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db1, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() db1 error = %v", err)
	}
	defer db1.Close()

	db2, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() db2 error = %v", err)
	}
	defer db2.Close()

	// Each database should be independent
	if db1.GetConnection() == db2.GetConnection() {
		t.Error("Expected different connections for different databases")
	}
}

func TestMigrate_IdempotentMigrations(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Run migrations again - should be idempotent
	err = db.migrate()
	if err != nil {
		t.Errorf("Second migrate() should be idempotent, got error = %v", err)
	}
}

func TestDatabase_Close(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestDatabase_CoreTables_Exist(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Only check core tables that always exist
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
		err := conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check table %s: %v", table, err)
			continue
		}
		if count == 0 {
			t.Errorf("Table %s does not exist", table)
		}
	}
}

func TestMigrate_ColumnMigrations(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	// Check that important columns exist in word_cards
	conn := db.GetConnection()
	
	columns := []string{"processed_at", "processing_error"}
	for _, col := range columns {
		var count int
		err := conn.QueryRow("SELECT COUNT(*) FROM pragma_table_info('word_cards') WHERE name=?", col).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check column %s: %v", col, err)
			continue
		}
		if count == 0 {
			t.Errorf("Column %s does not exist in word_cards", col)
		}
	}
}

func TestMigrate_TrainingCardsColumns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()

	conn := db.GetConnection()
	
	// Only check core columns
	columns := []string{"word_en", "word_ru", "meaning_en"}
	for _, col := range columns {
		var count int
		err := conn.QueryRow("SELECT COUNT(*) FROM pragma_table_info('training_cards') WHERE name=?", col).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check column %s: %v", col, err)
			continue
		}
		if count == 0 {
			t.Errorf("Column %s does not exist in training_cards", col)
		}
	}
}
