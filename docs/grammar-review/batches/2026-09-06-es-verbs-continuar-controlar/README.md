# ES verb forms: continuar, contribuir, controlar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `continuar`: 96/96 карточек проверено, `fixed` 96.
- `contribuir`: 96/96 карточек проверено, `fixed` 96.
- `controlar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 288 карточек, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок лиц и чисел и верхнеуровневые поля сохранены. Источники
точечно синхронизированы с `internal/grammartrainingpack`; `make training-pack`
не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 47 неверных
  `surface_form`: 14 в `continuar`, 8 в `contribuir` и 25 в `controlar`.
- В `continuar` восстановлены формы futuro de subjuntivo и futuro perfecto de
  subjuntivo, а также две формы imperfecto de subjuntivo.
- В `contribuir` исправлены `contribuís`, `contribuyere` и все шесть форм futuro
  perfecto de subjuntivo.
- В `controlar` восстановлены четыре последних scope subjuntivo: в исходнике их
  координаты и формы были циклически смещены. Также исправлено `controléis`.
- Нормализовано поле `tense` в 30 ранних карточках `continuar` и `contribuir`;
  у 24 карточек `controlar` исправлены `scope` и `tense`.
- В исходных файлах обнаружено 46 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Шаблонные контексты и переводы заменены примерами с нужным временным значением.
  Редкие времена явно поданы как книжные или регламентные употребления.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
18 контрольных точек: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-continuar-controlar/check.py`
  — 288/288 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная выборка всех 16 scopes для каждой леммы — 48 примеров, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех трёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25009 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
