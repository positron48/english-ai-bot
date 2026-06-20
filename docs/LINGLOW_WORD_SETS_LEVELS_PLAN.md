# Привязка словарных наборов (word sets) к уровню изучения языка (Linglow / es_ru)

## Context

В Linglow есть две параллельные модели контента:

1. **Legacy "word sets" модель** — `word_set_categories` → `word_sets` → `word_set_items` → `word_cards`. Создавалась миграциями `000002`/`000004` для испанского курса. У набора нет понятия уровня (CEFR), он не привязан к району города.
2. **Новая Linglow v2 модель** — `courses` → `districts` (1 район = 1 уровень A0..C1) → `locations` (одна из них `word_market`) → `modules`/`learning_items`. Слова из word_sets синхронизируются в эту модель функцией `CourseRepository.MapLegacyContent` (`internal/repository/course_repository.go:2594`), вызываемой один раз при старте бота (`internal/bot/bot.go:143`).

Корень проблемы: в `mapWordSetModulesSQL` и `mapWordCardItemsSQL` (`course_repository.go:2883-2939`) уровень **захардкожен как `'A0'`** (`'A0' AS level` в CTE `src`), потому что у `word_sets`/`word_set_categories` нет колонки уровня. Поэтому весь словарный контент сейчас стекается в `word_market` района A0, прогресс по `word_market` в других районах всегда пуст, а district map считает "освоено" только через `learning_items`/`srs_items` (см. `CityMapView.vue`, коммит `1562e1b3`), не через legacy `user_word_knowledge`.

Задача — убрать испанские наборы как контент, дать наборам/категориям явный CEFR-уровень, пересоздать наборы миграциями с привязкой к уровню+категориям/подкатегориям (контент готовит пользователь сам, добавлять в миграции пока не нужно — просто схема + механизм), завести синхронизацию словарных наборов с районом нужного уровня, и научить админку всему этому, сохранив текущий пайплайн генерации карточек/тренировочных карточек/озвучки.

## Изменения в БД (новые миграции, начиная с `000032`)

1. **`000032_word_sets_add_level.sql`**
   - `ALTER TABLE word_set_categories ADD COLUMN IF NOT EXISTS level_code TEXT;`
   - `ALTER TABLE word_sets ADD COLUMN IF NOT EXISTS level_code TEXT;`
   - `CHECK (level_code IS NULL OR level_code IN ('A0','A1','A2','B1','B2','C1'))` на обеих таблицах.
   - Индексы `idx_word_set_categories_course_level (course_code, level_code)`, `idx_word_sets_course_level (course_code, level_code)`.
   - Наборы наследуют уровень от категории, если у набора уровень не задан явно (бизнес-правило в коде, не в БД — см. ниже).

2. **`000033_delete_spanish_word_sets.sql`**
   - Удаляет весь словарный контент испанского курса (`course_code = 'es_ru'` либо, если course_code исторически NULL у старых строк 000002/000004 — нужно дополнительно матчить по `course_code IS NULL` для записей, которые остались нетронутыми с момента сидинга, ИЛИ просто фильтровать по тому, что реально принадлежит es_ru через сопоставление с уже промигрированными `course_code`).
   - Порядок: `DELETE FROM word_set_items WHERE word_set_id IN (SELECT id FROM word_sets WHERE course_code = 'es_ru')`, затем `DELETE FROM word_sets WHERE course_code = 'es_ru'`, затем `DELETE FROM word_set_categories WHERE course_code = 'es_ru'`.
   - Это НЕ удаляет `word_cards`/`training_cards`/`tts_generation_status` (они общие по слову, могут быть переиспользованы новыми наборами) — только наборы/категории/привязки.
   - Перед этим стоит явно проверить в проде, что у старых строк 000002/000004 `course_code` действительно заполнен (после 000022 добавлен бэкафилл при старте, нужно убедиться что есть аналог для word_set_categories/word_sets — если нет, на это нужно либо добавить бэкафилл в эту же миграцию, либо матчить по `course_code IS NULL AND target_lang='es'` контексту приложения при старте). План: проверить во время реализации, как именно es_ru backfill сейчас работает для `word_sets.course_code`, и при необходимости включить `UPDATE ... SET course_code='es_ru' WHERE course_code IS NULL` ограниченный по содержимому, созданному 000002/000004 (по списку известных названий категорий/наборов), перед удалением — чтобы не затронуть на будущее английский контент, если он когда-либо тоже использовал word_sets с NULL course_code.

## Изменения в Go-коде

### Модели (`internal/models/word_set.go`)
- `WordSetCategory.LevelCode *string`, `WordSet.LevelCode *string`.

### Репозитории
- `WordSetCategoryRepository`: добавить `level_code` во все CRUD/scan-методы (`CreateCategory`, `Get*`, `GetAllCategoriesForCourse`, `UpdateCategory*`).
- `WordSetRepository`: то же самое для `word_sets` (`CreateWordSet`, `GetWordSet*`, `ListWordSets*`, `UpdateWordSet`, scan-хелперы).
- Добавить метод получения эффективного уровня набора (`EffectiveLevelCode`): уровень набора, если задан, иначе уровень его категории. Использовать его в синхронизации с districts.

### Синхронизация word_market ↔ district (`internal/repository/course_repository.go`)
- В `mapWordSetModulesSQL` (`:2883`) заменить `'A0' AS level` на `COALESCE(ws.level_code, wsc.level_code, 'A0') AS level`, добавив `LEFT JOIN word_set_categories wsc ON wsc.id = ws.category_id` в CTE `src`.
- В `mapWordCardItemsSQL` (`:2908`) аналогично — уровень слова определяется уровнем набора (или его категории), через который слово попало в word_market; если слово в нескольких наборах разных уровней — берётся уровень того набора, через который попало (текущая логика `DISTINCT ON (wc.id) ... ORDER BY wc.id, wsi.sort_order` уже даёт детерминированный выбор одного набора per word; этого достаточно).
- Это автоматически распределит слова по нужным районам через существующий `levelDistrictJoinSQL` (`:2668`), без необходимости менять схему `districts`/`locations`/`learning_items`.
- Прогон `MapLegacyContent` уже идемпотентен (`ON CONFLICT ... DO UPDATE`) и запускается на старте — после смены уровня набора слова в `learning_items` переедут в правильный район при следующем старте бота. Дополнительно стоит дать админке кнопку «Синхронизировать с городом» (вызов `MapLegacyContentForLearning` вручную), чтобы не ждать перезапуска бота после правок в админке.

### Прогресс/«изучено» на уровне района
- Прогресс word_market уже считается через `learning_items`/`srs_items.state='mastered'`, сгруппированные по `district_id` (см. `CourseRepository`, поля `mastered_items`, `by_location` в `:872`, `:1277` и коммит `1562e1b3`). Подтверждено: легаси-прогресс слова (`user_word_knowledge.status='known'`) уже реактивно синхронизируется в `srs_items.state` через `SyncWordLearningItemForUser` (`internal/repository/linglow_word_srs_backfill_repository.go:115-160`), вызывается после каждой проверки слова из `training.go`/`training_handler.go` (`MirrorWordReview`). Эта синхронизация не зависит от происхождения слова (word_set/custom), так что слова из новых наборов будут покрыты без дополнительных доработок — достаточно правильно проставить им `district_id` через шаг "Синхронизация word_market ↔ district" выше.
- Открытие района как "следующего" уровня при полном освоении текущего — уже существующая механика `district_progress`/`CityMapView`, не требует изменений, кроме корректных данных по `word_market`.

### Админка (Go API) — `internal/web/admin_word_sets.go`
- `handleAdminWordSetCategories` (POST/PUT): принимать и валидировать `level_code` (если задан — один из A0..C1), сохранять через репозиторий.
- `handleAdminWordSets` (POST/PUT): то же самое для наборов; по умолчанию (если `level_code` не передан) — наследовать уровень категории (на чтение/в UI явно показывать «уровень не задан — наследуется от категории: X»).
- Добавить эндпоинт (или параметр к существующему) ручного триггера ресинхронизации легаси-контента в Linglow v2 (`POST /api/admin/word-sets/sync-districts` → `courseRepo.MapLegacyContentForLearning`), под тем же `PermissionWordSetsEdit`.
- `ProcessWordSetItemsForCourse` (`internal/service/word_set_service.go:586`) не меняется по сути — пайплайн создания `word_cards` (`EnsureWordCardExistsMinimal`) → асинхронное заполнение `TrainingWorker` → `pronunciationService.ScheduleWord` остаётся как есть и продолжит работать для новых наборов любого уровня, так как уровень не участвует в этой цепочке (он участвует только в привязке набора к district через отдельную синхронизацию).

### Админка (Vue) — `webapp/src/views/AdminWordSetsView.vue`
- В формах создания/редактирования категории и набора добавить `<select>` "Уровень (CEFR)" с опциями `A0..C1` + "не задано / наследуется от категории" (для набора).
- В списке категорий/наборов показывать бейдж уровня рядом с published-бейджем.
- Добавить кнопку "Синхронизировать с городом" (вызывает новый sync-эндпоинт), с тостом об успехе/ошибке.

## Verification

1. `go build ./...` / `go vet ./...`.
2. Прогнать миграции локально на копии БД linglow_unified (или sqlite/postgres тестовом инстансе) — проверить, что 000032 и 000033 идемпотентны при повторном запуске.
3. Создать через админку категорию с уровнем `B1` и набор без явного уровня → убедиться что эффективный уровень = B1.
4. Вызвать sync (вручную через новую кнопку/эндпоинт) → проверить в БД, что `modules`/`learning_items` для созданного набора имеют `district_id`, соответствующий B1-району курса `es_ru`.
5. Добавить слово в набор через `PUT /api/admin/word-sets/{id}/items` → убедиться, что `word_cards`/`training_cards`/`tts_generation_status` создаются как раньше (TrainingWorker + PronunciationService отрабатывают).
6. Открыть веб-апп, перейти в B1-район → убедиться, что в `word_market` видны слова из набора, и что счётчик "изучено" двигается только при `known`/`mastered`, не при простом наличии в словаре.
7. Прогнать существующие тесты в `internal/repository` и `internal/service` для word_set/course_repository.

## Открытые вопросы для уточнения перед/во время реализации
- Подтвердить точный список категорий/наборов 000002/000004, которые надо удалить (или просто удалять всё с `course_code = 'es_ru'` после проверки backfill).
