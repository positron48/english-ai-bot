# Редактура EN 095–098

2026-09-05 вычитано 488 вопросов в 27 исходных файлах: 4 основных банка и
23 тренировочных блока. Исправлено 229 вопросов, 259 оставлены без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 095 | Past Perfect и последовательность событий | 7 | 129 | 81 | 48 |
| EN 096 | Future Perfect / Future Continuous | 7 | 131 | 77 | 54 |
| EN 097 | биография, опыт и флэшбек | 7 | 121 | 50 | 71 |
| EN 098 | основы пассива `be + V3` | 6 | 107 | 21 | 86 |
| Всего | | 27 | 488 | 229 | 259 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. Все 62 назначенных
source/final/validation и embedded-пути до начала партии были чистыми. Placement и
отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- EN 095: устранено ложное требование Past Perfect после `before/after`; задания
  теперь явно просят выделить более раннее действие. Исправлены два неверных ключа
  для процесса в прошлом, разорванные пропуски, дубли вариантов и невозможные
  последовательности событий.
- EN 096: исправлен ключ `false` → `true` у значения `will be driving`. Разведены
  результат к сроку и процесс в будущий момент; убраны утверждения, будто `by` или
  точное время механически требуют Future Perfect / Future Continuous.
- EN 097: исправлен шаблон `I ___ already sent`, устранены альтернативно верные
  ответы с `for/since`, `already`, Past Simple и Past Perfect. Неоднозначный
  американский `gotten` заменён на однозначное `received`.
- EN 098: задания отделяют выбор пассива как фокуса от общей грамматичности активной
  формы. Исправлены неоднозначные времена с `every day`, объяснения про неизвестного
  исполнителя и вопросы, которые называли ошибочную V3 «непереходным глаголом».

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Изменены
только ошибочные ключи EN 095 и EN 096 и открытые ответы, где прежняя форма была
некорректной или неоднозначной. Signatures изменённых тренировочных вопросов
пересчитаны штатной функцией; новых дублей signatures по всему EN training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят;
  обычный AJV до и после сохраняет exit 1 только из-за незагруженной meta-schema
  Draft 2020-12. Выводы валидаторов до и после совпадают, служебные
  `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 488 контрактов и решений, fingerprints, ответы,
  theory bindings, signatures, сохранность validation и равенство source → final →
  embedded: [check.log](check.log), exit 0.
- Все 27 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 3740 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 21925/29095 вопросов основных банков, из них 224
`done` и 21701 `awaiting_verification`; 7170 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 097–100: 330 вопросов в 10 файлах. Reports отсутствуют; все назначенные source,
final и embedded-пути чистые и согласованы. Полная цель обоих курсов остаётся активной.

## Закрытые исходники EN 095–098

### EN 095

Итого: 129/129 (`fixed` 81, `ok` 48); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/095.en.grammar.perfect_aspect_experience_result.past_perfect_past_before/03-questions.json`
  Report: `docs/grammar-review/reports/0caa64e1aefd8c5c957b603c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/095/en.01.b1_theory_form_usage.questions.json`
  Report: `docs/grammar-review/reports/9787b2ac9fee19fd1dc7e83e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/095/en.02.b2_theory_sequence_before_past.questions.json`
  Report: `docs/grammar-review/reports/d0aedcf205aaa60be7f1bdd7.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/095/en.03.b3_theory_time_markers.questions.json`
  Report: `docs/grammar-review/reports/da2680a09abe83ee02a47bcf.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/095/en.04.b4_theory_contrast_past_simple.questions.json`
  Report: `docs/grammar-review/reports/3e0b2d85098d74a9786d82b6.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/095/en.05.b5_theory_questions_negatives.questions.json`
  Report: `docs/grammar-review/reports/81fb24e4e38c555cdaf8f0ec.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/095/en.06.b6_theory_narrative_context.questions.json`
  Report: `docs/grammar-review/reports/d25aeb35b7fd253a4ef0b221.json`

### EN 096

Итого: 131/131 (`fixed` 77, `ok` 54); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/096.en.grammar.perfect_aspect_experience_result.future_perfect_future_continuous/03-questions.json`
  Report: `docs/grammar-review/reports/825ed7b2c06c0469305f48f6.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/096/en.01.b1_theory_future_perfect_form.questions.json`
  Report: `docs/grammar-review/reports/9db0019d65f7bae86d16b8d5.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/096/en.02.b2_theory_future_perfect_result.questions.json`
  Report: `docs/grammar-review/reports/9ed1bba3b20eee38ff2afc50.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/096/en.03.b3_theory_future_continuous_form.questions.json`
  Report: `docs/grammar-review/reports/e8fbcd543465f632b53bf8ad.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/096/en.04.b4_theory_future_continuous_process.questions.json`
  Report: `docs/grammar-review/reports/911138707511488119a2387e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/096/en.05.b5_theory_contrast_future_forms.questions.json`
  Report: `docs/grammar-review/reports/7030c0944763ee8a2e341963.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/096/en.06.b6_theory_time_markers.questions.json`
  Report: `docs/grammar-review/reports/6512b99c45e2cdbbc34ff906.json`

### EN 097

Итого: 121/121 (`fixed` 50, `ok` 71); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/097.en.grammar.perfect_aspect_experience_result.build_speak_biography_experience/03-questions.json`
  Report: `docs/grammar-review/reports/042de519b6486aea86134d2c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/097/en.01.b1_theory_biography_frame.questions.json`
  Report: `docs/grammar-review/reports/6b51b8833ba36c218f0dc128.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/097/en.02.b2_theory_experience_markers.questions.json`
  Report: `docs/grammar-review/reports/60a280f4008532183c99d513.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/097/en.03.b3_theory_details_past_simple.questions.json`
  Report: `docs/grammar-review/reports/51db26afe700d02434481634.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/097/en.04.b4_theory_past_perfect_flashback.questions.json`
  Report: `docs/grammar-review/reports/e3fd5e22ddce47207fe7bb91.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/097/en.05.b5_theory_time_linkers.questions.json`
  Report: `docs/grammar-review/reports/c3389b26a94e5c4e2385dade.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/097/en.06.b6_theory_story_structure.questions.json`
  Report: `docs/grammar-review/reports/33e1a1d85719b4c3ea38a3d4.json`

### EN 098

Итого: 107/107 (`fixed` 21, `ok` 86); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/098.en.grammar.voice_focus_passive_and.passive_basics_be_v3/03-questions.json`
  Report: `docs/grammar-review/reports/0ed2db70c8d76dd679c67820.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/098/en.01.b1_theory_form_be_v3.questions.json`
  Report: `docs/grammar-review/reports/d591b7cb401c38116835c5b9.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/098/en.02.b2_theory_when_use.questions.json`
  Report: `docs/grammar-review/reports/ce8742424d59baeb31cc6374.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/098/en.03.b3_theory_active_to_passive.questions.json`
  Report: `docs/grammar-review/reports/02f44141669fceffc6a8686e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/098/en.04.b4_theory_be_tense_agreement.questions.json`
  Report: `docs/grammar-review/reports/09bee5893477ed6eafbf82bd.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/098/en.05.b5_theory_transitive_only.questions.json`
  Report: `docs/grammar-review/reports/8af4234c536240d3870a82a1.json`

## Следующая партия

- ES 097: 72 вопроса — безличное и пассивное `se`.
- ES 098: 72 вопроса — `ser/estar` и результат/состояние с причастиями.
- ES 099: 72 вопроса — связный текст о бытовых происшествиях, правилах и новостях.
- ES 100: 114 вопросов в 7 файлах — придаточные с `que/si` после речи, мысли и вопроса.
- Итого: 330 вопросов в 10 файлах; reports отсутствуют.
- Все назначенные source/final/embedded-пути чистые и сейчас совпадают.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
