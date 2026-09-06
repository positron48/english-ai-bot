# ES verb forms: partir, pasar, pedir, pegar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `partir`: 96/96 карточек проверено, `fixed` 96.
- `pasar`: 96/96 карточек проверено, `fixed` 96.
- `pedir`: 96/96 карточек проверено, `fixed` 96.
- `pegar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- `Partir` сверено с полной парадигмой и переходным значением
  [DLE RAE](https://dle.rae.es/partir), поскольку леммы нет в Jehle. Остальные
  формы сверены с локальной базой Jehle.
- Исправлено 38 неверных `surface_form`: 19 в `partir`, 0 в `pasar`, 13 в
  `pedir`, 6 в `pegar`.
- У `partir` восстановлены pretérito anterior и три времени subjuntivo. У
  `pedir` исправлено ударение imperfecto и оба будущих времени subjuntivo; у
  `pegar` — отдельные формы presente и futuro de subjuntivo.
- В исходных файлах обнаружено 52 карточки с повторяющимися вариантами: 27 в
  `partir`, 3 в `pasar`, 11 в `pedir`, 11 в `pegar`. Теперь в каждой карточке
  четыре уникальные формы того же scope, правильная встречается ровно один раз.
- Контексты последовательно используют `partir` как деление на части, `pasar`
  как прохождение проверки, `pedir` как запрос и `pegar` как приклеивание.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-partir-pegar/check.py`
  — 384/384 карточек совпадают с Jehle или DLE RAE, контракты и fingerprints
  актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25358 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
