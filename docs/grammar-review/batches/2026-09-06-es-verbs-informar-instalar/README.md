# ES verb forms: informar, iniciar, insistir, instalar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `informar`: 96/96 карточек проверено, `fixed` 96.
- `iniciar`: 96/96 карточек проверено, `fixed` 96.
- `insistir`: 96/96 карточек проверено, `fixed` 96.
- `instalar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 45 неверных
  `surface_form`: 7 в `informar`, 14 в `iniciar`, 13 в `insistir`, 11 в
  `instalar`.
- У `informar` и `instalar` восстановлены формы futuro perfecto de subjuntivo.
  У `iniciar` исправлены futuro, pretérito perfecto compuesto, pretérito
  anterior и futuro de subjuntivo.
- У `insistir` восстановлены все шесть форм pretérito anterior и pretérito
  perfecto de subjuntivo. У `instalar` дополнительно исправлены опечатки
  `instabas`, `instalarem` и смешение вариантов imperfecto de subjuntivo.
- В исходных файлах обнаружено 52 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по устойчивым управлениям: `informar sobre`, `iniciar`
  процесс, `insistir en` и `instalar` оборудование. Русские переводы согласованы
  с лицом, числом и временем.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-informar-instalar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; управления
  `informar sobre` и `insistir en` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25223 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
