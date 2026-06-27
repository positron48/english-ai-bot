-- Seed Spanish-only NPC quest chain: Profesor Martin Sol, a B1 university teacher.
-- NPC speaks Spanish; titles are RU, persona/scene/criteria are model-facing EN instructions.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('university_martin_join_seminar',      'B1', 'Представиться на семинаре'),
        ('university_martin_explain_topic',     'B1', 'Выбрать тему мини-доклада'),
        ('university_martin_interview_neighbor','B1', 'Подготовить вопросы для интервью'),
        ('university_martin_compare_versions',  'B1', 'Сравнить две версии легенды'),
        ('university_martin_defend_argument',   'B1', 'Защитить аргумент'),
        ('university_martin_public_talk',       'B1', 'Выступить с выводом')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Profesor Martin Sol scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('university_martin_join_seminar', 'university', 'B1', 'Представиться на семинаре',
            'Profesor Martín Sol', 'martin_professor', NULL,
            'a thoughtful cultural studies teacher who conducts the seminar in clear Spanish and encourages B1 learners to explain ideas, ask questions, and connect city stories with everyday life',
            'The learner joins Profesor Martin Sol''s university seminar about city stories connected to the plaza, the market, and the old theatre. You welcome them and ask them to introduce themself and their interest in the seminar.',
            true, 30, 40000, 0),
        ('university_martin_explain_topic', 'university', 'B1', 'Выбрать тему мини-доклада',
            'Profesor Martín Sol', 'martin_professor', 'university_martin_join_seminar',
            'a thoughtful cultural studies teacher who helps students choose focused research topics and express reasons in clear Spanish',
            'After the first seminar, the learner must choose a short presentation topic. You offer several possible directions connected to the plaza, the market, and the old theatre, then ask the learner to choose and explain why.',
            true, 30, 40000, 1),
        ('university_martin_interview_neighbor', 'university', 'B1', 'Подготовить вопросы для интервью',
            'Profesor Martín Sol', 'martin_professor', 'university_martin_explain_topic',
            'a thoughtful cultural studies teacher who prepares students for respectful interviews and keeps the conversation in clear Spanish',
            'The learner will interview a local resident about a city story. You help them prepare respectful questions, decide what information matters, and plan how to begin and end the interview.',
            true, 30, 40000, 2),
        ('university_martin_compare_versions', 'university', 'B1', 'Сравнить две версии легенды',
            'Profesor Martín Sol', 'martin_professor', 'university_martin_interview_neighbor',
            'a thoughtful cultural studies teacher who guides learners through comparison, uncertainty, and evidence using accessible Spanish',
            'The learner has heard two different versions of the same local legend. In seminar, you ask them to compare the versions and notice what is similar, what is different, and which details may matter.',
            true, 30, 40000, 3),
        ('university_martin_defend_argument', 'university', 'B1', 'Защитить аргумент',
            'Profesor Martín Sol', 'martin_professor', 'university_martin_compare_versions',
            'a thoughtful cultural studies teacher who asks supportive follow-up questions and helps learners defend a simple argument in Spanish',
            'The learner has formed an interpretation of the legend. You ask them to state their argument, support it with details, respond to a gentle challenge, and refine the idea.',
            true, 30, 40000, 4),
        ('university_martin_public_talk', 'university', 'B1', 'Выступить с выводом',
            'Profesor Martín Sol', 'martin_professor', 'university_martin_defend_argument',
            'a thoughtful cultural studies teacher who helps learners present conclusions clearly and confidently in Spanish',
            'It is the final seminar meeting. The learner gives a short public conclusion about the city story. You help them organize the ending, address the audience, and reflect on what they learned.',
            true, 30, 40000, 5)
)
INSERT INTO conversation_scenarios
    (course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
     title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest,
     max_turns, token_budget, sort_order, status)
SELECT c.id, d.id, l.id, li.id, s.code, s.place_type, s.cefr_level,
       s.title, s.npc_name, s.npc_code, COALESCE(s.prerequisite_code, ''), s.npc_persona, s.scene_setup,
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
        -- university_martin_join_seminar
        ('university_martin_join_seminar', 'greet_professor',      0, true,  'Поздороваться с профессором',
            'The learner greets Profesor Martin Sol appropriately for a seminar setting.'),
        ('university_martin_join_seminar', 'introduce_self',       1, true,  'Кратко представиться',
            'The learner introduces themself with relevant personal or study information.'),
        ('university_martin_join_seminar', 'state_interest',       2, true,  'Объяснить интерес к семинару',
            'The learner explains why they are interested in city stories, culture, or the seminar topic.'),
        ('university_martin_join_seminar', 'ask_seminar_question', 3, true,  'Задать вопрос о семинаре',
            'The learner asks a relevant question about the seminar, expectations, or course work.'),
        ('university_martin_join_seminar', 'confirm_participation',4, false, 'Подтвердить участие',
            'The learner confirms they understand the next step and are ready to participate.'),

        -- university_martin_explain_topic
        ('university_martin_explain_topic', 'recall_assignment', 0, true,  'Уточнить задание',
            'The learner checks or summarizes what the mini-presentation assignment requires.'),
        ('university_martin_explain_topic', 'name_topic',        1, true,  'Назвать тему',
            'The learner names a possible topic for the mini-presentation.'),
        ('university_martin_explain_topic', 'explain_reason',    2, true,  'Объяснить выбор',
            'The learner gives a reason for choosing the topic.'),
        ('university_martin_explain_topic', 'narrow_focus',      3, true,  'Сузить фокус темы',
            'The learner narrows the topic to a manageable question, place, person, or event.'),
        ('university_martin_explain_topic', 'agree_next_step',   4, false, 'Согласовать следующий шаг',
            'The learner agrees on a next step for preparing the mini-presentation.'),

        -- university_martin_interview_neighbor
        ('university_martin_interview_neighbor', 'state_goal',      0, true,  'Сформулировать цель интервью',
            'The learner states what they want to learn from the interview.'),
        ('university_martin_interview_neighbor', 'choose_person',   1, true,  'Выбрать собеседника',
            'The learner identifies a suitable local resident, neighbour, or witness to interview.'),
        ('university_martin_interview_neighbor', 'prepare_questions',2, true, 'Подготовить вопросы',
            'The learner proposes several interview questions that invite useful information.'),
        ('university_martin_interview_neighbor', 'plan_politeness', 3, true,  'Продумать вежливое начало',
            'The learner explains how they will begin the interview politely and respectfully.'),
        ('university_martin_interview_neighbor', 'plan_notes',      4, false, 'Решить, как фиксировать ответы',
            'The learner says how they will remember, record, or organize the answers.'),

        -- university_martin_compare_versions
        ('university_martin_compare_versions', 'summarize_first',  0, true,  'Кратко пересказать первую версию',
            'The learner summarizes the first version of the legend or story.'),
        ('university_martin_compare_versions', 'summarize_second', 1, true,  'Кратко пересказать вторую версию',
            'The learner summarizes the second version of the legend or story.'),
        ('university_martin_compare_versions', 'name_similarity',  2, true,  'Назвать сходство',
            'The learner identifies an important similarity between the two versions.'),
        ('university_martin_compare_versions', 'name_difference',  3, true,  'Назвать различие',
            'The learner identifies an important difference between the two versions.'),
        ('university_martin_compare_versions', 'say_preference',   4, false, 'Сказать, какая версия убедительнее',
            'The learner says which version seems more convincing, memorable, or useful and gives a reason.'),

        -- university_martin_defend_argument
        ('university_martin_defend_argument', 'state_argument',   0, true,  'Сформулировать аргумент',
            'The learner states a clear interpretation or argument about the city story.'),
        ('university_martin_defend_argument', 'give_evidence',    1, true,  'Привести подтверждение',
            'The learner supports the argument with at least one detail from the story, interview, or comparison.'),
        ('university_martin_defend_argument', 'answer_challenge', 2, true,  'Ответить на возражение',
            'The learner responds to a gentle challenge or alternative interpretation.'),
        ('university_martin_defend_argument', 'refine_argument',  3, true,  'Уточнить позицию',
            'The learner improves, narrows, or clarifies the argument after discussion.'),
        ('university_martin_defend_argument', 'ask_feedback',     4, false, 'Попросить обратную связь',
            'The learner asks for feedback or confirms how to strengthen the argument.'),

        -- university_martin_public_talk
        ('university_martin_public_talk', 'open_talk',       0, true,  'Начать выступление',
            'The learner opens the public talk and introduces the topic.'),
        ('university_martin_public_talk', 'present_findings',1, true,  'Представить основные выводы',
            'The learner presents the main findings or conclusion in an organized way.'),
        ('university_martin_public_talk', 'mention_sources', 2, true,  'Упомянуть источники',
            'The learner refers to interviews, story versions, observations, or other sources used.'),
        ('university_martin_public_talk', 'answer_question', 3, true,  'Ответить на вопрос',
            'The learner answers at least one follow-up question from the professor or imagined audience.'),
        ('university_martin_public_talk', 'close_talk',      4, false, 'Завершить выступление',
            'The learner closes the talk with a final thought, reflection, or thanks.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'es_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
