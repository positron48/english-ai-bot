# Редактура EN 075–078

2026-09-05 вычитан 481 вопрос: 251 основной и 230 тренировочных в 27
исходных файлах. Исправлено 187 вопросов, 294 оставлены без изменений. Все
исправления затрагивают редакторское содержание. Все отчёты имеют
`awaiting_verification`; независимая проверка остаётся `pending`.

| Глава | Тема | Всего | Исправлены | Без изменений |
| --- | --- | ---: | ---: | ---: |
| EN 075 | предлоги времени | 102 | 24 | 78 |
| EN 076 | предлоги места | 127 | 68 | 59 |
| EN 077 | предлоги движения | 131 | 52 | 79 |
| EN 078 | глаголы и прилагательные с предлогами | 121 | 43 | 78 |
| Всего | | 481 | 187 | 294 |

Редактор: `Codex editor 2026-09-05`. Правки локальные. Baseline сабмодуля EN —
`1ec076ada14b88fba983a482f3cfcdc6d37ee7ca`. Все назначенные исходники до
начала партии были чистыми, reports отсутствовали. Placement и отдельная очередь
форм глаголов исключены.

## Существенные исправления

- EN 075: устранены двусмысленные задания на `during`/`for`, смешение русской
  фразы с английским пропуском и варианты с несколькими правильными исправлениями.
  Контексты для `by`/`until` теперь явно различают дедлайн и продолжение до момента.
- EN 076: задания на `at`/`in`/`on` получили английские предложения и явную
  пространственную логику. Исправлено механическое правило «between только для
  двух»: отдельные перечисленные ориентиры отличаются от неразделённой группы.
  Для `opposite`, `next to`, `near`, `behind` и `in front of` добавлены наблюдаемые
  признаки положения.
- EN 077: исправлен неверный ключ `get off the car` на `get out of the car`.
  В заданиях на `to`/`into`, `across`/`through`, `off`/`out of` и `along`/`past`
  убраны нормальные фразы, ошибочно названные ошибками, и добавлены признаки
  входа, выхода, поверхности, маршрута и конечной точки.
- EN 078: устранена ложная взаимозаменяемость `depend on` и `rely on`, признано
  отдельное употребление `wait on`, а `different from` описано как нейтральная
  международная форма наряду с региональными `different than` и `different to`.
  Уточнены различия `agree with`/`agree to` и цель сочетаний `apply`, `ask`,
  `pay` + `for`.

IDs, порядок, типы, difficulty, theory bindings и choice IDs сохранены. Для
тренировок штатно пересчитаны signatures; новых дубликатов во всём EN-курсе нет.

## Замечания к теории

- EN 075 использует удобные базовые пары `during + event` и `for + duration`, но
  в живой речи выбор также зависит от того, представлен ли период как рамка события
  или как измеряемая длительность.
- EN 076 сводит `at`/`in`/`on` к точке, объёму и поверхности. Это полезная модель,
  однако одно место допускает разные предлоги в зависимости от того, выступает ли
  оно точкой назначения, учреждением, территорией или физическим объёмом.
- EN 077 аналогично требует контекста для различения назначения и входа, а также
  открытой поверхности и пространства, через которое проходит маршрут.
- EN 078 не должна представлять `different from` как единственную допустимую форму
  или считать глаголы с одним предлогом автоматически взаимозаменяемыми по смыслу.

Теория не входит в текущий scope и не изменялась.

## Проверки и интеграция

- Все четыре главы собраны и проходят `validate-chapter.sh` и явный AJV Draft
  2020. Raw AJV без выбора draft2020 во всех главах воспроизводит прежнюю ошибку
  метасхемы; логи до и после совпадают. Служебные `05-validation.json`
  восстановлены из baseline после проверок.
- [check.py](check.py) проверил 481 контракт и решение против baseline,
  актуальные fingerprints, ответы, theory bindings, signatures, отсутствие новых
  дубликатов и равенство source → final → embedded. Итог: [check.log](check.log),
  exit 0.
- Четыре final-главы и 23 тренировочных файла точечно синхронизированы с
  embedded. Неопросные поля final сохранены, кроме штатного `updated_at` сборки.
- Все 27 scoped-check ожидаемо возвращают exit 1 только из-за незавершённой
  независимой проверки: [reports-check.log](reports-check.log).
- Целевые Go-тесты прошли: [go-tests.log](go-tests.log). Три проверки
  `git diff --check` прошли: [diff-check.log](diff-check.log).
- Все 13 тестов механизма учёта прошли: [review-tool-tests.log](review-tool-tests.log).
- Контрольный снимок подтвердил сохранность всех 3205 ранее изменённых сторонних
  файлов: [preservation.log](preservation.log).

Общее редакторское покрытие: 18095/29095 вопросов основных банков, из них 224
`done` и 17871 `awaiting_verification`; 11000 ещё не прошли первичную редактуру.
Отдельная очередь форм глаголов ES содержит 34848 карточек. Следующая партия —
ES 077–080: 366 вопросов в 16 файлах. Все назначаемые исходники чистые, reports
отсутствуют. Полная цель обоих курсов остаётся активной.

## Закрытые исходники EN 075–078

- Source: `courses/english-grammar/chapters/075.en.grammar.prepositions_verb_patterns.time_prepositions_at_in/03-questions.json`
  Report: `docs/grammar-review/reports/b01353440fc7f15dd7700aee.json`
  Редактура: 52/52 (`fixed` 3, `ok` 49); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/075/en.01.b1_at_time_points.questions.json`
  Report: `docs/grammar-review/reports/8d7a680fc33fc18512c85038.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/075/en.02.b2_on_days_dates.questions.json`
  Report: `docs/grammar-review/reports/dda22b4b2ac6c172135c40b6.json`
  Редактура: 10/10 (`fixed` 1, `ok` 9); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/075/en.03.b3_in_periods_parts_of_day.questions.json`
  Report: `docs/grammar-review/reports/e0579d72cf27e0c408b57c05.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/075/en.04.b4_by_vs_until.questions.json`
  Report: `docs/grammar-review/reports/e8a44016e19e800f1408fdb7.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/075/en.05.b5_during_vs_for.questions.json`
  Report: `docs/grammar-review/reports/a85281deb89881d8919d590e.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/076.en.grammar.prepositions_verb_patterns.place_prepositions_at_in/03-questions.json`
  Report: `docs/grammar-review/reports/e49e104cfe9fd062de7b1c5a.json`
  Редактура: 67/67 (`fixed` 13, `ok` 54); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/076/en.01.b1_place_logic_point_area_surface.questions.json`
  Report: `docs/grammar-review/reports/8b895a17969d5da3b472ef03.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/076/en.02.b2_place_at_points.questions.json`
  Report: `docs/grammar-review/reports/a4c2260e93401ead7df2d6ff.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/076/en.03.b3_place_in_enclosed.questions.json`
  Report: `docs/grammar-review/reports/a223abb3027c8fe3cb1da44d.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/076/en.04.b4_place_on_surfaces.questions.json`
  Report: `docs/grammar-review/reports/ba5dfbfb3e5ddbdcd0a20791.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/076/en.05.b5_place_between_among.questions.json`
  Report: `docs/grammar-review/reports/8867cf76eaa0a4ba152eef2a.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/076/en.06.b6_place_next_to_near.questions.json`
  Report: `docs/grammar-review/reports/c88768f0fc1eafe63ba4a129.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/077.en.grammar.prepositions_verb_patterns.movement_to_into_across/03-questions.json`
  Report: `docs/grammar-review/reports/2221f5942ee48c65df948dbb.json`
  Редактура: 71/71 (`fixed` 14, `ok` 57); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/077/en.01.b1_move_direction_to_into.questions.json`
  Report: `docs/grammar-review/reports/ed2a3f2f107221541ba7e271.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/077/en.02.b2_move_across_through.questions.json`
  Report: `docs/grammar-review/reports/fd6ac417809b2869b990eca7.json`
  Редактура: 10/10 (`fixed` 4, `ok` 6); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/077/en.03.b3_move_out_of.questions.json`
  Report: `docs/grammar-review/reports/640e2251a00437fa9d1369d6.json`
  Редактура: 10/10 (`fixed` 2, `ok` 8); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/077/en.04.b4_move_off.questions.json`
  Report: `docs/grammar-review/reports/aa373dcad3eb7f032e644233.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/077/en.05.b5_move_along_past.questions.json`
  Report: `docs/grammar-review/reports/3d7eb74619c1d948dd5cbef0.json`
  Редактура: 10/10 (`fixed` 2, `ok` 8); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/077/en.06.b6_move_into_vs_in.questions.json`
  Report: `docs/grammar-review/reports/847172301a13b2ab2efca96d.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/chapters/078.en.grammar.prepositions_verb_patterns.verb_preposition_look_for/03-questions.json`
  Report: `docs/grammar-review/reports/5de488b27eaf5a4a332baff1.json`
  Редактура: 61/61 (`fixed` 9, `ok` 52); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/078/en.01.b1_verb_prep_search_listen.questions.json`
  Report: `docs/grammar-review/reports/313eaf7c17259e5e639fda5b.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/078/en.02.b2_verb_prep_depend_on.questions.json`
  Report: `docs/grammar-review/reports/0d383d61a83a5c553f07b1aa.json`
  Редактура: 10/10 (`fixed` 0, `ok` 10); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/078/en.03.b3_verb_prep_wait_for.questions.json`
  Report: `docs/grammar-review/reports/45ba8e91851cd17185b72a72.json`
  Редактура: 10/10 (`fixed` 5, `ok` 5); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/078/en.04.b4_verb_prep_apply_for.questions.json`
  Report: `docs/grammar-review/reports/2641dbcb2e63480af8b97f40.json`
  Редактура: 10/10 (`fixed` 10, `ok` 0); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/078/en.05.b5_verb_prep_agree_with.questions.json`
  Report: `docs/grammar-review/reports/fdbe516f7a6e75873b2a90f7.json`
  Редактура: 10/10 (`fixed` 7, `ok` 3); phase: `awaiting_verification`.
- Source: `courses/english-grammar/training_pack/chapters/078/en.06.b6_verb_prep_different_from.questions.json`
  Report: `docs/grammar-review/reports/af72d86c6ba4fc23f70368ce.json`
  Редактура: 10/10 (`fixed` 7, `ok` 3); phase: `awaiting_verification`.
