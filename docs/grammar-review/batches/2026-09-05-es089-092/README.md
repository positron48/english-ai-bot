# Редактура ES 089–092

2026-09-05 вычитано 288 вопросов основных банков в 4 исходных файлах.
Исправлен 71 вопрос, 217 оставлены без изменений. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| ES 089 | косвенные объектные местоимения `le/les` | 72 | 11 | 61 |
| ES 090 | позиция объектных местоимений | 72 | 36 | 36 |
| ES 091 | двойные местоимения `me lo`, `se lo`, `te la` | 72 | 5 | 67 |
| ES 092 | базовая модель `le/lo/la` и leísmo | 72 | 19 | 53 |
| Всего | | 288 | 71 | 217 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. В ES 090 и ES 092 уже были
локальные изменения source/final и embedded-копий; они включены в побитовый baseline
и сохранены. Placement и отдельная очередь форм глаголов исключены.

## Существенные исправления

- ES 089: примеры с `le` приведены к естественной речи. Уточнено нормативное
  дублирование получателя в `Le doy el libro a Ana`, исправлен пример с ошибочным
  `Le llamo`, а личный leísmo отделён от базовой модели прямого дополнения курса.
- ES 090: в 30 заданиях входящий порядок вариантов не соответствовал неизменяемому
  `correct_answer`: правильные тексты и feedback перенесены к существующим ID ответов.
  Исправлены объяснения `el regalo → lo`, ударений в `leyéndolo` и `haciéndolo`,
  отрицательных команд и позиции двух местоимений.
- ES 091: уточнены образование `explícamelo` и `dámelo`, описание неизменяемого `me`
  перед `lo`, русский перевод `No me lo digas` и правило порядка `me lo`.
- ES 092: базовая модель курса сохранена, но она больше не объявляет нормативно
  допустимый личный leísmo безусловной ошибкой. Устранены ненормативные удвоения
  `Lo veo a Pedro`, исправлена роль дополнения в спорном примере заменой на
  однозначное `Le escribo un mensaje`, уточнены границы leísmo для лица мужского рода.

IDs, порядок, типы, difficulty, theory bindings, choice IDs и IDs правильных
ответов сохранены. Для проверки формулировок о leísmo использовано официальное
определение RAE: https://www.rae.es/diccionario-estudiante/le%C3%ADsmo.

## Замечания к валидации и baseline

- Внутренний `validate-chapter.sh` проходит все четыре главы.
- Явный AJV Draft 2020 проходит ES 090. В ES 089, 091 и 092 он до и после партии
  возвращает exit 1 из-за существующего несоответствия типа `matching` схеме.
- Обычный AJV сохраняет baseline exit 1 для всех четырёх глав из-за незагруженной
  meta-schema Draft 2020-12. Вывод каждого валидатора до и после совпадает дословно.
- Служебные `05-validation.json` восстановлены побайтно к текущему входному состоянию.
  Теория не изменялась.

## Проверки и интеграция

- Все четыре главы собраны и точечно синхронизированы с embedded-копиями.
- [check.py](check.py) проверил 288 контрактов и решений, fingerprints, ответы,
  theory bindings, сохранность validation и равенство source → final → embedded:
  [check.log](check.log), exit 0.
- Четыре scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 3632 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 20717/29095 вопросов основных банков, из них 224
`done` и 20493 `awaiting_verification`; 8378 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 091–094: 438 вопросов в 25 файлах. Reports отсутствуют, назначенные файлы чистые.
Полная цель обоих курсов остаётся активной.

## Закрытые исходники ES 089–092

- Source: `courses/spanish-grammar/chapters/089.es.grammar.pronouns_direct_indirect.indirect_object_pronouns_le_les_recipient_meaning/03-questions.json`
  Report: `docs/grammar-review/reports/67d4bac7d7c71326b5bc0967.json`
  Редактура: 72/72 (`fixed` 11, `ok` 61); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/090.es.grammar.pronouns_direct_indirect.placement_rules_before_verb_attached_after/03-questions.json`
  Report: `docs/grammar-review/reports/1342c650bda10640988f89ed.json`
  Редактура: 72/72 (`fixed` 36, `ok` 36); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/091.es.grammar.pronouns_direct_indirect.double_pronouns_me_lo_se_lo_te_la/03-questions.json`
  Report: `docs/grammar-review/reports/eda91a5ea3aa08a452f6192f.json`
  Редактура: 72/72 (`fixed` 5, `ok` 67); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/092.es.grammar.pronouns_direct_indirect.le_lo_la_baseline_norm_learners_should_treat/03-questions.json`
  Report: `docs/grammar-review/reports/861ecea081a563a4b4598036.json`
  Редактура: 72/72 (`fixed` 19, `ok` 53); phase: `awaiting_verification`.

## Следующая партия

- EN 091: 109 вопросов в 6 файлах — Present Perfect как результат сейчас.
- EN 092: 125 вопросов в 7 файлах — жизненный опыт, `ever/never`, `been/gone`.
- EN 093: 97 вопросов в 6 файлах — длительность с `for/since/how long`.
- EN 094: 107 вопросов в 6 файлах — Present Perfect и Past Simple.
- Итого: 438 вопросов в 25 файлах; reports отсутствуют.
- Все назначенные source/final/validation и embedded-файлы чистые.
- Baseline HEAD EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
