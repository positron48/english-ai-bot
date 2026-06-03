# PWA/APK: офлайн-режим грамматики

Документ описывает фактически реализованную модель Android PWA/TWA APK и офлайн-грамматики для English/Spanish инстансов `english-ai-bot`.

## Итоговая карта реализации

Что было добавлено в рамках PWA/APK/offline-работ:

- Android TWA/PWA APK для двух приложений: English (`ru.qantrix.english`, `https://qantrix.ru/app/`) и Spanish (`ru.qantrix.spanish`, `https://es.qantrix.ru/app/`).
- PWA endpoints в web/backend: `/sw.js`, `/app/manifest.webmanifest`, `/app/icons/*`, `/.well-known/assetlinks.json`.
- Подпись APK через GitHub Actions secrets и Bubblewrap.
- GitHub Release assets для APK: `qantrix-english-<tag>.apk`, `qantrix-spanish-<tag>.apk`, `checksums.txt`.
- Offline grammar preload: manifest, главы, question bank, chapter/category/placement tests, Grammar Training/SRS pack.
- Offline word training preload: current due/new word cards для обычной multiple-choice тренировки.
- IndexedDB-хранилища для скачанных grammar/word данных и очередей попыток.
- Idempotent sync на сервере через `client_attempt_id`, чтобы повторная отправка не создавала дубли.
- Service worker cache для app shell и всех Vite assets, чтобы lazy routes открывались офлайн.
- Offline UI: preload panels, статус online/offline, pending sync, toast при потере/возврате сети, disabled/redirect для online-only разделов.
- Android/TWA polish: `fallbackType=webview`, светлый status bar/theme color, проверка Digital Asset Links через `assetlinks.json`.

Ключевое разделение:

- APK содержит только wrapper metadata: package id, server URL, icons/manifest metadata, подпись, TWA config.
- Web frontend, service worker и API приходят с сервера.
- Курсы, тесты и тренировочные данные не зашиваются в APK; пользователь скачивает их кнопками preload после логина.
- Offline source of truth на устройстве — IndexedDB.
- Server source of truth сохраняется после sync; сервер повторно применяет canonical grading/SRS logic.

## Что работает офлайн

После предзагрузки грамматики офлайн доступны:

- список категорий грамматики;
- список глав внутри категорий;
- теория глав;
- inline quiz внутри главы;
- тест по главе;
- тест по категории;
- placement test на основе локально сохранённого question bank;
- локальное отображение прогресса по результатам офлайн-тестов;
- очередь результатов с последующей синхронизацией на сервер;
- отдельный экран `Grammar Training` / SRS-тренировка грамматики (`/learning/grammar/training`) на основе предзагруженного training pack.
- обычная тренировка слов (`/training`) в offline-v1 multiple-choice режиме после отдельного `Preload words`.

Не работают офлайн в текущей версии:
- словарь и word sets;
- spell/type варианты тренировки слов и server-only варианты генерации training queue;
- chat/LLM;
- reading;
- speaking;
- админка;
- любые серверные операции вне offline grammar/word training API.

Важно: в UI есть два похожих понятия. “Тесты по грамматике” (chapter/category/placement tests) и “Grammar Training” как отдельный режим повторения/SRS теперь оба работают офлайн после preload, но синхронизируются разными очередями.

## Как пользователь получает APK

GitHub tag release в `english-ai-bot` собирает два TWA APK:

- `qantrix-english-<tag>.apk`: Android package `ru.qantrix.english`, открывает `https://qantrix.ru/app/`;
- `qantrix-spanish-<tag>.apk`: Android package `ru.qantrix.spanish`, открывает `https://es.qantrix.ru/app/`.

APK не содержит курс внутри себя. В APK фактически зашиты package id, URL сервера, TWA metadata и подпись. Сам web shell, service worker и курс приходят с выбранного сервера.

Для verified TWA сервер должен отдавать `/.well-known/assetlinks.json`. SHA-256 fingerprint релизного Android signing certificate задаётся в GitOps через `WEBAPP_ANDROID_CERT_FINGERPRINTS` в соответствующем ConfigMap English/Spanish.

Что лежит в APK:

- package id (`ru.qantrix.english` или `ru.qantrix.spanish`);
- host/start URL (`https://qantrix.ru/app/` или `https://es.qantrix.ru/app/`);
- TWA config, app name, цвета status/navigation bar;
- ссылки на web manifest/icon URL;
- Android signing metadata.

Что не лежит в APK:

- grammar course JSON;
- grammar tests/question bank;
- Grammar Training pack;
- word training pack;
- пользовательский progress;
- web frontend bundle как постоянный embedded asset.

Это значит: большинство frontend/offline исправлений раскатывается через серверный webapp и не требует переустановки APK. Переустановка APK нужна, если поменялись package id, TWA config, fallback strategy, signing, icon/start URL или нужно проверить новый release asset.

## CI и GitHub Releases

Основной workflow: `.github/workflows/ci.yml`.

На обычный push в `master` запускаются:

- webapp dependency install;
- `npm run type-check`;
- `npm run build`;
- Go dependency verify;
- Go tests;
- golangci-lint.

На tag push дополнительно запускаются:

- `release` job: собирает server binary artifacts и загружает их в GitHub Release;
- `docker-image` job: собирает и публикует `ghcr.io/<owner>/english:<tag/latest>` и `ghcr.io/<owner>/spanish:<tag/latest>`;
- `android-apk` matrix job: собирает English/Spanish APK и загружает их в тот же GitHub Release.

Важно: `android-apk` сейчас оставлен в ускоренном режиме и не ждёт полный `release` job. Это временная настройка, чтобы быстрее стабилизировать APK CI. После стабилизации можно вернуть зависимость:

```yaml
needs: [release]
```

APK job делает следующее:

1. Ставит Node.js, JDK 17 и Android SDK.
2. Декодирует `ANDROID_KEYSTORE_BASE64` в `release.keystore`.
3. Создаёт GitHub Release, если его ещё нет.
4. Запускает `scripts/build-twa-apks.sh` для `english` и `spanish`.
5. Скрипт генерирует `dist/twa-<app>/twa-manifest.json`.
6. Скрипт заранее пишет `${HOME}/.bubblewrap/config.json`, чтобы Bubblewrap не спрашивал про JDK интерактивно.
7. Запускает `bubblewrap update --skipVersionUpgrade`, чтобы появился Android project и checksum без prompt.
8. Принудительно проверяет/патчит `fallbackType=webview`.
9. Запускает `bubblewrap build --skipPwaValidation`.
10. Кладёт APK в `dist/qantrix-<app>-<tag>.apk`.
11. Обновляет `checksums.txt` и release assets через `gh release upload --clobber`.

GitHub Actions secrets для APK:

```text
ANDROID_KEYSTORE_BASE64
ANDROID_KEYSTORE_PASSWORD
ANDROID_KEY_ALIAS
ANDROID_KEY_PASSWORD
```

Локальный signing runbook: `docs/ANDROID_APK_SIGNING_RU.md`.

## GitOps/k3s часть

Для APK и offline frontend важны оба слоя: новый Docker image должен выкатиться на prod до успешной APK-сборки, потому что Bubblewrap скачивает web manifest/icon URL с реального домена.

English:

- namespace: `english`;
- deployment: `english`;
- image: `ghcr.io/positron48/english`;
- GitOps manifests: `devops-time-host/apps/english/base/`;
- public URL: `https://qantrix.ru`.

Spanish:

- namespace: `spanish`;
- deployment: `spanish`;
- image: `ghcr.io/positron48/spanish`;
- GitOps manifests: `devops-time-host/apps/spanish/base/`;
- public URL: `https://es.qantrix.ru`.

Digital Asset Links:

- backend отдаёт `/.well-known/assetlinks.json`;
- fingerprint берётся из env `WEBAPP_ANDROID_CERT_FINGERPRINTS`;
- значение хранится в GitOps ConfigMap для English/Spanish;
- если fingerprint не совпал с APK signing cert, verified TWA не поднимется и Android может открыть fallback.

Перед rerun APK job проверять prod URL:

```bash
curl -fsSI https://qantrix.ru/app/manifest.webmanifest
curl -fsSI https://qantrix.ru/app/icons/english-512.png
curl -fsSI https://qantrix.ru/app/asset-manifest.json
curl -fsSI https://es.qantrix.ru/app/manifest.webmanifest
curl -fsSI https://es.qantrix.ru/app/icons/spanish-512.png
curl -fsSI https://es.qantrix.ru/app/asset-manifest.json
```

Если там `404`, APK CI будет падать на скачивании manifest/icon или offline lazy routes не будут надежно precache-иться.

## Когда скачивается курс грамматики

Курс скачивается только по явному действию пользователя:

1. Пользователь устанавливает APK.
2. Открывает приложение с интернетом.
3. Логинится, чтобы получить обычные JWT/refresh tokens.
4. Заходит в грамматику.
5. Нажимает кнопку `Preload grammar` в блоке `Offline grammar`.

После нажатия frontend вызывает:

- `GET /api/learning/grammar/offline/manifest` — получает список опубликованных категорий/глав, версию bundle, URL глав, текущий прогресс пользователя и примерный размер.
- Для каждой опубликованной главы: `GET /api/learning/grammar/offline/chapters/{chapter_id}` — получает полный JSON главы, включая `question_bank` и `chapter_test`, чтобы локально показывать теорию и проверять ответы.
- `GET /api/learning/grammar/offline/training-pack` — получает доступные пользователю вопросы Grammar Training/SRS из `grammartrainingpack`.

Данные сохраняются в IndexedDB браузера/TWA:

- metadata bundle (`version_hash`, `bundle_id`, `target_lang`, дата скачивания);
- sections/categories;
- chapters;
- локальный progress snapshot;
- очередь офлайн-попыток по тестам;
- очередь офлайн-попыток Grammar Training/SRS.

Service worker отдельно кэширует web shell (`/app`, `/app/`, manifest, vendored Telegram script) и все Vite assets из `/app/asset-manifest.json`, чтобы lazy-loaded страницы грамматики могли открываться без сети даже после прямого перехода на route. Контент курса хранится не в CacheStorage, а в IndexedDB.

`/app/asset-manifest.json` генерируется на каждом production build скриптом `webapp/scripts/write-asset-manifest.mjs`. Если офлайн-переход в APK даёт чёрный экран, в первую очередь проверять, что на сервере доступен `/app/asset-manifest.json`, а service worker обновился до актуальной версии cache.

## Что видно на фронте

На странице грамматики есть блок `Offline grammar`:

- показывает `Online` или `Offline` badge по `navigator.onLine`;
- показывает, готова ли предзагрузка (`Ready: X/Y chapters`);
- показывает количество результатов, ожидающих синхронизации;
- даёт кнопки `Preload grammar` / `Update preload`, `Sync results`, `Delete`.

Если пользователь офлайн и пытается открыть режимы вне поддержанного набора grammar offline, router перенаправляет его обратно в грамматику. В текущей минимальной реализации нет отдельного глобального баннера на всех экранах; явный статус сети находится в блоке грамматики.

Для диагностики runtime-падений frontend сохраняет последние ошибки в `localStorage`:

- `qantrix-offline-debug-state` — последнее действие/переход в grammar offline UI;
- `qantrix-runtime-error` — Vue/window/router error с URL, online-state и stack/message.

В офлайн-режиме такие ошибки показываются поверх приложения отдельным debug overlay вместо пустого чёрного экрана.

## Как работают офлайн-тесты

Когда приложение офлайн (`navigator.onLine === false`) или любой grammar API-запрос упал сетевой ошибкой:

1. `grammarClient` читает нужные главы из IndexedDB.
2. Для chapter test выбирает вопросы из локального `chapter_test.pool_question_ids` / `question_bank`.
3. Для category test собирает вопросы из опубликованных глав категории.
4. Для placement test выбирает вопросы из локально сохранённых глав.
5. Ответы проверяются локально той же базовой логикой сравнения: строки нормализуются, массивы сравниваются без учёта порядка, true/false поддерживает `true/false`, `да/нет`, `yes/no`, `1/0`.
6. Результат сразу показывается пользователю и обновляет локальный progress snapshot.
7. Попытка сохраняется в IndexedDB queue с `client_attempt_id`.

Важно для Android/TWA: `navigator.onLine` на реальных устройствах может быть слишком оптимистичным. Поэтому frontend не полагается только на него: категории, главы, проверки доступа, chapter/category/placement tests и Grammar Training сначала пробуют сервер, а при сетевой ошибке автоматически переходят на IndexedDB.

Ограничение: локальная генерация тестов не пытается полностью воспроизвести все серверные нюансы отбора/placement-level алгоритма. Это минимальная offline-v1 модель: пользователь может заниматься и видеть результат сразу, а сервер при синхронизации пересчитает и сохранит результат по своему canonical logic.


## Как работает офлайн Grammar Training / SRS

После preload frontend сохраняет доступные пользователю вопросы из training pack. В офлайне экран `/learning/grammar/training`:

1. Проверяет наличие локальных training questions в IndexedDB.
2. Собирает сессию локально: группирует вопросы по `theory_block_id`, выбирает до 20 блоков, берёт один случайный вопрос на блок.
3. Для теоретической подсказки читает уже предзагруженную главу из IndexedDB.
4. Проверяет ответ локально той же базовой логикой сравнения.
5. Показывает feedback сразу.
6. Кладёт ответ в отдельную очередь `training_queue` с `client_attempt_id`.

При восстановлении сети очередь отправляется на:

```http
POST /api/learning/grammar/offline/sync-training-attempts
```

Backend повторно проверяет ответ через canonical `SubmitGrammarSrsAnswerWithClientAttemptID`, обновляет `grammar_theory_memory` и пишет `grammar_attempts`. Для защиты от дублей у `grammar_attempts` добавлен `client_attempt_id` и уникальный индекс `(user_id, client_attempt_id) WHERE client_attempt_id IS NOT NULL`.

Ограничение v1: локальная SRS-сессия не воспроизводит полностью серверный due-order по `next_review_at`, если устройство долго было офлайн. Она даёт рабочую тренировку и очередь ответов; canonical SRS-состояние пересчитывается на сервере при sync.

## Когда происходит синхронизация

Синхронизация офлайн-результатов происходит несколькими способами:

- автоматически по browser event `online`;
- автоматически каждые 30 секунд, если `navigator.onLine !== false`;
- вручную кнопкой `Sync results` в блоке `Offline grammar`;
- дополнительно при возврате в grammar page статус очереди перечитывается из IndexedDB.

Frontend отправляет очередь на:

```http
POST /api/learning/grammar/offline/sync-attempts
```

Payload содержит массив попыток:

- `client_attempt_id`;
- `scope`: `chapter` или `category`;
- `scope_id`;
- `answers`;
- `course_version`.

Backend вызывает обычную серверную проверку `SubmitTestWithClientAttemptID`, сохраняет попытку в `grammar_test_attempts` и обновляет `grammar_progress` так же, как обычный online submit. Grammar Training queue синхронизируется отдельно через `sync-training-attempts` и сохраняется в `grammar_attempts` / `grammar_theory_memory`.

## Idempotency и повторы

В БД добавлены поля:

```sql
grammar_test_attempts.client_attempt_id
grammar_attempts.client_attempt_id
```

И уникальные индексы:

```sql
(user_id, client_attempt_id) WHERE client_attempt_id IS NOT NULL
```

Это значит:

- если телефон отправит одну и ту же офлайн-попытку повторно, сервер не создаст дубль ни для grammar tests, ни для Grammar Training/SRS;
- frontend может безопасно повторять sync после обрыва связи;
- успешные элементы удаляются из локальной очереди;
- неуспешные остаются в очереди до следующей попытки.

## Что происходит при истёкшем токене

Офлайн-результаты не теряются: они остаются в IndexedDB queue.

Если access/refresh token истёк или сервер вернул auth error, sync не сможет завершиться. После повторного логина пользователь может снова нажать `Sync results` или дождаться автоматической синхронизации при восстановленной авторизации.

## Обновление курса

Кнопка `Update preload` повторно получает manifest и главы с сервера и перезаписывает локальный bundle metadata/chapters. Версия определяется `version_hash`, который строится сервером из embedded grammar bundle index/sections.

Если курс обновился на сервере, пользователь должен обновить preload, чтобы офлайн-режим использовал новую версию. Автоматического фонового скачивания новых глав в текущей v1 нет.

## Офлайн-тренировка слов

Добавлен отдельный offline-v1 режим для обычной тренировки слов (`/training`). Он работает независимо от grammar preload и хранится в отдельной IndexedDB базе `qantrix-word-training-offline`.

### Что скачивается

На экране тренировки слов есть блок `Offline word training` с кнопкой `Preload words`. При нажатии онлайн frontend вызывает:

```http
GET /api/training/offline/pack
```

Сервер собирает текущий тренировочный пул пользователя:

- due cards из `user_cards`, у которых `next_due_at` пустой или уже наступил;
- новые карточки пользователя (`state = new`);
- связанные `training_cards` и базовые поля слова;
- готовые варианты ответа и правильный ответ для multiple choice;
- подсказку, пример, транскрипцию и morph metadata, если они есть.

В APK ничего из слов не зашивается. На устройстве лежат только данные, которые пользователь явно скачал кнопкой preload после логина.

### Что работает офлайн

Офлайн поддержана обычная word training с выбором варианта ответа (`card` / multiple choice). Экран `/training` разрешён router-ом офлайн только если word training pack уже скачан.

Пока не воспроизводятся локально spell/type варианты тренировки слов. Причина: live-режим на сервере подменяет часть карточек на spell/type на основании актуального mastering score и server-side queue generation. В offline-v1 для минимального риска используется стабильный multiple-choice pack, а SRS пересчитывается сервером после синхронизации.

### Локальная попытка и видимость статуса

Когда пользователь офлайн:

1. `Start Training` создаёт локальную сессию из скачанного pack, максимум 30 карточек.
2. `Reveal` показывает заранее скачанные options.
3. Ответ проверяется локально по `correct_answer`.
4. Пользователь сразу видит correct/incorrect, hint/example и локальное завершение сессии.
5. Попытка кладётся в IndexedDB queue с `client_attempt_id`.

В блоке `Offline word training` видно:

- `Online` / `Offline`;
- сколько карточек скачано;
- сколько результатов ждёт синхронизации;
- кнопки `Update preload`, `Sync results`, `Delete preload`.

### Синхронизация слов

Синхронизация запускается:

- автоматически при событии браузера `online`;
- периодически из `main.ts` раз в 30 секунд, если сеть есть;
- вручную кнопкой `Sync results` на экране тренировки.

Frontend отправляет очередь в:

```http
POST /api/training/offline/sync-attempts
```

Backend для каждой попытки:

- проверяет `client_attempt_id` в `review_events`;
- если такая попытка уже есть у пользователя, возвращает success без повторного SRS update;
- проверяет, что `user_card_id` принадлежит текущему пользователю;
- применяет canonical `SRSService.GradeCard` к актуальному `user_cards`;
- создаёт `review_events` с `client_attempt_id`;
- для неправильных ответов обновляет wrong answers;
- закрывает синтетическую training session и пересчитывает `user_word_mastering` для затронутых слов.

Idempotency хранится в `review_events.client_attempt_id`; уникальный индекс: `(user_id, client_attempt_id)` where `client_attempt_id is not null`.

### Новые public API

```http
GET /api/training/offline/pack
POST /api/training/offline/sync-attempts
```

### Ограничения v1

- В pack попадают только карточки, уже созданные для пользователя в `user_cards`; каталоги слов сами по себе не скачиваются как новые user cards.
- Локальная тренировка слов поддерживает multiple choice, но не spell/type.
- Если пользователь долго занимался офлайн, сервер при sync применяет попытки последовательно к текущему состоянию карточек; это ожидаемое поведение и сохраняет сервер как source of truth.

## Offline UI behavior fixes

Текущая модель UI после доработки:

- В обычном браузере offline preload панели скрыты по умолчанию, чтобы не шуметь в web-версии. Они показываются в установленном PWA/TWA, при отсутствии сети, при уже скачанном bundle/pack или при pending sync.
- При потере/возврате сети показывается один глобальный toast с автоскрытием и кнопкой закрытия. Постоянный retry-banner на экране тренировки не используется для ожидаемого offline-mode.
- `/dashboard`, `/training` и offline grammar routes разрешены в router без редиректа на grammar root. Остальные online-only разделы офлайн редиректятся на dashboard с offline-state.
- Dashboard офлайн показывает только локально предзагруженные агрегаты: word training pack и grammar progress из IndexedDB.
- Завершенная офлайн word-training сессия не восстанавливается повторно как активная при новом входе на `/training`; пройденные локально карточки вычитаются из локального word pack до следующей предзагрузки.
- Offline grammar chapter access на фронте повторяет progression: первая глава доступной категории открыта, следующие открываются только после `passed` предыдущей главы. Доступность всей категории больше не раскрывает все главы сразу.
- Offline grammar training pack скачивает все опубликованные grammar training questions, чтобы тренировка могла стартовать из предзагруженного курса без зависимости от текущего server-side due/progress фильтра.
- Android status bar / PWA theme color выставлен в основной фон приложения (`#f5f5f5` для light theme) вместо language-specific brand color.

## Troubleshooting по уже пойманным проблемам

### APK CI: нет signing secret

Симптом:

```text
ANDROID_KEYSTORE_BASE64 is required
```

Причина: secrets не добавлены в GitHub Actions или добавлены не в тот repo.

Исправление: добавить 4 secrets из `docs/ANDROID_APK_SIGNING_RU.md`.

### Bubblewrap спрашивает JDK или regenerate

Симптомы:

```text
Do you want Bubblewrap to install the JDK?
No checksum file was found ... would you like to regenerate your project?
```

Причина: Bubblewrap запущен интерактивно.

Исправление уже в `scripts/build-twa-apks.sh`:

- заранее пишется `${HOME}/.bubblewrap/config.json`;
- перед build выполняется `bubblewrap update --skipVersionUpgrade`.

### Bubblewrap `EISDIR`

Симптом:

```text
cli ERROR EISDIR: illegal operation on a directory, read
```

Причина: Bubblewrap получил директорию вместо manifest file.

Исправление уже в `scripts/build-twa-apks.sh`: `--manifest="twa-manifest.json"` внутри `dist/twa-<app>/`.

### 404 на `/app/manifest.webmanifest` или `/app/icons/*.png`

Симптом:

```text
Failed to download Web Manifest ... 404
Failed to download icon ... 404
```

Причина: prod ещё не выкатил новый image, запрос попадает в старый pod, или pod падает до отдачи новых routes/assets.

Проверки:

```bash
kubectl get pods -n english -l app=english
kubectl logs -n english deploy/english --tail=200
kubectl get pods -n spanish -l app=spanish
kubectl logs -n spanish deploy/spanish --tail=200
```

В одном из инцидентов новый pod падал из-за порядка Postgres migration: индекс по `client_attempt_id` создавался до добавления колонки. Это исправлено отдельным релизом, но при повторении симптома сначала смотреть startup logs.

### APK показывает верхнюю browser/domain панель

Причины:

- Digital Asset Links не verified;
- `WEBAPP_ANDROID_CERT_FINGERPRINTS` не совпадает с сертификатом APK;
- домен отдаёт старый/битый `assetlinks.json`;
- Android использует fallback вместо verified TWA.

Проверки:

```bash
curl -sS https://qantrix.ru/.well-known/assetlinks.json | jq .
curl -sS https://es.qantrix.ru/.well-known/assetlinks.json | jq .
```

В актуальном APK fallback strategy выставлен в `webview`, чтобы при сбое verified TWA не показывать Chrome Custom Tab с доменной панелью.

### Чёрный экран после offline navigation

Причина, которую уже ловили: direct offline route открывался, но lazy JS chunk для страницы не был в CacheStorage.

Исправление:

- `npm run build` генерирует `/app/asset-manifest.json`;
- `/sw.js` precache-ит все assets из этого manifest;
- grammar navigation работает через Vue Router без forced full reload.

Проверка:

```bash
curl -fsS https://es.qantrix.ru/app/asset-manifest.json | jq '.assets | length'
```

После выката пользователь должен один раз открыть приложение онлайн, чтобы обновился service worker и asset cache.

### Offline результаты не синхронизируются

Ожидаемое поведение:

- если нет сети, попытки остаются в IndexedDB;
- если auth истёк, попытки остаются в IndexedDB до повторного логина;
- sync повторяется по `online` event, раз в 30 секунд и вручную кнопкой `Sync results`.

Если после логина и сети sync не уходит, смотреть API:

```http
POST /api/learning/grammar/offline/sync-attempts
POST /api/learning/grammar/offline/sync-training-attempts
POST /api/training/offline/sync-attempts
```

Повторы безопасны из-за unique `(user_id, client_attempt_id)`.
