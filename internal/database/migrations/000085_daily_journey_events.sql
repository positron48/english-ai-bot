-- Route lifecycle only: do not copy exercise payloads, answers or chat messages.
CREATE TABLE daily_journey_events (
    id BIGSERIAL PRIMARY KEY,
    journey_id TEXT NOT NULL REFERENCES daily_journeys(id) ON DELETE CASCADE,
    step_id TEXT REFERENCES daily_journey_steps(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('journey_started','step_completed','step_skipped','step_unavailable','journey_completed','journey_unavailable')),
    revision BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX daily_journey_event_once ON daily_journey_events(journey_id,event_type,COALESCE(step_id,''));
CREATE INDEX daily_journey_events_created ON daily_journey_events(created_at,event_type);
