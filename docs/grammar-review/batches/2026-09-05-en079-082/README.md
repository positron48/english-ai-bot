# Редактура EN 079–082

2026-09-05 вычитано 495 вопросов: 255 основных и 240 тренировочных в 28
исходных файлах. Исправлено 190 вопросов, 305 оставлены без изменений. Все
исправления затрагивают редакторское содержание. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| EN 079 | adjective + preposition | 130 | 66 | 64 |
| EN 080 | фразовые глаголы и порядок дополнения | 118 | 48 | 70 |
| EN 081 | направления и описание местоположения | 120 | 35 | 85 |
| EN 082 | and, but, or, so и пунктуация | 127 | 41 | 86 |
| Всего | | 495 | 190 | 305 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. Все назначенные исходники до
начала партии были чистыми, reports отсутствовали. Placement исключён.

## Существенные исправления

- EN 079: исправлены все тренировочные шаблоны вида `good at at math`;
  русские формулировки теперь задают нужную оценку навыка. Убраны ложная
  маркировка британского `different to` как ненормативного и неоднозначные
  противопоставления `bored with / bored by`. Блок `useful for / suitable for`
  теперь содержит полноценные английские предложения с пропусками.
- EN 080: исправлены ошибочные ключи, где правильные `get over it` и
  `put up with it` отвергались в пользу `get it over` и `put it up with`.
  Правило о позиции местоимения ограничено разделяемыми фразовыми глаголами;
  для `look after him` и других неразделяемых связок восстановлен правильный
  порядок. Длинное дополнение описано как естественное предпочтение, а не
  абсолютный запрет другого грамматического порядка.
- EN 081: задания различают `across`, `across from`, `through` и `into` по
  заданному маршруту. Убраны ложные утверждения, что `near to` всегда
  невозможно и что `opposite the bank` неграмматично; неоднозначные задания
  получили явные ориентиры движения и положения.
- EN 082: исправлен союз в `I would join you, but I have a meeting`, разведены
  причинно-следственные и контрастные чтения. Правило о запятой сформулировано
  как нейтральная письменная норма, а не абсолют; упражнения на Oxford comma
  и баланс теперь прямо называют выбранный стиль или требуемое преобразование.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Для
тренировок штатно пересчитаны signatures; новых дубликатов во всём EN-курсе нет.

## Замечания к теории и baseline

- Теоретический блок EN 080 о позиции местоимения сформулирован без явного
  ограничения «для разделяемых фразовых глаголов». Соседний блок отдельно
  вводит неразделяемые глаголы; вопросы теперь явно проговаривают границу правила.
- Теория EN 079 допускает и `bored with`, и `bored by`, поэтому задания больше
  не выдают одну из этих грамматичных форм за безусловно неверную.
- Теория EN 082 использует корректные оговорки `обычно` и `часто` для запятых;
  эти оговорки теперь сохранены и в вопросах.
- Обычный вызов AJV без `--spec=draft2020` не загружает meta-schema Draft 2020-12
  и штатно возвращает exit 1. Явный Draft 2020 и `validate-chapter.sh` принимают
  все четыре главы; результаты до и после редактуры совпадают.

Теория не входит в текущий scope и не изменялась.

## Проверки и интеграция

- Все четыре главы собраны и проходят `validate-chapter.sh` без предупреждений.
  Явный AJV Draft 2020 проходит с exit 0; baseline-ошибка обычного AJV полностью
  совпадает до и после. Служебные validation-файлы возвращены в исходное состояние.
- [check.py](check.py) проверил 495 контрактов и решений против baseline,
  актуальные fingerprints, ответы, theory bindings, signatures, отсутствие новых
  дубликатов и равенство source → final → embedded. Итог: [check.log](check.log),
  exit 0.
- Четыре final-главы и 24 тренировочных файла точечно синхронизированы с embedded.
  Неопросные поля final сохранены, кроме штатного `updated_at` сборки.
- Все 28 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Три проверки
  `git diff --check` прошли: [diff-check.log](diff-check.log).
- Все 13 тестов механизма учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Контрольный снимок подтвердил сохранность всех 3348 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 18956/29095 вопросов основных банков, из них 224
`done` и 18732 `awaiting_verification`; 10139 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 081–084: 288 вопросов в 4 файлах. Все назначаемые исходники чистые, reports
отсутствуют. Полная цель обоих курсов остаётся активной.

## Закрытые исходники EN 079–082

- Source: `courses/english-grammar/chapters/079.en.grammar.prepositions_verb_patterns.adjective_preposition_afraid_of/03-questions.json`
  Report: `docs/grammar-review/reports/a5a5a8dd4184297702fc7bce.json`
  Редактура: 70/70 (`fixed` 16, `ok` 54); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/079/en.01.b1_adj_prep_emotions.questions.json`
  Report: `docs/grammar-review/reports/23bf538e5dbe8f234b424e89.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/079/en.02.b2_adj_prep_ability.questions.json`
  Report: `docs/grammar-review/reports/b6156495d59fe763ad606118.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/079/en.03.b3_adj_prep_interest.questions.json`
  Report: `docs/grammar-review/reports/e3e3031b90e12be0262181bc.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/079/en.04.b4_adj_prep_ready.questions.json`
  Report: `docs/grammar-review/reports/67f02cb7ca3b65b271815a44.json`
  Редактура: 10/10 (`fixed` 3, `ok` 7); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/079/en.05.b5_adj_prep_comparison.questions.json`
  Report: `docs/grammar-review/reports/2acc72a65f2a4643ebdd6b79.json`
  Редактура: 10/10 (`fixed` 7, `ok` 3); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/079/en.06.b6_adj_prep_useful.questions.json`
  Report: `docs/grammar-review/reports/4e3271073d7eda7053848996.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/080.en.grammar.prepositions_verb_patterns.phrasal_verbs_as_grammar/03-questions.json`
  Report: `docs/grammar-review/reports/2e545a9713d8d85ccff5778b.json`
  Редактура: 58/58 (`fixed` 21, `ok` 37); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/080/en.01.b1_phrasal_what_is.questions.json`
  Report: `docs/grammar-review/reports/272696d8ebbd009e7f08dadb.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/080/en.02.b2_phrasal_separable.questions.json`
  Report: `docs/grammar-review/reports/fdf05b1600e94c3a3a0bb19f.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/080/en.03.b3_phrasal_pronoun_rule.questions.json`
  Report: `docs/grammar-review/reports/7b363b8abaea91b4de840898.json`
  Редактура: 10/10 (`fixed` 2, `ok` 8); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/080/en.04.b4_phrasal_inseparable.questions.json`
  Report: `docs/grammar-review/reports/d7ea149633ed4dd29a7be73f.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/080/en.05.b5_phrasal_object_length.questions.json`
  Report: `docs/grammar-review/reports/89a806a599fa9ab5265c59ea.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/080/en.06.b6_phrasal_word_order_traps.questions.json`
  Report: `docs/grammar-review/reports/e8cced805803fa86b84ac0ab.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/081.en.grammar.prepositions_verb_patterns.build_speak_directions_location/03-questions.json`
  Report: `docs/grammar-review/reports/f02c082fb720d77506086a11.json`
  Редактура: 60/60 (`fixed` 8, `ok` 52); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/081/en.01.b1_directions_core_phrases.questions.json`
  Report: `docs/grammar-review/reports/30ab5c48ea8387ac323941df.json`
  Редактура: 10/10 (`fixed` 3, `ok` 7); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/081/en.02.b2_location_prepositions_review.questions.json`
  Report: `docs/grammar-review/reports/e11bcad3fccface1e7fd245b.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/081/en.03.b3_movement_prepositions_review.questions.json`
  Report: `docs/grammar-review/reports/6ef97ef63077cb3e5d3afe07.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/081/en.04.b4_asking_finding.questions.json`
  Report: `docs/grammar-review/reports/b21a1a8d6e62593ab95670e1.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/081/en.05.b5_landmarks_distance.questions.json`
  Report: `docs/grammar-review/reports/067d928fe02f8cb2db3be2fc.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/081/en.06.b6_build_speak_dialogue.questions.json`
  Report: `docs/grammar-review/reports/6b027b35293ecaaba49ffadd.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/082.en.grammar.complex_sentences_1_connecting.and_but_or_so/03-questions.json`
  Report: `docs/grammar-review/reports/0c1ae170176870561b1710cb.json`
  Редактура: 67/67 (`fixed` 10, `ok` 57); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/082/en.01.b1_coordination_and.questions.json`
  Report: `docs/grammar-review/reports/b86b5ef84f538d30a9c1df65.json`
  Редактура: 10/10 (`fixed` 2, `ok` 8); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/082/en.02.b2_coordination_but.questions.json`
  Report: `docs/grammar-review/reports/3dd4b664cc47ee9ade644d79.json`
  Редактура: 10/10 (`fixed` 1, `ok` 9); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/082/en.03.b3_coordination_or.questions.json`
  Report: `docs/grammar-review/reports/f2f4eaf1ef2815d1e4970450.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/082/en.04.b4_coordination_so.questions.json`
  Report: `docs/grammar-review/reports/8f2c6902c987afa4519c0352.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/082/en.05.b5_punctuation_comma.questions.json`
  Report: `docs/grammar-review/reports/c110b2a1c2c07197dbf4d171.json`
  Редактура: 10/10 (`fixed` 8, `ok` 2); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/082/en.06.b6_sentence_balance.questions.json`
  Report: `docs/grammar-review/reports/2ae1f2b81f8aa73ebf812494.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
