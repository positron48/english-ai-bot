# ES verb forms verification: `alquilar`–`anunciar`

Дата: 2026-09-06.

## Результат

Независимо проверены четыре ранее отредактированных банка по 96 карточек:

- `alquilar` — verifier `edit_utilizar`, report `6097fcbb0eedc24adc99e46f`;
- `amenazar` — verifier `edit_vaciar`, report `1c0f9a97e4b6a9775de24b00`;
- `andar` — verifier `edit_valer`, report `0da887f8d8e7a7ac87b0fa62`;
- `anunciar` — verifier `Codex coordinator`, report `d15aa5af8e8c484f5a740a45`.

Все 384 карточки прочитаны с подставленными формами, русскими переводами и
вариантами. После repair и повторной проверки все 384 имеют `verification: ok`,
четыре reports находятся в `phase: done`.

## Найденные и исправленные ошибки

Проверка вернула 52 карточки отдельным repair-редакторам:

- `alquilar`: 7 карточек — одинаковый субъект и 6 незавершённых futuro perfecto
  de subjuntivo;
- `amenazar`: 13 карточек — 6 двусмысленных `sin demora`, одинаковый субъект и
  6 незавершённых futuro perfecto de subjuntivo;
- `andar`: 25 карточек — 6 неестественных завершённых контекстов, 12 неточных
  видовых и регистровых переводов, одинаковый субъект и 6 незавершённых futuro
  perfecto de subjuntivo;
- `anunciar`: 7 карточек — одинаковый субъект и 6 незавершённых futuro perfecto
  de subjuntivo.

Каждую правку выполнил участник, который не проверял эту лемму. Затем исходный
независимый verifier повторно принял исправленные карточки. Source и embedded
для всех четырёх лемм побайтно совпадают.

## Проверки

- `check.py` — PASS: 384 карточки, native `validate_artifact`, identity/order,
  fingerprints, reports и независимые editor/verifier.
- Точечный `grammar-review.py check --source` — `done` для четырёх файлов.
- `preservation-check.py` — 25760 файлов вне разрешённого набора сохранены;
  0 изменённых, пропавших и неожиданных новых.

Финальные fingerprints:

- `alquilar`: `25a02a6902c4075dae6efed69ad7abc9b2a2684fdfa188c92cbfdd4fb4a9b621`;
- `amenazar`: `8f14c3786d0d6198a1956aa0637be009140651d621f4b21a8c903a5ae9c38828`;
- `andar`: `1092afc95f3add09c5b042871976b34e626582015ec14d13e2eabe91dfafc874`;
- `anunciar`: `c285da3d10f28dd6629c1740fade5d00d5291f68dd3664cd4750232fa2c3769b`.

Commit и push не выполнялись.
