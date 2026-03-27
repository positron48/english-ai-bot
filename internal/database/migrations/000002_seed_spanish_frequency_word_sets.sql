-- Seed: Spanish frequency word set categories + sets (no items).
-- Idempotent-ish: inserts are guarded by NOT EXISTS checks.

-- Root categories (start from Core/Extended; higher-level navigation lives in UI)
WITH ins_core AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    NULL,
    'Core Frequency',
    'High-frequency everyday words used across most conversations and texts.',
    1,
    0
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id IS NULL AND c.name = 'Core Frequency'
  )
  RETURNING id
),
ins_ext AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    NULL,
    'Extended Frequency',
    'Mid-frequency words that expand everyday vocabulary beyond the core.',
    1,
    1
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id IS NULL AND c.name = 'Extended Frequency'
  )
  RETURNING id
),
core AS (
  SELECT id FROM word_set_categories WHERE parent_id IS NULL AND name = 'Core Frequency' LIMIT 1
),
ext AS (
  SELECT id FROM word_set_categories WHERE parent_id IS NULL AND name = 'Extended Frequency' LIMIT 1
),
-- Child categories under Core
ins_core_verbs AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    (SELECT id FROM core),
    'Top 500 Verbs',
    'High-frequency verbs used across everyday speech and texts.',
    1,
    0
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id = (SELECT id FROM core) AND c.name = 'Top 500 Verbs'
  )
  RETURNING id
),
ins_core_nouns AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    (SELECT id FROM core),
    'Top 500 Nouns',
    'High-frequency everyday nouns used across most conversations and texts.',
    1,
    1
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id = (SELECT id FROM core) AND c.name = 'Top 500 Nouns'
  )
  RETURNING id
),
ins_core_adjs AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    (SELECT id FROM core),
    'Top 500 Adjectives',
    'High-frequency adjectives that help you describe people, things, and situations.',
    1,
    2
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id = (SELECT id FROM core) AND c.name = 'Top 500 Adjectives'
  )
  RETURNING id
),
ins_core_advs AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    (SELECT id FROM core),
    'Top Adverbs',
    'High-frequency adverbs used to express time, frequency, degree, manner, and certainty.',
    1,
    3
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id = (SELECT id FROM core) AND c.name = 'Top Adverbs'
  )
  RETURNING id
),
ins_core_nouns_500_1000 AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    (SELECT id FROM core),
    'Top 500–1000 Nouns',
    'Mid-frequency nouns that expand core vocabulary and improve precision.',
    1,
    4
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id = (SELECT id FROM core) AND c.name = 'Top 500–1000 Nouns'
  )
  RETURNING id
),
-- Child categories under Extended (split 1001–2000 into two categories of 500)
ins_ext_nouns_1001_1500 AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    (SELECT id FROM ext),
    'Top 1001–1500 Nouns',
    'Extended-frequency nouns (ranks 1001–1500 within nouns).',
    1,
    0
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id = (SELECT id FROM ext) AND c.name = 'Top 1001–1500 Nouns'
  )
  RETURNING id
),
ins_ext_nouns_1501_2000 AS (
  INSERT INTO word_set_categories (parent_id, name, description, is_published, sort_order)
  SELECT
    (SELECT id FROM ext),
    'Top 1501–2000 Nouns',
    'Extended-frequency nouns (ranks 1501–2000 within nouns).',
    1,
    1
  WHERE NOT EXISTS (
    SELECT 1 FROM word_set_categories c
    WHERE c.parent_id = (SELECT id FROM ext) AND c.name = 'Top 1501–2000 Nouns'
  )
  RETURNING id
),
core_verbs AS (
  SELECT id FROM word_set_categories WHERE parent_id = (SELECT id FROM core) AND name = 'Top 500 Verbs' LIMIT 1
),
core_nouns AS (
  SELECT id FROM word_set_categories WHERE parent_id = (SELECT id FROM core) AND name = 'Top 500 Nouns' LIMIT 1
),
core_adjs AS (
  SELECT id FROM word_set_categories WHERE parent_id = (SELECT id FROM core) AND name = 'Top 500 Adjectives' LIMIT 1
),
core_advs AS (
  SELECT id FROM word_set_categories WHERE parent_id = (SELECT id FROM core) AND name = 'Top Adverbs' LIMIT 1
),
core_nouns_500_1000 AS (
  SELECT id FROM word_set_categories WHERE parent_id = (SELECT id FROM core) AND name = 'Top 500–1000 Nouns' LIMIT 1
),
ext_nouns_1001_1500 AS (
  SELECT id FROM word_set_categories WHERE parent_id = (SELECT id FROM ext) AND name = 'Top 1001–1500 Nouns' LIMIT 1
),
ext_nouns_1501_2000 AS (
  SELECT id FROM word_set_categories WHERE parent_id = (SELECT id FROM ext) AND name = 'Top 1501–2000 Nouns' LIMIT 1
)
SELECT 1;

-- Core Frequency / Top 500 Verbs: 10 sets of 50
INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
SELECT
  (SELECT id FROM word_set_categories WHERE name='Top 500 Verbs' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1),
  s.title,
  s.description,
  1,
  s.sort_order,
  'verb'
FROM (
  VALUES
    (0,  'Core Verbs — Top 50 (Ranks 1–50)',     'Verbs ranked 1–50 by frequency (within verbs).'),
    (1,  'Core Verbs — Top 50 (Ranks 51–100)',   'Verbs ranked 51–100 by frequency (within verbs).'),
    (2,  'Core Verbs — Top 50 (Ranks 101–150)',  'Verbs ranked 101–150 by frequency (within verbs).'),
    (3,  'Core Verbs — Top 50 (Ranks 151–200)',  'Verbs ranked 151–200 by frequency (within verbs).'),
    (4,  'Core Verbs — Top 50 (Ranks 201–250)',  'Verbs ranked 201–250 by frequency (within verbs).'),
    (5,  'Core Verbs — Top 50 (Ranks 251–300)',  'Verbs ranked 251–300 by frequency (within verbs).'),
    (6,  'Core Verbs — Top 50 (Ranks 301–350)',  'Verbs ranked 301–350 by frequency (within verbs).'),
    (7,  'Core Verbs — Top 50 (Ranks 351–400)',  'Verbs ranked 351–400 by frequency (within verbs).'),
    (8,  'Core Verbs — Top 50 (Ranks 401–450)',  'Verbs ranked 401–450 by frequency (within verbs).'),
    (9,  'Core Verbs — Top 50 (Ranks 451–500)',  'Verbs ranked 451–500 by frequency (within verbs).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE name='Top 500 Verbs' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1)
);

-- Core Frequency / Top 500 Nouns: 10 sets of 50
INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
SELECT
  (SELECT id FROM word_set_categories WHERE name='Top 500 Nouns' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1),
  s.title,
  s.description,
  1,
  s.sort_order,
  'noun'
FROM (
  VALUES
    (0,  'Core Nouns — Top 50 (Ranks 1–50)',     'Nouns ranked 1–50 by frequency (within nouns).'),
    (1,  'Core Nouns — Top 50 (Ranks 51–100)',   'Nouns ranked 51–100 by frequency (within nouns).'),
    (2,  'Core Nouns — Top 50 (Ranks 101–150)',  'Nouns ranked 101–150 by frequency (within nouns).'),
    (3,  'Core Nouns — Top 50 (Ranks 151–200)',  'Nouns ranked 151–200 by frequency (within nouns).'),
    (4,  'Core Nouns — Top 50 (Ranks 201–250)',  'Nouns ranked 201–250 by frequency (within nouns).'),
    (5,  'Core Nouns — Top 50 (Ranks 251–300)',  'Nouns ranked 251–300 by frequency (within nouns).'),
    (6,  'Core Nouns — Top 50 (Ranks 301–350)',  'Nouns ranked 301–350 by frequency (within nouns).'),
    (7,  'Core Nouns — Top 50 (Ranks 351–400)',  'Nouns ranked 351–400 by frequency (within nouns).'),
    (8,  'Core Nouns — Top 50 (Ranks 401–450)',  'Nouns ranked 401–450 by frequency (within nouns).'),
    (9,  'Core Nouns — Top 50 (Ranks 451–500)',  'Nouns ranked 451–500 by frequency (within nouns).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE name='Top 500 Nouns' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1)
);

-- Core Frequency / Top 500 Adjectives: 10 sets of 50
INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
SELECT
  (SELECT id FROM word_set_categories WHERE name='Top 500 Adjectives' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1),
  s.title,
  s.description,
  1,
  s.sort_order,
  'adjective'
FROM (
  VALUES
    (0,  'Core Adjectives — Top 50 (Ranks 1–50)',     'Adjectives ranked 1–50 by frequency (within adjectives).'),
    (1,  'Core Adjectives — Top 50 (Ranks 51–100)',   'Adjectives ranked 51–100 by frequency (within adjectives).'),
    (2,  'Core Adjectives — Top 50 (Ranks 101–150)',  'Adjectives ranked 101–150 by frequency (within adjectives).'),
    (3,  'Core Adjectives — Top 50 (Ranks 151–200)',  'Adjectives ranked 151–200 by frequency (within adjectives).'),
    (4,  'Core Adjectives — Top 50 (Ranks 201–250)',  'Adjectives ranked 201–250 by frequency (within adjectives).'),
    (5,  'Core Adjectives — Top 50 (Ranks 251–300)',  'Adjectives ranked 251–300 by frequency (within adjectives).'),
    (6,  'Core Adjectives — Top 50 (Ranks 301–350)',  'Adjectives ranked 301–350 by frequency (within adjectives).'),
    (7,  'Core Adjectives — Top 50 (Ranks 351–400)',  'Adjectives ranked 351–400 by frequency (within adjectives).'),
    (8,  'Core Adjectives — Top 50 (Ranks 401–450)',  'Adjectives ranked 401–450 by frequency (within adjectives).'),
    (9,  'Core Adjectives — Top 50 (Ranks 451–500)',  'Adjectives ranked 451–500 by frequency (within adjectives).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE name='Top 500 Adjectives' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1)
);

-- Core Frequency / Top Adverbs: create up to 6 blocks (1–50 ... 251–300). Data availability is handled at import-time.
INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
SELECT
  (SELECT id FROM word_set_categories WHERE name='Top Adverbs' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1),
  s.title,
  s.description,
  1,
  s.sort_order,
  'adverb'
FROM (
  VALUES
    (0,  'Core Adverbs — Top 50 (Ranks 1–50)',     'Adverbs ranked 1–50 by frequency (within adverbs).'),
    (1,  'Core Adverbs — Top 50 (Ranks 51–100)',   'Adverbs ranked 51–100 by frequency (within adverbs).'),
    (2,  'Core Adverbs — Top 50 (Ranks 101–150)',  'Adverbs ranked 101–150 by frequency (within adverbs).'),
    (3,  'Core Adverbs — Top 50 (Ranks 151–200)',  'Adverbs ranked 151–200 by frequency (within adverbs).'),
    (4,  'Core Adverbs — Top 50 (Ranks 201–250)',  'Adverbs ranked 201–250 by frequency (within adverbs).'),
    (5,  'Core Adverbs — Top 50 (Ranks 251–300)',  'Adverbs ranked 251–300 by frequency (within adverbs).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE name='Top Adverbs' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1)
);

-- Core Frequency / Top 500–1000 Nouns: 10 sets of 50, ranks 501–1000 within nouns
INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
SELECT
  (SELECT id FROM word_set_categories WHERE name='Top 500–1000 Nouns' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1),
  s.title,
  s.description,
  1,
  s.sort_order,
  'noun'
FROM (
  VALUES
    (0, 'Core Nouns — Top 50 (Ranks 501–550)',  'Nouns ranked 501–550 by frequency (within nouns).'),
    (1, 'Core Nouns — Top 50 (Ranks 551–600)',  'Nouns ranked 551–600 by frequency (within nouns).'),
    (2, 'Core Nouns — Top 50 (Ranks 601–650)',  'Nouns ranked 601–650 by frequency (within nouns).'),
    (3, 'Core Nouns — Top 50 (Ranks 651–700)',  'Nouns ranked 651–700 by frequency (within nouns).'),
    (4, 'Core Nouns — Top 50 (Ranks 701–750)',  'Nouns ranked 701–750 by frequency (within nouns).'),
    (5, 'Core Nouns — Top 50 (Ranks 751–800)',  'Nouns ranked 751–800 by frequency (within nouns).'),
    (6, 'Core Nouns — Top 50 (Ranks 801–850)',  'Nouns ranked 801–850 by frequency (within nouns).'),
    (7, 'Core Nouns — Top 50 (Ranks 851–900)',  'Nouns ranked 851–900 by frequency (within nouns).'),
    (8, 'Core Nouns — Top 50 (Ranks 901–950)',  'Nouns ranked 901–950 by frequency (within nouns).'),
    (9, 'Core Nouns — Top 50 (Ranks 951–1000)', 'Nouns ranked 951–1000 by frequency (within nouns).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE name='Top 500–1000 Nouns' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Core Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1)
);

-- Extended Frequency / Top 1001–1500 Nouns: 10 sets of 50
INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
SELECT
  (SELECT id FROM word_set_categories WHERE name='Top 1001–1500 Nouns' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Extended Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1),
  s.title,
  s.description,
  1,
  s.sort_order,
  'noun'
FROM (
  VALUES
    (0, 'Extended Nouns — Top 50 (Ranks 1001–1050)', 'Nouns ranked 1001–1050 by frequency (within nouns).'),
    (1, 'Extended Nouns — Top 50 (Ranks 1051–1100)', 'Nouns ranked 1051–1100 by frequency (within nouns).'),
    (2, 'Extended Nouns — Top 50 (Ranks 1101–1150)', 'Nouns ranked 1101–1150 by frequency (within nouns).'),
    (3, 'Extended Nouns — Top 50 (Ranks 1151–1200)', 'Nouns ranked 1151–1200 by frequency (within nouns).'),
    (4, 'Extended Nouns — Top 50 (Ranks 1201–1250)', 'Nouns ranked 1201–1250 by frequency (within nouns).'),
    (5, 'Extended Nouns — Top 50 (Ranks 1251–1300)', 'Nouns ranked 1251–1300 by frequency (within nouns).'),
    (6, 'Extended Nouns — Top 50 (Ranks 1301–1350)', 'Nouns ranked 1301–1350 by frequency (within nouns).'),
    (7, 'Extended Nouns — Top 50 (Ranks 1351–1400)', 'Nouns ranked 1351–1400 by frequency (within nouns).'),
    (8, 'Extended Nouns — Top 50 (Ranks 1401–1450)', 'Nouns ranked 1401–1450 by frequency (within nouns).'),
    (9, 'Extended Nouns — Top 50 (Ranks 1451–1500)', 'Nouns ranked 1451–1500 by frequency (within nouns).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE name='Top 1001–1500 Nouns' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Extended Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1)
);

-- Extended Frequency / Top 1501–2000 Nouns: 10 sets of 50
INSERT INTO word_sets (category_id, title, description, is_published, sort_order, preferred_pos)
SELECT
  (SELECT id FROM word_set_categories WHERE name='Top 1501–2000 Nouns' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Extended Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1),
  s.title,
  s.description,
  1,
  s.sort_order,
  'noun'
FROM (
  VALUES
    (0, 'Extended Nouns — Top 50 (Ranks 1501–1550)', 'Nouns ranked 1501–1550 by frequency (within nouns).'),
    (1, 'Extended Nouns — Top 50 (Ranks 1551–1600)', 'Nouns ranked 1551–1600 by frequency (within nouns).'),
    (2, 'Extended Nouns — Top 50 (Ranks 1601–1650)', 'Nouns ranked 1601–1650 by frequency (within nouns).'),
    (3, 'Extended Nouns — Top 50 (Ranks 1651–1700)', 'Nouns ranked 1651–1700 by frequency (within nouns).'),
    (4, 'Extended Nouns — Top 50 (Ranks 1701–1750)', 'Nouns ranked 1701–1750 by frequency (within nouns).'),
    (5, 'Extended Nouns — Top 50 (Ranks 1751–1800)', 'Nouns ranked 1751–1800 by frequency (within nouns).'),
    (6, 'Extended Nouns — Top 50 (Ranks 1801–1850)', 'Nouns ranked 1801–1850 by frequency (within nouns).'),
    (7, 'Extended Nouns — Top 50 (Ranks 1851–1900)', 'Nouns ranked 1851–1900 by frequency (within nouns).'),
    (8, 'Extended Nouns — Top 50 (Ranks 1901–1950)', 'Nouns ranked 1901–1950 by frequency (within nouns).'),
    (9, 'Extended Nouns — Top 50 (Ranks 1951–2000)', 'Nouns ranked 1951–2000 by frequency (within nouns).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE name='Top 1501–2000 Nouns' AND parent_id = (SELECT id FROM word_set_categories WHERE name='Extended Frequency' AND parent_id IS NULL LIMIT 1) LIMIT 1)
);

