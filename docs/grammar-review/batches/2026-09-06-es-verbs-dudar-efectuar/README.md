# ES verb forms: dudar, durar, echar, efectuar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `dudar`: 96/96 карточек проверено, `fixed` 96.
- `durar`: 96/96 карточек проверено, `fixed` 96.
- `echar`: 96/96 карточек проверено, `fixed` 96.
- `efectuar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 63 неверных
  `surface_form`: 16 в `dudar`, 22 в `durar`, 13 в `echar`, 12 в `efectuar`.
- У `dudar` исправлены несуществующие `ducho`, `duhas`, `duddes`, форма
  `dudáremos` и составные времена subjuntivo. У `durar` исправлены лица в
  presente, pretérito и pretérito anterior, а также три последних scopes
  subjuntivo.
- У `echar` исправлено ударение в `echarais`, восстановлены futuro и futuro
  perfecto de subjuntivo. У `efectuar` исправлены futuro и futuro perfecto de
  subjuntivo.
- В исходных файлах обнаружены 52 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- `dudar` последовательно используется с `de`, `durar` — в значении
  «продержаться», `echar` — в сочетаниях `echar un vistazo/una mirada`,
  `efectuar` — с естественными формальными объектами: платежом, переводом,
  проверкой, операцией и необходимыми действиями.
- Редкие futuro и futuro perfecto de subjuntivo помещены в формальные
  регламенты, pretérito anterior — в явно книжный контекст.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-dudar-efectuar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25110 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
