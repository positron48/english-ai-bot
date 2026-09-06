# ES verb forms: conducir, confesar, confiar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `conducir`: 96/96 карточек проверено, `fixed` 96.
- `confesar`: 96/96 карточек проверено, `fixed` 96.
- `confiar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Для `confiar` системно соблюдено управление
`confiar en` и русское управление дательным падежом. Источники точечно
синхронизированы с `internal/grammartrainingpack`; `make training-pack`
не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 47 неверных
  `surface_form`: 12 в `conducir`, 16 в `confesar` и 19 в `confiar`.
- В `conducir` восстановлены формы futuro de subjuntivo с основой
  `condujer-` и futuro perfecto de subjuntivo с `hubiere`.
- В `confesar` исправлено чередование `e/ie` в presente de subjuntivo,
  futuro de subjuntivo и futuro perfecto de subjuntivo.
- В `confiar` восстановлены pretérito perfecto de indicativo, первое лицо
  pretérito anterior, futuro de subjuntivo и pretérito perfecto de subjuntivo.
- В исходных файлах обнаружено 127 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены шаблонные контексты, управление и переводы без нужного временного
  значения. Безличная формулировка об ответственности корректна для всех шести
  лиц. Редкие времена отмечены как книжные, юридические, транспортные или
  регламентные употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-conducir-confiar/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24974 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
