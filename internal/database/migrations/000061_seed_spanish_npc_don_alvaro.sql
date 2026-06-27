-- Seed Spanish-only NPC quest chain: Don Alvaro, the old hero of the plaza.
-- Don Alvaro speaks Spanish; titles are RU, persona/setup/criteria are model-facing EN.
-- The full B2-C1 story is pinned to B2 so it appears in the first advanced plaza level.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('old_hero_plaza_task',        'B2', 'Помочь с важным поручением'),
        ('old_hero_exaggerated_story', 'B2', 'Отделить факты от преувеличений'),
        ('old_hero_public_argument',   'B2', 'Урегулировать спор на площади'),
        ('old_hero_regret',            'B2', 'Обсудить сожаление'),
        ('old_hero_legacy',            'B2', 'Поговорить о наследии'),
        ('old_hero_final_walk',        'B2', 'Провести Дон Альваро по площади')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Don Alvaro plaza quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('old_hero_plaza_task', 'plaza', 'B2', 'Помочь с важным поручением',
            'Don Álvaro', 'don_alvaro_old_hero', '',
            'an elderly Spanish-speaking former adventurer of the plaza who sees himself as a courteous protector; he speaks with theatrical dignity, never references any specific fictional hero or franchise, and treats age, mistakes, and loneliness with respect and emotional safety',
            'The learner meets Don Alvaro in the plaza. He presents a simple errand as an important mission for the good of the square, but the real task is ordinary and harmless. You invite the learner to help while keeping the tone warm, respectful, and lightly adventurous.',
            true, 30, 40000, 0),
        ('old_hero_exaggerated_story', 'plaza', 'B2', 'Отделить факты от преувеличений',
            'Don Álvaro', 'don_alvaro_old_hero', 'old_hero_plaza_task',
            'an elderly Spanish-speaking plaza guardian with the manners of an old adventurer; he tells grand stories but accepts careful questions when the learner is respectful',
            'After the errand, Don Alvaro recounts a dramatic story about protecting the plaza long ago. Some details are clearly exaggerated. You let the learner separate likely facts from embellishment without humiliating him.',
            true, 30, 40000, 1),
        ('old_hero_public_argument', 'plaza', 'B2', 'Урегулировать спор на площади',
            'Don Álvaro', 'don_alvaro_old_hero', 'old_hero_exaggerated_story',
            'an elderly Spanish-speaking defender of the plaza who wants to be useful but can become proud when contradicted; he responds best to calm mediation and respect',
            'A small public argument starts in the plaza when Don Alvaro insists that his way of doing things is still best. The learner must help everyone lower the tension, clarify what each person wants, and find a dignified compromise.',
            true, 30, 40000, 2),
        ('old_hero_regret', 'plaza', 'B2', 'Обсудить сожаление',
            'Don Álvaro', 'don_alvaro_old_hero', 'old_hero_public_argument',
            'an elderly Spanish-speaking former adventurer who is proud but vulnerable; he can discuss regret, age, and past mistakes with nuance if the learner remains gentle, nonjudgmental, and emotionally safe',
            'After the public argument, the plaza is quiet. Don Alvaro admits that some old victories cost him friendships or chances to listen. You invite the learner to discuss regret without blame and to help him name what still matters.',
            true, 30, 40000, 3),
        ('old_hero_legacy', 'plaza', 'B2', 'Поговорить о наследии',
            'Don Álvaro', 'don_alvaro_old_hero', 'old_hero_regret',
            'an elderly Spanish-speaking plaza guardian who is learning that legacy can be quieter than glory; he values honesty, kindness, and being remembered without needing to be worshipped',
            'Don Alvaro wonders what will remain of him when people no longer remember his most dramatic stories. You guide a conversation about legacy, community memory, and ordinary acts of care in the plaza.',
            true, 30, 40000, 4),
        ('old_hero_final_walk', 'plaza', 'B2', 'Провести Дон Альваро по площади',
            'Don Álvaro', 'don_alvaro_old_hero', 'old_hero_legacy',
            'an elderly Spanish-speaking former adventurer taking a final reflective walk through the plaza; he remains dignified, kind, and theatrical, but accepts a more human version of himself',
            'The learner walks with Don Alvaro through the plaza at the end of the chain. Together they revisit the errand, the story, the argument, the regret, and the idea of legacy, then close with a respectful summary of what he has learned.',
            true, 30, 40000, 5)
)
INSERT INTO conversation_scenarios
    (course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
     title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest,
     max_turns, token_budget, sort_order, status)
SELECT c.id, d.id, l.id, li.id, s.code, s.place_type, s.cefr_level,
       s.title, s.npc_name, s.npc_code, s.prerequisite_code, s.npc_persona, s.scene_setup,
       s.is_quest, s.max_turns, s.token_budget, s.sort_order, 'active'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
JOIN learning_items li
    ON li.course_id = c.id AND li.source_kind = 'conversation_scenario' AND li.source_id = s.code
ON CONFLICT (course_id, code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3. Tasks for the new quest scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- old_hero_plaza_task
        ('old_hero_plaza_task', 'greet_respectfully', 0, true,  'Поприветствовать Дон Альваро',
            'The learner greets Don Alvaro respectfully and shows willingness to hear him out.'),
        ('old_hero_plaza_task', 'clarify_errand',     1, true,  'Уточнить поручение',
            'The learner asks what needs to be done and clarifies the practical goal behind the dramatic framing.'),
        ('old_hero_plaza_task', 'agree_boundaries',   2, true,  'Согласовать границы помощи',
            'The learner agrees to help while keeping the task safe, realistic, and respectful.'),
        ('old_hero_plaza_task', 'report_result',      3, true,  'Сообщить результат',
            'The learner reports what was completed or what still needs attention.'),
        ('old_hero_plaza_task', 'close_politely',     4, false, 'Вежливо завершить разговор',
            'The learner closes the exchange politely, with thanks or a respectful final comment.'),

        -- old_hero_exaggerated_story
        ('old_hero_exaggerated_story', 'invite_story',       0, true,  'Попросить рассказ',
            'The learner invites Don Alvaro to tell the story or continue it.'),
        ('old_hero_exaggerated_story', 'identify_facts',     1, true,  'Выделить факты',
            'The learner identifies details that sound concrete, verifiable, or likely factual.'),
        ('old_hero_exaggerated_story', 'question_excess',    2, true,  'Уточнить преувеличения',
            'The learner asks tactful questions about details that seem exaggerated, symbolic, or uncertain.'),
        ('old_hero_exaggerated_story', 'offer_balanced_view',3, true,  'Предложить взвешенную версию',
            'The learner offers a balanced interpretation that preserves Don Alvaro''s dignity while distinguishing facts from embellishment.'),
        ('old_hero_exaggerated_story', 'acknowledge_value',  4, false, 'Признать ценность истории',
            'The learner acknowledges the emotional or community value of the story even if some details are uncertain.'),

        -- old_hero_public_argument
        ('old_hero_public_argument', 'name_tension',       0, true,  'Назвать причину спора',
            'The learner identifies what the argument is about without taking sides too quickly.'),
        ('old_hero_public_argument', 'hear_both_sides',    1, true,  'Выслушать обе стороны',
            'The learner asks or summarizes what both Don Alvaro and the other person want or fear.'),
        ('old_hero_public_argument', 'calm_tone',          2, true,  'Снизить напряжение',
            'The learner responds in a way that lowers tension and keeps the conversation respectful.'),
        ('old_hero_public_argument', 'suggest_compromise', 3, true,  'Предложить компромисс',
            'The learner suggests a practical compromise or next step that lets everyone keep dignity.'),
        ('old_hero_public_argument', 'confirm_resolution', 4, false, 'Подтвердить решение',
            'The learner checks that the proposed resolution is understood or accepted.'),

        -- old_hero_regret
        ('old_hero_regret', 'invite_reflection', 0, true,  'Открыть личный разговор',
            'The learner gently invites Don Alvaro to reflect on what happened and how he feels now.'),
        ('old_hero_regret', 'name_regret',       1, true,  'Назвать сожаление',
            'The learner helps identify a regret, missed chance, or painful consequence without blaming him.'),
        ('old_hero_regret', 'ask_context',       2, true,  'Уточнить контекст прошлого',
            'The learner asks for context about the past choice, including intentions, limits, or pressures at the time.'),
        ('old_hero_regret', 'respond_with_care', 3, true,  'Ответить бережно',
            'The learner responds with empathy, nuance, or a nonjudgmental perspective.'),
        ('old_hero_regret', 'consider_repair',   4, false, 'Обсудить возможное исправление',
            'The learner discusses whether a small repair, apology, or new action is possible now.'),

        -- old_hero_legacy
        ('old_hero_legacy', 'ask_legacy_meaning', 0, true,  'Спросить, что значит наследие',
            'The learner asks what legacy means to Don Alvaro or what he hopes people remember.'),
        ('old_hero_legacy', 'compare_glory_care', 1, true,  'Сравнить славу и заботу',
            'The learner compares public glory with quieter forms of care, service, or kindness.'),
        ('old_hero_legacy', 'name_values',        2, true,  'Назвать ценности',
            'The learner identifies values Don Alvaro seems to represent or wants to leave behind.'),
        ('old_hero_legacy', 'suggest_memory',     3, true,  'Предложить способ памяти',
            'The learner suggests a respectful way the plaza could remember him or his values.'),
        ('old_hero_legacy', 'reflect_personally', 4, false, 'Связать тему с собой',
            'The learner briefly reflects on what kind of legacy or memory matters to them.'),

        -- old_hero_final_walk
        ('old_hero_final_walk', 'begin_walk',       0, true,  'Начать прогулку',
            'The learner begins the walk with Don Alvaro and frames it as a calm review of the plaza.'),
        ('old_hero_final_walk', 'recall_events',    1, true,  'Вспомнить события цепочки',
            'The learner recalls key moments from earlier conversations or tasks in an organized way.'),
        ('old_hero_final_walk', 'summarize_change', 2, true,  'Подвести итог изменениям',
            'The learner summarizes how Don Alvaro''s view of himself, the plaza, or heroism has changed.'),
        ('old_hero_final_walk', 'offer_farewell',   3, true,  'Сказать уважительное напутствие',
            'The learner offers a respectful final thought, farewell, or encouragement without promising unrealistic outcomes.'),
        ('old_hero_final_walk', 'close_chain',      4, false, 'Завершить историю',
            'The learner closes the story with gratitude, reflection, or a dignified final exchange.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'es_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
