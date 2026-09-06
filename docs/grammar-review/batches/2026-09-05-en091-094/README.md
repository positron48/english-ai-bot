# Редактура EN 091–094

2026-09-05 вычитано 438 вопросов в 25 исходных файлах: 4 основных банка и
21 тренировочный блок. Исправлено 145 вопросов, 293 оставлены без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 091 | результат Present Perfect | 6 | 109 | 48 | 61 |
| EN 092 | опыт, `ever/never`, `been/gone` | 7 | 125 | 41 | 84 |
| EN 093 | длительность с `for/since/how long` | 6 | 97 | 21 | 76 |
| EN 094 | Present Perfect и Past Simple | 6 | 107 | 35 | 72 |
| Всего | | 25 | 438 | 145 | 293 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. Все назначенные исходники,
final/validation и embedded-файлы до начала партии были чистыми. Placement и
отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- EN 091: устранены неоднозначные пары Past Simple/Present Perfect, альтернативно
  правильные позиции `yet`, дублирующиеся ответы и неестественные сценарии текущего
  результата. Тренировочные пропуски теперь образуют полные английские предложения.
- EN 092: убраны ложные утверждения, будто `before` нельзя употреблять с Present
  Perfect, а `never before` избыточно. Исправлены `been/gone`, сценарии открытого
  периода и десять тренировочных вопросов, в каждом из которых было два правильных
  варианта Past Simple.
- EN 093: сохранены допустимые Present Perfect Continuous и разные способы ответа
  на `How long`. Контекст завершённости добавлен там, где Past Simple раньше нельзя
  было доказать; разведены несколько альтернативно правильных ответов с `since`.
- EN 094: уточнено, что отсутствие временного маркера само по себе не определяет
  время. Убраны неоднозначности с американским Past Simple после `just/already`,
  открытыми периодами и `before`; исправлен неестественный пример `slept late`.

IDs, порядок, типы, difficulty, theory bindings, choice IDs и IDs правильных
ответов сохранены. Signatures изменённых тренировочных вопросов пересчитаны штатной
функцией; новых дублей signatures по всему EN training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят;
  обычный AJV до и после сохраняет exit 1 только из-за незагруженной meta-schema
  Draft 2020-12. Выводы валидаторов до и после совпадают дословно, служебные
  `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 438 контрактов и решений, fingerprints, ответы,
  theory bindings, signatures, сохранность validation и равенство source → final →
  embedded: [check.log](check.log), exit 0.
- Все 25 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 3651 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 21155/29095 вопросов основных банков, из них 224
`done` и 20931 `awaiting_verification`; 7940 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 093–096: 282 вопроса в 4 файлах. Reports отсутствуют; входящие локальные
изменения всех четырёх глав нужно включить в baseline. Полная цель обоих курсов
остаётся активной.

## Закрытые исходники EN 091–094

### EN 091

Итого: 109/109 (`fixed` 48, `ok` 61); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/091.en.grammar.perfect_aspect_experience_result.result_i_ve_lost/03-questions.json`
  Report: `docs/grammar-review/reports/ca9b6e73faec0113aa86a816.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/091/en.01.b1_theory_result_link_now.questions.json`
  Report: `docs/grammar-review/reports/a658e29ba19d35ee88123b9e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/091/en.02.b2_theory_state_vs_event.questions.json`
  Report: `docs/grammar-review/reports/897792acc4cc3d5eae6ae294.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/091/en.03.b3_theory_markers.questions.json`
  Report: `docs/grammar-review/reports/4588bac6219ed06a64418702.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/091/en.04.b4_theory_word_order.questions.json`
  Report: `docs/grammar-review/reports/81dadca40a5e042a7c576a6a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/091/en.05.b5_theory_evidence.questions.json`
  Report: `docs/grammar-review/reports/4e0bb26bd49b887320ff60d7.json`

### EN 092

Итого: 125/125 (`fixed` 41, `ok` 84); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/092.en.grammar.perfect_aspect_experience_result.experience_have_you_ever/03-questions.json`
  Report: `docs/grammar-review/reports/bbc658debd48ee6c6103d090.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/092/en.01.b1_theory_experience_meaning.questions.json`
  Report: `docs/grammar-review/reports/0ea210932c019130192778cd.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/092/en.02.b2_theory_markers_ever_never.questions.json`
  Report: `docs/grammar-review/reports/c9c371ef0da828281ec1e8d0.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/092/en.03.b3_theory_questions_short_answers.questions.json`
  Report: `docs/grammar-review/reports/531510b56dc70340a8859d09.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/092/en.04.b4_theory_open_time_periods.questions.json`
  Report: `docs/grammar-review/reports/5b3da0287b921e958e4b7ca7.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/092/en.05.b5_theory_been_vs_gone.questions.json`
  Report: `docs/grammar-review/reports/fca65aa2ccc1abe351b9ef0f.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/092/en.06.b6_theory_time_marker_traps.questions.json`
  Report: `docs/grammar-review/reports/b04dc2ffa5b3d12f976bc680.json`

### EN 093

Итого: 97/97 (`fixed` 21, `ok` 76); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/093.en.grammar.perfect_aspect_experience_result.duration_for_since_how/03-questions.json`
  Report: `docs/grammar-review/reports/9300a13ec9b29538e1c4b235.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/093/en.01.b1_theory_duration_up_to_now.questions.json`
  Report: `docs/grammar-review/reports/0bee66708e8d7442aa1b5d8a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/093/en.02.b2_theory_for_vs_since.questions.json`
  Report: `docs/grammar-review/reports/4684fb72b2f0a0748b44f0ce.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/093/en.03.b3_theory_how_long.questions.json`
  Report: `docs/grammar-review/reports/96ce6495368852b8d5a2e816.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/093/en.04.b4_theory_since_clauses_positions.questions.json`
  Report: `docs/grammar-review/reports/c6fa0aff1cbd20b54f89f302.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/093/en.05.b5_theory_duration_vs_finished_time.questions.json`
  Report: `docs/grammar-review/reports/557f41b742f5e3b449282b74.json`

### EN 094

Итого: 107/107 (`fixed` 35, `ok` 72); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/094.en.grammar.perfect_aspect_experience_result.present_perfect_vs_past/03-questions.json`
  Report: `docs/grammar-review/reports/c3cfd24ccaabd3d40c704f57.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/094/en.01.b1_theory_finished_time.questions.json`
  Report: `docs/grammar-review/reports/60bc0ca4b264104cf121bf9c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/094/en.02.b2_theory_result_now.questions.json`
  Report: `docs/grammar-review/reports/413d71b50351902897277557.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/094/en.03.b3_theory_time_markers.questions.json`
  Report: `docs/grammar-review/reports/615e96a129adec468da21922.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/094/en.04.b4_theory_questions_follow_up.questions.json`
  Report: `docs/grammar-review/reports/84a71c3b9eeaf06abed594ae.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/094/en.05.b5_theory_open_closed_time.questions.json`
  Report: `docs/grammar-review/reports/33c14042ae7c3c9faf3852c7.json`

## Следующая партия

- ES 093: 70 вопросов — практика `dar/mostrar/comprar/decir` с объектными местоимениями.
- ES 094: 68 вопросов — возвратные и взаимные действия.
- ES 095: 72 вопроса — местоименные глаголы с изменением значения.
- ES 096: 72 вопроса — accidental `se`: `se me cayó`, `se nos olvidó`.
- Итого: 282 вопроса в 4 файлах; reports отсутствуют.
- Во всех четырёх главах уже изменены source/final и embedded-копии; текущие final
  и embedded совпадают. Эти изменения нужно включить в побитовый baseline и сохранить.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
