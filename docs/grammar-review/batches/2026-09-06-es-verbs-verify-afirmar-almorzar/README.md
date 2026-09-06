# ES verb forms verification: `afirmar`–`almorzar`

Дата: 2026-09-06.

## Результат

Независимо проверены четыре ранее отредактированных банка по 96 карточек:

- `afirmar` — verifier `edit_utilizar`, report `3d8a6b9dfa2717b8ffb11970`;
- `agradecer` — verifier `edit_vaciar`, report `f14f851ec99e68ffe9e19f5c`;
- `alcanzar` — verifier `edit_valer`, report `cac1f3b103c1cfe918926785`;
- `almorzar` — verifier `Codex coordinator`, report `89392692d12dcc3bc9dbf0c5`.

Все 384 карточки прочитаны с подставленными формами, русскими переводами и
вариантами. После repair и повторной проверки все 384 имеют `verification: ok`,
четыре reports находятся в `phase: done`.

## Найденные и исправленные ошибки

Проверка вернула 58 карточек отдельным repair-редакторам:

- `afirmar`: 22 карточки — 16 неточных переводов `afirmar su inocencia`, один
  одинаковый субъект и 6 незавершённых futuro perfecto de subjuntivo;
- `agradecer`: 22 карточки — 16 контекстов с благодарностью самому субъекту,
  один одинаковый субъект и 6 незавершённых futuro perfecto de subjuntivo;
- `alcanzar`: 7 карточек — неестественный контекст и 6 незавершённых futuro
  perfecto de subjuntivo;
- `almorzar`: 7 карточек — одинаковый субъект и 6 незавершённых futuro perfecto
  de subjuntivo.

Каждую правку выполнил участник, который не проверял эту лемму. Затем исходный
независимый verifier повторно принял исправленные карточки. Source и embedded
для всех четырёх лемм побайтно совпадают.

## Проверки

- `check.py` — PASS: 384 карточки, native `validate_artifact`, identity/order,
  fingerprints, reports и независимые editor/verifier.
- Точечный `grammar-review.py check --source` — `done` для четырёх файлов.
- `preservation-check.py` — 25750 файлов вне разрешённого набора сохранены;
  0 изменённых, пропавших и неожиданных новых.
- Repair-журналы и точные findings сохранены в каталоге партии.

Финальные fingerprints:

- `afirmar`: `b0c5e86273ac75aa802f69c47d170cf5cc7eea735bde8d07a7a9266a04c75031`;
- `agradecer`: `b6b46221d66777b76fde373aa45780068bbd4a0d848224b24f3c13a681472c1e`;
- `alcanzar`: `633a99dea410ed86720417a2a0c28db80f6e2c16448172aba0b18eb7aa11088d`;
- `almorzar`: `a429805b7315b188d5811e117779380613b4dd11998734bdd2aafb775480684d`.

Commit и push не выполнялись.
