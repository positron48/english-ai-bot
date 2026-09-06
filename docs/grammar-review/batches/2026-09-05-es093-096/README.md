# Редактура ES 093–096

2026-09-05 вычитано 282 вопроса основных банков в 4 исходных файлах.
Исправлено 49 вопросов, 233 оставлены без изменений. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| ES 093 | dar, mostrar, comprar и decir с объектными местоимениями | 70 | 10 | 60 |
| ES 094 | возвратные и взаимные действия | 68 | 11 | 57 |
| ES 095 | ir/irse, quedar/quedarse и смена значения | 72 | 12 | 60 |
| ES 096 | accidental `se` | 72 | 16 | 56 |
| Всего | | 282 | 49 | 233 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. Во всех четырёх главах до начала
партии уже были локальные изменения source/final и embedded-копий. Все 12 путей
включены в побитовый baseline и сохранены. Placement и отдельная очередь форм
глаголов исключены.

## Существенные исправления

- ES 093: нормативное дублирование косвенного дополнения в `Le doy el regalo a Ana`
  больше не объявляется ошибкой. Исправлены перспектива `te` в ответной реплике,
  примеры с `comprar` и различие конструкций с `a` и `para`, а также feedback к
  императивам и двойным местоимениям.
- ES 094: устранено сомнительное обобщение о взаимном прочтении
  `Los niños se lavan las manos`. Убраны альтернативно допустимые варианты с
  непереходным `vestir`, бенефактивным `lavarse` и `abrazarse a alguien`; задания
  на лицо, число и внешний объект получили однозначные контексты.
- ES 095: исправлен вопрос с инвертированной логикой «что не является новой задачей».
  Нормативное `Me voy al cine` больше не называется лишним `se`; уточнены границы
  `caer/caerse`, различие `me duermo` и `me acuesto`, а также допустимость `Vamos`
  вне явного требования использовать `irse`.
- ES 096: `Me cayó el vaso` больше не объявляется безусловно неграмматичным — явно
  отделено значение «стакан упал на меня» от `Se me cayó el vaso` «я случайно уронил».
  Для акцентного участника дано `A mí se me cayó`, уточнено согласование с предметом
  события и убран ложный критерий, будто любое `se` без `me/te/le` обязательно
  пассивное или безличное.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Новых точных
дублей вопросов в курсе не появилось. Теория не изменялась.

## Замечания к валидации и baseline

- Внутренний `validate-chapter.sh` проходит все четыре главы.
- Явный AJV Draft 2020 проходит ES 093. В ES 094–096 он до и после партии
  возвращает exit 1 из-за существующего несоответствия типа `matching` схеме.
- Обычный AJV сохраняет baseline exit 1 для всех четырёх глав из-за незагруженной
  meta-schema Draft 2020-12. Вывод каждого валидатора до и после совпадает дословно.
- Служебные `05-validation.json` восстановлены побайтно к входному состоянию.

## Проверки и интеграция

- Все четыре главы собраны и точечно синхронизированы с embedded-копиями.
- [check.py](check.py) проверил 282 контракта и решения, fingerprints, ответы,
  theory bindings, сохранность validation и равенство source → final → embedded:
  [check.log](check.log), exit 0.
- Четыре scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 3721 ранее изменённого стороннего
  файла: [preservation.log](preservation.log).

Общее редакторское покрытие: 21437/29095 вопросов основных банков, из них 224
`done` и 21213 `awaiting_verification`; 7658 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 095–098: 488 вопросов в 27 файлах. Reports отсутствуют, назначенные файлы чистые.
Полная цель обоих курсов остаётся активной.

## Закрытые исходники ES 093–096

- Source: `courses/spanish-grammar/chapters/093.es.grammar.pronouns_direct_indirect.build_speak_giving_showing_buying_telling/03-questions.json`
  Report: `docs/grammar-review/reports/111d2223787b3e0045a187b7.json`
  Редактура: 70/70 (`fixed` 10, `ok` 60); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/094.es.grammar.se_system_pronominal_verbs.reflexive_reciprocal_actions/03-questions.json`
  Report: `docs/grammar-review/reports/79cf2a54c17405642019fe5b.json`
  Редактура: 68/68 (`fixed` 11, `ok` 57); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/095.es.grammar.se_system_pronominal_verbs.pronominal_verbs_that_change_meaning_ir_irse_quedar/03-questions.json`
  Report: `docs/grammar-review/reports/30c1d9fd7f3d83859b73d9c9.json`
  Редактура: 72/72 (`fixed` 12, `ok` 60); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/096.es.grammar.se_system_pronominal_verbs.accidental_se_se_me_cayo_se_nos_olvido/03-questions.json`
  Report: `docs/grammar-review/reports/c2f37ea7b2d6c3739bb07d11.json`
  Редактура: 72/72 (`fixed` 16, `ok` 56); phase: `awaiting_verification`.

## Следующая партия

- EN 095: 129 вопросов в 7 файлах — Past Perfect и последовательность прошлых событий.
- EN 096: 131 вопрос в 7 файлах — Future Perfect и Future Continuous.
- EN 097: 121 вопрос в 7 файлах — связная биография и жизненный опыт.
- EN 098: 107 вопросов в 6 файлах — основы пассива `be + V3`.
- Итого: 488 вопросов в 27 файлах; reports отсутствуют.
- Все 62 назначенных source/final/validation и embedded-пути чистые.
- Baseline HEAD EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
