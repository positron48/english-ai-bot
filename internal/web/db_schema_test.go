package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestRouter_handleDBSchema(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	t.Run("GET request returns schema", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/db-schema", nil)
		w := httptest.NewRecorder()

		router.handleDBSchema(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var schema SchemaResponse
		if err := json.NewDecoder(w.Body).Decode(&schema); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(schema.Tables) == 0 {
			t.Error("Expected at least one table in schema")
		}

		// Check that we have expected tables
		tableNames := make(map[string]bool)
		for _, table := range schema.Tables {
			tableNames[table.Name] = true
		}

		expectedTables := []string{"users", "word_cards", "training_cards", "user_cards"}
		for _, expectedTable := range expectedTables {
			if !tableNames[expectedTable] {
				t.Errorf("Expected table %q not found in schema", expectedTable)
			}
		}
	})

	t.Run("POST request returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/db-schema", nil)
		w := httptest.NewRecorder()

		router.handleDBSchema(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

func TestRouter_getDBSchema(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	schema, err := router.getDBSchema()
	if err != nil {
		t.Fatalf("getDBSchema() error = %v", err)
	}

	if schema == nil {
		t.Fatal("getDBSchema() should not return nil")
	}

	if len(schema.Tables) == 0 {
		t.Error("Expected at least one table in schema")
	}

	// Verify table structure
	for _, table := range schema.Tables {
		if table.Name == "" {
			t.Error("Table name should not be empty")
		}
		if len(table.Columns) == 0 {
			t.Errorf("Table %q should have at least one column", table.Name)
		}

		// Check that columns have required fields
		for _, col := range table.Columns {
			if col.Name == "" {
				t.Error("Column name should not be empty")
			}
			if col.Type == "" {
				t.Error("Column type should not be empty")
			}
		}
	}
}

func TestRouter_getTableNames(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	tables, err := router.getTableNames()
	if err != nil {
		t.Fatalf("getTableNames() error = %v", err)
	}

	if len(tables) == 0 {
		t.Error("Expected at least one table")
	}

	// Verify that system tables are excluded
	for _, table := range tables {
		if table == "sqlite_sequence" || table == "sqlite_master" {
			t.Errorf("System table %q should be excluded", table)
		}
	}

	// Verify expected tables are present
	tableMap := make(map[string]bool)
	for _, table := range tables {
		tableMap[table] = true
	}

	expectedTables := []string{"users", "word_cards", "training_cards", "user_cards"}
	for _, expectedTable := range expectedTables {
		if !tableMap[expectedTable] {
			t.Errorf("Expected table %q not found", expectedTable)
		}
	}
}

func TestRouter_getTableColumns(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	t.Run("Get columns for existing table", func(t *testing.T) {
		columns, err := router.getTableColumns("users")
		if err != nil {
			t.Fatalf("getTableColumns() error = %v", err)
		}

		if len(columns) == 0 {
			t.Error("Expected at least one column")
		}

		// Verify column structure
		columnNames := make(map[string]bool)
		for _, col := range columns {
			columnNames[col.Name] = true
			if col.Name == "" {
				t.Error("Column name should not be empty")
			}
			if col.Type == "" {
				t.Error("Column type should not be empty")
			}
		}

		// Check for expected columns
		expectedColumns := []string{"id", "telegram_id"}
		for _, expectedCol := range expectedColumns {
			if !columnNames[expectedCol] {
				t.Errorf("Expected column %q not found", expectedCol)
			}
		}
	})

	t.Run("Get columns for invalid table name", func(t *testing.T) {
		_, err := router.getTableColumns("'; DROP TABLE users; --")
		if err == nil {
			t.Error("getTableColumns() should return error for invalid table name")
		}
	})

	t.Run("Get columns for non-existent table", func(t *testing.T) {
		columns, err := router.getTableColumns("non_existent_table")
		if err != nil {
			t.Fatalf("getTableColumns() error = %v", err)
		}
		if len(columns) != 0 {
			t.Errorf("Expected 0 columns for non-existent table, got %d", len(columns))
		}
	})
}

func TestRouter_getForeignKeys(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	defer db.Close()

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)

	t.Run("Get foreign keys for table with FKs", func(t *testing.T) {
		foreignKeys, err := router.getForeignKeys("word_forms")
		if err != nil {
			t.Fatalf("getForeignKeys() error = %v", err)
		}

		// word_forms should have a foreign key to word_cards
		if len(foreignKeys) == 0 {
			t.Error("Expected at least one foreign key for word_forms table")
		}

		found := false
		for _, fk := range foreignKeys {
			if fk.FromTable == "word_forms" && fk.ToTable == "word_cards" {
				found = true
				if fk.FromColumn == "" {
					t.Error("FromColumn should not be empty")
				}
				if fk.ToColumn == "" {
					t.Error("ToColumn should not be empty")
				}
			}
		}

		if !found {
			t.Error("Expected foreign key from word_forms to word_cards not found")
		}
	})

	t.Run("Get foreign keys for table without FKs", func(t *testing.T) {
		foreignKeys, err := router.getForeignKeys("users")
		if err != nil {
			t.Fatalf("getForeignKeys() error = %v", err)
		}

		// users table might not have foreign keys
		// Just verify the function doesn't error
		_ = foreignKeys
	})

	t.Run("Get foreign keys for invalid table name", func(t *testing.T) {
		_, err := router.getForeignKeys("'; DROP TABLE users; --")
		if err == nil {
			t.Error("getForeignKeys() should return error for invalid table name")
		}
	})
}

func TestRouter_isValidTableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid table name",
			input:    "users",
			expected: true,
		},
		{
			name:     "valid table name with underscore",
			input:    "word_cards",
			expected: true,
		},
		{
			name:     "valid table name with numbers",
			input:    "table123",
			expected: true,
		},
		{
			name:     "invalid - SQL injection attempt",
			input:    "'; DROP TABLE users; --",
			expected: false,
		},
		{
			name:     "invalid - starts with number",
			input:    "123table",
			expected: false,
		},
		{
			name:     "invalid - contains spaces",
			input:    "word cards",
			expected: false,
		},
		{
			name:     "invalid - contains special characters",
			input:    "word-cards",
			expected: false,
		},
		{
			name:     "invalid - empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid - only spaces",
			input:    "   ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTableName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidTableName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
