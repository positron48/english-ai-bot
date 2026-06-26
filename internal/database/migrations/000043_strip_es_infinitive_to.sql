-- Spanish infinitives must not carry the English "to " marker. Earlier generation
-- stored display forms like "to comer" for es_ru; strip the leading "to " so word
-- training (MCQ/spell/type) and the dictionary show bare Spanish infinitives.
-- "to %" (with the trailing space) only matches the infinitive marker, never a real
-- Spanish lemma, so this is safe.

UPDATE word_cards
SET display_en = substring(display_en FROM 4)
WHERE LOWER(course_code) = 'es_ru' AND display_en LIKE 'to %';

UPDATE training_cards
SET display_word = substring(display_word FROM 4)
WHERE LOWER(course_code) = 'es_ru' AND display_word LIKE 'to %';
