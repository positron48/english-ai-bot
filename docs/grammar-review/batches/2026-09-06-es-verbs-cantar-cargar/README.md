# ES verb forms: cantar, caracterizar, cargar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `cantar`: 96/96 карточек проверено, `fixed` 96.
- `caracterizar`: 96/96 карточек проверено, `fixed` 96.
- `cargar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Формы `caracterizar` и `cargar` сверены с локальной базой Jehle. `cantar`
  отсутствует в Jehle, поэтому его полная парадигма сверена с
  [официальной таблицей RAE](https://dle.rae.es/cantar).
- Исправлено 33 неверных `surface_form`: 6 в `cantar`, 13 в `caracterizar` и
  14 в `cargar`.
- В `cantar` восстановлен весь futuro de subjuntivo.
- В `caracterizar` исправлено лицо pretérito anterior, futuro de subjuntivo и
  futuro perfecto de subjuntivo.
- В `cargar` исправлены `cargaréis`, `cargarais`, futuro de subjuntivo и futuro
  perfecto de subjuntivo.
- В исходных файлах обнаружено 34 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены оборванные и шаблонные вопросы, неестественные дополнения и переводы
  без нужного временного значения. Редкие времена явно отмечены как книжные,
  юридические, конкурсные, отчётные или системные употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-cantar-cargar/check.py`
  — 288/288 карточек совпадают с авторитетными парадигмами, контракты и
  fingerprints актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24918 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
