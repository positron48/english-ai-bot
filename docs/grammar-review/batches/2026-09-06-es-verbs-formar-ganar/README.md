# ES verb forms: formar, fumar, funcionar, ganar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `formar`: 96/96 карточек проверено, `fixed` 96.
- `fumar`: 96/96 карточек проверено, `fixed` 96.
- `funcionar`: 96/96 карточек проверено, `fixed` 96.
- `ganar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 62 неверных
  `surface_form`: 10 в `formar`, 9 в `fumar`, 30 в `funcionar`, 13 в `ganar`.
- У `formar` и `fumar` исправлены futuro и futuro perfecto de subjuntivo. У
  `funcionar` восстановлены pretérito perfecto, pluscuamperfecto и pretérito
  anterior de indicativo, а также оба будущих времени subjuntivo. У `ganar`
  исправлено лицо в pretérito anterior и оба будущих времени subjuntivo.
- В исходных файлах обнаружена 81 карточка с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по устойчивым значениям: сформировать группу, курить
  конкретный предмет, работать в заданной роли и выиграть соревнование. Русские
  переводы согласованы с лицом, числом и временем.
- Для `funcionar` написаны ролевые контексты `como coordinador/enlace/...` и
  отдельные переводы futuro perfecto. Для `fumar` уточнён условный контекст
  регистрации нарушения. После ручной проверки партия заново собрана из
  baseline.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-formar-ganar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  нестандартные шаблоны `funcionar` и `fumar` повторно проверены во всех лицах.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25187 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
