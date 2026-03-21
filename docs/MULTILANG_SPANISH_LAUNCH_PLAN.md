# План подготовки universal language-learning сервиса и запуска Spanish-инстанса

## 0. Цель и рамки

Цель: на базе текущего `english-ai-bot` сделать единую кодовую базу для языковых пар (минимум RU->EN и RU->ES), где различия задаются через env + данные БД, без форка репозитория.

В этой версии плана заложен безопасный путь в 2 шага:
1. Сначала сделать код language-agnostic и сохранить текущий English prod без регрессий.
2. Затем поднять отдельный Spanish сервис в k3s (отдельный namespace/домен/БД), но на той же кодовой базе и том же CI/CD пайплайне.

Отдельный быстрый инфраструктурный чеклист к этапу запуска:
- [`SPANISH_K3S_ROLLOUT_CHECKLIST.md`](/Users/antonfilatov/www/my/k3s/english-ai-bot/docs/SPANISH_K3S_ROLLOUT_CHECKLIST.md)

## 1. Текущее состояние (что уже есть)

- БД создаётся кодом, без внешнего мигратора: [`internal/database/postgres_migrate.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/database/postgres_migrate.go).
- В схеме и моделях есть жёсткая привязка к EN/RU полям:
  - `word_cards.definition_ru`, `word_cards.display_en`
  - `training_cards.word_en`, `training_cards.word_ru`, `training_cards.meaning_en`, `training_cards.example_en`, `training_cards.example_ru`, `training_cards.distractors_en`, `training_cards.distractors_ru`
  - `tts_generation_status.word` как PK без языкового измерения.
- Логика word/training ориентирована на англ. леммы и RU<->EN направления: [`internal/models/word.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/word.go), [`internal/models/training.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/training.go), [`internal/repository/word_repository.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/repository/word_repository.go), [`internal/repository/training_card_repository.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/repository/training_card_repository.go).
- Word sets уже управляются через админку/API и текстовый список слов (не hardcoded seed): [`internal/web/admin_word_sets.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/web/admin_word_sets.go), [`internal/service/word_set_service.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/service/word_set_service.go).
- Грамматика уже частично универсальна на уровне контента (`ui_language`, `target_language`), но bundle в проекте сейчас английский: [`internal/repository/grammar_content_repository.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/repository/grammar_content_repository.go), [`internal/grammarbundle/index.json`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/grammarbundle/index.json).
- В k3s уже есть рабочий паттерн для `english`: [`devops-time-host/apps/english/base/deployment.yaml`](/Users/antonfilatov/www/my/k3s/devops-time-host/apps/english/base/deployment.yaml), [`devops-time-host/clusters/prod/infra/image-automation/english-image.yaml`](/Users/antonfilatov/www/my/k3s/devops-time-host/clusters/prod/infra/image-automation/english-image.yaml).

## 2. Архитектурное решение для multi-language

Рекомендуемый целевой контракт конфигурации:

- `LEARNING_TARGET_LANG=es` (изучаемый язык)
- `LEARNING_NATIVE_LANG=ru` (язык объяснений/переводов)
- `LEARNING_PAIR=ru-es` (служебный идентификатор пары)
- `LEARNING_PROFILE=spanish` (ID профиля/инстанса)
- `GRAMMAR_BUNDLE_DIR` или `GRAMMAR_BUNDLE_ID` (какой bundle подключить)
- `TTS_DICTIONARY_BASE_URL` под target language (для ES другой endpoint/провайдер)
- `AI_PROMPT_FILE` и `TRAINING_PROMPT_FILE` профильно-зависимые

Принцип:
- Код не знает «английский/испанский» напрямую.
- Код работает с `native_lang` и `target_lang` из runtime config.
- Данные и уникальные ключи учитывают `language_profile_id`.

## 3. Модель данных: где нужен language dimension

### 3.1 Таблицы, где language dimension обязателен

Добавить `language_profile_id` (FK) в:
- `word_cards`
- `word_forms`
- `word_request_history`
- `training_cards`
- `word_set_categories`
- `word_sets`
- `word_set_items`
- `grammar_published_items`
- `grammar_progress`
- `grammar_test_attempts`
- `grammar_placement_test`
- `tts_generation_status`

### 3.2 Таблицы, где можно оставить без language dimension

Если Spanish будет отдельным инстансом/ботом (рекомендуется на старте):
- `users`, `web_sessions`, `web_otps`, `training_sessions`, `review_events`, `training_nudges`, `circuit_breaker_state`, `app_settings`, `user_access_*`.

Если в будущем делать shared schema для нескольких языков в одном инстансе, тогда:
- `app_settings` сделать profile-scoped (`(language_profile_id, key)` unique).
- `grammar_placement_test` менять с `UNIQUE(user_id)` на `UNIQUE(user_id, language_profile_id)`.

### 3.3 Нормализация схемы (после стабилизации)

Текущие EN/RU колонки оставить на переходный период, но в коде перейти на нейтральные alias:
- `word_en` -> `word_target`
- `word_ru` -> `word_native`
- `meaning_en` -> `meaning_target`
- `example_en` -> `example_target`
- `example_ru` -> `example_native`
- `definition_ru` -> `definition_native`
- `display_en` -> `display_target`

Физическое переименование колонок делать отдельной поздней миграцией после запуска Spanish.

## 4. Стратегия миграций (backward-compatible)

Так как сейчас миграции в `migratePostgres()` (idempotent `CREATE TABLE IF NOT EXISTS`), нужно перейти к versioned SQL migration runner.

### 4.1 Шаги внедрения мигратора

1. Добавить таблицу `schema_migrations`.
2. Ввести SQL-файлы миграций (`internal/database/migrations/*.sql`) и runner.
3. `migratePostgres()` оставить как bootstrap только для новых пустых БД (или удалить после стабилизации).

### 4.2 Первая миграция multi-language

1. Создать таблицу:
- `language_profiles(id, code, name, native_lang, target_lang, is_default, created_at)`

2. Вставить дефолтный профиль для текущего English:
- `code='english-ru-en'`, `native_lang='ru'`, `target_lang='en'`, `is_default=true`.

3. Добавить nullable `language_profile_id` в таблицы из раздела 3.1.

4. Backfill существующих данных:
- `UPDATE ... SET language_profile_id = <default_profile_id> WHERE language_profile_id IS NULL`.

5. Добавить NOT NULL + индексы.

6. Пересобрать уникальности:
- `word_cards`: `UNIQUE(language_profile_id, lower(word))`
- `word_forms`: `UNIQUE(language_profile_id, lower(form))`
- `training_cards`: `UNIQUE(language_profile_id, word_card_id, sense_index)`
- `word_sets`: title uniqueness (если нужна) тоже scope по profile
- `tts_generation_status`: заменить PK `word` на surrogate `id` + `UNIQUE(language_profile_id, word)`.

7. Добавить FK `language_profile_id -> language_profiles(id)`.

### 4.3 Обратная совместимость кода

- На переходе все repository-методы получают `languageProfileID` параметр.
- Внешние handler/service получают profile из config (single-profile режим).
- Для старых мест временно оставить wrapper-методы без profile, внутри подставлять default profile.

## 5. Стратегия изоляции данных: отдельная БД vs shared schema

### Рекомендация для первого релиза Spanish

Использовать **отдельную БД и отдельный namespace/service** для Spanish.

Почему:
- минимум риска смешивания пользовательских данных;
- проще rollback;
- проще операционно (секреты, backup, SLO).

### Когда переходить к shared schema

Только если появится явная бизнес-потребность в одном инстансе с множеством языков и общими аккаунтами/аналитикой.

## 6. Кодовые доработки (декомпозиция по этапам)

### Этап A. Конфигурация языкового профиля

1. Добавить в [`internal/config/config.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/config/config.go) секцию `LearningConfig`:
- `Profile`, `Pair`, `NativeLang`, `TargetLang`, `LanguageProfileCode`.

2. Добавить в `env.example` новые переменные и значения для RU->EN как дефолт.

3. В startup (`cmd/bot/main.go`) загрузить/резолвить активный `language_profile_id`.

Готовность:
- приложение стартует при явном profile;
- при отсутствии profile использует default из БД.

### Этап B. Репозитории и модели

1. Добавить `LanguageProfileID` в модели:
- [`internal/models/word.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/word.go)
- [`internal/models/training.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/training.go)
- [`internal/models/word_set.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/word_set.go)
- модели grammar progress/attempt.

2. Обновить SQL в репозиториях:
- `word_repository`, `training_card_repository`, `word_set_*`, `grammar_*`, `tts_status_repository`.

3. Везде в SELECT/INSERT/UPDATE добавить фильтрацию по `language_profile_id`.

Готовность:
- тесты репозиториев проходят;
- невозможно прочитать чужой профиль без явного profile_id.

### Этап C. Сервисный слой и SRS

1. Протащить `languageProfileID` через сервисы:
- `WordService`, `TrainingService`, `WordSetService`, `GrammarService`, `PronunciationService`.

2. Сохранить текущую SRS механику без изменения алгоритма; меняется только scope данных.

3. Для TTS-кеша в БД и файловой системе добавить профилирование:
- БД: `language_profile_id`;
- файлы: подпапка `/app/data/tts/<profile>/...`.

Готовность:
- SRS поведение на English не меняется;
- Spanish не видит английский кэш/карточки.

### Этап D. AI prompt abstraction

1. Вынести prompt-переменные в шаблонные placeholders:
- `{{native_lang}}`, `{{target_lang}}`, `{{pair}}`.

2. Подготовить отдельные prompt files:
- `prompts/teacher-ru-en.txt`
- `prompts/teacher-ru-es.txt`
- `prompts/training-card-ru-en.txt`
- `prompts/training-card-ru-es.txt`

3. Пересмотреть JSON-контракты LLM:
- логические ключи нейтральные (`word_target`, `word_native`) в коде;
- на wire можно временно поддерживать старые alias для совместимости.

Готовность:
- один и тот же движок генерирует карточки и для EN, и для ES.

### Этап E. Grammar bundle routing

1. Ввести профильные bundle директории:
- `internal/grammarbundle/en/...`
- `internal/grammarbundle/es/...`

2. В `GrammarContentRepository` добавить загрузку bundle по profile/config.

3. Создать пустой/частичный испанский курс с тем же API-форматом chapter schema.

Готовность:
- English продолжает читать старый bundle;
- Spanish читает свой bundle (пусть даже сначала неполный).

### Этап F. API/UI контракт

1. Во внутренних DTO перейти на neutral naming.
2. В webapp оставить обратную совместимость на период миграции.
3. В отображении названий и подсказок использовать `target_lang` (не hardcoded English).

## 7. Загрузка испанских наборов слов и грамматики

### 7.1 Наборы слов (top lists)

Рекомендуемая структура источников в git:
- `data/wordsets/es/nouns_top_2000.txt`
- `data/wordsets/es/verbs_top_500.txt`
- `data/wordsets/es/adjectives_top_500.txt`
- `data/wordsets/es/adverbs_top_500.txt`
- `data/wordsets/es/pronouns_top_500.txt`

Добавить CLI импортер (`cmd/import_word_sets`) с параметрами:
- `--profile spanish`
- `--category "Nouns"`
- `--set "Top 2000 Nouns"`
- `--file data/wordsets/es/nouns_top_2000.txt`
- `--preferred-pos noun`

Импортер должен:
1. создать/обновить category;
2. создать/обновить word set;
3. загрузить слова через `ProcessWordSetItems`;
4. опционально запустить prewarm training cards.

### 7.2 Грамматика

1. Создать `courses/spanish-grammar` по аналогии с `courses/english-grammar`.
2. Сохранить schema-формат chapter JSON.
3. Добавить скрипт сборки bundle для `es`.
4. Публикацию разделов/глав оставить через существующий админ publish-механизм.

## 8. CI/CD и репозиторий (одна кодовая база)

Текущий workflow: [`english-ai-bot/.github/workflows/ci.yml`](/Users/antonfilatov/www/my/k3s/english-ai-bot/.github/workflows/ci.yml).

Что сделать:
1. Перейти с жесткого `IMAGE_NAME=.../english` на matrix/parameterized image target:
- `ghcr.io/<owner>/english`
- `ghcr.io/<owner>/spanish`

2. Для Spanish использовать тот же Dockerfile, но разные runtime env в k8s.

3. Для Flux image automation добавить отдельные ресурсы `spanish-image.yaml`.

## 9. Пошаговый план запуска Spanish в k3s

## 9.1 Подготовка GitOps манифестов

В `devops-time-host`:
1. Скопировать `apps/english` -> `apps/spanish`.
2. Переименовать ресурсы/namespace:
- `english` -> `spanish`
- `english-config` -> `spanish-config`
- `english-secrets` -> `spanish-secrets`
- `english-postgres` -> `spanish-postgres`
- PVC имена с префиксом `spanish-*`
3. В `apps/spanish/base/configmap.yaml` задать:
- `WEBAPP_PUBLIC_URL=https://<spanish-domain>`
- `LEARNING_PROFILE=spanish`
- `LEARNING_NATIVE_LANG=ru`
- `LEARNING_TARGET_LANG=es`
- `AI_PROMPT_FILE=prompts/teacher-ru-es.txt`
- `TRAINING_PROMPT_FILE=prompts/training-card-ru-es.txt`
- `TTS_DICTIONARY_BASE_URL` для испанского словаря/источника.
4. В `apps/spanish/base/ingress.yaml` поставить новый host и TLS secret.
5. Добавить `../../apps/spanish/prod` в [`devops-time-host/clusters/prod/kustomization.yaml`](/Users/antonfilatov/www/my/k3s/devops-time-host/clusters/prod/kustomization.yaml).

## 9.2 Flux image automation

1. Создать `clusters/prod/infra/image-automation/spanish-image.yaml` (аналог english).
2. Добавить его в kustomization image-automation.
3. В `apps/spanish/base/deployment.yaml` указать image policy comment для `spanish`.

## 9.3 Секреты на сервере

```bash
kubectl create namespace spanish --dry-run=client -o yaml | kubectl apply -f -

kubectl -n spanish create secret generic spanish-postgres \
  --from-literal=POSTGRES_DB='spanish' \
  --from-literal=POSTGRES_USER='spanish' \
  --from-literal=POSTGRES_PASSWORD='<SPANISH_POSTGRES_PASSWORD>' \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n spanish create secret generic spanish-secrets \
  --from-literal=DATABASE_URL='postgres://spanish:<SPANISH_POSTGRES_PASSWORD>@spanish-postgres:5432/spanish?sslmode=disable' \
  --from-literal=AI_API_KEY='<OPENROUTER_API_KEY>' \
  --from-literal=WEBAPP_JWT_SECRET='<WEBAPP_JWT_SECRET>' \
  --from-literal=TELEGRAM_TOKEN='<SPANISH_TELEGRAM_BOT_TOKEN>' \
  --from-literal=TTS_API_KEY='<OPENROUTER_API_KEY>' \
  --dry-run=client -o yaml | kubectl apply -f -
```

Если GHCR приватный:
```bash
kubectl -n spanish create secret docker-registry ghcr-creds \
  --docker-server=ghcr.io \
  --docker-username='<GITHUB_USERNAME>' \
  --docker-password='<GITHUB_PAT_WITH_read:packages>' \
  --docker-email='<EMAIL>'
```

## 9.4 Rollout и проверка

```bash
flux reconcile image repository spanish -n flux-system
flux reconcile image policy spanish -n flux-system
flux reconcile image update flux-system -n flux-system
flux reconcile kustomization flux-system -n flux-system --with-source

kubectl get pods -n spanish
kubectl logs -n spanish deploy/spanish --tail=200
kubectl get ingress -n spanish
```

Smoke checks:
1. `/health` OK
2. Telegram `/start` и `/help`
3. Запрос испанского слова -> карточка
4. Создание и прохождение тренировки
5. Пуш reminder только при неактивности
6. Miniapp: логин, словарь, тренировки, темы, переключение языка UI

## 10. Тест-план перед прод запуском

Обязательно:
- `make check` в `english-ai-bot`.
- интеграционные тесты для profile-scoped SQL (изоляция EN и ES).
- regression тесты SRS на English (идентичные метрики до/после).
- contract тесты LLM JSON parser для RU->ES.
- e2e smoke в отдельном namespace.

## 11. Порядок внедрения (рекомендуемый roadmap)

1. Ветка 1: `feature/multilang-config-and-profile`.
2. Ветка 2: `feature/db-profile-migrations`.
3. Ветка 3: `feature/repository-profile-scope`.
4. Ветка 4: `feature/prompts-ru-es-and-grammar-routing`.
5. Ветка 5: `feature/devops-spanish-app`.
6. Staging rollout Spanish.
7. Prod rollout Spanish (canary 1 день, потом full).

## 12. Риски и меры

- Риск: смешивание EN/ES данных.
  - Мера: обязательный `language_profile_id` + composite uniques + integration tests.

- Риск: просадка качества карточек для ES.
  - Мера: отдельные prompt templates и golden tests на 100-200 слов.

- Риск: TTS source для ES отличается от EN dictionary API.
  - Мера: абстракция dictionary provider по target language + fallback на OpenRouter TTS.

- Риск: CI сейчас пушит image только по tag.
  - Мера: синхронизировать политику с Flux (latest digest strategy) и обновить workflow.

## 13. Открытые вопросы перед реализацией

1. Spanish будет ориентирован только на RU-native пользователей (как сейчас), или сразу нужен EN-native UI тоже?
2. Требуется ли единый аккаунт пользователя между English и Spanish, или это два полностью независимых сервиса?
3. Для Spanish grammar на первом релизе достаточно MVP (A1-A2), или нужен минимум до B1/B2?
4. Какой источник считать canonical для испанского произношения (словарь vs только TTS)?

## 14. Краткий executable checklist

1. Добавить `LearningConfig` и profile resolution.
2. Внедрить SQL migration runner + `language_profiles`.
3. Протащить `language_profile_id` по репозиториям/сервисам.
4. Вынести prompt templates под RU->ES.
5. Добавить Spanish word sets import pipeline.
6. Подготовить Spanish grammar bundle routing.
7. Настроить CI image targets (`english` + `spanish`).
8. Создать `apps/spanish` + `spanish-image.yaml` в GitOps.
9. Применить k8s secrets для Spanish.
10. Выполнить rollout + smoke + наблюдение.
