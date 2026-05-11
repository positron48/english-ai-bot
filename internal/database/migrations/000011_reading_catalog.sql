CREATE TABLE IF NOT EXISTS reading_categories (
    category_id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    title_translations TEXT,
    level TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    text_ids TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS reading_texts (
    text_id TEXT NOT NULL PRIMARY KEY,
    category_id TEXT NOT NULL,
    title TEXT NOT NULL,
    title_translations TEXT,
    level TEXT NOT NULL,
    target_language TEXT NOT NULL,
    reading_passage TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reading_texts_category_id ON reading_texts(category_id);
