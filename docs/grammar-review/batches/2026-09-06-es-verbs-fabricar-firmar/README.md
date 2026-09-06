# ES verb forms: fabricar, faltar, fijar, firmar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `fabricar`: 96/96 карточек проверено, `fixed` 96.
- `faltar`: 96/96 карточек проверено, `fixed` 96.
- `fijar`: 96/96 карточек проверено, `fixed` 96.
- `firmar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 53 неверных
  `surface_form`: 13 в `fabricar`, 10 в `faltar`, 12 в `fijar`, 18 в `firmar`.
- У `fabricar` исправлено лицо в pretérito anterior и оба будущих времени
  subjuntivo. У `faltar` исправлены ударения и формы futuro, condicional,
  presente и futuro perfecto de subjuntivo. У `fijar` восстановлены futuro и
  futuro perfecto de subjuntivo. У `firmar` исправлены futuro, pretérito
  perfecto и futuro perfecto de subjuntivo.
- В исходных файлах обнаружено 35 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по устойчивым значениям: изготовить изделие, пропустить
  занятие, установить значение и подписать документ. Русские переводы
  согласованы с лицом, числом и временем.
- Для `faltar` используются конструкции `faltar a clase/al turno/...` и
  отдельный контекст pluscuamperfecto de subjuntivo. После ручной проверки
  испанский и русский шаблоны были синхронизированы, а партия заново собрана из
  baseline.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-fabricar-firmar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS;
  нестандартные шаблоны `faltar` повторно проверены во всех шести лицах.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25178 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
