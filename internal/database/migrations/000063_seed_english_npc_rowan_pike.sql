-- Seed English-only NPC quest chain: Professor Rowan Pike, a B1 university teacher.
-- NPC speaks English; titles are RU, persona/scene/criteria are model-facing EN instructions.
-- All scenarios are pinned to B1 (first from B1-B2); no A2 scenarios.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('university_join_class',      'B1', 'Представиться и объяснить интерес к курсу'),
        ('university_ask_assignment', 'B1', 'Уточнить задание'),
        ('university_present_legend', 'B1', 'Кратко пересказать городскую легенду'),
        ('university_debate_sources', 'B1', 'Обсудить, можно ли доверять источникам'),
        ('university_interview_plan', 'B1', 'Составить план интервью'),
        ('university_final_theory',   'B1', 'Защитить свою версию тайны')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Professor Rowan Pike scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('university_join_class', 'university', 'B1', 'Представиться и объяснить интерес к курсу',
            'Professor Rowan Pike', 'rowan_professor', NULL,
            'a thoughtful folklore and urban legends professor who teaches in clear English, encourages B1 learners to explain ideas, ask questions, and treat city stories as both culture and possible clues',
            'The learner joins Professor Rowan Pike''s university class on urban legends. Some local stories may be more than fiction, and students help collect and examine them. You welcome the learner and ask them to introduce themself and explain their interest in the course.',
            true, 30, 40000, 0),
        ('university_ask_assignment', 'university', 'B1', 'Уточнить задание',
            'Professor Rowan Pike', 'rowan_professor', 'university_join_class',
            'a thoughtful folklore professor who explains course work clearly in English and helps students understand what evidence and storytelling mean in legend research',
            'After the first class, the learner needs to clarify the fieldwork assignment for collecting urban legends. You explain what the task involves and help the learner check deadlines, expectations, and what kind of stories matter.',
            true, 30, 40000, 1),
        ('university_present_legend', 'university', 'B1', 'Кратко пересказать городскую легенду',
            'Professor Rowan Pike', 'rowan_professor', 'university_ask_assignment',
            'a thoughtful folklore professor who listens for structure, detail, and uncertainty when learners retell legends in clear English',
            'The learner has heard or found an urban legend connected to the city. In seminar, you ask them to retell it briefly, highlight the key details, and say why it feels memorable or suspicious.',
            true, 30, 40000, 2),
        ('university_debate_sources', 'university', 'B1', 'Обсудить, можно ли доверять источникам',
            'Professor Rowan Pike', 'rowan_professor', 'university_present_legend',
            'a thoughtful folklore professor who guides learners through source reliability, bias, and missing evidence using accessible English',
            'The learner is working with different accounts of the same legend. You ask them to discuss whether interviews, old articles, rumours, or online posts can be trusted and what makes one source stronger than another.',
            true, 30, 40000, 3),
        ('university_interview_plan', 'university', 'B1', 'Составить план интервью',
            'Professor Rowan Pike', 'rowan_professor', 'university_debate_sources',
            'a thoughtful folklore professor who prepares students for respectful interviews and keeps the conversation in clear English',
            'The learner will interview someone who may know part of the legend. You help them plan the goal of the interview, choose questions, decide how to begin politely, and think about what answers would matter most.',
            true, 30, 40000, 4),
        ('university_final_theory', 'university', 'B1', 'Защитить свою версию тайны',
            'Professor Rowan Pike', 'rowan_professor', 'university_interview_plan',
            'a thoughtful folklore professor who asks supportive follow-up questions and helps learners defend a reasoned theory in English when a legend may hide a real mystery',
            'The learner has formed their own version of what the urban legend might mean or hide. You ask them to defend their theory, support it with details from stories and sources, respond to a gentle challenge, and refine the idea.',
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
        -- university_join_class
        ('university_join_class', 'greet_professor',       0, true,  'Поздороваться с профессором',
            'The learner greets Professor Rowan Pike appropriately for a university class setting.'),
        ('university_join_class', 'introduce_self',        1, true,  'Кратко представиться',
            'The learner introduces themself with relevant personal or study information.'),
        ('university_join_class', 'state_interest',        2, true,  'Объяснить интерес к курсу',
            'The learner explains why they are interested in urban legends, folklore, or the course topic.'),
        ('university_join_class', 'ask_course_question',   3, true,  'Задать вопрос о курсе',
            'The learner asks a relevant question about the course, expectations, or legend research.'),
        ('university_join_class', 'confirm_participation', 4, false, 'Подтвердить участие',
            'The learner confirms they understand the next step and are ready to participate.'),

        -- university_ask_assignment
        ('university_ask_assignment', 'recall_assignment',  0, true,  'Уточнить задание',
            'The learner checks or summarizes what the fieldwork or legend-collection assignment requires.'),
        ('university_ask_assignment', 'ask_deadline',       1, true,  'Уточнить сроки',
            'The learner asks about deadlines, submission format, or when the work is due.'),
        ('university_ask_assignment', 'clarify_expectations',2, true, 'Уточнить ожидания',
            'The learner asks what kind of stories, evidence, or detail the professor expects.'),
        ('university_ask_assignment', 'ask_about_sources',  3, true,  'Спросить про источники',
            'The learner asks whether interviews, archives, rumours, or other sources are acceptable.'),
        ('university_ask_assignment', 'agree_next_step',    4, false, 'Согласовать следующий шаг',
            'The learner agrees on a next step for starting the assignment.'),

        -- university_present_legend
        ('university_present_legend', 'name_legend',        0, true,  'Назвать легенду',
            'The learner names or identifies the urban legend they will retell.'),
        ('university_present_legend', 'summarize_story',    1, true,  'Кратко пересказать сюжет',
            'The learner retells the main events of the legend in a clear, concise way.'),
        ('university_present_legend', 'mention_setting',     2, true,  'Указать место и время',
            'The learner mentions where and when the legend is set or where it is told in the city.'),
        ('university_present_legend', 'highlight_detail',    3, true,  'Выделить важную деталь',
            'The learner highlights a detail that makes the legend memorable, strange, or suspicious.'),
        ('university_present_legend', 'say_why_interesting', 4, false, 'Объяснить, почему легенда интересна',
            'The learner explains why the legend feels worth studying or may connect to something real.'),

        -- university_debate_sources
        ('university_debate_sources', 'name_source_types',  0, true,  'Назвать типы источников',
            'The learner names at least one type of source used for the legend, such as an interview, rumour, article, or online post.'),
        ('university_debate_sources', 'discuss_reliability',1, true,  'Обсудить надёжность',
            'The learner discusses whether a source can be trusted and gives a reason.'),
        ('university_debate_sources', 'compare_sources',    2, true,  'Сравнить источники',
            'The learner compares two sources or accounts and notes a strength or weakness.'),
        ('university_debate_sources', 'mention_missing_info',3, true, 'Отметить пробелы в данных',
            'The learner identifies missing information, bias, or uncertainty in the available sources.'),
        ('university_debate_sources', 'state_opinion',      4, false, 'Высказать мнение о доверии',
            'The learner states which source seems more trustworthy or useful and explains why.'),

        -- university_interview_plan
        ('university_interview_plan', 'state_goal',         0, true,  'Сформулировать цель интервью',
            'The learner states what they want to learn from the interview about the legend.'),
        ('university_interview_plan', 'choose_interviewee', 1, true,  'Выбрать собеседника',
            'The learner identifies a suitable person to interview, such as a local resident, witness, or storyteller.'),
        ('university_interview_plan', 'prepare_questions',  2, true,  'Подготовить вопросы',
            'The learner proposes several interview questions that could reveal useful information.'),
        ('university_interview_plan', 'plan_opening',       3, true,  'Продумать вежливое начало',
            'The learner explains how they will begin the interview politely and respectfully.'),
        ('university_interview_plan', 'plan_notes',         4, false, 'Решить, как фиксировать ответы',
            'The learner says how they will remember, record, or organize the answers.'),

        -- university_final_theory
        ('university_final_theory', 'state_theory',         0, true,  'Сформулировать версию тайны',
            'The learner states a clear theory or interpretation of what the urban legend may mean or hide.'),
        ('university_final_theory', 'give_evidence',        1, true,  'Привести подтверждение',
            'The learner supports the theory with at least one detail from the legend, interview, or source comparison.'),
        ('university_final_theory', 'answer_challenge',     2, true,  'Ответить на возражение',
            'The learner responds to a gentle challenge or alternative interpretation from the professor.'),
        ('university_final_theory', 'refine_theory',        3, true,  'Уточнить позицию',
            'The learner improves, narrows, or clarifies the theory after discussion.'),
        ('university_final_theory', 'close_defense',        4, false, 'Завершить защиту теории',
            'The learner closes the defense with a final thought, reflection, or summary of their version of the mystery.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'en_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
