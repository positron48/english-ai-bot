# ES verb forms: decidir, decir, declarar, dedicar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `decidir`: 96/96 карточек проверено, `fixed` 96.
- `decir`: 96/96 карточек проверено, `fixed` 96.
- `declarar`: 96/96 карточек проверено, `fixed` 96.
- `dedicar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел и верхнеуровневые поля сохранены. Источники
точечно синхронизированы с `internal/grammartrainingpack`; `make training-pack`
не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 40 неверных
  `surface_form`: 13 в `decidir`, 8 в `decir`, 13 в `declarar`, 6 в `dedicar`.
- В `decidir` исправлено `decidiéramos`, восстановлены futuro и futuro perfecto
  de subjuntivo.
- В `decir` исправлены pretérito anterior, futuro и futuro perfecto de
  subjuntivo; у последних четырёх scopes восстановлены правильные координаты.
- В `declarar` исправлено `declaremos`, восстановлены futuro и futuro perfecto
  de subjuntivo.
- В `dedicar` восстановлены все формы futuro de subjuntivo.
- В исходных файлах обнаружено 45 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Для `decir` выбраны естественные сочетания `decir la verdad/lo mismo/la frase`;
  у `declarar` сохранено единое значение декларирования дохода и имущества.
- Длительности `dedicar` ограничены часами и частями дня, поэтому они согласуются
  с контекстами «сегодня», «завтра» и «на этой неделе».

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-decidir-dedicar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная выборка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25046 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
