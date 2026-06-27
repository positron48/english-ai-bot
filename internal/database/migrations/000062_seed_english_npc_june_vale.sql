-- Seed English course NPC Dr. June Vale, a calm community-center therapist for A2 conversation quests.
-- Applies only to en_ru; June speaks English and all scenarios are pinned to A2.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('therapy_first_visit',    'A2', 'Рассказать имя, настроение и причину визита'),
        ('therapy_daily_routine',  'A2', 'Описать обычный день'),
        ('therapy_small_worry',    'A2', 'Рассказать о небольшой тревоге'),
        ('therapy_old_habit',      'A2', 'Объяснить привычку и трудности изменения'),
        ('therapy_set_goal',       'A2', 'Сформулировать цель на неделю'),
        ('therapy_letter_to_self', 'A2', 'Обсудить письмо себе из будущего')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Dr. June Vale therapy quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('therapy_first_visit', 'community_center', 'A2', 'Рассказать имя, настроение и причину визита',
            'Dr. June Vale', 'june_therapist', '',
            'a calm, warm city-center psychologist who speaks clear A2 English, asks gentle questions, and helps the learner introduce themselves and name simple feelings without pressure',
            'The learner comes to a first short therapy visit at the community center. You welcome them calmly, explain that this is a safe and simple conversation, and ask for their name, how they feel today, and why they came.',
            true, 30, 40000, 0),
        ('therapy_daily_routine', 'community_center', 'A2', 'Описать обычный день',
            'Dr. June Vale', 'june_therapist', 'therapy_first_visit',
            'a calm community-center psychologist who speaks clear A2 English and helps the learner describe an ordinary day in simple, concrete language',
            'The learner returns for another session. You ask about their typical day and invite them to walk through morning, work or study, and evening in simple steps.',
            true, 30, 40000, 1),
        ('therapy_small_worry', 'community_center', 'A2', 'Рассказать о небольшой тревоге',
            'Dr. June Vale', 'june_therapist', 'therapy_daily_routine',
            'a patient psychologist who speaks clear A2 English, listens carefully, and helps the learner talk about a small worry without making it feel huge',
            'The learner wants to talk about a small worry. You ask gentle follow-up questions about what bothers them, when it happens, and how it feels.',
            true, 30, 40000, 2),
        ('therapy_old_habit', 'community_center', 'A2', 'Объяснить привычку и трудности изменения',
            'Dr. June Vale', 'june_therapist', 'therapy_small_worry',
            'a supportive psychologist who speaks clear A2 English and helps the learner explain an old habit and why change feels difficult',
            'The learner wants to discuss a habit that is hard to change. You ask when it happens, what triggers it, and what makes changing it difficult.',
            true, 30, 40000, 3),
        ('therapy_set_goal', 'community_center', 'A2', 'Сформулировать цель на неделю',
            'Dr. June Vale', 'june_therapist', 'therapy_old_habit',
            'a practical, encouraging psychologist who speaks clear A2 English and helps the learner turn one weekly wish into a realistic goal',
            'The learner wants to set a goal for the week. You help them choose one achievable goal, explain why it matters, and name a first small step.',
            true, 30, 40000, 4),
        ('therapy_letter_to_self', 'community_center', 'A2', 'Обсудить письмо себе из будущего',
            'Dr. June Vale', 'june_therapist', 'therapy_set_goal',
            'a warm psychologist who speaks clear A2 English, handles reflection gently, and helps the learner discuss a letter from their future self',
            'The learner has written or imagined a letter from their future self. You ask what the letter says, what advice or encouragement it gives, and how reading it feels today.',
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
        -- therapy_first_visit
        ('therapy_first_visit', 'greet',        0, true,  'Поздороваться с доктором Вейл',
            'The learner greets Dr. Vale and starts the session politely.'),
        ('therapy_first_visit', 'name',         1, true,  'Назвать своё имя',
            'The learner gives their name or introduces themselves.'),
        ('therapy_first_visit', 'mood',         2, true,  'Описать настроение',
            'The learner names or describes how they feel today.'),
        ('therapy_first_visit', 'visit_reason', 3, true,  'Объяснить причину визита',
            'The learner gives a simple reason for coming to the session.'),
        ('therapy_first_visit', 'thank',        4, false, 'Поблагодарить',
            'The learner thanks Dr. Vale and/or closes the session politely.'),

        -- therapy_daily_routine
        ('therapy_daily_routine', 'greet',           0, true,  'Поздороваться с доктором Вейл',
            'The learner greets Dr. Vale and continues the therapy conversation.'),
        ('therapy_daily_routine', 'describe_day',    1, true,  'Описать обычный день',
            'The learner describes a typical day with at least two time periods or activities.'),
        ('therapy_daily_routine', 'routine_detail',  2, true,  'Рассказать про утро или вечер',
            'The learner gives simple details about one part of their daily routine.'),
        ('therapy_daily_routine', 'routine_feeling', 3, true,  'Сказать, что нравится или утомляет',
            'The learner says what they like, dislike, or find tiring about their routine.'),
        ('therapy_daily_routine', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks Dr. Vale and/or closes the session politely.'),

        -- therapy_small_worry
        ('therapy_small_worry', 'greet',          0, true,  'Поздороваться с доктором Вейл',
            'The learner greets Dr. Vale and agrees to talk about a worry.'),
        ('therapy_small_worry', 'name_worry',     1, true,  'Назвать небольшую тревогу',
            'The learner identifies a small worry or concern.'),
        ('therapy_small_worry', 'describe_worry', 2, true,  'Рассказать, о чём беспокоишься',
            'The learner describes what the worry is about in simple terms.'),
        ('therapy_small_worry', 'say_feeling',    3, true,  'Описать чувство',
            'The learner says how the worry makes them feel or when it appears.'),
        ('therapy_small_worry', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Dr. Vale and/or closes the session politely.'),

        -- therapy_old_habit
        ('therapy_old_habit', 'greet',         0, true,  'Поздороваться с доктором Вейл',
            'The learner greets Dr. Vale and agrees to talk about a habit.'),
        ('therapy_old_habit', 'name_habit',    1, true,  'Назвать привычку',
            'The learner names or identifies a habit they want to discuss.'),
        ('therapy_old_habit', 'explain_habit', 2, true,  'Объяснить, в чём она заключается',
            'The learner describes what the habit involves or when it happens.'),
        ('therapy_old_habit', 'say_difficult', 3, true,  'Сказать, почему трудно изменить',
            'The learner explains why the habit is hard to change or what makes change difficult.'),
        ('therapy_old_habit', 'thank',         4, false, 'Поблагодарить',
            'The learner thanks Dr. Vale and/or closes the session politely.'),

        -- therapy_set_goal
        ('therapy_set_goal', 'greet',       0, true,  'Поздороваться с доктором Вейл',
            'The learner greets Dr. Vale and says they want to set a weekly goal.'),
        ('therapy_set_goal', 'choose_goal', 1, true,  'Выбрать цель на неделю',
            'The learner names one achievable goal for the coming week.'),
        ('therapy_set_goal', 'say_reason',  2, true,  'Объяснить, зачем это важно',
            'The learner explains why the goal matters to them.'),
        ('therapy_set_goal', 'plan_step',   3, true,  'Назвать первый шаг',
            'The learner proposes a first small step, time, or action toward the goal.'),
        ('therapy_set_goal', 'thank',       4, false, 'Поблагодарить',
            'The learner thanks Dr. Vale and/or closes the session politely.'),

        -- therapy_letter_to_self
        ('therapy_letter_to_self', 'greet',           0, true,  'Поздороваться с доктором Вейл',
            'The learner greets Dr. Vale and agrees to discuss a letter from their future self.'),
        ('therapy_letter_to_self', 'describe_letter',1, true, 'Описать письмо из будущего',
            'The learner describes the letter or what their future self wrote about.'),
        ('therapy_letter_to_self', 'share_message',  2, true,  'Рассказать, что пишет будущий ты',
            'The learner shares the main message, advice, or encouragement from the letter.'),
        ('therapy_letter_to_self', 'say_feeling',    3, true,  'Описать чувство от письма',
            'The learner describes how reading or imagining the letter makes them feel.'),
        ('therapy_letter_to_self', 'thank',          4, false, 'Поблагодарить',
            'The learner thanks Dr. Vale and/or closes the session politely.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'en_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
