# ES verb forms verification: `apreciar`–`asegurar`

Дата: 2026-09-06.

## Результат

Независимо проверены 384 карточки: `apreciar`, `aprender`, `aprobar`, `asegurar`.
После repair и повторной проверки все четыре reports находятся в `phase: done`,
а все карточки имеют `verification: ok`.

## Найденные и исправленные ошибки

Проверка вернула 43 карточки отдельным repair-редакторам:

- по 7 в `apreciar`, `aprender`, `aprobar`: одинаковый субъект в конструкции
  `Me preocupa que yo...` и 6 незавершённых переводов futuro perfecto de
  subjuntivo;
- 22 в `asegurar`: 16 тавтологических `asegurar la seguridad`, одинаковый
  субъект и 6 незавершённых переводов futuro perfecto de subjuntivo.

Правки вносили участники, которые не проверяли соответствующие леммы. Исходные
verifiers повторно приняли каждое исправление.

## Проверки

- `check.py` — PASS: 384 карточки, native validation, identity/order,
  fingerprints и независимые editor/verifier.
- `preservation-check.py` — 25781 файл сохранён; 0 drift.
- Source и embedded всех четырёх лемм побайтно совпадают.

Финальные fingerprints:

- `apreciar`: `9429eba1a5857fb63d046acbfb4799f87e56ea4c0b63a00df1fa05781ed4253e`;
- `aprender`: `e1fa9ad0b0fe14aa39b29d82b615731623b5e0c04dc42caec40116073c677c3a`;
- `aprobar`: `8a1645f6e879fe731c1ffb491d18f0673443ca733f19f32dacf31095f7e35d98`;
- `asegurar`: `420d6c119b1528cec82c7114718a5fc05eb5789795534744ee4182adde811f6d`.

Commit и push не выполнялись.
