# ES verb forms: almorzar, alquilar, amenazar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `almorzar`: 96/96 карточек проверено, `fixed` 96.
- `alquilar`: 96/96 карточек проверено, `fixed` 96.
- `amenazar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 19 неверных
  `surface_form`: 11 в `almorzar`, по 4 в `alquilar` и `amenazar`.
- В `almorzar` восстановлены чередование `o` → `ue`, формы imperfecto de
  subjuntivo и оба блока futuro de subjuntivo. Исправлен перевод: `almorzar`
  означает «обедать», а не «завтракать».
- В `alquilar` исправлены ударения и окончания в presente и futuro de
  subjuntivo, включая `alquilemos`, `alquiláremos` и `alquilareis`.
- В `amenazar` формы futuro de subjuntivo больше не подменяются futuro
  indicativo; контексты заменены на естественное `amenazar con retirarse`.
- В исходных файлах обнаружено 49 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлено 12 несогласованных координат mood/tense у `almorzar`. Редкие времена
  явно отмечены как книжные или юридические употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-almorzar-amenazar/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24834 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
