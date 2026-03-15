package database

import (
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// DB wraps database connection
type DB struct {
	conn    *sql.DB
	logger  *zap.Logger
	dialect string
}

// openPostgresDBFunc is used by NewWithConfig; replaced in tests to cover migration failure path.
var openPostgresDBFunc = openPostgresDB

// NewWithConfig creates a database connection. Only PostgreSQL is supported.
// Requires DATABASE_URL (e.g. postgres://user:pass@host:5432/dbname?sslmode=disable).
func NewWithConfig(driver, path, url string, logger *zap.Logger) (*DB, error) {
	if !strings.EqualFold(strings.TrimSpace(driver), DialectPostgres) {
		return nil, fmt.Errorf("only postgres driver is supported, got: %q", driver)
	}
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	conn, err := openPostgresDBFunc(url)
	if err != nil {
		return nil, err
	}
	db := &DB{
		conn:    conn,
		logger:  logger,
		dialect: DialectPostgres,
	}
	registerConnDialect(conn, DialectPostgres)
	if err := db.migratePostgres(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetConnection returns the underlying database connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}
