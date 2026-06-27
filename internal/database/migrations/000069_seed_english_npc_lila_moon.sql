-- Seed English course NPC Lila Moon, a town-square street musician quest chain.
-- Applies only to en_ru; Lila speaks English and all scenarios are pinned to A1.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('musician_give_coin',      'A1', 'Похвалить музыку и дать монету'),
        ('musician_request_song',   'A1', 'Попросить песню'),
        ('musician_describe_mood',  'A1', 'Описать настроение песни'),
        ('musician_lost_melody',    'A1', 'Помочь вспомнить часть мелодии словами'),
        ('musician_family_story',   'A1', 'Обсудить семейную историю'),
        ('musician_square_concert', 'A1', 'Помочь подготовить маленький концерт')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Lila Moon town-square musician quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('musician_give_coin', 'town_square', 'A1', 'Похвалить музыку и дать монету',
            'Lila Moon', 'lila_musician', '',
            'a warm street musician in the town square who speaks clear, encouraging A1 English and invites the learner to answer with short, natural sentences',
            'The learner hears music in the town square and stops near Lila. You are finishing a simple song, smiling at passers-by, and you invite the learner to greet you, praise the music, and offer a coin.',
            true, 30, 40000, 0),
        ('musician_request_song', 'town_square', 'A1', 'Попросить песню',
            'Lila Moon', 'lila_musician', 'musician_give_coin',
            'a friendly town-square musician who speaks simple English, accepts song requests kindly, and helps the learner ask for music politely',
            'The learner returns to the square while Lila is choosing the next song. You ask what kind of song they want to hear and offer easy choices about topic or mood.',
            true, 30, 40000, 1),
        ('musician_describe_mood', 'town_square', 'A1', 'Описать настроение песни',
            'Lila Moon', 'lila_musician', 'musician_request_song',
            'a patient street musician who speaks clear A1 English and helps the learner describe music with simple feelings, pace, and images',
            'Lila has played the requested song. You ask the learner what feeling the song has and what simple image or moment it makes them imagine.',
            true, 30, 40000, 2),
        ('musician_lost_melody', 'town_square', 'A1', 'Помочь вспомнить часть мелодии словами',
            'Lila Moon', 'lila_musician', 'musician_describe_mood',
            'a creative town-square musician who speaks accessible English and asks the learner to help recover a forgotten melody line using simple words or sounds',
            'Lila is trying to finish a melody but one part is missing from memory. You hum what you remember, explain the mood in simple language, and ask the learner to suggest words or a short phrase that could bring the melody back.',
            true, 30, 40000, 3),
        ('musician_family_story', 'town_square', 'A1', 'Обсудить семейную историю',
            'Lila Moon', 'lila_musician', 'musician_lost_melody',
            'a gentle street musician who speaks slow, supportive English and shares a simple family story connected to her music without pushing for private details from the learner',
            'Lila wants the recovered melody to mean something personal. You tell a short family story behind one of your songs and invite the learner to ask questions and react with simple empathy.',
            true, 30, 40000, 4),
        ('musician_square_concert', 'town_square', 'A1', 'Помочь подготовить маленький концерт',
            'Lila Moon', 'lila_musician', 'musician_family_story',
            'a warm town-square musician who speaks clear English, celebrates the learner''s help, and plans a small concert together in simple, encouraging language',
            'Lila is preparing a short concert in the square using the melody and story the learner helped with. You ask for ideas about time, audience, song order, or a simple announcement and thank the learner for their support.',
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
        -- musician_give_coin
        ('musician_give_coin', 'greet_lila',       0, true,  'Поздороваться с Лилой',
            'The learner greets Lila politely and starts the town-square conversation.'),
        ('musician_give_coin', 'notice_music',     1, true,  'Сказать, что услышал музыку',
            'The learner shows that they noticed the music or stopped because of the song.'),
        ('musician_give_coin', 'praise_music',     2, true,  'Похвалить музыку',
            'The learner gives a positive reaction to the song, music, voice, or performance.'),
        ('musician_give_coin', 'give_coin',        3, true,  'Дать монету',
            'The learner offers a coin, tip, or other simple support for the street performance.'),
        ('musician_give_coin', 'thank',            4, false, 'Поблагодарить',
            'The learner thanks Lila or closes the first meeting politely.'),

        -- musician_request_song
        ('musician_request_song', 'greet_again',      0, true,  'Снова поздороваться',
            'The learner greets Lila again and reopens the conversation.'),
        ('musician_request_song', 'ask_for_song',     1, true,  'Попросить песню',
            'The learner asks Lila to sing or play a song.'),
        ('musician_request_song', 'choose_song_type', 2, true,  'Выбрать тип песни',
            'The learner chooses or describes the kind of song they want to hear.'),
        ('musician_request_song', 'answer_followup',  3, true,  'Ответить на уточнение',
            'The learner answers a simple follow-up question about topic, mood, person, or occasion.'),
        ('musician_request_song', 'react_to_start',   4, false, 'Отреагировать на начало песни',
            'The learner reacts when Lila accepts the request or begins the song.'),

        -- musician_describe_mood
        ('musician_describe_mood', 'say_listened',      0, true,  'Сказать, что слушал песню',
            'The learner confirms they listened to the song or paid attention to it.'),
        ('musician_describe_mood', 'name_mood',         1, true,  'Назвать настроение',
            'The learner describes the mood or feeling of the song.'),
        ('musician_describe_mood', 'give_reason',       2, true,  'Объяснить впечатление',
            'The learner gives a simple reason for their impression of the song.'),
        ('musician_describe_mood', 'describe_image',    3, true,  'Описать образ',
            'The learner describes a simple image, place, colour, or moment the song suggests.'),
        ('musician_describe_mood', 'ask_lila_opinion',  4, false, 'Спросить мнение Лилы',
            'The learner asks Lila what feeling or image she wanted the song to have.'),

        -- musician_lost_melody
        ('musician_lost_melody', 'ask_about_melody',  0, true,  'Спросить о забытой мелодии',
            'The learner asks about the missing melody part or what Lila is trying to remember.'),
        ('musician_lost_melody', 'understand_problem',1, true,  'Понять, что забыто',
            'The learner shows understanding of which part of the melody, mood, or line is missing.'),
        ('musician_lost_melody', 'suggest_words',     2, true,  'Предложить слова',
            'The learner suggests words, sounds, or a short phrase that could help recover the melody.'),
        ('musician_lost_melody', 'explain_choice',    3, true,  'Объяснить выбор',
            'The learner gives a simple reason why the suggested words or sounds fit the melody.'),
        ('musician_lost_melody', 'confirm_help',      4, false, 'Подтвердить, что идея подходит',
            'The learner reacts when Lila tries the suggestion and says whether it helps bring the melody back.'),

        -- musician_family_story
        ('musician_family_story', 'ask_about_story',   0, true,  'Спросить о семейной истории',
            'The learner asks about the family story behind Lila''s song or music.'),
        ('musician_family_story', 'understand_story',  1, true,  'Понять историю',
            'The learner shows understanding of the people, place, or event in Lila''s family story.'),
        ('musician_family_story', 'ask_followup',      2, true,  'Задать уточняющий вопрос',
            'The learner asks a simple follow-up question about the family story or its connection to the song.'),
        ('musician_family_story', 'share_reaction',    3, true,  'Отреагировать на историю',
            'The learner reacts with empathy, interest, or a simple opinion about the family story.'),
        ('musician_family_story', 'connect_to_song',   4, false, 'Связать историю с песней',
            'The learner connects the family story to the melody, mood, or meaning of the song.'),

        -- musician_square_concert
        ('musician_square_concert', 'greet_concert',    0, true,  'Поздороваться перед концертом',
            'The learner greets Lila while she is preparing the small square concert.'),
        ('musician_square_concert', 'offer_help',       1, true,  'Предложить помощь',
            'The learner offers to help with the concert preparation.'),
        ('musician_square_concert', 'suggest_idea',     2, true,  'Предложить идею',
            'The learner suggests an idea for timing, audience, song order, or a simple announcement.'),
        ('musician_square_concert', 'discuss_plan',     3, true,  'Обсудить план',
            'The learner discusses or agrees on a simple plan for the small concert.'),
        ('musician_square_concert', 'wish_luck',        4, false, 'Пожелать удачи',
            'The learner wishes Lila good luck or encourages her before the concert.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'en_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
