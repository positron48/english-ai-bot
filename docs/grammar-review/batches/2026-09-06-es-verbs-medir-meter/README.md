# ES verb forms: medir, mejorar, merecer, meter

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `medir`: 96/96 карточек проверено, `fixed` 96.
- `mejorar`: 96/96 карточек проверено, `fixed` 96.
- `merecer`: 96/96 карточек проверено, `fixed` 96.
- `meter`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `medir`, `merecer` и `meter` сверены с локальной базой Jehle. Лемма
  `mejorar` в Jehle отсутствует, поэтому её полная парадигма сверена с
  [Diccionario de la lengua española RAE](https://dle.rae.es/mejorar).
- Исправлено 63 неверных `surface_form`: 19 в `medir`, 13 в `mejorar`, 25 в
  `merecer`, 6 в `meter`.
- У `medir` восстановлены четыре времени subjuntivo и исправлены несуществующие
  формы `habere/haberen`. У `mejorar` исправлены `mejoraréis` и оба будущих
  времени subjuntivo.
- В `merecer` три последних блока были заполнены словом `merecida`; восстановлены
  pretérito perfecto, pluscuamperfecto и futuro perfecto de subjuntivo. У
  `meter` исправлена полная парадигма futuro simple de subjuntivo.
- В исходных файлах обнаружено 45 карточек с повторяющимися вариантами: 7 в
  `medir`, 10 в `mejorar`, 7 в `merecer`, 21 в `meter`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты закрепляют измерение объекта, улучшение результата, заслуженную
  оценку и конструкцию `meter … en …`; русские объекты у `meter` согласованы с
  глаголами «класть/положить» во всех временах.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-medir-meter/check.py`
  — 384/384 карточек совпадают с Jehle или RAE, контракты и fingerprints
  актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; сложные
  формы и пары объект/место у `meter` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25295 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
