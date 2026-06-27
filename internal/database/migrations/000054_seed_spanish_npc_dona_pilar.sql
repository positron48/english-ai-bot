-- Seed Spanish-only Doña Pilar pharmacy NPC chain.
-- Doña Pilar speaks Spanish; titles are RU, persona/setup/criteria are model-facing EN.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('pharmacy_pilar_cough',      'A1', 'Попросить что-то от кашля'),
        ('pharmacy_pilar_price',      'A1', 'Спросить цену и способ оплаты'),
        ('pharmacy_pilar_symptoms',   'A1', 'Описать симптомы'),
        ('pharmacy_pilar_for_friend', 'A1', 'Купить лекарство для друга'),
        ('pharmacy_pilar_warning',    'A1', 'Понять предупреждение о дозировке'),
        ('pharmacy_pilar_gossip',     'A1', 'Расспросить о странном посетителе')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. New Doña Pilar scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('pharmacy_pilar_cough', 'pharmacy', 'A1', 'Попросить что-то от кашля',
            'Doña Pilar', 'pilar_pharmacist', '',
            'an older Spanish-speaking pharmacist with a sharp tongue and a kind heart; she uses clear, simple A1 Spanish and keeps the scene safe by avoiding diagnosis, prescriptions, or medical advice beyond everyday store help and label reminders',
            'The learner enters a small neighbourhood pharmacy and asks for simple help with a cough. You are behind the counter, ask gentle basic questions, and keep the exchange about safe over-the-counter pharmacy service.',
            true, 30, 40000, 0),
        ('pharmacy_pilar_price', 'pharmacy', 'A1', 'Спросить цену и способ оплаты',
            'Doña Pilar', 'pilar_pharmacist', 'pharmacy_pilar_cough',
            'an older Spanish-speaking pharmacist who is practical, witty, and kind; she speaks slowly in simple A1 Spanish and keeps the conversation focused on price, payment, and polite service',
            'The learner has chosen a safe everyday pharmacy item. You tell them the price, answer payment questions, and keep the scene calm and practical.',
            true, 30, 40000, 1),
        ('pharmacy_pilar_symptoms', 'pharmacy', 'A1', 'Описать симптомы',
            'Doña Pilar', 'pilar_pharmacist', 'pharmacy_pilar_price',
            'an older Spanish-speaking pharmacist who listens carefully and uses simple, supportive Spanish; she does not diagnose, prescribe, or give risky advice, and suggests professional help for anything serious',
            'The learner returns to the pharmacy and wants to describe how they feel in basic words. You ask simple questions about everyday symptoms and keep the goal on communication practice, not medical decision-making.',
            true, 30, 40000, 2),
        ('pharmacy_pilar_for_friend', 'pharmacy', 'A1', 'Купить лекарство для друга',
            'Doña Pilar', 'pilar_pharmacist', 'pharmacy_pilar_symptoms',
            'an older Spanish-speaking pharmacist who knows the neighbourhood and speaks in clear, simple Spanish; she keeps requests safe and reminds the learner that another person should read labels and ask a professional when unsure',
            'The learner comes in to buy a simple pharmacy item for a friend. You ask who it is for, what they need, and whether they understand the basic purchase, without giving personal medical instructions.',
            true, 30, 40000, 3),
        ('pharmacy_pilar_warning', 'pharmacy', 'A1', 'Понять предупреждение о дозировке',
            'Doña Pilar', 'pilar_pharmacist', 'pharmacy_pilar_for_friend',
            'an older Spanish-speaking pharmacist who is direct but caring; she uses simple Spanish to explain that labels and professional instructions must be followed and avoids inventing any dosage or medical advice',
            'The learner is about to leave with a pharmacy item. You point to the label, give a general safety warning in simple words, and check that the learner understands the warning without stating a specific dosage.',
            true, 30, 40000, 4),
        ('pharmacy_pilar_gossip', 'pharmacy', 'A1', 'Расспросить о странном посетителе',
            'Doña Pilar', 'pilar_pharmacist', 'pharmacy_pilar_warning',
            'an older Spanish-speaking pharmacist with a sharp tongue, a kind heart, and many harmless neighbourhood stories; she speaks simple Spanish and shares gossip only when the learner is polite',
            'The pharmacy is quiet. A strange but harmless visitor has just left, and Doña Pilar hints that she knows something about the neighbourhood. You let the learner ask polite questions while keeping the scene safe and non-medical.',
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
-- 3. Tasks for the new scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- pharmacy_pilar_cough
        ('pharmacy_pilar_cough', 'greet',       0, true,  'Поздороваться с Доньей Пилар',
            'The learner greets Doña Pilar politely.'),
        ('pharmacy_pilar_cough', 'ask_help',    1, true,  'Попросить помощь от кашля',
            'The learner asks for a simple pharmacy item or help related to a cough.'),
        ('pharmacy_pilar_cough', 'answer_basic',2, true,  'Ответить на простой вопрос',
            'The learner answers a basic follow-up question about the situation without needing medical detail.'),
        ('pharmacy_pilar_cough', 'choose_item', 3, true,  'Подтвердить выбор',
            'The learner confirms the item or type of help they want.'),
        ('pharmacy_pilar_cough', 'thank',       4, false, 'Поблагодарить',
            'The learner thanks Doña Pilar and/or says goodbye.'),

        -- pharmacy_pilar_price
        ('pharmacy_pilar_price', 'greet',        0, true,  'Поздороваться с Доньей Пилар',
            'The learner greets Doña Pilar politely.'),
        ('pharmacy_pilar_price', 'ask_price',    1, true,  'Спросить цену',
            'The learner asks how much the item costs.'),
        ('pharmacy_pilar_price', 'ask_payment',  2, true,  'Спросить про оплату',
            'The learner asks whether they can pay by a chosen method or asks about payment options.'),
        ('pharmacy_pilar_price', 'confirm_buy',  3, true,  'Подтвердить покупку',
            'The learner confirms that they want to buy the item or accepts the price.'),
        ('pharmacy_pilar_price', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Doña Pilar and/or says goodbye.'),

        -- pharmacy_pilar_symptoms
        ('pharmacy_pilar_symptoms', 'greet',          0, true,  'Поздороваться с Доньей Пилар',
            'The learner greets Doña Pilar politely.'),
        ('pharmacy_pilar_symptoms', 'say_problem',    1, true,  'Сказать, что плохо себя чувствует',
            'The learner says they do not feel well or have a minor health concern.'),
        ('pharmacy_pilar_symptoms', 'describe_symptom',2, true, 'Описать симптом',
            'The learner describes one or more symptoms in simple, everyday terms.'),
        ('pharmacy_pilar_symptoms', 'answer_when',    3, true,  'Ответить, когда началось',
            'The learner answers a simple follow-up question about when the problem started or how long it has lasted.'),
        ('pharmacy_pilar_symptoms', 'confirm_next',   4, false, 'Подтвердить следующий шаг',
            'The learner confirms they understand the safe next step, such as reading the label or asking a professional if needed.'),

        -- pharmacy_pilar_for_friend
        ('pharmacy_pilar_for_friend', 'greet',        0, true,  'Поздороваться с Доньей Пилар',
            'The learner greets Doña Pilar politely.'),
        ('pharmacy_pilar_for_friend', 'say_for_friend',1, true, 'Сказать, что покупка для друга',
            'The learner says the pharmacy item is for another person.'),
        ('pharmacy_pilar_for_friend', 'describe_need',2, true,  'Объяснить, что нужно',
            'The learner describes what kind of simple pharmacy item or help the friend needs.'),
        ('pharmacy_pilar_for_friend', 'answer_detail',3, true,  'Ответить на уточнение',
            'The learner answers a basic follow-up question about the friend or the request.'),
        ('pharmacy_pilar_for_friend', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Doña Pilar and/or says goodbye.'),

        -- pharmacy_pilar_warning
        ('pharmacy_pilar_warning', 'greet',          0, true,  'Поздороваться с Доньей Пилар',
            'The learner greets Doña Pilar politely.'),
        ('pharmacy_pilar_warning', 'listen_warning', 1, true,  'Выслушать предупреждение',
            'The learner acknowledges that Doña Pilar is giving an important safety warning.'),
        ('pharmacy_pilar_warning', 'ask_clarify',    2, true,  'Попросить пояснить',
            'The learner asks for clarification about the warning or label in a general way.'),
        ('pharmacy_pilar_warning', 'repeat_meaning', 3, true,  'Показать, что понял',
            'The learner shows they understand the warning by summarizing its general meaning or saying what they will check.'),
        ('pharmacy_pilar_warning', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Doña Pilar and/or says goodbye.'),

        -- pharmacy_pilar_gossip
        ('pharmacy_pilar_gossip', 'greet',          0, true,  'Поздороваться с Доньей Пилар',
            'The learner greets Doña Pilar politely.'),
        ('pharmacy_pilar_gossip', 'ask_visitor',    1, true,  'Спросить о посетителе',
            'The learner asks who the visitor was or why the visitor seemed unusual.'),
        ('pharmacy_pilar_gossip', 'ask_detail',     2, true,  'Уточнить деталь',
            'The learner asks one polite follow-up question about the visitor or neighbourhood story.'),
        ('pharmacy_pilar_gossip', 'react_story',    3, true,  'Отреагировать на историю',
            'The learner reacts to the story with a simple opinion, feeling, or short comment.'),
        ('pharmacy_pilar_gossip', 'thank',          4, false, 'Поблагодарить и попрощаться',
            'The learner thanks Doña Pilar and/or says goodbye.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'es_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
