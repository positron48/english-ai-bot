-- Inline error corrections for conversation turns. The NPC role-play model may, in addition
-- to staying in character, point out mistakes in the learner's latest message. Those corrections
-- are stored on the assistant message that follows the user's message and shown to the learner
-- as a separate block under the reply.
ALTER TABLE conversation_messages ADD COLUMN IF NOT EXISTS corrections JSONB NOT NULL DEFAULT '[]'::jsonb;
