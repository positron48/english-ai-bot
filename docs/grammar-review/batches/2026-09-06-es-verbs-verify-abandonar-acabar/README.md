# ES verb forms verification: `abandonar`–`acabar`

Дата: 2026-09-06.

## Результат

Независимо проверены четыре ранее отредактированных банка по 96 карточек:

- `abandonar` — verifier `edit_utilizar`, report `f0566d23c167dd3b93172290`;
- `abordar` — verifier `edit_vaciar`, report `ca3615ffec3de616c6d263f7`;
- `abrir` — verifier `edit_valer`, report `d05efeee10bddf44584c3f13`;
- `acabar` — verifier `Codex coordinator`, report `7f938579fdbb300050b29afc`.

Все 384 карточки прочитаны с подставленными формами, русскими переводами и
вариантами. Проверяющие отличаются от сохранённого editor. Финальный результат:
384 `verification: ok`, четыре reports в `phase: done`.

## Найденные и исправленные ошибки

Проверка вернула 58 карточек отдельным repair-редакторам:

- `abandonar`: 32 неверных русских управления для `abandonar el plan` и
  `abandonar la búsqueda`;
- `abordar`: 2 неестественных контекста и 6 переводов futuro perfecto de
  subjuntivo без явно завершённого действия;
- `abrir`: 6 переводов futuro perfecto de subjuntivo без явно завершённого
  действия;
- `acabar`: 12 шаблонных habitual-контекстов, которые описывали многократное
  завершение одного и того же определённого задания, проекта или курса.

Каждую правку выполнил участник, который не проверял эту лемму. После ремонта тот
же независимый verifier повторно принял исправленные карточки. Source и embedded
для всех четырёх лемм побайтно совпадают.

## Проверки

- `check.py` — PASS: 384 карточки, native `validate_artifact`, identity/order,
  fingerprints, reports и независимые editor/verifier.
- Точечный `grammar-review.py check --source` — `done` для четырёх файлов.
- `preservation-check.py` — 25720 файлов вне разрешённого набора сохранены;
  0 изменённых, пропавших и неожиданных новых.
- Repair-журналы и точные findings сохранены в каталоге партии.

Финальные fingerprints:

- `abandonar`: `dc8b1b1a0d2fa64fb2bfda7d5c8343203a394968370a2b9ecef46929773258f5`;
- `abordar`: `25740dd9319f14da3ff90841704ab1624cdb77f70fc56f15f6f9f16acd414496`;
- `abrir`: `99a5bdf605a84a1d7783c13a13033e6ff55a811d1b950f2c4d3392f6976a2df7`;
- `acabar`: `c3b530c81f7a37a05e9c9f0363ca99d9f0538e273968575f10fe02f0a7aa5528`.

Commit и push не выполнялись.
