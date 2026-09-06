# Редактура ES 097–100

2026-09-05 вычитано 330 вопросов в 10 исходных файлах: 4 основных банка и
6 тренировочных блоков. Исправлено 150 вопросов, 180 оставлены без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| ES 097 | безличное и пассивное `se` | 1 | 72 | 27 | 45 |
| ES 098 | `ser/estar` и состояние с причастием | 1 | 72 | 37 | 35 |
| ES 099 | бытовые происшествия, правила и короткие новости | 1 | 72 | 31 | 41 |
| ES 100 | придаточные с `que/si` | 7 | 114 | 55 | 59 |
| Всего | | 10 | 330 | 150 | 180 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. Все 41 назначенный
source/final/validation и embedded-путь до начала партии были чистыми. Placement и
отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- ES 097: разведены пассивное и безличное `se`, устранены двусмысленные пары вроде
  `se habla` / `se hablan`, исправлены согласование, личное `a`, ключ и ложные
  объяснения про неизменность причастия после `haber`.
- ES 098: устранено системное смешение аналитического пассива `ser + participio` с
  прилагательными вроде `conocido`. Ненормативное `La ley es mal redactada` заменено
  на естественное `está mal redactada`; исправлены задания с двумя допустимыми
  ответами про `hecho a mano`, расписание и состояние.
- ES 099: удалено ложное утверждение, будто `Se me prohíbe fumar` само по себе
  смешивает accidental и безличное `se`. Уточнены возвратные, местоименные и
  взаимные формы; исправлены двусмысленные продолжения и тексты объявлений.
- ES 100: тренировочные вопросы получили однозначный требуемый смысл вместо выбора
  между двумя грамматически правильными утверждениями. Исправлены `estar seguro de
  si`, временная логика косвенной речи и ложные исправления нормативного imperfecto.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Изменён
только ошибочный ключ одного тренировочного вопроса ES 100. Signatures изменённых
тренировочных вопросов пересчитаны штатной функцией; новых дублей signatures по
всему ES training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator сохраняет exit 0. Явный AJV
  Draft 2020 сохраняет исходный результат: exit 1 для ES 097–099 и exit 0 для ES
  100; обычный AJV сохраняет exit 1 из-за незагруженной meta-schema Draft 2020-12.
  Выводы до и после совпадают, служебные `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 330 контрактов и решений, fingerprints, ответы,
  theory bindings, signatures, сохранность validation и равенство source → final →
  embedded: [check.log](check.log), exit 0.
- Все 10 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 3828 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 22255/29095 вопросов основных банков, из них 224
`done` и 22031 `awaiting_verification`; 6840 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 099–102: 427 вопросов в 24 файлах. Reports отсутствуют; все 56 назначенных
source/final/validation и embedded-путей чистые и согласованы. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники ES 097–100

### ES 097

Итого: 72/72 (`fixed` 27, `ok` 45); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/097.es.grammar.se_system_pronominal_verbs.impersonal_se_passive_se/03-questions.json`
  Report: `docs/grammar-review/reports/9e24901930333c2ed384ea95.json`

### ES 098

Итого: 72/72 (`fixed` 37, `ok` 35); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/098.es.grammar.se_system_pronominal_verbs.ser_estar_contrast_result_state_participles/03-questions.json`
  Report: `docs/grammar-review/reports/216da75db1420dfff04da369.json`

### ES 099

Итого: 72/72 (`fixed` 31, `ok` 41); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/099.es.grammar.se_system_pronominal_verbs.build_write_everyday_incidents_rules_short_news_items/03-questions.json`
  Report: `docs/grammar-review/reports/3aea328a2833e1ed496be101.json`

### ES 100

Итого: 114/114 (`fixed` 55, `ok` 59); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/100.es.grammar.complex_sentences_connecting.que_si_noun_clauses_after_saying_thinking_asking/03-questions.json`
  Report: `docs/grammar-review/reports/820d4dc1e42c3a205ee8c072.json`
- `training` Source: `courses/spanish-grammar/training_pack/chapters/055/es.01.b1_theory_noun_clause_role_after_main_verbs.questions.json`
  Report: `docs/grammar-review/reports/e464aad6d6bf56951db16f7f.json`
- `training` Source: `courses/spanish-grammar/training_pack/chapters/055/es.02.b2_theory_que_after_saying_thinking.questions.json`
  Report: `docs/grammar-review/reports/33f8b8b79eac3781257008a8.json`
- `training` Source: `courses/spanish-grammar/training_pack/chapters/055/es.03.b3_theory_si_for_indirect_yes_no_questions.questions.json`
  Report: `docs/grammar-review/reports/617c3194589552e774fa283c.json`
- `training` Source: `courses/spanish-grammar/training_pack/chapters/055/es.04.b4_theory_direct_vs_indirect_question_forms.questions.json`
  Report: `docs/grammar-review/reports/c7d4b4ead6df5d7e059bd5cf.json`
- `training` Source: `courses/spanish-grammar/training_pack/chapters/055/es.05.b5_theory_tense_linking_in_noun_clauses.questions.json`
  Report: `docs/grammar-review/reports/6ced3dec0d5beff47fa5c7e6.json`
- `training` Source: `courses/spanish-grammar/training_pack/chapters/055/es.06.b6_theory_common_traps_que_si_reporting.questions.json`
  Report: `docs/grammar-review/reports/521c5df5596d9286baa5c8e3.json`

## Следующая партия

- EN 099: 109 вопросов в 6 файлах — пассив в разных временах.
- EN 100: 105 вопросов в 6 файлах — когда указывать исполнителя с `by`.
- EN 101: 107 вопросов в 6 файлах — `get`-passive.
- EN 102: 106 вопросов в 6 файлах — causative `have/get something done`.
- Итого: 427 вопросов в 24 файлах; reports отсутствуют.
- Все 56 назначенных source/final/validation и embedded-путей чистые и согласованы.
- Baseline HEAD EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
