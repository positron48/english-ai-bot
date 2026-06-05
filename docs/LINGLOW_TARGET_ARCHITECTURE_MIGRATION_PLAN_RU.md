# Linglow target architecture migration plan

Дата: 2026-06-04

Цель: привести текущее приложение `english-ai-bot` к целевой концепции Linglow без потери пользовательских данных. Итоговая модель: одно приложение, один backend, одна PostgreSQL БД, несколько курсов (`en_ru`, `es_ru`), прогресс через `user_course_id`, контент импортируется из Git-файлов в БД и в runtime читается из БД.

Текущий статус:

- Phase 0/DB-first content: grammar content, grammar training, reading catalog и speaking catalog импортируются в БД; prod runtime переключен на `CONTENT_SOURCE=db`.
- Phase 1/schema foundation: добавлена миграция `000017_linglow_course_architecture.sql` с canonical Linglow v2 таблицами и seed для `en_ru`/`es_ru`, districts, locations и theme lines.
- Phase 2/user_courses bootstrap: добавлен zero-touch startup backfill, который по `LearningConfig` создаёт отсутствующие `user_courses` для всех существующих пользователей текущей БД (`en_ru` в English, `es_ru` в Spanish) без ручного `kubectl exec`.
- Phase 3/content mapping bootstrap: добавлен zero-touch startup mapping legacy/DB-first content в `modules` и `learning_items` для grammar, reading, speaking и word sets; mapping идемпотентный и пока не меняет runtime.
- Phase 4/dual-write foundation: добавлен feature flag `LINGLOW_EVENTS_WRITE_ENABLED`; при включении online/offline grammar test attempts, Grammar Training SRS attempts, web/PWA word training review events и Telegram bot word training review events зеркалятся в `exercise_attempts` и `learning_events` non-blocking, старые таблицы остаются source of truth.
- Phase 5/attempts-events backfill foundation: добавлен `cmd/backfill_linglow_events` с dry-run по умолчанию и `--commit`; команда сверяет и дозаливает исторические `grammar_test_attempts`, `grammar_attempts`, `review_events` в `exercise_attempts` + `learning_events` через тот же idempotent writer, что и runtime dual-write.
- Phase 5/word SRS snapshot foundation: добавлен `cmd/backfill_linglow_word_srs` с dry-run по умолчанию и `--commit`; команда переносит текущий legacy state из `user_cards` в canonical `srs_items` для mapped `word_card` learning items.
- Phase 5/grammar SRS snapshot foundation: добавлен canonical `grammar_theory_block` learning item, startup mapping theory blocks из DB-first training content и `cmd/backfill_linglow_grammar_srs`, который переносит `grammar_theory_memory` в `srs_items` без схлопывания нескольких blocks в chapter.
- Phase 5/attempt-SRS linking: runtime dual-write для word/grammar SRS attempts теперь проставляет `exercise_attempts.srs_item_id`; добавлен `cmd/backfill_linglow_attempt_srs_links`, который дозаполняет `srs_item_id` для исторических attempts и переводит grammar training attempts на `grammar_theory_block` item.
- Phase 5/reading-speaking progress foundation: добавлен `cmd/backfill_linglow_media_progress`, который переносит legacy `reading_text_progress` и `speaking_attempts` в `exercise_attempts` + `learning_events` для mapped `reading_text`/`speaking_task` items.
- Phase 6/course-aware read API foundation: добавлен protected endpoint `GET /api/learning/course`, который отдаёт карту course -> districts -> locations -> modules -> learning_items из Linglow v2 таблиц; во фронте добавлен `/city` read-only экран поверх этого API.
- Phase 6/course selection API foundation: добавлены protected endpoints `GET /api/courses`, `GET /api/user/courses/current`, `POST /api/user/courses/select`; текущий курс хранится в `users.settings_json.current_course_code`, selection идемпотентно создаёт `user_courses`.
- Phase 6/current-course city API: `GET /api/linglow/city` добавлен как новый endpoint карты города; `/api/learning/course` оставлен совместимым alias. Оба endpoint используют resolution `course_code query -> users.settings_json.current_course_code -> legacy env default`, а explicit read не меняет текущий курс пользователя.

## 1. Текущая точка

Что уже есть и что нужно сохранить:

- Go backend, PostgreSQL, Vue webapp, PWA/offline support, Android embedded WebView APK.
- Отдельные prod-инстансы English и Spanish в k3s: разные namespace, домены, БД, секреты, TTS PVC.
- Реальные пользовательские данные: пользователи, word progress, grammar progress, attempts, reading progress, speaking/chat data, offline sync attempts.
- Контентные источники в Git: grammar bundles, grammar training packs, reading catalog, speaking catalog, word frequency sets.
- Первый DB-first шаг уже сделан: `cmd/import_learning_content` импортирует grammar content, grammar training pack, reading catalog, speaking catalog в БД; runtime может переключаться через `CONTENT_SOURCE=db`.

Главное несовпадение с новой концепцией:

- Сейчас язык изолируется окружением и отдельной БД. В целевой модели язык является курсом в общей БД.
- Сейчас прогресс в основном привязан к `user_id`. В целевой модели весь прогресс должен быть привязан к `user_course_id`.
- Сейчас режимы живут отдельно: words, grammar, reading, speaking. В целевой модели они должны сходиться в единые `learning_items`, `srs_items`, `exercise_attempts`, `learning_events`.
- Сейчас UI режимный. В целевой модели главный UX - город, districts, locations/buildings, daily route, revisit content.

## 2. Принципы миграции

1. Не делать big bang rewrite.
   Сначала добавляем новую модель рядом со старой, потом постепенно переводим чтение/запись.

2. Никаких destructive migrations.
   Старые таблицы не удаляются и не переименовываются до конца миграции. Все новые связи добавляются nullable/parallel-режимом, затем backfill, затем включение runtime.

3. Каждый этап должен иметь rollback.
   Если новая логика ломается, можно вернуться к старым таблицам/старым API/`CONTENT_SOURCE=bundle`.

4. Все импорты должны быть идемпотентными.
   Повторный запуск импортера не должен создавать дубли или портить прогресс.

5. Источник истины для authored content остается в Git.
   Runtime source of truth для приложения - PostgreSQL после успешного импорта.

6. Старые English/Spanish БД считаются production data sources до финальной консолидации.
   Их нельзя изменять необратимо без backup и сверки счетчиков.

## 3. Целевая доменная модель

Минимальный набор таблиц для Linglow v2:

- `courses`
  Курс как продуктовая единица: `en_ru`, `es_ru`, target language, teaching locale, UI locale, slug, status.

- `user_courses`
  Связь пользователя с курсом. Все progress/state tables должны ссылаться на эту сущность.

- `districts`
  Уровни города: A0, A1, A2, B1, B2, C1.

- `locations`
  Buildings внутри district: Grammar Building, Word Market, Reading Spot, Conversation Hub, Review Station, Mistake Workshop.

- `theme_lines`
  Тематические линии, проходящие через уровни: travel, food/cafe, daily life, work и т.д.

- `modules`
  Контентные блоки внутри location/district/theme.

- `learning_objectives`
  Измеримые цели: grammar concept, vocab set, reading skill, speaking pattern.

- `learning_items`
  Унифицированные элементы обучения: word, grammar concept, grammar question, reading question, speaking task, chat correction, pronunciation item.

- `srs_items`
  Единый SRS-state для любого `learning_item`.

- `exercise_attempts`
  Унифицированные попытки упражнений.

- `learning_events`
  Append-only журнал всех учебных действий.

- Aggregate tables:
  `daily_course_stats`, `mode_daily_stats`, `district_progress`, `learning_item_stats`, `content_performance_stats`.

## 4. Фаза 0. Safety baseline

Цель: зафиксировать состояние данных и включить наблюдаемость перед изменениями.

Шаги:

1. Снять fresh backup English и Spanish БД через существующий k3s backup или ручной `pg_dump`.
2. Зафиксировать счетчики по основным таблицам обеих БД:
   - `users`
   - `word_cards`
   - `training_cards`
   - `user_cards`
   - `review_events`
   - `grammar_progress`
   - `grammar_test_attempts`
   - `grammar_attempts`
   - `reading_text_progress`
   - `speaking_sessions`
   - `speaking_attempts`
3. Сохранить SQL-скрипт audit queries в `docs` или `scripts`.
4. Проверить, что полный `make check` проходит на master.
5. Убедиться, что текущий prod rollout работает без pending migrations.

Готовность:

- Есть backup обеих БД.
- Есть baseline-счетчики.
- Есть команда для повторной сверки.
- `make check` проходит.

Rollback:

- Пока изменений runtime нет. Rollback не нужен.

## 5. Фаза 1. Course-aware schema рядом со старой моделью

Цель: добавить сущности Linglow v2 без изменения поведения текущего приложения.

Шаги:

1. Добавить SQL migration:
   - `courses`
   - `user_courses`
   - `districts`
   - `locations`
   - `theme_lines`
   - `modules`
   - `learning_objectives`
   - `learning_items`
   - `srs_items`
   - `exercise_attempts`
   - `learning_events`
   - aggregate tables

2. Засидить курсы:
   - `en_ru`: target `en`, teaching locale `ru`, UI locale `ru`, city name `Luminaria City`
   - `es_ru`: target `es`, teaching locale `ru`, UI locale `ru`, city name `Ciudad Luminaria`

3. Засидить districts:
   - A0 `Puerta de la Chispa`
   - A1 `Plaza Clara`
   - A2 `Barrio Vivo`
   - B1 `Puentes del Relato`
   - B2 `Distrito Alto`
   - C1 `Campus de Maestria`

4. Засидить locations для каждого district:
   - grammar
   - word_market
   - reading
   - conversation
   - review_station
   - mistake_workshop

5. Добавить repository/service слой для чтения новой модели, но пока не подключать к пользовательскому runtime.

6. Добавить тесты миграций и repository tests.

Готовность:

- Новые таблицы создаются на пустой и существующей БД.
- Старые экраны работают без изменений.
- `make check` проходит.

Rollback:

- Runtime не использует новые таблицы. При проблеме можно откатить код; данные в новых таблицах не влияют на старую работу.

## 6. Фаза 2. Backfill `user_courses`

Цель: создать course scope для существующих пользователей без изменения старого progress write path.

Переходная стратегия:

- В текущей English БД всем пользователям создается `user_course` для `en_ru`.
- В текущей Spanish БД всем пользователям создается `user_course` для `es_ru`.
- В финальной unified БД эти записи будут объединены.

Шаги:

1. Написать команду:
   `cmd/backfill_user_courses`

2. Команда должна:
   - принимать `--course-code=en_ru|es_ru`;
   - работать в dry-run по умолчанию;
   - с `--commit` создавать отсутствующие `user_courses`;
   - быть идемпотентной;
   - печатать счетчики: users scanned, created, existing, skipped.

3. Добавить уникальный constraint:
   `UNIQUE(user_id, course_id)`.

4. Добавить audit query:
   пользователи без `user_course` для текущего курса.

5. Прогнать на staging/local restore.

6. Прогнать в prod English и Spanish после backup.

Готовность:

- Для каждого существующего пользователя есть `user_course`.
- Старые таблицы не изменены.
- Счетчики до/после сохранены.

Rollback:

- Старый runtime не зависит от `user_courses`.
- При ошибке можно удалить только созданные rows по `course_id`, если команда была ошибочной.

## 7. Фаза 3. Content mapping в `learning_items`

Цель: связать существующий импортированный контент с новой универсальной моделью.

Шаги:

1. Расширить importer `cmd/import_learning_content` или добавить новый `cmd/import_linglow_content`.

2. Для grammar chapters создать:
   - `modules` под district/location grammar;
   - `learning_objectives` по concept/theory blocks;
   - `learning_items` для grammar chapter, theory block, training question.

3. Для word sets создать:
   - `modules` под word_market;
   - `learning_items` типа `word`;
   - связь с source word card/training card через stable external key.

4. Для reading catalog создать:
   - `modules` под reading;
   - `learning_items` типа `reading_text` и `reading_question`.

5. Для speaking catalog создать:
   - `modules` под conversation;
   - `learning_items` типа `speaking_task`.

6. Для каждого item хранить:
   - `course_id`
   - `district_id`
   - `location_id`
   - optional `theme_line_id`
   - `source_kind`
   - `source_id`
   - `content_hash`
   - `status`

7. Добавить проверку уникальности:
   `UNIQUE(course_id, source_kind, source_id)`.

8. Сохранить старые raw JSON таблицы как source payload на переходный период.

Готовность:

- Количество `learning_items` соответствует ожидаемым счетчикам импортированного контента.
- Повторный импорт не создает дубли.
- Старый runtime продолжает работать.

Rollback:

- Можно отключить новый importer.
- Старые content tables и bundle fallback остаются.

## 8. Фаза 4. Dual-write progress events

Цель: начать писать новые `learning_events` и `exercise_attempts` параллельно со старыми таблицами.

Шаги:

1. Ввести feature flag:
   `LINGLOW_EVENTS_WRITE_ENABLED=false` по умолчанию.

2. Добавить event writer:
   - принимает `user_course_id`;
   - принимает `learning_item_id`, если mapping найден;
   - не ломает основной request при ошибке записи event, но логирует error и метрику.

3. Подключить dual-write для:
   - word training attempts;
   - grammar test attempts;
   - grammar SRS attempts;
   - reading progress;
   - speaking attempts;
   - AI chat corrections, если есть stable mapping.

4. Добавить idempotency:
   - использовать существующие `client_attempt_id` для offline;
   - для online генерировать request/attempt id.

5. Включить flag на staging.

6. Сравнить:
   - старые attempts count;
   - новые `exercise_attempts`;
   - новые `learning_events`.

7. Включить flag в prod сначала на одном инстансе, затем на обоих.

Готовность:

- Новые events пишутся без заметного роста ошибок.
- Старые экраны и прогресс не меняют поведение.
- Offline sync остается идемпотентным.

Rollback:

- Выключить `LINGLOW_EVENTS_WRITE_ENABLED`.
- Старые tables остаются source of truth.

## 9. Фаза 5. Backfill historical progress в новую модель

Цель: перенести историю из старых таблиц в `learning_events`, `exercise_attempts`, `srs_items`, не потеряв прогресс.

Шаги:

1. Написать команды:
   - `cmd/backfill_linglow_events` - готово для attempts/events из `grammar_test_attempts`, `grammar_attempts`, `review_events`;
   - `cmd/backfill_linglow_word_srs` - готово для `srs_items`/word state snapshots из `user_cards`;
   - `cmd/backfill_linglow_grammar_srs` - готово для `srs_items`/grammar theory-block snapshots из `grammar_theory_memory`;
   - `cmd/backfill_linglow_attempt_srs_links` - готово для связи historical `exercise_attempts` с `srs_items`;
   - `cmd/backfill_linglow_media_progress` - готово для `reading_text_progress` и `speaking_attempts`.

2. Каждая команда:
   - dry-run по умолчанию;
   - `--commit` для записи;
   - course выбирается из `LearningConfig` текущего инстанса (`en_ru`/`es_ru`), если отдельный override не нужен;
   - `--since` optional;
   - идемпотентность через `exercise_attempts.source_table/source_pk` и runtime writer;
   - печатает scanned, inserted, skipped, unmapped.

3. Backfill порядок:
   - users/user_courses;
   - content/learning_items;
   - attempts/events;
   - SRS snapshots;
   - aggregates.

4. Для `srs_items`:
   - если есть старый review state, переносить интервалы/даты;
   - если state нет, восстанавливать из attempts;
   - если восстановить нельзя, создавать item без due state и логировать skipped reason.

5. Для конфликтов:
   - не перезаписывать существующий новый event;
   - сохранять `source_table`, `source_pk`, `source_hash`;
   - формировать отчет unmapped rows.

6. Прогнать backfill на копии prod БД.

7. Сверить агрегаты:
   - число пользователей;
   - число active users;
   - число attempts по режимам;
   - due items per user;
   - последние даты активности.

8. После сверки прогнать в prod.

Готовность:

- Исторический прогресс доступен через новую модель.
- Unmapped rows разобраны или явно признаны несущественными.
- Есть отчет миграции.

Rollback:

- Новые таблицы можно очистить по `backfill_run_id`.
- Старые таблицы остаются source of truth.

## 10. Фаза 6. Course-aware API

Цель: новый backend contract работает через `course_code` / `user_course_id`.

Шаги:

1. Добавить API:
   - `GET /api/courses` - готово;
   - `POST /api/user/courses/select` - готово;
   - `GET /api/user/courses/current` - готово;
   - `GET /api/linglow/city` - готово;
   - `GET /api/linglow/daily-route`
   - `POST /api/linglow/exercise-attempts`
   - `GET /api/linglow/review`
   - `GET /api/linglow/progress`

2. Старые API оставить совместимыми.

3. Ввести current course resolution:
   - из явного `course_code`;
   - из user preference;
   - из legacy env default для старых инстансов.

4. Все новые write endpoints должны писать через `user_course_id`.

5. Добавить permission checks:
   пользователь может читать/писать только свои `user_courses`.

6. Добавить integration tests:
   - один user, два courses;
   - прогресс не смешивается;
   - offline idempotency не пересекается между courses.

Готовность:

- Можно работать с `en_ru` и `es_ru` в одной БД на уровне API tests.
- Старые endpoints не сломаны.
- `make check` проходит.

Rollback:

- Не переводить frontend на новые endpoints.
- Старые endpoints остаются рабочими.

## 11. Фаза 7. Unified SRS and exercise engine

Цель: заменить разрозненные SRS/write paths единым движком.

Шаги:

1. Реализовать service:
   - `LearningItemResolver`
   - `ExerciseEngine`
   - `SRSService`
   - `MistakeService`
   - `ProgressAggregator`

2. Поддержать exercise templates:
   - 4-choice word;
   - letter assembly;
   - full word typing;
   - listening;
   - grammar drill;
   - reading question;
   - chat correction.

3. На первом этапе использовать существующие алгоритмы scoring/review interval, но хранить state в `srs_items`.

4. Добавить shadow-read сравнение:
   - старый due list vs новый due list;
   - старый mastery score vs новый confidence/stability.

5. Ввести flags:
   - `LINGLOW_SRS_READ_ENABLED`
   - `LINGLOW_SRS_WRITE_ENABLED`

6. Перевести сначала non-critical режим или internal user cohort.

Готовность:

- Новый SRS дает ожидаемые due/review результаты.
- Нет потери offline attempts.
- Старые progress tables больше не нужны для новых записей, но еще доступны для rollback.

Rollback:

- Вернуть read/write flags на legacy.

## 12. Фаза 8. City UX и Daily Route

Цель: перейти от режима "набор экранов" к Linglow city model.

Шаги:

1. Добавить course selector.

2. Добавить City Home:
   - city name;
   - districts A0-C1;
   - locked/unlocked states;
   - current daily route;
   - review pressure;
   - next openings.

3. Добавить District view:
   - buildings/locations;
   - foundation/confidence/stability вместо процента завершения;
   - weak items;
   - revisit tasks.

4. Добавить Daily Route:
   - review block;
   - grammar block;
   - vocabulary/reading/conversation block;
   - mistake workshop block, если есть ошибки.

5. Добавить Simple Mode:
   - прямой доступ к review, grammar, texts, vocab, settings;
   - без декоративного city layer для повторяющихся рабочих сценариев.

6. Сначала использовать существующие экраны внутри новых маршрутов.

7. Затем постепенно заменять старые screens на course-aware components.

Готовность:

- Новый пользователь может выбрать курс и начать A0/A1 flow.
- Старый пользователь видит свой прогресс после миграции.
- Offline grammar/word flows не ломаются.

Rollback:

- Feature flag на новый home/router.
- Старый dashboard остается доступен.

## 13. Фаза 9. Unified database migration

Цель: перейти от двух prod БД English/Spanish к одной Linglow БД.

Не начинать эту фазу, пока:

- `user_courses` есть в обеих БД.
- `learning_items` и progress backfill проверены.
- Новые course-aware API проходят integration tests.
- Есть dry-run merge report.

Шаги:

1. Создать новую PostgreSQL БД `linglow`.

2. Накатить все migrations.

3. Написать merge command:
   `cmd/merge_language_databases`

4. Merge command должен:
   - принимать English source DB;
   - принимать Spanish source DB;
   - принимать target Linglow DB;
   - работать dry-run по умолчанию;
   - маппить пользователей по stable identity;
   - маппить Telegram users осторожно: один Telegram user может иметь оба курса;
   - сохранять old ids в mapping tables.

5. Создать mapping tables:
   - `legacy_user_mappings`
   - `legacy_course_mappings`
   - `legacy_content_mappings`
   - `legacy_attempt_mappings`

6. Порядок merge:
   - users;
   - courses/user_courses;
   - content;
   - progress/events/attempts;
   - SRS;
   - subscriptions/settings;
   - chat/speaking/reading;
   - aggregates.

7. Конфликты:
   - один Telegram id в двух БД - объединять в одного user;
   - одинаковые emails/auth identities - объединять только при строгом совпадении verified identity;
   - несовпадения писать в conflict report, не угадывать.

8. Прогнать merge на staging.

9. Провести сверку:
   - users count;
   - user_courses count;
   - attempts by course;
   - active subscriptions/settings;
   - latest activity per user;
   - random sample пользователей.

10. Prod cutover:
    - включить maintenance/read-only window;
    - остановить writes или поставить queue;
    - fresh backup обеих БД;
    - final merge;
    - переключить backend `DATABASE_URL` на unified DB;
    - smoke tests;
    - снять maintenance.

Готовность:

- Один backend может обслуживать `en_ru` и `es_ru`.
- Пользователь с обоими курсами видит оба курса.
- Данные не потеряны по сверочным отчетам.

Rollback:

- До cutover: просто не переключать runtime.
- После cutover в коротком окне: вернуть `DATABASE_URL` на старые БД/деплойменты и отключить unified deployment.
- После длительного dual-write периода rollback сложнее, поэтому старые БД нужно держать read-only snapshot до стабилизации.

## 14. Фаза 10. Unified deployment and domains

Цель: один backend/app вместо отдельных English/Spanish приложений.

Шаги:

1. Добавить новый k3s app, например `linglow`.

2. Сохранить старые домены как redirects/deep links:
   - `qantrix.ru/app` -> Linglow course `en_ru`;
   - `es.qantrix.ru/app` -> Linglow course `es_ru`.

3. Ввести canonical public URL для Linglow.

4. Обновить config:
   - убрать обязательность `LEARNING_APP_CODE` как runtime tenant;
   - оставить default course только как fallback для legacy links.

5. Настроить secrets:
   - один `DATABASE_URL`;
   - общие AI secrets;
   - course-specific non-secret config в БД.

6. Обновить backup:
   - dump unified PostgreSQL;
   - TTS/cache strategy: либо общий PVC с course/language shard, либо object storage.

7. Обновить observability:
   - labels by `course_code`;
   - dashboards by course.

Готовность:

- Новый deployment обслуживает оба курса.
- Старые links не ломаются.
- Backup покрывает новую БД и критичные файлы.

Rollback:

- Вернуть ingress на старые deployments.
- Unified DB остается snapshot/standby до следующей попытки.

## 15. Фаза 11. Android strategy

Цель: перейти от отдельных English/Spanish APK к одному Linglow APK.

Варианты:

1. Transitional:
   - оставить старые APK `ru.qantrix.english` и `ru.qantrix.spanish`;
   - они открывают unified backend с default course;
   - использовать их как совместимость для текущих пользователей.

2. Target:
   - новый APK `ru.qantrix.linglow`;
   - course selector внутри приложения;
   - offline storage keyed by `course_code`.

Рекомендованный путь:

1. Сначала unified backend/web.
2. Затем выпустить новый `ru.qantrix.linglow`.
3. Старые APK оставить на период миграции как wrappers/deep links.
4. Не полагаться на перенос IndexedDB между package names. Offline cache можно потерять, серверный прогресс терять нельзя.

Готовность:

- Новый Android app открывает оба курса.
- Offline preload хранится отдельно по `course_code`.
- Старые APK не ломают существующих пользователей.

Rollback:

- Старые APK остаются доступными.

## 16. Фаза 12. Cleanup

Цель: убрать legacy только после стабильной работы новой модели.

Не начинать cleanup, пока:

- unified backend работает в prod минимум 2-4 недели;
- нет критичных unmapped rows;
- все новые writes идут через `user_course_id`;
- старые endpoints либо удалены из frontend, либо явно помечены legacy.

Шаги:

1. Перевести старые progress tables в read-only compatibility.
2. Удалить bundle runtime dependency, оставить только importer и tests.
3. Удалить legacy env-driven app split.
4. Удалить старые k3s deployments после срока совместимости.
5. Обновить docs/runbooks.

Готовность:

- Нет runtime reads из legacy progress tables.
- Нет prod traffic на старые deployments.
- Backup/restore проверены для unified модели.

Rollback:

- На этой фазе rollback уже дорогой. Делать только после периода стабильности.

## 17. Data protection checklist

Перед каждой фазой с записью в prod:

- Fresh backup сделан.
- Есть dry-run output.
- Есть счетчики до изменения.
- Есть expected counters после изменения.
- Есть rollback flag или понятная rollback command.
- Есть audit query для проверки.
- Изменение сначала прошло на копии prod или staging.
- После изменения сохранен отчет: дата, commit, команда, counts, ошибки.

Минимальный набор audit queries должен проверять:

- users without user_courses;
- user_courses without course;
- progress rows without mapped learning_item;
- duplicate idempotency keys;
- events without user_course_id;
- srs_items without learning_item_id;
- attempts count mismatch by mode/course;
- latest activity mismatch per user/course.

## 18. Рекомендуемый порядок ближайших работ

1. Закрепить DB-first content в prod:
   - выполнить import для English и Spanish;
   - сверить counts;
   - переключить `CONTENT_SOURCE=db`;
   - оставить bundle fallback.

2. Добавить Linglow v2 schema и seed courses/districts/locations.

3. Сделать `backfill_user_courses`.

4. Сделать content mapping в `learning_items`.

5. Включить dual-write `learning_events` под feature flag.

6. Сделать backfill historical progress.

7. Добавить course-aware API.

8. Собрать первый City Home/Daily Route поверх новых API.

9. После стабилизации готовить unified DB merge.

10. После unified DB выпускать один Linglow Android app.

## 19. Что не делать

- Не объединять English и Spanish БД до появления `user_courses` и mapping tables.
- Не удалять старые progress tables до полного backfill и периода стабильности.
- Не переименовывать старые колонки в первой волне.
- Не переносить Android в один package до unified backend и course-aware offline storage.
- Не делать city UI как чистую декоративную оболочку без новой data model.
- Не считать `CONTENT_SOURCE=db` финальной архитектурой: это только первый слой DB-first content.

## 20. Критерий завершения миграции

Миграция считается завершенной, когда:

- В prod работает один backend и одна БД.
- Пользователь может иметь несколько `user_courses`.
- Все progress writes идут через `user_course_id`.
- Контент runtime читается из БД.
- Git-файлы остаются authoring source и импортируются идемпотентно.
- SRS работает поверх `learning_items`.
- Все attempts/events пишутся в unified tables.
- City Home, Districts, Daily Route и Simple Mode используют course-aware API.
- Android target app один, с выбором курса.
- Старые English/Spanish deployments больше не нужны для штатной работы.
