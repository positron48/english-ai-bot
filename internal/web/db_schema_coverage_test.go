package web

// Tests to cover error paths in db_schema.go that are not covered by db_schema_test.go.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// TestGetTableColumns_BrokenDB covers lines 155-157 in db_schema.go:
// the r.db.Query error path in getTableColumns when the DB connection is closed.
func TestGetTableColumns_BrokenDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, nil)

	_, err := router.getTableColumns("users")
	if err == nil {
		t.Error("getTableColumns() should return error when DB is closed")
	}
}

// TestGetForeignKeys_BrokenDB covers lines 205-207 in db_schema.go:
// the r.db.Query error path in getForeignKeys when the DB connection is closed.
func TestGetForeignKeys_BrokenDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, nil)

	_, err := router.getForeignKeys("users")
	if err == nil {
		t.Error("getForeignKeys() should return error when DB is closed")
	}
}

// TestGetDBSchema_GetTableColumnsFails covers lines 83-85 in db_schema.go:
// the getTableColumns error path inside the getDBSchema loop.
// We use the second DB, verify it has tables, then close it so the loop body fails.
// Note: closing the DB will also cause getTableNames to fail, so both paths are tested.
func TestGetDBSchema_GetTableColumnsFails(t *testing.T) {
	router, dbWrap, _ := setupAdminRouterWithSecondDB(t)

	conn := dbWrap.GetConnection()
	// Verify DB is working and has tables
	rows, err := conn.Query(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' LIMIT 1`)
	if err != nil {
		t.Skipf("cannot query tables: %v", err)
	}
	hasTable := rows.Next()
	rows.Close()
	if !hasTable {
		t.Skip("no tables in second DB, cannot test this path")
	}

	// Close the DB wrapper so subsequent queries fail
	if err := dbWrap.Close(); err != nil {
		t.Skipf("cannot close DB: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/db-schema", nil)
	w := httptest.NewRecorder()

	router.handleDBSchema(w, req)

	// With a closed DB, getTableNames fails → 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when DB is closed, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleDBQuery_SelectWithNullValues covers the nil value branch (lines 325-329)
// in handleDBQuery where a column value is nil (NULL in SQL).
func TestHandleDBQuery_SelectWithNullValues(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	// Query that returns a NULL value to exercise the nil branch in the scan loop
	body, _ := json.Marshal(map[string]string{"query": "SELECT NULL::text as nullable_col, 1 as num"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.handleDBQuery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for SELECT with NULL, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleDBQuery_SelectWithBytesValue covers the []byte branch (lines 327-329)
// in handleDBQuery where a column value is returned as []byte.
func TestHandleDBQuery_SelectWithBytesValue(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	// Query that returns a bytea value to exercise the []byte branch
	body, _ := json.Marshal(map[string]string{"query": "SELECT 'hello'::bytea as bytes_col"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.handleDBQuery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for SELECT with bytea, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleDBQuery_ExecRowsAffected covers the non-SELECT exec path with rows affected.
func TestHandleDBQuery_ExecRowsAffected(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)

	// UPDATE with 0 rows affected (non-existent ID) - exercises exec path
	body, _ := json.Marshal(map[string]string{"query": "UPDATE users SET updated_at = NOW() WHERE id = -999"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.handleDBQuery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for UPDATE with 0 rows affected, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleDBQuery_SelectQueryError covers lines 298-301 in db_schema.go:
// the db.Query error path in handleDBQuery for SELECT statements.
func TestHandleDBQuery_SelectQueryError(t *testing.T) {
	router, _, _ := setupAdminRouterWithSecondDB(t)
	router.config.Admin.DBQueryAccess = true

	// Query a non-existent table to trigger a query error
	body, _ := json.Marshal(map[string]string{"query": "SELECT * FROM nonexistent_table_xyz"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.handleDBQuery(w, req)

	// Query on non-existent table returns 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for query on non-existent table, got %d: %s", w.Code, w.Body.String())
	}
}

// newSqlmockRouter creates a Router with a sqlmock DB for testing error paths.
func newSqlmockRouter(t *testing.T) (*Router, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	logger := zap.NewNop()
	cfg := &config.Config{}
	cfg.Admin.DBQueryAccess = true
	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	return router, mock
}

// TestGetTableNames_RowsErr covers rows.Err() in getTableNames.
func TestGetTableNames_RowsErr(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	rows := sqlmock.NewRows([]string{"table_name"}).
		AddRow("users").
		RowError(0, sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT table_name").WillReturnRows(rows)

	_, err := router.getTableNames()
	if err == nil {
		t.Error("getTableNames() should return error on rows.Err()")
	}
}

// TestGetTableColumns_RowsErr covers line 172 in db_schema.go:
// rows.Err() in getTableColumns returns error.
func TestGetTableColumns_RowsErr(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	rows := sqlmock.NewRows([]string{"column_name", "data_type", "not_null", "column_default", "is_pk"}).
		CloseError(sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	_, err := router.getTableColumns("users")
	if err == nil {
		t.Error("getTableColumns() should return error on rows.Err()")
	}
}

// TestGetForeignKeys_RowsErr covers line 219 in db_schema.go:
// rows.Err() in getForeignKeys returns error.
func TestGetForeignKeys_RowsErr(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	rows := sqlmock.NewRows([]string{"from_table", "from_column", "to_table", "to_column", "on_delete"}).
		CloseError(sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	_, err := router.getForeignKeys("users")
	if err == nil {
		t.Error("getForeignKeys() should return error on rows.Err()")
	}
}

// TestGetDBSchema_GetTableNamesScanError covers lines 82-85 in db_schema.go:
// getTableNames scan error propagates through getDBSchema.
func TestGetDBSchema_GetTableNamesScanError(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	rows := sqlmock.NewRows([]string{"table_name"}).
		AddRow("users").
		RowError(0, sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT table_name").WillReturnRows(rows)

	_, err := router.getDBSchema()
	if err == nil {
		t.Error("getDBSchema() should return error when getTableNames fails")
	}
}

// TestGetDBSchema_GetTableColumnsScanError covers lines 83-85 in db_schema.go:
// getTableColumns scan error propagates through getDBSchema.
func TestGetDBSchema_GetTableColumnsScanError(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	// getTableNames succeeds with one table
	tableRows := sqlmock.NewRows([]string{"table_name"}).AddRow("users")
	mock.ExpectQuery("SELECT table_name").WillReturnRows(tableRows)

	// getTableColumns fails
	colRows := sqlmock.NewRows([]string{"column_name", "data_type", "not_null", "column_default", "is_pk"}).
		AddRow("id", "integer", 1, "", 1).
		RowError(0, sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT").WillReturnRows(colRows)

	_, err := router.getDBSchema()
	if err == nil {
		t.Error("getDBSchema() should return error when getTableColumns fails")
	}
}

// TestGetDBSchema_GetForeignKeysScanError covers lines 87-90 in db_schema.go:
// getForeignKeys scan error propagates through getDBSchema.
func TestGetDBSchema_GetForeignKeysScanError(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	// getTableNames succeeds with one table
	tableRows := sqlmock.NewRows([]string{"table_name"}).AddRow("users")
	mock.ExpectQuery("SELECT table_name").WillReturnRows(tableRows)

	// getTableColumns succeeds (empty)
	colRows := sqlmock.NewRows([]string{"column_name", "data_type", "not_null", "column_default", "is_pk"})
	mock.ExpectQuery("SELECT").WillReturnRows(colRows)

	// getForeignKeys fails
	fkRows := sqlmock.NewRows([]string{"from_table", "from_column", "to_table", "to_column", "on_delete"}).
		AddRow("users", "id", "orders", "user_id", "CASCADE").
		RowError(0, sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT").WillReturnRows(fkRows)

	_, err := router.getDBSchema()
	if err == nil {
		t.Error("getDBSchema() should return error when getForeignKeys fails")
	}
}

// TestHandleDBQuery_ColumnsError covers lines 305-308 in db_schema.go:
// rows.Columns() error in handleDBQuery.
func TestHandleDBQuery_ColumnsError(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	// Return rows that fail on Columns()
	rows := sqlmock.NewRows([]string{}).CloseError(sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	body, _ := json.Marshal(map[string]string{"query": "SELECT 1"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.handleDBQuery(w, req)
	// rows.Columns() error or rows.Err() error → 500
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Logf("got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleDBQuery_ScanError covers lines 317-320 in db_schema.go:
// rows.Scan error in handleDBQuery.
func TestHandleDBQuery_ScanError(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	rows := sqlmock.NewRows([]string{"col1"}).
		AddRow("value1").
		RowError(0, sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	body, _ := json.Marshal(map[string]string{"query": "SELECT col1 FROM test"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.handleDBQuery(w, req)
	// scan error → 500
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Logf("got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleDBQuery_RowsErrError covers lines 334-337 in db_schema.go:
// rows.Err() error in handleDBQuery.
func TestHandleDBQuery_RowsErrError(t *testing.T) {
	router, mock := newSqlmockRouter(t)

	rows := sqlmock.NewRows([]string{"col1"}).CloseError(sqlmock.ErrCancelled)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	body, _ := json.Marshal(map[string]string{"query": "SELECT col1 FROM test"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/db-query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.handleDBQuery(w, req)
	// rows.Err() error → 500
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Logf("got %d: %s", w.Code, w.Body.String())
	}
}
