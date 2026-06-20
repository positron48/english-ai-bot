-- Delete all Spanish (es_ru) word-set content (categories, sets, items).
-- New es_ru word sets will be recreated with explicit level_code bindings.
-- word_cards / training_cards / tts_generation_status are intentionally left
-- untouched: they are shared by word and may be reused by the new sets.

DELETE FROM word_set_items
WHERE word_set_id IN (SELECT id FROM word_sets WHERE course_code = 'es_ru');

DELETE FROM word_sets WHERE course_code = 'es_ru';

DELETE FROM word_set_categories WHERE course_code = 'es_ru';

-- Clean up the Linglow v2 mirror (modules/learning_items) that MapLegacyContent
-- previously created for the deleted word sets, so the city map (word_market)
-- doesn't keep showing stale entries until the next legacy-content sync.
-- 'custom' is the user-personal-vocabulary module (not backed by a word set) and must stay.
DELETE FROM learning_items li
USING modules m, courses c
WHERE li.module_id = m.id
  AND m.course_id = c.id
  AND c.code = 'es_ru'
  AND m.module_type = 'word_set'
  AND m.source_id != 'custom';

DELETE FROM modules m
USING courses c
WHERE m.course_id = c.id
  AND c.code = 'es_ru'
  AND m.module_type = 'word_set'
  AND m.source_id != 'custom';
