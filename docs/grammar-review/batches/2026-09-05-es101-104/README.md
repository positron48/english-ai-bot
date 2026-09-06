# Редактура ES 101–104

С 2026-09-05 по 2026-09-06 вычитано 288 вопросов в четырёх основных банках.
Исправлено 74 вопроса, 214 оставлены без изменений. Все четыре отчёта имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| ES 101 | `porque`, `como`, `por eso`, `así que` | 72 | 15 | 57 |
| ES 102 | `pero`, `sino`, `aunque`, `sin embargo` | 72 | 22 | 50 |
| ES 103 | `cuando`, `antes de`, `después de`, `mientras` | 72 | 18 | 54 |
| ES 104 | `que`, `quien`, `donde`, `cuyo` | 72 | 19 | 53 |
| Всего | | 288 | 74 | 214 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. Все 16 назначенных
source/final/validation и embedded-путей до начала партии были чистыми. Placement и
отдельная очередь форм глаголов ES исключены.

## Существенные исправления

- ES 101: устранён ложный запрет на начальное `porque`; разведены союз `así que` и
  самостоятельный коннектор `por eso` с учётом пунктуации; исправлены допустимые
  вопросы, которые раньше назывались ошибочными. В q56 ключ изменён с перестройки
  предложения на минимальную пунктуационную правку.
- ES 102: восстановлено соответствие prompt и вариантов в вопросе о `sino que`;
  исправлены переводы `martes/miércoles`, дождя/снега и условия `si no estudias`;
  синтаксически разведены `aunque` и `sin embargo`; устранены дубли полных ответов.
- ES 103: `cuando` больше не объявляется неправильным для привычных ситуаций, а
  `mientras` выбирается при явном акценте на длительной одновременности. Уточнены
  `mientras que`, вопросительное `cuándo` и порядок действий с `antes/después de`.
- ES 104: допустимое вынесение объекта `La casa la compramos…` больше не выдаётся
  за ошибку относительного придаточного. Исправлены реальные дубли объектов после
  `que`, употребление `quien` после предлога, личное `a`, косвенное `le` и
  согласование `cuyo`.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Изменён один
ошибочный ключ — ES 101 q56. Полные правильные ответы внутри глав уникальны.

## Проверки и интеграция

- Все четыре главы собраны; source, final и embedded-копии совпадают.
- Внутренний validator проходит для всех глав. Явный AJV Draft 2020 проходит для
  ES 102–104. У ES 101 сохранены те же шесть baseline-ошибок формата тегов в theory;
  вопросная банка их не меняет. Обычный AJV до и после сохраняет exit 1 из-за
  незагруженной meta-schema Draft 2020-12. Служебные `05-validation.json`
  восстановлены побайтно.
- [check.py](check.py) проверил 288 контрактов и решений, fingerprints, ответы,
  theory bindings, сохранность validation и равенство source → final → embedded:
  [check.log](check.log), exit 0.
- Все четыре scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Все 13 тестов механизма
  учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Три проверки `git diff --check` прошли: [diff-check.log](diff-check.log).
- Контрольный снимок подтвердил сохранность всех 3944 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие основных банков: 22970/29095 вопросов, из них 224
`done` и 22746 `awaiting_verification`; 6125 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 103–106: 431 вопрос в 24 исходных файлах. Reports отсутствуют; все 56 назначенных
source/final/validation и embedded-путей чистые и согласованы. Полная цель обоих
курсов остаётся активной.

## Закрытые исходники ES 101–104

### ES 101

Итого: 72/72 (`fixed` 15, `ok` 57); phase `awaiting_verification`.

- Source: `courses/spanish-grammar/chapters/101.es.grammar.complex_sentences_connecting.porque_como_por_eso_asi_que_cause_result/03-questions.json`
  Report: `docs/grammar-review/reports/98ac09ca8493eebf737d0943.json`

### ES 102

Итого: 72/72 (`fixed` 22, `ok` 50); phase `awaiting_verification`.

- Source: `courses/spanish-grammar/chapters/102.es.grammar.complex_sentences_connecting.pero_sino_aunque_sin_embargo_contrast/03-questions.json`
  Report: `docs/grammar-review/reports/b5e38634c3959ede9b892a22.json`

### ES 103

Итого: 72/72 (`fixed` 18, `ok` 54); phase `awaiting_verification`.

- Source: `courses/spanish-grammar/chapters/103.es.grammar.complex_sentences_connecting.cuando_antes_de_despues_de_mientras_time_relations/03-questions.json`
  Report: `docs/grammar-review/reports/588f862f717e1eb0785815a9.json`

### ES 104

Итого: 72/72 (`fixed` 19, `ok` 53); phase `awaiting_verification`.

- Source: `courses/spanish-grammar/chapters/104.es.grammar.complex_sentences_connecting.relative_clauses_que_quien_donde_cuyo_starter_use/03-questions.json`
  Report: `docs/grammar-review/reports/b374012f8a1b80e58401e7c0.json`

## Следующая партия

- EN 103: 108 вопросов — безличный пассив и reporting structures.
- EN 104: 109 вопросов — новости, процессы и сервисный causative.
- EN 105: 107 вопросов — zero conditional.
- EN 106: 107 вопросов — first conditional.
- Итого: 431 вопрос в 24 файлах; reports отсутствуют.
- Все 56 назначенных source/final/validation и embedded-путей чистые и согласованы.
- Baseline HEAD EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
