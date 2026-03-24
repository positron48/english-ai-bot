# План подготовки universal language-learning сервиса и запуска Spanish-инстанса

## 0. Цель и рамки

Цель: на базе текущего `english-ai-bot` сделать единую кодовую базу для языковых пар (минимум RU->EN и RU->ES), где различия задаются через env + данные БД, без форка репозитория.

В этой версии плана заложен безопасный путь в 2 шага:
1. Сначала сделать код language-agnostic и сохранить текущий English prod без регрессий.
2. Затем поднять отдельный Spanish сервис в k3s (отдельный namespace/домен/БД), но на той же кодовой базе и том же CI/CD пайплайне.

Отдельный быстрый инфраструктурный чеклист к этапу запуска:
- [`SPANISH_K3S_ROLLOUT_CHECKLIST.md`](/Users/antonfilatov/www/my/k3s/english-ai-bot/docs/SPANISH_K3S_ROLLOUT_CHECKLIST.md)
- Пошаговый runbook (push репозиториев, tag, секреты, Flux): [`SPANISH_K3S_ROLLOUT_RUNBOOK.md`](/Users/antonfilatov/www/my/k3s/english-ai-bot/docs/SPANISH_K3S_ROLLOUT_RUNBOOK.md)

## 1. Текущее состояние (что уже есть)

- БД создаётся кодом, без внешнего мигратора: [`internal/database/postgres_migrate.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/database/postgres_migrate.go).
- В схеме и моделях есть жёсткая привязка к EN/RU полям:
  - `word_cards.definition_ru`, `word_cards.display_en`
  - `training_cards.word_en`, `training_cards.word_ru`, `training_cards.meaning_en`, `training_cards.example_en`, `training_cards.example_ru`, `training_cards.distractors_en`, `training_cards.distractors_ru`
  - `tts_generation_status.word` как PK без языкового измерения.
- Логика word/training ориентирована на англ. леммы и RU<->EN направления: [`internal/models/word.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/word.go), [`internal/models/training.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/training.go), [`internal/repository/word_repository.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/repository/word_repository.go), [`internal/repository/training_card_repository.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/repository/training_card_repository.go).
- Word sets уже управляются через админку/API и текстовый список слов (не hardcoded seed): [`internal/web/admin_word_sets.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/web/admin_word_sets.go), [`internal/service/word_set_service.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/service/word_set_service.go).
- Грамматика универсальна на уровне контента (`ui_language`, `target_language`); встроенные bundle лежат по языкам: [`internal/grammarbundle/en/`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/grammarbundle/en/), [`internal/grammarbundle/es/`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/grammarbundle/es/); выбор через `GRAMMAR_BUNDLE_ID` / [`GrammarContentRepository`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/repository/grammar_content_repository.go).
- В k3s уже есть рабочий паттерн для `english`: [`devops-time-host/apps/english/base/deployment.yaml`](/Users/antonfilatov/www/my/k3s/devops-time-host/apps/english/base/deployment.yaml), [`devops-time-host/clusters/prod/infra/image-automation/english-image.yaml`](/Users/antonfilatov/www/my/k3s/devops-time-host/clusters/prod/infra/image-automation/english-image.yaml).

## 2. Архитектурное решение для multi-language

Рекомендуемый целевой контракт конфигурации:

- `LEARNING_TARGET_LANG=es` (изучаемый язык)
- `LEARNING_NATIVE_LANG=ru` (язык объяснений/переводов)
- `LEARNING_PAIR=ru-es` (служебный идентификатор пары)
- `GRAMMAR_BUNDLE_DIR` или `GRAMMAR_BUNDLE_ID` (какой bundle подключить)
- `TTS_DICTIONARY_BASE_URL` под target language (для ES другой endpoint/провайдер)
- `AI_PROMPT_FILE` и `TRAINING_PROMPT_FILE` профильно-зависимые

Принцип:
- Код не знает «английский/испанский» напрямую.
- Код работает с `native_lang` и `target_lang` из runtime config.
- Изоляция данных обеспечивается отдельной БД на каждый сервис.

## 3. Модель данных при отдельной БД на сервис

### 3.1 Базовый подход

1. Не расширяем таблицы дополнительным tenant-измерением, так как сервисы разделены по БД.
2. Оставляем `direction` как рабочий механизм направления карточек.
3. Нейтрализуем терминологию через DTO/mapper, без обязательного SQL rename в первом этапе.

### 3.2 Нормализация схемы (после стабилизации)

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

Базовая схема по-прежнему поднимается в `migratePostgres()` (idempotent `CREATE TABLE IF NOT EXISTS`); поверх этого включён versioned SQL migration runner (таблица `schema_migrations`, файлы `internal/database/migrations/*.sql`). Дальнейшие DDL — отдельными пронумерованными SQL-файлами.

### 4.1 Шаги внедрения мигратора

1. Добавить таблицу `schema_migrations`.
2. Ввести SQL-файлы миграций (`internal/database/migrations/*.sql`) и runner.
3. `migratePostgres()` оставить как bootstrap только для новых пустых БД (или удалить после стабилизации).

### 4.2 Первая migration wave

1. Ввести `schema_migrations`.
2. Зафиксировать baseline текущей схемы.
3. Добавлять только миграции, нужные для language-agnostic поведения в архитектуре с отдельной БД:
- индексы и ограничения целостности;
- техдолг по качеству данных;
- точечные нейтральные поля при необходимости.

### 4.3 Обратная совместимость кода

- Для внешнего API поддерживать legacy-поля и нейтральные alias параллельно на переходном этапе.

## 5. Стратегия изоляции данных: отдельная БД vs shared schema

### Рекомендация для первого релиза Spanish

Использовать **отдельную БД и отдельный namespace/service** для Spanish.

Почему:
- минимум риска смешивания пользовательских данных;
- проще rollback;
- проще операционно (секреты, backup, SLO).

## 6. Кодовые доработки (декомпозиция по этапам)

### Этап A. Конфигурация языковой пары

1. Добавить в [`internal/config/config.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/config/config.go) секцию `LearningConfig`:
- `Pair`, `NativeLang`, `TargetLang`, `AppCode`, `GrammarBundleID`.

2. Добавить в `env.example` новые переменные и значения для RU->EN как дефолт.

3. В startup (`cmd/bot/main.go`) валидировать `LEARNING_PAIR` и передавать пару в сервисы.

4. Покрыть изменения тестами (конфиг/валидация старта) и прогнать `make check`.

Готовность:
- приложение стартует при корректной языковой паре;
- при некорректной паре завершает старт с понятной ошибкой;
- `make check` проходит полностью.

### Этап B. Репозитории и модели

1. Добавить нейтральные alias-поля на уровне моделей/DTO и мэпперов:
- [`internal/models/word.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/word.go)
- [`internal/models/training.go`](/Users/antonfilatov/www/my/k3s/english-ai-bot/internal/models/training.go)
- `internal/web/*` (DTO/response mapping).

2. Обновить SQL в репозиториях:
- `word_repository`, `training_card_repository`, `word_set_*`, `grammar_*`, `tts_status_repository`.

3. Проверить SQL-запросы на корректную работу в рамках одной БД сервиса.

4. Добавить/обновить unit + integration тесты для репозиториев и прогнать `make check`.

Готовность:
- тесты репозиториев проходят;
- данные English и Spanish изолированы на уровне отдельных БД;
- `make check` проходит полностью.

### Этап C. Сервисный слой и SRS

1. Протащить `LearningConfig` через сервисы:
- `WordService`, `TrainingService`, `WordSetService`, `GrammarService`, `PronunciationService`.

2. Сохранить текущую SRS механику без изменения алгоритма; меняется только scope данных.

3. Для TTS-кеша оставить изоляцию на уровне отдельного PVC/namespace сервиса.

4. Добавить/обновить unit + integration + regression тесты сервисного слоя и прогнать `make check`.

Готовность:
- SRS поведение на English не меняется;
- Spanish не видит английский кэш/карточки;
- `make check` проходит полностью.

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

4. Добавить тесты парсинга/валидации ответов LLM для RU->EN и RU->ES, затем прогнать `make check`.

Промпты `prompts/teacher-*.txt` и `prompts/training-card-*.txt` задают **нейтральные** ключи в JSON (`definition_native`, `example_target`, `gloss_native`, `distractors_target` и т.д.); сервер по-прежнему принимает старые имена (`definition_ru`, `example_en`, `distractors_en`) как alias.

Готовность:
- один и тот же движок генерирует карточки и для EN, и для ES;
- `make check` проходит полностью.

### Этап E. Grammar bundle routing

1. Ввести профильные bundle директории:
- `internal/grammarbundle/en/...`
- `internal/grammarbundle/es/...`

2. В `GrammarContentRepository` добавить загрузку bundle по `GRAMMAR_BUNDLE_ID`/config.

3. Создать пустой/частичный испанский курс с тем же API-форматом chapter schema.

4. Добавить тесты выбора bundle/контента и прогнать `make check`.

Готовность:
- English продолжает читать старый bundle;
- Spanish читает свой bundle (пусть даже сначала неполный);
- `make check` проходит полностью.

### Этап F. API/UI контракт

1. Во внутренних DTO перейти на neutral naming.
2. В webapp оставить обратную совместимость на период миграции.
3. В отображении названий и подсказок использовать `target_lang` (не hardcoded English).
4. Добавить/обновить unit + e2e/web regression тесты и прогнать `make check`.

Готовность:
- API сохраняет backward compatibility;
- UI корректно отображает целевой язык;
- `make check` проходит полностью.

## 7. Загрузка испанских наборов слов и грамматики

### 7.1 Грамматика

1. Конвейер авторинга: [`courses/spanish-grammar/`](/Users/antonfilatov/www/my/k3s/english-ai-bot/courses/spanish-grammar/) (Makefile, `scripts/`, `chapters/`, `config/generation-status.json`) — зеркально english-grammar. Сборка во встроенный bundle: из корня репо `./scripts/generate-grammar-bundle.sh es` или `... all` (обновляет `internal/grammarbundle/es/`).
2. Сохранить schema-формат chapter JSON (общий `02-chapter-schema.json`).
3. Публикацию разделов/глав — через существующий админ publish-механизм.
4. Расширять контент: новые главы в `courses/spanish-grammar/chapters/`, затем `generate-grammar-bundle.sh es`.

## 8. CI/CD и репозиторий (одна кодовая база)

Текущий workflow: [`english-ai-bot/.github/workflows/ci.yml`](/Users/antonfilatov/www/my/k3s/english-ai-bot/.github/workflows/ci.yml).

В репозитории `english-ai-bot`: job сборки Docker-образа в [`.github/workflows/ci.yml`](/Users/antonfilatov/www/my/k3s/english-ai-bot/.github/workflows/ci.yml) использует **matrix** `english` / `spanish` (один Dockerfile, два имени образа в GHCR на push tag).

Дальше по инфраструктуре:
1. Для Spanish использовать тот же образ, но разные runtime env в k8s.

2. Для Flux image automation в `devops-time-host` добавить отдельные ресурсы `spanish-image.yaml` (см. §9.2).

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
- `LEARNING_NATIVE_LANG=ru`
- `LEARNING_TARGET_LANG=es`
- `LEARNING_PAIR=ru-es`
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
- интеграционные тесты для изоляции EN и ES через отдельные БД/окружения.
- regression тесты SRS на English (идентичные метрики до/после).
- contract тесты LLM JSON parser для RU->ES.
- e2e smoke в отдельном namespace.

## 11. Порядок внедрения (рекомендуемый roadmap)

1. Ветка 1: `feature/multilang-config-and-pair`.
2. Ветка 2: `feature/domain-dto-neutralization`.
3. Ветка 3: `feature/prompts-ru-es-and-grammar-routing`.
4. Ветка 4: `feature/devops-spanish-app`.
5. Staging rollout Spanish.
6. Prod rollout Spanish (canary 1 день, потом full).

## 12. Риски и меры

- Риск: смешивание EN/ES данных.
  - Мера: отдельные БД/namespace/secrets + integration tests на изоляцию окружений.

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

Состояние на уровне **кода в `english-ai-bot`** (без k3s): пункты 1–4, 6–7 ниже закрыты в репозитории; пункт 5 (отдельный import pipeline для испанских word sets) и расширение контента §7 — по необходимости продуктом.

Остаётся **инфраструктура** (обычно `devops-time-host`): п. 8–10, плюс Flux/секреты из §8–9.

1. Добавить `LearningConfig` и pair resolution. — **сделано**
2. Внедрить SQL migration runner (baseline + `migrations/*.sql`). — **сделано**
3. Протащить `LearningConfig` по репозиториям/сервисам. — **сделано**
4. Вынести prompt templates под RU->ES (и нейтральные JSON-ключи). — **сделано**
5. Добавить Spanish word sets import pipeline (при необходимости массового импорта).
6. Подготовить Spanish grammar bundle routing (`internal/grammarbundle/es`). — **сделано** (MVP; полный конвейер `courses/spanish-grammar` — см. §7)
7. Настроить CI image targets (`english` + `spanish`). — **сделано** (matrix в CI репозитория)
8. Создать `apps/spanish` + `spanish-image.yaml` в GitOps.
9. Применить k8s secrets для Spanish.
10. Выполнить rollout + smoke + наблюдение.
