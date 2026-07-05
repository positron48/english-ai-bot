# Мгновенные основные экраны: обзор данных, кеша и план реализации

## Цель

Основные экраны Linglow должны открываться без повторяющихся экранов загрузки и без заметных прыжков интерфейса:

- `Dashboard` (`/app/dashboard`);
- `City` (`/app/city`, `/app/city/district/:districtCode`, `/app/city/daily-route`);
- `Practice` (`/app/learning`, `/app/training`, grammar/reading/speaking entry points);
- `Progress` (`/app/progress`).

Допустимая модель: первая загрузка может быть долгой, но после нее приложение работает из локального кеша. Сеть обновляет кеш в фоне, по кнопке обновления на главной и после действий пользователя, которые меняют прогресс, SRS, daily route, mastery или статистику.

Важно: это не должен быть HTTP/service-worker cache для `/api/*`. Нужен отдельный app-data cache уровня приложения, потому что данные персональные, course-scoped, зависят от пользователя, токена, текущего курса, pending offline attempts и бизнес-инвалидаторов.

## Текущая архитектура

### Frontend routes

Основной роутер находится в `webapp/src/router/index.ts`, base path — `/app`.

Основные вкладки описаны в `webapp/src/components/linglow/navTabs.ts`:

- `home` -> `/dashboard` -> `DashboardView.vue`;
- `city` -> `/city` -> `CityMapView.vue`;
- `practice` -> `/learning` -> `LearningView.vue`;
- `progress` -> `/progress` -> `ProgressView.vue`;
- `profile` -> `/settings` -> `SettingsView.vue`.

Глобального Pinia/store слоя нет. Состояние держится в module-level composables, API clients, IndexedDB modules и отдельных `localStorage` ключах.

### Текущие aggregate API

На backend уже есть слой overview, который уменьшает сетевой fan-out до одного round-trip на экран. Реализация: `internal/web/overview.go`.

- `GET /api/overview/dashboard`:
  - `dashboard` через `/api/dashboard?sections=counts`;
  - `progress` через `/api/linglow/progress?summary_only=1`;
  - `daily_route` через `/api/linglow/daily-route?limit=8`;
  - `continue_chapter`;
  - `sentence_today`.
- `GET /api/overview/city`:
  - `grammar_categories`;
  - `progress`;
  - `course_map?fields=districts`;
  - `word_levels`.
- `GET /api/overview/learning`:
  - `continue_chapter`;
  - `verb_upcoming`;
  - `vocab_summary`;
  - `sentence_today`.
- `GET /api/overview/progress`:
  - `stats`;
  - `progress`;
  - `dashboard?sections=totals`;
  - `history?days=7`.

Это хорошая основа, но сейчас каждый экран при `onMounted` снова вызывает свой overview endpoint и ставит локальный `loading`. Поэтому при возврате на уже виденный экран пользователь снова видит загрузку.

### Текущие offline/PWA/APK механики

Service Worker находится в `webapp/public/sw.js`.

Он кеширует shell и ассеты:

- `/app`, `/app/`, `/app/manifest.webmanifest`;
- Vite assets из `/app/asset-manifest.json`;
- `/app/assets/*`, `/app/fonts/*`, `/app/linglow/*`;
- `/telegram-web-app.js`, `/favicon.svg`.

Он специально не кеширует:

- `/api/*`;
- `/auth/*`;
- `/app/admin/*`.

Это правильно: shell и картинки города доступны быстро, но персональные данные экранов должны кешироваться отдельным IndexedDB/localStorage слоем.

Embedded APK определяется через `webapp/src/utils/runtime.ts` по `QantrixEmbeddedApp`. APK использует тот же web frontend, тот же service worker и те же browser storage APIs внутри WebView. Поэтому app-data cache должен быть единой frontend-механикой для web/PWA/APK.

Сейчас IndexedDB уже используется для offline grammar и word training:

- `webapp/src/api/grammarOfflineStore.ts` (`qantrix-grammar-offline`);
- `webapp/src/api/wordTrainingOfflineStore.ts` (`qantrix-word-training-offline`).

Фоновая синхронизация offline attempts:

- `webapp/src/api/offlineSyncRunner.ts`;
- запускается в `webapp/src/main.ts`;
- триггеры: `online`, `visibilitychange`, interval 30s;
- синхронизирует grammar attempts, word training attempts и content reports.

Фоновая загрузка offline pack:

- `webapp/src/composables/useOfflineAutoDownload.ts`;
- при включенной настройке `offlineAutoDownload`;
- не чаще раза в час;
- грузит grammar preload и word training preload.

Важно: это кеш контента/очередей для offline-тренировок, а не кеш overview экранов.

## Текущие экраны и их данные

### Dashboard

Файл: `webapp/src/views/DashboardView.vue`.

Текущая загрузка:

- при `isAuthenticated`;
- при смене `currentCourseCode`;
- вручную через `refreshData`;
- online: `GET /api/overview/dashboard?course_code=...`;
- offline: локальный fallback из `wordTrainingClient.getDashboard()` и `grammarClient.getStatistics()`.

Данные экрана:

- word counters из legacy `/api/dashboard`;
- canonical progress summary из `/api/linglow/progress`;
- daily route today block;
- continue grammar chapter;
- sentence today;
- streak отдельно через `useStats()` -> `/api/linglow/stats`;
- Lumi fact с отдельным localStorage cache.

Проблема:

- `loading=true` на каждый mount;
- кеш есть только для курса и части offline training/grammar, но не для whole dashboard payload;
- streak грузится отдельной механикой, не связан с overview cache.

### City map

Файл: `webapp/src/views/CityMapView.vue`.

Текущая загрузка:

- `ensureCourseLoaded()`;
- `GET /api/overview/city?course_code=...`;
- параллельно `ensureMe()` для Pro gates;
- после завершения ставится `mapReady=true`;
- картинки города идут из `/app/linglow/city/level*.jpg` и кешируются service worker.

Данные экрана:

- district titles/status;
- progress/mastery signals;
- grammar categories;
- word levels;
- Pro availability.

Проблема:

- `/city` сейчас запрещен offline router guard-ом;
- нет hydrate из последнего успешного city overview;
- canvas/labels могут появляться только после сети, хотя карта и картинки уже локально доступны.

### City district

Файл: `webapp/src/views/CityDistrictView.vue`.

Текущая загрузка:

- `courseClient.getCourseMap()`;
- `courseClient.getProgress()`;
- `useMe()` для Pro gates.

Проблема:

- не использует `/api/overview/city`;
- на каждый вход заново делает запросы;
- может быть мгновенным из того же city cache, потому что district detail в основном является проекцией course map + progress.

### Daily route

Файл: `webapp/src/views/CityDailyRouteView.vue`.

Текущая загрузка:

- `getDailyRoute(16)`;
- `getReviewQueue(16)`;
- `getProgress()`.

Проблема:

- нет собственного cached snapshot;
- daily route чувствителен к SRS writes, reading, conversation, picture quest, grammar/word attempts;
- после тренировки экран должен обновиться, но не обязан блокировать UI загрузкой.

### Practice hub

Файл: `webapp/src/views/LearningView.vue`.

Текущая загрузка:

- `ensureCourseLoaded()`;
- `ensureLearningLoaded()`;
- `GET /api/overview/learning?course_code=...`;
- fallback на отдельные calls.

Данные:

- continue grammar chapter;
- verb pool count;
- vocab summary;
- sentence availability;
- Pro gates через `/api/me`.

Проблема:

- состояние не переиспользуется между переходами;
- после training/grammar/reading/speaking нужно обновлять summary, но не обязательно показывать full-screen loading.

### Word training

Файл: `webapp/src/views/TrainingView.vue`.

Текущая модель:

- online endpoints `/api/training/*`;
- offline fallback и attempt queue уже есть в `wordTrainingClient.ts`;
- IndexedDB pack локально уменьшается после offline answer через `removeWordTrainingUserCards`;
- sync attempts идет через `/api/training/offline/sync-attempts`.

Влияние на кеш:

- после каждого ответа меняются due/new/review counts, daily route, progress, history, stats, SRS queue;
- после завершения сессии нужно инвалидировать dashboard, practice, city, progress, daily route/review;
- для мгновенного UX лучше делать optimistic patch локальных счетчиков сразу, а сеть/overview refresh запускать в фоне.

### Grammar

Файлы:

- `GrammarCategoriesView.vue`;
- `GrammarChaptersView.vue`;
- `GrammarChapterView.vue`;
- `GrammarTestView.vue`;
- `GrammarTrainingView.vue`;
- `grammarClient.ts`;
- `grammarOfflineStore.ts`.

Текущая модель:

- categories имеют короткий in-memory TTL 30 секунд;
- offline preload хранит manifest, chapters, training questions;
- offline test/training attempts queued и sync-ятся позже.

Влияние на кеш:

- submit chapter/category/placement test меняет grammar stats, continue chapter, city mastery, progress/history;
- grammar training answer меняет grammar SRS queue, daily route/review, progress/history;
- offline submit должен локально patch-ить доступ/score там, где это возможно, и пометить overview cache dirty до sync.

### Reading

Файл: `webapp/src/views/ReadingTextView.vue`.

Текущая загрузка:

- `GET /api/learning/reading/texts/{id}`;
- `ReadingPassageBlock` отправляет mark-read;
- view локально ставит `readingIsRead=true` после события `marked-read`.

Влияние на кеш:

- mark-read меняет progress, stats month texts, city district/location signals, daily route today reading_done;
- после mark-read нужен cache invalidation для dashboard, city, progress, daily route.

### Speaking / Conversation / Picture Quest

Файлы:

- `PlaceChatView.vue`;
- `PictureQuestChatView.vue`;
- `courseClient.ts`;
- backend handlers `internal/web/linglow_conversation.go`, `internal/web/linglow_picture_quest.go`, `internal/web/speaking.go`.

Влияние на кеш:

- сообщения и quest completion пишут `learning_events`/daily stats;
- completion влияет на city/progress/daily route;
- списки и текущие сессии стоит кешировать отдельно от overview, но overview должен инвалидироваться после quest progress.

### Progress

Файл: `webapp/src/views/ProgressView.vue`.

Текущая загрузка:

- `GET /api/overview/progress?course_code=...`;
- передает `dashboard` и `history` в `LgProgressCharts`, чтобы тот не делал повторный fetch.

Данные:

- `stats` из `daily_course_stats` + `learning_events`;
- `progress` из canonical course progress/mastery;
- `dashboard totals` из legacy dashboard;
- `history` из `exercise_attempts`/SRS history.

Проблема:

- `/progress` запрещен offline router guard-ом;
- повторный mount снова показывает loading, хотя последние данные подходят для мгновенного отображения.

## Backend read/write источники

### Read split

Dashboard counters по словам остаются legacy-first:

- `user_cards`;
- `user_word_knowledge`;
- `review_events`;
- `/api/dashboard`.

City/progress/mastery/history в основном canonical-first:

- `exercise_attempts`;
- `learning_events`;
- `srs_items`;
- `daily_course_stats`;
- `/api/linglow/progress`;
- `/api/linglow/history`;
- `/api/linglow/stats`.

Daily route/review зависит от `LINGLOW_SRS_READ_ENABLED`:

- `false`: legacy queue (`user_cards`, `grammar_theory_memory`);
- `true`: canonical `srs_items`.

По текущим operational notes для prod SRS read/write включены.

### Write actions, влияющие на основные экраны

Ниже перечислены действия, после которых app-data cache должен быть обновлен или помечен dirty.

| Действие | Endpoint | Что меняется | Какие кеши затронуты |
|---|---|---|---|
| Ответ в word training | `POST /api/training/answer` | legacy SRS, `review_events`, canonical attempts/events, SRS mirror | dashboard, learning, city, progress, daily-route, review, word offline pack |
| Offline word sync | `POST /api/training/offline/sync-attempts` | то же batch-ом | dashboard, learning, city, progress, daily-route, review |
| Grammar chapter/category/placement test | grammar submit endpoints | grammar progress, attempts/events | dashboard, learning, city, progress, continue chapter |
| Grammar training answer | grammar training answer/sync | grammar SRS, attempts/events | dashboard, city, progress, daily-route, review |
| Reading mark-read | reading `mark-read` | reading progress, events, stats | dashboard, city, progress, daily-route |
| Speaking submit | speaking submit/next | speaking attempts, daily stats | progress, city, daily-route |
| Conversation/Picture quest message | conversation/picture message endpoints | events, quest task progress, daily stats | city, progress, daily-route |
| Word set learn/know | word sets study endpoints | `user_cards`, known words | dashboard, learning, city, progress, daily-route, review |
| Course select | `POST /api/user/courses/select` | active course/user_course | all screen caches for current course namespace |
| Activity heartbeat | `POST /api/linglow/activity` | active minutes/streak/day stats | dashboard streak, progress stats |
| Offline sync success | sync runner | server catches up with queued attempts | all dirty caches for active course |

## Рекомендуемая целевая модель

### 1. Единый app-data cache

Добавить frontend слой, например:

- `webapp/src/api/appDataCache.ts` — IndexedDB storage;
- `webapp/src/composables/useScreenDataCache.ts` или набор `useDashboardData`, `useCityData`, `useLearningData`, `useProgressData`;
- `webapp/src/api/cacheInvalidation.ts` — события инвалидирования и tags.

Хранить данные в IndexedDB, не в `localStorage`, потому что payload city/progress может быть большим и должен быть версионируемым.

Предлагаемая DB:

- name: `linglow-app-data-cache`;
- version: `1`;
- stores:
  - `screens` с key `[userID|anonTokenScope, courseCode, screenKey]`;
  - `meta` для schema version, app build version, last refresh;
  - `mutations` или `dirtyTags` для pending offline/local invalidations.

Минимальная запись:

```ts
interface CachedScreenPayload<T = unknown> {
  key: 'dashboard' | 'city' | 'learning' | 'progress' | 'daily-route' | 'review' | 'district'
  userScope: string
  courseCode: string
  locale: string
  appVersion: string
  dataVersion: number
  payload: T
  fetchedAt: string
  staleAt: string
  dirtyTags: string[]
  pendingLocalMutations: number
}
```

`userScope` не должен хранить сам JWT. Достаточно стабильного user id, если он доступен из `/api/me`, или hash от access token/session scope. При logout кеш текущего пользователя нужно очищать или делать недоступным.

### 2. Hydrate-first стратегия

Для каждого основного экрана:

1. Синхронно/быстро открыть IndexedDB и показать последний payload.
2. Если payload есть:
   - не показывать full-screen loader;
   - показать маленький статус `Обновлено ...` или unobtrusive refresh indicator;
   - сохранить высоту блоков, чтобы UI не прыгал.
3. Если payload отсутствует:
   - показать текущий loader/skeleton;
   - загрузить overview endpoint.
4. В фоне выполнить refresh, если:
   - запись stale;
   - есть dirty tags;
   - был explicit pull/refresh;
   - приложение вернулось online/visible;
   - завершился offline sync.
5. После успешного refresh атомарно заменить cache и reactive state.

Важно: stale данные не равны invalid. Stale можно показывать. Invalid/dirty тоже можно показывать, но UI должен знать, что идет фоновое обновление.

### 3. Screen cache keys

Рекомендуемые ключи:

- `dashboard:${courseCode}`;
- `city:${courseCode}`;
- `city-district:${courseCode}:${districtCode}` или derived view из `city`;
- `daily-route:${courseCode}`;
- `review:${courseCode}`;
- `learning:${courseCode}`;
- `progress:${courseCode}`;
- `me/features` отдельно с TTL как сейчас в `useMe`;
- `courses` можно оставить в `useCourse`, но лучше позднее перенести в общий cache namespace.

Для `city-district` лучше не кешировать отдельный payload на первом этапе: строить экран из `city` cache + `progress` cache. Если нужны district extras, кешировать extras отдельно.

### 4. TTL и freshness

Предлагаемые TTL:

- `dashboard`: 5 минут, но dirty после любых attempts;
- `learning`: 5 минут, dirty после attempts, word set changes, sentence completion;
- `city`: 15 минут, dirty после attempts/reading/conversation/picture/word sets;
- `daily-route` и `review`: 1-2 минуты, dirty после SRS-related actions;
- `progress`: 10 минут, dirty после any progress action;
- `me/features`: 15 минут или существующий TTL;
- `courses`: 24 часа, dirty после course select.

TTL нужен только для автоматического background refresh. Он не должен блокировать показ cached UI.

### 5. Инвалидирование через доменные события

Вместо ручного `loadData()` в каждом view нужен небольшой event bus:

```ts
type AppDataEvent =
  | 'word-review-recorded'
  | 'word-review-session-completed'
  | 'grammar-test-submitted'
  | 'grammar-training-recorded'
  | 'reading-marked-read'
  | 'speaking-attempt-submitted'
  | 'conversation-progressed'
  | 'picture-quest-progressed'
  | 'word-set-updated'
  | 'activity-recorded'
  | 'offline-sync-completed'
  | 'course-selected'
  | 'manual-refresh'
```

Каждое событие мапится на tags:

- `srs`;
- `words`;
- `grammar`;
- `reading`;
- `speaking`;
- `conversation`;
- `picture`;
- `stats`;
- `progress`;
- `city`;
- `daily-route`;
- `course`.

Пример маппинга:

- `word-review-recorded` -> `words`, `srs`, `stats`, `progress`, `daily-route`;
- `grammar-test-submitted` -> `grammar`, `stats`, `progress`, `city`;
- `reading-marked-read` -> `reading`, `stats`, `progress`, `city`, `daily-route`;
- `course-selected` -> все cache keys другого current course перестают быть active, текущий course грузится hydrate-first.

### 6. Optimistic patch для мгновенности

После user action не обязательно ждать full overview refresh. Нужно локально patch-ить очевидные изменения:

- word answer:
  - уменьшить `dashboard.due_count`/`available_for_training` минимум на 1 для текущей карточки;
  - обновить training session state;
  - пометить progress/daily-route dirty.
- reading mark-read:
  - поставить `daily_route.today.reading_done=true`;
  - увеличить локальный `stats.month.texts_read`, если текст не был read;
  - пометить city/progress dirty.
- grammar test passed:
  - обновить локальный continue chapter, если backend result содержит next action;
  - пометить grammar/progress/city dirty.
- conversation/picture quest completion:
  - обновить current session UI сразу;
  - пометить city/progress/daily-route dirty.

Optimistic patch должен быть консервативным. Если точная бизнес-логика сложная, лучше только убрать loader и показать старые данные с фоновым refresh, чем изобретать неверные метрики.

### 7. Manual refresh на главной

Кнопка обновления на Dashboard должна стать глобальной:

- обновляет `dashboard`, `learning`, `city`, `progress`, `daily-route`, `review` для текущего курса;
- запускает `scheduleOfflineSync()` перед refresh, если online;
- показывает компактный статус, не full-screen loader, если данные уже есть;
- при ошибке оставляет старый кеш и показывает non-blocking warning.

Можно добавить `refresh all` в `useAppDataRefresh()`:

```ts
await refreshAppData({
  courseCode,
  reason: 'manual-refresh',
  screens: ['dashboard', 'learning', 'city', 'progress', 'daily-route', 'review'],
})
```

### 8. Router offline guard

После появления app-data cache можно разрешить offline просмотр:

- `/city`;
- `/city/district/:districtCode`;
- `/progress`.

Условие: cached payload для active course существует. Если нет payload, оставить текущий redirect на `/dashboard` или показать экран "нужно один раз открыть онлайн".

Для APK это особенно важно: после первого online запуска основные экраны должны быть доступны в WebView без сети.

### 9. Борьба с прыжками интерфейса

Помимо cache, нужно поменять loading UX:

- не сбрасывать screen state в `null` перед refresh;
- не ставить `loading=true`, если есть cached payload;
- заменить full-screen loaders на `refreshing` индикаторы;
- держать stable placeholders высоты для карточек, которые догружаются в фоне;
- для training idle screen не скрывать stats cards, пока refresh идет;
- для city canvas рисовать карту сразу, а labels/status брать из кеша.

## План реализации

### Этап 1. Базовый IndexedDB screen cache

1. Добавить `appDataCache.ts`:
   - typed get/set/delete;
   - schema version;
   - user/course/screen key;
   - stale/dirty metadata.
2. Добавить `useCachedOverviewScreen`:
   - принимает `screenKey`, `courseCode`, `fetcher`;
   - возвращает `data`, `hasCache`, `loadingInitial`, `refreshing`, `error`, `refresh()`.
3. Подключить только `DashboardView.vue`:
   - hydrate из cache;
   - background refresh `/api/overview/dashboard`;
   - ручная refresh-кнопка пишет cache.
4. Добавить unit tests для cache key/version/stale logic.

Критерий готовности: повторный вход в dashboard после первой загрузки не показывает full-screen loader.

### Этап 2. City и Progress hydrate-first

1. Подключить `CityMapView.vue` к `city` cache.
2. `CityDistrictView.vue` строить из cached city/progress, сеть только refresh в фоне.
3. Подключить `ProgressView.vue` к `progress` cache.
4. Разрешить offline `/city` и `/progress`, если cache есть.
5. Сохранять service-worker cached city images как есть.

Критерий готовности: после первого посещения city/progress открываются мгновенно в web и APK, даже при плохой сети.

### Этап 3. Practice hub и daily-route/review

1. Подключить `LearningView.vue` к `learning` cache.
2. Добавить отдельные cache keys для `daily-route` и `review`.
3. Использовать cached daily route на Dashboard и City Daily Route.
4. Убрать повторные загрузки при перелистывании между `dashboard` -> `learning` -> `progress`.

Критерий готовности: главные вкладки переключаются без loader при повторном посещении.

### Этап 4. Инвалидирование после user actions

1. Добавить `cacheInvalidation.ts` с event bus и tag mapping.
2. Вызвать события после успешных действий:
   - `wordTrainingClient.answer`;
   - offline word answer queue;
   - grammar test submit;
   - grammar training answer;
   - reading mark-read;
   - speaking submit;
   - conversation/picture message или quest completion;
   - word set learn/know;
   - course select;
   - offline sync success.
3. На событиях:
   - mark dirty соответствующие cache entries;
   - если экран активен и online, запустить debounced background refresh;
   - если offline, оставить dirty до sync/online.

Критерий готовности: после тренировки/грамматики/чтения пользователь видит локальное изменение сразу, а основные экраны обновляются без ручного reload.

### Этап 5. Optimistic patches

1. Добавить ограниченные patch helpers:
   - `patchDashboardAfterWordAttempt`;
   - `patchDailyRouteAfterReadingDone`;
   - `patchProgressStatsAfterReadingDone`;
   - `patchLearningAfterGrammarSubmit`.
2. Патчить только очевидные поля.
3. Все сложные поля (`mastery`, unlock, SRS due distribution) обновлять фоновым overview refresh.

Критерий готовности: после очевидных действий цифры не выглядят устаревшими до фонового refresh.

### Этап 6. Prefetch и warmup

1. После успешного login/course load загрузить в фоне:
   - dashboard;
   - learning;
   - city;
   - progress;
   - daily-route/review.
2. После открытия Dashboard впервые запустить `refresh all`.
3. В APK/PWA сохранить тот же путь; не добавлять Android-specific storage.
4. Не грузить тяжелые detail payloads без необходимости.

Критерий готовности: первый заход на dashboard может быть обычным, но последующее открытие остальных основных вкладок уже почти мгновенное.

### Этап 7. Наблюдаемость и отладка

Добавить debug panel или hidden console helpers:

- cache keys;
- fetchedAt/staleAt;
- dirty tags;
- pending offline attempts;
- last refresh error;
- current course/user scope.

Это поможет отлаживать APK/WebView, где storage и offline behavior труднее инспектировать.

## Риски и ограничения

- Нельзя blindly кешировать `/api/*` в service worker: данные персональные и зависят от Authorization/current course.
- Нужно разделять кеши по user scope и course code, иначе возможна утечка/смешение English/Spanish или разных пользователей.
- При logout нужно очистить или изолировать app-data cache.
- Pending offline attempts означают, что server state отстает. UI должен показывать cached+dirty состояние и синхронизировать при online.
- Dashboard word counts legacy-first, progress/city canonical-first. После action возможна краткая рассинхронизация, поэтому optimistic patches должны быть осторожными.
- Course switch должен мгновенно переключать namespace кеша, а не перетирать текущие данные другого курса.

## Рекомендуемая последовательность PR

1. `appDataCache` + Dashboard hydrate-first.
2. City/Progress hydrate-first + offline guard для cached routes.
3. Learning/DailyRoute/Review cache + Dashboard global refresh.
4. Cache invalidation event bus + wiring в основных write flows.
5. Optimistic patches для word training и reading.
6. APK/PWA polish: debug state, no-loader transitions, docs/tests.

## Краткий итог

У приложения уже есть правильная база: aggregate endpoints, service-worker shell cache, IndexedDB для offline grammar/word training и offline sync runner. Не хватает центрального app-data cache для персональных overview payloads и единой доменной инвалидизации после действий пользователя.

Самый быстрый путь к ощутимой мгновенности: сначала внедрить hydrate-first cache для `dashboard`, `city`, `learning`, `progress`, потом подключить invalidation events. Это даст мгновенное перелистывание основных вкладок без больших backend-изменений и будет одинаково работать в web, PWA и APK.
