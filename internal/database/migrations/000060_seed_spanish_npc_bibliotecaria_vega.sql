-- Seed Spanish-only NPC quest chain: Bibliotecaria Vega, a B1 library archivist.
-- NPC speaks Spanish; titles are RU, persona/scene/criteria are model-facing EN instructions.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('library_vega_card',           'B1', 'Оформить читательскую карточку'),
        ('library_vega_find_letter',    'B1', 'Описать письмо, которое ищешь'),
        ('library_vega_summarize_letter','B1', 'Пересказать содержание письма'),
        ('library_vega_infer_author',   'B1', 'Предположить, кто автор письма'),
        ('library_vega_ethics',         'B1', 'Обсудить этику чужих писем'),
        ('library_vega_unsent_truth',   'B1', 'Обсудить невысказанную правду')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Bibliotecaria Vega scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('library_vega_card', 'library', 'B1', 'Оформить читательскую карточку',
            'Bibliotecaria Vega', 'vega_librarian', NULL,
            'a careful city librarian and archive keeper who speaks clear Spanish, protects readers'' privacy, and guides B1 learners through respectful library conversations',
            'The learner visits the library for the first time. You welcome them at the archive desk, explain that the library also keeps a protected collection of unsent letters, and help them apply for a reader card without sharing private details from the archive.',
            true, 30, 40000, 0),
        ('library_vega_find_letter', 'library', 'B1', 'Описать письмо, которое ищешь',
            'Bibliotecaria Vega', 'vega_librarian', 'library_vega_card',
            'a careful librarian who helps readers describe research needs in Spanish while keeping private letters anonymous and handled only through library rules',
            'The learner now has a reader card and asks about the archive of unsent letters. You explain that only anonymized catalogue descriptions are available, then ask them to describe the kind of letter they are looking for by topic, period, mood, or visible catalogue details.',
            true, 30, 40000, 1),
        ('library_vega_summarize_letter', 'library', 'B1', 'Пересказать содержание письма',
            'Bibliotecaria Vega', 'vega_librarian', 'library_vega_find_letter',
            'a careful librarian who invites concise summaries in Spanish and reminds the learner to avoid exposing identifying or sensitive private information',
            'You provide an anonymized, permitted excerpt from an unsent letter. The learner must summarize the main situation, the writer''s feeling, and the possible message without revealing names, addresses, or private identifying details.',
            true, 30, 40000, 2),
        ('library_vega_infer_author', 'library', 'B1', 'Предположить, кто автор письма',
            'Bibliotecaria Vega', 'vega_librarian', 'library_vega_summarize_letter',
            'a careful librarian who encourages cautious hypotheses in Spanish, separates evidence from imagination, and never asks the learner to identify a real private person',
            'After the summary, you ask the learner to make a cautious, fictional inference about the anonymous writer: what kind of situation they might be in, what clues support that idea, and what remains uncertain. Keep the discussion respectful and non-invasive.',
            true, 30, 40000, 3),
        ('library_vega_ethics', 'library', 'B1', 'Обсудить этику чужих писем',
            'Bibliotecaria Vega', 'vega_librarian', 'library_vega_infer_author',
            'a careful librarian who leads a balanced Spanish discussion about consent, privacy, archives, and responsible reading without encouraging spying or misuse of personal information',
            'The learner and Vega pause before reading more. You ask whether it can ever be right to read letters that were never sent, what protections a library should require, and when the respectful choice is to stop reading or ask permission.',
            true, 30, 40000, 4),
        ('library_vega_unsent_truth', 'library', 'B1', 'Обсудить невысказанную правду',
            'Bibliotecaria Vega', 'vega_librarian', 'library_vega_ethics',
            'a careful librarian who supports nuanced but accessible Spanish reflection about truths people could not say, regrets, kindness, and the limits of what readers can know',
            'At closing time, Vega shows the learner a fully anonymized final fragment from the archive. You invite them to discuss a truth the writer did not dare to say, why someone might leave such words unsent, and how a reader can respond with empathy rather than judgment.',
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
        -- library_vega_card
        ('library_vega_card', 'greet_librarian',      0, true,  'Поздороваться с библиотекарем',
            'The learner greets Bibliotecaria Vega appropriately for a library setting.'),
        ('library_vega_card', 'state_purpose',        1, true,  'Сказать, зачем пришёл в библиотеку',
            'The learner explains why they came to the library or what they want to use there.'),
        ('library_vega_card', 'provide_basic_info',   2, true,  'Дать основные данные для карточки',
            'The learner provides ordinary registration information that would be appropriate for a library card.'),
        ('library_vega_card', 'ask_library_rule',     3, true,  'Спросить о правилах библиотеки',
            'The learner asks a relevant question about library rules, access, borrowing, or archive use.'),
        ('library_vega_card', 'confirm_card_ready',   4, false, 'Подтвердить получение карточки',
            'The learner confirms they understand the next step and that the reader card is ready or being prepared.'),

        -- library_vega_find_letter
        ('library_vega_find_letter', 'ask_archive',       0, true,  'Спросить об архиве писем',
            'The learner asks about the archive or how to search within it.'),
        ('library_vega_find_letter', 'describe_topic',    1, true,  'Описать тему письма',
            'The learner describes the kind of letter they are looking for by subject, situation, or catalogue theme.'),
        ('library_vega_find_letter', 'describe_context',  2, true,  'Уточнить контекст поиска',
            'The learner adds relevant non-identifying context such as approximate time, mood, relationship type, or reason for searching.'),
        ('library_vega_find_letter', 'respect_limits',    3, true,  'Принять ограничения приватности',
            'The learner acknowledges that private identifying details should not be requested or exposed.'),
        ('library_vega_find_letter', 'choose_catalogue',  4, false, 'Выбрать подходящую карточку каталога',
            'The learner chooses or confirms an anonymized catalogue entry to examine next.'),

        -- library_vega_summarize_letter
        ('library_vega_summarize_letter', 'identify_main_situation', 0, true,  'Определить основную ситуацию',
            'The learner identifies the main situation or conflict in the anonymized letter excerpt.'),
        ('library_vega_summarize_letter', 'summarize_message',       1, true,  'Кратко пересказать письмо',
            'The learner summarizes the letter excerpt in their own words.'),
        ('library_vega_summarize_letter', 'name_feeling',            2, true,  'Назвать чувство автора',
            'The learner describes the writer''s likely feeling or emotional state based on the excerpt.'),
        ('library_vega_summarize_letter', 'avoid_private_details',   3, true,  'Не раскрывать личные детали',
            'The learner keeps the summary respectful and avoids names, addresses, or other identifying private information.'),
        ('library_vega_summarize_letter', 'ask_clarification',       4, false, 'Задать уточняющий вопрос',
            'The learner asks a relevant clarification question about meaning, context, or what can be discussed ethically.'),

        -- library_vega_infer_author
        ('library_vega_infer_author', 'state_hypothesis', 0, true,  'Высказать осторожную гипотезу',
            'The learner makes a cautious, fictional hypothesis about the anonymous writer or their situation.'),
        ('library_vega_infer_author', 'cite_clue',        1, true,  'Назвать подсказку',
            'The learner supports the hypothesis with a detail from the anonymized excerpt or catalogue description.'),
        ('library_vega_infer_author', 'mark_uncertainty', 2, true,  'Показать неуверенность',
            'The learner makes clear what is uncertain or cannot be known from the available information.'),
        ('library_vega_infer_author', 'offer_alternative',3, true,  'Предложить другое объяснение',
            'The learner offers another possible interpretation without claiming private certainty.'),
        ('library_vega_infer_author', 'ask_vega_view',    4, false, 'Спросить мнение Веги',
            'The learner asks Vega for a professional or ethical view of the interpretation.'),

        -- library_vega_ethics
        ('library_vega_ethics', 'state_position',     0, true,  'Сформулировать позицию',
            'The learner states an opinion about whether reading unsent letters can be acceptable under library rules.'),
        ('library_vega_ethics', 'give_reason',        1, true,  'Объяснить причину',
            'The learner gives at least one reason connected to privacy, consent, history, learning, or care.'),
        ('library_vega_ethics', 'consider_other_side',2, true,  'Учесть другую сторону',
            'The learner considers an opposing or complicating argument respectfully.'),
        ('library_vega_ethics', 'set_boundary',       3, true,  'Предложить границу',
            'The learner proposes a responsible boundary, protection, or rule for reading archived private material.'),
        ('library_vega_ethics', 'agree_respectful_action',4, false, 'Согласовать уважительное действие',
            'The learner agrees on a respectful next action, including stopping, anonymizing, asking permission, or following archive rules.'),

        -- library_vega_unsent_truth
        ('library_vega_unsent_truth', 'describe_unsaid_truth', 0, true,  'Описать невысказанную правду',
            'The learner describes the truth, feeling, or message the anonymous writer could not say directly.'),
        ('library_vega_unsent_truth', 'explain_possible_reason',1, true, 'Объяснить возможную причину молчания',
            'The learner explains why someone might leave such words unsent.'),
        ('library_vega_unsent_truth', 'respond_with_empathy',  2, true,  'Ответить с эмпатией',
            'The learner responds to the writer''s situation with empathy rather than judgment or curiosity about private identity.'),
        ('library_vega_unsent_truth', 'reflect_on_reader_role',3, true,  'Осмыслить роль читателя',
            'The learner reflects on what a reader can learn or should avoid doing when meeting another person''s unsent truth.'),
        ('library_vega_unsent_truth', 'close_archive',         4, false, 'Завершить разговор об архиве',
            'The learner closes the conversation respectfully and confirms how they will treat the archive material.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'es_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
