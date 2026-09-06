# Редактура EN 123–126

2026-09-06 вычитано 433 вопроса в 24 основных и тренировочных банках.
Исправлено 146 вопросов, 287 оставлены без изменений. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 123 | связный пересказ разговора, новости и инструкций | 6 | 109 | 34 | 75 |
| EN 124 | do-emphasis | 6 | 106 | 29 | 77 |
| EN 125 | инверсия после negative/restrictive expressions | 6 | 109 | 31 | 78 |
| EN 126 | it-cleft, wh-cleft и reversed wh-cleft | 6 | 109 | 52 | 57 |
| Всего | | 24 | 433 | 146 | 287 |

Редактор: `Codex editor 2026-09-06`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. До начала партии все 56 назначенных
source/final/validation и embedded-путей были чистыми. Placement и отдельная
очередь форм глаголов ES исключены.

## Существенные исправления

- EN 123: общая схема пересказа больше не называет object говорящим; задания
  различают subject, адресата и содержание. Для backshift добавлена временная
  рамка. В десяти заданиях на команды и просьбы устранены по два-три допустимых
  ответа: требуемый reporting verb теперь явно указан. Исправлены слишком широкие
  утверждения об *asked to do* и *promise*.
- EN 124: обязательные *do/did* в вопросах и отрицаниях отделены от добавочного
  do-emphasis. Десять заданий блока контраста переписаны как утвердительные
  опровержения с ударным *do*. Устранены дублирующиеся правильные ответы при
  наречиях и модальных глаголах; отдельно оговорён emphatic imperative *Do be…*.
- EN 125: добавлен временной контекст для *Little do/did they know* и стандартной
  модели *No sooner + Past Perfect*. Уточнено, что инверсия после *not only* нужна
  при вынесении полной части в начало. В training-банках вторая часть после *but*
  больше не получает ложную инверсию, а нефронтированные *only*-обороты признаны
  грамматичными.
- EN 126: временная форма it-cleft больше не определяется одним прошедшим глаголом
  относительной части. Устранены двойные ключи с *which*, tag questions и
  вариативным согласованием. Basic wh-cleft отделена от reversed wh-cleft;
  reversed-примеры приведены к порядку *You are the person I trust*. Формальные
  *It was I/he who…* признаны допустимыми при правильном согласовании.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят
  для EN 123–126. Обычный AJV до и после сохраняет exit 1 из-за незагруженной
  meta-schema Draft 2020-12. Служебные `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 433 контракта и решения, fingerprints, ответы,
  theory bindings, training signatures, отсутствие новых course-wide дубликатов,
  сохранность validation и равенство source → final → embedded:
  [check.log](check.log), exit 0.
- Все 24 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [grammar-review-tests.log](grammar-review-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 4453 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 26916/29095 вопросов основных банков, из них 224
`done` и 26692 `awaiting_verification`; 2179 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 125–128: 282 вопроса в 4 файлах; reports отсутствуют. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники EN 123–126

### EN 123

Итого: 109/109 (`fixed` 34, `ok` 75); phase `awaiting_verification`.

- `chapter` 59 вопросов (`fixed` 5, `ok` 54)
  Source: `courses/english-grammar/chapters/123.en.grammar.reported_speech_meaning_transfer.build_write_retell_a/03-questions.json`
  Report: `docs/grammar-review/reports/e74439168dfcf71a79f666c1.json`
- `training` 10 вопросов (`fixed` 7, `ok` 3)
  Source: `courses/english-grammar/training_pack/chapters/123/en.01.b1_theory_retell_structure.questions.json`
  Report: `docs/grammar-review/reports/7c88b9135b8e6acb6dd2f5fd.json`
- `training` 10 вопросов (`fixed` 10, `ok` 0)
  Source: `courses/english-grammar/training_pack/chapters/123/en.02.b2_theory_mix_statements_questions.questions.json`
  Report: `docs/grammar-review/reports/75fc590f581734ef4495f606.json`
- `training` 10 вопросов (`fixed` 10, `ok` 0)
  Source: `courses/english-grammar/training_pack/chapters/123/en.03.b3_theory_commands_requests.questions.json`
  Report: `docs/grammar-review/reports/6aef91775d643c84ed92e14e.json`
- `training` 10 вопросов (`fixed` 0, `ok` 10)
  Source: `courses/english-grammar/training_pack/chapters/123/en.04.b4_theory_reporting_verbs_variety.questions.json`
  Report: `docs/grammar-review/reports/e5e4aee5bec051cc5893273a.json`
- `training` 10 вопросов (`fixed` 2, `ok` 8)
  Source: `courses/english-grammar/training_pack/chapters/123/en.05.b5_theory_cohesion.questions.json`
  Report: `docs/grammar-review/reports/b4d24937f2b9420ea81b53af.json`

### EN 124

Итого: 106/106 (`fixed` 29, `ok` 77); phase `awaiting_verification`.

- `chapter` 56 вопросов (`fixed` 5, `ok` 51)
  Source: `courses/english-grammar/chapters/124.en.grammar.information_structure_style.emphasis_do_emphasis/03-questions.json`
  Report: `docs/grammar-review/reports/db8ef550fc01b5e72f492a01.json`
- `training` 10 вопросов (`fixed` 1, `ok` 9)
  Source: `courses/english-grammar/training_pack/chapters/124/en.01.b1_theory_basic_emphasis.questions.json`
  Report: `docs/grammar-review/reports/c6230d98aef799ee6386ca18.json`
- `training` 10 вопросов (`fixed` 4, `ok` 6)
  Source: `courses/english-grammar/training_pack/chapters/124/en.02.b2_theory_past_emphasis.questions.json`
  Report: `docs/grammar-review/reports/12623e6ab5b119114b5da99a.json`
- `training` 10 вопросов (`fixed` 10, `ok` 0)
  Source: `courses/english-grammar/training_pack/chapters/124/en.03.b3_theory_contrast_correction.questions.json`
  Report: `docs/grammar-review/reports/421947328ecc79b0d5437a5c.json`
- `training` 10 вопросов (`fixed` 6, `ok` 4)
  Source: `courses/english-grammar/training_pack/chapters/124/en.04.b4_theory_position_limits.questions.json`
  Report: `docs/grammar-review/reports/9dc636440c3ee14f731e3c0c.json`
- `training` 10 вопросов (`fixed` 3, `ok` 7)
  Source: `courses/english-grammar/training_pack/chapters/124/en.05.b5_theory_emphatic_imperatives.questions.json`
  Report: `docs/grammar-review/reports/94932e122749ac466f091949.json`

### EN 125

Итого: 109/109 (`fixed` 31, `ok` 78); phase `awaiting_verification`.

- `chapter` 59 вопросов (`fixed` 8, `ok` 51)
  Source: `courses/english-grammar/chapters/125.en.grammar.information_structure_style.inversion_never_have_i/03-questions.json`
  Report: `docs/grammar-review/reports/60f6452b6a6899f1d68f577a.json`
- `training` 10 вопросов (`fixed` 3, `ok` 7)
  Source: `courses/english-grammar/training_pack/chapters/125/en.01.b1_theory_negative_adverbs.questions.json`
  Report: `docs/grammar-review/reports/ceb0d022659f4972709244ea.json`
- `training` 10 вопросов (`fixed` 10, `ok` 0)
  Source: `courses/english-grammar/training_pack/chapters/125/en.02.b2_theory_not_only.questions.json`
  Report: `docs/grammar-review/reports/862021cd630773b977ef3ba3.json`
- `training` 10 вопросов (`fixed` 6, `ok` 4)
  Source: `courses/english-grammar/training_pack/chapters/125/en.03.b3_theory_little_no_so.questions.json`
  Report: `docs/grammar-review/reports/5b6e262d9fe870f83149cff6.json`
- `training` 10 вопросов (`fixed` 2, `ok` 8)
  Source: `courses/english-grammar/training_pack/chapters/125/en.04.b4_theory_hardly_scarcely.questions.json`
  Report: `docs/grammar-review/reports/9a86ed95301c456cc83aeac0.json`
- `training` 10 вопросов (`fixed` 2, `ok` 8)
  Source: `courses/english-grammar/training_pack/chapters/125/en.05.b5_theory_emphasis_style.questions.json`
  Report: `docs/grammar-review/reports/40a5acad46acb28ef9bb93c0.json`

### EN 126

Итого: 109/109 (`fixed` 52, `ok` 57); phase `awaiting_verification`.

- `chapter` 59 вопросов (`fixed` 22, `ok` 37)
  Source: `courses/english-grammar/chapters/126.en.grammar.information_structure_style.cleft_sentences_it_s/03-questions.json`
  Report: `docs/grammar-review/reports/a32014968c67303cb00c2025.json`
- `training` 10 вопросов (`fixed` 6, `ok` 4)
  Source: `courses/english-grammar/training_pack/chapters/126/en.01.b1_theory_it_cleft_basic.questions.json`
  Report: `docs/grammar-review/reports/0fc008272ff14d604ced74e5.json`
- `training` 10 вопросов (`fixed` 0, `ok` 10)
  Source: `courses/english-grammar/training_pack/chapters/126/en.02.b2_theory_it_cleft_focus_types.questions.json`
  Report: `docs/grammar-review/reports/e917f7bd578ef7b1f5a849d9.json`
- `training` 10 вопросов (`fixed` 7, `ok` 3)
  Source: `courses/english-grammar/training_pack/chapters/126/en.03.b3_theory_wh_cleft_basic.questions.json`
  Report: `docs/grammar-review/reports/a31d04e5ce1564a3de5dd136.json`
- `training` 10 вопросов (`fixed` 10, `ok` 0)
  Source: `courses/english-grammar/training_pack/chapters/126/en.04.b4_theory_reversed_cleft.questions.json`
  Report: `docs/grammar-review/reports/557df49f57302610013e86f9.json`
- `training` 10 вопросов (`fixed` 7, `ok` 3)
  Source: `courses/english-grammar/training_pack/chapters/126/en.05.b5_theory_style_limits.questions.json`
  Report: `docs/grammar-review/reports/c3a60804e87108103f642f2a.json`

## Следующая партия

- ES 125: 70 вопросов — желания, рекомендации, эмоции и мнения.
- ES 126: 70 вопросов — *cuando, en cuanto, hasta que* с будущей отсылкой.
- ES 127: 70 вопросов — *para que, a fin de que, sin que*.
- ES 128: 72 вопроса — *antes/después de que* и смена подлежащего.
- Итого: 282 вопроса в 4 файлах; reports отсутствуют.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
