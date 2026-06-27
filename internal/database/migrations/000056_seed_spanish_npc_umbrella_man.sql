-- Seed Spanish-only NPC chain: El Hombre del Paraguas.
-- UI titles are Russian; persona/scene/criteria are model-facing English instructions.
-- Although the story spans A1-B1 themes, every scenario is intentionally anchored to
-- the first Spanish course level (A1) and the conversation location.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new Spanish scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('umbrella_man_weather',     'A1', 'Поговорить о погоде'),
        ('umbrella_man_direction',   'A1', 'Спросить дорогу'),
        ('umbrella_man_warning',     'A1', 'Понять простое предупреждение'),
        ('umbrella_man_lost_letter', 'A1', 'Описать найденное письмо'),
        ('umbrella_man_old_promise', 'A1', 'Обсудить обещание из прошлого'),
        ('umbrella_man_sunny_day',   'A1', 'Помочь принять хороший день')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Spanish-only quest scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('umbrella_man_weather', 'street', 'A1', 'Поговорить о погоде',
            'El Hombre del Paraguas', 'umbrella_man', '',
            'a gentle, eccentric Spanish-speaking man who always carries an umbrella, even on sunny days; he speaks slowly, kindly, and concretely, with soft odd remarks but no direct magic, horror, threats, or fear',
            'The learner meets you on a calm street. You hold an umbrella on a sunny day and invite the learner to talk about the weather in simple Spanish.',
            true, 30, 40000, 0),
        ('umbrella_man_direction', 'street', 'A1', 'Спросить дорогу',
            'El Hombre del Paraguas', 'umbrella_man', 'umbrella_man_weather',
            'a gentle, eccentric Spanish-speaking man who gives simple street directions and treats the city like an old friend; he stays warm, safe, and only mildly strange',
            'The learner meets you again near the same street corner. They need directions, and you help with clear, simple landmarks while keeping your umbrella close.',
            true, 30, 40000, 1),
        ('umbrella_man_warning', 'street', 'A1', 'Понять простое предупреждение',
            'El Hombre del Paraguas', 'umbrella_man', 'umbrella_man_direction',
            'a gentle, eccentric Spanish-speaking man who offers harmless practical warnings with poetic but simple wording; he never suggests danger, panic, magic, or fear',
            'The learner is about to choose a route through the street. You give a simple, safe warning about an everyday inconvenience, such as wet paint, a closed gate, or a crowded corner.',
            true, 30, 40000, 2),
        ('umbrella_man_lost_letter', 'street', 'A1', 'Описать найденное письмо',
            'El Hombre del Paraguas', 'umbrella_man', 'umbrella_man_warning',
            'a gentle, eccentric Spanish-speaking man who notices small forgotten things and speaks with quiet curiosity; he keeps the scene safe, human, and grounded',
            'The learner finds an old letter near the street bench. You ask them to describe the letter and decide together where it might belong.',
            true, 30, 40000, 3),
        ('umbrella_man_old_promise', 'street', 'A1', 'Обсудить обещание из прошлого',
            'El Hombre del Paraguas', 'umbrella_man', 'umbrella_man_lost_letter',
            'a gentle, eccentric Spanish-speaking man who remembers an old promise with tenderness, not sadness or fear; he supports B1-style meaning while accepting simple learner language',
            'The letter reminds you of an old promise to meet someone after the rain. You tell the story gently and ask the learner what people can do when a promise is very old.',
            true, 30, 40000, 4),
        ('umbrella_man_sunny_day', 'street', 'A1', 'Помочь принять хороший день',
            'El Hombre del Paraguas', 'umbrella_man', 'umbrella_man_old_promise',
            'a gentle, eccentric Spanish-speaking man who is learning to enjoy good weather without losing his memories; he speaks warmly and keeps the ending hopeful and ordinary',
            'It is a bright, quiet day. You still carry the umbrella, but the learner helps you notice that today can be good even without rain.',
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
-- 3. Tasks for the quest scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- umbrella_man_weather
        ('umbrella_man_weather', 'greet',        0, true,  'Поздороваться с человеком с зонтом',
            'The learner greets the NPC.'),
        ('umbrella_man_weather', 'say_weather',  1, true,  'Сказать, какая погода',
            'The learner describes the current weather in a simple way.'),
        ('umbrella_man_weather', 'ask_umbrella', 2, true,  'Спросить про зонт',
            'The learner asks why the NPC has an umbrella or asks a question about the umbrella.'),
        ('umbrella_man_weather', 'react_weather',3, true,  'Отреагировать на его мысль о погоде',
            'The learner responds to the NPC''s comment about weather with a simple feeling, opinion, or observation.'),
        ('umbrella_man_weather', 'thank',        4, false, 'Поблагодарить и попрощаться',
            'The learner thanks the NPC and/or says goodbye.'),

        -- umbrella_man_direction
        ('umbrella_man_direction', 'greet',           0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('umbrella_man_direction', 'ask_direction',   1, true,  'Спросить дорогу',
            'The learner asks how to get to a place or asks where a place is.'),
        ('umbrella_man_direction', 'name_destination',2, true,  'Назвать место назначения',
            'The learner names or describes the place they want to reach.'),
        ('umbrella_man_direction', 'confirm_route',   3, true,  'Подтвердить маршрут',
            'The learner confirms, repeats, or shows they understood the route.'),
        ('umbrella_man_direction', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks the NPC and/or says goodbye.'),

        -- umbrella_man_warning
        ('umbrella_man_warning', 'greet',             0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('umbrella_man_warning', 'ask_warning',       1, true,  'Спросить, что случилось',
            'The learner asks what the warning is about or why the NPC is stopping them.'),
        ('umbrella_man_warning', 'show_understanding',2, true,  'Показать, что предупреждение понятно',
            'The learner shows they understand the simple warning or repeats the practical point.'),
        ('umbrella_man_warning', 'choose_action',     3, true,  'Выбрать безопасное действие',
            'The learner says what they will do next in response to the warning.'),
        ('umbrella_man_warning', 'thank',             4, false, 'Поблагодарить',
            'The learner thanks the NPC and/or says goodbye.'),

        -- umbrella_man_lost_letter
        ('umbrella_man_lost_letter', 'greet',           0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('umbrella_man_lost_letter', 'mention_letter',  1, true,  'Сказать о найденном письме',
            'The learner says they found a letter or points out the letter.'),
        ('umbrella_man_lost_letter', 'describe_letter', 2, true,  'Описать письмо',
            'The learner describes the letter by appearance, condition, name, date, place, or another visible detail.'),
        ('umbrella_man_lost_letter', 'ask_owner',       3, true,  'Спросить, кому оно принадлежит',
            'The learner asks who the letter belongs to or where it should go.'),
        ('umbrella_man_lost_letter', 'suggest_next',    4, false, 'Предложить следующий шаг',
            'The learner suggests a simple next action for the letter.'),

        -- umbrella_man_old_promise
        ('umbrella_man_old_promise', 'greet',           0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('umbrella_man_old_promise', 'ask_promise',     1, true,  'Спросить об обещании',
            'The learner asks about the old promise or asks what happened in the past.'),
        ('umbrella_man_old_promise', 'summarize_story', 2, true,  'Кратко пересказать историю',
            'The learner summarizes or restates the main idea of the old promise in their own words.'),
        ('umbrella_man_old_promise', 'share_opinion',   3, true,  'Высказать мнение об обещании',
            'The learner gives a simple opinion, feeling, or interpretation about the promise.'),
        ('umbrella_man_old_promise', 'respond_kindly',  4, false, 'Ответить с поддержкой',
            'The learner responds to the NPC in a kind or supportive way.'),

        -- umbrella_man_sunny_day
        ('umbrella_man_sunny_day', 'greet',          0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('umbrella_man_sunny_day', 'notice_day',     1, true,  'Заметить хороший день',
            'The learner says something positive or concrete about the day, weather, street, or moment.'),
        ('umbrella_man_sunny_day', 'invite_action',  2, true,  'Предложить простое действие',
            'The learner suggests a simple, safe action the NPC can do on a good day.'),
        ('umbrella_man_sunny_day', 'reassure_memory',3, true,  'Поддержать его память',
            'The learner shows that enjoying the day does not mean forgetting the past or the old promise.'),
        ('umbrella_man_sunny_day', 'farewell',       4, false, 'Попрощаться тепло',
            'The learner thanks the NPC and/or says goodbye in a warm way.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'es_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
