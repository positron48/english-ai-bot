# Редактура ES 077–080

2026-09-05 вычитано 366 вопросов: 246 основных и 120 тренировочных в 16
исходных файлах. Исправлено 78 вопросов, 288 оставлены без изменений. Все
исправления затрагивают редакторское содержание. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| ES 077 | артикли: обобщение, части тела и абстрактные понятия | 114 | 36 | 78 |
| ES 078 | апокопа `buen`, `gran`, `primer` | 114 | 25 | 89 |
| ES 079 | нейтральное `lo` | 66 | 9 | 57 |
| ES 080 | итоговая практика сравнений и выбора | 72 | 8 | 64 |
| Всего | | 366 | 78 | 288 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. Все назначенные исходники до
начала партии были чистыми, reports отсутствовали. Placement и отдельная очередь
форм глаголов исключены.

## Существенные исправления

- ES 077: исправлены ключи и согласование в `Me duelen los ojos`,
  `Me duelen las piernas`, `Me duelen las orejas`, различие внешнего уха
  `oreja` и слуха `oído`, а также нормативное
  `el agua`. В тренировке абстрактных существительных восстановлены пропущенные
  существительные; неоднозначные варианты про боль заменены нейтральными моделями
  `me duele / me duelen + artículo + parte del cuerpo`.
- ES 078: исправлены неверные `buena apartamento` и `el primer película`,
  несогласованный род в русских формулировках и смешение физического размера с
  оценкой в `gran hombre`, `gran país`, `gran viaje`. Контексты теперь различают
  `gran` перед существительным и `grande` после него.
- ES 079: снят ложный запрет на стилистически маркированное `Importante es…`;
  нейтральная модель по-прежнему явно задаётся как `lo importante`. Разведены
  `lo bueno` и превосходное `lo mejor`, исправлена двусмысленная трактовка
  `lo dicho` и уточнено отличие `el/la + adjetivo` от нейтрального `lo`.
- ES 080: признано, что `Es la más bonita` может быть полным предложением, если
  группа известна из контекста. Исправлено нейтральное `Duermo lo suficiente`,
  устранено задание с двумя правильными ответами `muy difícil` и
  `demasiado difícil`, уточнены роли `lo bueno` и `el bueno`.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Для
тренировок штатно пересчитаны signatures; новых дубликатов во всём ES-курсе нет.

## Замечания к теории и baseline

- Outline ES 078 сводит `buen`, `gran` и `primer` к позиции перед мужским
  существительным в единственном числе. Это неверно для `gran`: форма используется
  также перед женскими существительными (`una gran oportunidad`).
- Формула ES 079 о «мужской форме» после `lo` полезна для `bueno → lo bueno`, но
  прилагательные вроде `importante` не различают род; точнее говорить о нейтральной
  форме единственного числа, совпадающей с мужской там, где род выражен.
- ES 080 представляет дополнение `de + группа` как обязательное для относительного
  суперлатива. Оно может опускаться, если группа уже восстановима из контекста.
- Draft 2020 schema всех четырёх глав не описывает часть уже существующих полей и
  типов: `accepted_answers` и/или `matching` с `pairs`. Набор ошибок до и после
  редактуры совпадает; обычный `validate-chapter.sh` принимает главы.

Теория не входит в текущий scope и не изменялась.

## Проверки и интеграция

- Все четыре главы собраны и проходят `validate-chapter.sh`. Явный AJV Draft 2020
  и raw AJV воспроизводят неизменные baseline-дефекты schema для существующих
  типов и полей. AJV-логи до и после совпадают, а `validate-chapter.sh` сохраняет
  успешный exit 0; служебные validation-файлы возвращены
  в исходное состояние.
- [check.py](check.py) проверил 366 контрактов и решений против baseline,
  актуальные fingerprints, ответы, theory bindings, signatures, отсутствие новых
  дубликатов и равенство source → final → embedded. Итог: [check.log](check.log),
  exit 0.
- Четыре final-главы и 12 тренировочных файлов точечно синхронизированы с embedded.
  Неопросные поля final сохранены, кроме штатного `updated_at` сборки.
- Все 16 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Три проверки
  `git diff --check` прошли: [diff-check.log](diff-check.log).
- Все 13 тестов механизма учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Контрольный снимок подтвердил сохранность всех 3293 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 18461/29095 вопросов основных банков, из них 224
`done` и 18237 `awaiting_verification`; 10634 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 079–082: 495 вопросов в 28 файлах. Все назначаемые исходники чистые, reports
отсутствуют. Полная цель обоих курсов остаётся активной.

## Закрытые исходники ES 077–080

- Source: `courses/spanish-grammar/chapters/077.es.grammar.noun_phrase_upgrade_precision.articles_generic_reference_body_parts_abstract_nouns/03-questions.json`
  Report: `docs/grammar-review/reports/1901ef79576689c40e5010d7.json`
  Редактура: 54/54 (`fixed` 1, `ok` 53); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/077/es.01.b1_theory_generic_reference_with_definite_article.questions.json`
  Report: `docs/grammar-review/reports/d6fead070b5eb5647c75844f.json`
  Редактура: 10/10 (`fixed` 6, `ok` 4); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/077/es.02.b2_theory_generic_reference_without_article.questions.json`
  Report: `docs/grammar-review/reports/fa068be6f44fee232ffce994.json`
  Редактура: 10/10 (`fixed` 3, `ok` 7); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/077/es.03.b3_theory_body_parts_reflexive_indirect_object.questions.json`
  Report: `docs/grammar-review/reports/025a3d8d03fa6021d77211e8.json`
  Редактура: 10/10 (`fixed` 7, `ok` 3); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/077/es.04.b4_theory_body_parts_article_vs_possessive.questions.json`
  Report: `docs/grammar-review/reports/e7828868bf9b1992e5e0b1ce.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/077/es.05.b5_theory_abstract_nouns_article_choice.questions.json`
  Report: `docs/grammar-review/reports/eadcd6deb3a13fbe202741e4.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/077/es.06.b6_theory_ru_speaker_traps_articles_generic_body_abstract.questions.json`
  Report: `docs/grammar-review/reports/ac4f75841b6c693518db7fca.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/078.es.grammar.noun_phrase_upgrade_precision.apocope_short_adjective_forms_buen_gran_primer/03-questions.json`
  Report: `docs/grammar-review/reports/72c5b76cce375d0a60610bd3.json`
  Редактура: 54/54 (`fixed` 3, `ok` 51); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/078/es.01.b1_theory_apocope_core_rule_before_singular_masculine.questions.json`
  Report: `docs/grammar-review/reports/270c0d352a14a60af7f5534b.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/078/es.02.b2_theory_bueno_buen_position_and_agreement.questions.json`
  Report: `docs/grammar-review/reports/b017bae72e0525ed0a07d35b.json`
  Редактура: 10/10 (`fixed` 3, `ok` 7); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/078/es.03.b3_theory_grande_gran_meaning_and_register.questions.json`
  Report: `docs/grammar-review/reports/f13d7b5f5daca74f87bd2aa3.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/078/es.04.b4_theory_primero_primer_and_order_expressions.questions.json`
  Report: `docs/grammar-review/reports/34ab5495902e9b535afd456a.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/078/es.05.b5_theory_exceptions_boundaries_and_fixed_phrases.questions.json`
  Report: `docs/grammar-review/reports/994f3ab4d613d357999148af.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/training_pack/chapters/078/es.06.b6_theory_ru_speaker_traps_apocope_short_forms.questions.json`
  Report: `docs/grammar-review/reports/862aba09eeae7e28096b2d19.json`
  Редактура: 10/10 (`fixed` 2, `ok` 8); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/079.es.grammar.noun_phrase_upgrade_precision.neuter_lo_lo_bueno_lo_importante/03-questions.json`
  Report: `docs/grammar-review/reports/057e178372fd37cc62a9a6d8.json`
  Редактура: 66/66 (`fixed` 9, `ok` 57); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/080.es.grammar.noun_phrase_upgrade_precision.build_speak_compare_cities_people_habits_options/03-questions.json`
  Report: `docs/grammar-review/reports/c4e480eb77eb4b94e7e90e85.json`
  Редактура: 72/72 (`fixed` 8, `ok` 64); phase: `awaiting_verification`.
