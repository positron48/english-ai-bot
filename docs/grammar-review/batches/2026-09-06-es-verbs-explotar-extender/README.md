# ES verb forms: explotar, exponer, expresar, extender

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `explotar`: 96/96 карточек проверено, `fixed` 96.
- `exponer`: 96/96 карточек проверено, `fixed` 96.
- `expresar`: 96/96 карточек проверено, `fixed` 96.
- `extender`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 55 неверных
  `surface_form`: 14 в `explotar`, 18 в `exponer`, 9 в `expresar`, 14 в
  `extender`.
- У `explotar` исправлены presente, futuro и futuro perfecto de subjuntivo. У
  `exponer` восстановлены весь pretérito indefinido (`expuse`, `expusiste` и
  т. д.) и оба будущих времени subjuntivo. У `expresar` исправлены ударения,
  одна форма presente и одна форма pluscuamperfecto de subjuntivo, а также
  futuro perfecto de subjuntivo. У `extender` исправлены presente и оба будущих
  времени subjuntivo.
- В исходных файлах обнаружено 87 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по устойчивым значениям: использовать возможность,
  излагать аргумент, выражать мнение и расширять охват. Русские переводы
  согласованы с лицом, числом и временем.
- После ручной проверки для `exponer` уточнён контекст pluscuamperfecto de
  subjuntivo, чтобы исключить повтор «изложить проблему — избежать проблемы»;
  партия после этого заново собрана из baseline.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-explotar-extender/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  уточнённый шаблон `exponer` повторно проверен во всех шести лицах.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25169 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
