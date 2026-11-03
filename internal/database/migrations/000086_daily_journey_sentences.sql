-- Sentence composition reuses the existing single-shot sentence attempts.
ALTER TABLE daily_journey_steps DROP CONSTRAINT daily_journey_steps_mode_check;
ALTER TABLE daily_journey_steps ADD CONSTRAINT daily_journey_steps_mode_check
    CHECK (mode IN ('words', 'reading', 'grammar', 'grammar_srs', 'verb_forms', 'sentences', 'picture', 'conversation'));
