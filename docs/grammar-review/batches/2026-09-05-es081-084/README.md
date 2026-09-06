# Редактура ES 081–084

2026-09-05 вычитано 288 вопросов основных банков в 4 исходных файлах.
Исправлено 23 вопроса, 265 оставлены без изменений. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| ES 081 | персональная `a` перед прямым дополнением | 72 | 6 | 66 |
| ES 082 | `a`, `de`, `en`, `con`, `sin`, `sobre` | 72 | 5 | 67 |
| ES 083 | центральное различие `por` и `para` | 72 | 5 | 67 |
| ES 084 | время и движение: `desde`, `hasta`, `hacia`, `por`, `para` | 72 | 7 | 65 |
| Всего | | 288 | 23 | 265 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. Все назначенные исходники до
начала партии были чистыми, reports отсутствовали. Placement и отдельная очередь
форм глаголов исключены.

## Существенные исправления

- ES 081: неоднозначное `Busco un fontanero` заменено контекстом с конкретным
  специалистом, где персональная `a` обязательна. Уточнены объяснения прямого
  дополнения, убраны смешанные русско-испанские варианты и добавлены полноценные
  примеры для персональной, направительной и косвенной `a`.
- ES 082: задания с `en` и `sobre` теперь прямо различают общее место и поверхность.
  В упражнении на тему разговора явно задан `sobre`, поскольку `hablar de` тоже
  грамматично. Убран второй правильный вариант с `en el suelo`.
- ES 083: уточнены значения `por ti`, естественная формулировка `para uso personal`
  и пример цены с `por tres euros`. Убран допустимый вариант `interesarse en`;
  ошибочный ключ true/false заменён однозначно ложным утверждением.
- ES 084: исправлены смешанные языки в вариантах, неестественные предложения и
  объяснение формы `nos quedamos`. Контексты теперь различают направление `hacia`,
  точку назначения `a/al` и маршрут `por`.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены.

## Замечания к теории и baseline

- ES 081 излагает персональную `a` в учебном упрощении. Для животных, групп людей
  и некоторых названий профессий употребление зависит от определённости и степени
  персонализации; исправленные вопросы получили достаточный контекст.
- В ES 082 `en` и `sobre` частично пересекаются в значении положения на поверхности,
  поэтому задания теперь явно называют требуемый оттенок.
- В ES 083 `por` с продолжительностью допустимо в некоторых вариантах испанского;
  формулировки не объявляют его универсально неверным.
- `validate-chapter.sh` для ES 082 до и после партии выдаёт одно и то же предупреждение
  о baseline-балансе true/false: 10 `true` и 2 `false`. Оно не вызвано редактурой.
- Обычный AJV без `--spec=draft2020` не загружает meta-schema Draft 2020-12 и штатно
  возвращает exit 1. Явный Draft 2020 проходит; результаты до и после совпадают.

Теория не входит в текущий scope и не изменялась.

## Проверки и интеграция

- Все четыре главы собраны. `validate-chapter.sh` и явный AJV Draft 2020 принимают
  их с тем же результатом, что baseline; служебные validation-файлы восстановлены.
- [check.py](check.py) проверил 288 контрактов и решений, fingerprints, ответы,
  theory bindings и равенство source → final → embedded: [check.log](check.log), exit 0.
- Все четыре final-главы точечно синхронизированы с embedded. Неопросные поля final
  сохранены, кроме штатного `updated_at` сборки.
- Четыре scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Три проверки
  `git diff --check` прошли: [diff-check.log](diff-check.log).
- Все 13 тестов механизма учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Контрольный снимок подтвердил сохранность всех 3439 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 19244/29095 вопросов основных банков, из них 224
`done` и 19020 `awaiting_verification`; 9851 ещё не прошёл первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 083–086: 447 вопросов в 25 файлах. Все назначаемые исходники чистые, reports
отсутствуют. Полная цель обоих курсов остаётся активной.

## Закрытые исходники ES 081–084

- Source: `courses/spanish-grammar/chapters/081.es.grammar.prepositions_verb_patterns.personal_object_marking/03-questions.json`
  Report: `docs/grammar-review/reports/6e74ca0dc0097f9c20eeb702.json`
  Редактура: 72/72 (`fixed` 6, `ok` 66); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/082.es.grammar.prepositions_verb_patterns.core_prepositions_de_en_con_sin_sobre/03-questions.json`
  Report: `docs/grammar-review/reports/f90bb4bc4b23f8354f0df854.json`
  Редактура: 72/72 (`fixed` 5, `ok` 67); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/083.es.grammar.prepositions_verb_patterns.por_para_central_contrast/03-questions.json`
  Report: `docs/grammar-review/reports/012c93a3eb6891ce2f60dc9d.json`
  Редактура: 72/72 (`fixed` 5, `ok` 67); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/084.es.grammar.prepositions_verb_patterns.prepositions_time_movement_desde_hasta_hacia_por_para/03-questions.json`
  Report: `docs/grammar-review/reports/a47abf0ea5c0aef229ce1002.json`
  Редактура: 72/72 (`fixed` 7, `ok` 65); phase: `awaiting_verification`.

## Следующая партия

- EN 083: 128 вопросов в 7 файлах — `because`, `so`, `although`, `but`, `while`, `whereas`.
- EN 084: 105 вопросов в 6 файлах — придаточные времени.
- EN 085: 103 вопроса в 6 файлах — noun clauses.
- EN 086: 111 вопросов в 6 файлах — косвенные вопросы.
- Итого: 447 вопросов в 25 файлах; назначаемые исходники чистые, reports отсутствуют.
- Baseline EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
