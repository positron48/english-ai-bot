ALTER TABLE conversation_scenarios
    ADD COLUMN IF NOT EXISTS image_url TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS conversation_npcs (
    id         BIGSERIAL PRIMARY KEY,
    course_id  BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    npc_code   TEXT NOT NULL,
    image_url  TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(course_id, npc_code)
);
