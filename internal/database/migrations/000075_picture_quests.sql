-- Picture description quests: the learner describes an admin-uploaded picture in the
-- target language, chatting with Lumi (the mascot). Content is authored in the admin
-- panel (image + admin-written image_description fed to the LLM + tasks).
-- Mirrors the conversation_* table set (000040) but stays fully separate.
-- Each quest gets a backing learning_items row (item_type='speaking_task',
-- source_kind='picture_quest') so district/location progress aggregation lights up.

CREATE TABLE IF NOT EXISTS picture_quests (
    id BIGSERIAL PRIMARY KEY,
    course_id BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    district_id BIGINT REFERENCES districts(id) ON DELETE SET NULL,
    location_id BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    learning_item_id BIGINT REFERENCES learning_items(id) ON DELETE SET NULL,
    code TEXT NOT NULL,
    cefr_level TEXT NOT NULL,
    title TEXT NOT NULL,
    image_url TEXT NOT NULL,
    image_description TEXT NOT NULL,
    max_turns INTEGER NOT NULL DEFAULT 20,
    token_budget INTEGER NOT NULL DEFAULT 15000,
    sort_order INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'draft',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(course_id, code),
    CHECK (status IN ('draft', 'active', 'locked', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_picture_quests_place
    ON picture_quests(course_id, district_id, cefr_level, sort_order)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS picture_quest_tasks (
    id BIGSERIAL PRIMARY KEY,
    quest_id BIGINT NOT NULL REFERENCES picture_quests(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_required BOOLEAN NOT NULL DEFAULT true,
    title TEXT NOT NULL,
    completion_criteria TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE(quest_id, code)
);

CREATE INDEX IF NOT EXISTS idx_picture_quest_tasks_quest
    ON picture_quest_tasks(quest_id, sort_order);

CREATE TABLE IF NOT EXISTS picture_quest_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_course_id BIGINT NOT NULL REFERENCES user_courses(id) ON DELETE CASCADE,
    quest_id BIGINT NOT NULL REFERENCES picture_quests(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'open',
    turn_count INTEGER NOT NULL DEFAULT 0,
    tokens_used INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    CHECK (status IN ('open', 'completed', 'abandoned'))
);

CREATE INDEX IF NOT EXISTS idx_picture_quest_sessions_user
    ON picture_quest_sessions(user_course_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_picture_quest_sessions_open
    ON picture_quest_sessions(user_course_id, quest_id)
    WHERE status = 'open';

CREATE TABLE IF NOT EXISTS picture_quest_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES picture_quest_sessions(id) ON DELETE CASCADE,
    seq INTEGER NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    corrections JSONB,
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, seq),
    CHECK (role IN ('user', 'assistant', 'system'))
);

CREATE TABLE IF NOT EXISTS picture_quest_task_progress (
    session_id BIGINT NOT NULL REFERENCES picture_quest_sessions(id) ON DELETE CASCADE,
    task_id BIGINT NOT NULL REFERENCES picture_quest_tasks(id) ON DELETE CASCADE,
    completed BOOLEAN NOT NULL DEFAULT false,
    completed_at TIMESTAMPTZ,
    completed_in_seq INTEGER,
    PRIMARY KEY (session_id, task_id)
);
