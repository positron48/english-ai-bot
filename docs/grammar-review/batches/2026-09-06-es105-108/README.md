# Редактура ES 105–108

2026-09-06 вычитано 270 вопросов в 4 основных банках. Исправлено 52 вопроса,
218 оставлены без изменений. Все отчёты имеют `awaiting_verification`;
независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| ES 105 | косвенные вопросы | 72 | 13 | 59 |
| ES 106 | связный текст 120–180 слов | 66 | 8 | 58 |
| ES 107 | инфинитив после предлогов | 66 | 5 | 61 |
| ES 108 | gerundio | 66 | 26 | 40 |
| Всего | | 270 | 52 | 218 |

Редактор: `Codex editor 2026-09-06`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. Все 16 назначенных
source/final/validation и embedded-путей до начала партии были чистыми. Placement и
отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- ES 105: прямой вопрос отделён от внешнего вопроса со встроенной частью. Убрано
  чрезмерное правило про «утвердительный порядок» и «инверсию»; пунктуация и порядок
  слов теперь объясняются через синтаксическую роль. Конструкция *qué es lo que*
  корректно описана как грамматичная усилительная модель. Исправлены примеры с
  *preguntar*, согласование времени и несколько неточных переводов.
- ES 106: исправлено *en mi opinión*, а ненадёжное различение *porque / por qué* по
  произношению заменено смысловым и орфографическим правилом. Устранены два правильных
  ответа в задании с *cuyo*. Правило о запятой теперь зависит от синтаксиса и смысла,
  а не от длины пояснения; улучшены уступительный пример и заключение.
- ES 107: исправлены неестественное *ayudar contigo* и оборот *por + infinitivo* с
  неверно приписанным субъектом. Примеры теперь явно различают общий субъект,
  собственный субъект инфинитива и придаточное с *para que / antes de que*.
- ES 108: у 24 MCQ после перестановки вариантов ключи указывали на неправильные ID;
  все ключи восстановлены по фактическим вариантам. Исправлены объяснение формы
  *durmiendo*, неполный вопрос с *estar + gerundio*, термин для составного герундия и
  пропущенная запятая в reorder-ответе.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. В
[check.py](check.py) добавлена отдельная регрессионная проверка восстановленных ключей
ES 108.

## Проверки и интеграция

- Все четыре главы собраны; source, final и embedded-копии совпадают. Внутренний
  validator проходит. Явный AJV Draft 2020 сохраняет исходный результат: ES 107–108
  проходят, а ES 105–106 имеют прежние ошибки в не относящихся к вопросам
  `concept_refs`/theory tags. Обычный AJV до и после возвращает exit 1 из-за
  незагруженной meta-schema Draft 2020-12. Служебные `05-validation.json`
  восстановлены побайтно.
- [check.py](check.py) проверил 270 контрактов и решений, fingerprints, ответы,
  theory bindings, сохранность validation и равенство source → final → embedded:
  [check.log](check.log), exit 0.
- Все 4 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 4042 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 23671/29095 вопросов основных банков, из них 224
`done` и 23447 `awaiting_verification`; 5424 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 107–110: 427 вопросов в 24 файлах. Reports отсутствуют; целевые source/final и
embedded-копии согласованы. Полная цель обоих курсов остаётся активной.

## Закрытые исходники ES 105–108

### ES 105

Итого: 72/72 (`fixed` 13, `ok` 59); phase `awaiting_verification`.

- Source: `courses/spanish-grammar/chapters/105.es.grammar.complex_sentences_connecting.indirect_questions_no_se_que_me_puedes_decir/03-questions.json`
  Report: `docs/grammar-review/reports/726fc8630ab5862fc217aa0e.json`

### ES 106

Итого: 66/66 (`fixed` 8, `ok` 58); phase `awaiting_verification`.

- Source: `courses/spanish-grammar/chapters/106.es.grammar.complex_sentences_connecting.build_write_120_180_word_text_connectors_relative/03-questions.json`
  Report: `docs/grammar-review/reports/bab6a21da647cc92de3ae9d6.json`

### ES 107

Итого: 66/66 (`fixed` 5, `ok` 61); phase `awaiting_verification`.

- Source: `courses/spanish-grammar/chapters/107.es.grammar.non_finite_forms_periphrases.infinitive_after_prepositions_conjunctions/03-questions.json`
  Report: `docs/grammar-review/reports/ae865d17e8f060cc71a6f1d5.json`

### ES 108

Итого: 66/66 (`fixed` 26, `ok` 40); phase `awaiting_verification`.

- Source: `courses/spanish-grammar/chapters/108.es.grammar.non_finite_forms_periphrases.gerundio_ongoing_action_manner_do/03-questions.json`
  Report: `docs/grammar-review/reports/5ed95f6f64549c21d3842d90.json`

## Следующая партия

- EN 107: 106 вопросов в 6 файлах — second conditional.
- EN 108: 108 вопросов в 6 файлах — third conditional.
- EN 109: 107 вопросов в 6 файлах — mixed conditionals.
- EN 110: 106 вопросов в 6 файлах — `unless`, `provided`, `as long as`.
- Итого: 427 вопросов в 24 файлах; reports отсутствуют.
- Source/final и embedded-копии назначенных файлов согласованы.
- Baseline HEAD EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
