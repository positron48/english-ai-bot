CREATE TABLE IF NOT EXISTS reading_text_progress (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chapter_id TEXT NOT NULL,
    read_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, chapter_id)
);

CREATE INDEX IF NOT EXISTS idx_reading_text_progress_user_read_at
    ON reading_text_progress(user_id, read_at DESC);

