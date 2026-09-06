# ES verb forms: aspirar, atacar, atender

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `aspirar`: 96/96 карточек проверено, `fixed` 96.
- `atacar`: 96/96 карточек проверено, `fixed` 96.
- `atender`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок и слоты лица/числа сохранены; у `atender` исправлены
ошибочные scope/mood/tense последних четырёх блоков. Верхнеуровневые поля
сохранены. Источники точечно синхронизированы с `internal/grammartrainingpack`;
`make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 52 неверных
  `surface_form`: 19 в `aspirar`, 8 в `atacar`, 25 в `atender`.
- У `aspirar` исправлены presente de subjuntivo, pretérito perfecto de
  subjuntivo и оба блока futuro de subjuntivo.
- У `atacar` восстановлены `ataqué`, `atacáremos` и весь futuro perfecto de
  subjuntivo.
- У `atender` восстановлены четыре последних блока subjuntivo; исправлены 24
  сдвинутые координаты scope/mood/tense.
- В исходных файлах обнаружена 21 карточка с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены оборванные задания, неполные составные формы и буквальные русские
  переводы. Редкие времена явно отмечены как книжные или юридические употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-aspirar-atender/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24869 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
