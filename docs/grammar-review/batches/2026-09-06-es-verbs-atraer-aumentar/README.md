# ES verb forms: atraer, atravesar, aumentar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `atraer`: 96/96 карточек проверено, `fixed` 96.
- `atravesar`: 96/96 карточек проверено, `fixed` 96.
- `aumentar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 60 неверных
  `surface_form`: 38 в `atraer`, 15 в `atravesar`, 7 в `aumentar`.
- В `atraer` восстановлены irregular stems `traj-`, полные составные формы и
  оба блока futuro de subjuntivo.
- В `atravesar` исправлены чередование `e` → `ie` и формы futuro de subjuntivo.
- В `aumentar` исправлены `aumentáremos` и весь pretérito perfecto de subjuntivo.
- В исходных файлах обнаружена 21 карточка с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены оборванные задания, неверные вспомогательные глаголы и переводы,
  которые не выражали заданное время. Редкие времена явно отмечены как книжные
  или юридические употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-atraer-aumentar/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24876 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
