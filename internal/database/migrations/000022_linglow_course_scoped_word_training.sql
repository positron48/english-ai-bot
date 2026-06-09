-- Path A (keep the existing word-training engine + SRS algorithm unchanged): make the
-- legacy per-user and content tables course-aware so multiple courses can coexist in the
-- unified Linglow DB. The engine keeps reading models.UserCard/models.TrainingCard; we only
-- add a course dimension to scope queries by the user's currently selected course.
--
-- STRICTLY ADDITIVE / SAFE ON LIVE PROD:
--   * Only nullable columns and IF NOT EXISTS indexes are added here.
--   * No UNIQUE constraint is changed and nothing is dropped.
--   * On single-course prod DBs (English / Spanish) these columns stay NULL until a
--     zero-touch startup backfill tags them with the instance course; the engine treats
--     NULL course_code as "matches current course" during that transition.
--   * The word_cards UNIQUE(word) -> UNIQUE(word, course_code) change and the per-course
--     content re-import are intentionally NOT done here: they apply only to the unified DB
--     and are handled by the merge tooling, so prod is never touched destructively.

-- Content tables -------------------------------------------------------------
ALTER TABLE word_cards ADD COLUMN IF NOT EXISTS course_code TEXT;
ALTER TABLE training_cards ADD COLUMN IF NOT EXISTS course_code TEXT;

-- Per-user progress / history tables -----------------------------------------
ALTER TABLE user_cards ADD COLUMN IF NOT EXISTS course_code TEXT;
ALTER TABLE review_events ADD COLUMN IF NOT EXISTS course_code TEXT;
ALTER TABLE user_word_mastering ADD COLUMN IF NOT EXISTS course_code TEXT;
ALTER TABLE user_word_knowledge ADD COLUMN IF NOT EXISTS course_code TEXT;

-- Indexes for course-scoped engine lookups (mirror existing user_id indexes with course).
CREATE INDEX IF NOT EXISTS idx_user_cards_course_due
    ON user_cards (user_id, course_code, next_due_at);

CREATE INDEX IF NOT EXISTS idx_user_cards_course_state
    ON user_cards (user_id, course_code, state);

CREATE INDEX IF NOT EXISTS idx_word_cards_course
    ON word_cards (course_code);

CREATE INDEX IF NOT EXISTS idx_training_cards_course
    ON training_cards (course_code);

CREATE INDEX IF NOT EXISTS idx_review_events_course
    ON review_events (user_id, course_code);
