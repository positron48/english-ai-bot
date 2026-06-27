-- Extend Mara's A0 cafe chain with a small sequential story before free chat.
-- Applies to en_ru and es_ru; Mara speaks the course target language.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('cafe_choose_seat',    'A0', 'Выбрать место в кафе'),
        ('cafe_wrong_order',    'A0', 'Исправить заказ в кафе'),
        ('cafe_favorite_drink', 'A0', 'Рассказать о любимом напитке'),
        ('cafe_busy_morning',   'A0', 'Помочь в загруженное утро'),
        ('cafe_mara_dream',     'A0', 'Поговорить с Марой о мечте')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code IN ('en_ru', 'es_ru')
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. New Mara scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('cafe_choose_seat', 'cafe', 'A0', 'Выбрать место в кафе',
            'Mara', 'mara_barista', 'cafe_order_pastry',
            'a warm, patient cafe barista who keeps sentences very short and simple; she now recognizes the learner as a regular guest',
            'The learner comes back to the cafe. You are behind the counter. The cafe is quiet, and you can help them choose a place to sit.',
            true, 30, 40000, 2),
        ('cafe_wrong_order', 'cafe', 'A0', 'Исправить заказ в кафе',
            'Mara', 'mara_barista', 'cafe_choose_seat',
            'a warm, patient cafe barista who apologizes kindly and keeps sentences very short and simple',
            'The learner is at their table. You bring the wrong drink by mistake and wait for them to tell you what is wrong.',
            true, 30, 40000, 3),
        ('cafe_favorite_drink', 'cafe', 'A0', 'Рассказать о любимом напитке',
            'Mara', 'mara_barista', 'cafe_wrong_order',
            'a warm, patient cafe barista who likes learning what regular guests enjoy; she keeps sentences very short and simple',
            'The order is fixed. The cafe is calm, and you ask the learner about drinks they like. You share one simple detail about your own favourite drink.',
            true, 30, 40000, 4),
        ('cafe_busy_morning', 'cafe', 'A0', 'Помочь в загруженное утро',
            'Mara', 'mara_barista', 'cafe_favorite_drink',
            'a warm but busy cafe barista who speaks clearly and thanks the learner for small help; she keeps sentences very short and simple',
            'It is a busy morning. You are behind the counter with many orders. You ask the learner simple questions so you can prepare their order quickly.',
            true, 30, 40000, 5),
        ('cafe_mara_dream', 'cafe', 'A0', 'Поговорить с Марой о мечте',
            'Mara', 'mara_barista', 'cafe_busy_morning',
            'a warm, patient cafe barista who dreams of opening a small evening cafe for stories; she keeps sentences simple and encouraging',
            'The morning rush is over. The learner is the last guest at the counter. You tell them, in simple words, that you dream of a small evening cafe with stories.',
            true, 30, 40000, 6)
)
INSERT INTO conversation_scenarios
    (course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
     title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest,
     max_turns, token_budget, sort_order, status)
SELECT c.id, d.id, l.id, li.id, s.code, s.place_type, s.cefr_level,
       s.title, s.npc_name, s.npc_code, s.prerequisite_code, s.npc_persona, s.scene_setup,
       s.is_quest, s.max_turns, s.token_budget, s.sort_order, 'active'
FROM scen s
JOIN courses c ON c.code IN ('en_ru', 'es_ru')
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
JOIN learning_items li
    ON li.course_id = c.id AND li.source_kind = 'conversation_scenario' AND li.source_id = s.code
ON CONFLICT (course_id, code) DO NOTHING;

-- Move the existing free chat to the end of the extended Mara chain. It intentionally remains
-- non-quest, so no later scenario should depend on it as a prerequisite.
UPDATE conversation_scenarios
SET prerequisite_code = 'cafe_mara_dream',
    sort_order = 7,
    updated_at = CURRENT_TIMESTAMP
WHERE npc_code = 'mara_barista' AND code = 'cafe_free_chat';

-- ---------------------------------------------------------------------------
-- 3. Tasks for the new scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- cafe_choose_seat
        ('cafe_choose_seat', 'greet',      0, true,  'Поздороваться с Марой',
            'The learner greets Mara.'),
        ('cafe_choose_seat', 'ask_seat',   1, true,  'Попросить место',
            'The learner asks for a place to sit or says where they want to sit.'),
        ('cafe_choose_seat', 'confirm',    2, true,  'Подтвердить выбор',
            'The learner confirms the seat or table Mara offers.'),
        ('cafe_choose_seat', 'thank',      3, false, 'Поблагодарить',
            'The learner thanks Mara and/or says goodbye.'),

        -- cafe_wrong_order
        ('cafe_wrong_order', 'greet',          0, true,  'Поздороваться с Марой',
            'The learner greets Mara.'),
        ('cafe_wrong_order', 'say_problem',    1, true,  'Сказать, что заказ не тот',
            'The learner says that the order or drink is not correct.'),
        ('cafe_wrong_order', 'state_order',    2, true,  'Повторить правильный заказ',
            'The learner says what they ordered or what they wanted.'),
        ('cafe_wrong_order', 'accept_fix',     3, true,  'Принять исправление',
            'The learner accepts Mara fixing or replacing the order.'),
        ('cafe_wrong_order', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Mara and/or says goodbye.'),

        -- cafe_favorite_drink
        ('cafe_favorite_drink', 'greet',          0, true,  'Поздороваться с Марой',
            'The learner greets Mara.'),
        ('cafe_favorite_drink', 'name_drink',     1, true,  'Назвать любимый напиток',
            'The learner names a drink they like.'),
        ('cafe_favorite_drink', 'say_reason',     2, true,  'Сказать простую причину',
            'The learner gives a simple reason why they like the drink.'),
        ('cafe_favorite_drink', 'ask_mara',       3, true,  'Спросить Мару о её любимом напитке',
            'The learner asks Mara what drink she likes or what her favourite drink is.'),
        ('cafe_favorite_drink', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Mara and/or says goodbye.'),

        -- cafe_busy_morning
        ('cafe_busy_morning', 'greet',        0, true,  'Поздороваться с Марой',
            'The learner greets Mara.'),
        ('cafe_busy_morning', 'order_fast',   1, true,  'Быстро сделать заказ',
            'The learner orders a drink or food item clearly.'),
        ('cafe_busy_morning', 'answer_size',  2, true,  'Ответить про размер или количество',
            'The learner answers a simple follow-up question about size, number, or quantity.'),
        ('cafe_busy_morning', 'offer_help',   3, true,  'Предложить небольшую помощь',
            'The learner offers small help or says they can wait because the cafe is busy.'),
        ('cafe_busy_morning', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Mara and/or says goodbye.'),

        -- cafe_mara_dream
        ('cafe_mara_dream', 'greet',          0, true,  'Поздороваться с Марой',
            'The learner greets Mara.'),
        ('cafe_mara_dream', 'ask_dream',      1, true,  'Спросить о мечте Мары',
            'The learner asks Mara about her dream or future cafe.'),
        ('cafe_mara_dream', 'react',          2, true,  'Отреагировать на мечту',
            'The learner reacts to Mara''s dream with a simple opinion, feeling, or supportive response.'),
        ('cafe_mara_dream', 'share_dream',    3, true,  'Рассказать о своей мечте',
            'The learner shares one simple dream, wish, or future plan of their own.'),
        ('cafe_mara_dream', 'thank',          4, false, 'Поблагодарить и попрощаться',
            'The learner thanks Mara and/or says goodbye.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
