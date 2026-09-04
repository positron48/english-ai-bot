# Точка продолжения вычитки

Обновлено: 2026-09-04.

- Пилот ES 001 + EN 001: 224 вопроса, 13 файлов done. Независимая проверка подтверждена пользователем: «я проверил, ок, правим дальше».
- Активная партия: ES 002–003, включая все тренировки: 250 вопросов в 14 файлах (один пустой).
- Редактор: Codex editor 2026-09-04; проверка новой партии пока не проводилась.
- Пользователь разрешил commit/push текущей вычитки и продолжение следующей партии. Placement и production исключены.
- Исходники новой партии до начала чистые. Предыдущие и сторонние изменения сохранены.
- Редактура завершена: 145 содержательных правок, 22 только signature, 83 без изменений.
- Обе главы собраны и валидны; встроенные копии совпадают; проверки контрактов/дубликатов, 13 тестов учёта и целевые Go-тесты прошли.
- Итог: [отчёт партии](batches/2026-09-04-es002-003/README.md). Все 14 файлов awaiting_verification; редакторских pending/blocked/needs_fix нет.

## Исходники и отчёты

- Source: `courses/spanish-grammar/chapters/002.es.grammar.orientation_alphabet_sounds.vowels_syllables_diphthongs_hiatus/03-questions.json`
  Report: `docs/grammar-review/reports/6d4a9c74b72e5baf68a1cd44.json`
  Редактура: 60/60; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/002/es.01.b1_theory_vowel_types.questions.json`
  Report: `docs/grammar-review/reports/756ef2eeb05bbf4070baa072.json`
  Редактура: 20/20; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/002/es.02.b2_theory_syllable_nucleus.questions.json`
  Report: `docs/grammar-review/reports/ead1a40341218fd0c969984d.json`
  Редактура: 17/17; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/002/es.03.b3_theory_basic_segmentation.questions.json`
  Report: `docs/grammar-review/reports/788cd0675544ee8b7f19fb61.json`
  Редактура: 16/16; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/002/es.04.b4_theory_diphthongs.questions.json`
  Report: `docs/grammar-review/reports/ead2b1930d3476dff7bfbf76.json`
  Редактура: 13/13; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/002/es.05.b5_theory_hiatus.questions.json`
  Report: `docs/grammar-review/reports/6e7e2d6dbeaba905443abd70.json`
  Редактура: 10/10; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/002/es.06.b6_theory_practical_algorithm.questions.json`
  Report: `docs/grammar-review/reports/47835e339e184e620c7f9515.json`
  Редактура: 12/12; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/chapters/003.es.grammar.orientation_alphabet_sounds.stress_rules_written_accents/03-questions.json`
  Report: `docs/grammar-review/reports/96dc542563c52c3b42df70dc.json`
  Редактура: 60/60; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/003/es.01.b1_theory_stressed_syllable.questions.json`
  Report: `docs/grammar-review/reports/65d4989e37bff2260bf74412.json`
  Редактура: 16/16; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/003/es.02.b2_theory_default_stress_vowel_n_s.questions.json`
  Report: `docs/grammar-review/reports/a769a3633b95f9bcdc2c3db4.json`
  Редактура: 1/1; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/003/es.03.b3_theory_default_stress_other_consonants.questions.json`
  Report: `docs/grammar-review/reports/317f6d525a6d428a7eb0347c.json`
  Редактура: 0/0; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/003/es.04.b4_theory_agudas_llanas_written_accent.questions.json`
  Report: `docs/grammar-review/reports/63a9a74bc357ce3c289109ba.json`
  Редактура: 3/3; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/003/es.05.b5_theory_esdrujulas_always_accented.questions.json`
  Report: `docs/grammar-review/reports/39de998aa45a5f96763972c6.json`
  Редактура: 12/12; следующий нерешённый ID: awaiting_verification; независимая проверка pending.
- Source: `courses/spanish-grammar/training_pack/chapters/003/es.06.b6_theory_reading_algorithm_and_traps.questions.json`
  Report: `docs/grammar-review/reports/2fd13d63d4134d3ec85b0c9f.json`
  Редактура: 10/10; следующий нерешённый ID: awaiting_verification; независимая проверка pending.

Следующая безопасная команда: `python3 scripts/grammar-review.py status`.
Следующее действие: независимая проверка всех 250 вопросов текущей партии и подтверждение пустого файла. Субагенты не разрешены; редактор не подменяет проверяющего. По поручению пользователя продолжить редактуру следующей партии EN 002–004 (282 вопроса), сохранив ожидание независимой проверки ES 002–003. Пилот повторно не редактировать.
