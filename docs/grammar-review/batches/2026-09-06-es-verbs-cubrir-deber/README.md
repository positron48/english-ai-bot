# ES verb forms: cubrir, cumplir, dar, deber

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `cubrir`: 96/96 карточек проверено, `fixed` 96.
- `cumplir`: 96/96 карточек проверено, `fixed` 96.
- `dar`: 96/96 карточек проверено, `fixed` 96.
- `deber`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел и верхнеуровневые поля сохранены. Источники
точечно синхронизированы с `internal/grammartrainingpack`; `make training-pack`
не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 53 неверных
  `surface_form`: 14 в `cubrir`, 6 в `cumplir`, 20 в `dar`, 13 в `deber`.
- В `cubrir` исправлены `cubrís`, pretérito anterior, futuro и futuro perfecto
  de subjuntivo.
- В `cumplir` восстановлены все формы futuro de subjuntivo и нормализованы 18
  меток `tense`.
- В `dar` исправлены формы imperfecto, futuro и condicional perfecto de
  indicativo, а также futuro de subjuntivo.
- В `deber` исправлено ударение в `debiéramos`, futuro и futuro perfecto de
  subjuntivo.
- В исходных файлах обнаружено 72 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Для `deber` последовательно используется значение обязанности с инфинитивом;
  составные времена переданы русскими конструкциями «пришлось / придётся».
- Контексты `cubrir` и `cumplir` приведены к естественным сочетаниям
  «покрыть поверхность» и `cumplir el encargo/requisito/obligación`.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-cubrir-deber/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная выборка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25038 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
