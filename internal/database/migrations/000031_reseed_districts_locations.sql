-- Reseed districts and locations for all courses.
-- Uses ON CONFLICT DO NOTHING so it is safe to run on prod DBs that already have this data.
-- Needed for linglow_unified where schema_migrations was pre-populated from a prod dump
-- but the actual seed INSERTs from 000017 never ran.

WITH district_seed(level_code, code, title, sort_order) AS (
    VALUES
        ('A0', 'a0_spark_gate',     'Puerta de la Chispa', 0),
        ('A1', 'a1_clear_plaza',    'Plaza Clara',          1),
        ('A2', 'a2_living_quarter', 'Barrio Vivo',          2),
        ('B1', 'b1_story_bridges',  'Puentes del Relato',   3),
        ('B2', 'b2_high_district',  'Distrito Alto',        4),
        ('C1', 'c1_mastery_campus', 'Campus de Maestria',   5)
)
INSERT INTO districts (course_id, code, level_code, title, sort_order, status)
SELECT c.id, ds.code, ds.level_code, ds.title, ds.sort_order, 'active'
FROM courses c
CROSS JOIN district_seed ds
WHERE c.code IN ('en_ru', 'es_ru')
ON CONFLICT (course_id, code) DO NOTHING;

WITH location_seed(code, location_type, title, sort_order) AS (
    VALUES
        ('grammar',          'grammar',          'Grammar Building',  0),
        ('word_market',      'word_market',      'Word Market',       1),
        ('reading',          'reading',          'Reading Spot',      2),
        ('conversation',     'conversation',     'Conversation Hub',  3),
        ('review_station',   'review_station',   'Review Station',    4),
        ('mistake_workshop', 'mistake_workshop', 'Mistake Workshop',  5)
)
INSERT INTO locations (course_id, district_id, code, location_type, title, sort_order, status)
SELECT d.course_id, d.id, ls.code, ls.location_type, ls.title, ls.sort_order, 'active'
FROM districts d
CROSS JOIN location_seed ls
ON CONFLICT (district_id, code) DO NOTHING;
