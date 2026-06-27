-- Seed English-only Nora Finch pharmacy NPC chain.
-- Nora Finch speaks English; titles are RU, persona/setup/criteria are model-facing EN.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('pharmacy_buy_medicine',        'A1', 'Попросить простое лекарство'),
        ('pharmacy_ask_dosage',          'A1', 'Спросить, как принимать'),
        ('pharmacy_describe_symptoms',   'A1', 'Описать симптомы'),
        ('pharmacy_allergy_warning',     'A1', 'Сказать об аллергии или ограничении'),
        ('pharmacy_help_neighbor',       'A1', 'Купить лекарство для другого человека'),
        ('pharmacy_strange_prescription','A1', 'Обсудить странный рецепт')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. New Nora Finch scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('pharmacy_buy_medicine', 'pharmacy', 'A1', 'Попросить простое лекарство',
            'Nora Finch', 'nora_pharmacist', '',
            'a friendly neighbourhood pharmacist who knows many local residents and speaks clear, simple A1 English; she keeps the scene safe by avoiding diagnosis, prescriptions, or medical advice beyond everyday store help and label reminders',
            'The learner enters a small neighbourhood pharmacy and asks for a simple over-the-counter item. You are behind the counter, ask gentle basic questions, and keep the exchange about safe everyday pharmacy service.',
            true, 30, 40000, 0),
        ('pharmacy_ask_dosage', 'pharmacy', 'A1', 'Спросить, как принимать',
            'Nora Finch', 'nora_pharmacist', 'pharmacy_buy_medicine',
            'a friendly neighbourhood pharmacist who speaks slowly in simple A1 English; she points to labels and general instructions but never invents a dosage or gives risky medical advice',
            'The learner has chosen a safe everyday pharmacy item. You help them ask how to take it in simple words, refer to the label, and keep the conversation practical and safe.',
            true, 30, 40000, 1),
        ('pharmacy_describe_symptoms', 'pharmacy', 'A1', 'Описать симптомы',
            'Nora Finch', 'nora_pharmacist', 'pharmacy_ask_dosage',
            'a friendly neighbourhood pharmacist who listens carefully and uses simple, supportive English; she does not diagnose, prescribe, or give risky advice, and suggests professional help for anything serious',
            'The learner returns to the pharmacy and wants to describe how they feel in basic words. You ask simple questions about everyday symptoms and keep the goal on communication practice, not medical decision-making.',
            true, 30, 40000, 2),
        ('pharmacy_allergy_warning', 'pharmacy', 'A1', 'Сказать об аллергии или ограничении',
            'Nora Finch', 'nora_pharmacist', 'pharmacy_describe_symptoms',
            'a friendly neighbourhood pharmacist who is careful and kind; she uses simple English to ask about allergies or restrictions and reminds the learner to read labels and ask a professional when unsure',
            'The learner needs to mention an allergy or dietary restriction before buying a pharmacy item. You ask simple follow-up questions and give a general safety reminder without diagnosing or prescribing.',
            true, 30, 40000, 3),
        ('pharmacy_help_neighbor', 'pharmacy', 'A1', 'Купить лекарство для другого человека',
            'Nora Finch', 'nora_pharmacist', 'pharmacy_allergy_warning',
            'a friendly neighbourhood pharmacist who knows the area well and speaks in clear, simple English; she keeps requests safe and reminds the learner that another person should read labels and ask a professional when unsure',
            'The learner comes in to buy a simple pharmacy item for a neighbour or family member. You ask who it is for, what they need, and whether they understand the basic purchase, without giving personal medical instructions.',
            true, 30, 40000, 4),
        ('pharmacy_strange_prescription', 'pharmacy', 'A1', 'Обсудить странный рецепт',
            'Nora Finch', 'nora_pharmacist', 'pharmacy_help_neighbor',
            'a friendly neighbourhood pharmacist with many harmless local stories; she speaks simple English, notices odd but non-mysterious details about prescriptions or requests, and never gives dangerous advice or turns the scene into fantasy',
            'The pharmacy is quiet. A customer has just left with an unusual but harmless prescription note, and Nora mentions it in passing. You let the learner ask polite questions while keeping the scene safe, practical, and non-medical.',
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
-- 3. Tasks for the new scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- pharmacy_buy_medicine
        ('pharmacy_buy_medicine', 'greet',        0, true,  'Поздороваться с Норой',
            'The learner greets Nora Finch politely.'),
        ('pharmacy_buy_medicine', 'ask_medicine', 1, true,  'Попросить простое лекарство',
            'The learner asks for a simple over-the-counter pharmacy item or everyday help.'),
        ('pharmacy_buy_medicine', 'answer_basic', 2, true,  'Ответить на простой вопрос',
            'The learner answers a basic follow-up question about the situation without needing medical detail.'),
        ('pharmacy_buy_medicine', 'confirm_item', 3, true,  'Подтвердить выбор',
            'The learner confirms the item or type of help they want.'),
        ('pharmacy_buy_medicine', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Nora Finch and/or says goodbye.'),

        -- pharmacy_ask_dosage
        ('pharmacy_ask_dosage', 'greet',             0, true,  'Поздороваться с Норой',
            'The learner greets Nora Finch politely.'),
        ('pharmacy_ask_dosage', 'ask_how_to_take',   1, true,  'Спросить, как принимать',
            'The learner asks how to take the pharmacy item in simple, general terms.'),
        ('pharmacy_ask_dosage', 'ask_when',          2, true,  'Спросить, когда принимать',
            'The learner asks when to take the item or asks about timing in a general way.'),
        ('pharmacy_ask_dosage', 'confirm_understand',3, true,  'Показать, что понял',
            'The learner shows they understand the general instructions, such as reading the label or following professional advice.'),
        ('pharmacy_ask_dosage', 'thank',             4, false, 'Поблагодарить',
            'The learner thanks Nora Finch and/or says goodbye.'),

        -- pharmacy_describe_symptoms
        ('pharmacy_describe_symptoms', 'greet',            0, true,  'Поздороваться с Норой',
            'The learner greets Nora Finch politely.'),
        ('pharmacy_describe_symptoms', 'say_feel_bad',     1, true,  'Сказать, что плохо себя чувствует',
            'The learner says they do not feel well or have a minor health concern.'),
        ('pharmacy_describe_symptoms', 'describe_symptom', 2, true,  'Описать симптом',
            'The learner describes one or more symptoms in simple, everyday terms.'),
        ('pharmacy_describe_symptoms', 'answer_when',      3, true,  'Ответить, когда началось',
            'The learner answers a simple follow-up question about when the problem started or how long it has lasted.'),
        ('pharmacy_describe_symptoms', 'confirm_next',     4, false, 'Подтвердить следующий шаг',
            'The learner confirms they understand the safe next step, such as reading the label or asking a professional if needed.'),

        -- pharmacy_allergy_warning
        ('pharmacy_allergy_warning', 'greet',            0, true,  'Поздороваться с Норой',
            'The learner greets Nora Finch politely.'),
        ('pharmacy_allergy_warning', 'mention_allergy',  1, true,  'Сказать об аллергии или ограничении',
            'The learner mentions an allergy, food restriction, or other limitation relevant to a pharmacy purchase.'),
        ('pharmacy_allergy_warning', 'answer_followup',  2, true,  'Ответить на уточняющий вопрос',
            'The learner answers a basic follow-up question about the allergy or restriction.'),
        ('pharmacy_allergy_warning', 'confirm_safe',     3, true,  'Подтвердить, что понимает предупреждение',
            'The learner shows they understand Nora''s general safety warning or what they should check on the label.'),
        ('pharmacy_allergy_warning', 'thank',            4, false, 'Поблагодарить',
            'The learner thanks Nora Finch and/or says goodbye.'),

        -- pharmacy_help_neighbor
        ('pharmacy_help_neighbor', 'greet',         0, true,  'Поздороваться с Норой',
            'The learner greets Nora Finch politely.'),
        ('pharmacy_help_neighbor', 'say_for_other', 1, true,  'Сказать, что покупка для другого человека',
            'The learner says the pharmacy item is for another person, such as a neighbour or family member.'),
        ('pharmacy_help_neighbor', 'describe_need', 2, true,  'Объяснить, что нужно',
            'The learner describes what kind of simple pharmacy item or help the other person needs.'),
        ('pharmacy_help_neighbor', 'answer_detail', 3, true,  'Ответить на уточнение',
            'The learner answers a basic follow-up question about the other person or the request.'),
        ('pharmacy_help_neighbor', 'thank',         4, false, 'Поблагодарить',
            'The learner thanks Nora Finch and/or says goodbye.'),

        -- pharmacy_strange_prescription
        ('pharmacy_strange_prescription', 'greet',              0, true,  'Поздороваться с Норой',
            'The learner greets Nora Finch politely.'),
        ('pharmacy_strange_prescription', 'mention_prescription',1, true, 'Рассказать о странном рецепте',
            'The learner asks about or comments on the unusual prescription or request Nora mentioned.'),
        ('pharmacy_strange_prescription', 'ask_question',       2, true,  'Задать уточняющий вопрос',
            'The learner asks one polite follow-up question about the prescription note or the customer.'),
        ('pharmacy_strange_prescription', 'react_safely',       3, true,  'Отреагировать безопасно',
            'The learner reacts to Nora''s story with a simple comment or opinion without asking for medical advice or turning it into mystery.'),
        ('pharmacy_strange_prescription', 'thank',              4, false, 'Поблагодарить и попрощаться',
            'The learner thanks Nora Finch and/or says goodbye.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'en_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
