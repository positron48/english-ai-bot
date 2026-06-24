-- Re-seed conversation scenarios for en_ru and es_ru.
-- The original seed (000040) ran once; if a course (e.g. es_ru) or its districts/
-- locations were created afterwards, that course got no scenarios. This migration
-- re-applies the same seed idempotently (ON CONFLICT DO NOTHING) so any missing
-- per-course copies are filled in without duplicating existing rows.

-- Backing learning_items (one per scenario per course).
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('cafe_order_coffee',  'A0', 'Заказать кофе в кафе'),
        ('shop_buy_water',     'A1', 'Купить воду в магазине'),
        ('police_report_lost', 'A2', 'Заявить о пропаже в полиции'),
        ('cafe_free_chat',     'A0', 'Свободная беседа в кафе')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code IN ('en_ru', 'es_ru')
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- Scenarios (linked to backing learning_items).
WITH scen(code, place_type, cefr_level, title, npc_name, npc_persona, scene_setup, is_quest, max_turns, sort_order) AS (
    VALUES
        ('cafe_order_coffee', 'cafe', 'A0', 'Заказать кофе в кафе',
            'Mara', 'a warm, patient cafe barista who keeps sentences short and simple',
            'The learner walks into a cozy neighborhood cafe. You are behind the counter, ready to take their order.',
            true, 16, 0),
        ('shop_buy_water', 'shop', 'A1', 'Купить воду в магазине',
            'Sam', 'a friendly corner-shop assistant who speaks clearly and slowly',
            'The learner enters a small corner shop. You are the assistant standing at the counter.',
            true, 18, 1),
        ('police_report_lost', 'police_station', 'A2', 'Заявить о пропаже в полиции',
            'Officer Park', 'a calm, patient police officer at the station front desk',
            'The learner comes to the police station front desk to report a lost item.',
            true, 22, 2),
        ('cafe_free_chat', 'cafe', 'A0', 'Свободная беседа в кафе',
            'Mara', 'a chatty cafe barista who loves easy small talk',
            'A quiet afternoon in the cafe. You make friendly small talk with the visitor about their day.',
            false, 30, 3)
)
INSERT INTO conversation_scenarios
    (course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
     title, npc_name, npc_persona, scene_setup, is_quest, max_turns, sort_order, status)
SELECT c.id, d.id, l.id, li.id, s.code, s.place_type, s.cefr_level,
       s.title, s.npc_name, s.npc_persona, s.scene_setup, s.is_quest, s.max_turns, s.sort_order, 'active'
FROM scen s
JOIN courses c ON c.code IN ('en_ru', 'es_ru')
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
JOIN learning_items li
    ON li.course_id = c.id AND li.source_kind = 'conversation_scenario' AND li.source_id = s.code
ON CONFLICT (course_id, code) DO NOTHING;

-- Tasks (matched to every course's copy of the scenario by code).
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        ('cafe_order_coffee', 'greet',     0, true,  'Поздороваться с бариста',
            'The user greets the barista in the target language (any common greeting).'),
        ('cafe_order_coffee', 'order',     1, true,  'Заказать кофе с молоком',
            'The user orders a coffee with milk (e.g. cafe con leche / coffee with milk).'),
        ('cafe_order_coffee', 'sugar',     2, true,  'Попросить 2 кусочка сахара',
            'The user asks for two sugars (e.g. dos azucares / two sugars) for the coffee.'),
        ('cafe_order_coffee', 'thank',     3, false, 'Поблагодарить и попрощаться',
            'The user thanks the barista and/or says goodbye.'),
        ('shop_buy_water', 'greet',        0, true,  'Поздороваться с продавцом',
            'The user greets the shop assistant in the target language.'),
        ('shop_buy_water', 'ask_water',    1, true,  'Попросить бутылку воды',
            'The user asks for a bottle of water.'),
        ('shop_buy_water', 'ask_price',    2, true,  'Спросить цену',
            'The user asks how much it costs / the price.'),
        ('shop_buy_water', 'pay_thank',    3, false, 'Заплатить и поблагодарить',
            'The user indicates they pay and thanks the assistant.'),
        ('police_report_lost', 'greet',        0, true,  'Поздороваться с полицейским',
            'The user greets the police officer.'),
        ('police_report_lost', 'report_lost',  1, true,  'Сообщить о потерянной вещи',
            'The user reports that they have lost something.'),
        ('police_report_lost', 'describe',     2, true,  'Описать потерянную вещь',
            'The user describes the lost item (what it is, colour, size, or where it was lost).'),
        ('police_report_lost', 'give_contact', 3, false, 'Оставить контакт для связи',
            'The user provides contact information (a name or phone number) for follow-up.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
