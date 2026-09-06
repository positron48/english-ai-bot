# ES verb forms: afirmar, agradecer, alcanzar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `afirmar`: 96/96 карточек проверено, `fixed` 96.
- `agradecer`: 96/96 карточек проверено, `fixed` 96.
- `alcanzar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Все карточки получили `fixed`: для каждой вычитаны и переписаны испанский
контекст, русский перевод и варианты ответа. Идентичность (`scope`, `person`,
`number`), порядок и верхнеуровневые поля сохранены. Источники точечно
синхронизированы с `internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 53 неверных
  `surface_form`: 18 в `afirmar`, 23 в `agradecer`, 12 в `alcanzar`.
- В `afirmar` исправлены второе лицо presente de subjuntivo, pretérito perfecto
  и futuro perfecto de subjuntivo; будущая парадигма восстановлена полностью.
- В `agradecer` устранены формы `agrades`, `agradeceros` и конструкция
  `hubiera habido agradecido`; все четыре блока subjuntivo приведены к Jehle.
- В `alcanzar` futuro simple и futuro perfecto de subjuntivo больше не
  подменяются futuro indicativo.
- В исходных файлах обнаружено 46 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены падежи и вид в русских переводах, оборванные задания и контексты,
  которые не выражали нужное время. Редкие времена явно отмечены как книжные
  или юридические употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-afirmar-alcanzar/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24835 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
