package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
)

const (
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
		return DialectPostgres
	}
	if v, ok := connDialects.Load(db); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return DialectPostgres
}

func InsertAndReturnID(db *sql.DB, query string, args ...interface{}) (int64, error) {
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
