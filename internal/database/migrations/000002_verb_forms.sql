CREATE TABLE IF NOT EXISTS verb_lemmas (
    id BIGSERIAL PRIMARY KEY,
    lemma TEXT NOT NULL,
    language TEXT NOT NULL,
    source TEXT,
    source_version TEXT,
    checksum TEXT,
    metadata_json TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(lemma, language)
);

CREATE TABLE IF NOT EXISTS verb_forms_dict (
    id BIGSERIAL PRIMARY KEY,
    verb_lemma_id BIGINT NOT NULL REFERENCES verb_lemmas(id) ON DELETE CASCADE,
    mood TEXT NOT NULL,
    tense TEXT NOT NULL,
    person TEXT NOT NULL,
    number TEXT NOT NULL,
    surface_form TEXT NOT NULL,
    is_irregular INTEGER NOT NULL DEFAULT 0,
    tags_json TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(verb_lemma_id, mood, tense, person, number)
);

CREATE TABLE IF NOT EXISTS word_verb_lemmas (
    word_card_id BIGINT PRIMARY KEY REFERENCES word_cards(id) ON DELETE CASCADE,
    verb_lemma_id BIGINT NOT NULL REFERENCES verb_lemmas(id) ON DELETE CASCADE,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    source TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verb_form_examples (
    id BIGSERIAL PRIMARY KEY,
    verb_form_dict_id BIGINT NOT NULL REFERENCES verb_forms_dict(id) ON DELETE CASCADE,
    example_target TEXT NOT NULL,
    gloss_native TEXT,
    source TEXT,
    quality_score INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(verb_form_dict_id, example_target)
);

CREATE TABLE IF NOT EXISTS verb_training_cards (
    id BIGSERIAL PRIMARY KEY,
    word_card_id BIGINT NOT NULL REFERENCES word_cards(id) ON DELETE CASCADE,
    verb_form_dict_id BIGINT NOT NULL REFERENCES verb_forms_dict(id) ON DELETE CASCADE,
    card_type TEXT NOT NULL,
    prompt_json TEXT NOT NULL,
    answer_json TEXT NOT NULL,
    distractors_json TEXT,
    example_id BIGINT REFERENCES verb_form_examples(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(word_card_id, verb_form_dict_id, card_type)
);

CREATE TABLE IF NOT EXISTS user_verb_cards (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    verb_training_card_id BIGINT NOT NULL REFERENCES verb_training_cards(id) ON DELETE CASCADE,
    state TEXT NOT NULL DEFAULT 'new',
    ef DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    reps INTEGER NOT NULL DEFAULT 0,
    interval_days INTEGER NOT NULL DEFAULT 0,
    learning_step INTEGER NOT NULL DEFAULT 0,
    lapse_count INTEGER NOT NULL DEFAULT 0,
    next_due_at TIMESTAMPTZ,
    last_review_at TIMESTAMPTZ,
    last_quality INTEGER,
    stats_json TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, verb_training_card_id)
);

CREATE TABLE IF NOT EXISTS verb_training_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMPTZ,
    planned_count INTEGER NOT NULL DEFAULT 0,
    done_count INTEGER NOT NULL DEFAULT 0,
    session_json TEXT
);

CREATE TABLE IF NOT EXISTS verb_review_events (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT REFERENCES verb_training_sessions(id) ON DELETE SET NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_verb_card_id BIGINT NOT NULL REFERENCES user_verb_cards(id) ON DELETE CASCADE,
    shown_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    answered_at TIMESTAMPTZ,
    is_correct INTEGER NOT NULL,
    quality INTEGER NOT NULL,
    metrics_json TEXT,
    srs_before_json TEXT,
    srs_after_json TEXT
);

CREATE INDEX IF NOT EXISTS idx_verb_forms_scope ON verb_forms_dict(verb_lemma_id, tense, mood, person, number);
CREATE INDEX IF NOT EXISTS idx_user_verb_cards_due ON user_verb_cards(user_id, next_due_at);
CREATE INDEX IF NOT EXISTS idx_word_verb_lemmas_lemma ON word_verb_lemmas(verb_lemma_id);
