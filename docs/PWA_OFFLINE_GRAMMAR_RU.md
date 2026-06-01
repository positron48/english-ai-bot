# PWA/APK: офлайн-режим грамматики

Документ описывает фактически реализованную модель Android PWA/TWA APK и офлайн-грамматики для English/Spanish инстансов `english-ai-bot`.

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

Не работают офлайн в текущей версии:
- словарь и word sets;
- обычные training cards;
- chat/LLM;
- reading;
- speaking;
- админка;
- любые серверные операции вне offline grammar API.

Важно: в UI есть два похожих понятия. “Тесты по грамматике” (chapter/category/placement tests) и “Grammar Training” как отдельный режим повторения/SRS теперь оба работают офлайн после preload, но синхронизируются разными очередями.

## Как пользователь получает APK

GitHub tag release в `english-ai-bot` собирает два TWA APK:

- `qantrix-english-<tag>.apk`: Android package `ru.qantrix.english`, открывает `https://qantrix.ru/app/`;
- `qantrix-spanish-<tag>.apk`: Android package `ru.qantrix.spanish`, открывает `https://es.qantrix.ru/app/`.

APK не содержит курс внутри себя. В APK фактически зашиты package id, URL сервера, TWA metadata и подпись. Сам web shell, service worker и курс приходят с выбранного сервера.

Для verified TWA сервер должен отдавать `/.well-known/assetlinks.json`. SHA-256 fingerprint релизного Android signing certificate задаётся в GitOps через `WEBAPP_ANDROID_CERT_FINGERPRINTS` в соответствующем ConfigMap English/Spanish.

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

Service worker отдельно кэширует только web shell (`/app`, `/app/`, static assets, manifest), чтобы приложение могло открыться без сети. Контент курса хранится не в CacheStorage, а в IndexedDB.

## Что видно на фронте

На странице грамматики есть блок `Offline grammar`:

- показывает `Online` или `Offline` badge по `navigator.onLine`;
- показывает, готова ли предзагрузка (`Ready: X/Y chapters`);
- показывает количество результатов, ожидающих синхронизации;
- даёт кнопки `Preload grammar` / `Update preload`, `Sync results`, `Delete`.

Если пользователь офлайн и пытается открыть режимы вне поддержанного набора grammar offline, router перенаправляет его обратно в грамматику. В текущей минимальной реализации нет отдельного глобального баннера на всех экранах; явный статус сети находится в блоке грамматики.

## Как работают офлайн-тесты

Когда приложение офлайн (`navigator.onLine === false`) или отправка результата упала сетевой ошибкой:

1. `grammarClient` читает нужные главы из IndexedDB.
2. Для chapter test выбирает вопросы из локального `chapter_test.pool_question_ids` / `question_bank`.
3. Для category test собирает вопросы из опубликованных глав категории.
4. Для placement test выбирает вопросы из локально сохранённых глав.
5. Ответы проверяются локально той же базовой логикой сравнения: строки нормализуются, массивы сравниваются без учёта порядка, true/false поддерживает `true/false`, `да/нет`, `yes/no`, `1/0`.
6. Результат сразу показывается пользователю и обновляет локальный progress snapshot.
7. Попытка сохраняется в IndexedDB queue с `client_attempt_id`.

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
