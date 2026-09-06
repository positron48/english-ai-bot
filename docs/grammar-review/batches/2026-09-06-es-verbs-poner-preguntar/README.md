# ES verb forms: poner, practicar, preferir, preguntar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `poner`: 96/96 карточек проверено, `fixed` 96.
- `practicar`: 96/96 карточек проверено, `fixed` 96.
- `preferir`: 96/96 карточек проверено, `fixed` 96.
- `preguntar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Исходный порядок scopes, лиц и чисел и верхнеуровневые поля
сохранены. Источники точечно синхронизированы с `internal/grammartrainingpack`;
`make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 50 неверных
  `surface_form`: 18 в `poner`, 6 в `practicar`, 18 в `preferir`, 8 в
  `preguntar`.
- В исходных файлах обнаружена 51 карточка с повторяющимися вариантами: 25 в
  `poner`, 4 в `practicar`, 12 в `preferir`, 10 в `preguntar`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- В 18 карточках `preguntar` нормализованы ошибочные значения полей `mood` и
  `tense` в соответствии с их scope.
- Контексты последовательно используют `poner` для размещения предметов,
  переходный `practicar`, сравнение `preferir X a Y` и `preguntar por`.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-poner-preguntar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25385 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
