# ES verb forms: diseñar, disfrutar, disminuir, disponer

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `diseñar`: 96/96 карточек проверено, `fixed` 96.
- `disfrutar`: 96/96 карточек проверено, `fixed` 96.
- `disminuir`: 96/96 карточек проверено, `fixed` 96.
- `disponer`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел, координаты и верхнеуровневые поля,
включая `word_card_id` у `disponer`, сохранены. Источники точечно
синхронизированы с `internal/grammartrainingpack`; `make training-pack` не
запускался.

## Основные исправления

- Формы `diseñar`, `disfrutar` и `disminuir` сверены с локальной базой Jehle.
  Полная таблица `disponer` сверена с официальным спряжением
  [RAE](https://dle.rae.es/disponer), поскольку этой леммы нет в Jehle.
- Исправлено 89 неверных `surface_form`: 1 в `diseñar`, 10 в `disfrutar`, 12 в
  `disminuir`, 66 в `disponer`.
- У `diseñar` исправлено ударение в `diseñáremos`. У `disfrutar` восстановлены
  отдельные формы futuro и весь futuro perfecto de subjuntivo. У `disminuir`
  восстановлены futuro и futuro perfecto de subjuntivo.
- У `disponer` исправлены pretérito indefinido, futuro simple, все составные
  времена с неправильным `disponido`/`disponible` и четыре последних scopes
  subjuntivo. Используются нормативные `dispuse`, `dispondré`, `dispuesto`,
  `dispusiere` и `hubiere dispuesto`.
- В исходных файлах обнаружены 74 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная
  встречается ровно один раз.
- Для `disfrutar` последовательно сохранено управление с `de`. Контексты
  `diseñar`, `disminuir` и `disponer` используют естественные переходные
  значения и согласованы с русскими переводами.
- Редкие futuro и futuro perfecto de subjuntivo помещены в явно обозначенные
  формальные регламенты. Pretérito anterior используется только в книжном
  контексте.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-disenar-disponer/check.py`
  — 384/384 карточек совпадают с Jehle/RAE, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25094 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
