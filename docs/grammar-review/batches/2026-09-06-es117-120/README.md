# Редактура ES 117–120

2026-09-06 вычитано 280 вопросов в 4 основных банках. Исправлено 25 вопросов,
255 оставлены без изменений. Все отчёты имеют `awaiting_verification`;
независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| ES 117 | биография, flashback и причинно-следственные цепочки | 1 | 70 | 7 | 63 |
| ES 118 | образование presente de subjuntivo | 1 | 70 | 5 | 65 |
| ES 119 | желание и влияние | 1 | 70 | 6 | 64 |
| ES 120 | эмоции и оценка | 1 | 70 | 7 | 63 |
| Всего | | 4 | 280 | 25 | 255 |

Редактор: `Codex editor 2026-09-06`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. До начала партии все 16 назначенных
source/final/validation и embedded-путей были чистыми. Placement и отдельная
очередь форм глаголов ES исключены.

## Существенные исправления

- ES 117: уточнены условия входа во flashback и требования к anterioridad;
  устранён второй допустимый ответ с *había querido*. Исправлена невозможная
  хронология, где *años antes* относилось ко времени до рождения героини, и
  согласованы рамки presente/perfecto и прошлой цепочки.
- ES 118: устранены два правильных ответа в заданиях на основу subjuntivo и
  нерегулярные yo-формы. Исправлено образование *salg-*, отдельное задание на
  *dar → dé*, неверное утверждение о car/gar/zar и описание форм *sea/sé*.
- ES 119: исправлены значение безличного *hace falta que*, ударение в *prohíben*,
  объяснение субъектов в *espero verte / espero que vengas* и неоднозначное
  задание, где несколько вариантов не передавали исходную фразу.
- ES 120: исправлены опечатка, неверные варианты ответа для формы *fumes*,
  двойная ошибка *te me parece*, описание субъекта инфинитива и различие между
  общей оценкой, необходимостью и придаточным с отдельным субъектом.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator проходит. Явный AJV Draft 2020
  проходит для ES 117, 119 и 120; у ES 118 побайтно сохранён baseline schema-сбой
  из-за пробела в одном теге теоретического примера. Обычный AJV до и после
  сохраняет exit 1 из-за незагруженной meta-schema Draft 2020-12. Служебные
  `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 280 контрактов и решений, fingerprints, ответы,
  theory bindings, сохранность validation и равенство source → final → embedded:
  [check.log](check.log), exit 0.
- Все 4 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [grammar-review-tests.log](grammar-review-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 4339 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 25789/29095 вопросов основных банков, из них 224
`done` и 25565 `awaiting_verification`; 3306 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 119–122: 414 вопросов в 24 файлах; reports отсутствуют. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники ES 117–120

### ES 117

Итого: 70/70 (`fixed` 7, `ok` 63); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/117.es.grammar.compound_tenses_narration.build_write_biography_flashback_cause_effect_chains/03-questions.json`
  Report: `docs/grammar-review/reports/ec8c1b0aefc886a5c853942e.json`

### ES 118

Итого: 70/70 (`fixed` 5, `ok` 65); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/118.es.grammar.subjunctive_noun_clauses.form_present_subjunctive/03-questions.json`
  Report: `docs/grammar-review/reports/19488c58ad3d39a369c4c444.json`

### ES 119

Итого: 70/70 (`fixed` 6, `ok` 64); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/119.es.grammar.subjunctive_noun_clauses.desire_influence_quiero_que_necesito_que/03-questions.json`
  Report: `docs/grammar-review/reports/d920a7534b68a45c0a3db377.json`

### ES 120

Итого: 70/70 (`fixed` 7, `ok` 63); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/120.es.grammar.subjunctive_noun_clauses.emotion_evaluation_me_alegra_que_es_bueno_que/03-questions.json`
  Report: `docs/grammar-review/reports/05c4619bd2fd77567c8134e1.json`

## Следующая партия

- EN 119: 107 вопросов в 6 файлах — reported statements, `say` и `tell`.
- EN 120: 97 вопросов в 6 файлах — reported questions и порядок слов.
- EN 121: 107 вопросов в 6 файлах — reported commands и requests.
- EN 122: 103 вопроса в 6 файлах — reporting verbs `advise`, `suggest`, `insist`.
- Итого: 414 вопросов в 24 файлах; reports отсутствуют.
- Baseline HEAD EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
