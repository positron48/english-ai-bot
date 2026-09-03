CREATE TABLE placement_banks (
    course_code TEXT NOT NULL REFERENCES courses(code),
    version TEXT NOT NULL,
    data_json JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT false,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (course_code, version)
);
CREATE UNIQUE INDEX placement_banks_active ON placement_banks(course_code) WHERE active;

CREATE TABLE placement_sessions (
    id TEXT PRIMARY KEY,
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    idempotency_key TEXT NOT NULL,
    bank_version TEXT NOT NULL,
    policy_version TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'abandoned')),
    snapshot_json JSONB NOT NULL,
    result_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    UNIQUE (user_course_id, idempotency_key)
);
CREATE UNIQUE INDEX placement_sessions_active ON placement_sessions(user_course_id) WHERE status = 'active';
CREATE INDEX placement_sessions_history ON placement_sessions(user_course_id, created_at DESC);

-- Every start key is retained, including keys that merely resumed another tab's session.
CREATE TABLE placement_session_requests (
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    idempotency_key TEXT NOT NULL,
    session_id TEXT NOT NULL REFERENCES placement_sessions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_course_id, idempotency_key)
);

CREATE TABLE grammar_placement_access (
    user_course_id BIGINT PRIMARY KEY REFERENCES user_courses(id) ON DELETE CASCADE,
    score INTEGER NOT NULL DEFAULT 0,
    total_questions INTEGER NOT NULL DEFAULT 0,
    opened_sections_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    completed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source TEXT NOT NULL CHECK (source IN ('diagnostic', 'legacy', 'admin')),
    admin_override BOOLEAN NOT NULL DEFAULT false,
    cleared BOOLEAN NOT NULL DEFAULT false
);
