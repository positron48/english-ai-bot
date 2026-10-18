-- Separate daily itinerary; exercise engines keep ownership of their own attempts.
CREATE TABLE daily_journeys (
    id TEXT PRIMARY KEY,
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    journey_date DATE NOT NULL,
    day_timezone TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'completed', 'archived', 'unavailable')),
    algorithm_version INTEGER NOT NULL,
    revision BIGINT NOT NULL DEFAULT 0 CHECK (revision >= 0),
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_course_id, journey_date)
);
CREATE UNIQUE INDEX daily_journeys_one_active ON daily_journeys(user_course_id) WHERE status = 'active';
CREATE TABLE daily_journey_steps (
    id TEXT PRIMARY KEY,
    journey_id TEXT NOT NULL REFERENCES daily_journeys(id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position >= 0),
    mode TEXT NOT NULL CHECK (mode IN ('words', 'reading', 'grammar', 'grammar_srs', 'verb_forms', 'picture', 'conversation')),
    required BOOLEAN NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'active', 'completed', 'skipped', 'unavailable')),
    title TEXT NOT NULL,
    goal TEXT NOT NULL,
    content_ref_json JSONB NOT NULL,
    content_version TEXT NOT NULL,
    target_count INTEGER NOT NULL CHECK (target_count >= 0),
    completed_count INTEGER NOT NULL DEFAULT 0 CHECK (completed_count >= 0),
    state_json JSONB NOT NULL DEFAULT '{}',
    launch_json JSONB,
    feedback TEXT,
    result TEXT,
    result_json JSONB NOT NULL DEFAULT '{}',
    skipped_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    UNIQUE (journey_id, position),
    UNIQUE (journey_id, mode),
    UNIQUE (journey_id, id),
    CHECK (NOT required OR status <> 'skipped')
);
CREATE UNIQUE INDEX daily_journey_one_skip ON daily_journey_steps(journey_id) WHERE status = 'skipped';
CREATE TABLE daily_journey_actions (
    journey_id TEXT NOT NULL REFERENCES daily_journeys(id) ON DELETE CASCADE,
    client_action_id TEXT NOT NULL,
    step_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    payload_hash TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('processing', 'completed')),
    result_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (journey_id, client_action_id),
    FOREIGN KEY (journey_id, step_id) REFERENCES daily_journey_steps(journey_id, id) ON DELETE CASCADE
);
CREATE TABLE grammar_block_progress (
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    chapter_id TEXT NOT NULL,
    block_id TEXT NOT NULL,
    content_version TEXT NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_course_id, chapter_id, block_id, content_version)
);
