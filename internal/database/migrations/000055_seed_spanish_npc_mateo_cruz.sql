-- Seed Spanish-only Dr. Mateo Cruz clinic NPC chain.
-- Applies only to es_ru; Mateo speaks Spanish. Titles are RU, while persona,
-- scene setup, and completion criteria are model-facing EN instructions.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('clinic_mateo_appointment',   'A2', 'Записаться на приём'),
        ('clinic_mateo_pain',          'A2', 'Описать боль'),
        ('clinic_mateo_habits',        'A2', 'Ответить про сон, еду и работу'),
        ('clinic_mateo_when_started',  'A2', 'Рассказать историю болезни'),
        ('clinic_mateo_follow_up',     'A2', 'Обсудить, помогли ли советы'),
        ('clinic_mateo_shared_dream',  'A2', 'Пересказать общий сон')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Dr. Mateo Cruz clinic scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('clinic_mateo_appointment', 'clinic', 'A2', 'Записаться на приём',
            'Dr. Mateo Cruz', 'mateo_doctor', '',
            'a practical, calm doctor in a small market clinic who speaks Spanish clearly for A2 learners; he keeps the exchange as safe language practice and never gives real medical diagnoses, prescriptions, dosage advice, or urgent-care instructions',
            'The learner comes to the clinic reception area and wants a simple appointment. You are Dr. Mateo Cruz, available for a safe everyday conversation about scheduling, not for real medical care.',
            true, 30, 40000, 0),
        ('clinic_mateo_pain', 'clinic', 'A2', 'Описать боль',
            'Dr. Mateo Cruz', 'mateo_doctor', 'clinic_mateo_appointment',
            'a practical, calm doctor who asks simple symptom questions in Spanish and redirects anything serious to real professional help without giving treatment instructions',
            'The learner is in a basic clinic conversation and needs to describe discomfort. Keep the scene focused on language practice: body area, intensity, and everyday context, with no real diagnosis or treatment plan.',
            true, 30, 40000, 1),
        ('clinic_mateo_habits', 'clinic', 'A2', 'Ответить про сон, еду и работу',
            'Dr. Mateo Cruz', 'mateo_doctor', 'clinic_mateo_pain',
            'a practical, calm doctor who uses simple Spanish questions about daily habits; he gives only general, non-medical, low-risk lifestyle reflections and avoids personalized medical advice',
            'After hearing the learner describe discomfort, you ask safe everyday questions about sleep, food, work, and routine. This is language practice in a friendly clinic, not a real consultation.',
            true, 30, 40000, 2),
        ('clinic_mateo_when_started', 'clinic', 'A2', 'Рассказать историю болезни',
            'Dr. Mateo Cruz', 'mateo_doctor', 'clinic_mateo_habits',
            'a practical, calm doctor who helps the learner tell a simple health-related timeline in Spanish while avoiding diagnosis, medication choices, or risky advice',
            'The learner returns to explain when the problem started and what changed over time. Ask for a safe everyday timeline and keep the focus on narration and sequencing.',
            true, 30, 40000, 3),
        ('clinic_mateo_follow_up', 'clinic', 'A2', 'Обсудить, помогли ли советы',
            'Dr. Mateo Cruz', 'mateo_doctor', 'clinic_mateo_when_started',
            'a practical, calm doctor who checks in with simple Spanish follow-up questions; he discusses only harmless general suggestions like rest, water, and observation, and does not present them as medical treatment',
            'The learner comes back for a follow-up conversation. You ask whether harmless general suggestions helped and what changed, while clearly keeping the interaction as safe language practice.',
            true, 30, 40000, 4),
        ('clinic_mateo_shared_dream', 'clinic', 'A2', 'Пересказать сон и сравнить с другими',
            'Dr. Mateo Cruz', 'mateo_doctor', 'clinic_mateo_follow_up',
            'a practical, calm doctor who has noticed several patients mention the same strange but harmless dream; he speaks Spanish clearly and treats the topic as storytelling, not medical or psychological advice',
            'The clinic is quiet near the market. You tell the learner that several people described a similar strange dream, and you ask them to retell their dream and compare details in safe everyday language.',
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
        -- clinic_mateo_appointment
        ('clinic_mateo_appointment', 'greet',          0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Mateo Cruz.'),
        ('clinic_mateo_appointment', 'ask_appointment',1, true,  'Попросить записать на приём',
            'The learner asks for an appointment or says they need to see the doctor.'),
        ('clinic_mateo_appointment', 'give_reason',    2, true,  'Кратко назвать причину',
            'The learner gives a simple non-emergency reason for the visit.'),
        ('clinic_mateo_appointment', 'choose_time',    3, true,  'Согласовать время',
            'The learner answers about a day, time, or availability for the appointment.'),
        ('clinic_mateo_appointment', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Dr. Mateo Cruz and/or says goodbye.'),

        -- clinic_mateo_pain
        ('clinic_mateo_pain', 'greet',          0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Mateo Cruz.'),
        ('clinic_mateo_pain', 'name_area',      1, true,  'Сказать, где болит',
            'The learner identifies a body area or general place of discomfort.'),
        ('clinic_mateo_pain', 'describe_pain',  2, true,  'Описать боль',
            'The learner describes the discomfort with a simple quality, intensity, or feeling.'),
        ('clinic_mateo_pain', 'answer_context', 3, true,  'Ответить на уточняющий вопрос',
            'The learner answers a follow-up question about everyday context, activity, or timing.'),
        ('clinic_mateo_pain', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Dr. Mateo Cruz and/or says goodbye.'),

        -- clinic_mateo_habits
        ('clinic_mateo_habits', 'greet',       0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Mateo Cruz.'),
        ('clinic_mateo_habits', 'say_sleep',   1, true,  'Рассказать про сон',
            'The learner says something about recent sleep or rest.'),
        ('clinic_mateo_habits', 'say_food',    2, true,  'Рассказать про еду',
            'The learner says something about meals, appetite, or food routine.'),
        ('clinic_mateo_habits', 'say_work',    3, true,  'Рассказать про работу',
            'The learner says something about work, study, stress, or daily activity.'),
        ('clinic_mateo_habits', 'ask_next',    4, false, 'Спросить, что дальше',
            'The learner asks what happens next or asks for a safe general next step in the conversation.'),

        -- clinic_mateo_when_started
        ('clinic_mateo_when_started', 'greet',       0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Mateo Cruz.'),
        ('clinic_mateo_when_started', 'say_started', 1, true,  'Сказать, когда началось',
            'The learner says when the issue or situation started.'),
        ('clinic_mateo_when_started', 'describe_before_after', 2, true, 'Сравнить раньше и сейчас',
            'The learner compares the earlier situation with the current situation.'),
        ('clinic_mateo_when_started', 'mention_change', 3, true, 'Назвать изменение',
            'The learner mentions one change, event, habit, or circumstance connected to the timeline.'),
        ('clinic_mateo_when_started', 'thank',       4, false, 'Поблагодарить',
            'The learner thanks Dr. Mateo Cruz and/or says goodbye.'),

        -- clinic_mateo_follow_up
        ('clinic_mateo_follow_up', 'greet',        0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Mateo Cruz.'),
        ('clinic_mateo_follow_up', 'say_result',   1, true,  'Сказать, стало ли лучше',
            'The learner says whether the situation improved, stayed the same, or got worse.'),
        ('clinic_mateo_follow_up', 'say_action',   2, true,  'Сказать, что делал',
            'The learner says what safe everyday action or routine change they tried.'),
        ('clinic_mateo_follow_up', 'describe_now', 3, true,  'Описать самочувствие сейчас',
            'The learner describes their current general condition or feeling.'),
        ('clinic_mateo_follow_up', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Dr. Mateo Cruz and/or says goodbye.'),

        -- clinic_mateo_shared_dream
        ('clinic_mateo_shared_dream', 'greet',        0, true,  'Поздороваться с доктором',
            'The learner greets Dr. Mateo Cruz.'),
        ('clinic_mateo_shared_dream', 'retell_dream', 1, true,  'Пересказать сон',
            'The learner retells a simple dream or remembered scene.'),
        ('clinic_mateo_shared_dream', 'name_detail',  2, true,  'Назвать важную деталь',
            'The learner names one important detail, image, place, person, or object from the dream.'),
        ('clinic_mateo_shared_dream', 'compare_others', 3, true, 'Сравнить с чужими снами',
            'The learner compares their dream with what other people described or says what is similar or different.'),
        ('clinic_mateo_shared_dream', 'react',        4, false, 'Отреагировать на историю',
            'The learner reacts with a simple opinion, feeling, or question about the shared dream story.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'es_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
