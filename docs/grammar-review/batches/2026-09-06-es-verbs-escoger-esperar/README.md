# ES verb forms: escoger, escribir, escuchar, esperar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `escoger`: 96/96 карточек проверено, `fixed` 96.
- `escribir`: 96/96 карточек проверено, `fixed` 96.
- `escuchar`: 96/96 карточек проверено, `fixed` 96.
- `esperar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 34 неверных
  `surface_form`: 2 в `escoger`, 18 в `escribir`, 8 в `escuchar`, 6 в
  `esperar`.
- У `escoger` исправлены формы `escogiéremos` и `escogiereis`. У `escribir`
  восстановлены futuro, pretérito perfecto и futuro perfecto de subjuntivo. У
  `escuchar` исправлены две формы futuro de subjuntivo и весь pretérito
  perfecto de subjuntivo. У `esperar` восстановлен futuro perfecto de
  subjuntivo.
- В исходных файлах обнаружены 67 карточек с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по значениям: выбирать объект, писать текст, слушать
  материал и ждать объект или результат. Русские переводы согласованы с лицом,
  числом и временем.
- Редкие futuro и futuro perfecto de subjuntivo помещены в формальные
  регламенты, pretérito anterior — в явно книжный контекст. Для `esperar`
  книжные и формальные примеры дополнительно перепроверены во всех шести лицах.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-escoger-esperar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; шаблоны
  `esperar` после уточнения повторно проверены для всех шести лиц.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25142 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
