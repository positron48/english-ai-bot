# Редактура EN 119–122

2026-09-06 вычитано 414 вопросов в 24 исходных файлах: 4 основных банка и
20 тренировочных блоков. Исправлено 113 вопросов, 301 оставлен без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 119 | reported statements, `say` и `tell` | 6 | 107 | 32 | 75 |
| EN 120 | reported questions и порядок слов | 6 | 97 | 33 | 64 |
| EN 121 | reported commands и requests | 6 | 107 | 25 | 82 |
| EN 122 | reporting verbs `advise`, `suggest`, `insist` | 6 | 103 | 23 | 80 |
| Всего | | 24 | 414 | 113 | 301 |

Редактор: `Codex editor 2026-09-06`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. До начала партии все 56 назначенных
source/final/validation и embedded-путей были чистыми. Placement и отдельная
очередь форм глаголов ES исключены.

## Существенные исправления

- EN 119: исправлены неполные fill-blank конструкции `had seen`, `the next day`
  и `the day before`. Уточнены условия backshift, устранены конкурирующие
  правильные варианты с `say to` и `tell`, восстановлены смена местоимений и
  временных указателей в тренировочных заданиях.
- EN 120: исправлено обобщение `are → was`, неверные объяснения Present Perfect,
  задания с несколькими правильными косвенными вопросами и ложное правило о
  `if` в официальной речи. Вопросы теперь отдельно проверяют порядок слов,
  стандартный backshift и выбор `if/whether`.
- EN 121: конструкции `to not` больше не объявляются безусловно неграмматичными:
  задания требуют нейтральный порядок `not to` и отмечают контрастивную
  альтернативу. Исправлены неверный ключ для просьбы `asked me not to be late`,
  значения `ask/tell` и смена `me → him/her`.
- EN 122: задания на `advise` теперь прямо называют проверяемый паттерн, не
  отрицая допустимые V-ing и that-clause. Для `insist` разграничены значения
  «требовать» и «утверждать»; исправлена идиома `take a break` и устранены
  конкурирующие правильные ответы.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Signatures
изменённых тренировочных вопросов пересчитаны штатной функцией; новых дублей
signatures по всему EN training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят;
  обычный AJV до и после сохраняет exit 1 только из-за незагруженной meta-schema
  Draft 2020-12. Служебные `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 414 контрактов и решений, fingerprints, ответы,
  theory bindings, signatures, сохранность validation и равенство source → final →
  embedded: [check.log](check.log), exit 0.
- Все 24 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [grammar-review-tests.log](grammar-review-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 4358 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 26203/29095 вопросов основных банков, из них 224
`done` и 25979 `awaiting_verification`; 2892 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 121–124: 280 вопросов в 4 файлах; reports отсутствуют. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники EN 119–122

### EN 119

Итого: 107/107 (`fixed` 32, `ok` 75); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/119.en.grammar.reported_speech_meaning_transfer.reported_statements_say_tell/03-questions.json`
  Report: `docs/grammar-review/reports/b14a5156519e536e99513854.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/119/en.01.b1_theory_overview.questions.json`
  Report: `docs/grammar-review/reports/a9c7a027fe1960e765f4fee8.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/119/en.02.b2_theory_say_tell.questions.json`
  Report: `docs/grammar-review/reports/94f5c90eb266c17541ce75cb.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/119/en.03.b3_theory_backshift.questions.json`
  Report: `docs/grammar-review/reports/3d343793e0deaf2bf464d67a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/119/en.04.b4_theory_time_place.questions.json`
  Report: `docs/grammar-review/reports/32fbd63e513134fc5af991a3.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/119/en.05.b5_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/9e724e12df6d9bd3f2920c8b.json`

### EN 120

Итого: 97/97 (`fixed` 33, `ok` 64); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/120.en.grammar.reported_speech_meaning_transfer.reported_questions_word_order/03-questions.json`
  Report: `docs/grammar-review/reports/5c27994493428996024e23ec.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/120/en.01.b1_theory_overview.questions.json`
  Report: `docs/grammar-review/reports/86e5b617088d8423c1e79711.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/120/en.02.b2_theory_if_whether.questions.json`
  Report: `docs/grammar-review/reports/00e066b13075a9fe51c4cdbc.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/120/en.03.b3_theory_wh_questions.questions.json`
  Report: `docs/grammar-review/reports/aa2fab145828a8df9efe50b2.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/120/en.04.b4_theory_backshift.questions.json`
  Report: `docs/grammar-review/reports/c275dd1d592fcdea65fd7db8.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/120/en.05.b5_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/17a165a7ea82db4da34cfd8b.json`

### EN 121

Итого: 107/107 (`fixed` 25, `ok` 82); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/121.en.grammar.reported_speech_meaning_transfer.reported_commands_requests_told/03-questions.json`
  Report: `docs/grammar-review/reports/55db12068917b0dbdc01968f.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/121/en.01.b1_theory_command_pattern.questions.json`
  Report: `docs/grammar-review/reports/73a39786b2d383a43cb34d77.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/121/en.02.b2_theory_tell_vs_ask.questions.json`
  Report: `docs/grammar-review/reports/e8df2bb4224ab2ba740ab606.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/121/en.03.b3_theory_negative_commands.questions.json`
  Report: `docs/grammar-review/reports/f3f748d19b3fcad6fb748dbf.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/121/en.04.b4_theory_reporting_verbs_requests.questions.json`
  Report: `docs/grammar-review/reports/0e4b7b0cbae25f5b318ec604.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/121/en.05.b5_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/1aa03fcdb729d5e22bfebd07.json`

### EN 122

Итого: 103/103 (`fixed` 23, `ok` 80); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/122.en.grammar.reported_speech_meaning_transfer.reporting_verbs_advise_suggest/03-questions.json`
  Report: `docs/grammar-review/reports/f442a89aefe44ff75fcb0f28.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/122/en.01.b1_theory_advise_pattern.questions.json`
  Report: `docs/grammar-review/reports/06dbd244cc12bc6aabf43a63.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/122/en.02.b2_theory_suggest_gerund.questions.json`
  Report: `docs/grammar-review/reports/826765d8cca6dd14db07553f.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/122/en.03.b3_theory_suggest_that_clause.questions.json`
  Report: `docs/grammar-review/reports/e821f02a64415ae30876320e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/122/en.04.b4_theory_insist_that_clause.questions.json`
  Report: `docs/grammar-review/reports/ea670262be3b8ba3df95af20.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/122/en.05.b5_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/dd303e821c456d6ac914d098.json`

## Следующая партия

- ES 121: 70 вопросов — сомнение, отрицание, возможность и уверенность.
- ES 122: 70 вопросов — безличные выражения и триггеры уверенности.
- ES 123: 70 вопросов — отрицательные и вежливые команды с subjuntivo.
- ES 124: 70 вопросов — выбор indicativo/subjuntivo в noun clauses.
- Итого: 280 вопросов в 4 файлах; reports отсутствуют.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
