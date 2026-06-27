-- Seed English-only NPC quest chain: Victor North, a retired city protector.
-- Victor speaks English; titles are RU, persona/setup/criteria are model-facing EN.
-- The full B2-C1 story is pinned to B2 so it appears in the first advanced level.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('retired_hero_small_errand',      'B2', 'Помочь с обычным бытовым делом'),
        ('retired_hero_public_attention',  'B2', 'Обсудить нежелательное внимание'),
        ('retired_hero_moral_choice',      'B2', 'Выбрать между двумя плохими решениями'),
        ('retired_hero_cost_of_fame',      'B2', 'Обсудить цену славы'),
        ('retired_hero_apology',           'B2', 'Подготовить сложное извинение'),
        ('retired_hero_new_identity',      'B2', 'Решить, кем он хочет быть теперь')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Victor North retired-hero quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('retired_hero_small_errand', 'community_center', 'B2', 'Помочь с обычным бытовым делом',
            'Victor North', 'victor_old_hero', NULL,
            'a tired English-speaking former city protector who once kept people safe but now lives quietly in plain clothes; he speaks plainly, never wears costumes or emblems, never uses codenames or superhero references, and treats ordinary help as meaningful without dramatizing it',
            'The learner meets Victor North at a community center. He needs help with a simple everyday task—carrying groceries, fixing a notice board, or picking up a prescription—and awkwardly frames it as if it still mattered like old crises. You invite the learner to help practically while staying kind and grounded.',
            true, 30, 40000, 0),
        ('retired_hero_public_attention', 'street', 'B2', 'Обсудить нежелательное внимание',
            'Victor North', 'victor_old_hero', 'retired_hero_small_errand',
            'a weary English-speaking former city protector who dislikes being recognized, photographed, or praised in public; he values dignity, privacy, and calm conversation without hero worship or franchise-style language',
            'On the street, strangers stare, ask for photos, or repeat exaggerated stories about Victor''s past. You help the learner discuss unwanted attention, personal boundaries, and how public memory can outlive the person.',
            true, 30, 40000, 1),
        ('retired_hero_moral_choice', 'community_center', 'B2', 'Выбрать между двумя плохими решениями',
            'Victor North', 'victor_old_hero', 'retired_hero_public_attention',
            'a reflective English-speaking former protector who still feels responsible for others but knows every choice has a cost; he speaks without superhero tropes, emblems, or theatrical hero language',
            'Victor faces a dilemma where both options are bad—perhaps reporting a neighbor''s minor wrongdoing versus staying silent, or stepping into a dispute versus letting it escalate. You guide the learner to explore both sides and help him articulate a reasoned choice.',
            true, 30, 40000, 2),
        ('retired_hero_cost_of_fame', 'street', 'B2', 'Обсудить цену славы',
            'Victor North', 'victor_old_hero', 'retired_hero_moral_choice',
            'an honest English-speaking former city protector who can discuss lost privacy, broken relationships, and exhaustion from being needed; he never names fictional heroes or wears symbols of his old role',
            'Victor reflects on what protecting the city cost him—sleep, friendships, a normal life, and the right to be forgotten. You invite the learner to discuss the price of fame and recognition without glorifying or mocking his past.',
            true, 30, 40000, 3),
        ('retired_hero_apology', 'community_center', 'B2', 'Подготовить сложное извинение',
            'Victor North', 'victor_old_hero', 'retired_hero_cost_of_fame',
            'a humble English-speaking former protector who knows he hurt someone while doing his job or chasing danger; he wants to apologize sincerely without excuses, emblems, or heroic self-justification',
            'Victor needs to apologize to someone he wronged years ago—a friend he neglected, a partner he worried, or a bystander he frightened. You help the learner prepare the apology: what happened, responsibility, tone, and how to deliver it.',
            true, 30, 40000, 4),
        ('retired_hero_new_identity', 'street', 'B2', 'Решить, кем он хочет быть теперь',
            'Victor North', 'victor_old_hero', 'retired_hero_apology',
            'a quieter English-speaking former city protector exploring life beyond the protector role; plain clothes, ordinary name, no emblems, and hope for a human future rather than legend',
            'Victor wonders who he is now that he no longer patrols or rescues. You guide a conversation about ordinary roles—neighbor, volunteer, friend—and close the chain with dignity and a sense of what he wants next.',
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
-- 3. Tasks for the new quest scenarios.
-- ---------------------------------------------------------------------------
WITH task(scenario_code, code, sort_order, is_required, title, completion_criteria) AS (
    VALUES
        -- retired_hero_small_errand
        ('retired_hero_small_errand', 'greet_victor',     0, true,  'Поприветствовать Виктора',
            'The learner greets Victor North respectfully and shows willingness to hear what he needs.'),
        ('retired_hero_small_errand', 'clarify_errand',   1, true,  'Уточнить, что нужно сделать',
            'The learner asks what the everyday task involves and clarifies the practical goal behind any dramatic framing.'),
        ('retired_hero_small_errand', 'agree_help',       2, true,  'Согласиться помочь',
            'The learner agrees to help while keeping the task safe, realistic, and respectful.'),
        ('retired_hero_small_errand', 'report_done',      3, true,  'Сообщить о результате',
            'The learner reports what was completed or what still needs attention.'),
        ('retired_hero_small_errand', 'close_politely',   4, false, 'Вежливо завершить разговор',
            'The learner closes the exchange politely, with thanks or a respectful final comment.'),

        -- retired_hero_public_attention
        ('retired_hero_public_attention', 'notice_attention',    0, true,  'Заметить нежелательное внимание',
            'The learner identifies that Victor is receiving unwanted stares, questions, photos, or public praise.'),
        ('retired_hero_public_attention', 'ask_how_feels',       1, true,  'Спросить, как он себя чувствует',
            'The learner asks how the attention makes Victor feel or invites him to describe the discomfort.'),
        ('retired_hero_public_attention', 'discuss_boundaries',  2, true,  'Обсудить личные границы',
            'The learner discusses boundaries Victor wants, such as privacy, refusal of photos, or not being treated as a legend.'),
        ('retired_hero_public_attention', 'suggest_response',    3, true,  'Предложить, как реагировать',
            'The learner suggests a calm, dignified way Victor could respond to strangers or exaggerated stories.'),
        ('retired_hero_public_attention', 'offer_support',       4, false, 'Предложить поддержку',
            'The learner offers emotional support or a practical way to reduce the unwanted attention.'),

        -- retired_hero_moral_choice
        ('retired_hero_moral_choice', 'understand_dilemma',  0, true,  'Понять суть дилеммы',
            'The learner identifies the core problem Victor faces and why both paths feel difficult.'),
        ('retired_hero_moral_choice', 'name_bad_options',    1, true,  'Назвать оба плохих варианта',
            'The learner names or summarizes both undesirable options without oversimplifying them.'),
        ('retired_hero_moral_choice', 'weigh_consequences',  2, true,  'Взвесить последствия',
            'The learner discusses likely consequences, harms, or tradeoffs of each option.'),
        ('retired_hero_moral_choice', 'help_decide',         3, true,  'Помочь сделать выбор',
            'The learner helps Victor articulate a reasoned choice or a clear next step, even if neither option is ideal.'),
        ('retired_hero_moral_choice', 'reflect_choice',       4, false, 'Осмыслить решение',
            'The learner briefly reflects on what the choice says about responsibility, limits, or values.'),

        -- retired_hero_cost_of_fame
        ('retired_hero_cost_of_fame', 'invite_reflection',     0, true,  'Пригласить к размышлению',
            'The learner gently invites Victor to reflect on what fame or public recognition has cost him.'),
        ('retired_hero_cost_of_fame', 'name_costs',            1, true,  'Назвать цену славы',
            'The learner helps identify specific costs such as privacy, relationships, rest, fear, or lost ordinary life.'),
        ('retired_hero_cost_of_fame', 'compare_public_private',2, true,  'Сравнить публичную и личную жизнь',
            'The learner compares how Victor was seen publicly with how he lived or felt privately.'),
        ('retired_hero_cost_of_fame', 'acknowledge_tradeoffs', 3, true,  'Признать компромиссы',
            'The learner acknowledges tradeoffs without glorifying heroism or dismissing Victor''s sacrifices.'),
        ('retired_hero_cost_of_fame', 'connect_to_present',    4, false, 'Связать с настоящим',
            'The learner connects the past costs to how Victor wants to live now or what he still carries.'),

        -- retired_hero_apology
        ('retired_hero_apology', 'identify_wronged',  0, true,  'Определить, перед кем извиняться',
            'The learner identifies who Victor needs to apologize to and why that person was hurt.'),
        ('retired_hero_apology', 'clarify_events',    1, true,  'Уточнить, что произошло',
            'The learner helps clarify what happened, Victor''s role in it, and what was left unsaid or unresolved.'),
        ('retired_hero_apology', 'draft_apology',     2, true,  'Составить извинение',
            'The learner helps draft key elements of the apology: acknowledgment, responsibility, and what Victor regrets.'),
        ('retired_hero_apology', 'refine_tone',       3, true,  'Отточить тон извинения',
            'The learner refines the tone so the apology sounds sincere, specific, and free of excuses or hero language.'),
        ('retired_hero_apology', 'plan_delivery',     4, false, 'Спланировать, как извиниться',
            'The learner discusses when, where, or how Victor could deliver the apology respectfully.'),

        -- retired_hero_new_identity
        ('retired_hero_new_identity', 'ask_future_self',       0, true,  'Спросить о будущем',
            'The learner asks who Victor wants to be now or what kind of life he hopes for beyond his old role.'),
        ('retired_hero_new_identity', 'name_values',           1, true,  'Назвать личные ценности',
            'The learner identifies values Victor still holds, such as care, honesty, service, or rest.'),
        ('retired_hero_new_identity', 'explore_ordinary_roles',2, true,  'Обсудить обычные роли',
            'The learner explores ordinary roles Victor could embrace, such as neighbor, volunteer, mentor, or friend.'),
        ('retired_hero_new_identity', 'summarize_identity',      3, true,  'Подвести итог новой идентичности',
            'The learner summarizes a more human version of Victor''s identity that is not defined by past protection or fame.'),
        ('retired_hero_new_identity', 'close_chain',             4, false, 'Завершить историю',
            'The learner closes the story with gratitude, reflection, or a dignified final exchange.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN courses c ON c.code = 'en_ru'
JOIN conversation_scenarios cs ON cs.course_id = c.id AND cs.code = t.scenario_code
ON CONFLICT (scenario_id, code) DO NOTHING;
