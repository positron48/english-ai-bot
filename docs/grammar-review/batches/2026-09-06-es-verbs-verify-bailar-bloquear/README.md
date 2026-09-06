# ES verb forms verification: `bailar`–`bloquear`

Дата: 2026-09-06.

## Результат

Независимо проверены 384 карточки: `bailar`, `bajar`, `beber`, `bloquear`.
После repair и повторной проверки все четыре reports находятся в `phase: done`,
а все карточки имеют `verification: ok`.

## Найденные и исправленные ошибки

Проверка вернула 34 карточки отдельным repair-редакторам:

- по 7 в `bailar`, `bajar`, `bloquear`: одинаковый субъект в конструкции
  `Me preocupa que yo...` и 6 незавершённых переводов futuro perfecto de
  subjuntivo;
- 13 в `beber`: 6 переводов потеряли юридический регистр, одна конструкция
  получила внешний субъект и 6 переводов теперь явно передают завершённость
  futuro perfecto de subjuntivo.

Правки вносили участники, которые не проверяли соответствующие леммы. Исходные
verifiers повторно приняли каждое исправление. Дополнительно исправлены и
проверены 24 устаревших report ID в `beber`.

## Проверки

- `check.py` — PASS: 384 карточки, native validation, identity/order, report ID,
  fingerprints и независимые editor/verifier.
- `preservation-check.py` — 25830 файлов сохранены; 0 drift.
- Source и embedded всех четырёх лемм побайтно совпадают.

Финальные fingerprints:

- `bailar`: `cf80f9e42b0c3945a3b873e780470b2cd34b8c4af59558fd517df41019f42b8a`;
- `bajar`: `dc433b90462fcb7f1a73bfad26df6e258001206ae4d81bfb4e8385c24a0c9b3d`;
- `beber`: `154cbac495636d58a00cda8fc48e4b4ad1c5b960c0d9a2a20a328da549c5bc7d`;
- `bloquear`: `edfa0ca361417f584500b7d43c5ddbce9412b551ec99df1ff655b23419d38fc2`.

Commit и push не выполнялись.
