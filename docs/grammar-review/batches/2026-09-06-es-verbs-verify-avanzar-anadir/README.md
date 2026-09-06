# ES verb forms verification: `avanzar`–`añadir`

Дата: 2026-09-06.

## Результат

Независимо проверены 384 карточки: `avanzar`, `avisar`, `ayudar`, `añadir`.
После repair и повторной проверки все четыре reports находятся в `phase: done`,
а все карточки имеют `verification: ok`.

## Найденные и исправленные ошибки

Проверка вернула 28 карточек отдельным repair-редакторам: по 7 в каждой лемме.
Во всех четырёх исправлена конструкция `Me preocupa que yo...` с одинаковым
субъектом, а в шести карточках futuro perfecto de subjuntivo русский перевод
теперь явно передаёт завершённость действия к указанной дате или часу.

Правки вносили участники, которые не проверяли соответствующие леммы. Исходные
verifiers повторно приняли каждое исправление.

## Проверки

- `check.py` — PASS: 384 карточки, native validation, identity/order, report ID,
  fingerprints и независимые editor/verifier.
- `preservation-check.py` — 25817 файлов сохранены; 0 drift.
- Source и embedded всех четырёх лемм побайтно совпадают.

Финальные fingerprints:

- `avanzar`: `da86029ef4c941c6db974b9631a1236093b6841c7b3c9cd45051f5eda8824774`;
- `avisar`: `2d097876c86ef5ef2c0686b591bb2bf07dd61c7b262d7e13602fa0f7e5cd064e`;
- `ayudar`: `035f2c556a97d76f56d32004ada10e48ab444d9205eeb8e8f97e240e5e2941a8`;
- `añadir`: `a7ab7820f9cf8ed72e6e6df9976fcee04b0c8e8b08ba5278066966ae66c20bbb`.

Commit и push не выполнялись.
