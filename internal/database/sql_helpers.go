package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
)

const (
	DialectSQLite   = "sqlite"
	DialectPostgres = "postgres"
)

var connDialects sync.Map // map[*sql.DB]string

func registerConnDialect(db *sql.DB, dialect string) {
	if db == nil {
		return
	}
	connDialects.Store(db, dialect)
}

func GetDialect(db *sql.DB) string {
	if db == nil {
		return DialectSQLite
	}
	if v, ok := connDialects.Load(db); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return DialectSQLite
}

func isPostgresDB(db *sql.DB) bool {
	return GetDialect(db) == DialectPostgres
}

func InsertAndReturnID(db *sql.DB, query string, args ...interface{}) (int64, error) {
	if isPostgresDB(db) {
		q := query
		if !strings.Contains(strings.ToUpper(q), "RETURNING") {
			q += " RETURNING id"
		}
		var id int64
		if err := db.QueryRow(q, args...).Scan(&id); err != nil {
			return 0, fmt.Errorf("failed to insert row with returning id: %w", err)
		}
		return id, nil
	}

	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
