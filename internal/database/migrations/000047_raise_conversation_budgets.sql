-- Raise conversation turn/token budgets. The original seed used token_budget=6000, but each turn
-- replays the system prompt + recent history (~1-1.5k tokens), so a quest could exhaust the budget
-- and end after only ~4-6 messages ("as if there were a message counter"). Bump every existing
-- scenario to a generous floor so normal quests run to a natural finish.

UPDATE conversation_scenarios
SET max_turns = GREATEST(max_turns, 30),
    token_budget = GREATEST(token_budget, 40000),
    updated_at = CURRENT_TIMESTAMP
WHERE token_budget < 40000 OR max_turns < 30;
