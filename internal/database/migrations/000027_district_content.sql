-- District content: descriptions, images, Lumi tips and building labels.
-- Content previously hardcoded in webapp views / Linglow Markup prototype.

WITH meta(code, desc_ru, image, buildings) AS (VALUES
    ('a0_spark_gate',     'Ворота в город. Первые шаги, первые слова.',                    'dist_vida',
     '[{"name":"Jardín de Frases","type":"grammar","x":18,"y":22},{"name":"Mercado de Palabras","type":"word_market","x":72,"y":18},{"name":"Quiosco de Lectura","type":"reading","x":12,"y":55},{"name":"Cabinas de Conversación","type":"conversation","x":76,"y":52}]'),
    ('a1_clear_plaza',    'Центр города. Здесь всё началось — говори, читай, общайся.',     'dist_parques',
     '[{"name":"Jardín de Frases","type":"grammar","x":18,"y":22},{"name":"Mercado de Palabras","type":"word_market","x":72,"y":18},{"name":"Quiosco de Lectura","type":"reading","x":12,"y":55},{"name":"Cabinas de Conversación","type":"conversation","x":76,"y":52}]'),
    ('a2_living_quarter', 'Жилые кварталы — повседневная речь, жизнь города.',              'dist_mercados',
     '[{"name":"Jardín de Frases","type":"grammar","x":18,"y":22},{"name":"Mercado de Palabras","type":"word_market","x":72,"y":18},{"name":"Quiosco de Lectura","type":"reading","x":12,"y":55},{"name":"Cabinas de Conversación","type":"conversation","x":76,"y":52}]'),
    ('b1_story_bridges',  'Мосты историй — длинные тексты и живые диалоги.',                'dist_cafeterias',
     '[{"name":"Jardín de Frases","type":"grammar","x":18,"y":22},{"name":"Mercado de Palabras","type":"word_market","x":72,"y":18},{"name":"Quiosco de Lectura","type":"reading","x":12,"y":55},{"name":"Cabinas de Conversación","type":"conversation","x":76,"y":52}]'),
    ('b2_high_district',  'Верхний город — сложная речь, нюансы и стиль.',                  'dist_viajes',
     '[{"name":"Jardín de Frases","type":"grammar","x":18,"y":22},{"name":"Mercado de Palabras","type":"word_market","x":72,"y":18},{"name":"Quiosco de Lectura","type":"reading","x":12,"y":55},{"name":"Cabinas de Conversación","type":"conversation","x":76,"y":52}]'),
    ('c1_mastery_campus', 'Кампус мастерства — финальная ступень владения языком.',          'dist_viajes',
     '[{"name":"Jardín de Frases","type":"grammar","x":18,"y":22},{"name":"Mercado de Palabras","type":"word_market","x":72,"y":18},{"name":"Quiosco de Lectura","type":"reading","x":12,"y":55},{"name":"Cabinas de Conversación","type":"conversation","x":76,"y":52}]')
)
UPDATE districts d
SET description = m.desc_ru,
    metadata_json = d.metadata_json || jsonb_build_object(
        'image', m.image,
        'desc_i18n', jsonb_build_object('ru', m.desc_ru),
        'lumi_tips', jsonb_build_object(
            'low',  'Продолжи путь отсюда! Ты справишься 🔥',
            'mid',  'Тут хорошо учить слова для жизни! 🌿',
            'high', '¡Ya hablas con confianza aquí! ✨'
        ),
        'buildings', m.buildings::jsonb
    ),
    updated_at = CURRENT_TIMESTAMP
FROM meta m
WHERE d.code = m.code
  AND (d.description IS NULL OR d.description = '');
