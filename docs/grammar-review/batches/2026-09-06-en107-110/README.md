# Редактура EN 107–110

2026-09-06 вычитано 427 вопросов в 24 исходных файлах: 4 основных банка и
20 тренировочных блоков. Исправлено 152 вопроса, 275 оставлены без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 107 | second conditional | 6 | 106 | 46 | 60 |
| EN 108 | third conditional | 6 | 108 | 34 | 74 |
| EN 109 | mixed conditionals | 6 | 107 | 31 | 76 |
| EN 110 | `unless`, `provided`, `as long as` | 6 | 106 | 41 | 65 |
| Всего | | 24 | 427 | 152 | 275 |

Редактор: `Codex editor 2026-09-06`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. Все 56 назначенных
source/final/validation и embedded-путей до начала партии были чистыми. Placement и
отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- EN 107: заполнены пропущенные части шаблонов `would be`; формулировки на
  `would/could/might` теперь задают требуемый смысл. `Was` больше не объявляется
  безусловной ошибкой: `were` привязано к формальной нереальной модели. Исправлены
  задания, где несколько вариантов одновременно подходили без контекста.
- EN 108: исправлены прошлые условия с обладанием (`had had`), нелогичные причины и
  результаты, подсказка `started` и противоречивое отрицание. Тренировочные задания
  теперь различают базовый third conditional, mixed conditional и оттенки
  `could/might have`, не называя грамматичные альтернативы ошибками.
- EN 109: исправлены шесть неполных конструкций `would be/work`, Past Perfect
  `had known` и пунктуационные задания с двумя правильными ответами. Десять
  тренировочных вопросов past → present получили точный смысловой контекст;
  признаки present → past теперь описывают устойчивые нынешние свойства.
- EN 110: устранено ложное требование обязательного `that` после
  `provided/providing`. `Unless + not` корректно описано как конструкция, которая
  меняет смысл, а не автоматически становится неграмматичной. Исправлены `wait for`
  и десять неоднозначных заданий на `as long as`.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Signatures
изменённых тренировочных вопросов пересчитаны штатной функцией; новых дублей
signatures по всему EN training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят;
  обычный AJV до и после сохраняет exit 1 только из-за незагруженной meta-schema
  Draft 2020-12. Служебные `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 427 контрактов и решений, fingerprints, ответы,
  theory bindings, signatures, сохранность validation и равенство source → final →
  embedded: [check.log](check.log), exit 0.
- Все 24 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 4061 ранее изменённого стороннего
  файла: [preservation.log](preservation.log).

Общее редакторское покрытие: 24098/29095 вопросов основных банков, из них 224
`done` и 23874 `awaiting_verification`; 4997 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 109–112: 264 вопроса в 4 файлах; reports отсутствуют. Полная цель обоих курсов
остаётся активной.

## Закрытые исходники EN 107–110

### EN 107

Итого: 106/106 (`fixed` 46, `ok` 60); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/107.en.grammar.conditionals_hypotheticals.second_conditional/03-questions.json`
  Report: `docs/grammar-review/reports/36a6942692a92e37bc87c42e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/107/en.01.b1_theory_meaning.questions.json`
  Report: `docs/grammar-review/reports/825e4982cd8a513cb3a0cb5b.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/107/en.02.b2_theory_form.questions.json`
  Report: `docs/grammar-review/reports/8dc2c98b4460be9b5f5a7db4.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/107/en.03.b3_theory_were.questions.json`
  Report: `docs/grammar-review/reports/bc33818d662f74b23925c833.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/107/en.04.b4_theory_modals.questions.json`
  Report: `docs/grammar-review/reports/863be7cad865d3e6f770911a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/107/en.05.b5_theory_first_vs_second.questions.json`
  Report: `docs/grammar-review/reports/00f729478dce8e02fb9fcd97.json`

### EN 108

Итого: 108/108 (`fixed` 34, `ok` 74); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/108.en.grammar.conditionals_hypotheticals.third_conditional/03-questions.json`
  Report: `docs/grammar-review/reports/12d5b7d0d710848d3b9caf23.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/108/en.01.b1_theory_meaning.questions.json`
  Report: `docs/grammar-review/reports/8e23c399ba01d50800f68165.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/108/en.02.b2_theory_form.questions.json`
  Report: `docs/grammar-review/reports/eb1d3c0e47630954250ed2ca.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/108/en.03.b3_theory_modals.questions.json`
  Report: `docs/grammar-review/reports/28c3fdc3fa4917fd2fcc5d35.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/108/en.04.b4_theory_inversion.questions.json`
  Report: `docs/grammar-review/reports/0e48b4eb31b3d660e22a5a26.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/108/en.05.b5_theory_second_vs_third.questions.json`
  Report: `docs/grammar-review/reports/608636e588f537dcce247330.json`

### EN 109

Итого: 107/107 (`fixed` 31, `ok` 76); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/109.en.grammar.conditionals_hypotheticals.mixed_conditionals/03-questions.json`
  Report: `docs/grammar-review/reports/75ed29f9b5615dad491ec8d1.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/109/en.01.b1_theory_meaning.questions.json`
  Report: `docs/grammar-review/reports/17733715d22ad75213de90b3.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/109/en.02.b2_theory_past_to_present.questions.json`
  Report: `docs/grammar-review/reports/3a2b8a9510c137a8774a61bd.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/109/en.03.b3_theory_present_to_past.questions.json`
  Report: `docs/grammar-review/reports/d8596b298eb6742431fef7f9.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/109/en.04.b4_theory_word_order.questions.json`
  Report: `docs/grammar-review/reports/755e6618c8f72fe5c2db85aa.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/109/en.05.b5_theory_compare.questions.json`
  Report: `docs/grammar-review/reports/6c0103fa8d976a519f4985bf.json`

### EN 110

Итого: 106/106 (`fixed` 41, `ok` 65); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/110.en.grammar.conditionals_hypotheticals.unless_provided_as_long/03-questions.json`
  Report: `docs/grammar-review/reports/cc2d291d88ce81dec257db18.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/110/en.01.b1_theory_unless.questions.json`
  Report: `docs/grammar-review/reports/a2cb558e2c899b79afa86e64.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/110/en.02.b2_theory_provided.questions.json`
  Report: `docs/grammar-review/reports/298b0c31e0a60677080cf083.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/110/en.03.b3_theory_as_long_as.questions.json`
  Report: `docs/grammar-review/reports/2358c48c4e91746b2b8df54e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/110/en.04.b4_theory_tenses.questions.json`
  Report: `docs/grammar-review/reports/8fb52b43222654e3c351358d.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/110/en.05.b5_theory_contrast.questions.json`
  Report: `docs/grammar-review/reports/a43ffd8c9019c870bf900d2e.json`

## Следующая партия

- ES 109: 66 вопросов — причастия как прилагательные и состояния-результаты.
- ES 110: 66 вопросов — `seguir + gerundio`, `acabar de`, `volver a`, `dejar de`.
- ES 111: 66 вопросов — `soler`, `llevar + gerundio`, `terminar por`.
- ES 112: 66 вопросов — рутина, прогресс, прерывание и смена деятельности.
- Итого: 264 вопроса в 4 файлах; reports отсутствуют.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
