-- Reported Spanish vocabulary corrections; preserve IDs and learner history.
UPDATE training_cards
SET word_ru = 'пляж',
    example_ru = 'Летом на пляже много людей.',
    transcription = 'ˈplaʝa'
WHERE id = 7088 AND word_card_id = 5715
  AND word_en = 'playa' AND course_code = 'es_ru';

UPDATE training_cards
SET word_ru = 'медбрат',
    example_ru = 'Медбрат ухаживал за пациентами всю ночь.'
WHERE id = 13558 AND word_card_id = 56735
  AND word_en = 'enfermero' AND course_code = 'es_ru';

UPDATE training_cards
SET word_ru = 'чат',
    meaning_en = 'Conversación por escrito a través de internet.',
    example_en = 'Hablamos por el chat del grupo.',
    example_ru = 'Мы общаемся в групповом чате.',
    distractors_ru = '["звонок", "письмо", "встреча"]',
    distractors_en = '["llamada", "carta", "reunión"]',
    hint = 'Conversación escrita en internet.'
WHERE id = 13565 AND word_card_id = 56746
  AND word_en = 'chat' AND course_code = 'es_ru';

UPDATE training_cards
SET word_ru = 'квартира на последнем этаже',
    meaning_en = 'Vivienda situada en el último piso de un edificio.',
    example_en = 'Desde nuestro ático se ve toda la ciudad.',
    example_ru = 'Из нашей квартиры на последнем этаже виден весь город.',
    distractors_ru = '["подвал", "гараж", "кухня"]',
    distractors_en = '["sótano", "garaje", "cocina"]',
    hint = 'Vivienda en la parte más alta de un edificio.'
WHERE id = 19233 AND word_card_id = 298441
  AND word_en = 'ático' AND course_code = 'es_ru';

UPDATE training_cards
SET word_ru = 'аранский язык',
    example_ru = 'Аранский язык имеет официальный статус в долине Аран.'
WHERE id = 19209 AND word_card_id = 298394
  AND word_en = 'aranés' AND course_code = 'es_ru';

UPDATE word_cards
SET definition_ru = 'пляж',
    transcription = 'ˈplaʝa',
    examples_json = '[{"example_en":"La playa está cerca del hotel.","example_target":"La playa está cerca del hotel.","gloss_native":"Пляж находится рядом с отелем.","gloss_ru":"Пляж находится рядом с отелем."},{"example_en":"En verano vamos a la playa.","example_target":"En verano vamos a la playa.","gloss_native":"Летом мы ходим на пляж.","gloss_ru":"Летом мы ходим на пляж."}]',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 5715 AND word = 'playa' AND course_code = 'es_ru';

UPDATE word_cards
SET definition_ru = 'медбрат; медицинский брат',
    transcription = 'emfeɾˈmeɾo',
    examples_json = '[{"example_en":"El enfermero cuida a los pacientes.","example_target":"El enfermero cuida a los pacientes.","gloss_native":"Медбрат заботится о пациентах.","gloss_ru":"Медбрат заботится о пациентах."},{"example_en":"Mi hermano es enfermero en el hospital.","example_target":"Mi hermano es enfermero en el hospital.","gloss_native":"Мой брат работает медбратом в больнице.","gloss_ru":"Мой брат работает медбратом в больнице."}]',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 56735 AND word = 'enfermero' AND course_code = 'es_ru';

UPDATE word_cards
SET definition_ru = 'аранский язык',
    transcription = 'aɾaˈnes',
    examples_json = '[{"example_en":"El aranés se habla en el valle de Arán.","example_target":"El aranés se habla en el valle de Arán.","gloss_native":"На аранском языке говорят в долине Аран.","gloss_ru":"На аранском языке говорят в долине Аран."},{"example_en":"Ella estudia aranés.","example_target":"Ella estudia aranés.","gloss_native":"Она изучает аранский язык.","gloss_ru":"Она изучает аранский язык."}]',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 298394 AND word = 'aranés' AND course_code = 'es_ru';

UPDATE training_cards
SET display_word = 'Extremadura'
WHERE id = 19217 AND word_card_id = 298413
  AND word_en = 'extremadura' AND course_code = 'es_ru';

UPDATE word_cards
SET display_en = 'Extremadura',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 298413 AND word = 'extremadura' AND course_code = 'es_ru';
