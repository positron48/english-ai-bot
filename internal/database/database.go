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
	conn, err := sql.Open("sqlite", dbPath+"?_foreign_keys=1&_busy_timeout=5000")
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
	// Migrate training_cards: rename meaning_ru to word_ru if needed
	if err := db.migrateTrainingCardsColumn(); err != nil {
		return fmt.Errorf("failed to migrate training_cards column: %w", err)
	}

	// Migrate word_cards: add processed_at and processing_error columns if needed
	if err := db.migrateWordCardsProcessingColumns(); err != nil {
		return fmt.Errorf("failed to migrate word_cards processing columns: %w", err)
	}

	// Migrate word_cards: add lemma-related columns (pos, transcription, etc.)
	if err := db.migrateWordCardsLemmaColumns(); err != nil {
		return fmt.Errorf("failed to migrate word_cards lemma columns: %w", err)
	}

	// Migrate word_request_history: add word_card_id and input_word
	if err := db.migrateWordRequestHistory(); err != nil {
		return fmt.Errorf("failed to migrate word_request_history: %w", err)
	}

	// Migrate training_cards: add pos and display_word
	if err := db.migrateTrainingCardsDisplayColumns(); err != nil {
		return fmt.Errorf("failed to migrate training_cards display columns: %w", err)
	}

	// Create word_forms table
	if err := db.migrateWordFormsTable(); err != nil {
		return fmt.Errorf("failed to migrate word_forms table: %w", err)
	}

	queries := []string{
		// Existing tables
		`CREATE TABLE IF NOT EXISTS word_cards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			word TEXT NOT NULL UNIQUE,
			definition TEXT NOT NULL,
			pos TEXT,
			transcription TEXT,
			definition_ru TEXT,
			examples_json TEXT,
			verb_forms_json TEXT,
			display_en TEXT,
			processed_at DATETIME,
			processing_error TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS word_forms (
			form TEXT PRIMARY KEY,
			word_card_id INTEGER NOT NULL,
			FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS word_request_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			word TEXT,
			word_card_id INTEGER,
			input_word TEXT,
			requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
		)`,
		
		// Training system tables
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id INTEGER NOT NULL UNIQUE,
			telegram_username TEXT,
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
			pos TEXT,
			display_word TEXT,
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
		
		// Web app tables
		`CREATE TABLE IF NOT EXISTS web_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			session_token TEXT NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		
		`CREATE TABLE IF NOT EXISTS web_otps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			code_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			consumed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		
		// Word sets tables
		`CREATE TABLE IF NOT EXISTS word_set_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			parent_id INTEGER,
			name TEXT NOT NULL,
			description TEXT,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (parent_id) REFERENCES word_set_categories(id) ON DELETE SET NULL
		)`,
		
		`CREATE TABLE IF NOT EXISTS word_sets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category_id INTEGER,
			title TEXT NOT NULL,
			description TEXT,
			is_published INTEGER DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			preferred_pos TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES word_set_categories(id) ON DELETE SET NULL
		)`,
		
		`CREATE TABLE IF NOT EXISTS word_set_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			word_set_id INTEGER NOT NULL,
			word_card_id INTEGER NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (word_set_id) REFERENCES word_sets(id) ON DELETE CASCADE,
			FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE,
			UNIQUE(word_set_id, word_card_id)
		)`,
		
		`CREATE TABLE IF NOT EXISTS user_word_knowledge (
			user_id INTEGER NOT NULL,
			word_card_id INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'known' CHECK(status IN ('known')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE,
			UNIQUE(user_id, word_card_id)
		)`,
		
		// Indexes for existing tables
		`CREATE INDEX IF NOT EXISTS idx_word_cards_word ON word_cards(word)`,
		`CREATE INDEX IF NOT EXISTS idx_word_forms_word_card_id ON word_forms(word_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_user_id ON word_request_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_word ON word_request_history(word)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_word_card_id ON word_request_history(word_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_requested_at ON word_request_history(requested_at)`,
		
		// Indexes for training tables
		`CREATE INDEX IF NOT EXISTS idx_training_cards_word_card_id ON training_cards(word_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_cards_user_id ON user_cards(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_cards_training_card_id ON user_cards(training_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_cards_next_due_at ON user_cards(user_id, next_due_at)`,
		`CREATE INDEX IF NOT EXISTS idx_user_cards_state ON user_cards(user_id, state)`,
		`CREATE INDEX IF NOT EXISTS idx_training_sessions_user_id ON training_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_review_events_user_id ON review_events(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_review_events_user_card_id ON review_events(user_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_training_nudges_user_date ON training_nudges(user_id, local_date)`,
		
		// Indexes for web app tables
		`CREATE INDEX IF NOT EXISTS idx_web_sessions_user_id ON web_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_web_sessions_token ON web_sessions(session_token)`,
		`CREATE INDEX IF NOT EXISTS idx_web_sessions_expires_at ON web_sessions(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_web_otps_user_id ON web_otps(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_web_otps_code_hash ON web_otps(code_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_web_otps_expires_at ON web_otps(expires_at)`,
		
		// Indexes for word sets tables
		`CREATE INDEX IF NOT EXISTS idx_word_set_categories_parent_id ON word_set_categories(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_sets_category_id ON word_sets(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_sets_is_published ON word_sets(is_published)`,
		`CREATE INDEX IF NOT EXISTS idx_word_set_items_word_set_id ON word_set_items(word_set_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_set_items_word_card_id ON word_set_items(word_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_word_knowledge_user_id ON user_word_knowledge(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_word_knowledge_word_card_id ON user_word_knowledge(word_card_id)`,
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

	// Migrate users table to add telegram_username if it doesn't exist
	// This must be done BEFORE creating index on telegram_username
	if err := db.migrateUsersTable(); err != nil {
		return fmt.Errorf("failed to migrate users table: %w", err)
	}

	// Create index on telegram_username AFTER the column is guaranteed to exist
	if _, err := db.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_users_telegram_username ON users(telegram_username)`); err != nil {
		return fmt.Errorf("failed to create index on telegram_username: %w", err)
	}

	// Migrate word_sets table to add sort_order if it doesn't exist
	if err := db.migrateWordSetsTable(); err != nil {
		return fmt.Errorf("failed to migrate word_sets table: %w", err)
	}

	// Migrate word_sets table to add preferred_pos if it doesn't exist
	if err := db.migrateWordSetsPreferredPOS(); err != nil {
		return fmt.Errorf("failed to migrate word_sets preferred_pos: %w", err)
	}

	db.logger.Info("database migration completed successfully")
	return nil
}

// migrateUsersTable adds telegram_username column if it doesn't exist
func (db *DB) migrateUsersTable() error {
	// Check if column exists
	var count int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') 
		WHERE name='telegram_username'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if count == 0 {
		// Column doesn't exist, add it
		_, err := db.conn.Exec(`ALTER TABLE users ADD COLUMN telegram_username TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add telegram_username column: %w", err)
		}
		db.logger.Info("added telegram_username column to users table")
	}

	return nil
}

// migrateTrainingCardsColumn migrates meaning_ru column to word_ru
func (db *DB) migrateTrainingCardsColumn() error {
	// Check if training_cards table exists
	var tableExists int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='training_cards'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists == 0 {
		// Table doesn't exist yet, will be created with correct schema
		return nil
	}

	// Check if meaning_ru column exists
	var columnExists int
	err = db.conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('training_cards') 
		WHERE name='meaning_ru'
	`).Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if columnExists > 0 {
		// Column exists, need to rename it
		// SQLite 3.25.0+ supports ALTER TABLE ... RENAME COLUMN
		// For older versions, we'd need to recreate the table, but we'll try the modern approach first
		_, err = db.conn.Exec(`ALTER TABLE training_cards RENAME COLUMN meaning_ru TO word_ru`)
		if err != nil {
			// If RENAME COLUMN is not supported, fall back to table recreation
			db.logger.Warn("RENAME COLUMN not supported, recreating table", zap.Error(err))
			return db.recreateTrainingCardsTable()
		}
		db.logger.Info("migrated training_cards.meaning_ru to word_ru")
	}

	return nil
}

// recreateTrainingCardsTable recreates training_cards table with word_ru column
func (db *DB) recreateTrainingCardsTable() error {
	// Start transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create new table with correct schema
	_, err = tx.Exec(`
		CREATE TABLE training_cards_new (
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
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// Copy data from old table to new table
	_, err = tx.Exec(`
		INSERT INTO training_cards_new 
		(id, word_card_id, word_en, transcription, sense_index, word_ru, meaning_en, example_en, example_ru, distractors_ru, distractors_en, hint, created_at)
		SELECT 
		id, word_card_id, word_en, transcription, sense_index, meaning_ru AS word_ru, meaning_en, example_en, example_ru, distractors_ru, distractors_en, hint, created_at
		FROM training_cards
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Drop old table
	_, err = tx.Exec(`DROP TABLE training_cards`)
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// Rename new table
	_, err = tx.Exec(`ALTER TABLE training_cards_new RENAME TO training_cards`)
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	// Recreate indexes
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_training_cards_word_card_id ON training_cards(word_card_id)`)
	if err != nil {
		return fmt.Errorf("failed to recreate index: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	db.logger.Info("recreated training_cards table with word_ru column")
	return nil
}

// migrateWordCardsProcessingColumns adds processed_at and processing_error columns to word_cards if they don't exist
func (db *DB) migrateWordCardsProcessingColumns() error {
	// Check if word_cards table exists
	var tableExists int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='word_cards'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists == 0 {
		// Table doesn't exist yet, will be created with correct schema
		return nil
	}

	// Check if processed_at column exists
	var processedAtExists int
	err = db.conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('word_cards') 
		WHERE name='processed_at'
	`).Scan(&processedAtExists)
	if err != nil {
		return fmt.Errorf("failed to check processed_at column existence: %w", err)
	}

	// Add processed_at column if it doesn't exist
	if processedAtExists == 0 {
		_, err = db.conn.Exec(`ALTER TABLE word_cards ADD COLUMN processed_at DATETIME NULL`)
		if err != nil {
			return fmt.Errorf("failed to add processed_at column: %w", err)
		}
		db.logger.Info("added processed_at column to word_cards")
	}

	// Check if processing_error column exists
	var processingErrorExists int
	err = db.conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('word_cards') 
		WHERE name='processing_error'
	`).Scan(&processingErrorExists)
	if err != nil {
		return fmt.Errorf("failed to check processing_error column existence: %w", err)
	}

	// Add processing_error column if it doesn't exist
	if processingErrorExists == 0 {
		_, err = db.conn.Exec(`ALTER TABLE word_cards ADD COLUMN processing_error TEXT NULL`)
		if err != nil {
			return fmt.Errorf("failed to add processing_error column: %w", err)
		}
		db.logger.Info("added processing_error column to word_cards")
	}

	return nil
}

// migrateWordCardsLemmaColumns adds lemma-related columns to word_cards
func (db *DB) migrateWordCardsLemmaColumns() error {
	// Check if word_cards table exists
	var tableExists int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='word_cards'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists == 0 {
		return nil
	}

	columns := []struct {
		name string
		typ  string
	}{
		{"pos", "TEXT"},
		{"transcription", "TEXT"},
		{"definition_ru", "TEXT"},
		{"examples_json", "TEXT"},
		{"verb_forms_json", "TEXT"},
		{"display_en", "TEXT"},
	}

	for _, col := range columns {
		var exists int
		err = db.conn.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('word_cards') 
			WHERE name=?
		`, col.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check column %s existence: %w", col.name, err)
		}

		if exists == 0 {
			_, err = db.conn.Exec(fmt.Sprintf(`ALTER TABLE word_cards ADD COLUMN %s %s`, col.name, col.typ))
			if err != nil {
				return fmt.Errorf("failed to add column %s: %w", col.name, err)
			}
			db.logger.Info("added column to word_cards", zap.String("column", col.name))
		}
	}

	return nil
}

// migrateWordRequestHistory migrates word_request_history to add word_card_id and input_word
func (db *DB) migrateWordRequestHistory() error {
	// Check if word_request_history table exists
	var tableExists int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='word_request_history'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists == 0 {
		return nil
	}

	// Check if word_card_id column exists
	var wordCardIDExists int
	err = db.conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('word_request_history') 
		WHERE name='word_card_id'
	`).Scan(&wordCardIDExists)
	if err != nil {
		return fmt.Errorf("failed to check word_card_id column: %w", err)
	}

	var inputWordExists int
	err = db.conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('word_request_history') 
		WHERE name='input_word'
	`).Scan(&inputWordExists)
	if err != nil {
		return fmt.Errorf("failed to check input_word column: %w", err)
	}

	// If both columns exist, migration already done
	if wordCardIDExists > 0 && inputWordExists > 0 {
		return nil
	}

	// Need to recreate table with new schema
	// Start transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create new table
	_, err = tx.Exec(`
		CREATE TABLE word_request_history_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			word TEXT,
			word_card_id INTEGER,
			input_word TEXT,
			requested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new table: %w", err)
	}

	// Copy data: try to resolve word to word_card_id
	// For existing records, we'll set input_word = word and try to find word_card_id
	_, err = tx.Exec(`
		INSERT INTO word_request_history_new (id, user_id, word, word_card_id, input_word, requested_at)
		SELECT 
			wrh.id,
			wrh.user_id,
			wrh.word,
			wc.id as word_card_id,
			wrh.word as input_word,
			wrh.requested_at
		FROM word_request_history wrh
		LEFT JOIN word_cards wc ON LOWER(wc.word) = LOWER(wrh.word)
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Drop old table
	_, err = tx.Exec(`DROP TABLE word_request_history`)
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// Rename new table
	_, err = tx.Exec(`ALTER TABLE word_request_history_new RENAME TO word_request_history`)
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	// Recreate indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_user_id ON word_request_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_word ON word_request_history(word)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_word_card_id ON word_request_history(word_card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_word_request_history_requested_at ON word_request_history(requested_at)`,
	}

	for _, idx := range indexes {
		if _, err = tx.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	db.logger.Info("migrated word_request_history table")
	return nil
}

// migrateTrainingCardsDisplayColumns adds pos and display_word columns to training_cards
func (db *DB) migrateTrainingCardsDisplayColumns() error {
	// Check if training_cards table exists
	var tableExists int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='training_cards'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists == 0 {
		return nil
	}

	columns := []struct {
		name string
		typ  string
	}{
		{"pos", "TEXT"},
		{"display_word", "TEXT"},
	}

	for _, col := range columns {
		var exists int
		err = db.conn.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('training_cards') 
			WHERE name=?
		`, col.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check column %s existence: %w", col.name, err)
		}

		if exists == 0 {
			_, err = db.conn.Exec(fmt.Sprintf(`ALTER TABLE training_cards ADD COLUMN %s %s`, col.name, col.typ))
			if err != nil {
				return fmt.Errorf("failed to add column %s: %w", col.name, err)
			}
			db.logger.Info("added column to training_cards", zap.String("column", col.name))
		}
	}

	return nil
}

// migrateWordFormsTable creates word_forms table if it doesn't exist
func (db *DB) migrateWordFormsTable() error {
	var tableExists int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='word_forms'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists > 0 {
		return nil
	}

	_, err = db.conn.Exec(`
		CREATE TABLE word_forms (
			form TEXT PRIMARY KEY,
			word_card_id INTEGER NOT NULL,
			FOREIGN KEY (word_card_id) REFERENCES word_cards(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create word_forms table: %w", err)
	}

	_, err = db.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_word_forms_word_card_id ON word_forms(word_card_id)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	db.logger.Info("created word_forms table")
	return nil
}

// migrateWordSetsTable adds sort_order column to word_sets if it doesn't exist
func (db *DB) migrateWordSetsTable() error {
	// Check if word_sets table exists
	var tableExists int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='word_sets'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists == 0 {
		// Table doesn't exist yet, will be created with correct schema
		return nil
	}

	// Check if sort_order column exists
	var columnExists int
	err = db.conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('word_sets') 
		WHERE name='sort_order'
	`).Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if columnExists == 0 {
		// Column doesn't exist, add it
		_, err := db.conn.Exec(`ALTER TABLE word_sets ADD COLUMN sort_order INTEGER DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("failed to add sort_order column: %w", err)
		}
		db.logger.Info("added sort_order column to word_sets table")
	}

	return nil
}

// migrateWordSetsPreferredPOS adds preferred_pos column to word_sets if it doesn't exist
func (db *DB) migrateWordSetsPreferredPOS() error {
	// Check if word_sets table exists
	var tableExists int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name='word_sets'
	`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if tableExists == 0 {
		// Table doesn't exist yet, will be created with correct schema
		return nil
	}

	// Check if preferred_pos column exists
	var columnExists int
	err = db.conn.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('word_sets') 
		WHERE name='preferred_pos'
	`).Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if columnExists == 0 {
		// Column doesn't exist, add it
		_, err := db.conn.Exec(`ALTER TABLE word_sets ADD COLUMN preferred_pos TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add preferred_pos column: %w", err)
		}
		db.logger.Info("added preferred_pos column to word_sets table")
	}

	return nil
}
