CREATE TABLE IF NOT EXISTS legacy_course_mappings (
    id BIGSERIAL PRIMARY KEY,
    source_app_code TEXT NOT NULL,
    source_db_label TEXT NOT NULL,
    source_course_code TEXT NOT NULL,
    source_course_id TEXT,
    target_course_id BIGINT REFERENCES courses(id) ON DELETE SET NULL,
    mapping_status TEXT NOT NULL DEFAULT 'pending',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_app_code, source_db_label, source_course_code),
    CHECK (mapping_status IN ('pending', 'mapped', 'conflict', 'ignored'))
);

CREATE INDEX IF NOT EXISTS idx_legacy_course_mappings_target
    ON legacy_course_mappings(target_course_id)
    WHERE target_course_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS legacy_user_mappings (
    id BIGSERIAL PRIMARY KEY,
    source_app_code TEXT NOT NULL,
    source_db_label TEXT NOT NULL,
    source_table TEXT NOT NULL DEFAULT 'users',
    source_user_id TEXT NOT NULL,
    target_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    stable_identity_type TEXT,
    stable_identity_value TEXT,
    mapping_status TEXT NOT NULL DEFAULT 'pending',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_app_code, source_db_label, source_table, source_user_id),
    CHECK (mapping_status IN ('pending', 'mapped', 'conflict', 'ignored'))
);

CREATE INDEX IF NOT EXISTS idx_legacy_user_mappings_target
    ON legacy_user_mappings(target_user_id)
    WHERE target_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_legacy_user_mappings_identity
    ON legacy_user_mappings(stable_identity_type, stable_identity_value)
    WHERE stable_identity_type IS NOT NULL AND stable_identity_value IS NOT NULL;

CREATE TABLE IF NOT EXISTS legacy_content_mappings (
    id BIGSERIAL PRIMARY KEY,
    source_app_code TEXT NOT NULL,
    source_db_label TEXT NOT NULL,
    source_table TEXT NOT NULL,
    source_pk TEXT NOT NULL,
    source_kind TEXT NOT NULL,
    source_id TEXT NOT NULL,
    target_course_id BIGINT REFERENCES courses(id) ON DELETE SET NULL,
    target_learning_item_id BIGINT REFERENCES learning_items(id) ON DELETE SET NULL,
    mapping_status TEXT NOT NULL DEFAULT 'pending',
    content_hash TEXT,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_app_code, source_db_label, source_table, source_pk),
    CHECK (mapping_status IN ('pending', 'mapped', 'conflict', 'ignored'))
);

CREATE INDEX IF NOT EXISTS idx_legacy_content_mappings_source_kind
    ON legacy_content_mappings(source_app_code, source_db_label, source_kind, source_id);

CREATE INDEX IF NOT EXISTS idx_legacy_content_mappings_target_item
    ON legacy_content_mappings(target_learning_item_id)
    WHERE target_learning_item_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS legacy_attempt_mappings (
    id BIGSERIAL PRIMARY KEY,
    source_app_code TEXT NOT NULL,
    source_db_label TEXT NOT NULL,
    source_table TEXT NOT NULL,
    source_pk TEXT NOT NULL,
    source_user_id TEXT,
    source_course_code TEXT,
    target_user_course_id BIGINT REFERENCES user_courses(id) ON DELETE SET NULL,
    target_exercise_attempt_id BIGINT REFERENCES exercise_attempts(id) ON DELETE SET NULL,
    target_learning_event_id BIGINT REFERENCES learning_events(id) ON DELETE SET NULL,
    mapping_status TEXT NOT NULL DEFAULT 'pending',
    source_hash TEXT,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(source_app_code, source_db_label, source_table, source_pk),
    CHECK (mapping_status IN ('pending', 'mapped', 'conflict', 'ignored'))
);

CREATE INDEX IF NOT EXISTS idx_legacy_attempt_mappings_user_course
    ON legacy_attempt_mappings(target_user_course_id)
    WHERE target_user_course_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_legacy_attempt_mappings_attempt
    ON legacy_attempt_mappings(target_exercise_attempt_id)
    WHERE target_exercise_attempt_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS legacy_merge_conflicts (
    id BIGSERIAL PRIMARY KEY,
    conflict_type TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'warning',
    source_app_code TEXT,
    source_db_label TEXT,
    source_table TEXT,
    source_pk TEXT,
    stable_identity_type TEXT,
    stable_identity_value TEXT,
    conflict_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    resolution_status TEXT NOT NULL DEFAULT 'open',
    resolution_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMPTZ,
    CHECK (severity IN ('info', 'warning', 'error')),
    CHECK (resolution_status IN ('open', 'resolved', 'ignored'))
);

CREATE INDEX IF NOT EXISTS idx_legacy_merge_conflicts_status
    ON legacy_merge_conflicts(resolution_status, severity, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_legacy_merge_conflicts_identity
    ON legacy_merge_conflicts(stable_identity_type, stable_identity_value)
    WHERE stable_identity_type IS NOT NULL AND stable_identity_value IS NOT NULL;
