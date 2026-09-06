# ES verb forms: llegar, llevar, llover, lograr

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `llegar`: 96/96 карточек проверено, `fixed` 96.
- `llevar`: 96/96 карточек проверено, `fixed` 96.
- `llover`: 96/96 карточек проверено, `fixed` 96.
- `lograr`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `llegar`, `llevar` и `lograr` сверены с локальной базой Jehle. Полная
  парадигма `llover` сверена с
  [Diccionario de la lengua española RAE](https://dle.rae.es/llover): словарь
  отмечает основное безличное употребление и одновременно приводит формы всех
  лиц.
- Исправлено 46 неверных `surface_form`: 12 в `llegar`, 7 в `llevar`, 20 в
  `llover`, 7 в `lograr`.
- У `llegar` восстановлены futuro simple и futuro perfecto de subjuntivo. У
  `llevar` исправлены форма `llevaremos` и futuro simple de subjuntivo. У
  `llover` восстановлены личные формы presente, pretérito anterior и три
  времени subjuntivo. У `lograr` исправлены `lográremos` и futuro perfecto de
  subjuntivo.
- В исходных файлах обнаружено 58 карточек с повторяющимися вариантами: 27 в
  `llegar`, 3 в `llevar`, 16 в `llover`, 12 в `lograr`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты закрепляют управление `llegar a`, перенос объекта, достижение
  результата и литературную персонификацию `llover`. Все личные примеры
  `llover` прямо помещены в сказку, стихотворение или метафору, чтобы не
  смешивать их с обычным безличным описанием погоды.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-llegar-lograr/check.py`
  — 384/384 карточек совпадают с Jehle или RAE, контракты и fingerprints
  актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; варианты
  объектов и литературные контексты `llover` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25268 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
