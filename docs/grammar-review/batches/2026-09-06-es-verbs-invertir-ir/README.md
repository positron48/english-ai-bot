# ES verb forms: invertir, investigar, invitar, ir

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `invertir`: 96/96 карточек проверено, `fixed` 96.
- `investigar`: 96/96 карточек проверено, `fixed` 96.
- `invitar`: 96/96 карточек проверено, `fixed` 96.
- `ir`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 60 неверных
  `surface_form`: 7 в `invertir`, 19 в `investigar`, 22 в `invitar`, 12 в `ir`.
- У `invertir` восстановлены основы `invirtier-` в imperfecto и futuro de
  subjuntivo. У `investigar` исправлены оба совершённых будущих времени и
  пропущенное причастие `investigado`.
- У `invitar` восстановлены лица presente, imperfecto, futuro, pretérito
  perfecto и futuro perfecto de subjuntivo. У неправильного `ir` исправлены все
  формы futuro simple и futuro perfecto de subjuntivo: `fuere` и `hubiere ido`.
- В исходных файлах обнаружено 90 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты закрепляют управления `invertir en`, расследование объекта,
  приглашение на событие и движение к месту. Русские переводы согласованы с
  лицом, числом и временем; у `invertir` явно указан подразумеваемый объект
  «средства».

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-invertir-ir/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; неправильные
  парадигмы `invertir` и `ir` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25241 файла вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
