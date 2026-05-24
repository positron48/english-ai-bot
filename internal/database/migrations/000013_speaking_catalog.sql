CREATE TABLE IF NOT EXISTS speaking_categories (
    category_id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    title_translations TEXT,
    level TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    task_ids TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS speaking_tasks (
    task_id TEXT NOT NULL PRIMARY KEY,
    category_id TEXT NOT NULL,
    title TEXT NOT NULL,
    level TEXT NOT NULL,
    task_type TEXT NOT NULL,
    target_language TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    task_json TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_speaking_tasks_category_id ON speaking_tasks(category_id);

CREATE TABLE IF NOT EXISTS speaking_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    task_ids TEXT NOT NULL,
    current_task_index INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_speaking_sessions_user_id ON speaking_sessions(user_id);

CREATE TABLE IF NOT EXISTS speaking_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id BIGINT NOT NULL REFERENCES speaking_sessions(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL,
    attempt_no INTEGER NOT NULL,
    mode TEXT NOT NULL DEFAULT 'initial',
    understood_answer TEXT,
    meaning_score INTEGER,
    grammar_score INTEGER,
    pronunciation_score INTEGER,
    fluency_score INTEGER,
    is_acceptable BOOLEAN,
    audio_quality TEXT,
    feedback_ru TEXT,
    better_version TEXT,
    repeat_task TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_speaking_attempts_session_id ON speaking_attempts(session_id);
CREATE INDEX IF NOT EXISTS idx_speaking_attempts_user_task ON speaking_attempts(user_id, task_id);
