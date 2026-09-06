# ES verb forms: cenar, cerrar, clasificar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `cenar`: 96/96 карточек проверено, `fixed` 96.
- `cerrar`: 96/96 карточек проверено, `fixed` 96.
- `clasificar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. В `cenar` используются естественные дополнения
без лишнего артикля. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 57 неверных
  `surface_form`: 14 в `cenar`, 13 в `cerrar` и 30 в `clasificar`.
- В `cenar` исправлены `cenas`, лицо pretérito anterior, futuro de subjuntivo и
  futuro perfecto de subjuntivo.
- В `cerrar` исправлена `cerréis`, futuro de subjuntivo и futuro perfecto de
  subjuntivo.
- В `clasificar` восстановлены 18 составных форм с `haber`, а также futuro de
  subjuntivo и futuro perfecto de subjuntivo.
- В исходных файлах обнаружено 63 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены лишние артикли, оборванные вопросы, неестественные дополнения и
  переводы без нужного временного значения. Редкие времена явно отмечены как
  книжные, юридические, регламентные, медицинские или системные употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-cenar-clasificar/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24932 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
