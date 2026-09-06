# Редактура EN 087–090

2026-09-05 вычитано 450 вопросов основных и тренировочных банков в 25 исходных файлах.
Исправлено 139 вопросов, 311 оставлены без изменений. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 087 | defining relative clauses | 6 | 99 | 43 | 56 |
| EN 088 | non-defining relative clauses | 6 | 109 | 43 | 66 |
| EN 089 | письменная практика раздела | 6 | 113 | 28 | 85 |
| EN 090 | форма Present Perfect | 7 | 129 | 25 | 104 |
| Всего | | 25 | 450 | 139 | 311 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline HEAD сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. EN 087, 088 и 090 были чистыми.
В EN 089 до начала партии уже были локальные изменения `03-questions.json`,
`05-final.json`, `05-validation.json` и embedded-копии; они вошли в побитовый
baseline и сохранены. Reports отсутствовали. Placement исключён.

## Существенные исправления

- EN 087: устранены десятки вопросов с несколькими правильными вариантами `who/that`
  или `which/that`. Исправлены висящие предлоги в примерах про фильм и письмо,
  неверный ключ с `what` и задания, где два разных исправления одновременно подходили
  под один `correct_answer`.
- EN 088: термины «непринадлежащий» и «неопределённый» заменены на точный
  `non-defining`. Исправлены неоднозначные контексты с запятыми, пропущенный `in`
  в `the park, in which we played`, примеры с `whose` и задания с несколькими
  ошибочными либо несколькими правильными ответами.
- EN 089: исправлены пунктуация `; therefore,`, non-defining clause в переводе про
  соседа, определение run-on sentence и смешанная русско-английская инструкция.
  Тренировки получили однозначные ключи и точные названия defining/non-defining clauses.
- EN 090: исправлены правила порядка в кратких ответах и образования сокращений,
  контекст для `ever` и вариативные позиции наречий. Десять оборванных упражнений
  have/has перестроены как полные конструкции `have/has + V3`, включая случаи,
  где прежний отвлекающий `has/have been` был грамматичнее ключа.

IDs, порядок, типы, difficulty, theory bindings, choice IDs и IDs правильных
ответов сохранены. Training signatures пересчитаны; новых дубликатов по курсу нет.

## Проверки и интеграция

- Все четыре главы собраны. `validate-chapter.sh` и явный AJV Draft 2020 проходят;
  обычный AJV сохраняет baseline exit 1 из-за незагруженной meta-schema Draft 2020-12.
  Результаты до и после каждой главы совпадают дословно.
- [check.py](check.py) проверил 450 контрактов и решений, fingerprints, ответы,
  theory bindings, signatures, source → final → embedded и восстановление входящего
  validation-файла EN 089: [check.log](check.log), exit 0.
- Четыре final-главы и 21 training-файл точечно синхронизированы с embedded-копиями.
- 25 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой независимой
  проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Три проверки
  `git diff --check` прошли: [diff-check.log](diff-check.log).
- Все 13 тестов механизма учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Контрольный снимок подтвердил сохранность всех 3555 сторонних файлов:
  [preservation.log](preservation.log).

Общее редакторское покрытие: 20429/29095 вопросов основных банков, из них 224
`done` и 20205 `awaiting_verification`; 8666 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 089–092: 288 вопросов в 4 файлах. Reports отсутствуют. В ES 090 и ES 092 уже
есть локальные изменения source/final и embedded-копий; перед началом их нужно
включить в baseline. Полная цель обоих курсов остаётся активной.

## Закрытые исходники EN 087–090

- Source: `courses/english-grammar/chapters/087.en.grammar.complex_sentences_1_connecting.relative_clauses_who_which/03-questions.json`
  Report: `docs/grammar-review/reports/2bb0c28030b375e1c8870bd5.json`
  Редактура: 49/49 (`fixed` 5, `ok` 44); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/087/en.01.b1_theory_defining_purpose.questions.json`
  Report: `docs/grammar-review/reports/ede66d7989edf61100d31500.json`
  Редактура: 10/10 (`fixed` 7, `ok` 3); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/087/en.02.b2_theory_pronoun_choice.questions.json`
  Report: `docs/grammar-review/reports/7d6c9bf463c55a6a184c0fd7.json`
  Редактура: 10/10 (`fixed` 9, `ok` 1); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/087/en.03.b3_theory_subject_object.questions.json`
  Report: `docs/grammar-review/reports/4865f4d9786dacc73153d3ee.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/087/en.04.b4_theory_prepositions.questions.json`
  Report: `docs/grammar-review/reports/5fa0a6b1d45614190c37ae29.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/087/en.05.b5_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/410ec5aa809add606b0fe875.json`
  Редактура: 10/10 (`fixed` 8, `ok` 2); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/088.en.grammar.complex_sentences_1_connecting.relative_clauses_commas_which/03-questions.json`
  Report: `docs/grammar-review/reports/d8b74d504afbf48722a74cab.json`
  Редактура: 59/59 (`fixed` 4, `ok` 55); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/088/en.01.b1_theory_non_defining_purpose_commas.questions.json`
  Report: `docs/grammar-review/reports/96c1ea9124141e7a40a8be2e.json`
  Редактура: 10/10 (`fixed` 9, `ok` 1); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/088/en.02.b2_theory_pronoun_choice_who_which.questions.json`
  Report: `docs/grammar-review/reports/742b30cfa52455f7cc04b7b4.json`
  Редактура: 10/10 (`fixed` 8, `ok` 2); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/088/en.03.b3_theory_no_omission.questions.json`
  Report: `docs/grammar-review/reports/654be27ee4202162687e7842.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/088/en.04.b4_theory_whose_possession.questions.json`
  Report: `docs/grammar-review/reports/db3130928161c494b7887666.json`
  Редактура: 10/10 (`fixed` 9, `ok` 1); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/088/en.05.b5_theory_common_errors_clause_which.questions.json`
  Report: `docs/grammar-review/reports/11c5e0ad55b5d341be513276.json`
  Редактура: 10/10 (`fixed` 3, `ok` 7); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/089.en.grammar.complex_sentences_1_connecting.build_write_a_120/03-questions.json`
  Report: `docs/grammar-review/reports/1b7904e680ebaba571b6ddcc.json`
  Редактура: 63/63 (`fixed` 7, `ok` 56); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/089/en.01.b1_theory_task_goal_structure.questions.json`
  Report: `docs/grammar-review/reports/4f9ec22fcb6f220196a9a655.json`
  Редактура: 10/10 (`fixed` 3, `ok` 7); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/089/en.02.b2_theory_connectors_purpose.questions.json`
  Report: `docs/grammar-review/reports/9644b57fdf4696bba4de2084.json`
  Редактура: 10/10 (`fixed` 1, `ok` 9); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/089/en.03.b3_theory_relative_clauses_in_text.questions.json`
  Report: `docs/grammar-review/reports/a7ad64a869267f6bd4635d5a.json`
  Редактура: 10/10 (`fixed` 8, `ok` 2); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/089/en.04.b4_theory_sentence_variety_punctuation.questions.json`
  Report: `docs/grammar-review/reports/85d5e460044751203afdbd59.json`
  Редактура: 10/10 (`fixed` 7, `ok` 3); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/089/en.05.b5_theory_editing_checklist.questions.json`
  Report: `docs/grammar-review/reports/795ecc99253fa761126ce36b.json`
  Редактура: 10/10 (`fixed` 2, `ok` 8); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/090.en.grammar.perfect_aspect_experience_result.present_perfect_form_have/03-questions.json`
  Report: `docs/grammar-review/reports/b7c6110c7801e0a08e7e0d23.json`
  Редактура: 69/69 (`fixed` 4, `ok` 65); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/090/en.01.b1_theory_core_form.questions.json`
  Report: `docs/grammar-review/reports/ca4761b80261fbde1a4151bd.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/090/en.02.b2_theory_have_vs_has.questions.json`
  Report: `docs/grammar-review/reports/4575128d6275d8b0d98eb36d.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/090/en.03.b3_theory_past_participle.questions.json`
  Report: `docs/grammar-review/reports/92e32e10fef0e2aa413f8c53.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/090/en.04.b4_theory_negatives_questions.questions.json`
  Report: `docs/grammar-review/reports/84d65530eab0d03013b3b139.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/090/en.05.b5_theory_short_answers_contractions.questions.json`
  Report: `docs/grammar-review/reports/efb2fb51617f59a1fa13b6ad.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/090/en.06.b6_theory_adverb_placement.questions.json`
  Report: `docs/grammar-review/reports/6eb2da9ee8f2a29e34c2a787.json`
  Редактура: 10/10 (`fixed` 6, `ok` 4); phase: `awaiting_verification`.

## Следующая партия

- ES 089: 72 вопроса — косвенные объектные местоимения `le/les`.
- ES 090: 72 вопроса — позиция местоимений перед глаголом и после него.
- ES 091: 72 вопроса — двойные местоимения `me lo`, `se lo`, `te la`.
- ES 092: 72 вопроса — базовая норма `le/lo/la` и leísmo.
- Итого: 288 вопросов в 4 файлах; reports отсутствуют.
- ES 089 и ES 091 чистые. В ES 090 и ES 092 уже изменены `03-questions.json`,
  `05-final.json` и соответствующие embedded-копии; final и embedded совпадают.
  Эти изменения считаются входным состоянием следующей партии.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
