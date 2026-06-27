-- Seed Carmen Sombra, a Spanish-only noir detective NPC for the es_ru course.
-- All scenarios are attached to the first B1 conversation level, including the
-- later quests that invite B2-style reasoning.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for Carmen's quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('detective_carmen_first_clue',    'B1', 'Описать первую подсказку'),
        ('detective_carmen_wall',          'B1', 'Описать рисунок на стене'),
        ('detective_carmen_alibi',         'B1', 'Объяснить, где был человек'),
        ('detective_carmen_symbol',        'B1', 'Обсудить значение символа'),
        ('detective_carmen_false_lead',    'B1', 'Признать ошибочную версию'),
        ('detective_carmen_mural_returns', 'B1', 'Понять, почему мурал исчез')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Carmen Sombra scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('detective_carmen_first_clue', 'street', 'B1', 'Описать первую подсказку',
            'Carmen Sombra', 'carmen_detective', '',
            'a noir-tinged passerby detective in a coat and brimmed hat; she speaks Spanish only, asks observant B1-level questions, and investigates the disappearance of a city mural rather than a crime',
            'The learner meets Carmen on a dim street near a blank wall where a familiar mural used to be. You ask what they saw earlier and keep the mood curious, poetic, and safe.',
            true, 30, 40000, 0),
        ('detective_carmen_wall', 'street', 'B1', 'Описать рисунок на стене',
            'Carmen Sombra', 'carmen_detective', 'detective_carmen_first_clue',
            'a careful street detective with a dry sense of humor; she speaks Spanish only and helps the learner describe visual details without naming any famous detective',
            'Carmen has found an old photo of the mural. You ask the learner to describe the wall, the colors, the figures, and anything that now feels missing.',
            true, 30, 40000, 1),
        ('detective_carmen_alibi', 'street', 'B1', 'Объяснить, где был человек',
            'Carmen Sombra', 'carmen_detective', 'detective_carmen_wall',
            'a patient noir street investigator who speaks Spanish only and guides the learner through time, place, and simple explanations',
            'A street vendor may know something about the night the mural vanished. Carmen asks the learner to explain where the person was and why that matters.',
            true, 30, 40000, 2),
        ('detective_carmen_symbol', 'street', 'B1', 'Обсудить значение символа',
            'Carmen Sombra', 'carmen_detective', 'detective_carmen_alibi',
            'a reflective detective of small city mysteries; she speaks Spanish only and invites B1-B2 reasoning about symbols, memory, and public art',
            'Carmen notices that one small symbol from the mural remains on the wall. You ask the learner what it could mean and how it might connect to the missing mural.',
            true, 30, 40000, 3),
        ('detective_carmen_false_lead', 'street', 'B1', 'Признать ошибочную версию',
            'Carmen Sombra', 'carmen_detective', 'detective_carmen_symbol',
            'a calm noir detective who values honesty over drama; she speaks Spanish only and encourages the learner to revise a theory without shame',
            'A promising clue turns out not to explain the mural at all. Carmen asks the learner to admit what was wrong in the earlier theory and suggest a better direction.',
            true, 30, 40000, 4),
        ('detective_carmen_mural_returns', 'street', 'B1', 'Понять, почему мурал исчез',
            'Carmen Sombra', 'carmen_detective', 'detective_carmen_false_lead',
            'a poetic street detective who speaks Spanish only and treats the case as a mystery about art, weather, memory, and neighborhood care rather than wrongdoing',
            'The mural begins to return after rain reveals a protected layer of paint. Carmen asks the learner to explain why it seemed to disappear and what the city should do next.',
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
-- 3. Quest tasks for Carmen's scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- detective_carmen_first_clue
        ('detective_carmen_first_clue', 'greet_carmen',       0, true,  'Поздороваться с Кармен',
            'The learner greets Carmen or acknowledges the meeting.'),
        ('detective_carmen_first_clue', 'say_what_seen',      1, true,  'Сказать, что видел',
            'The learner describes what they saw near the mural or the wall.'),
        ('detective_carmen_first_clue', 'give_time_or_place', 2, true,  'Уточнить время или место',
            'The learner gives a relevant time, place, or sequence detail.'),
        ('detective_carmen_first_clue', 'answer_followup',    3, true,  'Ответить на уточняющий вопрос',
            'The learner answers a follow-up question about the observation.'),
        ('detective_carmen_first_clue', 'agree_to_help',      4, false, 'Согласиться помочь',
            'The learner agrees to help Carmen continue the investigation or shows interest in the mystery.'),

        -- detective_carmen_wall
        ('detective_carmen_wall', 'greet_carmen',       0, true,  'Поздороваться с Кармен',
            'The learner greets Carmen or continues the conversation politely.'),
        ('detective_carmen_wall', 'describe_wall',      1, true,  'Описать стену',
            'The learner describes the wall or the place where the mural was.'),
        ('detective_carmen_wall', 'describe_mural',     2, true,  'Описать рисунок',
            'The learner describes the mural by color, shape, figure, mood, or another visual detail.'),
        ('detective_carmen_wall', 'notice_difference',  3, true,  'Заметить отличие',
            'The learner explains what seems different, missing, damaged, covered, or changed.'),
        ('detective_carmen_wall', 'ask_next_step',      4, false, 'Спросить о следующем шаге',
            'The learner asks what Carmen wants to check next or suggests continuing the search.'),

        -- detective_carmen_alibi
        ('detective_carmen_alibi', 'greet_carmen',      0, true,  'Поздороваться с Кармен',
            'The learner greets Carmen or reconnects with her investigation.'),
        ('detective_carmen_alibi', 'name_person',       1, true,  'Назвать человека',
            'The learner identifies the person being discussed or describes their role in the scene.'),
        ('detective_carmen_alibi', 'explain_location',  2, true,  'Объяснить, где он был',
            'The learner explains where the person was during the relevant time.'),
        ('detective_carmen_alibi', 'explain_reason',    3, true,  'Объяснить причину',
            'The learner gives a reason, context, or evidence that supports the location explanation.'),
        ('detective_carmen_alibi', 'evaluate_alibi',    4, false, 'Оценить версию',
            'The learner says whether the explanation seems believable, useful, incomplete, or doubtful.'),

        -- detective_carmen_symbol
        ('detective_carmen_symbol', 'greet_carmen',       0, true,  'Поздороваться с Кармен',
            'The learner greets Carmen or resumes the discussion.'),
        ('detective_carmen_symbol', 'describe_symbol',    1, true,  'Описать символ',
            'The learner describes the remaining symbol or mark in observable terms.'),
        ('detective_carmen_symbol', 'suggest_meaning',    2, true,  'Предложить значение',
            'The learner suggests what the symbol might mean.'),
        ('detective_carmen_symbol', 'connect_to_mural',   3, true,  'Связать символ с муралом',
            'The learner explains a possible connection between the symbol and the missing mural.'),
        ('detective_carmen_symbol', 'consider_alternative',4, false, 'Предложить другую версию',
            'The learner considers another possible meaning or admits that the symbol could be ambiguous.'),

        -- detective_carmen_false_lead
        ('detective_carmen_false_lead', 'greet_carmen',      0, true,  'Поздороваться с Кармен',
            'The learner greets Carmen or acknowledges the new evidence.'),
        ('detective_carmen_false_lead', 'state_old_theory',  1, true,  'Сформулировать старую версию',
            'The learner summarizes the earlier theory or assumption.'),
        ('detective_carmen_false_lead', 'admit_error',       2, true,  'Признать ошибку',
            'The learner says that the earlier theory was wrong, weak, or incomplete.'),
        ('detective_carmen_false_lead', 'explain_why_wrong', 3, true,  'Объяснить, почему версия не подходит',
            'The learner explains why the earlier theory does not fit the evidence.'),
        ('detective_carmen_false_lead', 'suggest_new_path',  4, false, 'Предложить новый путь',
            'The learner suggests a revised direction, next question, or better clue to follow.'),

        -- detective_carmen_mural_returns
        ('detective_carmen_mural_returns', 'greet_carmen',       0, true,  'Поздороваться с Кармен',
            'The learner greets Carmen or reacts to seeing the mural again.'),
        ('detective_carmen_mural_returns', 'notice_return',      1, true,  'Заметить возвращение мурала',
            'The learner describes what has reappeared or changed on the wall.'),
        ('detective_carmen_mural_returns', 'explain_disappear',  2, true,  'Объяснить исчезновение',
            'The learner explains why the mural seemed to disappear.'),
        ('detective_carmen_mural_returns', 'discuss_reason',     3, true,  'Обсудить настоящую причину',
            'The learner discusses a non-criminal cause such as weather, material, cleaning, repair, or neighborhood action.'),
        ('detective_carmen_mural_returns', 'propose_care',       4, false, 'Предложить, как сохранить мурал',
            'The learner suggests how people could protect, remember, or care for the mural in the future.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'es_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
