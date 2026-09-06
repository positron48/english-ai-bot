# Редактура EN 111–114

2026-09-06 вычитано 449 вопросов в 25 исходных файлах: 4 основных банка и
21 тренировочный блок. Исправлено 98 вопросов, 351 оставлен без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 111 | `wish`, `if only`, `would rather`, `it's time` | 6 | 107 | 33 | 74 |
| EN 112 | советы, сожаления и mixed conditional | 6 | 105 | 31 | 74 |
| EN 113 | gerund и to-infinitive: базовый выбор | 6 | 108 | 25 | 83 |
| EN 114 | глагольные модели и изменение смысла | 7 | 129 | 9 | 120 |
| Всего | | 25 | 449 | 98 | 351 |

Редактор: `Codex editor 2026-09-06`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. У EN 112 до начала партии уже была
синхронная правка q45 в source/final/validation/embedded; текущее согласованное
состояние принято как baseline и не включено в 98 новых исправлений. Остальные
54 из 58 назначенных source/final/validation и embedded-путей были чистыми.
Placement и отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- EN 111: восстановлены полные конструкции в заданиях с пропуском (`had read`,
  `had been`, `would be`) без нарушения однословного формата ответа. Исправлен
  неверный ключ в задании `I wish I were happy`, устранены конкурирующие
  правильные варианты в блоках `wish + would` и `would rather`, исправлены
  объяснения о функции Past Simple после `wish`.
- EN 112: задания, где исходная фраза уже была правильной, теперь прямо просят
  выбрать корректный mixed conditional. Исправлены обратная логика условий и
  результатов, неоднозначные варианты `If I were you`, ложное отнесение mixed
  conditional к third conditional и неполная форма `had booked`.
- EN 113: девять fill-blank заданий перестроены так, чтобы однословный ответ `to`
  действительно дополнял предложение. Во всех десяти вопросах блока о глаголах,
  допускающих обе формы, оставлен ровно один правильный вариант; задания на
  `like` и `prefer` теперь явно требуют сохранить to-infinitive.
- EN 114: восстановлено отрицание в `I didn’t mean to be rude`; исправлены
  неоднозначные задания на `remember`, `regret`, `try` и `suggest`, а также
  неверный перевод `meet the deadline`.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Signatures
изменённых тренировочных вопросов пересчитаны штатной функцией; новых дублей
signatures по всему EN training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят;
  обычный AJV до и после сохраняет exit 1 только из-за незагруженной meta-schema
  Draft 2020-12. Служебные `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 449 контрактов и решений, fingerprints, ответы,
  theory bindings, signatures, сохранность validation и равенство source → final →
  embedded: [check.log](check.log), exit 0.
- Все 25 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [grammar-review-tests.log](grammar-review-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 4155 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 24811/29095 вопросов основных банков, из них 224
`done` и 24587 `awaiting_verification`; 4284 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 113–116: 276 вопросов в 4 файлах; reports отсутствуют. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники EN 111–114

### EN 111

Итого: 107/107 (`fixed` 33, `ok` 74); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/111.en.grammar.conditionals_hypotheticals.wish_if_only_would/03-questions.json`
  Report: `docs/grammar-review/reports/39b7c13810b0c2ebd29ec730.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/111/en.01.b1_theory_wish_present.questions.json`
  Report: `docs/grammar-review/reports/aaded7e0e61780666cd94ea3.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/111/en.02.b2_theory_wish_past.questions.json`
  Report: `docs/grammar-review/reports/8f27dc83b45d54be0773675d.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/111/en.03.b3_theory_wish_would.questions.json`
  Report: `docs/grammar-review/reports/6d61b5d96f6fdfd8bc667683.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/111/en.04.b4_theory_would_rather.questions.json`
  Report: `docs/grammar-review/reports/c7b999c39065ec761260ccbf.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/111/en.05.b5_theory_its_time.questions.json`
  Report: `docs/grammar-review/reports/0d80ff121472b36255c902de.json`

### EN 112

Итого: 105/105 (`fixed` 31, `ok` 74); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/112.en.grammar.conditionals_hypotheticals.build_speak_advice_regrets/03-questions.json`
  Report: `docs/grammar-review/reports/12cc2292bf8daa0128ed3c74.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/112/en.01.b1_theory_advice.questions.json`
  Report: `docs/grammar-review/reports/9b5de3756746fb38bb81c68f.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/112/en.02.b2_theory_regrets.questions.json`
  Report: `docs/grammar-review/reports/5ab0946edcf1f1f1ebda4d7b.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/112/en.03.b3_theory_alternatives.questions.json`
  Report: `docs/grammar-review/reports/99ebaeb2b9307e6cf3169c7a.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/112/en.04.b4_theory_wish_if_only.questions.json`
  Report: `docs/grammar-review/reports/658782893783e51a8db2926e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/112/en.05.b5_theory_build_speak.questions.json`
  Report: `docs/grammar-review/reports/d558ac4978cf712e64155d5d.json`

### EN 113

Итого: 108/108 (`fixed` 25, `ok` 83); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/113.en.grammar.non_finite_forms_deep.gerund_vs_infinitive_basic/03-questions.json`
  Report: `docs/grammar-review/reports/af6a23bc68b0ed1ec2b5a272.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/113/en.01.b1_theory_overview.questions.json`
  Report: `docs/grammar-review/reports/0b4960efb46ab693168fa978.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/113/en.02.b2_theory_gerund.questions.json`
  Report: `docs/grammar-review/reports/9612db0ff7cf66ba868f33bc.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/113/en.03.b3_theory_infinitive.questions.json`
  Report: `docs/grammar-review/reports/da390ff9352d058fe530e788.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/113/en.04.b4_theory_both.questions.json`
  Report: `docs/grammar-review/reports/8767f85c79c7f3769f1cd100.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/113/en.05.b5_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/a8525c011c3c0493522b08ca.json`

### EN 114

Итого: 129/129 (`fixed` 9, `ok` 120); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/114.en.grammar.non_finite_forms_deep.verb_patterns_v_ing/03-questions.json`
  Report: `docs/grammar-review/reports/e08fa7544837d978c705860c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/114/en.01.b1_theory_overview.questions.json`
  Report: `docs/grammar-review/reports/5f9251f1d05550c5a9e1ac41.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/114/en.02.b2_theory_v_ing_only.questions.json`
  Report: `docs/grammar-review/reports/3c8353293d50f258d7b3b027.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/114/en.03.b3_theory_to_inf_only.questions.json`
  Report: `docs/grammar-review/reports/d598c8d8edccf41d63de75ff.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/114/en.04.b4_theory_memory_duty.questions.json`
  Report: `docs/grammar-review/reports/940ad330aa27751ea6248148.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/114/en.05.b5_theory_stop_try_mean.questions.json`
  Report: `docs/grammar-review/reports/a33946415224b96965aed6c1.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/114/en.06.b6_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/32a02c3f46d57f001bc41124.json`

## Следующая партия

- ES 113: 66 вопросов — pluscuamperfecto и действие до другого прошлого события.
- ES 114: 70 вопросов — futuro perfecto и condicional perfecto.
- ES 115: 70 вопросов — последовательность прошедших событий в пересказе.
- ES 116: 70 вопросов — управление временной линией и контраст прошедших времён.
- Итого: 276 вопросов в 4 файлах; reports отсутствуют.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
