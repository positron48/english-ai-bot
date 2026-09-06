# ES verb forms: colocar, combatir, comentar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `colocar`: 96/96 карточек проверено, `fixed` 96.
- `combatir`: 96/96 карточек проверено, `fixed` 96.
- `comentar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Для `combatir` русские переводы подобраны по
конкретному дополнению, чтобы завершённые времена передавали результат
естественно. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Формы `colocar` и `combatir` сверены с локальной базой Jehle. `comentar`
  отсутствует в Jehle, поэтому его полная парадигма сверена с
  [официальной таблицей RAE](https://dle.rae.es/comentar).
- Исправлено 54 неверных `surface_form`: 13 в `colocar`, 16 в `combatir`
  и 25 в `comentar`.
- В `colocar` исправлены `coloquéis`, весь futuro de subjuntivo и весь
  futuro perfecto de subjuntivo.
- В `combatir` исправлены `combatió`, `combatáis`, `combatan`,
  `combatiéramos`, futuro de subjuntivo и futuro perfecto de subjuntivo.
- В `comentar` восстановлены pretérito anterior, futuro de subjuntivo,
  pretérito perfecto de subjuntivo и futuro perfecto de subjuntivo; также
  исправлена форма `comentaras`.
- В исходных файлах обнаружено 59 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены шаблонные контексты и переводы без нужного временного значения.
  Редкие времена явно отмечены как книжные либо юридические/регламентные
  употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-colocar-comentar/check.py`
  — 288/288 карточек совпадают с Jehle/RAE, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24946 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
