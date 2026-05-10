package database

import "fmt"

func (db *DB) migratePostgres() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS word_cards (
			id BIGSERIAL PRIMARY KEY,
			word TEXT NOT NULL UNIQUE,
			definition TEXT NOT NULL,
			pos TEXT,
			noun_gender TEXT,
			opposite_gender_word TEXT,
			transcription TEXT,
			definition_ru TEXT,
			examples_json TEXT,
			verb_forms_json TEXT,
			display_en TEXT,
			processed_at TIMESTAMPTZ,
			processing_error TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS word_forms (
			form TEXT PRIMARY KEY,
			word_card_id BIGINT NOT NULL REFERENCES word_cards(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			telegram_id BIGINT NOT NULL UNIQUE,
			telegram_username TEXT,
			timezone TEXT DEFAULT 'Europe/Moscow',
			preferred_training_time TEXT DEFAULT '19:00',
			settings_json TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS word_request_history (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			word TEXT,
			word_card_id BIGINT REFERENCES word_cards(id) ON DELETE CASCADE,
			input_word TEXT,
			requested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS training_cards (
			id BIGSERIAL PRIMARY KEY,
			word_card_id BIGINT NOT NULL REFERENCES word_cards(id) ON DELETE CASCADE,
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
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(word_card_id, sense_index)
		)`,
		`CREATE TABLE IF NOT EXISTS user_cards (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			training_card_id BIGINT NOT NULL REFERENCES training_cards(id) ON DELETE CASCADE,
			direction TEXT NOT NULL,
			state TEXT DEFAULT 'new',
			ef DOUBLE PRECISION DEFAULT 2.5,
			reps INTEGER DEFAULT 0,
			interval_days INTEGER DEFAULT 0,
			learning_step INTEGER DEFAULT 0,
			lapse_count INTEGER DEFAULT 0,
			next_due_at TIMESTAMPTZ,
			last_review_at TIMESTAMPTZ,
			last_quality INTEGER,
			last_options_json TEXT,
			wrong_answers_json TEXT,
			stats_json TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, training_card_id, direction)
		)`,
		`CREATE TABLE IF NOT EXISTS training_sessions (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			started_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			ended_at TIMESTAMPTZ,
			source TEXT,
			planned_count INTEGER,
			done_count INTEGER DEFAULT 0,
			session_json TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS review_events (
			id BIGSERIAL PRIMARY KEY,
			session_id BIGINT REFERENCES training_sessions(id) ON DELETE SET NULL,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			user_card_id BIGINT NOT NULL REFERENCES user_cards(id) ON DELETE CASCADE,
			direction TEXT NOT NULL,
			shown_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			options_shown_at TIMESTAMPTZ,
			answered_at TIMESTAMPTZ,
			t_delay_ms INTEGER,
			early_reveal INTEGER DEFAULT 0,
			option_count INTEGER,
			options_json TEXT,
			chosen_option TEXT,
			is_correct INTEGER,
			quality INTEGER,
			metrics_json TEXT,
			srs_before_json TEXT,
			srs_after_json TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS training_nudges (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			local_date TEXT NOT NULL,
			sent_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			consumed_at TIMESTAMPTZ,
			due_count_at_send INTEGER,
			message_id INTEGER,
			UNIQUE(user_id, local_date)
		)`,
		`CREATE TABLE IF NOT EXISTS circuit_breaker_state (
			id INTEGER PRIMARY KEY,
			is_open INTEGER DEFAULT 0,
			failure_count INTEGER DEFAULT 0,
			last_failure_at TIMESTAMPTZ,
			last_failure_message TEXT,
			last_reset_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS web_sessions (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			session_token TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			last_seen_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS web_otps (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			code_hash TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			consumed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS word_set_categories (
			id BIGSERIAL PRIMARY KEY,
			parent_id BIGINT REFERENCES word_set_categories(id) ON DELETE SET NULL,
			name TEXT NOT NULL,
			description TEXT,
			is_published INTEGER DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS word_sets (
			id BIGSERIAL PRIMARY KEY,
			category_id BIGINT REFERENCES word_set_categories(id) ON DELETE SET NULL,
			title TEXT NOT NULL,
			description TEXT,
			is_published INTEGER DEFAULT 1,
			sort_order INTEGER DEFAULT 0,
			preferred_pos TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS word_set_items (
			id BIGSERIAL PRIMARY KEY,
			word_set_id BIGINT NOT NULL REFERENCES word_sets(id) ON DELETE CASCADE,
			word_card_id BIGINT NOT NULL REFERENCES word_cards(id) ON DELETE CASCADE,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(word_set_id, word_card_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_word_knowledge (
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			word_card_id BIGINT NOT NULL REFERENCES word_cards(id) ON DELETE CASCADE,
			status TEXT NOT NULL DEFAULT 'known',
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, word_card_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_word_mastering (
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			word_card_id BIGINT NOT NULL REFERENCES word_cards(id) ON DELETE CASCADE,
			mastering_score SMALLINT NOT NULL DEFAULT 0,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, word_card_id)
		)`,
		`CREATE TABLE IF NOT EXISTS grammar_published_items (
			id BIGSERIAL PRIMARY KEY,
			item_type TEXT NOT NULL,
			item_id TEXT NOT NULL,
			is_published INTEGER DEFAULT 0,
			name TEXT,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
			UNIQUE(item_type, item_id)
		)`,
		`CREATE TABLE IF NOT EXISTS grammar_test_attempts (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			scope_type TEXT NOT NULL,
			scope_id TEXT NOT NULL,
			started_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			finished_at TIMESTAMPTZ,
			score INTEGER NOT NULL,
			passed INTEGER NOT NULL DEFAULT 0,
			total_questions INTEGER NOT NULL,
			answers_json TEXT,
			results_json TEXT,
			course_version TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS grammar_progress (
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			chapter_id TEXT NOT NULL,
			best_score INTEGER DEFAULT 0,
			passed_at TIMESTAMPTZ,
			last_attempt_at TIMESTAMPTZ,
			UNIQUE(user_id, chapter_id)
		)`,
		`CREATE TABLE IF NOT EXISTS reading_text_progress (
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			chapter_id TEXT NOT NULL,
			read_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, chapter_id)
		)`,
		`CREATE TABLE IF NOT EXISTS grammar_placement_test (
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			score INTEGER NOT NULL,
			total_questions INTEGER NOT NULL,
			opened_sections_json TEXT NOT NULL,
			completed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tts_generation_status (
			word TEXT PRIMARY KEY,
			state TEXT NOT NULL DEFAULT 'pending',
			attempt_count INTEGER NOT NULL DEFAULT 0,
			max_attempts INTEGER NOT NULL DEFAULT 3,
			last_error_code TEXT,
			last_error_message TEXT,
			last_provider TEXT,
			audio_rel_path TEXT,
			last_attempt_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tts_generation_status_state ON tts_generation_status(state)`,
		`CREATE INDEX IF NOT EXISTS idx_tts_generation_status_updated_at ON tts_generation_status(updated_at)`,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_access_categories (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_access_user_categories (
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			category_id BIGINT NOT NULL REFERENCES user_access_categories(id) ON DELETE CASCADE,
			UNIQUE(user_id, category_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_access_category_permissions (
			category_id BIGINT NOT NULL REFERENCES user_access_categories(id) ON DELETE CASCADE,
			permission TEXT NOT NULL,
			UNIQUE(category_id, permission)
		)`,
		`ALTER TABLE word_cards ADD COLUMN IF NOT EXISTS noun_gender TEXT`,
		`ALTER TABLE word_cards ADD COLUMN IF NOT EXISTS opposite_gender_word TEXT`,
		`INSERT INTO circuit_breaker_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING`,
		`INSERT INTO app_settings (key, value) VALUES ('hide_placement_test_button', 'false') ON CONFLICT (key) DO NOTHING`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute postgres migration: %w", err)
		}
	}

	if err := runEmbeddedSQLMigrations(db); err != nil {
		return fmt.Errorf("failed to run sql file migrations: %w", err)
	}

	db.logger.Info("postgres database migration completed successfully")
	return nil
}
