# English Bot

Telegram бот для изучения английского языка с AI-ассистентом и системой тренировок на основе SRS (Spaced Repetition System).

## Быстрый старт

```bash
# 1. Скачать бинарник
curl -L -o universal-ai-bot https://github.com/positron48/universal-ai-bot/releases/latest/download/universal-ai-bot-linux_amd64
chmod +x universal-ai-bot

# 2. Создать .env
cat > .env << 'EOF'
TELEGRAM_TOKEN=your_bot_token_here
AI_URL=https://openrouter.ai/api/v1
AI_API_KEY=your_openrouter_api_key
AI_MODEL=qwen/qwen3-coder:free
AI_PROMPT=You are a helpful AI assistant.
DATABASE_PATH=./data/words.db
WEBAPP_JWT_SECRET=$(openssl rand -hex 32)
TRAINING_WORKER_ENABLED=true
TRAINING_PROMPT_FILE=prompts/training-card-generator.txt
EOF

# 3. Запустить
./universal-ai-bot
```

## Архитектура

### Компоненты

- **Telegram Bot** (`internal/bot/`) - обработка сообщений и команд
- **AI Service** (`internal/ai/`) - интеграция с OpenAI-совместимыми API
- **Training System** (`internal/service/training_*.go`) - система тренировок с SRS
- **Training Worker** (`internal/service/training_worker.go`) - фоновый воркер для генерации тренировочных карточек
- **Web App** (`internal/web/`, `webapp/`) - Vue SPA с JWT аутентификацией
- **Database** (`internal/database/`, `internal/repository/`) - SQLite с репозиториями

### Сущности БД

**Основные связи:**
```
word_cards (1) ←→ (N) training_cards (1) ←→ (2) user_cards
                                              ↑
                                         (ru_en, en_ru)
```

**Слова и карточки:**
- **word_cards** - словарные карточки с определениями (создаются при запросе слова)
  - Содержит: lemma, POS, transcription, определения, примеры, формы глаголов
- **word_forms** - маппинг словоформ на word_cards (для поиска по любой форме слова)
- **word_request_history** - история запросов слов пользователями
  - Связывает пользователя, исходное слово (input_word) и созданную карточку

**Пользователи:**
- **users** - пользователи бота
  - Telegram ID, username, timezone, настройки тренировок, пользовательские настройки

**Тренировки:**
- **training_cards** - тренировочные карточки (одно значение слова, создаются воркером)
  - Содержит: слово, перевод, примеры, отвлекающие варианты, подсказки
- **user_cards** - карточки пользователя для SRS (по 2 на training_card: RU→EN и EN→RU)
  - Хранит состояние SRS: EF, интервалы, шаги обучения, статистику
- **training_sessions** - сессии тренировок
  - Отслеживает начало/конец, источник (nudge/manual), количество карточек
- **review_events** - события ответов на карточки
  - Детальный лог: время показа, варианты, выбранный ответ, качество, метрики
- **training_nudges** - уведомления о тренировках
  - Отслеживает отправленные напоминания по датам

**Веб-приложение:**
- **web_sessions** - сессии веб-приложения (JWT refresh tokens)
- **web_otps** - одноразовые коды для входа через OTP

**Наборы слов:**
- **word_set_categories** - категории наборов слов (иерархическая структура)
- **word_sets** - наборы слов для изучения
- **word_set_items** - связи слов с наборами (word_set ←→ word_card)
- **user_word_knowledge** - знания пользователя о словах (статус "known")

**Служебные:**
- **circuit_breaker_state** - состояние circuit breaker для защиты от ошибок LLM

### SRS (Spaced Repetition System)

Алгоритм SM-2 с автоматическим определением качества ответа:

- **Quality 0** (Wrong) - неправильный ответ
- **Quality 1** (Hard) - правильный, но сложный (ранний показ вариантов или медленный ответ >8с)
- **Quality 2** (Good) - правильный, нормальный
- **Quality 3** (Easy) - правильный и легкий (быстрый ответ <2.5с без раннего показа)

**Состояния карточек:**
- `new` → `learning` → `review`
- При ошибке: сброс в `learning` с уменьшением EF (Easiness Factor)

**Шаги обучения:**
- RU→EN: [1, 3, 7, 14] дней (активное воспроизведение, сложнее)
- EN→RU: [1, 3, 7] дней (пассивное распознавание, проще)

### Тренировка

**UX карточки:**
1. Показ вопроса без вариантов (задержка `TRAINING_OPTIONS_DELAY_MS`, по умолчанию 5с)
2. Показ вариантов ответа (4 варианта по умолчанию)
3. Ответ пользователя → расчет качества → обновление SRS

**Генерация очереди:**
- Максимум 30 карточек за сессию (`DefaultMaxCardsPerSession`)
- Максимум 5 новых карточек (`DefaultMaxNewPerSession`)
- Приоритет: learning → review → new

### Training Worker

Фоновый процесс, который:
1. Находит необработанные `word_cards` (где `processed_at IS NULL`)
2. Генерирует `training_cards` через LLM (по одному на каждое значение слова)
3. Создает `user_cards` для всех пользователей, запросивших это слово (по 2 направления)
4. Использует Circuit Breaker для защиты от ошибок LLM

**Circuit Breaker:**
- Открывается при `CIRCUIT_BREAKER_THRESHOLD` ошибках подряд
- Автоматически закрывается через `CIRCUIT_BREAKER_AUTO_RESET_HOURS` часов

## Миграция существующих данных

После обновления схемы БД для работы с леммами и структурированными данными, необходимо перегенерировать существующие слова:

```bash
# Собрать инструмент миграции
make build-migrate

# Запустить миграцию
./bin/migrate_words

# Или через make (просто показывает инструкцию)
make migrate-words
```

Миграция:
- Проходит по всем существующим `word_cards`
- Для каждого слова запрашивает у LLM структурированные данные (lemma, POS, transcription, etc.)
- Обновляет `word_cards` с новыми полями
- Создает маппинги в `word_forms` для всех словоформ
- Пропускает уже мигрированные слова (если есть `pos`)

**⚠️ Важно:** Сделайте резервную копию базы данных перед запуском миграции!

### Миграция тренировочных карточек

После миграции слов необходимо обновить существующие тренировочные карточки:

```bash
# Собрать инструмент миграции
make build-migrate-training

# Запустить миграцию
./bin/migrate_training_cards

# Или через make (просто показывает инструкцию)
make migrate-training-cards
```

Миграция тренировочных карточек:
- Проходит по всем существующим `training_cards` без `pos` или `display_word`
- Для каждой карточки определяет часть речи (POS) на основе её `meaning_en` и `word_ru` через LLM
- Обновляет `pos` и `display_word` для каждой карточки отдельно
- Учитывает, что разные карточки одного слова могут иметь разные части речи

**⚠️ Важно:** Сделайте резервную копию базы данных перед запуском миграции!

## Команды

### Пользовательские
- `/start` - приветствие
- `/help` - справка
- `/train` - начать тренировку
- `/get_id` - получить Telegram ID
- `/unsubscribe` - отписаться от уведомлений
- `/notification [daily|never|N]` - настроить периодичность уведомлений
  - `daily` - ежедневно
  - `never` - никогда
  - `N` - каждые N дней (например, `/notification 3`)

### Админские
- `/reset_circuit` - сброс circuit breaker
- `/delete_train [word]` - удалить тренировочные карточки слова
- `/delete_train_all` - удалить все тренировочные карточки
- `/get_train_data [word]` - данные о тренировочных карточках

## Веб-приложение

Vue 3 SPA, встроенное в Go-бинарник через `go:embed`.

### Функциональность

**Обучение:**
- **Словарь** (`/vocab`) - просмотр всех изученных слов с карточками
- **Наборы слов** (`/learning/words`) - управление наборами слов для изучения
- **Грамматика** (`/learning/grammar`) - интерактивный курс грамматики (в разработке)
- **Тренировка** (`/training`) - система тренировок с SRS

**Взаимодействие:**
- **AI Чат** (`/chat`) - общение с AI-преподавателем
- **Настройки** (`/settings`) - настройки профиля и уведомлений

**Администрирование:**
- Управление circuit breaker
- Тестирование промптов
- Управление orphaned карточками
- Управление наборами слов
- Просмотр схемы базы данных

### Маршруты
- `/app` или `/app/#/login` - вход
- `/app/#/dashboard` - дашборд
- `/app/#/vocab` - словарь
- `/app/#/learning` - раздел обучения
  - `/app/#/learning/grammar` - курс грамматики (в разработке)
  - `/app/#/learning/words` - наборы слов
  - `/app/#/learning/words/:setId` - детали набора слов
  - `/app/#/learning/words/:setId/study` - изучение набора слов
- `/app/#/training` - тренировка (SRS)
- `/app/#/chat` - AI чат
- `/app/#/settings` - настройки
- `/app/#/admin` - админ-панель
  - `/app/#/admin/circuit-breaker` - управление circuit breaker
  - `/app/#/admin/prompt-tester` - тестирование промптов
  - `/app/#/admin/orphaned-cards` - управление orphaned карточками
  - `/app/#/admin/word-sets` - управление наборами слов
  - `/app/#/admin/db-schema` - схема базы данных

### Аутентификация

**JWT токены:**
- Access token (TTL: 24h) - для API запросов
- Refresh token (TTL: 30 дней) - для обновления access token

**Методы входа:**
1. Telegram Mini App - автоматическая аутентификация через `initData`
2. OTP - для внешнего браузера (запрос кода через бота)

### API

- `/swagger/` - Swagger UI документация
- `/health` - health check
- `/auth/*` - аутентификация
- `/app/*` - защищенные эндпоинты (требуют JWT)

## Конфигурация

### Обязательные переменные

```env
TELEGRAM_TOKEN=              # Токен бота от @BotFather
AI_URL=                      # URL AI провайдера
AI_API_KEY=                  # API ключ
AI_MODEL=                    # Модель AI (основная)
AI_MODEL_HIGH=               # Модель AI для сложных задач (опционально)
AI_PROMPT=                   # Системный промпт (или AI_PROMPT_FILE)
AI_PROMPT_FILE=              # Файл с промптом (альтернатива AI_PROMPT)
DATABASE_PATH=               # Путь к SQLite БД
WEBAPP_JWT_SECRET=           # Секрет для JWT (openssl rand -hex 32)
```

**Доступные промпты:**
- `prompts/simple-assistant.txt` - базовый помощник
- `prompts/customer-support.txt` - специалист поддержки
- `prompts/english-teacher.txt` - преподаватель английского (рекомендуется)

### Тренировки

```env
TRAINING_WORKER_ENABLED=true              # Включить воркер
TRAINING_WORKER_INTERVAL=30s              # Интервал обработки
TRAINING_WORKER_BATCH_SIZE=5              # Размер батча
TRAINING_LLM_WORKERS=4                    # Количество параллельных воркеров для LLM
TRAINING_PROMPT_FILE=prompts/training-card-generator.txt
CIRCUIT_BREAKER_THRESHOLD=5              # Порог для circuit breaker
CIRCUIT_BREAKER_AUTO_RESET_HOURS=24      # Автосброс circuit breaker (часы)
TRAINING_OPTIONS_DELAY_MS=5000           # Задержка показа вариантов
TRAINING_WRONG_ANSWER_DELAY_SECONDS=5    # Задержка после ошибки
ADMIN_TELEGRAM_ID=                        # ID админа для уведомлений (опционально)
```

### Веб-приложение

```env
WEBAPP_JWT_TTL_HOURS=24                  # TTL access token
WEBAPP_REFRESH_TTL_HOURS=720             # TTL refresh token (30 дней)
WEBAPP_OTP_TTL_SECONDS=300              # TTL OTP кода (5 минут)
WEBAPP_PUBLIC_URL=https://your-domain.com  # Публичный URL (для CORS)
WEBAPP_VITE_DEV_SERVER_URL=http://localhost:5173  # URL Vite dev server (только для разработки)
```

### Rate Limiting (опционально)

```env
# Лимиты запросов (по умолчанию используются безопасные значения)
WEBAPP_RATE_LIMIT_AUTH_REQUEST_OTP_PER_IP=10
WEBAPP_RATE_LIMIT_AUTH_REQUEST_OTP_PER_IP_USER=3
WEBAPP_RATE_LIMIT_AUTH_OTP_PER_IP=20
WEBAPP_RATE_LIMIT_AUTH_OTP_PER_IP_USER=5
WEBAPP_RATE_LIMIT_AUTH_TELEGRAM_UNSAFE_PER_IP=30
WEBAPP_RATE_LIMIT_AUTH_TELEGRAM_UNSAFE_PER_IP_USER=10
WEBAPP_RATE_LIMIT_AUTH_TELEGRAM_PER_IP=60
WEBAPP_RATE_LIMIT_AUTH_REFRESH_PER_IP=60
WEBAPP_RATE_LIMIT_APP_API_PER_USER=300
WEBAPP_RATE_LIMIT_APP_CHAT_PER_USER=60
WEBAPP_RATE_LIMIT_WINDOW_MINUTES=1       # Окно для rate limiting
WEBAPP_RATE_LIMIT_BURST_MULTIPLIER=2     # Множитель для burst
```

### Сервер и логирование

```env
SERVER_ADDRESS=:8080                     # Адрес сервера
SERVER_PORT=8080                         # Порт сервера
LOG_LEVEL=info                           # Уровень логирования (debug, info, warn, error)
TELEGRAM_DEBUG=false                     # Режим отладки Telegram API
TELEGRAM_UPDATES_TIMEOUT=30              # Таймаут обновлений Telegram
```

Полный список переменных см. в `env.example`.

## Структура проекта

```
english-bot/
├── cmd/
│   ├── bot/                      # Точка входа приложения
│   └── migrate_training_cards/   # Инструмент миграции тренировочных карточек
├── internal/
│   ├── ai/                       # AI сервис
│   ├── bot/                      # Telegram бот
│   ├── config/                   # Конфигурация
│   ├── database/                 # Инициализация БД
│   ├── integration/              # Интеграционные тесты
│   ├── models/                   # Модели данных
│   ├── repository/               # Репозитории (работа с БД)
│   ├── service/                  # Бизнес-логика
│   ├── utils/                    # Утилиты (Markdown)
│   └── web/                      # Веб-приложение (API + статика)
├── webapp/                       # Vue 3 SPA исходники
│   ├── src/
│   │   ├── api/                  # API клиент
│   │   ├── components/           # Vue компоненты
│   │   ├── composables/          # Vue composables
│   │   ├── layouts/              # Layouts
│   │   ├── router/               # Vue Router
│   │   ├── views/                # Страницы
│   │   └── styles/               # Стили
│   └── dist/                     # Собранный фронтенд (embed в Go)
├── courses/                      # Курсы обучения
│   └── english-grammar/          # Курс грамматики английского
│       ├── admin/                # Админ-панель для просмотра курсов
│       ├── chapters/             # Главы курса (JSON)
│       ├── config/               # Конфигурация генерации
│       ├── prompts/              # Промпты для генерации
│       ├── scripts/              # Скрипты генерации
│       └── test/                 # Тестовый веб-сервис для курса
├── prompts/                      # Промпты для AI
│   ├── english-teacher.txt       # Промпт преподавателя английского
│   └── training-card-generator.txt  # Промпт генератора карточек
├── scripts/                      # Скрипты деплоя и настройки
├── docs/                         # Документация
│   ├── JWT_AUTH.md               # Документация JWT аутентификации
│   ├── nginx-ssl-setup.md       # Настройка nginx с SSL
│   ├── training.md               # Документация системы тренировок
│   └── swagger/                  # Swagger документация API
└── data/                         # SQLite БД (создается автоматически)
```

## Разработка

```bash
make setup-local    # Настройка проекта (создает .env из env.example)
make build          # Сборка (включает сборку webapp)
make run            # Запуск
make dev            # Разработка (backend + frontend одновременно)
make test           # Тесты
make test-verbose   # Тесты с подробным выводом
make lint           # Линтинг
make fmt            # Форматирование кода
make check          # Все CI проверки (тесты, линтинг, coverage)
make swagger        # Генерация Swagger документации
```

### Интеграционные тесты LLM

```bash
# Требуют AI_URL и AI_API_KEY
make llm-words      # Тесты генерации карточек слов
make llm-cards      # Тесты генерации тренировочных карточек
make llm-all        # Все LLM тесты
```

### Фронтенд

```bash
cd webapp
npm install
npm run dev     # Vite dev server (прокси на :8184)
```

## Docker

```bash
make docker-build       # Сборка Docker образа
make docker-run         # Запуск через docker compose
make docker-stop        # Остановка
make docker-logs        # Просмотр логов
make docker-clean       # Очистка (остановка + удаление образа)
make docker-rebuild     # Пересборка (clean + build + run)
```

Для работы с Docker необходимо создать `.env` файл с необходимыми переменными (см. `env.example`).

## База данных

SQLite база создается автоматически при первом запуске.

**Просмотр:**
```bash
sqlite3 ./data/words.db
```

**Бэкап:**
```bash
sqlite3 ./data/words.db ".backup './data/words.db.backup'"
```

## Курсы грамматики

Проект включает систему генерации и изучения курсов грамматики английского языка.

### Структура курсов

Курсы находятся в `courses/english-grammar/` и включают:
- **Главы** (`chapters/`) - JSON файлы с теорией и упражнениями
- **Конфигурация** (`config/`) - статус генерации и настройки
- **Админ-панель** (`admin/`) - веб-интерфейс для просмотра и валидации курсов
- **Тестовый сервис** (`test/`) - веб-сервис для изучения курсов

### Генерация курсов

Подробная информация о генерации курсов находится в `courses/english-grammar/README.md`.

### Просмотр курсов

Для просмотра сгенерированных курсов:
1. Сгенерируйте индекс: `cd courses/english-grammar/admin && node generate-index.js`
2. Запустите локальный сервер: `python3 -m http.server 8000`
3. Откройте `http://localhost:8000/admin/` в браузере

## Деплой

См. [DEPLOYMENT.md](DEPLOYMENT.md)

Подробная документация по:
- Настройке сервера
- Деплою через GitHub releases
- Настройке systemd сервиса
- Настройке nginx с SSL

## Лицензия

MIT License
