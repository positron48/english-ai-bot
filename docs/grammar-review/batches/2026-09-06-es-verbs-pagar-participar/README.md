# ES verb forms: pagar, parar, parecer, participar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `pagar`: 96/96 карточек проверено, `fixed` 96.
- `parar`: 96/96 карточек проверено, `fixed` 96.
- `parecer`: 96/96 карточек проверено, `fixed` 96.
- `participar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 62 неверных
  `surface_form`: 24 в `pagar`, 14 в `parar`, 18 в `parecer`, 6 в `participar`.
- В `pagar` были перепутаны четыре времени subjuntivo. В `parar` исправлены
  отдельные простые и оба сложных будущих времени subjuntivo. У `parecer`
  восстановлены futuro simple, pretérito perfecto и futuro perfecto de
  subjuntivo; у `participar` — futuro simple de subjuntivo.
- В исходных файлах обнаружена 81 карточка с повторяющимися вариантами: 25 в
  `pagar`, 5 в `parar`, 6 в `parecer`, 45 в `participar`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты закрепляют оплату, остановку объектов, личное `parecer` с
  согласованной именной частью и участие в мероприятиях.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-pagar-participar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  согласование дополнений `parecer` проверено отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25349 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
