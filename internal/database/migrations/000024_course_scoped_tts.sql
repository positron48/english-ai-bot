-- TTS cache/status must be isolated by course in the unified Linglow database.
ALTER TABLE tts_generation_status
    ADD COLUMN IF NOT EXISTS course_code TEXT NOT NULL DEFAULT '';

ALTER TABLE tts_generation_status
    DROP CONSTRAINT IF EXISTS tts_generation_status_pkey;

ALTER TABLE tts_generation_status
    ADD PRIMARY KEY (course_code, word);

DROP INDEX IF EXISTS idx_tts_generation_status_state;
DROP INDEX IF EXISTS idx_tts_generation_status_updated_at;

CREATE INDEX IF NOT EXISTS idx_tts_generation_status_course_state
    ON tts_generation_status (course_code, state);
CREATE INDEX IF NOT EXISTS idx_tts_generation_status_course_updated
    ON tts_generation_status (course_code, updated_at);
