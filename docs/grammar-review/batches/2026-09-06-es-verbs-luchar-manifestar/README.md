# ES verb forms: luchar, mandar, manejar, manifestar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `luchar`: 96/96 карточек проверено, `fixed` 96.
- `mandar`: 96/96 карточек проверено, `fixed` 96.
- `manejar`: 96/96 карточек проверено, `fixed` 96.
- `manifestar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `luchar`, `mandar` и `manejar` сверены с локальной базой Jehle. Лемма
  `manifestar` в Jehle отсутствует, поэтому её полная парадигма сверена с
  [Diccionario de la lengua española RAE](https://dle.rae.es/manifestar).
- Исправлено 67 неверных `surface_form`: 14 в `luchar`, 12 в `mandar`, 26 в
  `manejar`, 15 в `manifestar`.
- У `luchar` восстановлены pretérito anterior и оба будущих времени subjuntivo.
  У `mandar` исправлены pretérito anterior, `mandáremos` и futuro perfecto de
  subjuntivo. У `manejar` восстановлены четыре последних времени subjuntivo и
  исправлены `manejaréis`, `manejarais`. У `manifestar` восстановлено
  чередование `e` → `ie` в presente и исправлены три времени subjuntivo.
- В исходных файлах обнаружено 49 карточек с повторяющимися вариантами: 7 в
  `luchar`, 16 в `mandar`, 23 в `manejar`, 3 в `manifestar`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты закрепляют `luchar por`, отправку объекта, управление транспортом и
  выражение позиции. В русских переводах `manejar` использованы естественные
  конструкции «ездить на», а pretérito anterior явно завершает поездку.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-luchar-manifestar/check.py`
  — 384/384 карточек совпадают с Jehle или RAE, контракты и fingerprints
  актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  аспектуальные переводы `luchar` и `manejar` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25277 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
