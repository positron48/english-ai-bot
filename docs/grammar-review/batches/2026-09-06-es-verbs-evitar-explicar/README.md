# ES verb forms: evitar, exigir, existir, explicar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `evitar`: 96/96 карточек проверено, `fixed` 96.
- `exigir`: 96/96 карточек проверено, `fixed` 96.
- `existir`: 96/96 карточек проверено, `fixed` 96.
- `explicar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `evitar`, `exigir` и `explicar` сверены с локальной базой Jehle. Формы
  `existir`, отсутствующего в Jehle, сверены с официальной
  [таблицей спряжения RAE DLE](https://dle.rae.es/existir#conjugaciongI18z6v).
- Исправлено 67 неверных `surface_form`: 19 в `evitar`, 12 в `exigir`, 24 в
  `existir`, 12 в `explicar`.
- У `evitar` восстановлены pretérito anterior, одна форма presente de
  subjuntivo и оба будущих времени subjuntivo. У `exigir` исправлены imperfecto
  и futuro de subjuntivo. У `existir` восстановлены pretérito anterior, futuro,
  pretérito perfecto и futuro perfecto de subjuntivo. У `explicar` исправлены
  futuro и futuro perfecto de subjuntivo.
- В исходных файлах обнаружено 49 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по устойчивым значениям: избежать риска, потребовать
  документ, существовать в заданной среде и объяснить материал. Русские
  переводы согласованы с лицом, числом и временем.
- Для непереходного `existir` написаны отдельные контексты для всех 16 scopes;
  обычные переходные шаблоны не использовались.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-evitar-explicar/check.py`
  — 384/384 карточек совпадают с Jehle или RAE DLE, контракты и fingerprints
  актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  нестандартные контексты `existir` проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25160 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
