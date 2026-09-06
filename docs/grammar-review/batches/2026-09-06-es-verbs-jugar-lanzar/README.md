# ES verb forms: jugar, juzgar, lamentar, lanzar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `jugar`: 96/96 карточек проверено, `fixed` 96.
- `juzgar`: 96/96 карточек проверено, `fixed` 96.
- `lamentar`: 96/96 карточек проверено, `fixed` 96.
- `lanzar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `jugar`, `lamentar` и `lanzar` сверены с локальной базой Jehle. Лемма
  `juzgar` в Jehle отсутствует, поэтому её полная парадигма сверена с
  [Diccionario de la lengua española RAE](https://dle.rae.es/juzgar).
- Исправлено 59 неверных `surface_form`: 25 в `jugar`, 12 в `juzgar`, 0 в
  `lamentar`, 22 в `lanzar`.
- У `jugar` восстановлены четыре последних времени subjuntivo и исправлено
  `jugarais`. У `juzgar` исправлены оба будущих времени subjuntivo по RAE.
- У `lanzar` восстановлены futuro perfecto и condicional perfecto de indicativo,
  формы imperfecto de subjuntivo и futuro perfecto de subjuntivo.
- В исходных файлах обнаружено 45 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты закрепляют управление `jugar a`, оценку объекта, сожаление о событии
  и запуск продукта или системы. Русские переводы согласованы с лицом, числом и
  временем.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-jugar-lanzar/check.py`
  — 384/384 карточек совпадают с зафиксированными парадигмами, контракты и
  fingerprints актуальны, source и embedded совпадают, дубликатов prompt+answer
  нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  орфографические изменения `jugar`, `juzgar` и `lanzar` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25250 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
