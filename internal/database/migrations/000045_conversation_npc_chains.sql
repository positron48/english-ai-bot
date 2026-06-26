-- NPC chains: group conversation scenarios under a recurring NPC and let them unlock
-- one after another. A scenario with a non-empty prerequisite_code stays locked for a
-- learner until they have completed the scenario named by that code (same course).
--
-- Both columns are nullable-by-default empty strings so existing scenarios keep working
-- unchanged (no NPC grouping, no prerequisite = always unlocked).

ALTER TABLE conversation_scenarios
    ADD COLUMN IF NOT EXISTS npc_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS prerequisite_code TEXT NOT NULL DEFAULT '';

-- Lookups when resolving a chain (an NPC's ordered scenarios) and prerequisite checks.
CREATE INDEX IF NOT EXISTS idx_conv_scenarios_npc
    ON conversation_scenarios(course_id, npc_code, sort_order)
    WHERE npc_code <> '';
