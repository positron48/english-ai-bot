# ES verb forms: ofrecer, olvidar, oponer, organizar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `ofrecer`: 96/96 карточек проверено, `fixed` 96.
- `olvidar`: 96/96 карточек проверено, `fixed` 96.
- `oponer`: 96/96 карточек проверено, `fixed` 96.
- `organizar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 39 неверных
  `surface_form`: 12 в `ofrecer`, 1 в `olvidar`, 13 в `oponer`, 13 в `organizar`.
- У `ofrecer` восстановлены pretérito perfecto compuesto и futuro de
  subjuntivo. У `oponer` исправлена одна форма presente и оба будущих времени
  subjuntivo. У `organizar` исправлено ударение imperfecto и оба будущих времени
  subjuntivo.
- В исходных файлах обнаружено 54 карточки с повторяющимися вариантами: 31 в
  `ofrecer`, 1 в `olvidar`, 19 в `oponer`, 3 в `organizar`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты закрепляют предложение помощи или условий, забывание данных,
  переходное `oponer` с возражениями и доводами и организацию мероприятий.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-ofrecer-organizar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25340 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
