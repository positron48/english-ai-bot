-- Word sets: add explicit CEFR level binding (A0-C1) to categories and sets.
-- A word set without its own level_code inherits the level of its category
-- (resolved in application code, not enforced here).

ALTER TABLE word_set_categories ADD COLUMN IF NOT EXISTS level_code TEXT;
ALTER TABLE word_sets ADD COLUMN IF NOT EXISTS level_code TEXT;

ALTER TABLE word_set_categories DROP CONSTRAINT IF EXISTS word_set_categories_level_code_check;
ALTER TABLE word_set_categories ADD CONSTRAINT word_set_categories_level_code_check
    CHECK (level_code IS NULL OR level_code IN ('A0', 'A1', 'A2', 'B1', 'B2', 'C1'));

ALTER TABLE word_sets DROP CONSTRAINT IF EXISTS word_sets_level_code_check;
ALTER TABLE word_sets ADD CONSTRAINT word_sets_level_code_check
    CHECK (level_code IS NULL OR level_code IN ('A0', 'A1', 'A2', 'B1', 'B2', 'C1'));

CREATE INDEX IF NOT EXISTS idx_word_set_categories_course_level
    ON word_set_categories (course_code, level_code);

CREATE INDEX IF NOT EXISTS idx_word_sets_course_level
    ON word_sets (course_code, level_code);
