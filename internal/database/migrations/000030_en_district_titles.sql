-- Update district titles for the English course (en_ru) to English names.
-- The Spanish course (es_ru) keeps its original Spanish titles.
UPDATE districts d
SET title = CASE d.code
    WHEN 'a0_spark_gate'     THEN 'Spark Gate'
    WHEN 'a1_clear_plaza'    THEN 'Clear Plaza'
    WHEN 'a2_living_quarter' THEN 'Living Quarter'
    WHEN 'b1_story_bridges'  THEN 'Story Bridges'
    WHEN 'b2_high_district'  THEN 'High District'
    WHEN 'c1_mastery_campus' THEN 'Mastery Campus'
END,
updated_at = CURRENT_TIMESTAMP
FROM courses c
WHERE d.course_id = c.id
  AND c.code = 'en_ru'
  AND d.code IN ('a0_spark_gate','a1_clear_plaza','a2_living_quarter',
                 'b1_story_bridges','b2_high_district','c1_mastery_campus');
