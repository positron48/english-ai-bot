# ES verb forms: necesitar, negar, negociar, nevar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `necesitar`: 96/96 карточек проверено, `fixed` 96.
- `negar`: 96/96 карточек проверено, `fixed` 96.
- `negociar`: 96/96 карточек проверено, `fixed` 96.
- `nevar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `necesitar`, `negar` и `negociar` сверены с локальной базой Jehle. Для
  `nevar` использована полная парадигма и переходное значение из
  [Diccionario de la lengua española RAE](https://dle.rae.es/nevar), поскольку
  Jehle хранит только безличные формы третьего лица.
- Исправлено 39 неверных `surface_form`: 6 в `necesitar`, 0 в `negar`, 11 в
  `negociar`, 22 в `nevar`.
- У `necesitar` восстановлено futuro de subjuntivo, у `negociar` — оба будущих
  времени subjuntivo. У `nevar` исправлены личные формы настоящего времени,
  `nevó`, presente de subjuntivo и оба будущих времени subjuntivo.
- В исходных файлах обнаружено 105 карточек с повторяющимися вариантами: 34 в
  `necesitar`, 26 в `negar`, 12 в `negociar`, 33 в `nevar`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты закрепляют необходимость, отрицание и переговоры. Личные формы
  `nevar` даны в зафиксированном DLE переходном значении «покрывать белым»;
  объекты помещены в сказки, басни, романы, легенды и другие вымышленные тексты.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-necesitar-nevar/check.py`
  — 384/384 карточек совпадают с Jehle или DLE RAE, контракты и fingerprints
  актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  переходное употребление и полная парадигма `nevar` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25322 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
