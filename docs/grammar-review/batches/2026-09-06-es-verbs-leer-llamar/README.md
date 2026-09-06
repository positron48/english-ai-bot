# ES verb forms: leer, levantar, limpiar, llamar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `leer`: 96/96 карточек проверено, `fixed` 96.
- `levantar`: 96/96 карточек проверено, `fixed` 96.
- `limpiar`: 96/96 карточек проверено, `fixed` 96.
- `llamar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлена 51 неверная
  `surface_form`: 12 в `leer`, 25 в `levantar`, 7 в `limpiar`, 7 в `llamar`.
- У `leer` восстановлены futuro simple и futuro perfecto de subjuntivo. У
  `levantar` исправлены четыре последних времени subjuntivo и форма `levantarais`.
  У `limpiar` исправлены `limpiáremos` и futuro perfecto de subjuntivo. У
  `llamar` восстановлен futuro simple de subjuntivo и исправлена форма
  `llamarais`.
- В исходных файлах обнаружено 68 карточек с повторяющимися вариантами: 11 в
  `leer`, 21 в `levantar`, 25 в `limpiar`, 11 в `llamar`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты закрепляют чтение конкретного материала, поднятие предмета, очистку
  объекта и управление `llamar a`. Русские переводы согласованы с лицом, числом
  и временем.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-leer-llamar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; контексты
  дополнительно проверены для крайних вариантов объектов.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25259 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
