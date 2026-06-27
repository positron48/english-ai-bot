-- Seed Spanish course NPC Rafa, a bus driver for A2 conversation quests.
-- Applies only to es_ru; Rafa speaks Spanish and all scenarios are pinned to A2.

-- ---------------------------------------------------------------------------
-- 1. Backing learning_items for the new scenarios.
-- ---------------------------------------------------------------------------
WITH scen(code, cefr_level, title) AS (
    VALUES
        ('bus_rafa_ticket',          'A2', 'Купить билет в автобусе'),
        ('bus_rafa_next_stop',       'A2', 'Спросить о следующей остановке'),
        ('bus_rafa_wrong_bus',       'A2', 'Объяснить, что сел не туда'),
        ('bus_rafa_passenger_story', 'A2', 'Пересказать историю пассажира'),
        ('bus_rafa_old_route',       'A2', 'Узнать о старом маршруте'),
        ('bus_rafa_last_stop',       'A2', 'Решить, ехать ли до последней остановки')
)
INSERT INTO learning_items (course_id, district_id, location_id, item_type, source_kind, source_id, title, cefr_level, status)
SELECT c.id, d.id, l.id, 'speaking_task', 'conversation_scenario', s.code, s.title, s.cefr_level, 'published'
FROM scen s
JOIN courses c ON c.code = 'es_ru'
JOIN districts d ON d.course_id = c.id AND d.level_code = s.cefr_level
JOIN locations l ON l.district_id = d.id AND l.code = 'conversation'
ON CONFLICT (course_id, source_kind, source_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2. Rafa bus driver quest chain.
-- ---------------------------------------------------------------------------
WITH scen(code, place_type, cefr_level, title, npc_name, npc_code, prerequisite_code, npc_persona, scene_setup, is_quest, max_turns, token_budget, sort_order) AS (
    VALUES
        ('bus_rafa_ticket', 'bus', 'A2', 'Купить билет в автобусе',
            'Rafa', 'rafa_bus_driver', '',
            'a talkative but patient city bus driver who speaks clear A2 Spanish, knows every stop, and helps passengers with simple travel problems',
            'The learner gets on Rafa''s bus for the first time. You greet them from the driver''s seat, ask where they want to go, and help them buy the right ticket.',
            true, 30, 40000, 0),
        ('bus_rafa_next_stop', 'bus', 'A2', 'Спросить о следующей остановке',
            'Rafa', 'rafa_bus_driver', 'bus_rafa_ticket',
            'a friendly city bus driver who announces stops clearly, answers route questions, and keeps explanations suitable for A2 Spanish learners',
            'The learner is riding the bus and is not sure where they are. You help them understand the next stop, nearby places, and whether they should get ready.',
            true, 30, 40000, 1),
        ('bus_rafa_wrong_bus', 'bus', 'A2', 'Объяснить, что сел не туда',
            'Rafa', 'rafa_bus_driver', 'bus_rafa_next_stop',
            'a calm city bus driver who helps lost passengers without blame, speaks clear A2 Spanish, and gives practical next steps',
            'The learner realizes they may be on the wrong bus. You listen, ask where they wanted to go, and explain a simple way to fix the route.',
            true, 30, 40000, 2),
        ('bus_rafa_passenger_story', 'bus', 'A2', 'Пересказать историю пассажира',
            'Rafa', 'rafa_bus_driver', 'bus_rafa_wrong_bus',
            'a talkative bus driver who knows small stories from the city, speaks clear A2 Spanish, and invites the learner to retell events in simple order',
            'The bus is quieter now. You tell the learner about a passenger from an earlier stop and ask them to repeat the main story to check they understood.',
            true, 30, 40000, 3),
        ('bus_rafa_old_route', 'bus', 'A2', 'Узнать о старом маршруте',
            'Rafa', 'rafa_bus_driver', 'bus_rafa_passenger_story',
            'a nostalgic city bus driver who remembers old routes, explains changes simply, and keeps the conversation grounded in everyday A2 Spanish',
            'The bus passes a street Rafa used to drive through. You tell the learner there was once an older route and invite them to ask why it changed.',
            true, 30, 40000, 4),
        ('bus_rafa_last_stop', 'bus', 'A2', 'Решить, ехать ли до последней остановки',
            'Rafa', 'rafa_bus_driver', 'bus_rafa_old_route',
            'a thoughtful bus driver who knows the last stop well, gives simple pros and cons, and helps the learner make a practical travel decision in A2 Spanish',
            'It is late in the route. Rafa says the last stop has a view of the city but may not be useful for everyone. You help the learner decide whether to stay on the bus or get off earlier.',
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
        -- bus_rafa_ticket
        ('bus_rafa_ticket', 'greet',           0, true,  'Поздороваться с Рафой',
            'The learner greets Rafa and starts the bus conversation politely.'),
        ('bus_rafa_ticket', 'ask_ticket',      1, true,  'Попросить билет',
            'The learner asks to buy or get a bus ticket.'),
        ('bus_rafa_ticket', 'say_destination', 2, true,  'Назвать остановку или район',
            'The learner says where they want to go, naming a stop, area, or destination.'),
        ('bus_rafa_ticket', 'confirm_ticket',  3, true,  'Подтвердить билет',
            'The learner confirms the ticket, route, or destination Rafa offers.'),
        ('bus_rafa_ticket', 'thank',           4, false, 'Поблагодарить',
            'The learner thanks Rafa and/or ends the exchange politely.'),

        -- bus_rafa_next_stop
        ('bus_rafa_next_stop', 'greet',         0, true,  'Поздороваться с Рафой',
            'The learner greets Rafa or politely gets his attention.'),
        ('bus_rafa_next_stop', 'ask_next_stop', 1, true,  'Спросить о следующей остановке',
            'The learner asks what the next stop is or whether a specific stop is next.'),
        ('bus_rafa_next_stop', 'ask_landmark',  2, true,  'Уточнить ориентир',
            'The learner asks about a nearby place, landmark, or useful detail for recognizing the stop.'),
        ('bus_rafa_next_stop', 'confirm_ready', 3, true,  'Подтвердить, когда выходить',
            'The learner confirms whether they should get ready, wait, or get off soon.'),
        ('bus_rafa_next_stop', 'thank',         4, false, 'Поблагодарить',
            'The learner thanks Rafa and/or closes the exchange politely.'),

        -- bus_rafa_wrong_bus
        ('bus_rafa_wrong_bus', 'greet',              0, true,  'Обратиться к Рафе',
            'The learner politely gets Rafa''s attention.'),
        ('bus_rafa_wrong_bus', 'say_wrong_bus',      1, true,  'Сказать, что автобус не тот',
            'The learner explains that they may be on the wrong bus or route.'),
        ('bus_rafa_wrong_bus', 'say_target_place',   2, true,  'Назвать нужное место',
            'The learner says where they intended or needed to go.'),
        ('bus_rafa_wrong_bus', 'ask_next_action',    3, true,  'Спросить, что делать дальше',
            'The learner asks for a practical next step to fix the route problem.'),
        ('bus_rafa_wrong_bus', 'accept_instruction', 4, false, 'Принять совет',
            'The learner accepts or confirms Rafa''s suggested next action.'),

        -- bus_rafa_passenger_story
        ('bus_rafa_passenger_story', 'greet',          0, true,  'Поздороваться с Рафой',
            'The learner greets Rafa and agrees to listen to the passenger story.'),
        ('bus_rafa_passenger_story', 'ask_story',      1, true,  'Попросить рассказать историю',
            'The learner asks Rafa about the passenger or invites him to tell the story.'),
        ('bus_rafa_passenger_story', 'retell_events',  2, true,  'Пересказать основные события',
            'The learner retells the main events of the passenger story in a simple order.'),
        ('bus_rafa_passenger_story', 'react_story',    3, true,  'Отреагировать на историю',
            'The learner gives a simple reaction, opinion, or feeling about the story.'),
        ('bus_rafa_passenger_story', 'ask_next_detail',4, false, 'Спросить дополнительную деталь',
            'The learner asks one follow-up question about the passenger, place, or result.'),

        -- bus_rafa_old_route
        ('bus_rafa_old_route', 'greet',             0, true,  'Поздороваться с Рафой',
            'The learner greets Rafa and continues the bus conversation.'),
        ('bus_rafa_old_route', 'ask_old_route',     1, true,  'Спросить о старом маршруте',
            'The learner asks Rafa about the old route or what it was like before.'),
        ('bus_rafa_old_route', 'ask_why_changed',   2, true,  'Узнать, почему маршрут изменился',
            'The learner asks why the route changed or what happened to it.'),
        ('bus_rafa_old_route', 'compare_routes',    3, true,  'Сравнить старый и новый маршрут',
            'The learner compares the old and current routes using simple differences.'),
        ('bus_rafa_old_route', 'say_preference',    4, false, 'Сказать, какой маршрут лучше',
            'The learner says which route they prefer or which route seems better for passengers.'),

        -- bus_rafa_last_stop
        ('bus_rafa_last_stop', 'greet',          0, true,  'Поздороваться с Рафой',
            'The learner greets Rafa or starts the final route conversation politely.'),
        ('bus_rafa_last_stop', 'ask_last_stop',  1, true,  'Спросить о последней остановке',
            'The learner asks what is at the last stop or what happens there.'),
        ('bus_rafa_last_stop', 'ask_pros_cons',  2, true,  'Уточнить плюсы и минусы',
            'The learner asks for reasons to continue to the last stop or get off earlier.'),
        ('bus_rafa_last_stop', 'make_decision',  3, true,  'Принять решение',
            'The learner decides whether to ride to the last stop or leave the bus earlier.'),
        ('bus_rafa_last_stop', 'explain_reason', 4, false, 'Объяснить причину',
            'The learner gives a simple reason for the decision.')
)
INSERT INTO conversation_tasks (scenario_id, code, sort_order, is_required, title, completion_criteria)
SELECT cs.id, t.code, t.sort_order, t.is_required, t.title, t.completion_criteria
FROM task t
JOIN conversation_scenarios cs ON cs.code = t.scenario_code
JOIN courses c ON c.id = cs.course_id AND c.code = 'es_ru'
ON CONFLICT (scenario_id, code) DO NOTHING;
