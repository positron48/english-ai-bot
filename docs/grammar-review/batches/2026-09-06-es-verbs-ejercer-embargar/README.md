# ES verb forms: ejercer, elegir, eliminar, embargar

Дата редакторской проверки: 2026-09-06. Независимая проверка ожидается.

## Объём

- `ejercer`: 96/96 карточек проверено, `fixed` 96.
- `elegir`: 96/96 карточек проверено, `fixed` 96.
- `eliminar`: 96/96 карточек проверено, `fixed` 96.
- `embargar`: 96/96 карточек проверено, `fixed` 96.
- Всего: 384 карточки, 16 scopes × 6 лиц и чисел для каждой леммы.

Во всех карточках вычитаны и переписаны испанский контекст, русский перевод и
варианты ответа. Порядок scopes, лиц и чисел и верхнеуровневые поля сохранены.
Источники точечно синхронизированы с `internal/grammartrainingpack`; `make
training-pack` не запускался.

## Основные исправления

- Формы `ejercer`, `elegir` и `eliminar` сверены с локальной базой Jehle;
  `embargar` — с полной парадигмой RAE. Исправлено 90 неверных `surface_form`:
  20, 18, 20 и 32 соответственно.
- У `ejercer` восстановлены вспомогательные глаголы в составных временах,
  `ejerzamos`, `ejerzáis` и futuro de subjuntivo. У `elegir` исправлены futuro,
  pretérito perfecto и futuro perfecto de subjuntivo.
- У `eliminar` исправлены две формы imperfecto de subjuntivo и три следующих
  scopes. У `embargar` устранена подмена причастия `embargado` на `embarcado`,
  исправлены condicional, imperfecto de subjuntivo и последние четыре scopes.
- В исходных файлах обнаружена 41 карточка с повторяющимися вариантами. Теперь
  в каждой карточке четыре уникальные формы того же scope, правильная встречается
  ровно один раз.
- `ejercer` последовательно используется с функциями координации, контроля и
  представительства; `elegir` — с объектами выбора; `eliminar` — с удаляемыми
  объектами; `embargar` — только в юридическом значении ареста имущества.
- Редкие futuro и futuro perfecto de subjuntivo помещены в формальные
  регламенты, pretérito anterior — в явно книжный контекст.

## Контрольные точки

Решения сохранялись порциями по 12–18 карточек. В `checkpoints.jsonl` записано
24 контрольные точки: по шесть для каждой леммы.

## Проверки

- `python3 docs/grammar-review/batches/2026-09-06-es-verbs-ejercer-embargar/check.py`
  — 384/384 карточек совпадают с Jehle/RAE, контракты и fingerprints актуальны,
  source и embedded совпадают, дубликатов prompt+answer нет.
- Ручная проверка всех 16 scopes для каждой леммы — 64 примера, PASS.
- Строгий `validate_artifact()` генератора — PASS для всех четырёх лемм.
- `go test ./internal/verbtraining ./cmd/sync_verb_training_json` — PASS.
- `python3 scripts/grammar_review_test.py` — 13 тестов, PASS.
- `git diff --check` в корне и Spanish submodule — PASS.
- Сохранность 25118 файлов вне разрешённого набора — 0 изменённых, 0 пропавших.
- Scoped `grammar-review.py check` ожидаемо возвращает 1 только из-за
  `awaiting_verification`: независимый проверяющий ещё не назначен.

Baseline Spanish submodule: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
Commit и push в этой партии не выполнялись.
