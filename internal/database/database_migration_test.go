package database

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestNew_CreatesDatabase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	db, err := New(dbPath, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestNew_InMemory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Verify connection works
	conn := db.GetConnection()
	if conn == nil {
		t.Error("GetConnection() returned nil")
	}
	
	// Test we can query
	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if count == 0 {
		t.Error("Expected tables to be created")
	}
}

func TestNew_MigrationCreatesAllTables(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// List of expected tables
	expectedTables := []string{
		"word_cards",
		"word_forms",
		"word_request_history",
		"users",
		"training_cards",
		"user_cards",
		"training_sessions",
		"review_events",
		"training_nudges",
		"circuit_breaker_state",
		"web_sessions",
		"web_otps",
		"word_set_categories",
		"word_sets",
		"word_set_items",
		"user_word_knowledge",
	}
	
	conn := db.GetConnection()
	for _, tableName := range expectedTables {
		var count int
		err = conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
		if err != nil {
			t.Fatalf("Query for table %s failed: %v", tableName, err)
		}
		if count == 0 {
			t.Errorf("Table %s was not created", tableName)
		}
	}
}

func TestNew_CircuitBreakerStateInitialized(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	conn := db.GetConnection()
	
	var id, isOpen int
	err = conn.QueryRow("SELECT id, is_open FROM circuit_breaker_state WHERE id = 1").Scan(&id, &isOpen)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if id != 1 {
		t.Error("Expected circuit breaker state with id=1")
	}
	if isOpen != 0 {
		t.Error("Expected circuit breaker to be closed by default")
	}
}

func TestClose(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	
	err = db.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
	
	// Try to query - should fail
	conn := db.GetConnection()
	var count int
	err = conn.QueryRow("SELECT 1").Scan(&count)
	if err == nil {
		t.Error("Expected error after Close()")
	}
}

func TestNew_ExistingDatabase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	// Create first time
	db1, err := New(dbPath, logger)
	if err != nil {
		t.Fatalf("First New() error = %v", err)
	}
	
	// Insert test data
	_, err = db1.GetConnection().Exec("INSERT INTO users (telegram_id) VALUES (12345)")
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	db1.Close()
	
	// Open again
	db2, err := New(dbPath, logger)
	if err != nil {
		t.Fatalf("Second New() error = %v", err)
	}
	defer db2.Close()
	
	// Check data persists
	var telegramID int64
	err = db2.GetConnection().QueryRow("SELECT telegram_id FROM users WHERE telegram_id = 12345").Scan(&telegramID)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if telegramID != 12345 {
		t.Errorf("Expected telegram_id 12345, got %d", telegramID)
	}
}

func TestMigrateUsersTable_AddsColumn(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check telegram_username column exists
	var count int
	err = db.GetConnection().QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') 
		WHERE name='telegram_username'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if count == 0 {
		t.Error("Expected telegram_username column in users table")
	}
}

func TestMigrateWordSetsTable_AddsSortOrder(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check sort_order column exists
	var count int
	err = db.GetConnection().QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('word_sets') 
		WHERE name='sort_order'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if count == 0 {
		t.Error("Expected sort_order column in word_sets table")
	}
}

func TestMigrateWordSetsTable_AddsPreferredPOS(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check preferred_pos column exists
	var count int
	err = db.GetConnection().QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('word_sets') 
		WHERE name='preferred_pos'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if count == 0 {
		t.Error("Expected preferred_pos column in word_sets table")
	}
}

func TestMigrate_CreatesIndexes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check for some key indexes
	expectedIndexes := []string{
		"idx_word_cards_word",
		"idx_user_cards_user_id",
		"idx_training_cards_word_card_id",
		"idx_web_sessions_token",
	}
	
	conn := db.GetConnection()
	for _, indexName := range expectedIndexes {
		var count int
		err = conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&count)
		if err != nil {
			t.Fatalf("Query for index %s failed: %v", indexName, err)
		}
		if count == 0 {
			t.Errorf("Index %s was not created", indexName)
		}
	}
}

func TestMigrateWordCardsLemmaColumns_All(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check lemma-related columns exist
	expectedColumns := []string{
		"pos",
		"transcription",
		"definition_ru",
		"examples_json",
		"verb_forms_json",
		"display_en",
	}
	
	conn := db.GetConnection()
	for _, colName := range expectedColumns {
		var count int
		err = conn.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('word_cards') 
			WHERE name=?
		`, colName).Scan(&count)
		if err != nil {
			t.Fatalf("Query for column %s failed: %v", colName, err)
		}
		if count == 0 {
			t.Errorf("Column %s was not created in word_cards", colName)
		}
	}
}

func TestMigrateTrainingCardsDisplayColumns_All(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check display columns exist
	expectedColumns := []string{"pos", "display_word"}
	
	conn := db.GetConnection()
	for _, colName := range expectedColumns {
		var count int
		err = conn.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('training_cards') 
			WHERE name=?
		`, colName).Scan(&count)
		if err != nil {
			t.Fatalf("Query for column %s failed: %v", colName, err)
		}
		if count == 0 {
			t.Errorf("Column %s was not created in training_cards", colName)
		}
	}
}

func TestMigrateWordRequestHistory_All(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check word_card_id and input_word columns exist
	expectedColumns := []string{"word_card_id", "input_word"}
	
	conn := db.GetConnection()
	for _, colName := range expectedColumns {
		var count int
		err = conn.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('word_request_history') 
			WHERE name=?
		`, colName).Scan(&count)
		if err != nil {
			t.Fatalf("Query for column %s failed: %v", colName, err)
		}
		if count == 0 {
			t.Errorf("Column %s was not created in word_request_history", colName)
		}
	}
}

func TestMigrateWordCardsProcessingColumns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check processed_at and processing_error columns exist
	expectedColumns := []string{"processed_at", "processing_error"}
	
	conn := db.GetConnection()
	for _, colName := range expectedColumns {
		var count int
		err = conn.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('word_cards') 
			WHERE name=?
		`, colName).Scan(&count)
		if err != nil {
			t.Fatalf("Query for column %s failed: %v", colName, err)
		}
		if count == 0 {
			t.Errorf("Column %s was not created in word_cards", colName)
		}
	}
}

func TestMigrateWordFormsTable(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	db, err := New(":memory:", logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer db.Close()
	
	// Check word_forms table exists and has correct columns
	var count int
	err = db.GetConnection().QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='word_forms'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	if count == 0 {
		t.Error("word_forms table was not created")
	}
	
	// Check columns
	expectedColumns := []string{"form", "word_card_id"}
	conn := db.GetConnection()
	for _, colName := range expectedColumns {
		err = conn.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('word_forms') 
			WHERE name=?
		`, colName).Scan(&count)
		if err != nil {
			t.Fatalf("Query for column %s failed: %v", colName, err)
		}
		if count == 0 {
			t.Errorf("Column %s was not created in word_forms", colName)
		}
	}
}
