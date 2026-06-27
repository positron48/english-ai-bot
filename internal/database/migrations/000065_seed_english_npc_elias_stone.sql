-- Seed English-only Dr. Elias Stone clinic NPC chain.
-- Applies only to en_ru; Dr. Stone speaks English. Titles are RU, while persona,
-- scene setup, and completion criteria are model-facing EN instructions.
-- All scenarios are pinned to A2 (first from A2-B1).

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('clinic_make_appointment',  'A2', 'Записаться на приём'),
        ('clinic_describe_pain',       'A2', 'Описать боль или самочувствие'),
        ('clinic_answer_questions',    'A2', 'Ответить про сон, еду и температуру'),
        ('clinic_explain_history',     'A2', 'Рассказать, когда всё началось'),
        ('clinic_follow_advice',       'A2', 'Обсудить рекомендации врача'),
        ('clinic_city_pattern',        'A2', 'Заметить связь между жалобами жителей')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Dr. Elias Stone clinic scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('clinic_make_appointment', 'clinic', 'A2', 'Записаться на приём',
            'Dr. Elias Stone', 'elias_doctor', '',
            'a practical, calm doctor in a small neighbourhood clinic who speaks English clearly for A2 learners; he keeps the exchange as safe language practice and never gives real medical diagnoses, prescriptions, dosage advice, or urgent-care instructions',
            'The learner comes to the clinic reception area and wants a simple appointment. You are Dr. Elias Stone, available for a safe everyday conversation about scheduling, not for real medical care.',
            true, 30, 40000, 0),
        ('clinic_describe_pain', 'clinic', 'A2', 'Описать боль или самочувствие',
            'Dr. Elias Stone', 'elias_doctor', 'clinic_make_appointment',
            'a practical, calm doctor who asks simple symptom questions in English and redirects anything serious to real professional help without giving treatment instructions',
            'The learner is in a basic clinic conversation and needs to describe discomfort or how they feel. Keep the scene focused on language practice: body area, intensity, and everyday context, with no real diagnosis or treatment plan.',
            true, 30, 40000, 1),
        ('clinic_answer_questions', 'clinic', 'A2', 'Ответить про сон, еду и температуру',
            'Dr. Elias Stone', 'elias_doctor', 'clinic_describe_pain',
            'a practical, calm doctor who uses simple English questions about daily habits; he gives only general, non-medical, low-risk lifestyle reflections and avoids personalized medical advice',
            'After hearing the learner describe how they feel, you ask safe everyday questions about sleep, food, and temperature. This is language practice in a friendly clinic, not a real consultation.',
            true, 30, 40000, 2),
        ('clinic_explain_history', 'clinic', 'A2', 'Рассказать, когда всё началось',
            'Dr. Elias Stone', 'elias_doctor', 'clinic_answer_questions',
            'a practical, calm doctor who helps the learner tell a simple health-related timeline in English while avoiding diagnosis, medication choices, or risky advice',
            'The learner returns to explain when the problem started and what changed over time. Ask for a safe everyday timeline and keep the focus on narration and sequencing.',
            true, 30, 40000, 3),
        ('clinic_follow_advice', 'clinic', 'A2', 'Обсудить рекомендации врача',
            'Dr. Elias Stone', 'elias_doctor', 'clinic_explain_history',
            'a practical, calm doctor who explains harmless general suggestions in simple English, such as rest, water, and observation; he discusses them as language practice and does not present them as medical treatment',
            'The learner comes back to discuss the doctor''s general recommendations. You ask whether they understand the suggestions, what they plan to do, and how they feel now, while clearly keeping the interaction as safe language practice.',
            true, 30, 40000, 4),
        ('clinic_city_pattern', 'clinic', 'A2', 'Заметить связь между жалобами жителей',
            'Dr. Elias Stone', 'elias_doctor', 'clinic_follow_advice',
            'a practical, calm doctor who has noticed several residents mention similar mild complaints; he speaks English clearly and treats the topic as everyday observation and pattern-finding, not epidemiology, diagnosis, or risky advice',
            'The clinic is quiet. You tell the learner that several neighbourhood residents described similar mild complaints, and you ask them to compare the stories and notice what might connect them in safe everyday language.',
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
-- 3. Tasks for the new quest scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- clinic_make_appointment
        ('clinic_make_appointment', 'greet',           0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Elias Stone.'),
        ('clinic_make_appointment', 'ask_appointment', 1, true,  'Попросить записать на приём',
            'The learner asks for an appointment or says they need to see the doctor.'),
        ('clinic_make_appointment', 'give_reason',     2, true,  'Кратко назвать причину',
            'The learner gives a simple non-emergency reason for the visit.'),
        ('clinic_make_appointment', 'choose_time',     3, true,  'Согласовать время',
            'The learner answers about a day, time, or availability for the appointment.'),
        ('clinic_make_appointment', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks Dr. Elias Stone and/or says goodbye.'),

        -- clinic_describe_pain
        ('clinic_describe_pain', 'greet',           0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Elias Stone.'),
        ('clinic_describe_pain', 'name_area',       1, true,  'Сказать, где болит',
            'The learner identifies a body area or general place of discomfort.'),
        ('clinic_describe_pain', 'describe_feeling',2, true,  'Описать боль или самочувствие',
            'The learner describes the discomfort or general feeling with a simple quality, intensity, or everyday word.'),
        ('clinic_describe_pain', 'answer_context',  3, true,  'Ответить на уточняющий вопрос',
            'The learner answers a follow-up question about everyday context, activity, or timing.'),
        ('clinic_describe_pain', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks Dr. Elias Stone and/or says goodbye.'),

        -- clinic_answer_questions
        ('clinic_answer_questions', 'greet',            0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Elias Stone.'),
        ('clinic_answer_questions', 'say_sleep',        1, true,  'Рассказать про сон',
            'The learner says something about recent sleep or rest.'),
        ('clinic_answer_questions', 'say_food',         2, true,  'Рассказать про еду',
            'The learner says something about meals, appetite, or food routine.'),
        ('clinic_answer_questions', 'say_temperature',  3, true,  'Рассказать про температуру',
            'The learner says something about body temperature, fever, feeling hot or cold, or a recent measurement.'),
        ('clinic_answer_questions', 'ask_next',         4, false, 'Спросить, что дальше',
            'The learner asks what happens next or asks for a safe general next step in the conversation.'),

        -- clinic_explain_history
        ('clinic_explain_history', 'greet',                 0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Elias Stone.'),
        ('clinic_explain_history', 'say_started',           1, true,  'Сказать, когда началось',
            'The learner says when the issue or situation started.'),
        ('clinic_explain_history', 'describe_before_after', 2, true,  'Сравнить раньше и сейчас',
            'The learner compares the earlier situation with the current situation.'),
        ('clinic_explain_history', 'mention_change',        3, true,  'Назвать изменение',
            'The learner mentions one change, event, habit, or circumstance connected to the timeline.'),
        ('clinic_explain_history', 'thank',                 4, false, 'Поблагодарить',
            'The learner thanks Dr. Elias Stone and/or says goodbye.'),

        -- clinic_follow_advice
        ('clinic_follow_advice', 'greet',              0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Elias Stone.'),
        ('clinic_follow_advice', 'repeat_advice',      1, true,  'Пересказать рекомендацию',
            'The learner repeats or summarizes at least one harmless general recommendation the doctor gave.'),
        ('clinic_follow_advice', 'ask_clarification',  2, true,  'Уточнить рекомендацию',
            'The learner asks a clarifying question about a recommendation or says what they did not understand.'),
        ('clinic_follow_advice', 'say_plan',           3, true,  'Сказать, что планирует делать',
            'The learner says what safe everyday action or routine change they plan to try.'),
        ('clinic_follow_advice', 'thank',              4, false, 'Поблагодарить',
            'The learner thanks Dr. Elias Stone and/or says goodbye.'),

        -- clinic_city_pattern
        ('clinic_city_pattern', 'greet',             0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Elias Stone.'),
        ('clinic_city_pattern', 'hear_complaints',   1, true,  'Выслушать жалобы жителей',
            'The learner listens to or acknowledges the complaints other residents described.'),
        ('clinic_city_pattern', 'name_similarity',   2, true,  'Назвать общую деталь',
            'The learner names one detail, symptom, timing, or situation that appears in more than one complaint.'),
        ('clinic_city_pattern', 'suggest_connection',3, true,  'Предположить связь',
            'The learner suggests a possible everyday connection between the residents'' complaints.'),
        ('clinic_city_pattern', 'react',             4, false, 'Отреагировать на наблюдение',
            'The learner reacts with a simple opinion, feeling, or question about the shared pattern.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'en_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
