-- Seed Iris Lane, an English-only noir detective NPC for the en_ru course.
-- All scenarios are attached to the B1 conversation level as a linked quest chain.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for Iris's quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('detective_street_question',  'B1', 'Ответить, кого или что видел на улице'),
        ('detective_describe_place',   'B1', 'Описать место'),
        ('detective_follow_clue',      'B1', 'Объяснить маршрут по подсказкам'),
        ('detective_compare_versions', 'B1', 'Сравнить две версии событий'),
        ('detective_hidden_motive',    'B1', 'Предположить мотив'),
        ('detective_rainy_confession', 'B1', 'Обсудить финальный поворот')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Iris Lane scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('detective_street_question', 'street', 'B1', 'Ответить, кого или что видел на улице',
            'Iris Lane', 'iris_detective', '',
            'a noir-tinged passerby detective in a long coat; she speaks English only, asks observant B1-level questions, and investigates small street mysteries without naming any famous fictional detective',
            'The learner meets Iris on a wet evening street where something odd happened earlier. You ask what they saw and keep the mood curious, calm, and safe.',
            true, 30, 40000, 0),
        ('detective_describe_place', 'street', 'B1', 'Описать место',
            'Iris Lane', 'iris_detective', 'detective_street_question',
            'a careful street detective with a dry sense of humor; she speaks English only and helps the learner describe places and surroundings in clear B1 English',
            'Iris leads the learner to the spot that matters. You ask them to describe the place, the layout, the light, and anything that feels important.',
            true, 30, 40000, 1),
        ('detective_follow_clue', 'street', 'B1', 'Объяснить маршрут по подсказкам',
            'Iris Lane', 'iris_detective', 'detective_describe_place',
            'a patient noir street investigator who speaks English only and guides the learner through directions, sequence, and simple route explanations',
            'Iris has a few clues about where someone went. You ask the learner to explain the route step by step using the hints she gives.',
            true, 30, 40000, 2),
        ('detective_compare_versions', 'street', 'B1', 'Сравнить две версии событий',
            'Iris Lane', 'iris_detective', 'detective_follow_clue',
            'a reflective detective of small city mysteries; she speaks English only and invites B1 reasoning about two different accounts of the same event',
            'Two people tell Iris different stories about what happened on the street. You ask the learner to compare the versions and notice what does not match.',
            true, 30, 40000, 3),
        ('detective_hidden_motive', 'street', 'B1', 'Предположить мотив',
            'Iris Lane', 'iris_detective', 'detective_compare_versions',
            'a calm noir detective who values careful thinking over drama; she speaks English only and encourages the learner to guess why someone acted a certain way',
            'The facts are clearer now, but the reason is still unclear. Iris asks the learner to suggest a possible motive and support it with what they know.',
            true, 30, 40000, 4),
        ('detective_rainy_confession', 'street', 'B1', 'Обсудить финальный поворот',
            'Iris Lane', 'iris_detective', 'detective_hidden_motive',
            'a poetic street detective who speaks English only and treats the case as a mystery about memory, weather, and honest mistakes rather than danger',
            'Rain starts again and the truth comes out in a quiet confession. Iris asks the learner to discuss the final twist and what it changes about the story.',
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
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
JOIN learning_items li
    ON li.course_id = c.id AND li.source_kind = 'conversation_scenario' AND li.source_id = s.code
ON CONFLICT (course_id, code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3. Quest tasks for Iris's scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- detective_street_question
        ('detective_street_question', 'greet_iris',         0, true,  'Поздороваться с Айрис',
            'The learner greets Iris or acknowledges the meeting.'),
        ('detective_street_question', 'say_what_seen',      1, true,  'Сказать, кого или что видел',
            'The learner describes who or what they saw on the street.'),
        ('detective_street_question', 'give_time_or_place', 2, true,  'Уточнить время или место',
            'The learner gives a relevant time, place, or sequence detail.'),
        ('detective_street_question', 'answer_followup',    3, true,  'Ответить на уточняющий вопрос',
            'The learner answers a follow-up question about the observation.'),
        ('detective_street_question', 'agree_to_help',      4, false, 'Согласиться помочь',
            'The learner agrees to help Iris continue the investigation or shows interest in the mystery.'),

        -- detective_describe_place
        ('detective_describe_place', 'greet_iris',          0, true,  'Поздороваться с Айрис',
            'The learner greets Iris or continues the conversation politely.'),
        ('detective_describe_place', 'describe_place',      1, true,  'Описать место',
            'The learner describes the place or the street location under discussion.'),
        ('detective_describe_place', 'describe_details',    2, true,  'Добавить детали',
            'The learner adds visual or spatial details such as light, objects, layout, or atmosphere.'),
        ('detective_describe_place', 'notice_something_odd', 3, true,  'Заметить необычное',
            'The learner points out something unusual, missing, changed, or worth attention.'),
        ('detective_describe_place', 'ask_next_step',       4, false, 'Спросить о следующем шаге',
            'The learner asks what Iris wants to check next or suggests continuing the search.'),

        -- detective_follow_clue
        ('detective_follow_clue', 'greet_iris',           0, true,  'Поздороваться с Айрис',
            'The learner greets Iris or reconnects with her investigation.'),
        ('detective_follow_clue', 'restate_clue',         1, true,  'Пересказать подсказку',
            'The learner restates or confirms one of the clues Iris mentioned.'),
        ('detective_follow_clue', 'explain_route',        2, true,  'Объяснить маршрут',
            'The learner explains the route or path step by step.'),
        ('detective_follow_clue', 'name_landmarks',       3, true,  'Назвать ориентиры',
            'The learner names landmarks, turns, or stops along the route.'),
        ('detective_follow_clue', 'confirm_direction',    4, false, 'Подтвердить направление',
            'The learner confirms the direction, corrects a step, or asks Iris to verify the route.'),

        -- detective_compare_versions
        ('detective_compare_versions', 'greet_iris',            0, true,  'Поздороваться с Айрис',
            'The learner greets Iris or resumes the discussion.'),
        ('detective_compare_versions', 'state_first_version',   1, true,  'Пересказать первую версию',
            'The learner summarizes the first account of what happened.'),
        ('detective_compare_versions', 'state_second_version',  2, true,  'Пересказать вторую версию',
            'The learner summarizes the second account of what happened.'),
        ('detective_compare_versions', 'compare_difference',    3, true,  'Сравнить отличия',
            'The learner compares the two versions and notes a difference or contradiction.'),
        ('detective_compare_versions', 'judge_believability',   4, false, 'Оценить правдоподобность',
            'The learner says which version seems more believable, incomplete, or doubtful.'),

        -- detective_hidden_motive
        ('detective_hidden_motive', 'greet_iris',              0, true,  'Поздороваться с Айрис',
            'The learner greets Iris or acknowledges the new question.'),
        ('detective_hidden_motive', 'name_person_or_situation', 1, true,  'Назвать человека или ситуацию',
            'The learner identifies the person or situation whose motive is being discussed.'),
        ('detective_hidden_motive', 'suggest_motive',          2, true,  'Предложить мотив',
            'The learner suggests a possible motive or reason for the action.'),
        ('detective_hidden_motive', 'give_reason',             3, true,  'Объяснить, почему так думает',
            'The learner explains why that motive seems plausible based on the evidence.'),
        ('detective_hidden_motive', 'consider_alternative',    4, false, 'Предложить другую версию',
            'The learner considers another possible motive or admits the reason could be uncertain.'),

        -- detective_rainy_confession
        ('detective_rainy_confession', 'greet_iris',         0, true,  'Поздороваться с Айрис',
            'The learner greets Iris or reacts to the final revelation.'),
        ('detective_rainy_confession', 'react_to_twist',     1, true,  'Отреагировать на поворот',
            'The learner reacts to the final twist or confession.'),
        ('detective_rainy_confession', 'explain_final_turn', 2, true,  'Объяснить финальный поворот',
            'The learner explains what the final twist changes about the story.'),
        ('detective_rainy_confession', 'discuss_stakes',     3, true,  'Обсудить последствия',
            'The learner discusses feelings, consequences, or why the truth matters.'),
        ('detective_rainy_confession', 'suggest_closure',    4, false, 'Предложить, что делать дальше',
            'The learner suggests how Iris or the people involved should move forward after the confession.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'en_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
