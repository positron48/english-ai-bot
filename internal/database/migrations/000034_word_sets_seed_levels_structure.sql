-- Recreate es_ru word-set categories/sets per docs/LINGLOW_WORD_SETS_LEVELS_PLAN.md vocabulary scope plan.
-- Structure only (no words yet): level -> topic category -> word sets, each bound to its CEFR level.
-- A0-A2 / B1-B2 sets map 1:1 onto the planned word-set rows (already sized 8-45 words).
-- C1 rows are split into <= 30 word chunks so no single set is unreasonably large.

-- Self-cleaning: 000033 ran before word_sets/word_set_categories were backfilled with
-- course_code='es_ru', so its DELETE matched nothing and the legacy rows were retagged
-- afterwards. Re-run the same cleanup here so this migration is idempotent going forward.
DELETE FROM word_set_items
WHERE word_set_id IN (SELECT id FROM word_sets WHERE course_code = 'es_ru');

DELETE FROM word_sets WHERE course_code = 'es_ru';

DELETE FROM word_set_categories WHERE course_code = 'es_ru';

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

-- ===== Level A0 =====
INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru', NULL, 'Уровень A0', 'Стартовый уровень: базовые слова для первых фраз и понимания простых обращений.', 1, 0, 'A0'
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.parent_id IS NULL AND c.name = 'Уровень A0'
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1),
  'Служебные слова и грамматический минимум', 'Тематический блок уровня A0 (~35 слов).', 1, 0, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Служебные слова и грамматический минимум'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Служебные слова и грамматический минимум' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Личные местоимения и обращения', 'Личные местоимения и обращения (~8 слов).'),
    (1, 'Указательные и притяжательные основы', 'Указательные и притяжательные основы (~6 слов).'),
    (2, 'Вопросительные слова', 'Вопросительные слова (~8 слов).'),
    (3, 'Самые частые предлоги и союзы', 'Самые частые предлоги и союзы (~8 слов).'),
    (4, 'Частицы, согласие, отрицание, степень', 'Частицы, согласие, отрицание, степень (~5 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Служебные слова и грамматический минимум' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1),
  'Люди и контакт', 'Тематический блок уровня A0 (~25 слов).', 1, 1, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Люди и контакт'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Люди и контакт' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Приветствия, прощания, вежливость', 'Приветствия, прощания, вежливость (~8 слов).'),
    (1, 'Семья и близкие люди', 'Семья и близкие люди (~8 слов).'),
    (2, 'Страны, языки, национальности: минимум', 'Страны, языки, национальности: минимум (~5 слов).'),
    (3, 'Простые роли: человек, друг, мужчина, женщина', 'Простые роли: человек, друг, мужчина, женщина (~4 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Люди и контакт' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1),
  'Базовые действия', 'Тематический блок уровня A0 (~25 слов).', 1, 2, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Базовые действия'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Базовые действия' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Быть, находиться, иметь, делать', 'Быть, находиться, иметь, делать (~8 слов).'),
    (1, 'Движение и повседневные действия', 'Движение и повседневные действия (~10 слов).'),
    (2, 'Хотеть, мочь, знать, понимать, нужно', 'Хотеть, мочь, знать, понимать, нужно (~7 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Базовые действия' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1),
  'Предметы вокруг', 'Тематический блок уровня A0 (~20 слов).', 1, 3, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Предметы вокруг'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Предметы вокруг' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Дом и комната: минимум', 'Дом и комната: минимум (~8 слов).'),
    (1, 'Личные вещи и документы', 'Личные вещи и документы (~7 слов).'),
    (2, 'Телефон, учёба, интернет: минимум', 'Телефон, учёба, интернет: минимум (~5 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Предметы вокруг' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1),
  'Еда, город, покупки', 'Тематический блок уровня A0 (~20 слов).', 1, 4, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Еда, город, покупки'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Еда, город, покупки' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Еда и напитки: минимум', 'Еда и напитки: минимум (~8 слов).'),
    (1, 'Места в городе', 'Места в городе (~6 слов).'),
    (2, 'Транспорт, покупка, оплата: минимум', 'Транспорт, покупка, оплата: минимум (~6 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Еда, город, покупки' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1),
  'Числа, время, свойства', 'Тематический блок уровня A0 (~25 слов).', 1, 5, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Числа, время, свойства'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Числа, время, свойства' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Числа: базовый набор', 'Числа: базовый набор (~12 слов).'),
    (1, 'Дни, части дня, сейчас/потом', 'Дни, части дня, сейчас/потом (~5 слов).'),
    (2, 'Цвета, размер, базовые признаки', 'Цвета, размер, базовые признаки (~8 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Числа, время, свойства' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A0' LIMIT 1) LIMIT 1)
);

-- ===== Level A1 =====
INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru', NULL, 'Уровень A1', 'Простой быт: рассказать о себе, купить, спросить дорогу, заказать еду, описать день.', 1, 1, 'A1'
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.parent_id IS NULL AND c.name = 'Уровень A1'
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1),
  'Служебные слова A1', 'Тематический блок уровня A1 (~55 слов).', 1, 0, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Служебные слова A1'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Служебные слова A1' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Местоимения дополнения: основы', 'Местоимения дополнения: основы (~12 слов).'),
    (1, 'Притяжательные, указательные, неопределённые', 'Притяжательные, указательные, неопределённые (~12 слов).'),
    (2, 'Предлоги места, направления, времени', 'Предлоги места, направления, времени (~12 слов).'),
    (3, 'Союзы и связки простых фраз', 'Союзы и связки простых фраз (~9 слов).'),
    (4, 'Наречия частотности, степени, порядка', 'Наречия частотности, степени, порядка (~10 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Служебные слова A1' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1),
  'Повседневные действия и рутина', 'Тематический блок уровня A1 (~50 слов).', 1, 1, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Повседневные действия и рутина'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Повседневные действия и рутина' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Утро, вечер, бытовой день', 'Утро, вечер, бытовой день (~12 слов).'),
    (1, 'Готовить, есть, покупать', 'Готовить, есть, покупать (~10 слов).'),
    (2, 'Перемещаться, ехать, приходить', 'Перемещаться, ехать, приходить (~10 слов).'),
    (3, 'Просить, давать, помогать, брать', 'Просить, давать, помогать, брать (~10 слов).'),
    (4, 'Чувствовать, думать, помнить: основы', 'Чувствовать, думать, помнить: основы (~8 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Повседневные действия и рутина' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1),
  'Дом, семья, личная информация', 'Тематический блок уровня A1 (~45 слов).', 1, 2, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Дом, семья, личная информация'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Дом, семья, личная информация' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Анкета, контакты, личные данные', 'Анкета, контакты, личные данные (~10 слов).'),
    (1, 'Семья, возраст, семейный статус', 'Семья, возраст, семейный статус (~10 слов).'),
    (2, 'Жильё, комнаты, мебель', 'Жильё, комнаты, мебель (~12 слов).'),
    (3, 'Одежда и внешность', 'Одежда и внешность (~8 слов).'),
    (4, 'Базовые документы', 'Базовые документы (~5 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Дом, семья, личная информация' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1),
  'Еда, покупки, услуги', 'Тематический блок уровня A1 (~55 слов).', 1, 3, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Еда, покупки, услуги'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Еда, покупки, услуги' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Продукты и напитки A1', 'Продукты и напитки A1 (~15 слов).'),
    (1, 'Кафе, ресторан, меню', 'Кафе, ресторан, меню (~10 слов).'),
    (2, 'Магазин, цена, оплата', 'Магазин, цена, оплата (~12 слов).'),
    (3, 'Услуги: аптека, банк, почта', 'Услуги: аптека, банк, почта (~8 слов).'),
    (4, 'Упаковка, количество, единицы', 'Упаковка, количество, единицы (~10 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Еда, покупки, услуги' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1),
  'Город, транспорт, путешествие', 'Тематический блок уровня A1 (~45 слов).', 1, 4, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Город, транспорт, путешествие'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Город, транспорт, путешествие' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Места в городе A1', 'Места в городе A1 (~12 слов).'),
    (1, 'Общественный транспорт', 'Общественный транспорт (~10 слов).'),
    (2, 'Направления и маршрут', 'Направления и маршрут (~10 слов).'),
    (3, 'Гостиница и аренда: минимум', 'Гостиница и аренда: минимум (~6 слов).'),
    (4, 'Багаж, билеты, поездка', 'Багаж, билеты, поездка (~7 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Город, транспорт, путешествие' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1),
  'Время, погода, календарь', 'Тематический блок уровня A1 (~45 слов).', 1, 5, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Время, погода, календарь'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Время, погода, календарь' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Дни, месяцы, сезоны', 'Дни, месяцы, сезоны (~12 слов).'),
    (1, 'Часы и расписание', 'Часы и расписание (~10 слов).'),
    (2, 'Частотность и длительность', 'Частотность и длительность (~8 слов).'),
    (3, 'Погода A1', 'Погода A1 (~8 слов).'),
    (4, 'Числа 20-100 и порядок', 'Числа 20-100 и порядок (~7 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Время, погода, календарь' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1),
  'Описание и состояния', 'Тематический блок уровня A1 (~55 слов).', 1, 6, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Описание и состояния'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Описание и состояния' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Прилагательные для людей', 'Прилагательные для людей (~12 слов).'),
    (1, 'Прилагательные для вещей', 'Прилагательные для вещей (~10 слов).'),
    (2, 'Эмоции и самочувствие', 'Эмоции и самочувствие (~10 слов).'),
    (3, 'Вкусы и предпочтения', 'Вкусы и предпочтения (~8 слов).'),
    (4, 'Противоположности и простые сравнения', 'Противоположности и простые сравнения (~15 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Описание и состояния' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A1' LIMIT 1) LIMIT 1)
);

-- ===== Level A2 =====
INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru', NULL, 'Уровень A2', 'Самостоятельное выживание: объяснить проблему, договориться, решить бытовой вопрос.', 1, 2, 'A2'
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.parent_id IS NULL AND c.name = 'Уровень A2'
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1),
  'Служебные слова и связность A2', 'Тематический блок уровня A2 (~90 слов).', 1, 0, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Служебные слова и связность A2'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Служебные слова и связность A2' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Местоименные формы и неопределённые слова', 'Местоименные формы и неопределённые слова (~18 слов).'),
    (1, 'Предлоги и выражения времени, причины, цели', 'Предлоги и выражения времени, причины, цели (~18 слов).'),
    (2, 'Союзы контраста, условия, последовательности', 'Союзы контраста, условия, последовательности (~16 слов).'),
    (3, 'Наречия вероятности, частотности, способа', 'Наречия вероятности, частотности, способа (~18 слов).'),
    (4, 'Маркеры диалога: уточнение, переспрос, реакции', 'Маркеры диалога: уточнение, переспрос, реакции (~20 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Служебные слова и связность A2' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1),
  'Быт, дом, личные дела', 'Тематический блок уровня A2 (~90 слов).', 1, 1, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Быт, дом, личные дела'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Быт, дом, личные дела' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Уборка, ремонт, бытовые предметы', 'Уборка, ремонт, бытовые предметы (~18 слов).'),
    (1, 'Одежда, уход, покупки для дома', 'Одежда, уход, покупки для дома (~18 слов).'),
    (2, 'Счета, аренда, коммуналка', 'Счета, аренда, коммуналка (~18 слов).'),
    (3, 'Документы, анкеты, записи', 'Документы, анкеты, записи (~16 слов).'),
    (4, 'Проблемы и просьбы в быту', 'Проблемы и просьбы в быту (~20 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Быт, дом, личные дела' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1),
  'Работа, учёба, цифровая жизнь', 'Тематический блок уровня A2 (~110 слов).', 1, 2, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Работа, учёба, цифровая жизнь'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Работа, учёба, цифровая жизнь' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Профессии и рабочие роли', 'Профессии и рабочие роли (~20 слов).'),
    (1, 'Офисные действия и процессы', 'Офисные действия и процессы (~20 слов).'),
    (2, 'Учёба, курсы, задания', 'Учёба, курсы, задания (~18 слов).'),
    (3, 'Компьютер, телефон, приложения', 'Компьютер, телефон, приложения (~22 слов).'),
    (4, 'Переписка, сообщения, звонки', 'Переписка, сообщения, звонки (~18 слов).'),
    (5, 'Базовые IT- и интернет-слова', 'Базовые IT- и интернет-слова (~12 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Работа, учёба, цифровая жизнь' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1),
  'Здоровье, тело, спорт', 'Тематический блок уровня A2 (~80 слов).', 1, 3, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Здоровье, тело, спорт'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Здоровье, тело, спорт' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Части тела A2', 'Части тела A2 (~16 слов).'),
    (1, 'Симптомы и простые болезни', 'Симптомы и простые болезни (~18 слов).'),
    (2, 'Аптека и врач', 'Аптека и врач (~18 слов).'),
    (3, 'Спорт и активность', 'Спорт и активность (~14 слов).'),
    (4, 'Привычки и самочувствие', 'Привычки и самочувствие (~14 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Здоровье, тело, спорт' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1),
  'Путешествия, город, жильё', 'Тематический блок уровня A2 (~100 слов).', 1, 4, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Путешествия, город, жильё'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Путешествия, город, жильё' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Транспорт A2', 'Транспорт A2 (~18 слов).'),
    (1, 'Поезд, аэропорт, билеты', 'Поезд, аэропорт, билеты (~18 слов).'),
    (2, 'Ориентирование в городе', 'Ориентирование в городе (~16 слов).'),
    (3, 'Аренда, гостиница, жильё', 'Аренда, гостиница, жильё (~18 слов).'),
    (4, 'Места и учреждения', 'Места и учреждения (~16 слов).'),
    (5, 'Безопасность, потеря, поломка', 'Безопасность, потеря, поломка (~14 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Путешествия, город, жильё' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1),
  'Общение, мнения, эмоции', 'Тематический блок уровня A2 (~90 слов).', 1, 5, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Общение, мнения, эмоции'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Общение, мнения, эмоции' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Простые мнения и оценки', 'Простые мнения и оценки (~18 слов).'),
    (1, 'Согласие, несогласие, сомнение', 'Согласие, несогласие, сомнение (~16 слов).'),
    (2, 'Эмоции и характер', 'Эмоции и характер (~18 слов).'),
    (3, 'Отношения и социальные ситуации', 'Отношения и социальные ситуации (~16 слов).'),
    (4, 'Просьбы, извинения, договорённости', 'Просьбы, извинения, договорённости (~22 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Общение, мнения, эмоции' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1),
  'Медиа, культура, досуг', 'Тематический блок уровня A2 (~80 слов).', 1, 6, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Медиа, культура, досуг'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Медиа, культура, досуг' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Фильмы, музыка, книги', 'Фильмы, музыка, книги (~16 слов).'),
    (1, 'Хобби и свободное время', 'Хобби и свободное время (~16 слов).'),
    (2, 'События, праздники, приглашения', 'События, праздники, приглашения (~18 слов).'),
    (3, 'Интернет-контент и соцсети', 'Интернет-контент и соцсети (~16 слов).'),
    (4, 'Культурные места', 'Культурные места (~14 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Медиа, культура, досуг' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1),
  'Природа, погода, животные, еда', 'Тематический блок уровня A2 (~60 слов).', 1, 7, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Природа, погода, животные, еда'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Природа, погода, животные, еда' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Природа и ландшафт', 'Природа и ландшафт (~14 слов).'),
    (1, 'Животные и растения', 'Животные и растения (~12 слов).'),
    (2, 'Погода и климат A2', 'Погода и климат A2 (~14 слов).'),
    (3, 'Еда, кухня, рецепты A2', 'Еда, кухня, рецепты A2 (~20 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Природа, погода, животные, еда' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень A2' LIMIT 1) LIMIT 1)
);

-- ===== Level B1 =====
INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru', NULL, 'Уровень B1', 'Нормальное бытовое и рабочее общение: причины, истории, административные вопросы.', 1, 3, 'B1'
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.parent_id IS NULL AND c.name = 'Уровень B1'
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Связность, грамматика речи, дискурс', 'Тематический блок уровня B1 (~140 слов).', 1, 0, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Связность, грамматика речи, дискурс'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Связность, грамматика речи, дискурс' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Вводные слова и структура мысли', 'Вводные слова и структура мысли (~25 слов).'),
    (1, 'Причина, следствие, цель, условие', 'Причина, следствие, цель, условие (~25 слов).'),
    (2, 'Уступка, контраст, исключение', 'Уступка, контраст, исключение (~20 слов).'),
    (3, 'Степень уверенности и вероятности', 'Степень уверенности и вероятности (~20 слов).'),
    (4, 'Временные отношения и последовательность', 'Временные отношения и последовательность (~20 слов).'),
    (5, 'Пересказ, уточнение, примеры', 'Пересказ, уточнение, примеры (~30 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Связность, грамматика речи, дискурс' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Работа и карьера', 'Тематический блок уровня B1 (~150 слов).', 1, 1, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Работа и карьера'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Работа и карьера' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Вакансии, резюме, собеседование', 'Вакансии, резюме, собеседование (~25 слов).'),
    (1, 'Обязанности, навыки, условия', 'Обязанности, навыки, условия (~25 слов).'),
    (2, 'Офисные процессы и коммуникация', 'Офисные процессы и коммуникация (~25 слов).'),
    (3, 'Проекты, сроки, задачи', 'Проекты, сроки, задачи (~25 слов).'),
    (4, 'Ошибки, конфликты, решения', 'Ошибки, конфликты, решения (~20 слов).'),
    (5, 'Удалённая работа и фриланс', 'Удалённая работа и фриланс (~30 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Работа и карьера' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Образование и саморазвитие', 'Тематический блок уровня B1 (~100 слов).', 1, 2, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Образование и саморазвитие'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Образование и саморазвитие' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Школа, университет, предметы', 'Школа, университет, предметы (~20 слов).'),
    (1, 'Курсы, экзамены, оценка', 'Курсы, экзамены, оценка (~20 слов).'),
    (2, 'Навыки и обучение', 'Навыки и обучение (~20 слов).'),
    (3, 'Чтение, исследование, источники', 'Чтение, исследование, источники (~20 слов).'),
    (4, 'Цели, прогресс, результаты', 'Цели, прогресс, результаты (~20 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Образование и саморазвитие' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Администрация, переезд, документы', 'Тематический блок уровня B1 (~150 слов).', 1, 3, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Администрация, переезд, документы'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Администрация, переезд, документы' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Госучреждения и процедуры', 'Госучреждения и процедуры (~25 слов).'),
    (1, 'Виза, резиденция, регистрация', 'Виза, резиденция, регистрация (~25 слов).'),
    (2, 'Банк, страховка, налоги: минимум', 'Банк, страховка, налоги: минимум (~25 слов).'),
    (3, 'Аренда, договор, права, обязанности', 'Аренда, договор, права, обязанности (~25 слов).'),
    (4, 'Письма, заявления, формы', 'Письма, заявления, формы (~25 слов).'),
    (5, 'Чрезвычайные и юридические ситуации', 'Чрезвычайные и юридические ситуации (~25 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Администрация, переезд, документы' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Здоровье и бытовые проблемы', 'Тематический блок уровня B1 (~120 слов).', 1, 4, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Здоровье и бытовые проблемы'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Здоровье и бытовые проблемы' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Врач, клиника, запись', 'Врач, клиника, запись (~25 слов).'),
    (1, 'Симптомы, травмы, боль', 'Симптомы, травмы, боль (~25 слов).'),
    (2, 'Лечение, лекарства, рекомендации', 'Лечение, лекарства, рекомендации (~25 слов).'),
    (3, 'Ремонт, поломки, сервис', 'Ремонт, поломки, сервис (~20 слов).'),
    (4, 'Бытовые риски и безопасность', 'Бытовые риски и безопасность (~25 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Здоровье и бытовые проблемы' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Отношения, характер, психология', 'Тематический блок уровня B1 (~130 слов).', 1, 5, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Отношения, характер, психология'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Отношения, характер, психология' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Характер и личные качества', 'Характер и личные качества (~25 слов).'),
    (1, 'Эмоции и состояния B1', 'Эмоции и состояния B1 (~25 слов).'),
    (2, 'Дружба, семья, отношения', 'Дружба, семья, отношения (~25 слов).'),
    (3, 'Конфликты, извинения, границы', 'Конфликты, извинения, границы (~25 слов).'),
    (4, 'Привычки, мотивация, стресс', 'Привычки, мотивация, стресс (~30 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Отношения, характер, психология' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Медиа, культура, мнения', 'Тематический блок уровня B1 (~120 слов).', 1, 6, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Медиа, культура, мнения'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Медиа, культура, мнения' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Новости и общественные темы', 'Новости и общественные темы (~25 слов).'),
    (1, 'Кино, сериалы, книги', 'Кино, сериалы, книги (~25 слов).'),
    (2, 'Музыка, искусство, события', 'Музыка, искусство, события (~20 слов).'),
    (3, 'Отзывы, рецензии, впечатления', 'Отзывы, рецензии, впечатления (~25 слов).'),
    (4, 'Аргументация вкуса', 'Аргументация вкуса (~25 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Медиа, культура, мнения' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Деньги, покупки, потребление', 'Тематический блок уровня B1 (~110 слов).', 1, 7, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Деньги, покупки, потребление'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Деньги, покупки, потребление' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Доходы, расходы, бюджет', 'Доходы, расходы, бюджет (~25 слов).'),
    (1, 'Покупки, качество, возврат', 'Покупки, качество, возврат (~25 слов).'),
    (2, 'Банк, карты, переводы', 'Банк, карты, переводы (~20 слов).'),
    (3, 'Рынок, цены, скидки', 'Рынок, цены, скидки (~20 слов).'),
    (4, 'Потребительские проблемы', 'Потребительские проблемы (~20 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Деньги, покупки, потребление' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Путешествия, город, среда', 'Тематический блок уровня B1 (~120 слов).', 1, 8, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Путешествия, город, среда'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Путешествия, город, среда' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Планирование поездки', 'Планирование поездки (~20 слов).'),
    (1, 'Транспортные проблемы', 'Транспортные проблемы (~20 слов).'),
    (2, 'Районы и инфраструктура', 'Районы и инфраструктура (~20 слов).'),
    (3, 'Природа, экология, погода B1', 'Природа, экология, погода B1 (~20 слов).'),
    (4, 'Жильё, соседи, городская жизнь', 'Жильё, соседи, городская жизнь (~20 слов).'),
    (5, 'Впечатления от мест', 'Впечатления от мест (~20 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Путешествия, город, среда' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1),
  'Технологии и интернет', 'Тематический блок уровня B1 (~160 слов).', 1, 9, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Технологии и интернет'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Технологии и интернет' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Устройства и настройки', 'Устройства и настройки (~25 слов).'),
    (1, 'Приложения, аккаунты, безопасность', 'Приложения, аккаунты, безопасность (~25 слов).'),
    (2, 'Разработка и IT-работа: базовый слой', 'Разработка и IT-работа: базовый слой (~30 слов).'),
    (3, 'Данные, файлы, сервисы', 'Данные, файлы, сервисы (~25 слов).'),
    (4, 'Ошибки, поддержка, инструкции', 'Ошибки, поддержка, инструкции (~25 слов).'),
    (5, 'Онлайн-коммуникация и контент', 'Онлайн-коммуникация и контент (~30 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Технологии и интернет' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B1' LIMIT 1) LIMIT 1)
);

-- ===== Level B2 =====
INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru', NULL, 'Уровень B2', 'Уверенная речь: аргументация, работа, общество, технологии, финансы, культура.', 1, 4, 'B2'
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.parent_id IS NULL AND c.name = 'Уровень B2'
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Связность, аргументация, модальность', 'Тематический блок уровня B2 (~250 слов).', 1, 0, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Связность, аргументация, модальность'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Связность, аргументация, модальность' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Аргументация и логические связки', 'Аргументация и логические связки (~40 слов).'),
    (1, 'Вероятность, обязанность, допущение', 'Вероятность, обязанность, допущение (~35 слов).'),
    (2, 'Причинно-следственные цепочки', 'Причинно-следственные цепочки (~35 слов).'),
    (3, 'Контраргументы и нюансирование', 'Контраргументы и нюансирование (~35 слов).'),
    (4, 'Дискуссия, дебаты, переговоры', 'Дискуссия, дебаты, переговоры (~40 слов).'),
    (5, 'Формальный и нейтральный регистр', 'Формальный и нейтральный регистр (~35 слов).'),
    (6, 'Резюмирование и выводы', 'Резюмирование и выводы (~30 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Связность, аргументация, модальность' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Абстрактные понятия и оценка', 'Тематический блок уровня B2 (~250 слов).', 1, 1, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Абстрактные понятия и оценка'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Абстрактные понятия и оценка' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Качество, эффективность, результат', 'Качество, эффективность, результат (~35 слов).'),
    (1, 'Изменение, развитие, тенденции', 'Изменение, развитие, тенденции (~35 слов).'),
    (2, 'Проблема, риск, решение', 'Проблема, риск, решение (~35 слов).'),
    (3, 'Ценности, мораль, ответственность', 'Ценности, мораль, ответственность (~35 слов).'),
    (4, 'Восприятие, интерпретация, отношение', 'Восприятие, интерпретация, отношение (~35 слов).'),
    (5, 'Сравнение, приоритеты, компромиссы', 'Сравнение, приоритеты, компромиссы (~35 слов).'),
    (6, 'Общеупотребительные метафорические слова', 'Общеупотребительные метафорические слова (~40 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Абстрактные понятия и оценка' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Работа, бизнес, управление', 'Тематический блок уровня B2 (~300 слов).', 1, 2, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Работа, бизнес, управление'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Работа, бизнес, управление' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Управление проектами', 'Управление проектами (~45 слов).'),
    (1, 'Команда и лидерство', 'Команда и лидерство (~40 слов).'),
    (2, 'Стратегия, цели, KPI', 'Стратегия, цели, KPI (~40 слов).'),
    (3, 'Продажи, клиенты, рынок', 'Продажи, клиенты, рынок (~40 слов).'),
    (4, 'Договорные и юридические рабочие слова', 'Договорные и юридические рабочие слова (~35 слов).'),
    (5, 'HR, performance, оценка', 'HR, performance, оценка (~35 слов).'),
    (6, 'Конфликты, переговоры, решения', 'Конфликты, переговоры, решения (~35 слов).'),
    (7, 'Удалёнка и международная работа', 'Удалёнка и международная работа (~30 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Работа, бизнес, управление' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Технологии, наука, ИИ', 'Тематический блок уровня B2 (~300 слов).', 1, 3, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Технологии, наука, ИИ'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Технологии, наука, ИИ' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Разработка ПО и архитектура', 'Разработка ПО и архитектура (~45 слов).'),
    (1, 'Данные, аналитика, ML, AI', 'Данные, аналитика, ML, AI (~45 слов).'),
    (2, 'Безопасность, приватность, риски', 'Безопасность, приватность, риски (~35 слов).'),
    (3, 'Инфраструктура, cloud, DevOps', 'Инфраструктура, cloud, DevOps (~40 слов).'),
    (4, 'Наука: методы, исследования, гипотезы', 'Наука: методы, исследования, гипотезы (~40 слов).'),
    (5, 'Продукты, UX, метрики', 'Продукты, UX, метрики (~35 слов).'),
    (6, 'Ошибки, инциденты, качество', 'Ошибки, инциденты, качество (~35 слов).'),
    (7, 'Цифровое общество', 'Цифровое общество (~25 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Технологии, наука, ИИ' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Политика, общество, право', 'Тематический блок уровня B2 (~300 слов).', 1, 4, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Политика, общество, право'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Политика, общество, право' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Государство и институты', 'Государство и институты (~40 слов).'),
    (1, 'Выборы, партии, общественное мнение', 'Выборы, партии, общественное мнение (~35 слов).'),
    (2, 'Право, суд, договоры', 'Право, суд, договоры (~40 слов).'),
    (3, 'Миграция, гражданство, статус', 'Миграция, гражданство, статус (~35 слов).'),
    (4, 'Социальные группы и неравенство', 'Социальные группы и неравенство (~35 слов).'),
    (5, 'Преступность и безопасность', 'Преступность и безопасность (~35 слов).'),
    (6, 'Международные отношения', 'Международные отношения (~40 слов).'),
    (7, 'Общественные услуги и бюрократия', 'Общественные услуги и бюрократия (~40 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Политика, общество, право' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Экономика, финансы, недвижимость', 'Тематический блок уровня B2 (~250 слов).', 1, 5, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Экономика, финансы, недвижимость'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Экономика, финансы, недвижимость' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Макроэкономика: минимум', 'Макроэкономика: минимум (~35 слов).'),
    (1, 'Личные финансы и инвестиции', 'Личные финансы и инвестиции (~35 слов).'),
    (2, 'Недвижимость, ипотека, аренда', 'Недвижимость, ипотека, аренда (~35 слов).'),
    (3, 'Налоги, зарплата, соцвзносы', 'Налоги, зарплата, соцвзносы (~35 слов).'),
    (4, 'Банки, кредиты, страхование', 'Банки, кредиты, страхование (~35 слов).'),
    (5, 'Бизнес-модели и ценообразование', 'Бизнес-модели и ценообразование (~35 слов).'),
    (6, 'Кризисы, инфляция, рынки', 'Кризисы, инфляция, рынки (~40 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Экономика, финансы, недвижимость' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Культура, медиа, литература', 'Тематический блок уровня B2 (~250 слов).', 1, 6, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Культура, медиа, литература'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Культура, медиа, литература' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Жанры и формы искусства', 'Жанры и формы искусства (~35 слов).'),
    (1, 'Кино, литература, критика', 'Кино, литература, критика (~35 слов).'),
    (2, 'Журналистика и медиаформаты', 'Журналистика и медиаформаты (~35 слов).'),
    (3, 'История и культурная память', 'История и культурная память (~35 слов).'),
    (4, 'Традиции, праздники, идентичность', 'Традиции, праздники, идентичность (~30 слов).'),
    (5, 'Рецензии и интерпретация', 'Рецензии и интерпретация (~40 слов).'),
    (6, 'Культурные индустрии', 'Культурные индустрии (~40 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Культура, медиа, литература' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Психология, отношения, социальное взаимодействие', 'Тематический блок уровня B2 (~200 слов).', 1, 7, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Психология, отношения, социальное взаимодействие'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Психология, отношения, социальное взаимодействие' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Эмоции и внутренние состояния B2', 'Эмоции и внутренние состояния B2 (~35 слов).'),
    (1, 'Коммуникация и границы', 'Коммуникация и границы (~30 слов).'),
    (2, 'Конфликты, манипуляции, доверие', 'Конфликты, манипуляции, доверие (~35 слов).'),
    (3, 'Семья, партнёрство, сообщество', 'Семья, партнёрство, сообщество (~30 слов).'),
    (4, 'Личность, привычки, поведение', 'Личность, привычки, поведение (~35 слов).'),
    (5, 'Стресс, выгорание, баланс', 'Стресс, выгорание, баланс (~35 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Психология, отношения, социальное взаимодействие' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Здоровье, медицина, образ жизни', 'Тематический блок уровня B2 (~200 слов).', 1, 8, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Здоровье, медицина, образ жизни'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Здоровье, медицина, образ жизни' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Система здравоохранения', 'Система здравоохранения (~35 слов).'),
    (1, 'Диагностика, лечение, обследования', 'Диагностика, лечение, обследования (~35 слов).'),
    (2, 'Хронические состояния и профилактика', 'Хронические состояния и профилактика (~30 слов).'),
    (3, 'Питание, сон, спорт', 'Питание, сон, спорт (~30 слов).'),
    (4, 'Медицина и технологии', 'Медицина и технологии (~35 слов).'),
    (5, 'Риски, побочные эффекты, инструкции', 'Риски, побочные эффекты, инструкции (~35 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Здоровье, медицина, образ жизни' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1),
  'Экология, география, урбанистика', 'Тематический блок уровня B2 (~200 слов).', 1, 9, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Экология, география, урбанистика'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Экология, география, урбанистика' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Климат и экологические процессы', 'Климат и экологические процессы (~35 слов).'),
    (1, 'Городское планирование', 'Городское планирование (~35 слов).'),
    (2, 'Транспорт, энергия, инфраструктура', 'Транспорт, энергия, инфраструктура (~35 слов).'),
    (3, 'География, регионы, ландшафты', 'География, регионы, ландшафты (~30 слов).'),
    (4, 'Жильё, районы, качество среды', 'Жильё, районы, качество среды (~30 слов).'),
    (5, 'Катастрофы, устойчивость, ресурсы', 'Катастрофы, устойчивость, ресурсы (~35 слов).')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Экология, география, урбанистика' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень B2' LIMIT 1) LIMIT 1)
);

-- ===== Level C1 =====
INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru', NULL, 'Уровень C1', 'Точные оттенки, регистры, абстрактная лексика и доменные ветки терминологии.', 1, 5, 'C1'
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.parent_id IS NULL AND c.name = 'Уровень C1'
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Дискурс, регистр, прагматика', 'Тематический блок уровня C1 (~500 слов).', 1, 0, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Дискурс, регистр, прагматика'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Дискурс, регистр, прагматика' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Академические связки и метатекст (1-30)', 'Академические связки и метатекст: слова 1-30 из ~70 в этом наборе.'),
    (1, 'Академические связки и метатекст (31-60)', 'Академические связки и метатекст: слова 31-60 из ~70 в этом наборе.'),
    (2, 'Академические связки и метатекст (61-70)', 'Академические связки и метатекст: слова 61-70 из ~70 в этом наборе.'),
    (3, 'Риторические маркеры и оттенки позиции (1-30)', 'Риторические маркеры и оттенки позиции: слова 1-30 из ~70 в этом наборе.'),
    (4, 'Риторические маркеры и оттенки позиции (31-60)', 'Риторические маркеры и оттенки позиции: слова 31-60 из ~70 в этом наборе.'),
    (5, 'Риторические маркеры и оттенки позиции (61-70)', 'Риторические маркеры и оттенки позиции: слова 61-70 из ~70 в этом наборе.'),
    (6, 'Вежливость, имплицитность, дистанция (1-30)', 'Вежливость, имплицитность, дистанция: слова 1-30 из ~60 в этом наборе.'),
    (7, 'Вежливость, имплицитность, дистанция (31-60)', 'Вежливость, имплицитность, дистанция: слова 31-60 из ~60 в этом наборе.'),
    (8, 'Ирония, сарказм, эвфемизмы (1-30)', 'Ирония, сарказм, эвфемизмы: слова 1-30 из ~60 в этом наборе.'),
    (9, 'Ирония, сарказм, эвфемизмы (31-60)', 'Ирония, сарказм, эвфемизмы: слова 31-60 из ~60 в этом наборе.'),
    (10, 'Переговорные формулировки (1-30)', 'Переговорные формулировки: слова 1-30 из ~60 в этом наборе.'),
    (11, 'Переговорные формулировки (31-60)', 'Переговорные формулировки: слова 31-60 из ~60 в этом наборе.'),
    (12, 'Юридически и административно нейтральный регистр (1-30)', 'Юридически и административно нейтральный регистр: слова 1-30 из ~60 в этом наборе.'),
    (13, 'Юридически и административно нейтральный регистр (31-60)', 'Юридически и административно нейтральный регистр: слова 31-60 из ~60 в этом наборе.'),
    (14, 'Разговорные устойчивые реакции (1-30)', 'Разговорные устойчивые реакции: слова 1-30 из ~60 в этом наборе.'),
    (15, 'Разговорные устойчивые реакции (31-60)', 'Разговорные устойчивые реакции: слова 31-60 из ~60 в этом наборе.'),
    (16, 'Резюмирование, переформулирование, оговорки (1-30)', 'Резюмирование, переформулирование, оговорки: слова 1-30 из ~60 в этом наборе.'),
    (17, 'Резюмирование, переформулирование, оговорки (31-60)', 'Резюмирование, переформулирование, оговорки: слова 31-60 из ~60 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Дискурс, регистр, прагматика' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Абстрактные концепты и точные оттенки', 'Тематический блок уровня C1 (~600 слов).', 1, 1, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Абстрактные концепты и точные оттенки'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Абстрактные концепты и точные оттенки' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Причинность и сложные зависимости (1-30)', 'Причинность и сложные зависимости: слова 1-30 из ~75 в этом наборе.'),
    (1, 'Причинность и сложные зависимости (31-60)', 'Причинность и сложные зависимости: слова 31-60 из ~75 в этом наборе.'),
    (2, 'Причинность и сложные зависимости (61-75)', 'Причинность и сложные зависимости: слова 61-75 из ~75 в этом наборе.'),
    (3, 'Власть, контроль, влияние (1-30)', 'Власть, контроль, влияние: слова 1-30 из ~75 в этом наборе.'),
    (4, 'Власть, контроль, влияние (31-60)', 'Власть, контроль, влияние: слова 31-60 из ~75 в этом наборе.'),
    (5, 'Власть, контроль, влияние (61-75)', 'Власть, контроль, влияние: слова 61-75 из ~75 в этом наборе.'),
    (6, 'Идентичность, принадлежность, память (1-30)', 'Идентичность, принадлежность, память: слова 1-30 из ~75 в этом наборе.'),
    (7, 'Идентичность, принадлежность, память (31-60)', 'Идентичность, принадлежность, память: слова 31-60 из ~75 в этом наборе.'),
    (8, 'Идентичность, принадлежность, память (61-75)', 'Идентичность, принадлежность, память: слова 61-75 из ~75 в этом наборе.'),
    (9, 'Нормы, отклонения, легитимность (1-30)', 'Нормы, отклонения, легитимность: слова 1-30 из ~75 в этом наборе.'),
    (10, 'Нормы, отклонения, легитимность (31-60)', 'Нормы, отклонения, легитимность: слова 31-60 из ~75 в этом наборе.'),
    (11, 'Нормы, отклонения, легитимность (61-75)', 'Нормы, отклонения, легитимность: слова 61-75 из ~75 в этом наборе.'),
    (12, 'Качество, достоверность, надёжность (1-30)', 'Качество, достоверность, надёжность: слова 1-30 из ~75 в этом наборе.'),
    (13, 'Качество, достоверность, надёжность (31-60)', 'Качество, достоверность, надёжность: слова 31-60 из ~75 в этом наборе.'),
    (14, 'Качество, достоверность, надёжность (61-75)', 'Качество, достоверность, надёжность: слова 61-75 из ~75 в этом наборе.'),
    (15, 'Неопределённость, вероятность, сомнение (1-30)', 'Неопределённость, вероятность, сомнение: слова 1-30 из ~75 в этом наборе.'),
    (16, 'Неопределённость, вероятность, сомнение (31-60)', 'Неопределённость, вероятность, сомнение: слова 31-60 из ~75 в этом наборе.'),
    (17, 'Неопределённость, вероятность, сомнение (61-75)', 'Неопределённость, вероятность, сомнение: слова 61-75 из ~75 в этом наборе.'),
    (18, 'Изменения, кризисы, трансформации (1-30)', 'Изменения, кризисы, трансформации: слова 1-30 из ~75 в этом наборе.'),
    (19, 'Изменения, кризисы, трансформации (31-60)', 'Изменения, кризисы, трансформации: слова 31-60 из ~75 в этом наборе.'),
    (20, 'Изменения, кризисы, трансформации (61-75)', 'Изменения, кризисы, трансформации: слова 61-75 из ~75 в этом наборе.'),
    (21, 'Эмоционально-оценочные оттенки (1-30)', 'Эмоционально-оценочные оттенки: слова 1-30 из ~75 в этом наборе.'),
    (22, 'Эмоционально-оценочные оттенки (31-60)', 'Эмоционально-оценочные оттенки: слова 31-60 из ~75 в этом наборе.'),
    (23, 'Эмоционально-оценочные оттенки (61-75)', 'Эмоционально-оценочные оттенки: слова 61-75 из ~75 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Абстрактные концепты и точные оттенки' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Политика, право, государство', 'Тематический блок уровня C1 (~600 слов).', 1, 2, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Политика, право, государство'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Политика, право, государство' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Конституционная система и институты (1-30)', 'Конституционная система и институты: слова 1-30 из ~75 в этом наборе.'),
    (1, 'Конституционная система и институты (31-60)', 'Конституционная система и институты: слова 31-60 из ~75 в этом наборе.'),
    (2, 'Конституционная система и институты (61-75)', 'Конституционная система и институты: слова 61-75 из ~75 в этом наборе.'),
    (3, 'Законодательство, суд, процесс (1-30)', 'Законодательство, суд, процесс: слова 1-30 из ~75 в этом наборе.'),
    (4, 'Законодательство, суд, процесс (31-60)', 'Законодательство, суд, процесс: слова 31-60 из ~75 в этом наборе.'),
    (5, 'Законодательство, суд, процесс (61-75)', 'Законодательство, суд, процесс: слова 61-75 из ~75 в этом наборе.'),
    (6, 'Административное право и бюрократия (1-30)', 'Административное право и бюрократия: слова 1-30 из ~75 в этом наборе.'),
    (7, 'Административное право и бюрократия (31-60)', 'Административное право и бюрократия: слова 31-60 из ~75 в этом наборе.'),
    (8, 'Административное право и бюрократия (61-75)', 'Административное право и бюрократия: слова 61-75 из ~75 в этом наборе.'),
    (9, 'Международная политика и дипломатия (1-30)', 'Международная политика и дипломатия: слова 1-30 из ~75 в этом наборе.'),
    (10, 'Международная политика и дипломатия (31-60)', 'Международная политика и дипломатия: слова 31-60 из ~75 в этом наборе.'),
    (11, 'Международная политика и дипломатия (61-75)', 'Международная политика и дипломатия: слова 61-75 из ~75 в этом наборе.'),
    (12, 'Партии, идеологии, кампании (1-30)', 'Партии, идеологии, кампании: слова 1-30 из ~75 в этом наборе.'),
    (13, 'Партии, идеологии, кампании (31-60)', 'Партии, идеологии, кампании: слова 31-60 из ~75 в этом наборе.'),
    (14, 'Партии, идеологии, кампании (61-75)', 'Партии, идеологии, кампании: слова 61-75 из ~75 в этом наборе.'),
    (15, 'Права человека и социальная политика (1-30)', 'Права человека и социальная политика: слова 1-30 из ~75 в этом наборе.'),
    (16, 'Права человека и социальная политика (31-60)', 'Права человека и социальная политика: слова 31-60 из ~75 в этом наборе.'),
    (17, 'Права человека и социальная политика (61-75)', 'Права человека и социальная политика: слова 61-75 из ~75 в этом наборе.'),
    (18, 'Преступления, ответственность, санкции (1-30)', 'Преступления, ответственность, санкции: слова 1-30 из ~75 в этом наборе.'),
    (19, 'Преступления, ответственность, санкции (31-60)', 'Преступления, ответственность, санкции: слова 31-60 из ~75 в этом наборе.'),
    (20, 'Преступления, ответственность, санкции (61-75)', 'Преступления, ответственность, санкции: слова 61-75 из ~75 в этом наборе.'),
    (21, 'Публичное управление и реформы (1-30)', 'Публичное управление и реформы: слова 1-30 из ~75 в этом наборе.'),
    (22, 'Публичное управление и реформы (31-60)', 'Публичное управление и реформы: слова 31-60 из ~75 в этом наборе.'),
    (23, 'Публичное управление и реформы (61-75)', 'Публичное управление и реформы: слова 61-75 из ~75 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Политика, право, государство' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Экономика, бизнес, рынки', 'Тематический блок уровня C1 (~550 слов).', 1, 3, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Экономика, бизнес, рынки'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Экономика, бизнес, рынки' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Макроэкономика и финансы (1-30)', 'Макроэкономика и финансы: слова 1-30 из ~70 в этом наборе.'),
    (1, 'Макроэкономика и финансы (31-60)', 'Макроэкономика и финансы: слова 31-60 из ~70 в этом наборе.'),
    (2, 'Макроэкономика и финансы (61-70)', 'Макроэкономика и финансы: слова 61-70 из ~70 в этом наборе.'),
    (3, 'Рынки, конкуренция, регулирование (1-30)', 'Рынки, конкуренция, регулирование: слова 1-30 из ~70 в этом наборе.'),
    (4, 'Рынки, конкуренция, регулирование (31-60)', 'Рынки, конкуренция, регулирование: слова 31-60 из ~70 в этом наборе.'),
    (5, 'Рынки, конкуренция, регулирование (61-70)', 'Рынки, конкуренция, регулирование: слова 61-70 из ~70 в этом наборе.'),
    (6, 'Корпоративное управление (1-30)', 'Корпоративное управление: слова 1-30 из ~70 в этом наборе.'),
    (7, 'Корпоративное управление (31-60)', 'Корпоративное управление: слова 31-60 из ~70 в этом наборе.'),
    (8, 'Корпоративное управление (61-70)', 'Корпоративное управление: слова 61-70 из ~70 в этом наборе.'),
    (9, 'Бухгалтерия, налоги, отчётность (1-30)', 'Бухгалтерия, налоги, отчётность: слова 1-30 из ~65 в этом наборе.'),
    (10, 'Бухгалтерия, налоги, отчётность (31-60)', 'Бухгалтерия, налоги, отчётность: слова 31-60 из ~65 в этом наборе.'),
    (11, 'Бухгалтерия, налоги, отчётность (61-65)', 'Бухгалтерия, налоги, отчётность: слова 61-65 из ~65 в этом наборе.'),
    (12, 'Инвестиции, риски, активы (1-30)', 'Инвестиции, риски, активы: слова 1-30 из ~70 в этом наборе.'),
    (13, 'Инвестиции, риски, активы (31-60)', 'Инвестиции, риски, активы: слова 31-60 из ~70 в этом наборе.'),
    (14, 'Инвестиции, риски, активы (61-70)', 'Инвестиции, риски, активы: слова 61-70 из ~70 в этом наборе.'),
    (15, 'Недвижимость, ипотека, оценка (1-30)', 'Недвижимость, ипотека, оценка: слова 1-30 из ~65 в этом наборе.'),
    (16, 'Недвижимость, ипотека, оценка (31-60)', 'Недвижимость, ипотека, оценка: слова 31-60 из ~65 в этом наборе.'),
    (17, 'Недвижимость, ипотека, оценка (61-65)', 'Недвижимость, ипотека, оценка: слова 61-65 из ~65 в этом наборе.'),
    (18, 'Трудовой рынок, зарплаты, соцгарантии (1-30)', 'Трудовой рынок, зарплаты, соцгарантии: слова 1-30 из ~70 в этом наборе.'),
    (19, 'Трудовой рынок, зарплаты, соцгарантии (31-60)', 'Трудовой рынок, зарплаты, соцгарантии: слова 31-60 из ~70 в этом наборе.'),
    (20, 'Трудовой рынок, зарплаты, соцгарантии (61-70)', 'Трудовой рынок, зарплаты, соцгарантии: слова 61-70 из ~70 в этом наборе.'),
    (21, 'Кризисы, долг, инфляция (1-30)', 'Кризисы, долг, инфляция: слова 1-30 из ~70 в этом наборе.'),
    (22, 'Кризисы, долг, инфляция (31-60)', 'Кризисы, долг, инфляция: слова 31-60 из ~70 в этом наборе.'),
    (23, 'Кризисы, долг, инфляция (61-70)', 'Кризисы, долг, инфляция: слова 61-70 из ~70 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Экономика, бизнес, рынки' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Наука, технологии, инженерия', 'Тематический блок уровня C1 (~650 слов).', 1, 4, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Наука, технологии, инженерия'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Наука, технологии, инженерия' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Методология исследований (1-30)', 'Методология исследований: слова 1-30 из ~80 в этом наборе.'),
    (1, 'Методология исследований (31-60)', 'Методология исследований: слова 31-60 из ~80 в этом наборе.'),
    (2, 'Методология исследований (61-80)', 'Методология исследований: слова 61-80 из ~80 в этом наборе.'),
    (3, 'Статистика, данные, моделирование (1-30)', 'Статистика, данные, моделирование: слова 1-30 из ~80 в этом наборе.'),
    (4, 'Статистика, данные, моделирование (31-60)', 'Статистика, данные, моделирование: слова 31-60 из ~80 в этом наборе.'),
    (5, 'Статистика, данные, моделирование (61-80)', 'Статистика, данные, моделирование: слова 61-80 из ~80 в этом наборе.'),
    (6, 'Software engineering advanced (1-30)', 'Software engineering advanced: слова 1-30 из ~85 в этом наборе.'),
    (7, 'Software engineering advanced (31-60)', 'Software engineering advanced: слова 31-60 из ~85 в этом наборе.'),
    (8, 'Software engineering advanced (61-85)', 'Software engineering advanced: слова 61-85 из ~85 в этом наборе.'),
    (9, 'Cybersecurity, privacy, compliance (1-30)', 'Cybersecurity, privacy, compliance: слова 1-30 из ~75 в этом наборе.'),
    (10, 'Cybersecurity, privacy, compliance (31-60)', 'Cybersecurity, privacy, compliance: слова 31-60 из ~75 в этом наборе.'),
    (11, 'Cybersecurity, privacy, compliance (61-75)', 'Cybersecurity, privacy, compliance: слова 61-75 из ~75 в этом наборе.'),
    (12, 'AI, ML, NLP, data science (1-30)', 'AI, ML, NLP, data science: слова 1-30 из ~85 в этом наборе.'),
    (13, 'AI, ML, NLP, data science (31-60)', 'AI, ML, NLP, data science: слова 31-60 из ~85 в этом наборе.'),
    (14, 'AI, ML, NLP, data science (61-85)', 'AI, ML, NLP, data science: слова 61-85 из ~85 в этом наборе.'),
    (15, 'Инженерия, производство, материалы (1-30)', 'Инженерия, производство, материалы: слова 1-30 из ~75 в этом наборе.'),
    (16, 'Инженерия, производство, материалы (31-60)', 'Инженерия, производство, материалы: слова 31-60 из ~75 в этом наборе.'),
    (17, 'Инженерия, производство, материалы (61-75)', 'Инженерия, производство, материалы: слова 61-75 из ~75 в этом наборе.'),
    (18, 'Энергия, транспорт, инфраструктура (1-30)', 'Энергия, транспорт, инфраструктура: слова 1-30 из ~80 в этом наборе.'),
    (19, 'Энергия, транспорт, инфраструктура (31-60)', 'Энергия, транспорт, инфраструктура: слова 31-60 из ~80 в этом наборе.'),
    (20, 'Энергия, транспорт, инфраструктура (61-80)', 'Энергия, транспорт, инфраструктура: слова 61-80 из ~80 в этом наборе.'),
    (21, 'Научная коммуникация и публикации (1-30)', 'Научная коммуникация и публикации: слова 1-30 из ~90 в этом наборе.'),
    (22, 'Научная коммуникация и публикации (31-60)', 'Научная коммуникация и публикации: слова 31-60 из ~90 в этом наборе.'),
    (23, 'Научная коммуникация и публикации (61-90)', 'Научная коммуникация и публикации: слова 61-90 из ~90 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Наука, технологии, инженерия' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Культура, искусство, гуманитарные темы', 'Тематический блок уровня C1 (~550 слов).', 1, 5, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Культура, искусство, гуманитарные темы'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Культура, искусство, гуманитарные темы' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Литература и стилистика (1-30)', 'Литература и стилистика: слова 1-30 из ~70 в этом наборе.'),
    (1, 'Литература и стилистика (31-60)', 'Литература и стилистика: слова 31-60 из ~70 в этом наборе.'),
    (2, 'Литература и стилистика (61-70)', 'Литература и стилистика: слова 61-70 из ~70 в этом наборе.'),
    (3, 'Кино, театр, визуальное искусство (1-30)', 'Кино, театр, визуальное искусство: слова 1-30 из ~70 в этом наборе.'),
    (4, 'Кино, театр, визуальное искусство (31-60)', 'Кино, театр, визуальное искусство: слова 31-60 из ~70 в этом наборе.'),
    (5, 'Кино, театр, визуальное искусство (61-70)', 'Кино, театр, визуальное искусство: слова 61-70 из ~70 в этом наборе.'),
    (6, 'Музыка, архитектура, дизайн (1-30)', 'Музыка, архитектура, дизайн: слова 1-30 из ~65 в этом наборе.'),
    (7, 'Музыка, архитектура, дизайн (31-60)', 'Музыка, архитектура, дизайн: слова 31-60 из ~65 в этом наборе.'),
    (8, 'Музыка, архитектура, дизайн (61-65)', 'Музыка, архитектура, дизайн: слова 61-65 из ~65 в этом наборе.'),
    (9, 'История и историография (1-30)', 'История и историография: слова 1-30 из ~70 в этом наборе.'),
    (10, 'История и историография (31-60)', 'История и историография: слова 31-60 из ~70 в этом наборе.'),
    (11, 'История и историография (61-70)', 'История и историография: слова 61-70 из ~70 в этом наборе.'),
    (12, 'Философия, этика, религии как культурная тема (1-30)', 'Философия, этика, религии как культурная тема: слова 1-30 из ~65 в этом наборе.'),
    (13, 'Философия, этика, религии как культурная тема (31-60)', 'Философия, этика, религии как культурная тема: слова 31-60 из ~65 в этом наборе.'),
    (14, 'Философия, этика, религии как культурная тема (61-65)', 'Философия, этика, религии как культурная тема: слова 61-65 из ~65 в этом наборе.'),
    (15, 'Языкознание, этимология, перевод (1-30)', 'Языкознание, этимология, перевод: слова 1-30 из ~70 в этом наборе.'),
    (16, 'Языкознание, этимология, перевод (31-60)', 'Языкознание, этимология, перевод: слова 31-60 из ~70 в этом наборе.'),
    (17, 'Языкознание, этимология, перевод (61-70)', 'Языкознание, этимология, перевод: слова 61-70 из ~70 в этом наборе.'),
    (18, 'Культурные конфликты и память (1-30)', 'Культурные конфликты и память: слова 1-30 из ~70 в этом наборе.'),
    (19, 'Культурные конфликты и память (31-60)', 'Культурные конфликты и память: слова 31-60 из ~70 в этом наборе.'),
    (20, 'Культурные конфликты и память (61-70)', 'Культурные конфликты и память: слова 61-70 из ~70 в этом наборе.'),
    (21, 'Критика и эссеистика (1-30)', 'Критика и эссеистика: слова 1-30 из ~70 в этом наборе.'),
    (22, 'Критика и эссеистика (31-60)', 'Критика и эссеистика: слова 31-60 из ~70 в этом наборе.'),
    (23, 'Критика и эссеистика (61-70)', 'Критика и эссеистика: слова 61-70 из ~70 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Культура, искусство, гуманитарные темы' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Медиа, риторика, общественная дискуссия', 'Тематический блок уровня C1 (~450 слов).', 1, 6, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Медиа, риторика, общественная дискуссия'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Медиа, риторика, общественная дискуссия' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Политическая риторика (1-30)', 'Политическая риторика: слова 1-30 из ~60 в этом наборе.'),
    (1, 'Политическая риторика (31-60)', 'Политическая риторика: слова 31-60 из ~60 в этом наборе.'),
    (2, 'Пропаганда, манипуляция, фрейминг (1-30)', 'Пропаганда, манипуляция, фрейминг: слова 1-30 из ~60 в этом наборе.'),
    (3, 'Пропаганда, манипуляция, фрейминг (31-60)', 'Пропаганда, манипуляция, фрейминг: слова 31-60 из ~60 в этом наборе.'),
    (4, 'Журналистские жанры (1-30)', 'Журналистские жанры: слова 1-30 из ~55 в этом наборе.'),
    (5, 'Журналистские жанры (31-55)', 'Журналистские жанры: слова 31-55 из ~55 в этом наборе.'),
    (6, 'Соцсети, инфлюенсеры, платформы (1-30)', 'Соцсети, инфлюенсеры, платформы: слова 1-30 из ~55 в этом наборе.'),
    (7, 'Соцсети, инфлюенсеры, платформы (31-55)', 'Соцсети, инфлюенсеры, платформы: слова 31-55 из ~55 в этом наборе.'),
    (8, 'Дезинформация и fact-checking (1-30)', 'Дезинформация и fact-checking: слова 1-30 из ~60 в этом наборе.'),
    (9, 'Дезинформация и fact-checking (31-60)', 'Дезинформация и fact-checking: слова 31-60 из ~60 в этом наборе.'),
    (10, 'Репутация, скандалы, кризисные коммуникации (1-30)', 'Репутация, скандалы, кризисные коммуникации: слова 1-30 из ~55 в этом наборе.'),
    (11, 'Репутация, скандалы, кризисные коммуникации (31-55)', 'Репутация, скандалы, кризисные коммуникации: слова 31-55 из ~55 в этом наборе.'),
    (12, 'Публичные выступления и презентации (1-30)', 'Публичные выступления и презентации: слова 1-30 из ~55 в этом наборе.'),
    (13, 'Публичные выступления и презентации (31-55)', 'Публичные выступления и презентации: слова 31-55 из ~55 в этом наборе.'),
    (14, 'Метрики медиа и аудитория (1-30)', 'Метрики медиа и аудитория: слова 1-30 из ~50 в этом наборе.'),
    (15, 'Метрики медиа и аудитория (31-50)', 'Метрики медиа и аудитория: слова 31-50 из ~50 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Медиа, риторика, общественная дискуссия' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Медицина, психология, биология', 'Тематический блок уровня C1 (~450 слов).', 1, 7, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Медицина, психология, биология'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Медицина, психология, биология' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Анатомия и физиология advanced (1-30)', 'Анатомия и физиология advanced: слова 1-30 из ~60 в этом наборе.'),
    (1, 'Анатомия и физиология advanced (31-60)', 'Анатомия и физиология advanced: слова 31-60 из ~60 в этом наборе.'),
    (2, 'Болезни и симптомы: нюансы (1-30)', 'Болезни и симптомы: нюансы: слова 1-30 из ~60 в этом наборе.'),
    (3, 'Болезни и симптомы: нюансы (31-60)', 'Болезни и симптомы: нюансы: слова 31-60 из ~60 в этом наборе.'),
    (4, 'Диагностика, анализы, процедуры (1-30)', 'Диагностика, анализы, процедуры: слова 1-30 из ~60 в этом наборе.'),
    (5, 'Диагностика, анализы, процедуры (31-60)', 'Диагностика, анализы, процедуры: слова 31-60 из ~60 в этом наборе.'),
    (6, 'Лекарства, побочные эффекты, инструкции (1-30)', 'Лекарства, побочные эффекты, инструкции: слова 1-30 из ~55 в этом наборе.'),
    (7, 'Лекарства, побочные эффекты, инструкции (31-55)', 'Лекарства, побочные эффекты, инструкции: слова 31-55 из ~55 в этом наборе.'),
    (8, 'Психология и психотерапия (1-30)', 'Психология и психотерапия: слова 1-30 из ~60 в этом наборе.'),
    (9, 'Психология и психотерапия (31-60)', 'Психология и психотерапия: слова 31-60 из ~60 в этом наборе.'),
    (10, 'Общественное здоровье (1-30)', 'Общественное здоровье: слова 1-30 из ~55 в этом наборе.'),
    (11, 'Общественное здоровье (31-55)', 'Общественное здоровье: слова 31-55 из ~55 в этом наборе.'),
    (12, 'Биология, генетика, эволюция (1-30)', 'Биология, генетика, эволюция: слова 1-30 из ~50 в этом наборе.'),
    (13, 'Биология, генетика, эволюция (31-50)', 'Биология, генетика, эволюция: слова 31-50 из ~50 в этом наборе.'),
    (14, 'Этика медицины (1-30)', 'Этика медицины: слова 1-30 из ~50 в этом наборе.'),
    (15, 'Этика медицины (31-50)', 'Этика медицины: слова 31-50 из ~50 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Медицина, психология, биология' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Экология, география, город, инфраструктура', 'Тематический блок уровня C1 (~400 слов).', 1, 8, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Экология, география, город, инфраструктура'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Экология, география, город, инфраструктура' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Климатология и энергетика (1-30)', 'Климатология и энергетика: слова 1-30 из ~55 в этом наборе.'),
    (1, 'Климатология и энергетика (31-55)', 'Климатология и энергетика: слова 31-55 из ~55 в этом наборе.'),
    (2, 'Экосистемы и биоразнообразие (1-30)', 'Экосистемы и биоразнообразие: слова 1-30 из ~50 в этом наборе.'),
    (3, 'Экосистемы и биоразнообразие (31-50)', 'Экосистемы и биоразнообразие: слова 31-50 из ~50 в этом наборе.'),
    (4, 'Ресурсы, отходы, потребление (1-30)', 'Ресурсы, отходы, потребление: слова 1-30 из ~50 в этом наборе.'),
    (5, 'Ресурсы, отходы, потребление (31-50)', 'Ресурсы, отходы, потребление: слова 31-50 из ~50 в этом наборе.'),
    (6, 'Урбанистика, жильё, районы (1-30)', 'Урбанистика, жильё, районы: слова 1-30 из ~55 в этом наборе.'),
    (7, 'Урбанистика, жильё, районы (31-55)', 'Урбанистика, жильё, районы: слова 31-55 из ~55 в этом наборе.'),
    (8, 'Транспорт, инфраструктура, логистика (1-30)', 'Транспорт, инфраструктура, логистика: слова 1-30 из ~55 в этом наборе.'),
    (9, 'Транспорт, инфраструктура, логистика (31-55)', 'Транспорт, инфраструктура, логистика: слова 31-55 из ~55 в этом наборе.'),
    (10, 'География Испании, Латинской Америки и мира (1-30)', 'География Испании, Латинской Америки и мира: слова 1-30 из ~50 в этом наборе.'),
    (11, 'География Испании, Латинской Америки и мира (31-50)', 'География Испании, Латинской Америки и мира: слова 31-50 из ~50 в этом наборе.'),
    (12, 'Риски, катастрофы, адаптация (1-30)', 'Риски, катастрофы, адаптация: слова 1-30 из ~45 в этом наборе.'),
    (13, 'Риски, катастрофы, адаптация (31-45)', 'Риски, катастрофы, адаптация: слова 31-45 из ~45 в этом наборе.'),
    (14, 'Сельское хозяйство, пища, земля (1-30)', 'Сельское хозяйство, пища, земля: слова 1-30 из ~40 в этом наборе.'),
    (15, 'Сельское хозяйство, пища, земля (31-40)', 'Сельское хозяйство, пища, земля: слова 31-40 из ~40 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Экология, география, город, инфраструктура' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);

INSERT INTO word_set_categories (course_code, parent_id, name, description, is_published, sort_order, level_code)
SELECT
  'es_ru',
  (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1),
  'Испанский мир, региональная вариативность, идиоматика', 'Тематический блок уровня C1 (~250 слов).', 1, 9, NULL
WHERE NOT EXISTS (
  SELECT 1 FROM word_set_categories c
  WHERE c.course_code = 'es_ru' AND c.name = 'Испанский мир, региональная вариативность, идиоматика'
    AND c.parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1)
);

INSERT INTO word_sets (course_code, category_id, title, description, is_published, sort_order, level_code)
SELECT
  'es_ru', (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Испанский мир, региональная вариативность, идиоматика' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1), s.title, s.description, 1, s.sort_order, NULL
FROM (
  VALUES
    (0, 'Испания: автономии, институты, реалии (1-30)', 'Испания: автономии, институты, реалии: слова 1-30 из ~40 в этом наборе.'),
    (1, 'Испания: автономии, институты, реалии (31-40)', 'Испания: автономии, институты, реалии: слова 31-40 из ~40 в этом наборе.'),
    (2, 'Латинская Америка: страны, регионы, реалии (1-30)', 'Латинская Америка: страны, регионы, реалии: слова 1-30 из ~40 в этом наборе.'),
    (3, 'Латинская Америка: страны, регионы, реалии (31-40)', 'Латинская Америка: страны, регионы, реалии: слова 31-40 из ~40 в этом наборе.'),
    (4, 'Варианты испанского и региональная лексика (1-30)', 'Варианты испанского и региональная лексика: слова 1-30 из ~40 в этом наборе.'),
    (5, 'Варианты испанского и региональная лексика (31-40)', 'Варианты испанского и региональная лексика: слова 31-40 из ~40 в этом наборе.'),
    (6, 'Устойчивые выражения и коллокации C1 (1-30)', 'Устойчивые выражения и коллокации C1: слова 1-30 из ~45 в этом наборе.'),
    (7, 'Устойчивые выражения и коллокации C1 (31-45)', 'Устойчивые выражения и коллокации C1: слова 31-45 из ~45 в этом наборе.'),
    (8, 'Фразовые шаблоны из СМИ и документов (1-30)', 'Фразовые шаблоны из СМИ и документов: слова 1-30 из ~35 в этом наборе.'),
    (9, 'Фразовые шаблоны из СМИ и документов (31-35)', 'Фразовые шаблоны из СМИ и документов: слова 31-35 из ~35 в этом наборе.'),
    (10, 'Бытовая идиоматика без жёсткого жаргона (1-30)', 'Бытовая идиоматика без жёсткого жаргона: слова 1-30 из ~35 в этом наборе.'),
    (11, 'Бытовая идиоматика без жёсткого жаргона (31-35)', 'Бытовая идиоматика без жёсткого жаргона: слова 31-35 из ~35 в этом наборе.'),
    (12, 'Ложные друзья и тонкие пары слов', 'Ложные друзья и тонкие пары слов: слова 1-15 из ~15 в этом наборе.')
) AS s(sort_order, title, description)
WHERE NOT EXISTS (
  SELECT 1 FROM word_sets ws
  WHERE ws.title = s.title
    AND ws.category_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND name = 'Испанский мир, региональная вариативность, идиоматика' AND parent_id = (SELECT id FROM word_set_categories WHERE course_code = 'es_ru' AND parent_id IS NULL AND name = 'Уровень C1' LIMIT 1) LIMIT 1)
);
