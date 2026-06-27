-- Seed Spanish course NPC Lucía Canta, a plaza singer quest chain.
-- Applies only to es_ru; Lucía speaks Spanish and all scenarios are pinned to A1.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('singer_lucia_greeting',      'A1', 'Поздороваться с Лусией'),
        ('singer_lucia_request',       'A1', 'Попросить песню'),
        ('singer_lucia_describe_song', 'A1', 'Описать настроение песни'),
        ('singer_lucia_missing_word',  'A1', 'Предложить слово для строки'),
        ('singer_lucia_memory',        'A1', 'Рассказать воспоминание для песни'),
        ('singer_lucia_new_song',      'A1', 'Обсудить готовую песню')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Lucía Canta plaza singer quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('singer_lucia_greeting', 'plaza', 'A1', 'Поздороваться с Лусией',
            'Lucía Canta', 'lucia_singer', '',
            'a warm street singer in the plaza who speaks clear, encouraging A1 Spanish and invites the learner to answer with short, natural sentences',
            'The learner hears music in the plaza and stops near Lucía. You are finishing a simple song, smiling at passers-by, and you invite the learner to greet you and react to the music.',
            true, 30, 40000, 0),
        ('singer_lucia_request', 'plaza', 'A1', 'Попросить песню',
            'Lucía Canta', 'lucia_singer', 'singer_lucia_greeting',
            'a friendly plaza singer who speaks simple Spanish, accepts song requests kindly, and helps the learner ask for music politely',
            'The learner returns to the plaza while Lucía is choosing the next song. You ask what kind of song they want to hear and offer easy choices about topic or mood.',
            true, 30, 40000, 1),
        ('singer_lucia_describe_song', 'plaza', 'A1', 'Описать настроение песни',
            'Lucía Canta', 'lucia_singer', 'singer_lucia_request',
            'a patient street singer who speaks clear A1 Spanish and helps the learner describe music with simple feelings, pace, and images',
            'Lucía has played the requested song. You ask the learner what feeling the song has and what simple image or moment it makes them imagine.',
            true, 30, 40000, 2),
        ('singer_lucia_missing_word', 'plaza', 'A1', 'Предложить слово для строки',
            'Lucía Canta', 'lucia_singer', 'singer_lucia_describe_song',
            'a creative plaza singer who speaks accessible Spanish and asks for simple, safe word ideas for a new line about the city',
            'Lucía is writing a short line for a city song but one word is missing. You explain the idea in simple language and ask the learner to suggest a word or small idea that could fit.',
            true, 30, 40000, 3),
        ('singer_lucia_memory', 'plaza', 'A1', 'Рассказать воспоминание для песни',
            'Lucía Canta', 'lucia_singer', 'singer_lucia_missing_word',
            'a gentle street singer who speaks slow, supportive Spanish and turns small personal memories into simple song ideas without pushing for private details',
            'Lucía wants the city song to include a real memory from someone in the plaza. You ask the learner to share a simple memory about a place, person, sound, or feeling.',
            true, 30, 40000, 4),
        ('singer_lucia_new_song', 'plaza', 'A1', 'Обсудить готовую песню',
            'Lucía Canta', 'lucia_singer', 'singer_lucia_memory',
            'a warm plaza singer who speaks clear Spanish, celebrates the learner''s help, and discusses a finished song in simple, encouraging language',
            'Lucía has finished a small song using the learner''s ideas. You play it in the plaza, thank the learner, and ask what they think about the song and where it should be sung next.',
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
        -- singer_lucia_greeting
        ('singer_lucia_greeting', 'greet_lucia',     0, true,  'Поздороваться с Лусией',
            'The learner greets Lucía politely and starts the plaza conversation.'),
        ('singer_lucia_greeting', 'notice_music',    1, true,  'Сказать, что услышал музыку',
            'The learner shows that they noticed the music or stopped because of the song.'),
        ('singer_lucia_greeting', 'praise_song',     2, true,  'Похвалить песню',
            'The learner gives a positive reaction to the song, music, voice, or performance.'),
        ('singer_lucia_greeting', 'ask_name_or_role',3, true,  'Спросить о Лусии',
            'The learner asks Lucía who she is, what she is doing, or whether she sings here often.'),
        ('singer_lucia_greeting', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks Lucía or closes the first meeting politely.'),

        -- singer_lucia_request
        ('singer_lucia_request', 'greet_again',      0, true,  'Снова поздороваться',
            'The learner greets Lucía again and reopens the conversation.'),
        ('singer_lucia_request', 'ask_for_song',     1, true,  'Попросить песню',
            'The learner asks Lucía to sing or play a song.'),
        ('singer_lucia_request', 'choose_song_type', 2, true,  'Выбрать тип песни',
            'The learner chooses or describes the kind of song they want to hear.'),
        ('singer_lucia_request', 'answer_followup',  3, true,  'Ответить на уточнение',
            'The learner answers a simple follow-up question about topic, mood, person, or occasion.'),
        ('singer_lucia_request', 'react_to_start',   4, false, 'Отреагировать на начало песни',
            'The learner reacts when Lucía accepts the request or begins the song.'),

        -- singer_lucia_describe_song
        ('singer_lucia_describe_song', 'say_listened',     0, true,  'Сказать, что слушал песню',
            'The learner confirms they listened to the song or paid attention to it.'),
        ('singer_lucia_describe_song', 'name_mood',        1, true,  'Назвать настроение',
            'The learner describes the mood or feeling of the song.'),
        ('singer_lucia_describe_song', 'give_reason',      2, true,  'Объяснить впечатление',
            'The learner gives a simple reason for their impression of the song.'),
        ('singer_lucia_describe_song', 'describe_image',   3, true,  'Описать образ',
            'The learner describes a simple image, place, colour, or moment the song suggests.'),
        ('singer_lucia_describe_song', 'ask_lucia_opinion',4, false, 'Спросить мнение Лусии',
            'The learner asks Lucía what feeling or image she wanted the song to have.'),

        -- singer_lucia_missing_word
        ('singer_lucia_missing_word', 'ask_about_line',  0, true,  'Спросить о строке',
            'The learner asks about the unfinished line or what kind of word Lucía needs.'),
        ('singer_lucia_missing_word', 'understand_theme',1, true,  'Понять тему строки',
            'The learner shows understanding of the line theme, mood, or city image.'),
        ('singer_lucia_missing_word', 'suggest_word',    2, true,  'Предложить слово',
            'The learner suggests one word, image, or small idea that could complete the line.'),
        ('singer_lucia_missing_word', 'explain_choice',  3, true,  'Объяснить выбор',
            'The learner gives a simple reason why the suggested word or idea fits.'),
        ('singer_lucia_missing_word', 'accept_edit',     4, false, 'Согласиться с правкой',
            'The learner reacts to Lucía changing, accepting, or adapting the suggestion.'),

        -- singer_lucia_memory
        ('singer_lucia_memory', 'ask_memory_prompt',0, true,  'Уточнить, какое воспоминание нужно',
            'The learner asks what kind of memory Lucía wants for the song.'),
        ('singer_lucia_memory', 'choose_memory',    1, true,  'Выбрать воспоминание',
            'The learner identifies a personal memory, moment, place, person, or sound they can share.'),
        ('singer_lucia_memory', 'tell_memory',      2, true,  'Рассказать воспоминание',
            'The learner shares the memory with enough detail for Lucía to understand it.'),
        ('singer_lucia_memory', 'say_feeling',      3, true,  'Назвать чувство',
            'The learner says what feeling is connected to the memory.'),
        ('singer_lucia_memory', 'allow_song_use',   4, false, 'Разрешить использовать идею',
            'The learner agrees that Lucía can use the memory, feeling, or image in the song.'),

        -- singer_lucia_new_song
        ('singer_lucia_new_song', 'greet_final',      0, true,  'Поздороваться перед выступлением',
            'The learner greets Lucía before or after hearing the finished song.'),
        ('singer_lucia_new_song', 'listen_or_ask',    1, true,  'Попросить услышать готовую песню',
            'The learner asks to hear the finished song or shows they are ready to listen.'),
        ('singer_lucia_new_song', 'react_to_song',    2, true,  'Отреагировать на песню',
            'The learner gives an opinion or emotional reaction to the finished song.'),
        ('singer_lucia_new_song', 'mention_own_part', 3, true,  'Упомянуть свой вклад',
            'The learner recognizes how their earlier word, idea, or memory appears in the song.'),
        ('singer_lucia_new_song', 'discuss_next_place',4, false, 'Обсудить, где петь дальше',
            'The learner discusses where, when, or for whom Lucía could sing the song next.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'es_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
