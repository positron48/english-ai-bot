# ES verb forms: convencer, convertir, copiar, correr

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `convencer`: 96/96 карточек проверено, `fixed` 96.
- `convertir`: 96/96 карточек проверено, `fixed` 96.
- `copiar`: 96/96 карточек проверено, `fixed` 96.
- `correr`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 90 неверных
  `surface_form`: 20 в `convencer`, 21 в `convertir`, 19 в `copiar` и 30 в
  `correr`.
- В `convencer` исправлены формы imperfecto de indicativo, pretérito anterior,
  presente и futuro de subjuntivo, а также futuro perfecto de subjuntivo.
- В `convertir` восстановлены формы presente, imperfecto, futuro, pretérito
  perfecto и futuro perfecto de subjuntivo.
- В `copiar` исправлены pretérito anterior, futuro, pretérito perfecto и futuro
  perfecto de subjuntivo.
- В `correr` восстановлены все формы трёх составных прошедших времён indicativo,
  futuro и futuro perfecto de subjuntivo.
- В исходных файлах обнаружено 95 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены шаблонные контексты и переводы без нужного временного значения.
  Редкие времена явно поданы как книжные, юридические или регламентные
  употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-convencer-correr/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная выборка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25014 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
