//go:build test

package database

import (
	"database/sql"
)

// OpenFailingDB returns a *sql.DB that will fail on all operations.
// It uses a DSN pointing to a non-existent server (port 1) so that
// sql.Open succeeds but every query/exec will return a connection error.
// This is intended for use in tests that need to exercise error paths.
func OpenFailingDB() (*sql.DB, error) {
	registerPostgresCompatDriver()
	db, err := sql.Open("postgres_compat", "host=127.0.0.1 port=1 user=test dbname=test sslmode=disable")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	return db, nil
}
