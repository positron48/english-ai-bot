-- Dedicated circuit breaker for pronunciation/TTS provider calls, separate from the
-- existing AI word-card generation breaker (circuit_breaker_state). Provider-level
-- failures (e.g. OpenRouter "insufficient balance") should stop further network
-- calls instead of repeatedly hitting a known-broken provider for every word.
CREATE TABLE IF NOT EXISTS tts_circuit_breaker_state (
    id INTEGER PRIMARY KEY,
    is_open INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    last_failure_at TIMESTAMPTZ,
    last_failure_message TEXT,
    last_reset_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO tts_circuit_breaker_state (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
