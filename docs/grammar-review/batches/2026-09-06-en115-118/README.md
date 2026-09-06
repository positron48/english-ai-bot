# Редактура EN 115–118

2026-09-06 вычитано 422 вопроса в 25 исходных файлах: 4 основных банка и
21 тренировочный блок. Исправлено 86 вопросов, 336 оставлены без изменений.
Все отчёты имеют `awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| EN 115 | perfect infinitive и perfect gerund | 6 | 106 | 24 | 82 |
| EN 116 | participle clauses и `having + V3` | 7 | 117 | 11 | 106 |
| EN 117 | dangling participles | 6 | 93 | 37 | 56 |
| EN 118 | итоговая письменная практика | 6 | 106 | 14 | 92 |
| Всего | | 25 | 422 | 86 | 336 |

Редактор: `Codex editor 2026-09-06`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. До начала партии все 58 назначенных
source/final/validation и embedded-путей были чистыми. Placement и отдельная
очередь форм глаголов ES исключены.

## Существенные исправления

- EN 115: исправлена форма после модального `must` (`have left` без `to`), убраны
  конкурирующие правильные ответы после `regret` и `admit`, а задания на perfect
  gerund теперь прямо отличают целевую форму от допустимого простого герундия.
  Восстановлены полные `to have + V3` в тренировочных вариантах и исправлена
  обратная временная логика с `before`.
- EN 116: исправлены неполные fill-blank конструкции `having lost/edited`, связь
  субъектов при преобразовании придаточного, неоднозначные ключи для активных и
  пассивных причастных оборотов и задания, где несколько вариантов содержали
  dangling participle.
- EN 117: исправлен категоричный и неверный тезис «V3 всегда пассивен». В серии
  заданий на обнаружение и исправление dangling participle устранены множественные
  правильные ответы, ложные исходные предпосылки и неверные ключи; устойчивые
  вводные маркеры теперь сформулированы отдельно от обычных причастных оборотов.
- EN 118: устранён dangling participle в примере с анализом feedback, уточнены
  задания на пассивные V3-обороты и пунктуацию, а искусственный запрет на несколько
  неличных форм заменён проверкой реальной перегруженности предложения.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Signatures
изменённых тренировочных вопросов пересчитаны штатной функцией; новых дублей
signatures по всему EN training pack нет.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator и явный AJV Draft 2020 проходят;
  обычный AJV до и после сохраняет exit 1 только из-за незагруженной meta-schema
  Draft 2020-12. Служебные `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 422 контракта и решения, fingerprints, ответы,
  theory bindings, signatures, сохранность validation и равенство source → final →
  embedded: [check.log](check.log), exit 0.
- Все 25 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [grammar-review-tests.log](grammar-review-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 4257 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 25509/29095 вопросов основных банков, из них 224
`done` и 25285 `awaiting_verification`; 3586 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 117–120: 280 вопросов в 4 файлах; reports отсутствуют. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники EN 115–118

### EN 115

Итого: 106/106 (`fixed` 24, `ok` 82); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/115.en.grammar.non_finite_forms_deep.perfect_infinitive_gerund_to/03-questions.json`
  Report: `docs/grammar-review/reports/91a9a1a56e65c2f97cbecab3.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/115/en.01.b1_theory_overview.questions.json`
  Report: `docs/grammar-review/reports/f8bd06d3f6a9930f3e69d320.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/115/en.02.b2_theory_perfect_infinitive.questions.json`
  Report: `docs/grammar-review/reports/1a454600c5ebda2192bf9e1c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/115/en.03.b3_theory_perfect_gerund.questions.json`
  Report: `docs/grammar-review/reports/417f32c664ef4fb12365ff5c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/115/en.04.b4_theory_comparison.questions.json`
  Report: `docs/grammar-review/reports/d2aac0c88f57960b9b4ca0b9.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/115/en.05.b5_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/d45b889a53ba34471f653e9e.json`

### EN 116

Итого: 117/117 (`fixed` 11, `ok` 106); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/116.en.grammar.non_finite_forms_deep.participle_clauses_having_finished/03-questions.json`
  Report: `docs/grammar-review/reports/cfe2a9f41dbad8df2e9c1a54.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/116/en.01.b1_theory_overview.questions.json`
  Report: `docs/grammar-review/reports/9497fb8a0fc792103a1428a5.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/116/en.02.b2_theory_present_participle.questions.json`
  Report: `docs/grammar-review/reports/9ce909a907cd2706f8046840.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/116/en.03.b3_theory_past_participle.questions.json`
  Report: `docs/grammar-review/reports/ca683026810a84d209fc1ca4.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/116/en.04.b4_theory_perfect_participle.questions.json`
  Report: `docs/grammar-review/reports/d062fa8de7e8a54f3470801c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/116/en.05.b5_theory_style_markers.questions.json`
  Report: `docs/grammar-review/reports/6f8669121650c95af8d143cc.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/116/en.06.b6_theory_common_errors.questions.json`
  Report: `docs/grammar-review/reports/32886b02d02fd6851e28d0ff.json`

### EN 117

Итого: 93/93 (`fixed` 37, `ok` 56); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/117.en.grammar.non_finite_forms_deep.dangling_participles_common_mistake/03-questions.json`
  Report: `docs/grammar-review/reports/e185e81035addf83a4729358.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/117/en.01.b1_theory_definition.questions.json`
  Report: `docs/grammar-review/reports/7a9ec05f16506ae770e97b2c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/117/en.02.b2_theory_detection.questions.json`
  Report: `docs/grammar-review/reports/eff59060e3e1f2d91bfe2ca3.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/117/en.03.b3_theory_fixes.questions.json`
  Report: `docs/grammar-review/reports/784ae4199667157b306ed952.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/117/en.04.b4_theory_safe_phrases.questions.json`
  Report: `docs/grammar-review/reports/ecfeca60e2f0c3953ab7ad20.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/117/en.05.b5_theory_common_traps.questions.json`
  Report: `docs/grammar-review/reports/5956c904cb0ba4ed4a061515.json`

### EN 118

Итого: 106/106 (`fixed` 14, `ok` 92); phase `awaiting_verification`.

- `chapter` Source: `courses/english-grammar/chapters/118.en.grammar.non_finite_forms_deep.build_write_a_short/03-questions.json`
  Report: `docs/grammar-review/reports/e59f692ef200171eda8a92b4.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/118/en.01.b1_theory_task_overview.questions.json`
  Report: `docs/grammar-review/reports/d209959a4b17a19e61d96a05.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/118/en.02.b2_theory_construction_menu.questions.json`
  Report: `docs/grammar-review/reports/b13c5685014f11a2d20f006e.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/118/en.03.b3_theory_structure.questions.json`
  Report: `docs/grammar-review/reports/ca6d6fcb18c684f72fe89003.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/118/en.04.b4_theory_accuracy.questions.json`
  Report: `docs/grammar-review/reports/bfb47dfd8dafb1f6924b564c.json`
- `training` Source: `courses/english-grammar/training_pack/chapters/118/en.05.b5_theory_checklist.questions.json`
  Report: `docs/grammar-review/reports/020bc3eba6686af09f62183b.json`

## Следующая партия

- ES 117: 70 вопросов — биография, flashback и причинно-следственные цепочки.
- ES 118: 70 вопросов — образование presente de subjuntivo.
- ES 119: 70 вопросов — желание и влияние: `quiero que`, `necesito que`.
- ES 120: 70 вопросов — эмоции и оценка: `me alegra que`, `es bueno que`.
- Итого: 280 вопросов в 4 файлах; reports отсутствуют.
- Baseline HEAD ES: `501044644954d5aaa474f7d1eaee99518a4ebd9d`.
