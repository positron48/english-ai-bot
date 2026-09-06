# ES verb forms verification: `aparecer`–`apoyar`

Дата: 2026-09-06.

## Результат

Независимо проверены четыре ранее отредактированных банка по 96 карточек:

- `aparecer` — verifier `edit_utilizar`, report `89de62208d14b43c5fe8d0ec`;
- `aplicar` — verifier `edit_vaciar`, report `920aef2eb16b243a945f39e1`;
- `apostar` — verifier `edit_valer`, report `783da360bf746916a9b697de`;
- `apoyar` — verifier `Codex coordinator`, report `10d25122aafc2b328ebab030`.

После полной вычитки, repair и повторной проверки все 384 карточки имеют
`verification: ok`; четыре reports находятся в `phase: done`.

## Найденные и исправленные ошибки

Проверка вернула 28 карточек отдельным repair-редакторам: по семь в каждой
лемме. Исправлены неестественные конструкции `Me preocupa que yo...` и русские
переводы futuro perfecto de subjuntivo, в которых не была явно выражена
завершённость действия. Каждую правку внёс участник, который не проверял эту
лемму; исходный verifier затем повторно принял исправления.

## Проверки

- `check.py` — PASS: 384 карточки, native `validate_artifact`, identity/order,
  fingerprints, reports и независимые editor/verifier.
- `preservation-check.py` — 25771 файл вне разрешённого набора сохранён;
  0 изменённых, пропавших и неожиданных новых.
- Source и embedded всех четырёх лемм побайтно совпадают.

Финальные fingerprints:

- `aparecer`: `6dfd941a9b6b6b042087117b92a3a48913d216cb03a2c15f91ab7b7b2f069e16`;
- `aplicar`: `b842af6c7b4796d4fae554fb641b7d6112aff0162e44fe8f70443519a726879b`;
- `apostar`: `47321c1e8413bc4f61977c141d6d6292efbb1b42ccf16bdb68aa4f7082b3f561`;
- `apoyar`: `be7403287df1d2f48fa8532021dfa87561e2545765c5f2f7455bc5fd5f40c50e`.

Commit и push не выполнялись.
