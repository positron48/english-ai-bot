# ES verb forms: descender, describir, descubrir, desear

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `descender`: 96/96 карточек проверено, `fixed` 96.
- `describir`: 96/96 карточек проверено, `fixed` 96.
- `descubrir`: 96/96 карточек проверено, `fixed` 96.
- `desear`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел, координаты и верхнеуровневые поля
сохранены. Источники точечно синхронизированы с
`internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 86 неверных
  `surface_form`: 34 в `descender`, 4 в `describir`, 12 в `descubrir`, 36 в
  `desear`.
- У `descender` восстановлены чередование `e → ie`, составные формы последних
  трёх scopes и все формы futuro de subjuntivo; исправлены pretérito anterior и
  ударение в `descendiéramos`.
- У `describir` исправлены четыре формы futuro de subjuntivo. У `descubrir`
  восстановлены futuro и futuro perfecto de subjuntivo.
- У `desear` восстановлены три полных блока составных времён indicativo,
  futuro и futuro perfecto de subjuntivo, а также отдельные формы futuro,
  condicional и imperfecto de subjuntivo.
- В исходных файлах обнаружена 61 карточка с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная
  встречается ровно один раз.
- Для `descender` используются естественные направления движения; у
  `describir` и `descubrir` объекты согласованы со значением и русским
  переводом. `Desear` последовательно используется в конструкции пожелания
  кому-либо успеха, удачи, счастья или всего лучшего.
- Редкие futuro и futuro perfecto de subjuntivo помещены в явно обозначенные
  формальные регламенты. Pretérito anterior используется только в книжном
  контексте.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-descender-desear/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25070 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
