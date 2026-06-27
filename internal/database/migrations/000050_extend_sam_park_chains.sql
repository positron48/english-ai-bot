-- Extend existing Sam (A1 shop) and Officer Park (A2 police) NPC chains.
-- Applies to en_ru and es_ru; NPCs speak the course target language.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        -- Sam
        ('shop_find_gift',        'A1', 'Выбрать подарок в магазине'),
        ('shop_compare_items',    'A1', 'Сравнить два товара'),
        ('shop_lost_receipt',     'A1', 'Объяснить, что чек потерян'),
        ('shop_mystery_object',   'A1', 'Описать странный предмет'),
        ('shop_note_in_bottle',   'A1', 'Обсудить записку в бутылке'),
        ('shop_secret_shelf',     'A1', 'Свободная беседа у секретной полки'),
        -- Officer Park
        ('police_missing_pet',        'A2', 'Сообщить о пропавшем питомце'),
        ('police_timeline',           'A2', 'Восстановить порядок событий'),
        ('police_witness_statement',  'A2', 'Дать свидетельское описание'),
        ('police_noise_complaint',    'A2', 'Пожаловаться на ночной шум'),
        ('police_connect_clues',      'A2', 'Связать две подсказки'),
        ('police_case_closed',        'A2', 'Узнать итог дела')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code IN ('en_ru', 'es_ru')
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. New scenarios linked to their existing NPC chains.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        -- Sam: a friendly shop mystery that stays within A1 language.
        ('shop_find_gift', 'shop', 'A1', 'Выбрать подарок в магазине',
            'Sam', 'sam_shop', 'shop_return_item',
            'a friendly corner-shop assistant who speaks clearly, slowly, and uses simple A1 sentences; he likes helping regular customers choose small gifts',
            'The learner comes back to Sam''s shop and wants a small gift. You are at the counter and suggest simple, everyday items.',
            true, 30, 40000, 3),
        ('shop_compare_items', 'shop', 'A1', 'Сравнить два товара',
            'Sam', 'sam_shop', 'shop_find_gift',
            'a friendly corner-shop assistant who compares everyday items with simple words and short sentences',
            'The learner is choosing between two small items in the shop. You help them compare price, size, colour, or use.',
            true, 30, 40000, 4),
        ('shop_lost_receipt', 'shop', 'A1', 'Объяснить, что чек потерян',
            'Sam', 'sam_shop', 'shop_compare_items',
            'a patient corner-shop assistant who stays polite when there is a small problem; he speaks clearly and slowly',
            'The learner comes back to the shop with a problem but has lost the receipt. You try to understand what they bought and how you can help.',
            true, 30, 40000, 5),
        ('shop_mystery_object', 'shop', 'A1', 'Описать странный предмет',
            'Sam', 'sam_shop', 'shop_lost_receipt',
            'a friendly corner-shop assistant who keeps a box of forgotten things and speaks in simple, curious sentences',
            'Sam opens a small box of things people forgot in the shop. You show the learner one strange object and ask what it looks like.',
            true, 30, 40000, 6),
        ('shop_note_in_bottle', 'shop', 'A1', 'Обсудить записку в бутылке',
            'Sam', 'sam_shop', 'shop_mystery_object',
            'a friendly corner-shop assistant who is curious but not dramatic; he speaks slowly and uses simple A1 sentences',
            'Inside the box, Sam finds a small bottle with a folded note. You ask the learner to help understand what it is and what to do next.',
            true, 30, 40000, 7),
        ('shop_secret_shelf', 'shop', 'A1', 'Свободная беседа у секретной полки',
            'Sam', 'sam_shop', 'shop_note_in_bottle',
            'a friendly corner-shop assistant who shows trusted regulars a small secret shelf of forgotten objects and easy city stories',
            'The learner has helped Sam with the note. The shop is quiet, and Sam shows them a small secret shelf with odd forgotten objects and simple stories.',
            false, 30, 40000, 8),

        -- Officer Park: small safe city cases, still A2.
        ('police_missing_pet', 'police_station', 'A2', 'Сообщить о пропавшем питомце',
            'Officer Park', 'park_police', 'police_ask_next_steps',
            'a calm, patient police officer who asks clear follow-up questions and keeps language suitable for A2 learners',
            'The learner comes to the police station again. A pet is missing, and you help them make a simple report.',
            true, 30, 40000, 3),
        ('police_timeline', 'police_station', 'A2', 'Восстановить порядок событий',
            'Officer Park', 'park_police', 'police_missing_pet',
            'a calm, patient police officer who helps people put events in order using simple past-time language',
            'Following the missing pet report, you ask the learner to explain what happened first, next, and later.',
            true, 30, 40000, 4),
        ('police_witness_statement', 'police_station', 'A2', 'Дать свидетельское описание',
            'Officer Park', 'park_police', 'police_timeline',
            'a calm, patient police officer taking a witness statement with short, clear questions',
            'A neighbour may have seen something. You ask the learner to give a simple witness statement about a person, place, or action.',
            true, 30, 40000, 5),
        ('police_noise_complaint', 'police_station', 'A2', 'Пожаловаться на ночной шум',
            'Officer Park', 'park_police', 'police_witness_statement',
            'a calm police officer who listens carefully to small neighbourhood complaints and explains simple next steps',
            'The learner reports strange but harmless night noise near their building. You ask when it happened and what it sounded like.',
            true, 30, 40000, 6),
        ('police_connect_clues', 'police_station', 'A2', 'Связать две подсказки',
            'Officer Park', 'park_police', 'police_noise_complaint',
            'a calm police officer who explains clues simply and invites the learner to compare details without making the scene scary',
            'Officer Park has two small clues from the pet case and the night noise. You ask the learner to compare them and suggest a simple connection.',
            true, 30, 40000, 7),
        ('police_case_closed', 'police_station', 'A2', 'Узнать итог дела',
            'Officer Park', 'park_police', 'police_connect_clues',
            'a calm, patient police officer who explains how a small city case ended and thanks the learner for helping',
            'The small case is solved. You explain the result in simple words and let the learner ask what happened and what to do next.',
            true, 30, 40000, 8)
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

-- ---------------------------------------------------------------------------
-- 3. Tasks for quest scenarios. shop_secret_shelf is free chat and has no tasks.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- shop_find_gift
        ('shop_find_gift', 'greet',          0, true,  'Поздороваться с Сэмом',
            'The learner greets Sam.'),
        ('shop_find_gift', 'ask_gift',       1, true,  'Попросить помочь с подарком',
            'The learner says they need a gift or asks Sam to help choose one.'),
        ('shop_find_gift', 'say_person',     2, true,  'Сказать, для кого подарок',
            'The learner says who the gift is for.'),
        ('shop_find_gift', 'choose_item',    3, true,  'Выбрать товар',
            'The learner chooses a gift or says which item they want.'),
        ('shop_find_gift', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Sam and/or says goodbye.'),

        -- shop_compare_items
        ('shop_compare_items', 'greet',          0, true,  'Поздороваться с Сэмом',
            'The learner greets Sam.'),
        ('shop_compare_items', 'ask_compare',    1, true,  'Попросить сравнить товары',
            'The learner asks Sam to compare two items or asks about the difference between them.'),
        ('shop_compare_items', 'ask_price',      2, true,  'Спросить цену',
            'The learner asks the price of one or both items.'),
        ('shop_compare_items', 'say_preference', 3, true,  'Сказать, что нравится больше',
            'The learner says which item they prefer or like more.'),
        ('shop_compare_items', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Sam and/or says goodbye.'),

        -- shop_lost_receipt
        ('shop_lost_receipt', 'greet',            0, true,  'Поздороваться с Сэмом',
            'The learner greets Sam.'),
        ('shop_lost_receipt', 'say_no_receipt',   1, true,  'Сказать, что чек потерян',
            'The learner says they lost the receipt or do not have it.'),
        ('shop_lost_receipt', 'describe_purchase',2, true,  'Описать покупку',
            'The learner describes what they bought or when they bought it.'),
        ('shop_lost_receipt', 'ask_help',         3, true,  'Попросить помощи',
            'The learner asks Sam what they can do or asks for help with the problem.'),
        ('shop_lost_receipt', 'thank',            4, false, 'Поблагодарить',
            'The learner thanks Sam and/or says goodbye.'),

        -- shop_mystery_object
        ('shop_mystery_object', 'greet',          0, true,  'Поздороваться с Сэмом',
            'The learner greets Sam.'),
        ('shop_mystery_object', 'ask_box',        1, true,  'Спросить о коробке',
            'The learner asks about the box of forgotten things or what is inside it.'),
        ('shop_mystery_object', 'describe_object',2, true,  'Описать предмет',
            'The learner describes the strange object by colour, size, shape, or possible use.'),
        ('shop_mystery_object', 'ask_owner',      3, true,  'Спросить, чей это предмет',
            'The learner asks who the object belongs to or where it came from.'),
        ('shop_mystery_object', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Sam and/or says goodbye.'),

        -- shop_note_in_bottle
        ('shop_note_in_bottle', 'greet',          0, true,  'Поздороваться с Сэмом',
            'The learner greets Sam.'),
        ('shop_note_in_bottle', 'ask_note',       1, true,  'Спросить о записке',
            'The learner asks about the note or the bottle.'),
        ('shop_note_in_bottle', 'describe_note',  2, true,  'Описать записку',
            'The learner describes what the note looks like or what information is on it.'),
        ('shop_note_in_bottle', 'suggest_action', 3, true,  'Предложить, что делать дальше',
            'The learner suggests a simple next action with the note or bottle.'),
        ('shop_note_in_bottle', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Sam and/or says goodbye.'),

        -- police_missing_pet
        ('police_missing_pet', 'greet',           0, true,  'Поздороваться с полицейским',
            'The learner greets Officer Park.'),
        ('police_missing_pet', 'report_pet',      1, true,  'Сообщить о пропавшем питомце',
            'The learner reports that a pet is missing.'),
        ('police_missing_pet', 'describe_pet',    2, true,  'Описать питомца',
            'The learner describes the pet by animal type, colour, size, name, or another simple detail.'),
        ('police_missing_pet', 'say_last_place',  3, true,  'Сказать, где его видели в последний раз',
            'The learner says where or when the pet was last seen.'),
        ('police_missing_pet', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks Officer Park and/or says goodbye.'),

        -- police_timeline
        ('police_timeline', 'greet',          0, true,  'Поздороваться с полицейским',
            'The learner greets Officer Park.'),
        ('police_timeline', 'say_first',      1, true,  'Сказать, что было сначала',
            'The learner says what happened first.'),
        ('police_timeline', 'say_next',       2, true,  'Сказать, что было потом',
            'The learner says what happened next or later.'),
        ('police_timeline', 'answer_time',    3, true,  'Ответить про время',
            'The learner answers a follow-up question about time, day, or order of events.'),
        ('police_timeline', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Officer Park and/or says goodbye.'),

        -- police_witness_statement
        ('police_witness_statement', 'greet',        0, true,  'Поздороваться с полицейским',
            'The learner greets Officer Park.'),
        ('police_witness_statement', 'say_witnessed',1, true,  'Сказать, что видел',
            'The learner says what they saw or heard.'),
        ('police_witness_statement', 'describe_place',2, true, 'Описать место',
            'The learner describes where it happened or what the place looked like.'),
        ('police_witness_statement', 'describe_action',3, true,'Описать действие',
            'The learner describes what a person or animal was doing.'),
        ('police_witness_statement', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Officer Park and/or says goodbye.'),

        -- police_noise_complaint
        ('police_noise_complaint', 'greet',        0, true,  'Поздороваться с полицейским',
            'The learner greets Officer Park.'),
        ('police_noise_complaint', 'report_noise', 1, true,  'Сообщить о шуме',
            'The learner reports a noise problem or says there was strange noise.'),
        ('police_noise_complaint', 'say_time',     2, true,  'Сказать, когда это было',
            'The learner says when the noise happened.'),
        ('police_noise_complaint', 'describe_sound',3, true, 'Описать звук',
            'The learner describes the sound or what it was like.'),
        ('police_noise_complaint', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Officer Park and/or says goodbye.'),

        -- police_connect_clues
        ('police_connect_clues', 'greet',         0, true,  'Поздороваться с полицейским',
            'The learner greets Officer Park.'),
        ('police_connect_clues', 'name_clue_one', 1, true,  'Назвать первую подсказку',
            'The learner names or repeats one clue.'),
        ('police_connect_clues', 'name_clue_two', 2, true,  'Назвать вторую подсказку',
            'The learner names or repeats a second clue.'),
        ('police_connect_clues', 'suggest_link',  3, true,  'Предложить связь',
            'The learner suggests a simple connection between the clues.'),
        ('police_connect_clues', 'thank',         4, false, 'Поблагодарить',
            'The learner thanks Officer Park and/or says goodbye.'),

        -- police_case_closed
        ('police_case_closed', 'greet',        0, true,  'Поздороваться с полицейским',
            'The learner greets Officer Park.'),
        ('police_case_closed', 'ask_result',   1, true,  'Спросить, чем всё закончилось',
            'The learner asks what happened, what the result was, or whether the case is solved.'),
        ('police_case_closed', 'ask_pet',      2, true,  'Спросить о питомце',
            'The learner asks about the missing pet or whether it is safe.'),
        ('police_case_closed', 'react_result', 3, true,  'Отреагировать на итог',
            'The learner reacts to the result with a simple opinion, feeling, or short comment.'),
        ('police_case_closed', 'thank',        4, false, 'Поблагодарить и попрощаться',
            'The learner thanks Officer Park and/or says goodbye.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
