# ES verb forms: comenzar, comer, compartir

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `comenzar`: 96/96 карточек проверено, `fixed` 96.
- `comer`: 96/96 карточек проверено, `fixed` 96.
- `compartir`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 19 неверных
  `surface_form`: 13 в `comenzar` и 6 в `compartir`; формы `comer`
  уже совпадали с парадигмой.
- В `comenzar` исправлена форма первого лица pretérito anterior, весь futuro
  de subjuntivo и весь futuro perfecto de subjuntivo.
- В `compartir` восстановлены все шесть форм futuro de subjuntivo:
  `compartiere`–`compartieren`.
- В исходных файлах обнаружено 19 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- У всех трёх лемм заменены шаблонные контексты и переводы без нужного временного
  значения. Редкие времена явно отмечены как книжные, юридические, медицинские
  или регламентные употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-comenzar-compartir/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24953 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
