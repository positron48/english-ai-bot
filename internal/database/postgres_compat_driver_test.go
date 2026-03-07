package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"sync"
	"testing"
)

// mockDriver и mockConn позволяют проверить, что compatConn передаёт в базовый
// драйвер запрос после rebindQuestionToDollar (?) -> ($1, $2, ...).
type mockDriver struct {
	mu    sync.Mutex
	lastPrepareQuery string
	lastPrepareContextQuery string
}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{driver: d}, nil
}

type mockConn struct {
	driver *mockDriver
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	c.driver.mu.Lock()
	c.driver.lastPrepareQuery = query
	c.driver.mu.Unlock()
	return &mockStmt{}, nil
}

func (c *mockConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	c.driver.mu.Lock()
	c.driver.lastPrepareContextQuery = query
	c.driver.mu.Unlock()
	return &mockStmt{}, nil
}

func (c *mockConn) Close() error              { return nil }
func (c *mockConn) Begin() (driver.Tx, error)  { return nil, nil }

// mockConnNoPrepareContext реализует только driver.Conn (без PrepareContext),
// чтобы проверить fallback compatConn.PrepareContext -> Prepare.
type mockConnNoPrepareContext struct {
	driver *mockDriverNoPrepareContext
}

func (c *mockConnNoPrepareContext) Prepare(query string) (driver.Stmt, error) {
	c.driver.mu.Lock()
	c.driver.lastPrepareQuery = query
	c.driver.mu.Unlock()
	return &mockStmt{}, nil
}

func (c *mockConnNoPrepareContext) Close() error              { return nil }
func (c *mockConnNoPrepareContext) Begin() (driver.Tx, error)  { return nil, nil }

type mockDriverNoPrepareContext struct {
	mu               sync.Mutex
	lastPrepareQuery string
}

func (d *mockDriverNoPrepareContext) Open(name string) (driver.Conn, error) {
	return &mockConnNoPrepareContext{driver: d}, nil
}

type mockStmt struct{}

func (s *mockStmt) Close() error   { return nil }
func (s *mockStmt) NumInput() int   { return -1 }
func (s *mockStmt) Exec([]driver.Value) (driver.Result, error) { return nil, nil }
func (s *mockStmt) Query([]driver.Value) (driver.Rows, error)  { return nil, nil }

func TestCompatConn_Prepare_rebindsQuestionToDollar(t *testing.T) {
	// sql.DB.Prepare() в современных версиях Go вызывает PrepareContext,
	// поэтому проверяем lastPrepareContextQuery.
	md := &mockDriver{}
	sql.Register("postgres_compat_mock_prepare", &compatDriver{base: md})
	conn, err := sql.Open("postgres_compat_mock_prepare", "test")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	stmt, err := conn.Prepare("SELECT ? FROM t WHERE x = ?")
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer stmt.Close()

	md.mu.Lock()
	got := md.lastPrepareContextQuery
	if got == "" {
		got = md.lastPrepareQuery
	}
	md.mu.Unlock()
	want := "SELECT $1 FROM t WHERE x = $2"
	if got != want {
		t.Errorf("Prepare received query = %q, want %q", got, want)
	}
}

func TestCompatConn_PrepareContext_fallbackWhenNoPrepareContext(t *testing.T) {
	md := &mockDriverNoPrepareContext{}
	sql.Register("postgres_compat_mock_noprepctx", &compatDriver{base: md})
	conn, err := sql.Open("postgres_compat_mock_noprepctx", "test")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	stmt, err := conn.PrepareContext(ctx, "SELECT ?")
	if err != nil {
		t.Fatalf("PrepareContext: %v", err)
	}
	defer stmt.Close()

	md.mu.Lock()
	got := md.lastPrepareQuery
	md.mu.Unlock()
	if got != "SELECT $1" {
		t.Errorf("PrepareContext fallback received query = %q, want SELECT $1", got)
	}
}

func TestCompatConn_PrepareContext_rebindsQuestionToDollar(t *testing.T) {
	md := &mockDriver{}
	sql.Register("postgres_compat_mock_ctx", &compatDriver{base: md})
	conn, err := sql.Open("postgres_compat_mock_ctx", "test")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	stmt, err := conn.PrepareContext(ctx, "INSERT INTO t (a,b) VALUES (?, ?)")
	if err != nil {
		t.Fatalf("PrepareContext: %v", err)
	}
	defer stmt.Close()

	md.mu.Lock()
	got := md.lastPrepareContextQuery
	md.mu.Unlock()
	want := "INSERT INTO t (a,b) VALUES ($1, $2)"
	if got != want {
		t.Errorf("PrepareContext received query = %q, want %q", got, want)
	}
}

// Тесты rebindQuestionToDollar через Prepare (косвенно).
func Test_rebindQuestionToDollar(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"empty", "", ""},
		{"no placeholders", "SELECT 1", "SELECT 1"},
		{"single question", "SELECT ?", "SELECT $1"},
		{"two questions", "SELECT ? FROM t WHERE x = ?", "SELECT $1 FROM t WHERE x = $2"},
		{"inside single quotes", "SELECT '?' FROM t", "SELECT '?' FROM t"},
		{"inside double quotes", `SELECT "?" FROM t`, `SELECT "?" FROM t`},
		{"mixed quoted and unquoted", "SELECT ? FROM t WHERE c = '?' AND d = ?", "SELECT $1 FROM t WHERE c = '?' AND d = $2"},
		{"double quote inside single", "SELECT 'say \"?\"' FROM t WHERE a = ?", "SELECT 'say \"?\"' FROM t WHERE a = $1"},
		{"single inside double", `SELECT "it's ?" FROM t WHERE a = ?`, `SELECT "it's ?" FROM t WHERE a = $1`},
		{"only in string", "SELECT 'hello' FROM t", "SELECT 'hello' FROM t"},
		{"multiple placeholders", "?, ?, ?", "$1, $2, $3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &mockDriver{}
			driverName := "postgres_compat_rebind_" + strings.ReplaceAll(tt.name, " ", "_")
			sql.Register(driverName, &compatDriver{base: md})
			conn, err := sql.Open(driverName, "test")
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			defer conn.Close()
			stmt, err := conn.Prepare(tt.query)
			if err != nil {
				t.Fatalf("Prepare: %v", err)
			}
			_ = stmt.Close()
			md.mu.Lock()
			got := md.lastPrepareContextQuery
			if got == "" {
				got = md.lastPrepareQuery
			}
			md.mu.Unlock()
			if got != tt.want {
				t.Errorf("rebindQuestionToDollar(%q) => %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}
