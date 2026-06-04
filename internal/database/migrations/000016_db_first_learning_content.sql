CREATE TABLE IF NOT EXISTS learning_content_import_runs (
    id BIGSERIAL PRIMARY KEY,
    app_code TEXT NOT NULL,
    bundle_id TEXT NOT NULL,
    target_lang TEXT NOT NULL,
    source TEXT NOT NULL,
    source_hash TEXT NOT NULL,
    mode TEXT NOT NULL,
    sections_count INTEGER NOT NULL DEFAULT 0,
    chapters_count INTEGER NOT NULL DEFAULT 0,
    training_questions_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMPTZ,
    error TEXT
);

CREATE TABLE IF NOT EXISTS grammar_content_sections (
    bundle_id TEXT NOT NULL,
    section_id TEXT NOT NULL,
    title TEXT NOT NULL,
    title_translations_json TEXT,
    level TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    chapter_ids_json TEXT NOT NULL,
    raw_json TEXT NOT NULL,
    source_hash TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (bundle_id, section_id)
);

CREATE TABLE IF NOT EXISTS grammar_content_chapters (
    bundle_id TEXT NOT NULL,
    chapter_id TEXT NOT NULL,
    section_id TEXT NOT NULL,
    title TEXT NOT NULL,
    title_translations_json TEXT,
    title_short TEXT,
    description TEXT,
    ui_language TEXT NOT NULL,
    target_language TEXT NOT NULL,
    level TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    estimated_minutes INTEGER NOT NULL DEFAULT 0,
    raw_json TEXT NOT NULL,
    source_hash TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (bundle_id, chapter_id)
);

CREATE INDEX IF NOT EXISTS idx_grammar_content_chapters_section
    ON grammar_content_chapters(bundle_id, section_id, sort_order);

CREATE TABLE IF NOT EXISTS grammar_content_bundle_meta (
    bundle_id TEXT PRIMARY KEY,
    app_code TEXT NOT NULL,
    native_lang TEXT NOT NULL,
    target_lang TEXT NOT NULL,
    version TEXT NOT NULL,
    generated_at TEXT,
    source_hash TEXT NOT NULL,
    sections_json TEXT NOT NULL,
    index_json TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS grammar_training_content_meta (
    bundle_id TEXT PRIMARY KEY,
    language TEXT NOT NULL,
    course_id TEXT NOT NULL,
    version TEXT NOT NULL,
    generated_at TEXT,
    index_json TEXT NOT NULL,
    source_hash TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS grammar_training_content_questions (
    bundle_id TEXT NOT NULL,
    question_id TEXT NOT NULL,
    chapter_id TEXT NOT NULL,
    theory_block_id TEXT NOT NULL,
    concept_id TEXT,
    difficulty INTEGER,
    raw_json TEXT NOT NULL,
    source_hash TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (bundle_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_grammar_training_content_chapter
    ON grammar_training_content_questions(bundle_id, chapter_id);

CREATE INDEX IF NOT EXISTS idx_grammar_training_content_block
    ON grammar_training_content_questions(bundle_id, theory_block_id);
