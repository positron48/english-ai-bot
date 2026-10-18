ALTER TABLE conversation_sessions ADD COLUMN daily_journey_step_id TEXT REFERENCES daily_journey_steps(id) ON DELETE CASCADE;
ALTER TABLE picture_quest_sessions ADD COLUMN daily_journey_step_id TEXT REFERENCES daily_journey_steps(id) ON DELETE CASCADE;
DROP INDEX idx_conv_sessions_open;
CREATE UNIQUE INDEX idx_conv_sessions_open ON conversation_sessions(user_course_id,scenario_id) WHERE status='open' AND daily_journey_step_id IS NULL;
DROP INDEX idx_picture_quest_sessions_open;
CREATE UNIQUE INDEX idx_picture_quest_sessions_open ON picture_quest_sessions(user_course_id,quest_id) WHERE status='open' AND daily_journey_step_id IS NULL;
CREATE UNIQUE INDEX conversation_session_daily_step ON conversation_sessions(daily_journey_step_id) WHERE daily_journey_step_id IS NOT NULL;
CREATE UNIQUE INDEX picture_session_daily_step ON picture_quest_sessions(daily_journey_step_id) WHERE daily_journey_step_id IS NOT NULL;
