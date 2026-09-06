# ES verb forms: gobernar, guardar, guiar, gustar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `gobernar`: 96/96 карточек проверено, `fixed` 96.
- `guardar`: 96/96 карточек проверено, `fixed` 96.
- `guiar`: 96/96 карточек проверено, `fixed` 96.
- `gustar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлена 91 неверная
  `surface_form`: 18 в `gobernar`, 12 в `guardar`, 10 в `guiar`, 51 в `gustar`.
- У `gobernar` восстановлено чередование `gobierno/gobiernas/gobierne`,
  исправлены ударения и оба будущих времени subjuntivo. У `guardar` исправлены
  futuro и futuro perfecto de subjuntivo. У `guiar` исправлены формы `guie`,
  `guio`, presente и futuro de subjuntivo. У `gustar` восстановлены лица в
  простых временах, три составных времени indicativo и три времени subjuntivo.
- В исходных файлах обнаружено 58 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по устойчивым значениям: управлять территорией, сохранять
  объект, вести группу по маршруту и нравиться определённой аудитории. Русские
  переводы согласованы с лицом, числом и временем.
- Для `gustar` используется полная личная модель `gusto al público`, `gustas a
  los lectores`, `gustan a los visitantes`. Для `guiar` маршруты и адресаты
  дополнительно проверены во всех шести лицах.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-gobernar-gustar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  нестандартные модели `gustar`, `gobernar` и `guiar` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25196 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
