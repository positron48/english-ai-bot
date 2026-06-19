# Linglow — план функционализации захардкоженных блоков вёрстки

Дата: 2026-06-12
Статус: план, к реализации
Связанные документы: `docs/LINGLOW_UI_MIGRATION_PLAN.md` (миграция вёрстки), `docs/LINGLOW_TARGET_ARCHITECTURE_MIGRATION_PLAN_RU.md`

Цель: после выполнения всех этапов каждый блок новой вёрстки показывает реальные данные, наполняется через БД/админку и не содержит фейковых вычислений. Этапы изолированы — каждый деплоится отдельно, порядок внутри «волн» можно менять.

---

## 0. Инвентаризация: что захардкожено сейчас

### 0.1 Полностью фейковые / статичные блоки

| # | Блок | Где | Что не так |
|---|------|-----|------------|
| 1 | **Lumi-факты** («Lumi знает») | `LgLumiFact.vue:24-29`, массив в `locales/ru.json:848` | 10 фактов про испанский зашиты в i18n; ротация по дню через `Date.now()`. Не зависят от курса (на английском курсе — факты про испанский), не зависят от экрана, нет управления контентом |
| 2 | **Минуты обучения** | `ProgressView.vue:211` | `minutes = learnedWords * 3.2` — выдумка. Время нигде не трекается |
| 3 | **Ритм недели** (активность по дням) | `ProgressView.vue:208` | `weekAct = ['done','done','done','done','done','today','empty']` — константа |
| 4 | **Streak** («N дней подряд») | `ProgressView.vue:187` (`streakDays = ref(6)`), `LgStreakBadge`/`LgPageHeader` никем не наполняются | Streak нигде не считается, ни на бэке, ни на фронте |
| 5 | **Сводка месяца** (минуты/слова/тексты/диалоги) | `ProgressView.vue:210-215` | Все 4 метрики — производные от `learnedWords` (×3.2, ×0.06, ×0.05). Вдобавок `/api/dashboard` не возвращает поля `total_words`/`learned_words`, которые читает view (`ProgressView.vue:200-201`) → всё всегда **0** |
| 6 | **Районы на экране Прогресс** | `ProgressView.vue:217-223` | Hardcode «Plaza Clara 88%, Distrito Alto 62%…» — хотя реальные данные есть в `/api/linglow/progress.by_district` |
| 7 | **Сильные стороны / навыки** | `ProgressView.vue:225-230` | Проценты — производные от `learnedPct` (±3, ×0.73, ×0.80) |
| 8 | **«Что подтянуть»** | `ProgressView.vue:232-236` | Три статичные i18n-строки, не зависят от реальных слабостей |
| 9 | **Ачивки** | `ProgressView.vue:238-244` | Полу-фейк: «исследователь = 4» зашито, «эксперт» всегда locked, тексты/слова — те же производные |
| 10 | **Любимая зона** | `ProgressView.vue:80-92` + i18n `progress.favZoneName` | Название зоны зашито в локали, донат — от `learnedPct` |
| 11 | **Здания района** (лейблы на иллюстрации) | `CityDistrictView.vue:191-196` | «Jardín de Frases» и др. + координаты в % зашиты в коде, одинаковы для всех районов |
| 12 | **Прогресс «Общение» и «Чтение» в районе** | `CityDistrictView.vue:142-147` | `chatPct = wordsPct*0.5`, `readingPct` fallback `= wordsPct*0.7` |
| 13 | **Описание района** | `CityDistrictView.vue:153-162` | Подбор i18n-строки эвристикой по подстроке кода района (`includes('plaza')`...), хотя в таблице `districts` есть пустующее поле `description` |
| 14 | **«Новые открытия»** в районе | `CityDistrictView.vue:199` | Заголовок — статичная фраза `city.discoveryPhrase`, ведёт просто в список чтения. Нет реальной рекомендации |
| 15 | **«Задачи дня» района** | нет вообще | В прототипе (`linglow-data.js` → `districtDetail.*.tasks`) есть блок задач с прогрессом «2/3» — в Vue не портирован, т.к. нет данных |
| 16 | **Профиль: карточка пользователя** | `SettingsView.vue:13-26`, `locales/ru.json:277-278` | Имя — «Linglow», чипы «🌿 Уровень A1» и «📚 Испанский» зашиты в локаль |
| 17 | **«Твой путь сегодня»** на главной | `DashboardView.vue:38-79` | 4 статичные ссылки; динамика только в количестве слов. Нет реального состава шагов и отметок выполнения |
| 18 | **Карта города** | `CityView.vue:3-10` | Статичная картинка `map_city.jpg` вместо интерактивной canvas-карты прототипа (полигоны районов, lock-уровни). Картинки районов в списке ротируются декоративно (`CityView.vue:129-130`), не привязаны к районам |
| 19 | **Lumi-tip района** | `CityDistrictView.vue:171-176` | Три пороговые i18n-фразы по confidence. Приемлемо как v1, но стоит влить в систему фактов/советов |

### 0.2 Что уже есть в БД/бэке и можно переиспользовать

- **`learning_events` / `exercise_attempts`** (миграция `000017`) — события пишутся для grammar test/training и word review (`internal/repository/linglow_event_repository.go`); есть backfill-команды `cmd/backfill_linglow_*`.
- **`daily_course_stats`** и **`mode_daily_stats`** — таблицы с `attempt_count`, `correct_count`, **`active_seconds`** уже в схеме, но **ничем не наполняются** (используются только в тестовых фикстурах). Это готовый фундамент для минут/ритма/streak.
- **`district_progress`**, `learning_item_stats`, `content_performance_stats` — тоже в схеме, не наполняются.
- **`districts.description`**, **`districts.metadata_json`**, `locations.metadata_json` — пустые поля под контент района.
- **`/api/linglow/progress`** — by_district / by_location с foundation/confidence/stability/weakness — реальные данные для экранов Прогресс и Район.
- **`/api/linglow/history`** — weekly_stats, words_added_stats, accuracy, **by_mode** (attempt/correct per mode).
- **`reading_text_progress`** — факт прочтения текста (user_id, chapter_id, read_at) — источник для «текстов прочитано».
- **Чат не персистится** (`/api/chat` — stateless): для метрики «диалогов» нужно начать писать событие.
- **`app_settings`** + `AdminAppSettingsView` — образец key-value настроек и админ-CRUD.
- Админка — отдельный entry (`admin.html`), новые админ-страницы добавляются по образцу существующих `Admin*View`.

---

## Архитектурные решения (общие для всех этапов)

1. **Один источник правды по активности — `learning_events` + дневные агрегаты.** Всё, что показывает «активность» (минуты, streak, ритм, метрики месяца), читается из `daily_course_stats`/`mode_daily_stats`, которые наполняются (а) синхронно при записи событий, (б) heartbeat'ом времени, (в) backfill-командой по историческим событиям.
2. **Локальная дата пользователя приходит с клиента.** В агрегатах ключ — `local_date`; клиент передаёт свою дату (`YYYY-MM-DD`) в heartbeat и при записи попыток (поле `client_day`). Сервер валидирует: `client_day ∈ [server_date-1, server_date+1]`, иначе берёт серверную. Так streak и «сегодня» не ломаются на часовых поясах.
3. **Чего нет в данных — блок не рендерим** (принцип v1 из UI-плана). На каждом этапе фейковая ветка либо заменяется реальной, либо блок прячется.
4. **Контент (факты, описания районов, ачивки) — per-course и per-locale.** Курс определяет язык изучения (факты про испанский ≠ факты про английский), locale — язык интерфейса.
5. **Каждый этап = отдельный PR** с миграцией (если нужна), бэком, фронтом и тестами. Сначала бэк (deployable, ничего не ломает), затем фронт.

---

## Волна 1 — фундамент данных

### Этап 1. События для всех активностей + дневные агрегаты

**Зачем:** сейчас события пишутся только для grammar и word review. Чтение, чат, speaking — невидимы. Без этого нет ни минут, ни streak, ни метрик месяца.

**БД** (миграция `000026_linglow_activity_aggregates.sql`):
- Новые `event_type` в `learning_events`: `reading_text_completed`, `chat_message_sent`, `speaking_turn_completed` (схема CHECK-ов на event_type нет — менять таблицу не нужно, только договориться о константах в Go).
- Индекс при необходимости уже есть (`idx_learning_events_course_time`).

**Бэкенд:**
- `LinglowEventRepository`: методы `RecordReadingCompleted`, `RecordChatMessage`, `RecordSpeakingTurn` — по образцу существующих `Record*` (резолв user_course по course_code, идемпотентность по `source_table`/`source_pk`).
- Врезки в существующие хендлеры:
  - `reading_text_progress` upsert (хендлер чтения) → событие `reading_text_completed` (source_pk = `user_id:chapter_id`, идемпотентно).
  - `/api/chat` (`handleChat`) → событие `chat_message_sent` на каждое user-сообщение (только course-scoped; в payload — длина сообщения).
  - speaking-сессии → `speaking_turn_completed`.
- **Апдейтер агрегатов** `internal/service/daily_stats_service.go`:
  - `BumpDaily(userCourseID, localDate, mode, isCorrect, isAttempt)` — upsert в `daily_course_stats` (+`mode_daily_stats`): `attempt_count`, `correct_count`, `review_count`/`new_count` по state SRS.
  - Вызывается из всех `Record*` методов событий и из `handleLinglowExerciseAttempts`.
- **Backfill-команда** `cmd/backfill_linglow_daily_stats`: проходит `learning_events`+`exercise_attempts`+`reading_text_progress`, наполняет `daily_course_stats`/`mode_daily_stats` за всю историю (local_date = серверная дата события; для истории это приемлемо). По образцу `cmd/backfill_linglow_events`.

**Фронт:** изменений нет.

**Готово, когда:** после тренировки/чтения/чата строка за сегодня в `daily_course_stats` растёт; backfill прогнан на проде; счётчики сходятся с `/api/linglow/history`.

---

### Этап 2. Трекинг времени (минуты обучения)

**Зачем:** «минуты» нельзя честно вывести из числа слов. Считаем реальное активное время.

**Модель:** клиент шлёт heartbeat «я активен», сервер аккумулирует `active_seconds` в `daily_course_stats` и `mode_daily_stats`.

**Фронтенд** — composable `useActivityTracker.ts` (подключается один раз в `PublicLayout.vue`):
- Тикает раз в секунду, считает секунду «активной», если: вкладка видима (`document.visibilityState === 'visible'`) **и** был ввод (pointer/key/scroll/touch) за последние 60 с.
- Раз в 60 с (и на `visibilitychange`→hidden / `pagehide` через `navigator.sendBeacon`) отправляет батч:
  `POST /api/linglow/activity` `{ course_code, client_day, seconds, mode }`
  - `mode` — грубая классификация по текущему роуту: `words | grammar | reading | chat | speaking | other` (маппинг от `route.meta.navTab`/пути; держать в `utils/activityMode.ts`).
- Оффлайн: батчи складываются в localStorage-очередь (по образцу `wordTrainingOfflineStore`), синк при появлении сети через `offlineSyncRunner`.

**Бэкенд:**
- `POST /api/linglow/activity` (auth): валидация `seconds ∈ (0, 120]` на один пинг, `client_day` по правилу из «Архитектурных решений»; дневной потолок (например, 16 ч) от мусора. Upsert: `active_seconds += seconds`.
- Swagger + тест.

**Исторические минуты:** в backfill из этапа 1 добавить эвристику — «минуты ≈ количество уникальных минут, в которые было ≥1 событие, ×1 мин» (нижняя оценка). Пометить в `stats_json: {"estimated": true}`.

**Готово, когда:** после 5 минут активного использования `active_seconds` за сегодня ≈ 300; фоновая вкладка время не накручивает; оффлайн-тренировка досылает время после синка.

---

### Этап 3. Stats API: `/api/linglow/stats`

**Зачем:** единая ручка для экрана Прогресс, streak-баджа и шапок. Читает только агрегаты (дёшево).

**Бэкенд** — `internal/web/linglow_stats.go`, `GET /api/linglow/stats?course_code=&month=YYYY-MM`:

```json
{
  "course": { ... },
  "streak": { "current_days": 6, "best_days": 14, "today_active": true },
  "today":  { "active_seconds": 1260, "attempt_count": 42 },
  "week":   [ { "date": "2026-06-06", "active_seconds": 900, "attempt_count": 30, "status": "done|today|empty" }, ... 7 дней ],
  "month":  {
    "active_minutes": 320,
    "words_learned": 58,        // srs_items: перешедшие в review/mastered за месяц (по updated_at+state) или words_added_stats
    "texts_read": 4,            // reading_text_progress за месяц
    "chat_messages": 31,        // learning_events type=chat_message_sent за месяц
    "active_days": 18
  },
  "skills": [ { "mode": "grammar", "attempt_count": 120, "correct_count": 98, "accuracy_percent": 81.6 }, ... ],
  "favorite_district": { "district_code": "plaza", "title": "Plaza Clara", "attempt_count": 88, "progress_percent": 64 },
  "generated_at": "..."
}
```

- **Streak:** идём по `daily_course_stats` от сегодня (локальная дата = максимум из присланных client_day за сегодня, иначе серверная) назад, пока `attempt_count>0 OR active_seconds>=60`. Если сегодня пусто — streak считается до вчера (не обнуляем в течение дня). `best_days` — одним проходом по всем дням (их немного).
- **skills** — из `mode_daily_stats` за месяц (или за всё время, если месяц пустой — флаг `period`).
- **favorite_district** — district с max `attempt_count` за месяц: `exercise_attempts → learning_items → modules → locations → districts` (либо, дешевле, добавить `district_id` в `mode_daily_stats.stats_json` на этапе 1; v1 — допустим прямой запрос с лимитом по месяцу).
- Кэш в памяти на 60 с per (user, course) — ручку будут дёргать все экраны.

**Фронт:** `api/statsClient.ts` + типы. Пока не подключаем (следующие этапы).

**Готово, когда:** swagger-тесты на стрик (граничные случаи: пустой сегодня, разрыв, первый день), ручка отвечает <50 мс на тёплом кэше.

---

## Волна 2 — экраны на реальные данные

### Этап 4. ProgressView без фейков

Полная замена данных в `ProgressView.vue` (вёрстка не меняется):

| Блок | Источник |
|------|----------|
| Сводка месяца (мин/слова/тексты/диалоги) | `stats.month` (этап 3). Подписи «в этом месяце» — оставить; если `estimated` минуты — без изменений UI |
| Ритм недели | `stats.week` → `weekAct` из `status` |
| Streak «N дней» | `stats.streak.current_days` |
| Районы | `/api/linglow/progress.by_district` (уже грузится на других экранах): name=title, pct=`progress_percent`, статус по порогам (≥75 «отлично», ≥40 «хорошо», ≥10 «в процессе», >0 «началось»), locked = `attempted_items===0 && district.status==='locked'` |
| Любимая зона | `stats.favorite_district` (имя + донат `progress_percent`). Нет данных → блок скрыть |
| Сильная сторона | максимум `accuracy_percent` из `stats.skills` (с min attempt_count ≥ 20, иначе скрыть) |
| Сильные стороны (список) | `stats.skills` → проценты = accuracy, лейблы mode → i18n |
| Что подтянуть | **временно скрыть** (реализуется в этапе 10) |
| Ачивки | **временно скрыть** (этап 9) |

Заодно убрать чтение несуществующих `total_words`/`learned_words` из `/api/dashboard` (баг — сейчас всегда 0).

**Готово, когда:** на проде Прогресс показывает живые числа; при пустом аккаунте — аккуратные нули/скрытые блоки, не NaN.

---

### Этап 5. Streak в шапках и навигации

- `LgTopBar` / `LgPageHeader` (`streak`-prop уже есть) — наполнить из `stats.streak.current_days`.
- Composable `useStats.ts`: грузит `/api/linglow/stats` один раз за сессию (+рефреш после завершения тренировки — событие через mitt/emit или просто re-fetch на router afterEach в тренировочные экраны), отдаёт `streak`, `today`.
- Показать бадж: Главная (шапка), Прогресс, профиль-чип (этап 6). На остальных экранах — по месту, где он есть в прототипе (`linglow-screens-home.jsx:183,380`).
- streak=0 → бадж не показываем (как в прототипе для новичка).

**Готово, когда:** после первой тренировки дня бадж появляется/инкрементится без перезагрузки.

---

### Этап 6. Профиль: реальные имя / уровень / курс

- **Бэкенд:** `GET /api/me` (или расширить существующий auth-ответ): `{ first_name, username, telegram_id, created_at }` из `users`.
- **Фронт** `SettingsView.vue`:
  - Имя: `first_name || username || 'Estudiante'`.
  - Чип уровня: `confirmed_level` из grammar statistics (уже есть в `/api/dashboard.grammar_stats`); нет уровня → чип «Новичок».
  - Чип курса: `currentCourse.title` из `useCourse` (вместо i18n «Испанский»).
  - Третий чип — streak (`🔥 N дней`) из `useStats`.
- Удалить ключи `settings.profileChipLevel/profileChipCourse` из локалей.

**Готово, когда:** профиль у пользователя с английским курсом показывает «English», не «Испанский».

---

## Волна 3 — контент-системы

### Этап 7. Lumi-факты: контент в БД + админка + контексты

**Модель данных** (миграция `000027_lumi_facts.sql`):

```sql
CREATE TABLE IF NOT EXISTS lumi_facts (
    id BIGSERIAL PRIMARY KEY,
    course_code TEXT NOT NULL DEFAULT '',      -- '' = для всех курсов
    context TEXT NOT NULL DEFAULT 'general',   -- general|grammar|reading|practice|progress|city
    locale TEXT NOT NULL DEFAULT 'ru',
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',     -- active|archived
    last_shown_on DATE,                        -- когда факт в последний раз был «фактом дня»
    shown_count INTEGER NOT NULL DEFAULT 0,
    created_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (context IN ('general','grammar','reading','practice','progress','city')),
    CHECK (status IN ('active','archived'))
);
CREATE INDEX idx_lumi_facts_pick ON lumi_facts(course_code, context, locale, status, last_shown_on);
```

**Ротация «факт дня»** — `GET /api/linglow/lumi-fact?context=grammar&course_code=es_ru`:
1. В транзакции: ищем факт с `last_shown_on = today` для (course, context, locale) — если есть, возвращаем его (все пользователи в один день видят один факт — это дёшево и предсказуемо).
2. Нет → выбираем `ORDER BY last_shown_on ASC NULLS FIRST, id` `LIMIT 1` `FOR UPDATE SKIP LOCKED`, ставим `last_shown_on = today`, `shown_count++`. Так факты идут по кругу, «самый давно не показанный — следующий», новые добавленные вклиниваются первыми.
3. Фоллбеки: нет фактов для контекста → берём `context='general'`; нет для курса → `course_code=''`; совсем пусто → 204, фронт прячет карточку.
4. Ответ кэшируется в памяти до конца дня per (course, context, locale).

**Наполнение «на много дней вперёд»:** просто массово добавляем факты в админке; очередь ротации сама растягивает их по дням. Отдельной «календарной» раскладки не нужно.

**Админка** — `AdminLumiFactsView.vue` (+ роут в admin-router, пункт в `AdminMenu`):
- Фильтры: курс, контекст, локаль, статус.
- Таблица: текст (обрезанный), context, last_shown_on, shown_count; действия: редактировать (textarea), архивировать.
- **Массовое добавление:** одна большая textarea «один факт = один абзац (разделение — пустая строка)» + селекты course/context/locale → `POST /api/admin/lumi-facts/bulk`. Это покрывает сценарий «нагенерировал в LLM 30 фактов → вставил → готово на месяц».
- (Опционально, позже) кнопка «Сгенерировать через LLM» по образцу `AdminPromptTesterView` — генерит в textarea, админ правит и сохраняет.
- API: `GET/POST/PUT /api/admin/lumi-facts`, `POST /api/admin/lumi-facts/bulk` (admin-permission по образцу других admin-ручек).

**Сидинг:** текущие 10 фактов из `locales/ru.json` перенести миграцией/seed-скриптом в `lumi_facts` (course=es_ru? — нет: `course_code=''` не годится, факты про испанский → `es_ru`; для `en_ru` добавить свои). Ключ `lg.lumiFacts` из локалей удалить.

**Фронт:**
- `LgLumiFact.vue`: prop `context` (default `general`); грузит факт через `factClient.getDailyFact(context)`; кэш в localStorage `lumi-fact:{course}:{context}:{date}` — повторные заходы и оффлайн в течение дня без сети. Нет факта → компонент не рендерится (родителю — `v-if` через emit или просто пустой рендер).
- Точки подключения: Dashboard (`context='general'`), LearningView (`practice`), ProgressView (`progress`), списки Грамматики (`grammar`) и Чтения (`reading`) — в прототипе факт-карточка есть на обоих (`linglow-screens-grammar.jsx:277`, `linglow-screens-reading.jsx:277`).
- `LgLumiTip` (совет в районе) пока остаётся на пороговых i18n-фразах — это «совет», не «факт»; при желании позже добавить context `city`.

**Готово, когда:** факт меняется в полночь; добавленный в админке факт появляется в ротации; на курсе en_ru нет испанских фактов; оффлайн показывается вчерашний/сегодняшний кэш.

---

### Этап 8. Район: контент из БД + реальные проценты + «открытия» + «задачи дня»

**8.1 Контент района в БД.** Используем существующие `districts.description` + `districts.metadata_json`:

```json
{
  "image": "dist_parques",                       // ключ ассета из фиксированного набора
  "desc_i18n": { "ru": "Центр города…", "en": "…" },
  "lumi_tips": { "low": "…", "mid": "…", "high": "…" },
  "buildings": [
    { "name": "Jardín de Frases", "type": "grammar",  "x": 18, "y": 22 },
    { "name": "Mercado de Palabras", "type": "word_market", "x": 72, "y": 18 }
  ]
}
```

- Наполнение: seed-миграция для существующих районов обоих курсов (контент берём из прототипа `linglow-data.js`), правка — через админ-CRUD районов (минимальный `AdminLinglowDistrictsView`: JSON-редактор metadata + description; полноценный визуальный редактор координат уже есть — `Linglow Markup/District Editor.html`, остаётся внутренним инструментом).
- `handleCourseMap` (`/api/linglow/city`): добавить в ответ district'а `description` и `metadata` (whitelist-поля, не сырой JSON).
- `CityDistrictView.vue`: `districtDesc` ← `description`/`desc_i18n[locale]` (удалить эвристику по подстрокам), `buildings` ← metadata (нет — лейблы не рендерим), картинка ← `image` (маппинг ключ→импорт ассета), `lumiTip` ← `lumi_tips` по порогу confidence (фоллбек — текущие i18n).
- `CityView.vue`: картинку карточки района брать из того же `metadata.image` вместо ротации по индексу.

**8.2 Реальные проценты по активностям района.**
- Грамматика — уже реальная (по level_code). Оставить.
- Слова — уже реальные (confidence). Оставить.
- Чтение — из `by_location` c `location_type='reading'` (уже считается в `typeProgress`); **убрать** fallback `wordsPct*0.7`: нет данных → 0 и meta «ещё не начато».
- Общение — **убрать** `wordsPct*0.5`. v1: если в курсе есть `conversation`-локации с item'ами — их `attempted/total`; иначе строка «Общение» без прогресс-бара (кнопка-ссылка в чат остаётся). После этапа 1 (chat-события) можно показывать «N диалогов на этой неделе» вместо процента — честнее, чем выдуманный %.

**8.3 «Новые открытия» — реальная рекомендация.**
- `GET /api/linglow/discovery?district_code=…&course_code=…`: первый опубликованный текст чтения уровня района (`level_code`), которого нет у пользователя в `reading_text_progress`. Ответ: `{ kind: 'reading_text', chapter_id, title, category }` или 204.
- Карточка: реальный заголовок текста, тап → прямо в `ReadingTextView` этого текста. 204 → карточку скрыть.

**8.4 «Задачи дня» района** (блок из прототипа, не портирован):
- Расширить `GET /api/linglow/daily-route` параметром `district_code` **или** (проще) собрать на бэке секцию `district_tasks` в ответе discovery-ручки/нового `GET /api/linglow/district-tasks?district_code=…`:

```json
{ "tasks": [
  { "kind": "review_words", "target": 10, "done": 7, "route": "/training" },
  { "kind": "read_text",    "target": 1,  "done": 0, "route": "/reading/…" },
  { "kind": "chat",         "target": 1,  "done": 1, "route": "/chat" }
] }
```

- Генерация: due-слова района (SRS по location'ам района, цель = min(due,10)), 1 непрочитанный текст уровня, 1 чат-сессия. `done` — из событий за сегодня (`daily_course_stats`/`learning_events` с фильтром по району, для чата — по курсу).
- Фронт: новый блок `LgDailyTasks.vue` в `CityDistrictView` (вёрстка «N/M + прогресс-полоска» из прототипа).

**Готово, когда:** у района нет ни одного литерала контента в коде; «открытие» ведёт в конкретный текст; задачи дня закрываются галочками по мере выполнения.

---

### Этап 9. «Твой путь сегодня» на главной — генерация и отметки выполнения

- **Бэкенд:** расширить ответ `/api/linglow/daily-route` секцией `today`:

```json
"today": {
  "words_due": 12, "words_done": 5,
  "route_total": 8, "route_done": 3,
  "reading_done": false, "reading_suggestion": { "chapter_id": "...", "title": "..." },
  "chat_done": true
}
```

  - `words_done` — попытки word-режима за сегодня из `mode_daily_stats`; `reading_done` — `reading_text_progress` за сегодня; `chat_done` — `chat_message_sent` за сегодня; `reading_suggestion` — та же логика discovery (уровень пользователя).
- **Фронт** `DashboardView.vue` блок `lg-path-*`:
  - Строки рендерятся из ответа, а не статично: «Повтори N слов» (если due>0), «Путь на сегодня» (с `route_done/route_total` в подписи), «Прочитай: {title}» (ведёт в конкретный текст; всё прочитано — строка скрывается или «Открой новый текст»), «Диалог с ИИ».
  - Выполненные шаги — галочка/зачёркивание (стиль из прототипа `linglow-screens-home.jsx`), порядок: невыполненные сверху.
- Карточка «N слов пора повторить» уже реальная — не трогаем.

**Готово, когда:** в течение дня шаги отмечаются выполненными без перезагрузки (re-fetch при возврате на Главную).

---

## Волна 4 — геймификация и доводка

### Этап 10. Ачивки

**v1 — вычисляемые на лету, без таблиц.** В `/api/linglow/stats` секция:

```json
"achievements": [
  { "code": "streak",    "value": 6,  "unlocked": true,  "tier": 1 },
  { "code": "reader",    "value": 4,  "unlocked": true },
  { "code": "collector", "value": 128,"unlocked": true },
  { "code": "explorer",  "value": 3,  "unlocked": true },   // районов с attempted>0
  { "code": "expert",    "value": 0,  "unlocked": false }   // все районы ≥80% confidence
]
```

Источники: streak (этап 3), `reading_text_progress` (всего), слова в review/mastered, `progress.by_district`. Иконки/тексты/пороги — на фронте (i18n), бэк отдаёт только code/value/unlocked.

**v2 (отдельный PR, по желанию):** таблица `user_achievements (user_course_id, code, unlocked_at)` для тоста «новая ачивка» при первом достижении + пуш через notification_service. В v1 не делаем.

**Фронт:** `ProgressView` — вернуть блок ачивок из этапа 4, маппинг code→иконка/тайтл/саб. Залоченные — как в текущей вёрстке.

---

### Этап 11. «Что подтянуть» — рекомендации

Правило-ориентированная генерация (без LLM) в `/api/linglow/stats`:

```json
"improvements": [
  { "kind": "mode_accuracy", "mode": "grammar", "accuracy": 62 },
  { "kind": "due_backlog",   "count": 34 },
  { "kind": "weak_district", "district_code": "barrio", "title": "Barrio Vivo", "weakness": 12 }
]
```

Кандидаты (берём топ-3 по приоритету):
1. mode с минимальной accuracy за месяц (≥20 попыток);
2. due-бэклог > 20 слов;
3. район с max `weakness`/min confidence из by_district (среди attempted);
4. «нет чтения за 7 дней» / «нет чата за 7 дней» (по mode_daily_stats);
5. пустой список → совет «продолжай в том же духе».

Тексты — i18n-шаблоны на фронте по `kind` (`progress.imp.modeAccuracy` и т.п.), каждая строка кликабельна → ведёт в соответствующий раздел. LLM-генерацию персональных советов можно добавить потом тем же контрактом (`kind: "custom", text: "…"`).

---

### Этап 12. Интерактивная карта города (опционально, последним)

Перенос canvas-рендера из прототипа (`linglow-screens-map.jsx` + `City Map.html`):
- Конфигурация полигонов: `districts.metadata_json.polygon` (массив точек в % от размера картинки) — размечается через `District Editor.html`, сохраняется админ-ручкой из этапа 8.
- `CityView`: canvas поверх `map_city.jpg`; полигоны подсвечиваются прогрессом (`by_district.progress_percent`), lock — `district.status='locked'` или «предыдущий уровень < порога»; тап по полигону → район.
- Производительность в Telegram webview: рисуем один раз на загрузку + при изменении данных (не rAF-луп); фоллбек `<img>` + список (текущая вёрстка) для устройств без canvas/при ошибке.
- Логика «открытия» районов (locked-уровни) — отдельное продуктовое решение: v1 — все районы открыты, lock только по `status` из БД.

---

## Сводный порядок PR'ов

| # | PR | Зависит от | Риск |
|---|-----|-----------|------|
| 1 | События reading/chat/speaking + daily-агрегаты + backfill | — | низкий (бэк-only) |
| 2 | Heartbeat времени (`/api/linglow/activity` + `useActivityTracker`) | 1 | средний (клиентский трекинг) |
| 3 | `/api/linglow/stats` (streak, week, month, skills, fav district) | 1–2 | низкий |
| 4 | ProgressView на реальные данные (+скрыть ачивки/improvements) | 3 | низкий |
| 5 | Streak в шапках (`useStats`) | 3 | низкий |
| 6 | Профиль: имя/уровень/курс/стрик | 3 | низкий |
| 7 | Lumi-факты: таблица + ротация + админка + контексты | — (параллелится с 1–6) | низкий |
| 8 | Район: metadata/description + реальные проценты + discovery + задачи дня | 1 | средний |
| 9 | «Твой путь сегодня»: генерация + отметки выполнения | 1, 8 (discovery-логика) | низкий |
| 10 | Ачивки (computed) | 3 | низкий |
| 11 | «Что подтянуть» | 3 | низкий |
| 12 | Интерактивная карта | 8 (полигоны в metadata) | высокий (canvas/webview) |

После PR 1–9 в приложении не остаётся ни одного фейкового числа; 10–11 возвращают на Прогресс скрытые блоки; 12 — чистый polish.

## Риски и открытые вопросы

- **Часовые пояса**: правило `client_day ± 1 день` покрывает 99% случаев; для надёжности можно позже хранить tz в user settings.
- **Накрутка времени**: потолки на пинг/день + проверка на сервере; метрика некритичная (не соревновательная), параноить не нужно.
- **Двойной счёт активности** при дублирующих событиях (legacy dual-write + canonical): агрегаты наполнять только из одного пути — канонических `Record*`-методов; backfill должен дедуплицировать по `source_table/source_pk`.
- **Streak и оффлайн**: оффлайн-попытки синкаются позже с `client_day` — `BumpDaily` должен принимать дату из события, а не `now()`, иначе оффлайн-день выпадет из стрика.
- **Контент фактов для en_ru**: на старте нужно ~30 фактов на курс; генерим LLM'ом, вычитываем, грузим bulk-формой.
- **`UNIQUE(course_id, level_code)` в districts** — при добавлении районов следить за уникальностью уровня; контент-этап (8) схему не меняет.
