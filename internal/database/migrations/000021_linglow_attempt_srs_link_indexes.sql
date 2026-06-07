-- Speed up Linglow attempt SRS link backfill on large exercise_attempts tables.
CREATE INDEX IF NOT EXISTS idx_exercise_attempts_unlinked_by_mode
    ON exercise_attempts (user_course_id, mode)
    WHERE srs_item_id IS NULL AND mode IN ('word_training', 'grammar_training');

CREATE INDEX IF NOT EXISTS idx_exercise_attempts_review_events_source
    ON exercise_attempts (source_pk)
    WHERE source_table = 'review_events' AND mode = 'word_training';

CREATE INDEX IF NOT EXISTS idx_exercise_attempts_grammar_attempts_source
    ON exercise_attempts (source_pk)
    WHERE source_table = 'grammar_attempts' AND mode = 'grammar_training';
