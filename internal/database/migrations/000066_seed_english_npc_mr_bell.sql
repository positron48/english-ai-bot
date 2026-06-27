-- Seed English-only NPC chain: Mr. Bell (the clock man).
-- UI titles are Russian; persona/scene/criteria are model-facing English instructions.
-- Although the story spans A1-B1 themes, every scenario is intentionally anchored to
-- the first English course level (A1) and the conversation location.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new English scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('clockman_first_warning', 'A1', 'Понять простое предупреждение'),
        ('clockman_ask_time',      'A1', 'Спросить время и дорогу'),
        ('clockman_lost_day',      'A1', 'Выслушать рассказ о потерянном дне'),
        ('clockman_find_photo',    'A1', 'Описать старую фотографию'),
        ('clockman_memory_story',  'A1', 'Обсудить воспоминание и чувства'),
        ('clockman_last_chime',    'A1', 'Помочь решить, что делать дальше')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. English-only quest scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('clockman_first_warning', 'town_square', 'A1', 'Понять простое предупреждение',
            'Mr. Bell', 'bell_clockman', '',
            'a strange but kind English-speaking man who tends an old town clock; he speaks in gentle riddles and simple metaphors, keeps the city''s memory, and stays lonely but never scary, magical, threatening, or frightening',
            'The learner meets you beside the old clock in the town square. You stop them with a simple, safe warning about an everyday inconvenience, such as wet steps, a closed path, or a busy corner.',
            true, 30, 40000, 0),
        ('clockman_ask_time', 'town_square', 'A1', 'Спросить время и дорогу',
            'Mr. Bell', 'bell_clockman', 'clockman_first_warning',
            'a strange but kind English-speaking clock man who answers simple questions about time and directions with patient, concrete landmarks; he may sound poetic but stays clear, safe, and mildly odd',
            'The learner meets you again by the clock. They need the time and directions, and you help with simple words while the square is calm around you.',
            true, 30, 40000, 1),
        ('clockman_lost_day', 'town_square', 'A1', 'Выслушать рассказ о потерянном дне',
            'Mr. Bell', 'bell_clockman', 'clockman_ask_time',
            'a strange but kind English-speaking clock man who feels he has lost a day from his life; he speaks in soft riddles but tells a grounded, human story without fear, horror, or supernatural danger',
            'The learner finds you beside the clock on a quiet afternoon. You say, in simple English, that one day seems missing from your memory, and you invite the learner to listen.',
            true, 30, 40000, 2),
        ('clockman_find_photo', 'street', 'A1', 'Описать старую фотографию',
            'Mr. Bell', 'bell_clockman', 'clockman_lost_day',
            'a strange but kind English-speaking clock man who notices small forgotten things and connects them to the city''s memory; he stays curious, safe, and gently lonely',
            'The learner finds an old photograph near a street bench close to the square. You ask them to describe what they see and help decide what the picture might mean.',
            true, 30, 40000, 3),
        ('clockman_memory_story', 'town_square', 'A1', 'Обсудить воспоминание и чувства',
            'Mr. Bell', 'bell_clockman', 'clockman_find_photo',
            'a strange but kind English-speaking clock man who remembers people and places with tenderness, not sadness or fear; he accepts simple learner language while inviting gentle feeling words',
            'The photograph reminds you of someone from the old square. You tell the memory slowly and ask the learner what memories can feel like and why they matter.',
            true, 30, 40000, 4),
        ('clockman_last_chime', 'town_square', 'A1', 'Помочь решить, что делать дальше',
            'Mr. Bell', 'bell_clockman', 'clockman_memory_story',
            'a strange but kind English-speaking clock man who is deciding what to do with an old memory and the clock''s next chime; he speaks warmly, keeps the ending hopeful and ordinary, and never suggests magic or danger',
            'Evening is coming to the square. You wonder whether to keep waiting, share the memory, or take one small next step. The learner helps you choose what to do next.',
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
-- 3. Tasks for the quest scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- clockman_first_warning
        ('clockman_first_warning', 'greet',              0, true,  'Поздороваться с часовщиком',
            'The learner greets the NPC.'),
        ('clockman_first_warning', 'ask_warning',        1, true,  'Спросить, в чём предупреждение',
            'The learner asks what the warning is about or why the NPC is stopping them.'),
        ('clockman_first_warning', 'show_understanding', 2, true,  'Показать, что предупреждение понятно',
            'The learner shows they understand the simple warning or repeats the practical point.'),
        ('clockman_first_warning', 'choose_action',      3, true,  'Выбрать безопасное действие',
            'The learner says what they will do next in response to the warning.'),
        ('clockman_first_warning', 'thank',              4, false, 'Поблагодарить',
            'The learner thanks the NPC and/or says goodbye.'),

        -- clockman_ask_time
        ('clockman_ask_time', 'greet',            0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('clockman_ask_time', 'ask_time',         1, true,  'Спросить, который час',
            'The learner asks what time it is or asks about the time.'),
        ('clockman_ask_time', 'ask_direction',    2, true,  'Спросить дорогу',
            'The learner asks how to get to a place or asks where a place is.'),
        ('clockman_ask_time', 'confirm_info',     3, true,  'Подтвердить время или маршрут',
            'The learner confirms, repeats, or shows they understood the time or directions.'),
        ('clockman_ask_time', 'thank',            4, false, 'Поблагодарить',
            'The learner thanks the NPC and/or says goodbye.'),

        -- clockman_lost_day
        ('clockman_lost_day', 'greet',           0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('clockman_lost_day', 'notice_mood',     1, true,  'Заметить, что с ним что-то не так',
            'The learner notices that the NPC seems troubled, confused, or different today.'),
        ('clockman_lost_day', 'listen_respond',  2, true,  'Показать, что слушаешь',
            'The learner responds to the lost-day story with a simple reaction, question, or acknowledgment that they are listening.'),
        ('clockman_lost_day', 'ask_about_day',   3, true,  'Спросить о потерянном дне',
            'The learner asks what day is missing, what happened, or asks a simple follow-up about the story.'),
        ('clockman_lost_day', 'respond_kindly',  4, false, 'Ответить с поддержкой',
            'The learner responds to the NPC in a kind or supportive way.'),

        -- clockman_find_photo
        ('clockman_find_photo', 'greet',           0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('clockman_find_photo', 'mention_photo',   1, true,  'Сказать о найденной фотографии',
            'The learner says they found a photograph or points out the photograph.'),
        ('clockman_find_photo', 'describe_photo',  2, true,  'Описать фотографию',
            'The learner describes the photograph by appearance, people, place, age, condition, or another visible detail.'),
        ('clockman_find_photo', 'ask_about_photo', 3, true,  'Спросить, что на фото',
            'The learner asks who is in the photo, where it was taken, or what it might mean.'),
        ('clockman_find_photo', 'suggest_next',    4, false, 'Предложить следующий шаг',
            'The learner suggests a simple next action for the photograph.'),

        -- clockman_memory_story
        ('clockman_memory_story', 'greet',           0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('clockman_memory_story', 'ask_memory',      1, true,  'Спросить о воспоминании',
            'The learner asks about the memory, the person in the photo, or what happened in the past.'),
        ('clockman_memory_story', 'summarize_story', 2, true,  'Кратко пересказать историю',
            'The learner summarizes or restates the main idea of the memory in their own words.'),
        ('clockman_memory_story', 'share_feeling',   3, true,  'Высказать чувство или мнение',
            'The learner gives a simple feeling, opinion, or interpretation about the memory.'),
        ('clockman_memory_story', 'respond_kindly',  4, false, 'Ответить с поддержкой',
            'The learner responds to the NPC in a kind or supportive way.'),

        -- clockman_last_chime
        ('clockman_last_chime', 'greet',            0, true,  'Поздороваться',
            'The learner greets the NPC.'),
        ('clockman_last_chime', 'notice_situation', 1, true,  'Заметить, что он колеблется',
            'The learner notices that the NPC is unsure what to do next or asks what he is thinking about.'),
        ('clockman_last_chime', 'suggest_action',   2, true,  'Предложить простое действие',
            'The learner suggests a simple, safe action the NPC can take next.'),
        ('clockman_last_chime', 'support_decision', 3, true,  'Поддержать его решение',
            'The learner supports the NPC''s choice or helps him feel okay about the next step.'),
        ('clockman_last_chime', 'farewell',         4, false, 'Попрощаться тепло',
            'The learner thanks the NPC and/or says goodbye in a warm way.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'en_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
