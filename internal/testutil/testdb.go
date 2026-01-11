package testutil

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates an in-memory SQLite database with all necessary tables for testing.
// This function creates all tables that are defined in internal/database/database.go
// to ensure tests use the same schema as production.
func SetupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create all tables matching the production schema
	createTables := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_id INTEGER UNIQUE NOT NULL,
		telegram_username TEXT,
		username TEXT,
		timezone TEXT DEFAULT '',
		preferred_training_time TEXT DEFAULT '',
		settings_json TEXT DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS word_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT UNIQUE NOT NULL,
		definition TEXT NOT NULL,
		pos TEXT,
		transcription TEXT,
		definition_ru TEXT,
		examples_json TEXT,
		verb_forms_json TEXT,
		display_en TEXT,
		processed_at TEXT,
		processing_error TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS word_forms (
		form TEXT PRIMARY KEY,
		word_card_id INTEGER NOT NULL,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS word_request_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		word TEXT,
		word_card_id INTEGER,
		input_word TEXT,
		requested_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS training_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word_card_id INTEGER NOT NULL,
		word_en TEXT NOT NULL,
		transcription TEXT,
		sense_index INTEGER NOT NULL,
		word_ru TEXT NOT NULL,
		meaning_en TEXT NOT NULL,
		example_en TEXT,
		example_ru TEXT,
		distractors_ru TEXT,
		distractors_en TEXT,
		hint TEXT,
		pos TEXT,
		display_word TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS user_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		training_card_id INTEGER NOT NULL,
		direction TEXT NOT NULL,
		state TEXT NOT NULL,
		ef REAL NOT NULL DEFAULT 2.5,
		reps INTEGER NOT NULL DEFAULT 0,
		interval_days INTEGER NOT NULL DEFAULT 0,
		learning_step INTEGER NOT NULL DEFAULT 0,
		lapse_count INTEGER NOT NULL DEFAULT 0,
		next_due_at TEXT,
		last_review_at TEXT,
		last_quality INTEGER,
		last_options_json TEXT,
		wrong_answers_json TEXT,
		stats_json TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS training_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		ended_at TEXT,
		source TEXT NOT NULL,
		planned_count INTEGER NOT NULL DEFAULT 0,
		done_count INTEGER NOT NULL DEFAULT 0,
		session_json TEXT DEFAULT ''
	);
	
	CREATE TABLE IF NOT EXISTS review_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER,
		user_id INTEGER NOT NULL,
		user_card_id INTEGER NOT NULL,
		direction TEXT NOT NULL,
		shown_at TEXT NOT NULL,
		options_shown_at TEXT,
		answered_at TEXT,
		t_delay_ms INTEGER,
		early_reveal INTEGER NOT NULL DEFAULT 0,
		option_count INTEGER NOT NULL,
		options_json TEXT,
		chosen_option TEXT,
		is_correct INTEGER NOT NULL DEFAULT 0,
		quality INTEGER NOT NULL,
		metrics_json TEXT,
		srs_before_json TEXT,
		srs_after_json TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS user_word_knowledge (
		user_id INTEGER NOT NULL,
		word_card_id INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'known' CHECK(status IN ('known')),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE,
		UNIQUE(user_id, word_card_id)
	);
	
	CREATE TABLE IF NOT EXISTS circuit_breaker_state (
		id INTEGER PRIMARY KEY CHECK(id = 1),
		is_open INTEGER DEFAULT 0,
		failure_count INTEGER DEFAULT 0,
		last_failure_at DATETIME,
		last_failure_message TEXT,
		last_reset_at DATETIME,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS web_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		session_token TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS web_otps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		code_hash TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		consumed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS word_set_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		parent_id INTEGER,
		name TEXT NOT NULL,
		description TEXT,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (parent_id) REFERENCES word_set_categories(id) ON DELETE SET NULL
	);
	
	CREATE TABLE IF NOT EXISTS word_sets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER,
		title TEXT NOT NULL,
		description TEXT,
		is_published INTEGER DEFAULT 1,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES word_set_categories(id) ON DELETE SET NULL
	);
	
	CREATE TABLE IF NOT EXISTS word_set_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word_set_id INTEGER NOT NULL,
		word_card_id INTEGER NOT NULL,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (word_set_id) REFERENCES word_sets(id) ON DELETE CASCADE,
		FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE,
		UNIQUE(word_set_id, word_card_id)
	);
	
	CREATE TABLE IF NOT EXISTS training_nudges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		local_date TEXT NOT NULL,
		sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		consumed_at DATETIME,
		due_count_at_send INTEGER,
		message_id INTEGER,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, local_date)
	);
	
	INSERT OR IGNORE INTO circuit_breaker_state (id) VALUES (1);
	`

	_, err = db.Exec(createTables)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return db
}
