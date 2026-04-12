-- Runtime verb example templates (optional DB layer; built-in code catalog remains fallback).
CREATE TABLE IF NOT EXISTS verb_example_templates (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    lemma_match TEXT NOT NULL,
    verb_class TEXT,
    mood TEXT,
    tense TEXT,
    es_suffix TEXT NOT NULL,
    ru_pattern TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_verb_example_templates_lemma
    ON verb_example_templates (lower(lemma_match));

COMMENT ON TABLE verb_example_templates IS 'Deterministic ES/RU cloze tails; merged at runtime with built-in catalog (duplicate code IDs are ignored if already defined in code).';
