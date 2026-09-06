# ES verb forms: admitir, advertir, afectar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `admitir`: 96/96 карточек проверено, `fixed` 96.
- `advertir`: 96/96 карточек проверено, `fixed` 96.
- `afectar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Все карточки получили `fixed`: для каждой вычитаны и переписаны испанский
контекст, русский перевод и варианты ответа. Идентичность (`scope`, `person`,
`number`), порядок и верхнеуровневые поля сохранены. Источники точечно
синхронизированы с `internal/grammartrainingpack`; `make training-pack` не запускался.

## Источники форм

- `admitir` и `advertir` сверены с локальной базой
  `resources/verbs/jehle_verb_database.csv`.
- `afectar` отсутствует в Jehle. Его полная парадигма сверена с официальной
  [таблицей спряжения RAE](https://dle.rae.es/afectar).

## Основные исправления

- Исправлено 45 неверных `surface_form`: 24 в `admitir`, 14 в `advertir`,
  7 в `afectar`.
- В `admitir` три составных времени indicativo содержали только `admitido` без
  вспомогательного `haber`; pretérito perfecto de subjuntivo был подменён
  imperfecto de subjuntivo.
- В `advertir` исправлены `advirtamos`, `advirtáis`, futuro simple и futuro
  perfecto de subjuntivo. Значение примеров унифицировано как «предупреждать».
- В `afectar` исправлены первое лицо pretérito anterior и все формы futuro
  simple de subjuntivo; 12 значений `tense` нормализованы по каноническим scopes.
- В исходных файлах обнаружены 42 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Исправлены оборванные задания, несовпадения лица и ошибочные переводы вроде
  «я предупрежу опасность» и «я признаю решение». Редкие времена явно помещены
  в книжный или юридический контекст.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-admitir-afectar/check.py`
  — 288/288 карточек совпадают с авторитетными парадигмами, контракты и
  fingerprints актуальны, source и embedded совпадают.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 24828 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
