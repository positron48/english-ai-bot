# ES verb forms: pensar, perder, permanecer, permitir

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `pensar`: 96/96 карточек проверено, `fixed` 96.
- `perder`: 96/96 карточек проверено, `fixed` 96.
- `permanecer`: 96/96 карточек проверено, `fixed` 96.
- `permitir`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 39 неверных
  `surface_form`: 0 в `pensar`, 24 в `perder`, 12 в `permanecer`, 3 в `permitir`.
- В `perder` были перепутаны четыре времени subjuntivo. У `permanecer`
  восстановлены оба будущих времени subjuntivo; у `permitir` исправлены одна
  форма pretérito anterior и две формы futuro de subjuntivo.
- В исходных файлах обнаружено 49 карточек с повторяющимися вариантами: 6 в
  `pensar`, 17 в `perder`, 9 в `permanecer`, 17 в `permitir`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты последовательно используют `pensar en`, потерю предметов,
  `permanecer en` и разрешение доступа или действий.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-pensar-permitir/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25367 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
