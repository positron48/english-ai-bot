# ES verb forms: despedir, despertar, destacar, destruir

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `despedir`: 96/96 карточек проверено, `fixed` 96.
- `despertar`: 96/96 карточек проверено, `fixed` 96.
- `destacar`: 96/96 карточек проверено, `fixed` 96.
- `destruir`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел, координаты и верхнеуровневые поля,
включая `word_card_id` у `destacar`, сохранены. Источники точечно
синхронизированы с `internal/grammartrainingpack`; `make training-pack` не
запускался.

## Основные исправления

- Формы `despedir` и `despertar` сверены с локальной базой Jehle. Формы
  `destacar` и futuro de subjuntivo у `destruir` дополнительно сверены с
  официальными таблицами RAE: [destacar](https://dle.rae.es/destacar) и
  [destruir](https://dle.rae.es/destruir).
- Исправлено 66 неверных `surface_form`: 20 в `despedir`, 26 в `despertar`,
  14 в `destacar`, 6 в `destruir`.
- У `despedir` восстановлены presente, futuro и futuro perfecto de subjuntivo,
  а также отдельные формы pretérito indefinido и pretérito anterior.
- У `despertar` восстановлены futuro de subjuntivo и три составных scope
  subjuntivo; исправлены формы `despertemos` и `despertéis`.
- У `destacar` исправлены `destaqué`, `destacáremos`, pretérito perfecto и
  futuro perfecto de subjuntivo. У `destruir` условное наклонение, ошибочно
  записанное в futuro de subjuntivo, заменено на `destruyere` и остальные лица.
- В исходных файлах обнаружено 66 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная
  встречается ровно один раз.
- `Despedir` последовательно используется в переходном значении увольнения,
  поэтому карточки не требуют пропущенного возвратного местоимения. `Despertar`
  используется в переходном значении «будить». Для `destacar` выбрано значение
  подчёркивания результата, аргумента или вывода.
- У `destruir` убраны агрессивные бытовые и городские сценарии: контексты
  описывают регламентированное уничтожение дефектных документов, повреждённых
  копий, загрязнённых образцов и просроченных материалов.
- Редкие futuro и futuro perfecto de subjuntivo помещены в явно обозначенные
  формальные регламенты. Pretérito anterior используется только в книжном
  контексте.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-despedir-destruir/check.py`
  — 384/384 карточек совпадают с Jehle/RAE, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25078 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
