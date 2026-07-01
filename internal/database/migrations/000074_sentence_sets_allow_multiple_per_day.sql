-- Manual admin generation can create more than one sentence set for the same
-- user/course/date. The automatic daily guard remains in application code.
ALTER TABLE sentence_sets
    DROP CONSTRAINT IF EXISTS sentence_sets_user_id_course_code_generation_date_key;
