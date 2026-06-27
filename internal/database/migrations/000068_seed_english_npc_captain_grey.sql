-- Seed English course NPC Captain Grey, a bus driver for A2 conversation quests.
-- Applies only to en_ru; Captain Grey speaks English and all scenarios are pinned to A2.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('bus_buy_ticket',       'A2', 'Купить билет и назвать остановку'),
        ('bus_missed_stop',      'A2', 'Объяснить, что проехал остановку'),
        ('bus_ask_route',        'A2', 'Спросить, как пересесть'),
        ('bus_talk_passenger',   'A2', 'Пересказать историю пассажира'),
        ('bus_last_route',       'A2', 'Обсудить последний рейс и странного пассажира'),
        ('bus_city_map_secret',  'A2', 'Понять, почему маршрут изменился')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'en_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Captain Grey bus driver quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('bus_buy_ticket', 'bus', 'A2', 'Купить билет и назвать остановку',
            'Captain Grey', 'grey_bus_driver', '',
            'a calm, experienced city bus driver who speaks clear A2 English, knows every stop on his route, and patiently helps passengers buy tickets and find the right stop',
            'The learner boards Captain Grey''s bus for the first time. You greet them from the driver''s seat, ask where they want to go, and help them buy the right ticket.',
            true, 30, 40000, 0),
        ('bus_missed_stop', 'bus', 'A2', 'Объяснить, что проехал остановку',
            'Captain Grey', 'grey_bus_driver', 'bus_buy_ticket',
            'a patient city bus driver who listens without blame, speaks clear A2 English, and gives practical advice when a passenger misses their stop',
            'The learner realizes they missed their stop. You listen calmly, ask where they wanted to get off, and explain what they can do next.',
            true, 30, 40000, 1),
        ('bus_ask_route', 'bus', 'A2', 'Спросить, как пересесть',
            'Captain Grey', 'grey_bus_driver', 'bus_missed_stop',
            'a helpful city bus driver who knows transfer points and connections, speaks clear A2 English, and explains route changes in simple steps',
            'The learner needs to change buses to reach their destination. You help them understand where to get off, which bus to take next, and how to find the right stop.',
            true, 30, 40000, 2),
        ('bus_talk_passenger', 'bus', 'A2', 'Пересказать историю пассажира',
            'Captain Grey', 'grey_bus_driver', 'bus_ask_route',
            'a talkative bus driver who remembers small stories from the city, speaks clear A2 English, and invites the learner to retell events in simple order',
            'The bus is quieter now. You tell the learner about a passenger from an earlier stop and ask them to repeat the main story to check they understood.',
            true, 30, 40000, 3),
        ('bus_last_route', 'bus', 'A2', 'Обсудить последний рейс и странного пассажира',
            'Captain Grey', 'grey_bus_driver', 'bus_talk_passenger',
            'a thoughtful bus driver who knows the last route of the day well, speaks clear A2 English, and shares simple observations about unusual passengers',
            'It is late in the day and the bus is nearly empty. Captain Grey mentions the last route and a strange passenger from earlier. You invite the learner to ask questions and share their thoughts.',
            true, 30, 40000, 4),
        ('bus_city_map_secret', 'bus', 'A2', 'Понять, почему маршрут изменился',
            'Captain Grey', 'grey_bus_driver', 'bus_last_route',
            'a nostalgic city bus driver who remembers old routes and city changes, explains why the map looks different now, and keeps the conversation grounded in everyday A2 English',
            'The bus passes a street Captain Grey used to drive through on an older route. You tell the learner the route changed recently and invite them to ask why and what is different now.',
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
        -- bus_buy_ticket
        ('bus_buy_ticket', 'greet',           0, true,  'Поздороваться с капитаном Греем',
            'The learner greets Captain Grey and starts the bus conversation politely.'),
        ('bus_buy_ticket', 'ask_ticket',      1, true,  'Попросить билет',
            'The learner asks to buy or get a bus ticket.'),
        ('bus_buy_ticket', 'say_destination', 2, true,  'Назвать остановку или район',
            'The learner says where they want to go, naming a stop, area, or destination.'),
        ('bus_buy_ticket', 'confirm_ticket',  3, true,  'Подтвердить билет',
            'The learner confirms the ticket, route, or destination Captain Grey offers.'),
        ('bus_buy_ticket', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks Captain Grey and/or ends the exchange politely.'),

        -- bus_missed_stop
        ('bus_missed_stop', 'greet',              0, true,  'Обратиться к капитану Грею',
            'The learner politely gets Captain Grey''s attention.'),
        ('bus_missed_stop', 'say_missed_stop',    1, true,  'Сказать, что проехал остановку',
            'The learner explains that they missed their stop or went too far.'),
        ('bus_missed_stop', 'say_target_stop',    2, true,  'Назвать нужную остановку',
            'The learner says which stop or place they wanted to get off at.'),
        ('bus_missed_stop', 'ask_next_action',    3, true,  'Спросить, что делать дальше',
            'The learner asks for a practical next step to fix the missed-stop problem.'),
        ('bus_missed_stop', 'accept_instruction', 4, false, 'Принять совет',
            'The learner accepts or confirms Captain Grey''s suggested next action.'),

        -- bus_ask_route
        ('bus_ask_route', 'greet',              0, true,  'Поздороваться с капитаном Греем',
            'The learner greets Captain Grey or politely continues the bus conversation.'),
        ('bus_ask_route', 'explain_need',       1, true,  'Объяснить, куда нужно попасть',
            'The learner explains where they need to go or that they need a different bus.'),
        ('bus_ask_route', 'ask_transfer',       2, true,  'Спросить, как пересесть',
            'The learner asks how to change buses, where to transfer, or which line to take.'),
        ('bus_ask_route', 'confirm_direction',  3, true,  'Подтвердить маршрут пересадки',
            'The learner confirms the transfer stop, next bus, or direction Captain Grey describes.'),
        ('bus_ask_route', 'thank',              4, false, 'Поблагодарить',
            'The learner thanks Captain Grey and/or closes the exchange politely.'),

        -- bus_talk_passenger
        ('bus_talk_passenger', 'greet',           0, true,  'Поздороваться с капитаном Греем',
            'The learner greets Captain Grey and agrees to listen to the passenger story.'),
        ('bus_talk_passenger', 'ask_story',       1, true,  'Попросить рассказать историю',
            'The learner asks Captain Grey about the passenger or invites him to tell the story.'),
        ('bus_talk_passenger', 'retell_events',   2, true,  'Пересказать основные события',
            'The learner retells the main events of the passenger story in a simple order.'),
        ('bus_talk_passenger', 'react_story',     3, true,  'Отреагировать на историю',
            'The learner gives a simple reaction, opinion, or feeling about the story.'),
        ('bus_talk_passenger', 'ask_next_detail', 4, false, 'Спросить дополнительную деталь',
            'The learner asks one follow-up question about the passenger, place, or result.'),

        -- bus_last_route
        ('bus_last_route', 'greet',                 0, true,  'Поздороваться с капитаном Греем',
            'The learner greets Captain Grey and starts the late-route conversation politely.'),
        ('bus_last_route', 'ask_last_route',          1, true,  'Спросить о последнем рейсе',
            'The learner asks about the last route, what happens at the end, or how the day finishes.'),
        ('bus_last_route', 'ask_strange_passenger', 2, true,  'Спросить о странном пассажире',
            'The learner asks Captain Grey about the strange passenger or what was unusual about them.'),
        ('bus_last_route', 'react_details',         3, true,  'Отреагировать на детали',
            'The learner reacts to the details about the last route or the strange passenger.'),
        ('bus_last_route', 'ask_conclusion',        4, false, 'Спросить, чем всё закончилось',
            'The learner asks what happened in the end or shares a simple guess about the passenger.'),

        -- bus_city_map_secret
        ('bus_city_map_secret', 'greet',             0, true,  'Поздороваться с капитаном Греем',
            'The learner greets Captain Grey and continues the bus conversation.'),
        ('bus_city_map_secret', 'ask_route_change',  1, true,  'Спросить, почему маршрут изменился',
            'The learner asks why the route changed or what is different on the map now.'),
        ('bus_city_map_secret', 'ask_old_route',     2, true,  'Узнать о старом маршруте',
            'The learner asks Captain Grey about the old route or what it was like before.'),
        ('bus_city_map_secret', 'compare_routes',    3, true,  'Сравнить старый и новый маршрут',
            'The learner compares the old and current routes using simple differences.'),
        ('bus_city_map_secret', 'say_preference',    4, false, 'Сказать, какой маршрут лучше',
            'The learner says which route they prefer or which route seems better for passengers.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'en_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
