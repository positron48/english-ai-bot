# Редактура ES 121–124

2026-09-06 вычитано 280 вопросов в 4 основных банках. Исправлено 44 вопроса,
236 оставлены без изменений. Все отчёты имеют `awaiting_verification`;
независимая проверка остаётся `pending`.

| Глава | Тема | Файлов | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: | ---: |
| ES 121 | сомнение, отрицание, возможность и уверенность | 1 | 70 | 5 | 65 |
| ES 122 | безличные оценки и триггеры уверенности | 1 | 70 | 19 | 51 |
| ES 123 | отрицательные и формальные команды | 1 | 70 | 10 | 60 |
| ES 124 | выбор indicativo/subjuntivo в noun clauses | 1 | 70 | 10 | 60 |
| Всего | | 4 | 280 | 44 | 236 |

Редактор: `Codex editor 2026-09-06`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. Placement и отдельная очередь форм
глаголов ES исключены.

До начала партии ES 123/q69 уже был синхронно изменён в source, final и embedded:
абстрактный чеклист заменён конкретным заданием *No hables*. Эта корректная правка
принята как baseline, не перезаписывалась и не включена в 44 исправления партии.
Остальные назначенные пути были чистыми.

## Существенные исправления

- ES 121: будущая уверенность приведена к *Creo que vendrá / Estoy seguro de que
  vendrá*; отрицание сообщения отделено от отрицания факта; правило для *quizás*
  больше не запрещает indicativo при более высокой уверенности.
- ES 122: восстановлены испанские диакритические знаки; переписан неудачный вариант
  о формальной уверенности; у задания с безличным *haber* устранены два неверных
  ответа и ошибочная маркировка времени. Уточнено, что безличный *haber* не
  согласуется с *errores*: *haya/hubiera*, не *hayan*.
- ES 123: разведены морфологическая форма presente de subjuntivo и функция
  imperativo; уточнены команды *usted/ustedes*, порядок местоимений и значение
  самостоятельного *Que hable…*. Исправлен пропущенный акцент в *habláis*.
- ES 124: устранён дословно повторявшийся distractor; снята неоднозначность задания
  *Me alegra que… / Me alegra verte*; уточнены функции существительных придаточных,
  безличного инфинитива и ограничения правила об отрицательной рамке.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены.

## Проверки и интеграция

- Все четыре главы собраны. Внутренний validator проходит. Явный AJV Draft 2020
  проходит для ES 121, 122 и 124; у ES 123 побайтно сохранён baseline schema-сбой:
  шесть тегов теоретических примеров нарушают допустимый шаблон тега. Обычный AJV
  до и после сохраняет exit 1 из-за незагруженной meta-schema Draft 2020-12.
  Служебные `05-validation.json` восстановлены побайтно.
- [check.py](check.py) проверил 280 контрактов и решений, fingerprints, ответы,
  theory bindings, отсутствие новых дубликатов, сохранность validation и равенство
  source → final → embedded: [check.log](check.log), exit 0.
- Все 4 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [grammar-review-tests.log](grammar-review-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- В первоначальном setup путь `spanish-grammar` был ошибочно опрошен дважды, один
  раз с префиксом `english-grammar`. До исправления выполнялись только записи в
  разрешённые пути ES 121–124, их reports и каталог партии. Снимок сторонних путей
  был переснят с правильными репозиториями; причина и способ восстановления записаны
  в [preservation-before.json](preservation-before.json). Итоговая проверка
  подтверждает сохранность 4434 сторонних файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 26483/29095 вопросов основных банков, из них 224
`done` и 26259 `awaiting_verification`; 2612 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 123–126: 433 вопроса в 24 файлах; reports отсутствуют. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники ES 121–124

### ES 121

Итого: 70/70 (`fixed` 5, `ok` 65); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/121.es.grammar.subjunctive_noun_clauses.doubt_denial_possibility_certainty/03-questions.json`
  Report: `docs/grammar-review/reports/6df4954e1ec0f573bb21f216.json`

### ES 122

Итого: 70/70 (`fixed` 19, `ok` 51); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/122.es.grammar.subjunctive_noun_clauses.impersonal_expressions_triggers_certainty/03-questions.json`
  Report: `docs/grammar-review/reports/7b1fd8238063b4f9567ef526.json`

### ES 123

Итого: 70/70 (`fixed` 10, `ok` 60); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/123.es.grammar.subjunctive_noun_clauses.negative_commands_polite_commands_subjunctive/03-questions.json`
  Report: `docs/grammar-review/reports/9efc7506257fa3ff409e1520.json`

### ES 124

Итого: 70/70 (`fixed` 10, `ok` 60); phase `awaiting_verification`.

- `chapter` Source: `courses/spanish-grammar/chapters/124.es.grammar.subjunctive_noun_clauses.indicative_subjunctive_decision_rules_noun_clauses/03-questions.json`
  Report: `docs/grammar-review/reports/c9ddfba1efeba66beac7690b.json`

## Следующая партия

- EN 123: 109 вопросов в 6 файлах — связный пересказ разговора, новости и инструкций.
- EN 124: 106 вопросов в 6 файлах — do-emphasis.
- EN 125: 109 вопросов в 6 файлах — инверсия после negative/restrictive expressions.
- EN 126: 109 вопросов в 6 файлах — it-cleft и wh-cleft.
- Итого: 433 вопроса в 24 файлах; reports отсутствуют.
- Baseline HEAD EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
