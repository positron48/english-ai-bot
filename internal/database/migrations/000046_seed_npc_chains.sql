-- Seed NPC chains: group the existing conversation scenarios under recurring NPCs and add
-- follow-up scenarios that unlock one after another. Applies to en_ru and es_ru (NPC speaks the
-- course target language; titles are RU, persona/scene/criteria are model-facing EN instructions).
--
-- Chains (within one CEFR level / district each):
--   Mara (cafe, A0):      cafe_order_coffee -> cafe_order_pastry -> cafe_free_chat
--   Sam (shop, A1):       shop_buy_water    -> shop_ask_directions -> shop_return_item
--   Officer Park (A2):    police_report_lost -> police_describe_person -> police_ask_next_steps

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the NEW scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('cafe_order_pastry',      'A0', 'Заказать выпечку в кафе'),
        ('shop_ask_directions',    'A1', 'Спросить, где найти товар'),
        ('shop_return_item',       'A1', 'Вернуть товар в магазин'),
        ('police_describe_person', 'A2', 'Описать человека в полиции'),
        ('police_ask_next_steps',  'A2', 'Узнать, что будет дальше')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code IN ('en_ru', 'es_ru')
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. New scenarios (linked to backing learning_items).
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, sort_order) AS (
    VALUES
        ('cafe_order_pastry', 'cafe', 'A0', 'Заказать выпечку в кафе',
            'Mara', 'mara_barista', 'cafe_order_coffee',
            'a warm, patient cafe barista who keeps sentences short and simple',
            'The learner is back at the cafe counter. You are the barista, ready to help them pick something to eat.',
            true, 16, 1),
        ('shop_ask_directions', 'shop', 'A1', 'Спросить, где найти товар',
            'Sam', 'sam_shop', 'shop_buy_water',
            'a friendly corner-shop assistant who speaks clearly and slowly',
            'The learner is in the corner shop again and cannot find what they need. You are the assistant who points them to the right aisle.',
            true, 16, 1),
        ('shop_return_item', 'shop', 'A1', 'Вернуть товар в магазин',
            'Sam', 'sam_shop', 'shop_ask_directions',
            'a friendly corner-shop assistant who stays polite even with complaints',
            'The learner comes back to the shop to return or exchange something they bought. You are the assistant at the counter.',
            true, 18, 2),
        ('police_describe_person', 'police_station', 'A2', 'Описать человека в полиции',
            'Officer Park', 'park_police', 'police_report_lost',
            'a calm, patient police officer taking notes at the front desk',
            'Following up on the report, you ask the learner to describe the person they saw. You take notes and ask short follow-up questions.',
            true, 22, 1),
        ('police_ask_next_steps', 'police_station', 'A2', 'Узнать, что будет дальше',
            'Officer Park', 'park_police', 'police_describe_person',
            'a calm, patient police officer who explains procedures clearly',
            'The report is filed. The learner wants to know what happens next. You explain the procedure step by step.',
            true, 20, 2)
)
INSERT INTO conversation_scenarios
    (course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
     title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, sort_order, status)
SELECT c.id, d.id, l.id, li.id, s.code, s.place_type, s.cefr_level,
       s.title, s.npc_name, s.npc_code, s.prerequisite_code, s.npc_persona, s.scene_setup, s.is_quest, s.max_turns, s.sort_order, 'active'
FROM scen s
JOIN courses c ON c.code IN ('en_ru', 'es_ru')
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
JOIN learning_items li
    ON li.course_id = c.id AND li.source_kind = 'conversation_scenario' AND li.source_id = s.code
ON CONFLICT (course_id, code) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3. Wire the EXISTING scenarios into their chains (npc_code, prerequisite, order).
-- ---------------------------------------------------------------------------
WITH chain(code, npc_code, prerequisite_code, sort_order) AS (
    VALUES
        ('cafe_order_coffee',  'mara_barista', '',                  0),
        ('cafe_free_chat',     'mara_barista', 'cafe_order_pastry', 2),
        ('shop_buy_water',     'sam_shop',     '',                  0),
        ('police_report_lost', 'park_police',  '',                  0)
)
UPDATE conversation_scenarios cs
SET npc_code = chain.npc_code,
    prerequisite_code = chain.prerequisite_code,
    sort_order = chain.sort_order,
    updated_at = CURRENT_TIMESTAMP
FROM chain
WHERE cs.code = chain.code;

-- ---------------------------------------------------------------------------
-- 4. Tasks for the NEW scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- cafe_order_pastry
        ('cafe_order_pastry', 'greet',       0, true,  'Поздороваться с бариста',
            'The user greets the barista in the target language.'),
        ('cafe_order_pastry', 'ask_options', 1, true,  'Спросить, что есть из выпечки',
            'The user asks what pastries / food are available, or asks for a specific item.'),
        ('cafe_order_pastry', 'order',       2, true,  'Заказать что-то конкретное',
            'The user orders a specific pastry or food item (e.g. a croissant).'),
        ('cafe_order_pastry', 'ask_price',   3, true,  'Спросить цену',
            'The user asks how much it costs / the price.'),
        ('cafe_order_pastry', 'thank',       4, false, 'Поблагодарить и попрощаться',
            'The user thanks the barista and/or says goodbye.'),
        -- shop_ask_directions
        ('shop_ask_directions', 'greet',        0, true,  'Поздороваться с продавцом',
            'The user greets the shop assistant in the target language.'),
        ('shop_ask_directions', 'ask_location', 1, true,  'Спросить, где находится товар',
            'The user asks where a product is (e.g. where is the bread / milk / water).'),
        ('shop_ask_directions', 'confirm',      2, true,  'Подтвердить, что понял направление',
            'The user confirms or repeats back the aisle / direction the assistant gave.'),
        ('shop_ask_directions', 'thank',        3, false, 'Поблагодарить',
            'The user thanks the assistant.'),
        -- shop_return_item
        ('shop_return_item', 'greet',           0, true,  'Поздороваться с продавцом',
            'The user greets the shop assistant.'),
        ('shop_return_item', 'explain_problem', 1, true,  'Объяснить, что хочет вернуть товар',
            'The user explains that there is a problem with an item or that they want to return it.'),
        ('shop_return_item', 'give_reason',     2, true,  'Назвать причину',
            'The user gives a reason (broken, wrong size, expired, not needed, etc.).'),
        ('shop_return_item', 'ask_resolution',  3, true,  'Попросить возврат или обмен',
            'The user asks for a refund or an exchange.'),
        ('shop_return_item', 'thank',           4, false, 'Поблагодарить и попрощаться',
            'The user thanks the assistant and/or says goodbye.'),
        -- police_describe_person
        ('police_describe_person', 'greet',              0, true,  'Поздороваться с полицейским',
            'The user greets the police officer.'),
        ('police_describe_person', 'describe_appearance', 1, true, 'Описать внешность человека',
            'The user describes the person''s appearance (height, hair, clothes, or age).'),
        ('police_describe_person', 'describe_action',     2, true, 'Рассказать, что делал человек',
            'The user describes what the person was doing or where they went.'),
        ('police_describe_person', 'answer_followup',     3, true, 'Ответить на уточняющий вопрос',
            'The user answers at least one follow-up question asked by the officer.'),
        ('police_describe_person', 'thank',               4, false, 'Поблагодарить',
            'The user thanks the officer and/or says goodbye.'),
        -- police_ask_next_steps
        ('police_ask_next_steps', 'greet',          0, true,  'Поздороваться с полицейским',
            'The user greets the police officer.'),
        ('police_ask_next_steps', 'ask_next',       1, true,  'Спросить, что будет дальше',
            'The user asks what will happen next or what they should do.'),
        ('police_ask_next_steps', 'ask_contact',    2, true,  'Спросить, как с ними свяжутся',
            'The user asks how or when the police will contact them.'),
        ('police_ask_next_steps', 'confirm',        3, true,  'Подтвердить, что всё понял',
            'The user confirms they understood the procedure.'),
        ('police_ask_next_steps', 'thank',          4, false, 'Поблагодарить и попрощаться',
            'The user thanks the officer and/or says goodbye.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
