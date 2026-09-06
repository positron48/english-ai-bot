# ES verb forms: empezar, emplear, encontrar, enseñar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `empezar`: 96/96 карточек проверено, `fixed` 96.
- `emplear`: 96/96 карточек проверено, `fixed` 96.
- `encontrar`: 96/96 карточек проверено, `fixed` 96.
- `enseñar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 33 неверных
  `surface_form`: 1 в `empezar`, 12 в `emplear`, 13 в `encontrar`, 7 в
  `enseñar`.
- У `empezar` исправлено `empecéis`. У `emplear` восстановлены futuro и futuro
  perfecto de subjuntivo. У `encontrar` исправлены pretérito anterior,
  `encontrarais` и futuro de subjuntivo. У `enseñar` исправлены `enseñáremos` и
  futuro perfecto de subjuntivo.
- В исходных файлах обнаружены 78 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Удалены смысловые и языковые ошибки: «найти Мадрид», «использовать проблему»,
  незаполненные объекты и русское «учу́л». `Enseñar` переводится по объекту как
  «преподавать» или «объяснять».
- Редкие futuro и futuro perfecto de subjuntivo помещены в формальные
  регламенты, pretérito anterior — в явно книжный контекст.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-empezar-ensenar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25126 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
