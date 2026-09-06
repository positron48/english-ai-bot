# ES verb forms: corresponder, cortar, costar, crear

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `corresponder`: 96/96 карточек проверено, `fixed` 96.
- `cortar`: 96/96 карточек проверено, `fixed` 96.
- `costar`: 96/96 карточек проверено, `fixed` 96.
- `crear`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел и верхнеуровневые поля сохранены. Источники
точечно синхронизированы с `internal/grammartrainingpack`; `make training-pack`
не запускался.

## Основные исправления

- `corresponder` отсутствует в локальной Jehle-базе. Его парадигма сверена с
  [официальной моделью второго спряжения RAE](https://www.rae.es/diccionario-estudiante/docs/conjugaciones-verbales.pdf).
  Формы `cortar`, `costar` и `crear` сверены с Jehle.
- Исправлено 67 неверных `surface_form`: 2 в `corresponder`, 24 в `cortar`,
  26 в `costar` и 15 в `crear`.
- В `corresponder` исправлены `correspondes` и `hubiereis correspondido`.
- В `cortar` восстановлены pluscuamperfecto и anterior de indicativo, futuro и
  pretérito perfecto de subjuntivo.
- В `costar` восстановлены четыре последних scope subjuntivo и формы `costemos`,
  `costéis`.
- В `crear` исправлены presente, futuro и futuro perfecto de subjuntivo.
- Нормализованы 24 ранние метки `tense` в `corresponder` и `cortar`; у 24
  карточек `costar` восстановлены правильные `scope` и `tense`.
- В исходных файлах обнаружено 87 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Для `costar` составные времена переданы завершёнными русскими формами
  «обошёлся / обойдётся», а контексты уточняют итоговую стоимость.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-corresponder-crear/check.py`
  — 384/384 карточек совпадают с авторитетными парадигмами, контракты и
  fingerprints актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная выборка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25022 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
