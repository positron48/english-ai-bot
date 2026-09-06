# ES verb forms: beber, bloquear, borrar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `beber`: 96/96 карточек проверено, `fixed` 96.
- `bloquear`: 96/96 карточек проверено, `fixed` 96.
- `borrar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел и верхнеуровневые поля сохранены. В `beber`
исправлена ошибочная перестановка последних четырёх scopes при сохранении полного
набора идентификаторов. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Формы `beber` и `borrar` сверены с локальной базой Jehle. `bloquear` отсутствует
  в Jehle, поэтому его полная парадигма сверена с
  [официальной таблицей RAE](https://dle.rae.es/bloquear).
- Исправлено 49 неверных `surface_form`: 25 в `beber`, по 12 в `bloquear` и
  `borrar`.
- В `beber` исправлена диакритика `bebiéramos`; последние 24 карточки возвращены
  в канонический порядок futuro simple, pretérito perfecto, pluscuamperfecto и
  futuro perfecto de subjuntivo.
- В `bloquear` восстановлены все формы pretérito anterior и futuro de subjuntivo.
- В `borrar` исправлена форма `borraréis`, futuro de subjuntivo и futuro perfecto
  de subjuntivo.
- В исходных файлах обнаружено 22 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены оборванные вопросы, служебные подчёркивания, неестественные дополнения
  и переводы без нужного временного значения. Редкие времена явно отмечены как
  книжные, юридические, регламентные или медицинские употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-beber-borrar/check.py`
  — 288/288 карточек совпадают с авторитетными парадигмами, контракты и
  fingerprints актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24897 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
