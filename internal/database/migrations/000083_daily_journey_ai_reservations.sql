-- Durable provider reservation; network I/O is never performed under a journey lock.
ALTER TABLE daily_journey_actions DROP CONSTRAINT daily_journey_actions_status_check;
ALTER TABLE daily_journey_actions ADD CONSTRAINT daily_journey_actions_status_check
 CHECK(status IN ('processing','retryable','completed'));
ALTER TABLE daily_journey_actions ADD COLUMN reservation_token TEXT;
ALTER TABLE daily_journey_actions ADD COLUMN lease_expires_at TIMESTAMPTZ;
ALTER TABLE daily_journey_actions ADD COLUMN prepared_json JSONB;
-- No live worker could have been created by the previous, synchronous API.
UPDATE daily_journey_actions SET status='retryable' WHERE status='processing' AND reservation_token IS NULL;
ALTER TABLE daily_journey_actions ADD CONSTRAINT daily_journey_ai_reservation_shape CHECK (
 (status='processing' AND reservation_token IS NOT NULL AND lease_expires_at IS NOT NULL AND prepared_json IS NOT NULL)
 OR (status<>'processing' AND reservation_token IS NULL AND lease_expires_at IS NULL)
);
CREATE UNIQUE INDEX daily_journey_one_processing_action ON daily_journey_actions(journey_id) WHERE status='processing';
