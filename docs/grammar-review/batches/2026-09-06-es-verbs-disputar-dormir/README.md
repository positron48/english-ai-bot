# ES verb forms: disputar, distinguir, distribuir, dormir

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `disputar`: 96/96 карточек проверено, `fixed` 96.
- `distinguir`: 96/96 карточек проверено, `fixed` 96.
- `distribuir`: 96/96 карточек проверено, `fixed` 96.
- `dormir`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `disputar` сверены с полной парадигмой RAE; остальные три леммы — с
  локальной базой Jehle. Исправлено 95 неверных `surface_form`: 30 в `disputar`,
  35 в `distinguir`, 13 в `distribuir`, 17 в `dormir`.
- У `disputar` восстановлены pretérito anterior, futuro de subjuntivo и три
  составных времени subjuntivo. У `distinguir` исправлены ударения в imperfecto
  и pretérito, futuro de subjuntivo и составные времена subjuntivo.
- У `distribuir` исправлены pretérito anterior первого лица и futuro и futuro
  perfecto de subjuntivo. У `dormir` исправлены первое лицо pretérito anterior,
  чередование основы в presente de subjuntivo, imperfecto de subjuntivo и оба
  будущих времени subjuntivo.
- В исходных файлах обнаружены 63 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты выдержаны в значениях «играть матч», «различать», «распределять» и
  «спать». Редкие futuro и futuro perfecto de subjuntivo помещены в формальные
  регламенты, pretérito anterior — в явно книжный контекст.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-disputar-dormir/check.py`
  — 384/384 карточек совпадают с Jehle/RAE, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25102 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
