# Редактура EN 083–086

2026-09-05 вычитано 447 вопросов: 237 основных и 210 тренировочных в 25
исходных файлах. Исправлено 140 вопросов, 307 оставлены без изменений. Все
исправления затрагивают редакторское содержание. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| EN 083 | because/so; although/but; while/whereas | 128 | 28 | 100 |
| EN 084 | придаточные времени | 105 | 45 | 60 |
| EN 085 | noun clauses | 103 | 44 | 59 |
| EN 086 | косвенные вопросы | 111 | 23 | 88 |
| Всего | | 447 | 140 | 307 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. Все назначенные исходники до
начала партии были чистыми, reports отсутствовали. Placement исключён.

## Существенные исправления

- EN 083: разведены причина, результат и контраст в неоднозначных контекстах.
  Упражнения больше не называют союзное `though he went...` наречием в конце
  предложения; для конечного `though` даны настоящие отдельные предложения.
  Исправлены задания с двумя правильными вариантами `while/whereas`, неточное
  описание запятой и ложный запрет совместного появления although и but, когда
  они связывают разные части.
- EN 084: временные последовательности теперь прямо задают порядок или
  немедленность, чтобы `when`, `after` и `as soon as` не конкурировали как
  допустимые ответы. Исправлен неверный ключ в отрицательном примере, заменены
  нелогичные ситуации, выправлено согласование времён и системная смесь
  `время-clause` / `придаточном времени clause`.
- EN 085: исправлено неграмматичное `I know that you mean` на `I know what you
  mean`. В тренировках устранены многочисленные пары, где варианты с that и без
  that, if и whether, present simple и present continuous были одновременно
  верны. Задания на negative raising теперь проверяют нейтральную степень
  уверенности, а не объявляют остальные грамматичные варианты ошибками.
- EN 086: в вежливых вопросах `Can/Could you tell me...` восстановлены
  вопросительные знаки. Уточнены контексты для времён и аспектов, исправлены
  лица при переводе прямого вопроса в косвенный и устранены вторые правильные
  варианты с прямым порядком слов.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Для
тренировок штатно пересчитаны signatures; новых дубликатов во всём EN-курсе нет.

## Замечания к теории и baseline

- EN 083 объединяет союзное though и его разговорное употребление в конце
  предложения. Вопросы теперь явно различают эти две конструкции.
- EN 084 формулирует базовое правило пунктуации для нейтрального стиля и правило
  present simple в будущих придаточных времени. Вопросы не распространяют их
  на контрастное обособление и особые волевые значения will.
- EN 085 описывает отрицание в главной части как более мягкое и нейтральное.
  Исправленные задания проверяют именно прагматическое предпочтение, сохраняя
  грамматичность вариантов с отрицанием внутри noun clause.
- EN 086 определяет конечный знак по типу главного предложения; все упражнения
  теперь согласованы с этим правилом.
- Обычный AJV без `--spec=draft2020` не загружает meta-schema Draft 2020-12 и
  штатно возвращает exit 1. Явный Draft 2020 и `validate-chapter.sh` принимают
  все четыре главы; результаты до и после редактуры совпадают.

Теория не входит в текущий scope и не изменялась.

## Проверки и интеграция

- Все четыре главы собраны и проходят `validate-chapter.sh` без предупреждений.
  Явный AJV Draft 2020 проходит с exit 0; baseline-ошибка обычного AJV полностью
  совпадает до и после. Служебные validation-файлы возвращены в исходное состояние.
- [check.py](check.py) проверил 447 контрактов и решений против baseline,
  актуальные fingerprints, ответы, theory bindings, signatures, отсутствие новых
  дубликатов и равенство source → final → embedded. Итог: [check.log](check.log),
  exit 0.
- Четыре final-главы и 21 тренировочный файл точечно синхронизированы с embedded.
  Неопросные поля final сохранены, кроме штатного `updated_at` сборки.
- Все 25 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Три проверки
  `git diff --check` прошли: [diff-check.log](diff-check.log).
- Все 13 тестов механизма учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Контрольный снимок подтвердил сохранность всех 3458 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 19691/29095 вопросов основных банков, из них 224
`done` и 19467 `awaiting_verification`; 9404 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 085–088: 288 вопросов в 4 файлах. Все назначаемые исходники чистые, reports
отсутствуют. Полная цель обоих курсов остаётся активной.

## Закрытые исходники EN 083–086

- Source: `courses/english-grammar/chapters/083.en.grammar.complex_sentences_1_connecting.because_so_although_but/03-questions.json`
  Report: `docs/grammar-review/reports/def45bc52087a2fab812fc7c.json`
  Редактура: 68/68 (`fixed` 6, `ok` 62); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/083/en.01.b1_because_cause.questions.json`
  Report: `docs/grammar-review/reports/c2249e59ad46eb450654df1f.json`
  Редактура: 10/10 (`fixed` 7, `ok` 3); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/083/en.02.b2_so_result.questions.json`
  Report: `docs/grammar-review/reports/32c24cf8e262e7a81b5f0d28.json`
  Редактура: 10/10 (`fixed` 2, `ok` 8); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/083/en.03.b3_although_contrast.questions.json`
  Report: `docs/grammar-review/reports/56b27c46b22e3d907704093a.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/083/en.04.b4_while_whereas.questions.json`
  Report: `docs/grammar-review/reports/684e1c24823464d70ffaf3e9.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/083/en.05.b5_clause_order_punct.questions.json`
  Report: `docs/grammar-review/reports/d8c2d79848f0225bcff6cf89.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/083/en.06.b6_common_errors.questions.json`
  Report: `docs/grammar-review/reports/8a02195f7bb48fe7c998e458.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.

- Source: `courses/english-grammar/chapters/084.en.grammar.complex_sentences_1_connecting.time_clauses_again_when/03-questions.json`
  Report: `docs/grammar-review/reports/f6c534e484a19562301e9310.json`
  Редактура: 55/55 (`fixed` 11, `ok` 44); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/084/en.01.b1_time_clause_when.questions.json`
  Report: `docs/grammar-review/reports/51ea11d9d24e37a88a4b8e50.json`
  Редактура: 10/10 (`fixed` 9, `ok` 1); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/084/en.02.b2_time_clause_before_after.questions.json`
  Report: `docs/grammar-review/reports/95d759aa70a523d5354979aa.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/084/en.03.b3_time_clause_as_soon_as.questions.json`
  Report: `docs/grammar-review/reports/f3a8f8288e7e5421c992dc31.json`
  Редактура: 10/10 (`fixed` 1, `ok` 9); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/084/en.04.b4_future_time_clause_no_will.questions.json`
  Report: `docs/grammar-review/reports/e76bc56406eb10b88b8a18cc.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/084/en.05.b5_time_clause_position_punctuation.questions.json`
  Report: `docs/grammar-review/reports/6474dbe4bdb8056463f4d857.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.

- Source: `courses/english-grammar/chapters/085.en.grammar.complex_sentences_1_connecting.noun_clauses_i_think/03-questions.json`
  Report: `docs/grammar-review/reports/8aa7ec31ac275490e593ef48.json`
  Редактура: 53/53 (`fixed` 13, `ok` 40); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/085/en.01.b1_noun_clause_basic.questions.json`
  Report: `docs/grammar-review/reports/e7ca6fd3110e9ab42318fa74.json`
  Редактура: 10/10 (`fixed` 6, `ok` 4); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/085/en.02.b2_that_optional.questions.json`
  Report: `docs/grammar-review/reports/fc288d9692031cef21acc751.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/085/en.03.b3_if_whether_noun_clause.questions.json`
  Report: `docs/grammar-review/reports/ba8891770d73b40416b11457.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/085/en.04.b4_question_word_order.questions.json`
  Report: `docs/grammar-review/reports/fdc901db50c4b61ed623ba4b.json`
  Редактура: 10/10 (`fixed` 7, `ok` 3); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/085/en.05.b5_negation_scope.questions.json`
  Report: `docs/grammar-review/reports/ffdf73fa6d3599c7ece3fdc7.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.

- Source: `courses/english-grammar/chapters/086.en.grammar.complex_sentences_1_connecting.indirect_questions_could_you/03-questions.json`
  Report: `docs/grammar-review/reports/407d5c6a8ca0f5646247602c.json`
  Редактура: 61/61 (`fixed` 3, `ok` 58); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/086/en.01.b1_theory_indirect_questions_purpose.questions.json`
  Report: `docs/grammar-review/reports/9ecdeb269307ff6274a9a042.json`
  Редактура: 10/10 (`fixed` 1, `ok` 9); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/086/en.02.b2_theory_statement_word_order.questions.json`
  Report: `docs/grammar-review/reports/e5f5233a89e13544d92feaaf.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/086/en.03.b3_theory_yes_no_if_whether.questions.json`
  Report: `docs/grammar-review/reports/fe4f742f502b5c1fbb461c2c.json`
  Редактура: 10/10 (`fixed` 6, `ok` 4); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/086/en.04.b4_theory_wh_indirect_questions.questions.json`
  Report: `docs/grammar-review/reports/a1217c90a7b8000141280d8c.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.

- Source: `courses/english-grammar/training_pack/chapters/086/en.05.b5_theory_punctuation_reporting_verbs.questions.json`
  Report: `docs/grammar-review/reports/4acbf9f26635e60ae7b09e78.json`
  Редактура: 10/10 (`fixed` 3, `ok` 7); phase: `awaiting_verification`.

## Следующая партия

- ES 085: 72 вопроса — глаголы с предлогами.
- ES 086: 72 вопроса — прилагательные с предлогами.
- ES 087: 72 вопроса — направления, причины, цели и отношения.
- ES 088: 72 вопроса — прямые объектные местоимения lo/la/los/las.
- Итого: 288 вопросов в 4 файлах; назначаемые исходники чистые, reports отсутствуют.
- Baseline ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
