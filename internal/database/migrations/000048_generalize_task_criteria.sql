-- Generalize conversation task completion_criteria: the original seeds embedded concrete
-- English/Spanish example phrases (e.g. "where is the bread / milk", "cafe con leche") which could
-- make the model judge a learner's message too literally and miss valid completions in the target
-- language. Rewrite the criteria as purely semantic, language-neutral instructions.

WITH crit(scenario_code, task_code, completion_criteria) AS (
    VALUES
        -- cafe_order_coffee
        ('cafe_order_coffee', 'order', 'The learner orders a coffee with milk.'),
        ('cafe_order_coffee', 'sugar', 'The learner asks for two sugars / two sugar cubes for the coffee.'),
        -- shop_buy_water
        ('shop_buy_water', 'ask_water', 'The learner asks for a bottle of water.'),
        ('shop_buy_water', 'ask_price', 'The learner asks the price or how much it costs.'),
        -- police_report_lost
        ('police_report_lost', 'describe', 'The learner describes the lost item (what it is, its colour or size, or where it was lost).'),
        -- cafe_order_pastry
        ('cafe_order_pastry', 'ask_options', 'The learner asks what pastries or food are available, or asks for a specific item.'),
        ('cafe_order_pastry', 'order', 'The learner orders a specific pastry or food item.'),
        ('cafe_order_pastry', 'ask_price', 'The learner asks the price or how much it costs.'),
        -- shop_ask_directions
        ('shop_ask_directions', 'ask_location', 'The learner asks where to find a product, or which aisle or section it is in.'),
        ('shop_ask_directions', 'confirm', 'The learner confirms or repeats back the location or directions the assistant gave.'),
        -- shop_return_item
        ('shop_return_item', 'explain_problem', 'The learner says they want to return or exchange an item, or that there is a problem with it.'),
        ('shop_return_item', 'give_reason', 'The learner gives a reason for the return (broken, wrong size, expired, not needed, etc.).'),
        ('shop_return_item', 'ask_resolution', 'The learner asks for a refund or an exchange.'),
        -- police_describe_person
        ('police_describe_person', 'describe_appearance', 'The learner describes the person''s appearance (height, hair, clothes, or age).'),
        ('police_describe_person', 'describe_action', 'The learner describes what the person was doing or where they went.'),
        ('police_describe_person', 'answer_followup', 'The learner answers at least one follow-up question asked by the officer.'),
        -- police_ask_next_steps
        ('police_ask_next_steps', 'ask_next', 'The learner asks what will happen next or what they should do.'),
        ('police_ask_next_steps', 'ask_contact', 'The learner asks how or when the police will contact them.'),
        ('police_ask_next_steps', 'confirm', 'The learner confirms they understood the procedure.')
)
UPDATE conversation_tasks t
SET completion_criteria = crit.completion_criteria
FROM crit
JOIN conversation_scenarios cs ON cs.code = crit.scenario_code
WHERE t.scenario_id = cs.id AND t.code = crit.task_code;
