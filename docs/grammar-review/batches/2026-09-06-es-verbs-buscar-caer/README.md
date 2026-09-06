# ES verb forms: buscar, caber, caer

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `buscar`: 96/96 карточек проверено, `fixed` 96.
- `caber`: 96/96 карточек проверено, `fixed` 96.
- `caer`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Для `caber` и `caer` использованы отдельные
смысловые шаблоны, соответствующие непереходному значению глаголов. Источники
точечно синхронизированы с `internal/grammartrainingpack`; `make training-pack`
не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 60 неверных
  `surface_form`: 14 в `buscar`, 31 в `caber` и 15 в `caer`.
- В `buscar` исправлены `buscasteis`, `buscarais`, futuro de subjuntivo и futuro
  perfecto de subjuntivo.
- В `caber` восстановлены нерегулярные `quepo`, вся парадигма pretérito,
  составные формы с `cabido`, futuro de subjuntivo и futuro perfecto de
  subjuntivo.
- В `caer` исправлены формы imperfecto de subjuntivo, futuro de subjuntivo и
  futuro perfecto de subjuntivo.
- В исходных файлах обнаружена 51 карточка с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены шаблонные и бессмысленные сочетания, а русские переводы различают
  процесс поиска и его завершение. Редкие времена явно отмечены как книжные,
  юридические, регламентные или отчётные употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-buscar-caer/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24904 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
