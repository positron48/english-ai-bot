-- Seed English-only NPC quest chain: Ada Quill, a B1 librarian guarding the quiet shelf.
-- NPC speaks English; titles are RU, persona/scene/criteria are model-facing EN instructions.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('library_get_card',              'B1', 'Оформить читательский билет'),
        ('library_find_book',             'B1', 'Описать нужную книгу'),
        ('library_summarize_chapter',     'B1', 'Пересказать главу'),
        ('library_argue_interpretation',  'B1', 'Обсудить смысл текста'),
        ('library_unreliable_narrator',   'B1', 'Понять ненадёжного рассказчика'),
        ('library_book_that_reads_you',   'B1', 'Обсудить книгу, которая «выбирает» читателя')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Ada Quill scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('library_get_card', 'library', 'B1', 'Оформить читательский билет',
            'Ada Quill', 'ada_librarian', '',
            'a calm, precise city librarian who speaks clear English, protects the quiet shelf where books appear on their own, and guides B1 learners through respectful library conversations',
            'The learner visits the library for the first time. You welcome them at the front desk, mention in passing that you also watch over a quiet shelf where books sometimes appear without being ordered, and help them apply for a reader card.',
            true, 30, 40000, 0),
        ('library_find_book', 'library', 'B1', 'Описать нужную книгу',
            'Ada Quill', 'ada_librarian', 'library_get_card',
            'a careful librarian who helps readers describe what they are looking for in English while keeping the quiet shelf mysterious and handled only through library rules',
            'The learner now has a reader card and asks about finding a book. You explain that some titles on the quiet shelf cannot be searched by author alone, then ask them to describe the kind of book they need by topic, mood, setting, or any details they remember.',
            true, 30, 40000, 1),
        ('library_summarize_chapter', 'library', 'B1', 'Пересказать главу',
            'Ada Quill', 'ada_librarian', 'library_find_book',
            'a careful librarian who invites concise chapter summaries in English and reminds the learner to focus on plot and meaning rather than memorizing every line',
            'You bring the learner a book from the quiet shelf and open it to an early chapter. The learner must summarize the main events, the central conflict, and the mood of the chapter in their own words.',
            true, 30, 40000, 2),
        ('library_argue_interpretation', 'library', 'B1', 'Обсудить смысл текста',
            'Ada Quill', 'ada_librarian', 'library_summarize_chapter',
            'a thoughtful librarian who encourages reasoned interpretation in English, asks for evidence from the text, and welcomes different readings without forcing one correct answer',
            'After the summary, you ask the learner what they think the chapter really means: what theme, symbol, or message might be hidden beneath the surface events. Invite them to support their view with details from the text and respond to a gentle challenge.',
            true, 30, 40000, 3),
        ('library_unreliable_narrator', 'library', 'B1', 'Понять ненадёжного рассказчика',
            'Ada Quill', 'ada_librarian', 'library_argue_interpretation',
            'a careful librarian who guides B1 learners to notice when a narrator may be misleading, biased, or incomplete, using clear English and examples from the quiet-shelf book',
            'You point out that the book''s narrator may not be telling the whole truth. Ask the learner to identify moments where the narrator seems unreliable, what clues suggest doubt, and how that changes the reader''s trust in the story.',
            true, 30, 40000, 4),
        ('library_book_that_reads_you', 'library', 'B1', 'Обсудить книгу, которая «выбирает» читателя',
            'Ada Quill', 'ada_librarian', 'library_unreliable_narrator',
            'a reflective librarian who supports nuanced but accessible English discussion about books, readers, coincidence, and the idea that some stories seem to find the person who needs them',
            'At closing time, Ada walks the learner past the quiet shelf. You invite them to discuss whether a book can somehow choose its reader, what it felt like when this title appeared for them, and how reading changes the person who opens it.',
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
        -- library_get_card
        ('library_get_card', 'greet_librarian',    0, true,  'Поздороваться с библиотекарем',
            'The learner greets Ada Quill appropriately for a library setting.'),
        ('library_get_card', 'state_purpose',      1, true,  'Сказать, зачем пришёл в библиотеку',
            'The learner explains why they came to the library or what they want to use there.'),
        ('library_get_card', 'provide_basic_info', 2, true,  'Дать основные данные для билета',
            'The learner provides ordinary registration information that would be appropriate for a library card.'),
        ('library_get_card', 'ask_library_rule',   3, true,  'Спросить о правилах библиотеки',
            'The learner asks a relevant question about library rules, access, borrowing, or the quiet shelf.'),
        ('library_get_card', 'confirm_card_ready', 4, false, 'Подтвердить получение билета',
            'The learner confirms they understand the next step and that the reader card is ready or being prepared.'),

        -- library_find_book
        ('library_find_book', 'ask_catalogue',      0, true,  'Спросить, как искать книгу',
            'The learner asks how to search for a book or how the quiet shelf works.'),
        ('library_find_book', 'describe_topic',     1, true,  'Описать тему книги',
            'The learner describes the kind of book they need by subject, genre, or theme.'),
        ('library_find_book', 'describe_mood',      2, true,  'Описать настроение или сеттинг',
            'The learner adds relevant details such as mood, setting, time period, or the kind of story they want.'),
        ('library_find_book', 'add_recall_detail',  3, true,  'Добавить деталь из памяти',
            'The learner mentions at least one remembered detail such as a character type, cover colour, or phrase they associate with the book.'),
        ('library_find_book', 'choose_title',       4, false, 'Выбрать подходящую книгу',
            'The learner chooses or confirms a book from Ada''s suggestions to read next.'),

        -- library_summarize_chapter
        ('library_summarize_chapter', 'identify_main_event', 0, true,  'Назвать главное событие',
            'The learner identifies the main event or turning point in the chapter.'),
        ('library_summarize_chapter', 'summarize_plot',      1, true,  'Кратко пересказать сюжет',
            'The learner summarizes the chapter''s plot in their own words.'),
        ('library_summarize_chapter', 'name_conflict',       2, true,  'Описать конфликт или проблему',
            'The learner describes the central conflict, problem, or tension in the chapter.'),
        ('library_summarize_chapter', 'describe_mood',       3, true,  'Описать настроение главы',
            'The learner describes the mood or atmosphere of the chapter.'),
        ('library_summarize_chapter', 'ask_clarification',   4, false, 'Задать уточняющий вопрос',
            'The learner asks a relevant clarification question about a character, event, or unclear moment.'),

        -- library_argue_interpretation
        ('library_argue_interpretation', 'state_interpretation', 0, true,  'Высказать своё толкование',
            'The learner states what they think the chapter means beyond the surface plot.'),
        ('library_argue_interpretation', 'cite_text_evidence',   1, true,  'Привести пример из текста',
            'The learner supports their interpretation with a detail, line, or moment from the chapter.'),
        ('library_argue_interpretation', 'name_theme',           2, true,  'Назвать тему или символ',
            'The learner identifies a possible theme, symbol, or deeper message in the text.'),
        ('library_argue_interpretation', 'respond_to_challenge', 3, true,  'Ответить на возражение',
            'The learner responds to Ada''s alternative reading or gentle challenge without abandoning their view entirely.'),
        ('library_argue_interpretation', 'ask_ada_view',         4, false, 'Спросить мнение Ады',
            'The learner asks Ada for her professional or personal view of the interpretation.'),

        -- library_unreliable_narrator
        ('library_unreliable_narrator', 'identify_doubt',       0, true,  'Заметить сомнительный момент',
            'The learner points to a moment where the narrator may be unreliable, biased, or incomplete.'),
        ('library_unreliable_narrator', 'cite_clue',            1, true,  'Назвать подсказку недоверия',
            'The learner names a clue from the text that suggests the narrator should not be fully trusted.'),
        ('library_unreliable_narrator', 'explain_effect',       2, true,  'Объяснить эффект на читателя',
            'The learner explains how the unreliable narration changes the reader''s understanding or trust.'),
        ('library_unreliable_narrator', 'offer_alternative',    3, true,  'Предложить другую версию',
            'The learner offers another possible version of events that the narrator may be hiding or distorting.'),
        ('library_unreliable_narrator', 'reflect_on_trust',     4, false, 'Осмыслить доверие к рассказчику',
            'The learner reflects on when a reader should trust or question a narrator.'),

        -- library_book_that_reads_you
        ('library_book_that_reads_you', 'describe_connection',  0, true,  'Описать связь с книгой',
            'The learner describes why this book feels meaningful, timely, or personally relevant to them.'),
        ('library_book_that_reads_you', 'discuss_coincidence',  1, true,  'Обсудить совпадение или «выбор»',
            'The learner discusses whether the book seemed to find them or whether that is only coincidence.'),
        ('library_book_that_reads_you', 'reflect_on_change',    2, true,  'Осмыслить, как чтение меняет',
            'The learner reflects on how reading the book may change their view, mood, or understanding.'),
        ('library_book_that_reads_you', 'compare_reader_role',  3, true,  'Сравнить роль читателя и книги',
            'The learner compares the reader''s role with the book''s role in creating meaning.'),
        ('library_book_that_reads_you', 'close_conversation',   4, false, 'Завершить разговор у тихой полки',
            'The learner closes the conversation respectfully and confirms what they will take away from the quiet shelf.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'en_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
