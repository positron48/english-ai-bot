# ES verb forms verification: `asistir`–`atacar`

Дата: 2026-09-06.

## Результат

Независимо проверены 384 карточки: `asistir`, `asociar`, `aspirar`, `atacar`.
После repair и повторной проверки все четыре reports находятся в `phase: done`,
а все карточки имеют `verification: ok`.

## Найденные и исправленные ошибки

Проверка вернула 46 карточек отдельным repair-редакторам:

- по 7 в `asistir`, `asociar`, `atacar`: одинаковый субъект в конструкции
  `Me preocupa que yo...` и 6 незавершённых переводов futuro perfecto de
  subjuntivo;
- 25 в `aspirar`: 6 неестественных мгновенных контекстов для `aspirar a`,
  6 неограниченных контекстов pretérito anterior, 6 неточных переводов futuro
  perfecto, одинаковый субъект и 6 незавершённых переводов futuro perfecto de
  subjuntivo.

Правки вносили участники, которые не проверяли соответствующие леммы. Исходные
verifiers повторно приняли каждое исправление.

## Проверки

- `check.py` — PASS: 384 карточки, native validation, identity/order,
  fingerprints и независимые editor/verifier.
- `preservation-check.py` — 25791 файл сохранён; 0 drift.
- Source и embedded всех четырёх лемм побайтно совпадают.

Финальные fingerprints:

- `asistir`: `369020f83def5a2105e1cdb1c46352fd2e2c91e4576840ce21e7982aecf5ab1b`;
- `asociar`: `039f8ec6d28fdc6f63bc4d8e3b4bdaf0ed1f62b812062f06e0a4bd8323a536a0`;
- `aspirar`: `e25d95c16be8249586397e2f29dd747f61f225f5a7faa4df1d08ed393ef96c21`;
- `atacar`: `e8bbb3a41c51187e1e95e387ae38fae68313991af7e80fbc89b9288d6e23c0df`.

Commit и push не выполнялись.
