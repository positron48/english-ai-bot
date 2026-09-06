# ES verb forms: desaparecer, desarrollar, desayunar, descansar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `desaparecer`: 96/96 карточек проверено, `fixed` 96.
- `desarrollar`: 96/96 карточек проверено, `fixed` 96.
- `desayunar`: 96/96 карточек проверено, `fixed` 96.
- `descansar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел, координаты и верхнеуровневые поля,
включая `word_card_id` у `descansar`, сохранены. Источники точечно
синхронизированы с `internal/grammartrainingpack`; `make training-pack` не
запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 42 неверных
  `surface_form`: 13 в `desaparecer`, 13 в `desarrollar`, 4 в `desayunar`,
  12 в `descansar`.
- У `desaparecer`, `desarrollar` и `descansar` восстановлены futuro и futuro
  perfecto de subjuntivo; дополнительно исправлены формы pretérito anterior у
  `desaparecer`, ударение в `desarrolléis` и отдельные формы `desayunar`.
- В исходных файлах обнаружено 75 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная
  встречается ровно один раз.
- Для `desaparecer` контексты согласованы с исчезновением с карты, из виду,
  реестра, файлов и системы. Для `desarrollar` используются естественные
  объекты: проект, предложение, прототип, стратегия, материалы и функции.
- У `desayunar` и `descansar` устранены неестественные и незавершённые фразы;
  места и обстоятельства согласованы с русским переводом.
- Редкие futuro и futuro perfecto de subjuntivo помещены в явно обозначенные
  формальные регламенты и протоколы. Pretérito anterior используется только в
  явно обозначенном книжном контексте.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-desaparecer-descansar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25062 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
