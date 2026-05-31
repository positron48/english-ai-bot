ALTER TABLE grammar_test_attempts
    ADD COLUMN IF NOT EXISTS client_attempt_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_grammar_test_attempts_client_attempt_id
    ON grammar_test_attempts(user_id, client_attempt_id)
    WHERE client_attempt_id IS NOT NULL;
