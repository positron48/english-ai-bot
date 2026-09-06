# Редактура ES 085–088

2026-09-05 вычитано 288 вопросов основных банков в 4 исходных файлах.
Исправлено 52 вопроса, 236 оставлены без изменений. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| ES 085 | глаголы с предлогами | 72 | 15 | 57 |
| ES 086 | прилагательные с предлогами | 72 | 14 | 58 |
| ES 087 | направления, причины, цели и отношения | 72 | 12 | 60 |
| ES 088 | прямые объектные местоимения `lo/la/los/las` | 72 | 11 | 61 |
| Всего | | 288 | 52 | 236 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля ES —
`501044644954d5aaa474f7d1eaee99518a4ebd9d`. Все назначенные исходники до
начала партии были чистыми, reports отсутствовали. Placement и отдельная очередь
форм глаголов исключены.

## Существенные исправления

- ES 085: исправлены неверные объяснения различий между предлогами, разведены
  `pensar en` и `pensar de`, `recordar` и `acordarse de`. Уточнено, почему
  `se trata de llamar` не выражает личную попытку; исправлены примеры с
  `dedicarse a`, `insistir en`, `alegrarse de`, `fiarse de` и `contar con`.
- ES 086: убраны ложные противопоставления допустимых предлогов и региональных
  вариантов `diferente a/de`, `distinto a/de`. Исправлены неоднозначные отвлекающие
  ответы, неестественные примеры и ошибочный пропуск для `parecido a este`.
- ES 087: `pensar sobre` больше не объявляется ошибкой. Исправлены значения
  адресатной `a`, пример канала `por teléfono`, контексты `para` и `contento con`.
  Ключ q61 теперь действительно ведёт к `Gracias por`; фраза q63 больше не
  противоречит собственному смыслу причины опоздания.
- ES 088: исправлен ключ q41, который раньше выбирал вариант с повтором полных
  имён вместо их замены местоимениями. Уточнена нормативная вариативность leísmo,
  устранены неоднозначные варианты и заменён спорный q63 на однозначную ошибку
  порядка `he lo leído`.

IDs, порядок, типы, difficulty, theory bindings, choice IDs и IDs правильных
ответов сохранены.

## Замечания к теории и baseline

- В ES 085–087 часть моделей допускает нормативные или региональные варианты:
  `pensar sobre`, `contento de`, `diferente a`, `distinto a`, `alegrarse con/por`.
  Вопросы теперь либо признают вариативность, либо прямо называют целевую модель главы.
- В ES 088 базовая система курса использует `lo/la/los/las`; ограниченный
  нормативный leísmo с лицом мужского рода не объявляется безусловной ошибкой.
- Явный AJV Draft 2020 до и после партии возвращает exit 1 из-за существующих
  вне текущего scope несоответствий: theory-tag в ES 085 и тип `matching` в
  ES 086–088. Обычный AJV также сохраняет baseline exit 1 из-за незагруженной
  meta-schema Draft 2020-12. Вывод до и после каждой главы совпадает дословно.
- Внутренний `validate-chapter.sh` проходит все четыре главы. Теория не изменялась.

## Проверки и интеграция

- Все четыре главы собраны; профили всех трёх валидаторов дословно совпадают с baseline.
  Служебные validation-файлы восстановлены.
- [check.py](check.py) проверил 288 контрактов и решений, fingerprints, ответы,
  theory bindings и равенство source → final → embedded: [check.log](check.log), exit 0.
- Все четыре final-главы точечно синхронизированы с embedded. Неопросные поля final
  сохранены, кроме штатного `updated_at` сборки.
- Четыре scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Три проверки
  `git diff --check` прошли: [diff-check.log](diff-check.log).
- Все 13 тестов механизма учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Контрольный снимок подтвердил сохранность всех 3540 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 19979/29095 вопросов основных банков, из них 224
`done` и 19755 `awaiting_verification`; 9116 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
EN 087–090: 450 вопросов в 25 файлах. Reports отсутствуют. В EN 089 уже были
локальные изменения source/final/validation и embedded-копии; перед началом их
нужно включить в baseline и сохранить. Полная цель обоих курсов остаётся активной.

## Закрытые исходники ES 085–088

- Source: `courses/spanish-grammar/chapters/085.es.grammar.prepositions_verb_patterns.verb_plus_preposition_patterns_pensar_en_depender_de/03-questions.json`
  Report: `docs/grammar-review/reports/ea91d276dce2fae754de8379.json`
  Редактура: 72/72 (`fixed` 15, `ok` 57); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/086.es.grammar.prepositions_verb_patterns.adjective_plus_preposition_patterns_contento_con_parecido_lleno/03-questions.json`
  Report: `docs/grammar-review/reports/8e57b8783aa3de865344fbd2.json`
  Редактура: 72/72 (`fixed` 14, `ok` 58); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/087.es.grammar.prepositions_verb_patterns.build_speak_directions_reasons_goals_relationships/03-questions.json`
  Report: `docs/grammar-review/reports/03bb78e95188ea776dd2ffa1.json`
  Редактура: 72/72 (`fixed` 12, `ok` 60); phase: `awaiting_verification`.
- Source: `courses/spanish-grammar/chapters/088.es.grammar.pronouns_direct_indirect.direct_object_pronouns_lo_la_los_las/03-questions.json`
  Report: `docs/grammar-review/reports/37713727c7b07bd089ff0af3.json`
  Редактура: 72/72 (`fixed` 11, `ok` 61); phase: `awaiting_verification`.

## Следующая партия

- EN 087: 99 вопросов в 6 файлах — defining relative clauses.
- EN 088: 109 вопросов в 6 файлах — non-defining relative clauses и запятые.
- EN 089: 113 вопросов в 6 файлах — итоговая письменная практика раздела.
- EN 090: 129 вопросов в 7 файлах — форма Present Perfect.
- Итого: 450 вопросов в 25 файлах; reports отсутствуют.
- EN 087, 088 и 090 чистые. В EN 089 уже изменены `03-questions.json`,
  `05-final.json`, `05-validation.json` и соответствующая embedded-копия; текущие
  final и embedded совпадают. Эти изменения считаются входным состоянием следующей партии.
- Baseline HEAD EN: `1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`.
