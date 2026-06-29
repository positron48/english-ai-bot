-- Sentence composition: daily Pro training where the user translates LLM-generated
-- Russian sentences (built from their well-learned words) into the target language,
-- graded teacher-style. One set per user/course/day; words rotate by least participation.

-- A generated set of sentences for one user/course/day.
CREATE TABLE IF NOT EXISTS sentence_sets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_code TEXT NOT NULL,
    generation_date DATE NOT NULL,
    scopes_json JSONB NOT NULL DEFAULT '[]'::jsonb,   -- verb scopes used at generation time
    status TEXT NOT NULL DEFAULT 'ready',             -- ready | started | completed
    started_at TIMESTAMPTZ,                           -- set on first attempt (consumption marker)
    completed_at TIMESTAMPTZ,
    star_count INTEGER NOT NULL DEFAULT 0,
    passed_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, course_code, generation_date),
    CHECK (status IN ('ready','started','completed'))
);

-- Latest-set lookup per user/course (regeneration guard + today's availability).
CREATE INDEX IF NOT EXISTS idx_sentence_sets_user_course_date
    ON sentence_sets (user_id, course_code, generation_date DESC);

-- Individual sentences within a set. Single-shot attempt result stored inline.
CREATE TABLE IF NOT EXISTS sentence_items (
    id BIGSERIAL PRIMARY KEY,
    set_id BIGINT NOT NULL REFERENCES sentence_sets(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    prompt_ru TEXT NOT NULL,            -- Russian sentence shown to the user
    reference_es TEXT NOT NULL,         -- LLM-suggested correct target translation (grading aid)
    word_card_ids JSONB NOT NULL DEFAULT '[]'::jsonb, -- word_cards actually used (participation)
    -- attempt result (single shot; NULL until attempted)
    attempted_at TIMESTAMPTZ,
    user_input TEXT,
    error_count INTEGER,
    outcome TEXT,                       -- star | passed | failed
    grading_json JSONB,                 -- teacher markup tokens
    UNIQUE (set_id, position),
    CHECK (outcome IS NULL OR outcome IN ('star','passed','failed'))
);

CREATE INDEX IF NOT EXISTS idx_sentence_items_set ON sentence_items (set_id, position);

-- Per-(user,word) participation counter driving least-used-first selection.
CREATE TABLE IF NOT EXISTS sentence_word_usage (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    word_card_id BIGINT NOT NULL REFERENCES word_cards(id) ON DELETE CASCADE,
    course_code TEXT NOT NULL,
    used_count INTEGER NOT NULL DEFAULT 0,
    last_used_on DATE,
    PRIMARY KEY (user_id, word_card_id, course_code)
);
