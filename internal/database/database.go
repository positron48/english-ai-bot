package database

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"
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
		// Existing tables
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
		
		// Training system tables
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id INTEGER NOT NULL UNIQUE,
			timezone TEXT DEFAULT 'Europe/Moscow',
			preferred_training_time TEXT DEFAULT '19:00',
			settings_json TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS training_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			word_card_id INTEGER NOT NULL,
			word_en TEXT NOT NULL,
			transcription TEXT,
			sense_index INTEGER NOT NULL DEFAULT 0,
			word_ru TEXT NOT NULL,
			meaning_en TEXT NOT NULL,
			example_en TEXT,
			example_ru TEXT,
			distractors_ru TEXT,
			distractors_en TEXT,
			hint TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE,
			UNIQUE(word_card_id, sense_index)
		)`,
		
		`CREATE TABLE IF NOT EXISTS user_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			training_card_id INTEGER NOT NULL,
			direction TEXT NOT NULL CHECK(direction IN ('ru_en', 'en_ru')),
			state TEXT DEFAULT 'new' CHECK(state IN ('new', 'learning', 'review')),
			ef REAL DEFAULT 2.5,
			reps INTEGER DEFAULT 0,
			interval_days INTEGER DEFAULT 0,
			learning_step INTEGER DEFAULT 0,
			lapse_count INTEGER DEFAULT 0,
			next_due_at DATETIME,
			last_review_at DATETIME,
			last_quality INTEGER,
			last_options_json TEXT,
			wrong_answers_json TEXT,
			stats_json TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (training_card_id) REFERENCES training_cards(id) ON DELETE CASCADE,
			UNIQUE(user_id, training_card_id, direction)
		)`,
		
		`CREATE TABLE IF NOT EXISTS training_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			ended_at DATETIME,
			source TEXT CHECK(source IN ('nudge', 'manual')),
			planned_count INTEGER,
			done_count INTEGER DEFAULT 0,
			session_json TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		
		`CREATE TABLE IF NOT EXISTS review_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER,
			user_id INTEGER NOT NULL,
			user_card_id INTEGER NOT NULL,
			direction TEXT NOT NULL,
			shown_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			options_shown_at DATETIME,
			answered_at DATETIME,
			t_delay_ms INTEGER,
			early_reveal INTEGER DEFAULT 0,
			option_count INTEGER,
			options_json TEXT,
			chosen_option TEXT,
			is_correct INTEGER,
			quality INTEGER,
			metrics_json TEXT,
			srs_before_json TEXT,
			srs_after_json TEXT,
			FOREIGN KEY (session_id) REFERENCES training_sessions(id) ON DELETE SET NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (user_card_id) REFERENCES user_cards(id) ON DELETE CASCADE
		)`,
		
		`CREATE TABLE IF NOT EXISTS training_nudges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			local_date TEXT NOT NULL,
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			consumed_at DATETIME,
			due_count_at_send INTEGER,
			message_id INTEGER,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(user_id, local_date)
		)`,
		
		`CREATE TABLE IF NOT EXISTS circuit_breaker_state (
			id INTEGER PRIMARY KEY CHECK(id = 1),
			is_open INTEGER DEFAULT 0,
			failure_count INTEGER DEFAULT 0,
			last_failure_at DATETIME,
			last_failure_message TEXT,
			last_reset_at DATETIME,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// Indexes for existing tables
		`CREATE INDEX IF NOT EXISTS idx_word_cards_word ON word_cards(word)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_user_id ON word_request_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_word ON word_request_history(word)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_requested_at ON word_request_history(requested_at)`,
		
		// Indexes for training tables
		`CREATE INDEX IF NOT EXISTS idx_training_cards_word_card_id ON training_cards(word_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_cards_user_id ON user_cards(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_cards_training_card_id ON user_cards(training_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_cards_next_due_at ON user_cards(user_id, next_due_at)`,
		`CREATE INDEX IF NOT EXISTS idx_user_cards_state ON user_cards(user_id, state)`,
		`CREATE INDEX IF NOT EXISTS idx_training_sessions_user_id ON training_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_review_events_user_id ON review_events(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_training_nudges_user_date ON training_nudges(user_id, local_date)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	// Initialize circuit breaker state
	_, err := db.conn.Exec(`INSERT OR IGNORE INTO circuit_breaker_state (id) VALUES (1)`)
	if err != nil {
		return fmt.Errorf("failed to initialize circuit breaker: %w", err)
	}

	db.logger.Info("database migration completed successfully")
	return nil
}
