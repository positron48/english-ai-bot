package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

// failingResponseWriter implements http.ResponseWriter and fails on Write to trigger encode error path.
type failingResponseWriter struct {
	http.ResponseWriter
	writeCalls int
}

func (w *failingResponseWriter) Write(p []byte) (n int, err error) {
	w.writeCalls++
	if w.writeCalls > 0 {
		return 0, errors.New("write failed")
	}
	return w.ResponseWriter.Write(p)
}

func TestRouter_handleDBSchema(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

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

		// DBQueryAccess should reflect config (default false when not set)
		if schema.DBQueryAccess != router.config.Admin.DBQueryAccess {
			t.Errorf("schema.DBQueryAccess = %v, want %v", schema.DBQueryAccess, router.config.Admin.DBQueryAccess)
		}
	})

	t.Run("GET request returns schema with DBQueryAccess true when configured", func(t *testing.T) {
		cfg := &config.Config{
			WebApp: config.WebAppConfig{JWTSecret: "test-secret"},
			Admin:  config.AdminConfig{DBQueryAccess: true},
		}
		router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
		req := httptest.NewRequest("GET", "/admin/db-schema", nil)
		w := httptest.NewRecorder()
		router.handleDBSchema(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", w.Code)
		}
		var schema SchemaResponse
		if err := json.NewDecoder(w.Body).Decode(&schema); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if !schema.DBQueryAccess {
			t.Error("Expected DBQueryAccess true when config.Admin.DBQueryAccess is true")
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

	t.Run("encode error returns 500", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/db-schema", nil)
		rec := httptest.NewRecorder()
		w := &failingResponseWriter{ResponseWriter: rec}

		router.handleDBSchema(w, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 on encode failure, got %d", rec.Code)
		}
	})
}

func TestRouter_handleDBSchema_Get_DBFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/admin/db-schema", nil)
	w := httptest.NewRecorder()
	router.handleDBSchema(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 when getDBSchema fails, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRouter_getDBSchema_DBFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, nil)

	_, err := router.getDBSchema()
	if err == nil {
		t.Error("getDBSchema() should return error when DB fails")
	}
}

func TestRouter_getDBSchema(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

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

	// Verify expected tables are present (Postgres information_schema already excludes system tables)
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

func TestHandleDBQuery_Select(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"SELECT 1 as num"}`))
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload dbQueryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(payload.Rows) != 1 || payload.Rows[0]["num"] != float64(1) {
		t.Fatalf("expected row with num=1, got %+v", payload.Rows)
	}
}

func TestHandleDBQuery_Exec(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"INSERT INTO users (telegram_id, created_at, updated_at) VALUES (123, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"}`))
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload dbQueryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if payload.RowsAffected != 1 || payload.Message != "OK" {
		t.Fatalf("expected rows_affected=1 and OK, got %+v", payload)
	}
}

func TestHandleDBQuery_Disabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = false

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"SELECT 1"}`))
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHandleDBQuery_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/db-query", nil)
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleDBQuery_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDBQuery_EmptyQuery(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDBQuery_ForbiddenSQL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"DROP TABLE users"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if body := w.Body.String(); body != "" && !strings.Contains(body, "Forbidden") && !strings.Contains(body, "forbidden") {
		t.Logf("response body: %s", body)
	}
}

func TestHandleDBQuery_QueryTooLong(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	// maxQueryLen is 50000
	longQuery := strings.Repeat("x", 50001)
	body, _ := json.Marshal(map[string]string{"query": longQuery})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for query too long, got %d", w.Code)
	}
}

func TestHandleDBQuery_MultipleStatements(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"SELECT 1; SELECT 2"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for multiple statements, got %d: %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); !strings.Contains(body, "single") {
		t.Errorf("expected body to mention single statement, got %s", body)
	}
}

func TestHandleDBQuery_SelectFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"SELECT 1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when query fails, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDBQuery_ExecFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"INSERT INTO users (telegram_id, created_at, updated_at) VALUES (1, NOW(), NOW())"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when exec fails, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDBQuery_WithSelect(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	// WITH (CTE) is treated as SELECT
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"WITH t AS (SELECT 1 AS n) SELECT * FROM t"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleDBQuery(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var payload dbQueryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload.Rows) != 1 || payload.Rows[0]["n"] != float64(1) {
		t.Fatalf("expected one row with n=1, got %+v", payload.Rows)
	}
}

func TestHandleDBQuery_EncodeFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewBufferString(`{"query":"SELECT 1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	w := &failingResponseWriter{ResponseWriter: rec}
	router.handleDBQuery(w, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when encode fails, got %d", rec.Code)
	}
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
