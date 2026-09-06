# ES verb forms: obligar, obtener, ocupar, ocurrir

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `obligar`: 96/96 карточек проверено, `fixed` 96.
- `obtener`: 96/96 карточек проверено, `fixed` 96.
- `ocupar`: 96/96 карточек проверено, `fixed` 96.
- `ocurrir`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `obligar` и `obtener` сверены с локальной базой Jehle. `Ocupar` сверено
  по полной парадигме [DLE RAE](https://dle.rae.es/ocupar), так как леммы нет в
  Jehle. Формы и личное значение `ocurrir` «обращаться к судье или власти»
  сверены по [DLE RAE](https://dle.rae.es/ocurrir), поскольку Jehle хранит для
  этой леммы только формы значения «происходить».
- Исправлено 79 неверных `surface_form`: 30 в `obligar`, 23 в `obtener`, 7 в
  `ocupar`, 19 в `ocurrir`.
- В `obligar` были перепутаны pretérito anterior и четыре времени subjuntivo.
  В `obtener` восстановлены pretérito anterior, presente perfecto и оба будущих
  времени subjuntivo. В `ocupar` восстановлено futuro de subjuntivo.
- Для `ocurrir` восстановлены личные формы presente и pretérito, а также оба
  будущих времени subjuntivo; все контексты используют подтверждённое личное
  управление `ocurrir a`.
- В исходных файлах обнаружено 111 карточек с повторяющимися вариантами: 23 в
  `obligar`, 34 в `obtener`, 20 в `ocupar`, 34 в `ocurrir`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-obligar-ocurrir/check.py`
  — 384/384 карточек совпадают с Jehle или DLE RAE, контракты и fingerprints
  актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; управление
  `ocurrir a` проверено отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25331 файла вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
