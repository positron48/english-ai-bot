package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
	"go.uber.org/zap"
)

// DB wraps database connection
type DB struct {
	conn   *sql.DB
	logger *zap.Logger
}

// New creates a new database connection
func New(dbPath string, logger *zap.Logger) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath+"?_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{
		conn:   conn,
		logger: logger,
	}

	if err := db.migrate(); err != nil {
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

// migrate creates necessary tables
func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS word_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			word TEXT NOT NULL UNIQUE,
			definition TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS word_request_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			word TEXT NOT NULL,
			requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (word) REFERENCES word_cards(word) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_word_cards_word ON word_cards(word)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_user_id ON word_request_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_word ON word_request_history(word)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_requested_at ON word_request_history(requested_at)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	db.logger.Info("database migration completed successfully")
	return nil
}

