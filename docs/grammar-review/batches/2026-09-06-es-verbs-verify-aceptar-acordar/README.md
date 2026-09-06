# ES verb forms verification: `aceptar`–`acordar`

Дата: 2026-09-06.

## Результат

Независимо проверены четыре ранее отредактированных банка по 96 карточек:

- `aceptar` — verifier `edit_utilizar`, report `1dc2c119af76486dcdb8988b`;
- `acercar` — verifier `edit_vaciar`, report `94810e5c4ce81e3756613144`;
- `acompañar` — verifier `edit_valer`, report `7f62e70e5743a19a220a744f`;
- `acordar` — verifier `Codex coordinator`, report `b7b548ceac6a3b82717d54eb`.

Все 384 карточки прочитаны с подставленными формами, русскими переводами и
вариантами. Проверяющие отличаются от сохранённого editor. Финальный результат:
384 `verification: ok`, четыре reports в `phase: done`.

## Найденные и исправленные ошибки

Проверка вернула 26 карточек отдельным repair-редакторам:

- `aceptar`: 6 переводов futuro perfecto de subjuntivo без явно завершённого
  действия;
- `acercar`: 1 неестественный контекст с одинаковым субъектом и 6 переводов
  futuro perfecto de subjuntivo без явно завершённого действия;
- `acompañar`: 6 переводов futuro perfecto de subjuntivo без явно завершённого
  действия;
- `acordar`: 1 неестественный контекст с одинаковым субъектом и 6 переводов
  futuro perfecto de subjuntivo без явно завершённого действия.

Каждую правку выполнил участник, который не проверял эту лемму. После ремонта тот
же независимый verifier повторно принял исправленные карточки. Source и embedded
для всех четырёх лемм побайтно совпадают.

## Проверки

- `check.py` — PASS: 384 карточки, native `validate_artifact`, identity/order,
  fingerprints, reports и независимые editor/verifier.
- Точечный `grammar-review.py check --source` — `done` для четырёх файлов.
- `preservation-check.py` — 25730 файлов вне разрешённого набора сохранены;
  0 изменённых, пропавших и неожиданных новых.
- Repair-журналы и точные findings сохранены в каталоге партии.

Финальные fingerprints:

- `aceptar`: `bc0e31be227241c7e5c25c6078c18549232937164342bfa385db15df3c3149ec`;
- `acercar`: `66e493951e2d644743b88d39117311a70ab5e0b9653dc53a4de999a1c3700d66`;
- `acompañar`: `d00d428e551c65dbd282c99b428a721d10b871a929f1937031c99d9e63b196ee`;
- `acordar`: `70511ef970160c1912911ba1d29e1f115e82026bd7fbefa5477a3e77b51566db`.

Commit и push не выполнялись.
