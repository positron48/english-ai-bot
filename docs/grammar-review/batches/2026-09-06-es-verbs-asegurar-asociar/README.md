# ES verb forms: asegurar, asistir, asociar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `asegurar`: 96/96 карточек проверено, `fixed` 96.
- `asistir`: 96/96 карточек проверено, `fixed` 96.
- `asociar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Идентичность (`scope`, `person`, `number`), порядок и
верхнеуровневые поля сохранены. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- `asistir` и `asociar` сверены с локальной базой Jehle. Отсутствующий в ней
  `asegurar` сверён по официальной таблице спряжения RAE:
  `https://dle.rae.es/asegurar`.
- Исправлено 38 неверных `surface_form`: 12 в `asegurar`, 20 в `asistir`, 6 в
  `asociar`.
- В `asegurar` восстановлены оба блока futuro de subjuntivo и исправлены шесть
  значений tense в pluscuamperfecto de subjuntivo.
- В `asistir` причастия без вспомогательного глагола заменены полными формами в
  трёх составных временах; исправлены формы futuro de subjuntivo.
- В `asociar` presente de subjuntivo больше не подменяет futuro de subjuntivo.
- В исходных файлах обнаружено 20 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены оборванные задания, буквальные русские переводы и сочетания без
  обязательных предлогов. Редкие времена явно отмечены как книжные или
  юридические употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-asegurar-asociar/check.py`
  — 288/288 карточек совпадают с RAE/Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24862 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
