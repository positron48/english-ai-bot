ALTER TABLE review_events
    ADD COLUMN IF NOT EXISTS client_attempt_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_review_events_client_attempt_id
    ON review_events(user_id, client_attempt_id)
    WHERE client_attempt_id IS NOT NULL;
