# ES verb forms: establecer, estar, estimar, estudiar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `establecer`: 96/96 карточек проверено, `fixed` 96.
- `estar`: 96/96 карточек проверено, `fixed` 96.
- `estimar`: 96/96 карточек проверено, `fixed` 96.
- `estudiar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Все формы сверены с локальной базой Jehle. Исправлено 56 неверных
  `surface_form`: 6 в `establecer`, 25 в `estar`, 11 в `estimar`, 14 в
  `estudiar`.
- У `establecer` восстановлен futuro de subjuntivo. У `estar` исправлена форма
  `estaréis`, восстановлен futuro de subjuntivo и добавлено `estado` во всех
  трёх составных временах subjuntivo. У `estimar` исправлены формы presente и
  imperfecto de subjuntivo, `estimáremos` и futuro perfecto de subjuntivo. У
  `estudiar` исправлены ударения, futuro и futuro perfecto de subjuntivo.
- В исходных файлах обнаружена 21 карточка с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- Контексты разделены по устойчивым значениям: установить правило, находиться
  в месте, оценить величину и изучить материал. Русские переводы согласованы с
  лицом, числом и временем.
- Для непереходного `estar` используются отдельные конструкции с местом и
  присутствием. Книжный pretérito anterior и оба формальных будущих времени
  subjuntivo дополнительно проверены во всех шести лицах.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-establecer-estudiar/check.py`
  — 384/384 карточек совпадают с Jehle, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; шаблоны
  `estar` для редких времён повторно проверены для всех шести лиц.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25151 файла вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
