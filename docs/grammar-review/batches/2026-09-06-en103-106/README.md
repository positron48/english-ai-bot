# Редактура EN 103–106

2026-09-06 вычитан 431 вопрос в 24 исходных файлах: 4 основных банка и
20 тренировочных блоков. Исправлено 117 вопросов, 314 оставлены без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 103 | безличный пассив | 6 | 108 | 27 | 81 |
| EN 104 | пассив в новостях и описании процессов | 6 | 109 | 32 | 77 |
| EN 105 | zero conditional | 6 | 107 | 42 | 65 |
| EN 106 | first conditional | 6 | 107 | 16 | 91 |
| Всего | | 24 | 431 | 117 | 314 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. Все 56 назначенных
source/final/validation и embedded-путей до начала партии были чистыми. Placement и
отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- EN 103: добавлен временной контекст для различения `to be` и `to have been`.
  Устранено ложное правило, будто `that` после `It is said/believed` можно просто
  удалить. Уточнены задания на точную безличную модель, perfect/continuous infinitive
  и отличие личного пассива от безличного.
- EN 104: разведены `have/get something done`, прошлое и настоящее время, а также
  регулярный процесс и единичное событие. Исправлены переводы сервисного каузатива,
  ложное требование общего вспомогательного глагола и задания с несколькими ошибками
  или несколькими допустимыми ответами.
- EN 105: грамматичный first conditional больше не объявляется ошибкой только потому,
  что упражнение просит zero conditional. Задания на `if/when` теперь задают нужный
  смысл; исправлены термин «условное предложение», неоднозначные варианты и ложный
  запрет zero conditional в прошедшем времени.
- EN 106: добавлен будущий контекст, уточнено правило о `will` в if-clause и исправлены
  задания, где грамматичная конструкция называлась ошибкой. Исправлены нелогичный
  пример смешения цветов, модальный пример и варианты с несколькими правильными
  предложениями.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Исправлены
три ошибочных ключа: форма инфинитива в EN 103, оценка допустимости прошедшего zero
conditional и выбор `if/when` в EN 105. Signatures изменённых тренировочных вопросов
пересчитаны штатной функцией; новых дублей signatures по всему EN training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят;
  обычный AJV до и после сохраняет exit 1 только из-за незагруженной meta-schema
  Draft 2020-12. Выводы до и после совпадают, служебные `05-validation.json`
  восстановлены побайтно.
- [check.py](check.py) проверил 431 контракт и решение, fingerprints, ответы, theory
  bindings, signatures, сохранность validation и равенство source → final → embedded:
  [check.log](check.log), exit 0.
- Все 24 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 3963 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 23401/29095 вопросов основных банков, из них 224
`done` и 23177 `awaiting_verification`; 5694 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 105–108: 270 вопросов в 4 файлах. Reports отсутствуют; все 16 назначенных
source/final/validation и embedded-путей чистые и согласованы. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники EN 103–106

### EN 103

Итого: 108/108 (`fixed` 27, `ok` 81); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/103.en.grammar.voice_focus_passive_and.impersonal_passive_it_is/03-questions.json`
  Report: `docs/grammar-review/reports/bc19823cf65e03280616115d.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/103/en.01.b1_theory_it_passive.questions.json`
  Report: `docs/grammar-review/reports/a526e895a9d44f0037fbaacb.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/103/en.02.b2_theory_he_passive.questions.json`
  Report: `docs/grammar-review/reports/19703a84bc544fe5466de24a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/103/en.03.b3_theory_tense_alignment.questions.json`
  Report: `docs/grammar-review/reports/05a933871a5a7f9ae431530a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/103/en.04.b4_theory_reporting_verbs.questions.json`
  Report: `docs/grammar-review/reports/45faa0dc25a9a2d9d294652b.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/103/en.05.b5_theory_style_contexts.questions.json`
  Report: `docs/grammar-review/reports/41a73aa3083a05faee1f2df5.json`

### EN 104

Итого: 109/109 (`fixed` 32, `ok` 77); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/104.en.grammar.voice_focus_passive_and.build_write_news_processes/03-questions.json`
  Report: `docs/grammar-review/reports/d9cec8bc57db135b28dd1ce5.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/104/en.01.b1_theory_news_structure.questions.json`
  Report: `docs/grammar-review/reports/07a33d331906fb191cb7d9a2.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/104/en.02.b2_theory_process_descriptions.questions.json`
  Report: `docs/grammar-review/reports/223d2000d3a8cdf2413f3355.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/104/en.03.b3_theory_service_causative.questions.json`
  Report: `docs/grammar-review/reports/7fac3391fbdbcb142b42d7b1.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/104/en.04.b4_theory_style_consistency.questions.json`
  Report: `docs/grammar-review/reports/02cfd8008e5a518a904621e3.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/104/en.05.b5_theory_examples_templates.questions.json`
  Report: `docs/grammar-review/reports/0b3f9fe9e1066092c70c34d7.json`

### EN 105

Итого: 107/107 (`fixed` 42, `ok` 65); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/105.en.grammar.conditionals_hypotheticals.zero_conditional/03-questions.json`
  Report: `docs/grammar-review/reports/c50baf5feb762cf9ace652b7.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/105/en.01.b1_theory_meaning.questions.json`
  Report: `docs/grammar-review/reports/7333626a0724ecdfb5cfe0ec.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/105/en.02.b2_theory_form.questions.json`
  Report: `docs/grammar-review/reports/17d0973ced0f7924dc00df5a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/105/en.03.b3_theory_if_when.questions.json`
  Report: `docs/grammar-review/reports/f45ce3e4a41d2fe96de04f03.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/105/en.04.b4_theory_word_order.questions.json`
  Report: `docs/grammar-review/reports/d270c23d6b6df3ba0d2f9480.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/105/en.05.b5_theory_zero_vs_first.questions.json`
  Report: `docs/grammar-review/reports/c465a0dcdd2583d1a7f7ab7b.json`

### EN 106

Итого: 107/107 (`fixed` 16, `ok` 91); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/106.en.grammar.conditionals_hypotheticals.first_conditional/03-questions.json`
  Report: `docs/grammar-review/reports/69727d7afb1ad5f33f654368.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/106/en.01.b1_theory_meaning.questions.json`
  Report: `docs/grammar-review/reports/456e3759ffb6157816249cdc.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/106/en.02.b2_theory_form.questions.json`
  Report: `docs/grammar-review/reports/a9de24c156e91c1b3ef60ccf.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/106/en.03.b3_theory_modals.questions.json`
  Report: `docs/grammar-review/reports/8f736e6f84c465c819c9cfe5.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/106/en.04.b4_theory_word_order.questions.json`
  Report: `docs/grammar-review/reports/7b596d54b4c0d6e04155282f.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/106/en.05.b5_theory_first_vs_zero.questions.json`
  Report: `docs/grammar-review/reports/52c8021f890eff27e23fd160.json`

## Следующая партия

- ES 105: 72 вопроса — косвенные вопросы.
- ES 106: 66 вопросов — письменный текст 120–180 слов со связками и относительными придаточными.
- ES 107: 66 вопросов — инфинитив после предлогов и рядом с союзными конструкциями.
- ES 108: 66 вопросов — gerundio: одновременное действие и образ действия.
- Итого: 270 вопросов в 4 файлах; reports отсутствуют.
- Все 16 назначенных source/final/validation и embedded-путей чистые и согласованы.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
