# ES verb forms: entender, entrar, entregar, enviar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `entender`: 96/96 карточек проверено, `fixed` 96.
- `entrar`: 96/96 карточек проверено, `fixed` 96.
- `entregar`: 96/96 карточек проверено, `fixed` 96.
- `enviar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 58 неверных
  `surface_form`: 12 в `entender`, 19 в `entrar`, 13 в `entregar`, 14 в
  `enviar`.
- У `entender` восстановлены futuro и futuro perfecto de subjuntivo. У `entrar`
  исправлены pretérito anterior, одна форма imperfecto de subjuntivo и оба
  будущих времени subjuntivo.
- У `entregar` исправлены `entregáremos`, pretérito perfecto и futuro perfecto
  de subjuntivo. У `enviar` исправлены `enviemos`, `enviéis`, futuro и futuro
  perfecto de subjuntivo.
- В исходных файлах обнаружены 92 карточки с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по устойчивым значениям: понимать сообщение, входить в
  помещение, передавать документ и отправлять сообщение. Убраны пустые объекты,
  смешение `entrar` с «приехать» и неестественные конструкции доставки.
- Редкие futuro и futuro perfecto de subjuntivo помещены в формальные
  регламенты, pretérito anterior — в явно книжный контекст.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-entender-enviar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25134 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
