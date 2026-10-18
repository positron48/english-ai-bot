-- Keep the reported Spanish attic sense consistent in dictionary and training.
UPDATE training_cards
SET word_ru = 'чердак',
    meaning_en = 'Espacio situado bajo el tejado de una casa, usado para guardar objetos.',
    example_en = 'Guardamos las cajas en el ático.',
    example_ru = 'Мы храним коробки на чердаке.',
    hint = 'Espacio bajo el tejado donde se guardan objetos.'
WHERE id = 19233 AND word_card_id = 298441
  AND word_en = 'ático' AND course_code = 'es_ru';

UPDATE word_cards
SET definition_ru = 'чердак',
    opposite_gender_word = NULL,
    examples_json = '[{"example_en":"Guardamos las cajas en el ático.","example_target":"Guardamos las cajas en el ático.","gloss_ru":"Мы храним коробки на чердаке.","gloss_native":"Мы храним коробки на чердаке."},{"example_en":"El ático está lleno de cosas viejas.","example_target":"El ático está lleno de cosas viejas.","gloss_ru":"Чердак полон старых вещей.","gloss_native":"Чердак полон старых вещей."}]',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 298441 AND word = 'ático' AND course_code = 'es_ru';
