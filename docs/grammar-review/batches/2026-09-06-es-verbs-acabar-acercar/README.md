# ES verb forms: acabar, aceptar, acercar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `acabar`: 96/96 карточек проверено, `fixed` 96.
- `aceptar`: 96/96 карточек проверено, `fixed` 96.
- `acercar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Все карточки получили `fixed`: испанский контекст, русский перевод и варианты
ответа вычитаны и переписаны для каждой формы. Идентичность карточек (`scope`,
`person`, `number`), порядок и верхнеуровневые поля файлов сохранены. Источники
точечно синхронизированы с `internal/grammartrainingpack`; `make training-pack`
не запускался.

## Основные исправления

- Все 288 форм сверены с локальной базой Jehle. Исправлено 33 неверных
  `surface_form`: 15 в `acabar`, 12 в `aceptar`, 6 в `acercar`.
- В `acabar` исправлены ошибочные ударения, первое лицо pretérito anterior и
  оба будущих времени subjuntivo.
- В `aceptar` формы futuro indicativo, ошибочно стоявшие в двух блоках futuro
  de subjuntivo, заменены правильными формами `aceptare`/`hubiere aceptado`.
- В `acercar` futuro perfecto de subjuntivo больше не подменён pretérito
  perfecto de subjuntivo. Во всех блоках сохранены орфографические чередования
  `c` → `qu`, подтверждённые Jehle.
- В исходных карточках найдено 62 случая повторяющихся вариантов. Теперь в
  каждой карточке ровно четыре уникальные формы того же scope, правильная форма
  встречается один раз.
- У `acercar` устранено смешение переходного `acercar algo a algo` с возвратным
  `acercarse`: все 96 примеров последовательно тренируют переходную лемму.
- Обрывочные или неграмотные пары (`Yo _ acercar`, `_ acercado`, «если бы был
  время») заменены полноценными ES/RU контекстами. Редкие pretérito anterior и
  futuro de subjuntivo явно отмечены как книжные или юридические употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-acabar-acercar/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24818 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: все решения редактора записаны, независимый
  проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
