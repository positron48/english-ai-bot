# ES verb forms: crecer, creer, criticar, cruzar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `crecer`: 96/96 карточек проверено, `fixed` 96.
- `creer`: 96/96 карточек проверено, `fixed` 96.
- `criticar`: 96/96 карточек проверено, `fixed` 96.
- `cruzar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 46 неверных
  `surface_form`: по 12 в `crecer` и `creer`, 7 в `criticar`, 15 в `cruzar`.
- В `crecer` и `creer` восстановлены futuro и futuro perfecto de subjuntivo.
- В `criticar` исправлены `criticáremos` и весь futuro perfecto de subjuntivo.
- В `cruzar` исправлены pretérito anterior, presente и imperfecto de subjuntivo,
  futuro de subjuntivo и весь pretérito perfecto de subjuntivo.
- В исходных файлах обнаружено 112 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Для `creer` закреплено переходное значение «поверить версии/показаниям», чтобы
  не смешивать его с конструкцией `creer en`. Для `cruzar` русские переводы
  приведены к естественному «пересечь мост/дорогу/границу».
- Редкие времена явно поданы как книжные, юридические или регламентные
  употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-crecer-cruzar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная выборка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25030 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
