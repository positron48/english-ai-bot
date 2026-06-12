-- 000025: Link reading_texts learning_items to districts by cefr_level.
--
-- reading_text items imported before this migration may have district_id = NULL
-- when the level-to-district mapping wasn't applied at import time.
-- This migration assigns them to the correct district within their course.

-- Link unassigned reading_text learning items to districts using cefr_level.
UPDATE learning_items li
SET district_id = d.id
FROM districts d
WHERE li.district_id IS NULL
  AND li.item_type = 'reading_text'
  AND li.cefr_level IS NOT NULL
  AND li.cefr_level != ''
  AND d.course_id = li.course_id
  AND d.level_code = li.cefr_level;

-- Same for grammar_chapter items that may be unlinked.
UPDATE learning_items li
SET district_id = d.id
FROM districts d
WHERE li.district_id IS NULL
  AND li.item_type IN ('grammar_chapter', 'grammar_theory_block')
  AND li.cefr_level IS NOT NULL
  AND li.cefr_level != ''
  AND d.course_id = li.course_id
  AND d.level_code = li.cefr_level;

-- Words: assign to the district matching their cefr_level (if known).
-- Words without a level stay unassigned (they show in the review station).
UPDATE learning_items li
SET district_id = d.id
FROM districts d
WHERE li.district_id IS NULL
  AND li.item_type = 'word'
  AND li.cefr_level IS NOT NULL
  AND li.cefr_level != ''
  AND d.course_id = li.course_id
  AND d.level_code = li.cefr_level;
