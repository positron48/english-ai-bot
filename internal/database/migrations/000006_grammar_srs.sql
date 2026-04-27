CREATE TABLE IF NOT EXISTS grammar_theory_memory (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    language VARCHAR(16) NOT NULL,
    course_id VARCHAR(128) NOT NULL,
    chapter_id VARCHAR(255) NOT NULL,
    theory_block_id VARCHAR(255) NOT NULL,
    concept_id VARCHAR(255),
    state VARCHAR(32) NOT NULL DEFAULT 'new',
    review_count INT NOT NULL DEFAULT 0,
    correct_count INT NOT NULL DEFAULT 0,
    wrong_count INT NOT NULL DEFAULT 0,
    lapse_count INT NOT NULL DEFAULT 0,
    correct_streak INT NOT NULL DEFAULT 0,
    wrong_streak INT NOT NULL DEFAULT 0,
    ease DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    interval_days INT NOT NULL DEFAULT 0,
    mastery_score INT NOT NULL DEFAULT 0,
    next_review_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_review_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, language, course_id, theory_block_id)
);

CREATE INDEX IF NOT EXISTS idx_grammar_theory_memory_due
    ON grammar_theory_memory (user_id, language, course_id, next_review_at);

CREATE TABLE IF NOT EXISTS grammar_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    language VARCHAR(16) NOT NULL,
    course_id VARCHAR(128) NOT NULL,
    chapter_id VARCHAR(255) NOT NULL,
    theory_block_id VARCHAR(255) NOT NULL,
    concept_id VARCHAR(255),
    question_id VARCHAR(255) NOT NULL,
    question_source VARCHAR(32) NOT NULL DEFAULT 'training_pack',
    is_correct BOOLEAN NOT NULL,
    answer_payload_json TEXT,
    correct_payload_json TEXT,
    answered_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_grammar_attempts_user_block
    ON grammar_attempts (user_id, language, course_id, theory_block_id, answered_at DESC);

