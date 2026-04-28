CREATE TABLE IF NOT EXISTS content_reports (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    word TEXT,
    translation_direction TEXT,
    word_card_id BIGINT REFERENCES word_cards(id) ON DELETE SET NULL,
    training_card_id BIGINT REFERENCES training_cards(id) ON DELETE SET NULL,
    user_card_id BIGINT REFERENCES user_cards(id) ON DELETE SET NULL,
    word_category TEXT,
    grammar_chapter_id TEXT,
    theory_block_id TEXT,
    grammar_question_id TEXT,
    payload_json TEXT,
    resolved_at TIMESTAMPTZ,
    resolved_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_content_reports_status_created_at ON content_reports(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_content_reports_source_type ON content_reports(source_type);
CREATE INDEX IF NOT EXISTS idx_content_reports_word_card_id ON content_reports(word_card_id);
CREATE INDEX IF NOT EXISTS idx_content_reports_grammar_question_id ON content_reports(grammar_question_id);
