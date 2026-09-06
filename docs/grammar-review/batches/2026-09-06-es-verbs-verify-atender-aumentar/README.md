# ES verb forms verification: `atender`–`aumentar`

Дата: 2026-09-06.

## Результат

Независимо проверены 384 карточки: `atender`, `atraer`, `atravesar`, `aumentar`.
После repair и повторной проверки все четыре reports находятся в `phase: done`,
а все карточки имеют `verification: ok`.

## Найденные и исправленные ошибки

Проверка вернула 43 карточки отдельным repair-редакторам:

- по 7 в `atender`, `atraer`, `aumentar`: одинаковый субъект в конструкции
  `Me preocupa que yo...` и 6 незавершённых переводов futuro perfecto de
  subjuntivo;
- 22 в `atravesar`: 16 неточных переводов `atravesar el túnel`, одинаковый
  субъект и 6 незавершённых переводов futuro perfecto de subjuntivo.

Правки вносили участники, которые не проверяли соответствующие леммы. Исходные
verifiers повторно приняли каждое исправление. Дополнительно исправлены 24
устаревших report ID в `atender`; исходный verifier подтвердил совпадение всех
96 ID с координатами source.

## Проверки

- `check.py` — PASS: 384 карточки, native validation, identity/order, report ID,
  fingerprints и независимые editor/verifier.
- `preservation-check.py` — 25804 файла сохранены; 0 drift.
- Source и embedded всех четырёх лемм побайтно совпадают.

Финальные fingerprints:

- `atender`: `35ee990e400cefe2d35c7f64523b68f85bdd29750344cbd9926cb3cffeaef3ad`;
- `atraer`: `0db882a80a758c03a969c771863bbecd1022781926815b46826f86cc8803ed93`;
- `atravesar`: `175b91d2ffe22d43b65419b17e085f13cf9d037edd283cb33a65b02bb7ce61ba`;
- `aumentar`: `04541a092081e3a782ac32eaef7b5c4fdc35f9e6333588a647aa5b026a34d42e`.

Commit и push не выполнялись.
