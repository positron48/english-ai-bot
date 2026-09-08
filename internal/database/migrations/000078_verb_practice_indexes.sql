-- Durable verb practice resumes the most recent session and aggregates a user's
-- rule evidence. Keep the existing JSON columns and historical rows unchanged.
CREATE INDEX IF NOT EXISTS idx_verb_practice_sessions_user
    ON verb_training_sessions(user_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_verb_practice_events_user
    ON verb_review_events(user_id, answered_at DESC);
