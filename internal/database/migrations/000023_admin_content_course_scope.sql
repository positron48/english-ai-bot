-- Make legacy admin-managed vocabulary content explicitly course-aware.
-- Existing single-course rows are tagged at startup by TagLegacyWordTablesForLearning.
ALTER TABLE word_set_categories ADD COLUMN IF NOT EXISTS course_code TEXT;
ALTER TABLE word_sets ADD COLUMN IF NOT EXISTS course_code TEXT;

CREATE INDEX IF NOT EXISTS idx_word_set_categories_course
    ON word_set_categories (course_code, sort_order, name);

CREATE INDEX IF NOT EXISTS idx_word_sets_course
    ON word_sets (course_code, sort_order, title);
