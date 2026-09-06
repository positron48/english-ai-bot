# Редактура EN 099–102

2026-09-05 вычитано 427 вопросов в 24 исходных файлах: 4 основных банка и
20 тренировочных блоков. Исправлено 177 вопросов, 250 оставлены без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 099 | пассив в разных временах | 6 | 109 | 40 | 69 |
| EN 100 | исполнитель с `by` | 6 | 105 | 59 | 46 |
| EN 101 | get-passive | 6 | 107 | 53 | 54 |
| EN 102 | causative `have/get something done` | 6 | 106 | 25 | 81 |
| Всего | | 24 | 427 | 177 | 250 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. Все 56 назначенных
source/final/validation и embedded-путей до начала партии были чистыми. Placement и
отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- EN 099: задания с `already`, `yet`, `just`, `recently` и повторяющимися событиями
  теперь задают явную временную линию. Исправлены положительное `has been approved
  yet`, позиция `just`, ложная обязательность времени по одному маркеру и дубли
  вариантов.
- EN 100: разграничены исполнитель, инструмент и канал. Устранено ложное правило,
  будто неопределённый `by a designer` неграмматичен; стилистический выбор активной
  формы и пропуск агента теперь привязаны к заданному информационному фокусу.
- EN 101: отделён get-passive от лексического `get` в `got a bonus / got the job`.
  Исправлено ложное запрещение get-passive в повторяющихся событиях: `Employees get
  paid every Friday` и `The system gets updated weekly` грамматичны. Формальная
  предпочтительность be-passive больше не выдаётся за грамматический запрет.
- EN 102: устранены неоднозначные временные контексты и пары, где `had` и `got` были
  одновременно правильны. Исправлены подсказка `trim` при ответе `cut`, дубль
  `renewed` и задания, которые называли пассив «неправильным каузативом».

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Изменены
только два ошибочных ключа: у регулярного get-passive и у нейтральной инструкции.
Signatures изменённых тренировочных вопросов пересчитаны штатной функцией; новых
дублей signatures по всему EN training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят;
  обычный AJV до и после сохраняет exit 1 только из-за незагруженной meta-schema
  Draft 2020-12. Выводы до и после совпадают, служебные `05-validation.json`
  восстановлены побайтно.
- [check.py](check.py) проверил 427 контрактов и решений, fingerprints, ответы,
  theory bindings, signatures, сохранность validation и равенство source → final →
  embedded: [check.log](check.log), exit 0.
- Все 24 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 3865 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 22682/29095 вопросов основных банков, из них 224
`done` и 22458 `awaiting_verification`; 6413 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 101–104: 288 вопросов в 4 файлах. Reports отсутствуют; все 16 назначенных
source/final/validation и embedded-путей чистые и согласованы. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники EN 099–102

### EN 099

Итого: 109/109 (`fixed` 40, `ok` 69); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/099.en.grammar.voice_focus_passive_and.passive_across_tenses/03-questions.json`
  Report: `docs/grammar-review/reports/a000f6f5cadb41b4747287c4.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/099/en.01.b1_theory_present_passive.questions.json`
  Report: `docs/grammar-review/reports/537663bd658c4dc5dfec290c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/099/en.02.b2_theory_past_passive.questions.json`
  Report: `docs/grammar-review/reports/987b2bce28cc6e4da78d0301.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/099/en.03.b3_theory_present_perfect_passive.questions.json`
  Report: `docs/grammar-review/reports/c890a8412f4695b69d04f4ca.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/099/en.04.b4_theory_active_vs_passive_tense.questions.json`
  Report: `docs/grammar-review/reports/bbccf131bbc92b3099d7fcb4.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/099/en.05.b5_theory_time_markers.questions.json`
  Report: `docs/grammar-review/reports/2abd500fda9c684820e89583.json`

### EN 100

Итого: 105/105 (`fixed` 59, `ok` 46); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/100.en.grammar.voice_focus_passive_and.by_agent_when_to/03-questions.json`
  Report: `docs/grammar-review/reports/ef8c8303551e8d9705d4e0e3.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/100/en.01.b1_theory_omit_agent.questions.json`
  Report: `docs/grammar-review/reports/8c68132126dd0377c1085a50.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/100/en.02.b2_theory_include_agent.questions.json`
  Report: `docs/grammar-review/reports/40e804295ceb03a489b27217.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/100/en.03.b3_theory_agent_types.questions.json`
  Report: `docs/grammar-review/reports/b4b64043474d87077d850c8a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/100/en.04.b4_theory_avoid_by.questions.json`
  Report: `docs/grammar-review/reports/752408eff0116418eac46d4f.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/100/en.05.b5_theory_news_reports.questions.json`
  Report: `docs/grammar-review/reports/1c255d7743910d009ff5a961.json`

### EN 101

Итого: 107/107 (`fixed` 53, `ok` 54); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/101.en.grammar.voice_focus_passive_and.get_passive_it_happened/03-questions.json`
  Report: `docs/grammar-review/reports/86558751d3483ffe3d6f3db3.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/101/en.01.b1_theory_meaning_get.questions.json`
  Report: `docs/grammar-review/reports/620fd514387ec6abda7563d9.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/101/en.02.b2_theory_form_get.questions.json`
  Report: `docs/grammar-review/reports/79e7ac41d5a280434daf5714.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/101/en.03.b3_theory_be_vs_get.questions.json`
  Report: `docs/grammar-review/reports/a65d57ee9f6b1197637118a0.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/101/en.04.b4_theory_common_contexts.questions.json`
  Report: `docs/grammar-review/reports/7357d21bf57c322b74218fad.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/101/en.05.b5_theory_limits.questions.json`
  Report: `docs/grammar-review/reports/65c353e90b6ff43c3a0fd039.json`

### EN 102

Итого: 106/106 (`fixed` 25, `ok` 81); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/102.en.grammar.voice_focus_passive_and.causative_have_get_something/03-questions.json`
  Report: `docs/grammar-review/reports/af2ea5403789377c43dc77f2.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/102/en.01.b1_theory_meaning_causative.questions.json`
  Report: `docs/grammar-review/reports/27d7ecfa54ee8f6b5468ab03.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/102/en.02.b2_theory_form_have_get.questions.json`
  Report: `docs/grammar-review/reports/1d14084eef122a0116568691.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/102/en.03.b3_theory_tenses.questions.json`
  Report: `docs/grammar-review/reports/e24b76858ae7b3fe6a59429c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/102/en.04.b4_theory_causative_vs_passive.questions.json`
  Report: `docs/grammar-review/reports/945f6356e21f6748cdaf7f42.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/102/en.05.b5_theory_common_contexts.questions.json`
  Report: `docs/grammar-review/reports/073dc4c1826cb1cf188a3760.json`

## Следующая партия

- ES 101: 72 вопроса — `porque`, `como`, `por eso`, `así que`.
- ES 102: 72 вопроса — `pero`, `sino`, `aunque`, `sin embargo`.
- ES 103: 72 вопроса — `cuando`, `antes de`, `después de`, `mientras`.
- ES 104: 72 вопроса — относительные `que`, `quien`, `donde`, `cuyo`.
- Итого: 288 вопросов в 4 файлах; reports отсутствуют.
- Все 16 назначенных source/final/validation и embedded-путей чистые и согласованы.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
