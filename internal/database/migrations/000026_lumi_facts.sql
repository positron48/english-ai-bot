-- Lumi facts: rotating "did you know" content per course/context/locale.
CREATE TABLE IF NOT EXISTS lumi_facts (
    id BIGSERIAL PRIMARY KEY,
    course_code TEXT NOT NULL DEFAULT '',
    context TEXT NOT NULL DEFAULT 'general',
    locale TEXT NOT NULL DEFAULT 'ru',
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    last_shown_on DATE,
    shown_count INTEGER NOT NULL DEFAULT 0,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (context IN ('general','grammar','reading','practice','progress','city')),
    CHECK (status IN ('active','archived'))
);

CREATE INDEX IF NOT EXISTS idx_lumi_facts_pick
    ON lumi_facts(course_code, context, locale, status, last_shown_on);

-- Seed: facts previously hardcoded in webapp locales (Spanish course, RU UI).
INSERT INTO lumi_facts (course_code, context, locale, body)
SELECT 'es_ru', 'general', 'ru', f.body
FROM (VALUES
    ('Испанский — второй язык мира по числу носителей: более 500 млн человек говорят на нём как на родном.'),
    ('«Шоколад» (chocolate) пришёл из языка ацтеков náhuatl: «xocolātl» — горький напиток из какао. Испанцы привезли его в Европу в XVI веке.'),
    ('В испанском около 10 000 слов арабского происхождения — наследие 700 лет мавританского присутствия на Пиренейском полуострове.'),
    ('«Guerrilla» — испанское слово, означающее «маленькая война». Оно стало международным термином после наполеоновских войн.'),
    ('«Mariposa» (бабочка) — одно из немногих слов неизвестного происхождения в испанском. Лингвисты до сих пор спорят о его истоках.'),
    ('Испанский восходит к народной латыни, которую принесли римские солдаты на Пиренеи в III веке до н. э.'),
    ('«Plaza» (площадь) пришло из греческого «plateía» — широкая улица. Через латынь оно попало в испанский.'),
    ('В испанском два глагола «быть»: ser — для постоянных качеств, estar — для временных. Такое разграничение редко встречается среди романских языков.'),
    ('«Tornado» — от испанского tornar (вращать). Слово вошло в английский через испанских моряков Атлантики.'),
    ('«Bonanza» по-испански означает «хорошая погода». В США оно прижилось в значении «удача» благодаря горнякам XIX века.')
) AS f(body)
WHERE NOT EXISTS (SELECT 1 FROM lumi_facts);

INSERT INTO lumi_facts (course_code, context, locale, body)
SELECT 'en_ru', 'general', 'ru', f.body
FROM (VALUES
    ('Английский — язык с самым большим словарём: в Оксфордском словаре более 600 000 слов, и каждый год добавляются новые.'),
    ('Около 60% английских слов имеют латинские или французские корни — наследие нормандского завоевания 1066 года.'),
    ('Слово «OK» — самое узнаваемое слово в мире. Его происхождение связывают с шуточным сокращением «oll korrect» из бостонских газет 1830-х.'),
    ('Шекспир ввёл в английский более 1700 слов, включая «lonely», «generous» и «eyeball».'),
    ('В английском нет официального регулятора языка — в отличие от испанского с его Real Academia Española.'),
    ('Самая частая буква английского — «e»: она встречается примерно в 11% всех слов.'),
    ('«Goodbye» — это сжатое «God be with ye» («Бог с тобой»), которое употреблялось с XVI века.'),
    ('Английский — официальный язык авиации: все пилоты мира обязаны знать его для международных полётов.'),
    ('Слово «set» имеет более 400 значений — одно из самых многозначных слов английского языка.'),
    ('Каждые два часа в английском появляется новое слово — около 4000 новых слов в год попадают в словари.')
) AS f(body)
WHERE NOT EXISTS (SELECT 1 FROM lumi_facts WHERE course_code = 'en_ru');
