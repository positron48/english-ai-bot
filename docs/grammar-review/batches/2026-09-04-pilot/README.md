# Пилот ES 001 → EN 001, 2026-09-04

Пилот закрыт: 224/224 вопроса, 13/13 файлов — `done`. Пользователь 2026-09-04 подтвердил независимую ручную проверку: «я проверил, ок, правим дальше». Координатор записал это подтверждение в каждый отчёт после сверки fingerprints. Ниже сохранены результаты редакторского этапа; самостоятельной проверки редактором под другим именем не было.

| Банк | Прочитано | Правки содержания | Только signature | Без изменений |
| --- | ---: | ---: | ---: | ---: |
| ES 001, основные вопросы | 60 | 26 | 0 | 34 |
| ES 001, 6 тренировок | 87 | 68 | 3 | 16 |
| EN 001, основные вопросы | 22 | 8 | 0 | 14 |
| EN 001, 5 тренировок | 55 | 38 | 0 | 17 |
| Всего | 224 | 140 | 3 | 81 |

Точные source/report пути и состояние каждого файла находятся в [ACTIVE.md](../../ACTIVE.md). Решения и причины сохранены по каждому ID в существующей схеме отчётов. Checkpoint обновлялся порциями до 20 вопросов; текущие source/context fingerprints совпадают.

Редактор: Codex editor 2026-09-04. Все вопросы прочитаны вместе с outline и теорией. Проверены GrammarQuestion.vue (перемешивание вариантов, разбор reorder из correct_answer, перевод, сравнение строк), grammar_service.go (нормализация регистра/пробелов, без альтернатив через слеш) и grammar_reorder.go (доступность reorder с переводом). Это редактура одного исполнителя, а не независимая верификация.

## Исправления

- ES: вопросы с несколькими нормативными ответами (`ye/i griega`, `uve doble/doble u`, `uve/ve`, `con tilde/con acento`); ложные объяснения, объявлявшие допустимые варианты ошибочными; смешение названия буквы и её звука; ложное отсутствие w в алфавите; неоднозначное написание вымышленного слова и вопрос с двумя правильными алфавитными рядами. Убраны метаинструкции, подсказки и явные повторы.
- EN: ошибочные абсолютные правила о подлежащем/порядке слов; два допустимых времени у `He read(s) a book`; полные и сокращённые отрицания как конкурирующие правильные ответы; необязательные дополнения у read/eat/drink; неверный разбор русского «Идёт дождь»; отсутствие предложения в вопросе об объекте; неполные группы подлежащего с артиклем. Неизученные упражнения на спряжение/вопросительный порядок заменены заданиями на структуру предложения в рамках прежнего theory_block_id.
- Три исходные signature уже не соответствовали содержимому: ES training b3 q14/q17 и b5 q6. Их текст оставлен, signature пересчитаны существующей функцией курса; они отмечены `fixed` с отдельной причиной.
- Сохранены IDs, порядок, типы, difficulty, chapter/concept/theory bindings и IDs вариантов. EN chapter q19 остаётся reorder, с согласованным переводом и повествовательным порядком; связанных полей приложения не меняли.

Нормативные основания: названия букв и допустимость региональных вариантов сверены по [RAE](https://www.rae.es/espanol-al-dia/un-solo-nombre-para-cada-letra); знак ударения также называется acento gráfico/ortográfico по [DPD: tilde](https://www.rae.es/dpd/tilde), функция диэрезиса — по [DPD: diéresis](https://www.rae.es/dpd/diéresis). Исключения из обязательности подлежащего и положения перед глаголом подтверждает [Cambridge: Subjects](https://dictionary.cambridge.org/grammar/british-grammar/subjects); переходное и непереходное употребление глаголов — [Cambridge: Verb patterns](https://dictionary.cambridge.org/grammar/british-grammar/transitive-and-intransitive-verbs).

## Валидация и интеграция

Из корня:

```bash
python3 docs/grammar-review/batches/2026-09-04-pilot/check.py
go test ./internal/grammarbundle ./internal/repository -run 'TestValidateEmbeddedBundleID|TestGrammar.*Bundle|TestGrammar.*Reorder' -count=1
git diff --check
git -C courses/spanish-grammar diff --check
git -C courses/english-grammar diff --check
```

Результат: pass. Проверка пилота read-only: сверяет 224 IDs/контракта с baseline commits, решения с фактическими изменениями, JSON, уникальность вариантов, правильные ключи, theory bindings, актуальность signatures и отсутствие новых дубликатов во всём training pack каждого курса; совпадение source → final → embedded и fingerprints. Go: оба пакета прошли (5.068s / 15.166s). Скрипт `check.py` фиксирует эти технические свойства и не оценивает смысл вместо человека/LLM.

Обе изменённые главы собраны `assemble-chapter.sh <dirname>` и проверены `validate-chapter.sh <dirname>` из своего курса. `is_valid=true`, `issues=[]` до и после. ES схема в сборщике проходит. У EN старый сборщик вызывает AJV без нужного draft и показывает предупреждение; отдельная проверка draft2020 выявила одинаковое старое ограничение до/после: `chapter_test.pool_question_ids` содержит 12 элементов при `minItems=20`. Это `known_baseline`, а не успешная полная проверка схемы. См. [до](en-schema-before.log) и [после](en-schema-after.log). Форматы дат в этой проверке отключены, как в существующем ES-сборщике.

Воспроизведение baseline EN из корня (не зависит от временного файла редактора):

```bash
git -C courses/english-grammar show 5f9f7d14a684baa83ef4fe759bdacd0a97c4f563:chapters/001.en.grammar.orientation_how_to_read.subject_verb_object_in/05-final.json > /tmp/grammar-pilot-en-final-before.json
ajv validate --spec=draft2020 --validate-formats=false --all-errors -s courses/english-grammar/02-chapter-schema.json -d /tmp/grammar-pilot-en-final-before.json
ajv validate --spec=draft2020 --validate-formats=false --all-errors -s courses/english-grammar/02-chapter-schema.json -d courses/english-grammar/chapters/001.en.grammar.orientation_how_to_read.subject_verb_object_in/05-final.json
```

Обе последние команды ожидаемо возвращают 1 с одной и той же ошибкой. Baseline ES: `7d0708cdd1d7c88c790f6238a82de9b903fa1ba9`. Исходники пилота перед работой были без редакторских изменений; предыдущие изменения других глав/пакетов сохранены.

Синхронизированы строго 2 файла глав в `internal/grammarbundle` и 11 файлов тренировок в `internal/grammartrainingpack`. Перед копированием подтверждено совпадение этих embedded файлов с исходными версиями; после — побайтовое совпадение новых версий. Полная перегенерация банков не запускалась.

## Остаток и передача

Редакторских pending/blocked в пилоте нет. Проверка подтверждена пользователем; scoped `grammar-review.py check` для всех 13 файлов проходит. Общий check остаётся незакрытым из-за остальных банков.

В теории остались исходные упрощения: EN b1/b2/b5 говорят об обязательности подлежащего/фиксированном порядке слишком широко, EN b3 смешивает понятие сказуемого и глагола, ES b4 описывает ü как отдельную гласную [u]. Вопросы уточнены без изменения теории. Эти места вынесены на оценку пилота; изменение теории потребует повторной сверки зависимых вопросов и context hashes.

Пользователь разрешил продолжить редактуру следующей партии. Коммитов, push, релиза, production-команд, фонового расписания не было. Placement и очереди других типов не затронуты.
