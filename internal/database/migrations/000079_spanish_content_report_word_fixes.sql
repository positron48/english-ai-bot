-- Correct reported Spanish cards. IDs and course guards keep this one-time data fix scoped.
UPDATE training_cards
SET word_ru = 'чемодан',
    example_ru = 'Я взял чемодан в аэропорт.',
    distractors_ru = '["рюкзак","коробка","портфель"]'
WHERE id = 13595 AND word_card_id = 56776 AND word_en = 'valija'
  AND course_code = 'es_ru';

UPDATE training_cards
SET word_ru = 'бензин'
WHERE id = 6682 AND word_card_id = 10493 AND word_en = 'gasolina'
  AND course_code = 'es_ru';

UPDATE word_cards
SET examples_json = replace(examples_json, 'Машине нужна бензин.', 'Машине нужен бензин.'),
    updated_at = CURRENT_TIMESTAMP
WHERE id = 10493 AND word = 'gasolina' AND course_code = 'es_ru'
  AND examples_json LIKE '%Машине нужна бензин.%';

UPDATE training_cards
SET word_ru = 'му (звук коровы)',
    meaning_en = 'Onomatopeya que imita la voz de una vaca.',
    example_en = 'La vaca hace mu en el establo.',
    example_ru = 'Корова мычит в стойле.',
    distractors_ru = '["мяу","гав","бе-е"]',
    hint = 'Es la voz de la vaca.'
WHERE id = 19227 AND word_card_id = 298434 AND word_en = 'mu'
  AND course_code = 'es_ru';

UPDATE word_cards
SET definition_ru = 'му; звук, издаваемый коровой',
    examples_json = '[{"example_en":"La vaca hace mu.","example_target":"La vaca hace mu.","gloss_native":"Корова мычит.","gloss_ru":"Корова мычит."},{"example_en":"Oí un mu cerca del establo.","example_target":"Oí un mu cerca del establo.","gloss_native":"Я услышал мычание возле стойла.","gloss_ru":"Я услышал мычание возле стойла."}]',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 298434 AND word = 'mu' AND course_code = 'es_ru';
