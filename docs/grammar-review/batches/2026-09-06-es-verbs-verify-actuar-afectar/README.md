# ES verb forms verification: `actuar`–`afectar`

Дата: 2026-09-06.

## Результат

Независимо проверены четыре ранее отредактированных банка по 96 карточек:

- `actuar` — verifier `edit_utilizar`, report `6748c8dfd1d0b8a74de00903`;
- `admitir` — verifier `edit_vaciar`, report `e8c7c65b5c5e4acbcb5385db`;
- `advertir` — verifier `edit_valer`, report `2f91f9a7058edd3d76945041`;
- `afectar` — verifier `Codex coordinator`, report `4887f9691e031cf9727ad2f6`.

Все 384 карточки прочитаны с подставленными формами, русскими переводами и
вариантами. После repair и повторной проверки все 384 имеют `verification: ok`,
четыре reports находятся в `phase: done`.

## Найденные и исправленные ошибки

Проверка вернула 44 карточки отдельным repair-редакторам:

- `actuar`: 22 карточки, включая неестественное «выступать в фильме», один
  одинаковый субъект и 6 незавершённых переводов futuro perfecto de subjuntivo;
- `admitir`: 8 карточек — одинаковый субъект, тавтология и 6 незавершённых
  переводов futuro perfecto de subjuntivo;
- `advertir`: 7 карточек — одинаковый субъект и 6 неточных переводов futuro
  perfecto de subjuntivo, включая значение `se archivará`;
- `afectar`: 7 карточек — одинаковый субъект и 6 незавершённых переводов futuro
  perfecto de subjuntivo.

Каждую правку выполнил участник, который не проверял эту лемму. Затем исходный
независимый verifier повторно принял исправленные карточки. Source и embedded
для всех четырёх лемм побайтно совпадают.

## Проверки

- `check.py` — PASS: 384 карточки, native `validate_artifact`, identity/order,
  fingerprints, reports и независимые editor/verifier.
- Точечный `grammar-review.py check --source` — `done` для четырёх файлов.
- `preservation-check.py` — 25740 файлов вне разрешённого набора сохранены;
  0 изменённых, пропавших и неожиданных новых.
- Repair-журналы и точные findings сохранены в каталоге партии.

Финальные fingerprints:

- `actuar`: `3e01a15076a85fedb08675769edb2c626aaa7d08de1e92b2d77abe4f68c4d757`;
- `admitir`: `22f8e5ba109b2cf7da7d0df98db58517fae0007c6c4090dee30222a29f851208`;
- `advertir`: `f348716b6571da50158b00bfce1b03353b22ce1ca12fb684aecf8f47f297f7fe`;
- `afectar`: `35f2ed3f0bf462b1fa3f05a62e43e6a575473ef4a834d5b17ff3d6d12441a78c`.

Commit и push не выполнялись.
