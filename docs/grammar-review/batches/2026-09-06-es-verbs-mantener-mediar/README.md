# ES verb forms: mantener, marcar, matar, mediar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `mantener`: 96/96 карточек проверено, `fixed` 96.
- `marcar`: 96/96 карточек проверено, `fixed` 96.
- `matar`: 96/96 карточек проверено, `fixed` 96.
- `mediar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `mantener`, `marcar` и `matar` сверены с локальной базой Jehle. Лемма
  `mediar` в Jehle отсутствует, поэтому её полная парадигма сверена с
  [Diccionario de la lengua española RAE](https://dle.rae.es/mediar).
- Исправлено 80 неверных `surface_form`: 25 в `mantener`, 9 в `marcar`, 14 в
  `matar`, 32 в `mediar`.
- У `mantener` восстановлены четыре последних времени subjuntivo. У `marcar`
  исправлены формы condicional и futuro de subjuntivo. У `matar` исправлены
  ошибочные ударения и оба будущих времени subjuntivo.
- У `mediar` восстановлен participio `mediado`: в исходнике сложные формы были
  обрезаны до вспомогательного глагола либо ошибочно образованы от `medir`.
  Исправлены также ударения imperfecto и futuro de subjuntivo.
- В исходных файлах обнаружено 80 карточек с повторяющимися вариантами: 15 в
  `mantener`, 26 в `marcar`, 24 в `matar`, 15 в `mediar`. Теперь в каждой
  карточке четыре уникальные формы того же scope, правильная встречается ровно
  один раз.
- Контексты закрепляют сохранение состояния, нанесение отметки, лабораторное
  уничтожение насекомых и микроорганизмов и управление `mediar en/entre`.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-mantener-mediar/check.py`
  — 384/384 карточек совпадают с Jehle или RAE, контракты и fingerprints
  актуальны, source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS; управление
  `mediar` и формы из RAE проверены отдельно.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25286 файлов вне разрешённого набора — 0 изменённых, 0 пропавших,
  0 неожиданных новых файлов.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
