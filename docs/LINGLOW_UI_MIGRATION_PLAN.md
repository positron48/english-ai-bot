# Linglow UI Migration Plan — внедрение новой вёрстки в webapp

Дата: 2026-06-11
Статус: план, к реализации

## 0. Исходные данные

**Что есть:**
- `Linglow Markup/` — React-прототип нового дизайна (CDN React 18 + Babel standalone, inline-styles, статичные данные в `linglow-data.js`). Экраны: Home, Map (canvas-город), District, Practice, Grammar, Lesson, Exercise (Choice/Tiles/Write), Reading (list + text), Chat, Progress, Profile. Общие компоненты и дизайн-токены в `linglow-components.jsx` + `linglow.html`. Ассеты (Lumi, районы, карта) в `assets/` (~30 файлов).
- `webapp/` — боевое Vue 3 SPA (vite, vue-router, vue-i18n, TS). ~37 view'шек, из них 13 админских.

**Решения (зафиксированы):**
- React в прод не тащим — **портируем вёрстку в Vue SFC**. Прототип остаётся референсом.
- Админка (`/admin/*`, `AdminLayout`) остаётся на старых стилях.
- **Публичка и админка разделяются на две сборки (Vite MPA: два entry в одном проекте)** — полная изоляция CSS, старые стили в публичный бандл просто не попадают. Логика (composables, api-клиенты, i18n, общие компоненты) остаётся общей.
- Вся публичка переезжает на новую вёрстку, старые публичные стили выкидываем.
- Делаем только для unified-инстанса Linglow; отдельные en/es прод-проекты не трогаем.
- Сначала «натягиваем» существующий функционал на новую вёрстку 1:1, новые фичи прототипа (которых нет в бэке) — заглушками не делаем, просто не показываем.

---

## Фаза 1. Фундамент: split-сборка, токены, шрифты, ассеты

Ключевое архитектурное решение фазы: **публичка и админка — два отдельных entry в одной Vite-сборке**. Это снимает обе CSS-проблемы разом: конфликт имён переменных (`--card-bg` есть и в старой, и в новой теме) и утечку старых глобальных правил (`body`, `.container`, ресеты) на публичные страницы после визита в админку — в SPA лениво подгруженный CSS остаётся жить навсегда, а при split старые стили в публичный бандл вообще не попадают.

### 1.0 Разделение сборки (Vite MPA)
- **Entry публички**: `webapp/index.html` + `src/main.ts` → публичный App + публичный роутер (все маршруты, кроме `/admin/*`). Импортирует только новую тему.
- **Entry админки**: `webapp/admin.html` + `src/admin-main.ts` + `src/AdminApp.vue` (текущий каркас из `App.vue` в минимальном виде) + `src/router/admin.ts` (поддерево `/admin` с base `/app/admin` + guard'ы requiresAuth/requiresAdmin из текущего роутера). Импортирует старые `style.css` / `styles/theme.css`.
- Общее (без дублирования): `composables/`, `api/`, `i18n`, `locales/`, `components/`, `utils/`. Auth — общий localStorage, тот же origin.
- **`vite.config.ts`**: `build.rollupOptions.input = { index, admin }`; в dev-плагине `spa-fallback` добавить ветку — `/app/admin*` → `/admin.html`, остальное как сейчас → `/index.html`.
- **`internal/web/webapp_routes.go`**: в SPA-fallback'е выбирать файл по пути — `/app/admin*` → `admin.html`, иначе `index.html` (~5 строк; `embed.go` менять не нужно, кладётся весь `dist`).
- Переходы публичка↔админка — полная перезагрузка (`<a href="/app/admin">` вместо `router-link`); в обратную сторону аналогично. Для админки приемлемо.
- SW/офлайн: `write-asset-manifest.mjs` уже кладёт весь `dist/assets` в манифест — оба бандла кешируются, офлайн для админки не требуется. Проверить, что `admin.html` не ломает app-shell-логику в `sw.js` (в `APP_SHELL_URLS` его добавлять не нужно).
- Тесты: `internal/web/webapp_routes_test.go` и соседние — дополнить кейсами fallback'а на `admin.html`.

### 1.1 Дизайн-токены
- Создать `webapp/src/styles/linglow-theme.css` — перенос блоков `:root[data-theme="light"]` / `:root[data-theme="dark"]` из `linglow.html` (строки 14–140: токены, скрытие скроллбара, transitions, dark-градиент body, фильтры Lumi). Имена переменных — **как в макете, без префиксов**: конфликт со старой темой невозможен, она живёт в другом entry.
- Подключается только в `src/main.ts` (публичный entry).
- Существующий `useTheme` уже ставит `data-theme` на `<html>` — новая тема цепляется к тому же атрибуту, переключатель темы переиспользуем как есть.

### 1.2 Шрифты
- Прототип тянет Lora + Inter с Google Fonts. Для PWA/Telegram mini app/офлайна — **self-host**: положить woff2 (Lora 400/600/700, Inter 400/500/600/700) в `webapp/public/fonts/`, `@font-face` в `linglow-theme.css`, `font-display: swap`. Проверить латиницу+кириллицу subsets.

### 1.3 Ассеты
- Скопировать `Linglow Markup/assets/*` → `webapp/src/assets/linglow/` (через vite import, попадут в манифест и кеш sw). Карта `map_city.jpg` и здания/районы jpg — крупные, прогнать через сжатие (webp + fallback или просто пережать jpg ≤200KB).
- Проверить, что `webapp/scripts/write-asset-manifest.mjs` и `sw.js` подхватывают новые ассеты для офлайна.

### 1.4 Публичный layout
- Создать `webapp/src/layouts/PublicLayout.vue` (аналог desktop/mobile-логики из `linglow-app.jsx`):
  - корневой контейнер с `background: var(--bg)`, `font-family: Inter`, `color: var(--text)` (body-стили из `linglow.html`);
  - `>= 900px`: `LgSideNav` слева + контент по центру (`max-width: 880px`, для карты 760px);
  - `< 900px`: контент full-width + `LgBottomNav` (fixed, blur, safe-area-inset-bottom);
  - флаг «полноэкранный режим» через `route.meta.fullscreen` — для exercise/chat-сессий навигация скрывается (как `showNav` в прототипе);
  - сюда же переносятся из `App.vue`: network-toast, auth-error banner, Alert/Confirm-модалки можно оставить глобально в App.vue.
- Публичный роутер: маршруты оборачиваются в `PublicLayout` (children), `/login` — без layout'а; маршруты `/admin/*` из него удаляются (живут в admin-entry). Guard'ы из `router/index.ts` сохраняются без изменений, admin-специфичные уезжают в `router/admin.ts`.
- Публичный `App.vue` худеет до `<router-view/>` + глобальные модалки/тосты; вся навигационная разметка (navbar-desktop, navbar-mobile, sidebar) уезжает в layout'ы или умирает вместе со старым entry.

### 1.5 Брейкпоинт
- Прототип использует 900px, текущее приложение — 768px. Принять **900px** как единый публичный брейкпоинт (константа/CSS custom media), `isMobile` в коде привести к нему.

**Результат фазы:** приложение выглядит по-старому внутри страниц, но уже живёт в новом каркасе с новой навигацией и темой. Это самый безопасный момент для первого деплоя.

---

## Фаза 2. Библиотека компонентов `components/linglow/`

Портировать `linglow-components.jsx` в Vue SFC, inline-styles переводить в scoped CSS на токенах новой темы. Имена компонентов с префиксом `Lg`:

| Прототип | Vue-компонент | Примечания |
|---|---|---|
| `LumiSVG` | `LgLumi.vue` | props: size, pose (default/book/pencil/map); фильтры по теме уже в css |
| `AppTopBar` | `LgTopBar.vue` | селектор языка → **подключить к `useCourse`** (реальные курсы вместо хардкода «Испанский/Английский»); streak — из API |
| `SideNav` | `LgSideNav.vue` | 5 табов: Главная/Город/Практика/Прогресс/Профиль → router-link'и; active-логика по `route.meta.navTab` |
| `BottomNav` | `LgBottomNav.vue` | то же + scroll-border эффект (capture-листенер) |
| `CenteredHeader`, `BackHeader` | `LgPageHeader.vue` | объединить: back-кнопка / гамбургер / титул / right-slot / streak |
| `ProgressBar` | `LgProgressBar.vue` | |
| `CircleRing` | `LgCircleRing.vue` | |
| `StreakBadge` | `LgStreakBadge.vue` | |
| `LevelChip`, `NewChip` | `LgChip.vue` | вариант через prop |
| `PrimaryBtn` | `LgButton.vue` | + вторичный/disabled варианты |
| `LumiTip`, `LumiFactCard` | `LgLumiTip.vue` / `LgLumiFact.vue` | факты — пока массив в коде, позже из API/i18n |
| `SpeechBubble` | `LgSpeechBubble.vue` | |
| иконки (Home, MapPin, Book, Dumbbell, Bar, User, Chevron*, X, Check, Sound, Moon, Sun, Send, Mic, Pin, Hamburger, Clock, Refresh) | расширить существующий `Icon.vue` или `LgIcon.vue` со встроенными svg | один компонент с name-prop, как уже принято в проекте |

Принципы:
- Все строки — через vue-i18n (в прототипе хардкод RU). Ключи в `locales/ru.json` + перевод в `es`/`en`.
- Никаких данных внутри компонентов — только props/slots.
- Тема — через существующий `useTheme` (toggle уже есть).

---

## Фаза 3. Миграция экранов (по одному, каждый — отдельный PR)

Маппинг «экран прототипа → существующие view» и порядок (от простого/изолированного к сложному):

### 3.1 Профиль / Настройки
- `ProfileScreen` → рестайл `SettingsView.vue`. Сюда же: переключатель темы, переключатель UI-языка (RU/target из App.vue), выбор курса, logout.
- Низкий риск, обкатывает компонентную базу.

### 3.2 Home (Главная)
- `HomeScreen` → `DashboardView.vue` (1515 строк, переписываем шаблон).
- Секции прототипа → данные: header (имя, streak — `/api/dashboard`), «Ciudad Luminaria» (карточка-вход в город — city API), «Твой путь сегодня» (daily route — `CityDailyRouteView`-данные), карточка активного района, «N слов» (SRS due — данные тренировок), «Совет от Lumi».
- Чего в API нет — секцию не рендерим (v1 = реальные данные, не моки).

### 3.3 Город: карта + район
- `MapScreen` (canvas: фон-jpg + полигоны районов + lock-уровни) → `CityView.vue`. Canvas-рендер из прототипа портируется почти как есть (чистые функции), полигоны районов — из `linglow-data.js` как статичная конфигурация, прогресс/уровни — из city API.
- `DistrictScreen` → `CityDistrictView.vue`: шапка района, confidence, «практикуешь сейчас», задачи дня, здания (Grammar/Reading/Conversation/Repaso → ссылки на соответствующие разделы).
- `CityDailyRouteView` → секция «путь сегодня» в новом стиле.
- Сверить коды районов прототипа (plaza/barrio/…) с реальными `districtCode` из БД — мэппинг или правка конфигурации полигонов. `District Editor.html` оставить как внутренний инструмент для разметки полигонов.

### 3.4 Практика
- `PracticeScreen` → объединяет текущие `LearningView` + `TrainingView`-хаб: hero-карточка «продолжить», quick-launches, грид 2×2 режимов (грамматика/слова/чтение/разговор), «мой словарь» (→ `VocabView`/`WordSets`), Lumi-факт.
- Таб «Практика» в навигации активен для всего поддерева `/learning/*`, `/training/*` (как `practiceIds` в прототипе).

### 3.5 Грамматика
- `GrammarScreen` (список глав со StatusBadge/ChapterIcon) → `GrammarCategoriesView` + `GrammarChaptersView`.
- `LessonScreen` (теория) → `GrammarChapterView` (+ `GrammarTheoryExamples`, markdown-стили: адаптировать `markdown-content.css` под `--lg-*`).
- Placement test и training-страницы — рестайл на тех же компонентах.

### 3.6 Упражнения (самый большой кусок)
- `ExerciseScreen` (shell: прогресс-хедер, FeedbackBanner, Choice/Tiles/Write) → рестайл `GrammarTestView` (1588), `TrainingView` (4966), `WordSetStudyView`, `VerbTrainingView`.
- Стратегия: вынести из прототипа **четыре переиспользуемых компонента** — `LgExerciseShell` (хедер с прогрессом + крестик выхода), `LgChoiceExercise`, `LgTilesExercise`, `LgWriteExercise`, `LgFeedbackBanner` — и подключать их к существующей логике view'шек, не трогая state-машины тренировок. Логику (SRS, офлайн-store, задержки ответов `useTrainingAnswerDelay`) не переписываем — меняем только presentation-слой.
- `route.meta.fullscreen = true` для сессий.
- Это 3–5 отдельных PR (по типу тренировки).

### 3.7 Чтение
- `ReadingListScreen` → `ReadingCategoriesView` + `ReadingChaptersView`.
- `ReadingTextScreen` → `ReadingTextView` (+ `ReadingPassageBlock`, кнопки озвучки через `useAudio`/TTS — иконка Sound из набора).

### 3.8 Чат
- `ChatScreen` → `ChatView.vue`: fullscreen, пузыри, инпут с Send/Mic (Mic — `SpeakingRecordingPanel`/`useAudioRecorder`, если включён speaking). Markdown в сообщениях (`marked`) — стилизовать под тему.
- Speaking-сессии (`SpeakingSessionView`) — тот же стиль чата.

### 3.9 Прогресс
- `ProgressScreen` (месячная сводка, ритм, районы, сильные стороны, ачивки) → **новый** view `ProgressView.vue` + маршрут `/progress` + таб в навигации. Наполняем тем, что реально отдаёт API (статистика из dashboard/stats); блоки без данных не показываем. Возможно, частично поглощает статистику из старого Dashboard.

### 3.10 Login
- В прототипе нет — рестайл `LoginView.vue` на новых токенах (логотип Lora, Lumi, PrimaryBtn). Telegram-auth логика не трогается.

---

## Фаза 4. Чистка и доводка

1. **Выкинуть старые публичные стили**: `style.css` и `styles/theme.css` импортируются только в `admin-main.ts` — из публичного entry они исчезают автоматически; внутри них удалить правила, которые нужны были только публичке. Удалить мёртвый CSS из мигрированных view (старые scoped-стили уходят вместе с шаблонами). Проверить bundle-анализом, что в публичный чанк не затащило админский код/стили через общие импорты.
2. **Breadcrumbs** — в новом дизайне их нет; компонент остаётся только если нужен в админке, из публички убрать.
3. **i18n-ревизия**: все новые ключи в ru/en/es, прогнать `npm run type-check`.
4. **Офлайн/PWA**: проверить precache новых ассетов/шрифтов в sw.js, офлайн-маршруты из router guard работают в новом layout.
5. **Telegram mini app**: `theme-color`/`setSystemBarsColor` теперь должен брать `--bg` новой темы на публичке (правка `updateThemeMetaColor`), safe-area для BottomNav, отсутствие logout в TG.
6. **Android embedded** (`android-embedded/`, `QantrixAndroid`) — smoke-тест webview.
7. Удалить `Linglow Markup/` из деплой-артефактов (в репо можно оставить как референс, в Docker-образ не копировать) — проверить `.dockerignore`/Dockerfile.

---

## Порядок PR'ов (итог)

1. Фаза 1, шаг 1: **split-сборка** (два entry, vite-конфиг, Go-fallback, dev-режим) — деплоится, визуально ничего не меняется.
2. Фаза 1, шаг 2: токены+шрифты+ассеты+PublicLayout+навигация — **деплоится, страницы работают по-старому внутри нового каркаса**.
3. Библиотека `Lg*`-компонентов (+ Storybook не нужен — есть прототип как референс).
4. Settings/Profile.
5. Home/Dashboard.
6. City: карта → район → daily route.
7. Practice-хаб.
8. Grammar: списки → глава/теория.
9. Exercise-компоненты + по одному PR на каждый тип тренировки (grammar test, word study, training, verbs).
10. Reading.
11. Chat + Speaking.
12. Progress (новый view).
13. Login.
14. Чистка (фаза 4).

Каждый шаг деплоится независимо: до конца миграции часть страниц старые, часть новые — это ок, общий каркас уже новый с фазы 1.

## Риски / открытые вопросы

- **Изоляция CSS** — закрыта split-сборкой (старые стили не попадают в публичный entry). Правило: ничего из старых css-файлов не импортировать в `main.ts`/публичные компоненты.
- **Split-сборка**: следить, чтобы общие компоненты не тянули admin-only зависимости в публичный чанк; fallback-логика в Go и vite-dev-плагине должна совпадать (`/app/admin*` → `admin.html`).
- **TrainingView (5k строк)** — самая дорогая часть; не пытаться переписать за раз, мигрировать по типам упражнений с переиспользованием `LgExercise*`.
- **Полигоны карты vs реальные районы** — коды районов в БД должны совпасть с конфигурацией карты; нужен seed/маппинг + District Editor для новых районов.
- **Шрифты+кириллица** — Lora и Inter имеют cyrillic subsets, обязательно включить в self-host.
- **Перф canvas-карты в Telegram webview** — проверить на слабом Android; fallback — статичная картинка без подсветки полигонов.
- **Verb training (испанские глаголы)** — экрана в прототипе нет; делаем на `LgExercise*`-компонентах по аналогии.
