-- Seed Spanish course NPC Ines Robles, a warm therapist for A2 conversation quests.
-- Applies only to es_ru; Ines speaks Spanish and all scenarios are pinned to A2.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('therapy_ines_first_session',  'A2', 'Рассказать, как себя чувствуешь'),
        ('therapy_ines_family_balance', 'A2', 'Описать семью или близких'),
        ('therapy_ines_bad_day',        'A2', 'Рассказать о плохом дне'),
        ('therapy_ines_say_no',         'A2', 'Потренироваться вежливо отказывать'),
        ('therapy_ines_future_plan',    'A2', 'Обсудить план изменений'),
        ('therapy_ines_old_photo',      'A2', 'Рассказать о чувствах из-за старого фото')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Ines Robles therapy quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('therapy_ines_first_session', 'community_center', 'A2', 'Рассказать, как себя чувствуешь',
            'Inés Robles', 'ines_therapist', '',
            'a warm, calm community-center therapist who speaks clear A2 Spanish, asks gentle questions, and helps the learner name simple feelings without pressure',
            'The learner comes to a first short session at the community center. You welcome them warmly, explain that this is a safe and simple conversation, and ask how they feel today.',
            true, 30, 40000, 0),
        ('therapy_ines_family_balance', 'community_center', 'A2', 'Описать семью или близких',
            'Inés Robles', 'ines_therapist', 'therapy_ines_first_session',
            'a warm community-center therapist who speaks clear A2 Spanish and helps the learner describe close people in simple, respectful language',
            'The learner returns for another session. You ask about their family or close people and invite them to describe who gives them support and who needs their time.',
            true, 30, 40000, 1),
        ('therapy_ines_bad_day', 'community_center', 'A2', 'Рассказать о плохом дне',
            'Inés Robles', 'ines_therapist', 'therapy_ines_family_balance',
            'a patient therapist who speaks clear A2 Spanish, listens carefully, and helps the learner tell a simple story about a difficult day',
            'The learner has had a bad day. You ask simple follow-up questions about what happened, how they felt, and what helped even a little.',
            true, 30, 40000, 2),
        ('therapy_ines_say_no', 'community_center', 'A2', 'Потренироваться вежливо отказывать',
            'Inés Robles', 'ines_therapist', 'therapy_ines_bad_day',
            'a supportive therapist who speaks clear A2 Spanish and helps the learner practise polite boundaries in everyday situations',
            'The learner wants to practise saying no politely. You offer a safe role-play with a simple request, then help them answer calmly and respectfully.',
            true, 30, 40000, 3),
        ('therapy_ines_future_plan', 'community_center', 'A2', 'Обсудить план изменений',
            'Inés Robles', 'ines_therapist', 'therapy_ines_say_no',
            'a practical, encouraging therapist who speaks clear A2 Spanish and helps the learner turn one small wish into a realistic plan',
            'The learner wants to change one small habit or part of daily life. You help them choose one step, a time, and a simple way to check progress.',
            true, 30, 40000, 4),
        ('therapy_ines_old_photo', 'community_center', 'A2', 'Рассказать о чувствах из-за старого фото',
            'Inés Robles', 'ines_therapist', 'therapy_ines_future_plan',
            'a warm therapist who speaks clear A2 Spanish, handles memories gently, and helps the learner describe feelings connected to an old photo',
            'The learner brings or mentions an old photo. You ask what they see, who or what they remember, and what feeling the photo brings today.',
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
        -- therapy_ines_first_session
        ('therapy_ines_first_session', 'greet',       0, true,  'Поздороваться с Инес',
            'The learner greets Ines and starts the session politely.'),
        ('therapy_ines_first_session', 'name_feeling',1, true,  'Назвать своё чувство',
            'The learner names or describes how they feel today.'),
        ('therapy_ines_first_session', 'say_reason',  2, true,  'Сказать простую причину',
            'The learner gives a simple reason or situation connected to that feeling.'),
        ('therapy_ines_first_session', 'answer_need', 3, true,  'Сказать, что сейчас нужно',
            'The learner says what they need, want, or hope for in the conversation.'),
        ('therapy_ines_first_session', 'thank',       4, false, 'Поблагодарить',
            'The learner thanks Ines and/or closes the session politely.'),

        -- therapy_ines_family_balance
        ('therapy_ines_family_balance', 'greet',          0, true,  'Поздороваться с Инес',
            'The learner greets Ines and continues the therapy conversation.'),
        ('therapy_ines_family_balance', 'name_people',    1, true,  'Назвать близких людей',
            'The learner names or identifies family members or close people.'),
        ('therapy_ines_family_balance', 'describe_person',2, true,  'Описать одного человека',
            'The learner describes one close person with simple personal details.'),
        ('therapy_ines_family_balance', 'say_balance',    3, true,  'Сказать, с кем легко или трудно',
            'The learner says which relationship feels easy, difficult, supportive, or stressful.'),
        ('therapy_ines_family_balance', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Ines and/or closes the session politely.'),

        -- therapy_ines_bad_day
        ('therapy_ines_bad_day', 'greet',        0, true,  'Поздороваться с Инес',
            'The learner greets Ines and agrees to talk about the day.'),
        ('therapy_ines_bad_day', 'say_problem',  1, true,  'Сказать, что случилось',
            'The learner describes the main difficult event or problem from the day.'),
        ('therapy_ines_bad_day', 'say_feeling',  2, true,  'Описать чувство',
            'The learner says how the event made them feel.'),
        ('therapy_ines_bad_day', 'say_helped',   3, true,  'Сказать, что немного помогло',
            'The learner names something that helped, could help, or made the day less difficult.'),
        ('therapy_ines_bad_day', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Ines and/or closes the session politely.'),

        -- therapy_ines_say_no
        ('therapy_ines_say_no', 'greet',           0, true,  'Поздороваться с Инес',
            'The learner greets Ines and agrees to practise a boundary.'),
        ('therapy_ines_say_no', 'describe_request',1, true,  'Описать просьбу',
            'The learner describes a request or situation where they need to refuse.'),
        ('therapy_ines_say_no', 'refuse_politely', 2, true,  'Вежливо отказать',
            'The learner gives a polite refusal without accepting the request.'),
        ('therapy_ines_say_no', 'give_reason',     3, true,  'Коротко объяснить причину',
            'The learner gives a short, respectful reason or boundary.'),
        ('therapy_ines_say_no', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks Ines and/or closes the role-play politely.'),

        -- therapy_ines_future_plan
        ('therapy_ines_future_plan', 'greet',       0, true,  'Поздороваться с Инес',
            'The learner greets Ines and says they want to discuss a change.'),
        ('therapy_ines_future_plan', 'choose_change',1, true, 'Выбрать одно изменение',
            'The learner names one small change, habit, or goal they want to work on.'),
        ('therapy_ines_future_plan', 'say_reason',  2, true,  'Объяснить, зачем это нужно',
            'The learner explains why the change matters to them.'),
        ('therapy_ines_future_plan', 'plan_step',   3, true,  'Назвать первый шаг',
            'The learner proposes a first small step, time, or action for the plan.'),
        ('therapy_ines_future_plan', 'thank',       4, false, 'Поблагодарить',
            'The learner thanks Ines and/or closes the session politely.'),

        -- therapy_ines_old_photo
        ('therapy_ines_old_photo', 'greet',        0, true,  'Поздороваться с Инес',
            'The learner greets Ines and agrees to talk about an old photo.'),
        ('therapy_ines_old_photo', 'describe_photo',1, true, 'Описать фото',
            'The learner describes what is in the old photo or who is connected to it.'),
        ('therapy_ines_old_photo', 'say_memory',   2, true,  'Рассказать воспоминание',
            'The learner shares a simple memory connected to the photo.'),
        ('therapy_ines_old_photo', 'say_feeling',  3, true,  'Описать чувство сейчас',
            'The learner describes the feeling the photo brings now.'),
        ('therapy_ines_old_photo', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Ines and/or closes the session politely.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'es_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
