# ES verb forms: abandonar, abordar, abrir

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `abandonar`: 96/96 карточек проверено, `fixed` 96.
- `abordar`: 96/96 карточек проверено, `fixed` 96.
- `abrir`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Все карточки получили `fixed`, потому что для каждой были заново вычитаны и
переписаны испанский контекст, русский перевод и набор вариантов. Идентичность
карточек (`scope`, `person`, `number`) и порядок сохранены. Источники точечно
синхронизированы с `internal/grammartrainingpack`; `make training-pack` не запускался.

## Основные исправления

- Все 288 форм сверены с локальной базой Jehle. Исправлено 47 ошибочных
  `surface_form`: 9 в `abandonar`, 8 в `abordar`, 30 в `abrir`.
- В `abandonar` восстановлены формы множественного числа presente de subjuntivo
  и все шесть форм futuro simple de subjuntivo.
- В `abordar` исправлены `abordas`, `abordáremos` и все формы futuro perfecto
  de subjuntivo.
- В `abrir` восстановлены полные формы pretérito anterior, futuro perfecto и
  condicional perfecto; исправлены оба будущих времени subjuntivo.
- У `abrir` нормализованы 18 значений `tense` по каноническому `scope`; раньше
  три блока использовали несовместимые названия.
- Удалены повторяющиеся и морфологически невозможные варианты. Теперь в каждой
  карточке ровно четыре уникальные формы того же scope, а правильная встречается
  один раз.
- Неестественные заготовки наподобие `_ de Madrid`, обрывочные переводы и
  несовпадения лица заменены полноценными парами ES/RU. Редкие pretérito anterior
  и futuro de subjuntivo явно помещены в книжный или юридический контекст.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В
`checkpoints.jsonl` записано 18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-abandonar-abrir/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 4716 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: все 288 решений редактора записаны, независимый
  проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
