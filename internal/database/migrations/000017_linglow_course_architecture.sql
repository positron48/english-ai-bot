CREATE TABLE IF NOT EXISTS courses (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    target_lang TEXT NOT NULL,
    teaching_locale TEXT NOT NULL,
    ui_locale TEXT NOT NULL,
    title TEXT NOT NULL,
    city_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (status IN ('draft', 'active', 'locked', 'archived'))
);

CREATE TABLE IF NOT EXISTS user_courses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMPTZ,
    settings_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, course_id),
    CHECK (status IN ('active', 'paused', 'completed', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_user_courses_user
    ON user_courses(user_id, status);

CREATE INDEX IF NOT EXISTS idx_user_courses_course
    ON user_courses(course_id, status);

CREATE TABLE IF NOT EXISTS districts (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    level_code TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(course_id, code),
    UNIQUE(course_id, level_code),
    CHECK (status IN ('draft', 'active', 'locked', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_districts_course_order
    ON districts(course_id, sort_order);

CREATE TABLE IF NOT EXISTS locations (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    district_id BIGINT NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    location_type TEXT NOT NULL,
    title TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(district_id, code),
    CHECK (location_type IN ('grammar', 'word_market', 'reading', 'conversation', 'review_station', 'mistake_workshop')),
    CHECK (status IN ('draft', 'active', 'locked', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_locations_course_district
    ON locations(course_id, district_id, sort_order);

CREATE TABLE IF NOT EXISTS theme_lines (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    title TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(course_id, code),
    CHECK (status IN ('draft', 'active', 'locked', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_theme_lines_course_order
    ON theme_lines(course_id, sort_order);

CREATE TABLE IF NOT EXISTS modules (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    district_id BIGINT REFERENCES districts(id) ON DELETE SET NULL,
    location_id BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    theme_line_id BIGINT REFERENCES theme_lines(id) ON DELETE SET NULL,
    code TEXT NOT NULL,
    module_type TEXT NOT NULL,
    title TEXT NOT NULL,
    source_kind TEXT,
    source_id TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'draft',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(course_id, code),
    CHECK (module_type IN ('grammar', 'word_set', 'reading', 'speaking', 'review', 'mistake_workshop', 'daily_route')),
    CHECK (status IN ('draft', 'published', 'locked', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_modules_course_location
    ON modules(course_id, location_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_modules_source
    ON modules(course_id, source_kind, source_id)
    WHERE source_kind IS NOT NULL AND source_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS learning_objectives (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    module_id BIGINT REFERENCES modules(id) ON DELETE SET NULL,
    code TEXT NOT NULL,
    objective_type TEXT NOT NULL,
    title TEXT NOT NULL,
    cefr_level TEXT,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(course_id, code),
    CHECK (objective_type IN ('grammar_concept', 'vocab_set', 'reading_skill', 'speaking_pattern', 'pronunciation', 'chat_correction'))
);

CREATE INDEX IF NOT EXISTS idx_learning_objectives_module
    ON learning_objectives(module_id);

CREATE TABLE IF NOT EXISTS learning_items (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    module_id BIGINT REFERENCES modules(id) ON DELETE SET NULL,
    objective_id BIGINT REFERENCES learning_objectives(id) ON DELETE SET NULL,
    district_id BIGINT REFERENCES districts(id) ON DELETE SET NULL,
    location_id BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    theme_line_id BIGINT REFERENCES theme_lines(id) ON DELETE SET NULL,
    item_type TEXT NOT NULL,
    source_kind TEXT NOT NULL,
    source_id TEXT NOT NULL,
    title TEXT,
    cefr_level TEXT,
    content_hash TEXT,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(course_id, source_kind, source_id),
    CHECK (item_type IN ('word', 'grammar_chapter', 'grammar_concept', 'grammar_question', 'reading_text', 'reading_question', 'speaking_task', 'chat_correction', 'pronunciation')),
    CHECK (status IN ('draft', 'published', 'locked', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_learning_items_course_type
    ON learning_items(course_id, item_type, status);

CREATE INDEX IF NOT EXISTS idx_learning_items_module
    ON learning_items(module_id, status);

CREATE TABLE IF NOT EXISTS srs_items (
    id BIGSERIAL PRIMARY KEY,
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    learning_item_id BIGINT NOT NULL REFERENCES learning_items(id) ON DELETE CASCADE,
    state TEXT NOT NULL DEFAULT 'new',
    stability DOUBLE PRECISION,
    difficulty DOUBLE PRECISION,
    due_at TIMESTAMPTZ,
    last_review_at TIMESTAMPTZ,
    reps INTEGER NOT NULL DEFAULT 0,
    lapse_count INTEGER NOT NULL DEFAULT 0,
    stats_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_course_id, learning_item_id),
    CHECK (state IN ('new', 'learning', 'review', 'relearning', 'mastered', 'suspended'))
);

CREATE INDEX IF NOT EXISTS idx_srs_items_due
    ON srs_items(user_course_id, state, due_at);

CREATE TABLE IF NOT EXISTS exercise_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    learning_item_id BIGINT REFERENCES learning_items(id) ON DELETE SET NULL,
    srs_item_id BIGINT REFERENCES srs_items(id) ON DELETE SET NULL,
    mode TEXT NOT NULL,
    client_attempt_id TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    answered_at TIMESTAMPTZ,
    is_correct BOOLEAN,
    score INTEGER,
    quality INTEGER,
    prompt_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    answer_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    result_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_table TEXT,
    source_pk TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exercise_attempts_client_attempt
    ON exercise_attempts(user_course_id, client_attempt_id)
    WHERE client_attempt_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_exercise_attempts_course_time
    ON exercise_attempts(user_course_id, answered_at DESC);

CREATE INDEX IF NOT EXISTS idx_exercise_attempts_item_time
    ON exercise_attempts(learning_item_id, answered_at DESC);

CREATE TABLE IF NOT EXISTS learning_events (
    id BIGSERIAL PRIMARY KEY,
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    learning_item_id BIGINT REFERENCES learning_items(id) ON DELETE SET NULL,
    exercise_attempt_id BIGINT REFERENCES exercise_attempts(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    event_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    mode TEXT,
    source_table TEXT,
    source_pk TEXT,
    event_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_learning_events_course_time
    ON learning_events(user_course_id, event_time DESC);

CREATE INDEX IF NOT EXISTS idx_learning_events_item_time
    ON learning_events(learning_item_id, event_time DESC);

CREATE TABLE IF NOT EXISTS daily_course_stats (
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    local_date DATE NOT NULL,
    review_count INTEGER NOT NULL DEFAULT 0,
    new_count INTEGER NOT NULL DEFAULT 0,
    correct_count INTEGER NOT NULL DEFAULT 0,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    active_seconds INTEGER NOT NULL DEFAULT 0,
    stats_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_course_id, local_date)
);

CREATE TABLE IF NOT EXISTS mode_daily_stats (
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    local_date DATE NOT NULL,
    mode TEXT NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    correct_count INTEGER NOT NULL DEFAULT 0,
    active_seconds INTEGER NOT NULL DEFAULT 0,
    stats_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_course_id, local_date, mode)
);

CREATE TABLE IF NOT EXISTS district_progress (
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    district_id BIGINT NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
    foundation_score INTEGER NOT NULL DEFAULT 0,
    confidence_score INTEGER NOT NULL DEFAULT 0,
    stability_score INTEGER NOT NULL DEFAULT 0,
    weak_items_count INTEGER NOT NULL DEFAULT 0,
    opened_at TIMESTAMPTZ,
    stats_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_course_id, district_id)
);

CREATE TABLE IF NOT EXISTS learning_item_stats (
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    learning_item_id BIGINT NOT NULL REFERENCES learning_items(id) ON DELETE CASCADE,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    correct_count INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    stats_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_course_id, learning_item_id)
);

CREATE TABLE IF NOT EXISTS content_performance_stats (
    learning_item_id BIGINT PRIMARY KEY REFERENCES learning_items(id) ON DELETE CASCADE,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    correct_count INTEGER NOT NULL DEFAULT 0,
    user_count INTEGER NOT NULL DEFAULT 0,
    stats_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO courses (code, target_lang, teaching_locale, ui_locale, title, city_name, status)
VALUES
    ('en_ru', 'en', 'ru', 'ru', 'English RU', 'Luminaria City', 'active'),
    ('es_ru', 'es', 'ru', 'ru', 'Spanish RU', 'Ciudad Luminaria', 'active')
ON CONFLICT (code) DO UPDATE SET
    target_lang = excluded.target_lang,
    teaching_locale = excluded.teaching_locale,
    ui_locale = excluded.ui_locale,
    title = excluded.title,
    city_name = excluded.city_name,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP;

WITH district_seed(level_code, code, title, sort_order) AS (
    VALUES
        ('A0', 'a0_spark_gate', 'Puerta de la Chispa', 0),
        ('A1', 'a1_clear_plaza', 'Plaza Clara', 1),
        ('A2', 'a2_living_quarter', 'Barrio Vivo', 2),
        ('B1', 'b1_story_bridges', 'Puentes del Relato', 3),
        ('B2', 'b2_high_district', 'Distrito Alto', 4),
        ('C1', 'c1_mastery_campus', 'Campus de Maestria', 5)
)
INSERT INTO districts (course_id, code, level_code, title, sort_order, status)
SELECT c.id, ds.code, ds.level_code, ds.title, ds.sort_order, 'active'
FROM courses c
CROSS JOIN district_seed ds
WHERE c.code IN ('en_ru', 'es_ru')
ON CONFLICT (course_id, code) DO UPDATE SET
    level_code = excluded.level_code,
    title = excluded.title,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP;

WITH location_seed(code, location_type, title, sort_order) AS (
    VALUES
        ('grammar', 'grammar', 'Grammar Building', 0),
        ('word_market', 'word_market', 'Word Market', 1),
        ('reading', 'reading', 'Reading Spot', 2),
        ('conversation', 'conversation', 'Conversation Hub', 3),
        ('review_station', 'review_station', 'Review Station', 4),
        ('mistake_workshop', 'mistake_workshop', 'Mistake Workshop', 5)
)
INSERT INTO locations (course_id, district_id, code, location_type, title, sort_order, status)
SELECT d.course_id, d.id, ls.code, ls.location_type, ls.title, ls.sort_order, 'active'
FROM districts d
CROSS JOIN location_seed ls
ON CONFLICT (district_id, code) DO UPDATE SET
    location_type = excluded.location_type,
    title = excluded.title,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP;

WITH theme_seed(code, title, sort_order) AS (
    VALUES
        ('travel', 'Travel', 0),
        ('food_cafe', 'Food and Cafe', 1),
        ('daily_life', 'Daily Life', 2),
        ('work', 'Work', 3)
)
INSERT INTO theme_lines (course_id, code, title, sort_order, status)
SELECT c.id, ts.code, ts.title, ts.sort_order, 'active'
FROM courses c
CROSS JOIN theme_seed ts
WHERE c.code IN ('en_ru', 'es_ru')
ON CONFLICT (course_id, code) DO UPDATE SET
    title = excluded.title,
    sort_order = excluded.sort_order,
    status = excluded.status,
    updated_at = CURRENT_TIMESTAMP;
