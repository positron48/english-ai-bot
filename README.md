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

```
word_cards (1) ←→ (N) training_cards (1) ←→ (2) user_cards
                                              ↑
                                         (ru_en, en_ru)
```

- **word_cards** - словарные карточки с определениями (создаются при запросе слова)
- **training_cards** - тренировочные карточки (одно значение слова, создаются воркером)
- **user_cards** - карточки пользователя для SRS (по 2 на training_card: RU→EN и EN→RU)
- **training_sessions** - сессии тренировок
- **review_events** - события ответов на карточки

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

## Команды

### Пользовательские
- `/start` - приветствие
- `/help` - справка
- `/train` - начать тренировку
- `/get_id` - получить Telegram ID

### Админские
- `/reset_circuit` - сброс circuit breaker
- `/delete_train [word]` - удалить тренировочные карточки слова
- `/delete_train_all` - удалить все тренировочные карточки
- `/get_train_data [word]` - данные о тренировочных карточках

## Веб-приложение

Vue 3 SPA, встроенное в Go-бинарник через `go:embed`.

### Маршруты
- `/app` или `/app/#/login` - вход
- `/app/#/dashboard` - дашборд
- `/app/#/vocab` - словарь
- `/app/#/training` - тренировка
- `/app/#/chat` - AI чат
- `/app/#/admin` - админ-панель

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
AI_MODEL=                    # Модель AI
AI_PROMPT=                   # Системный промпт (или AI_PROMPT_FILE)
DATABASE_PATH=               # Путь к SQLite БД
WEBAPP_JWT_SECRET=           # Секрет для JWT (openssl rand -hex 32)
```

### Тренировки

```env
TRAINING_WORKER_ENABLED=true              # Включить воркер
TRAINING_WORKER_INTERVAL=30s              # Интервал обработки
TRAINING_WORKER_BATCH_SIZE=5              # Размер батча
TRAINING_PROMPT_FILE=prompts/training-card-generator.txt
CIRCUIT_BREAKER_THRESHOLD=5              # Порог для circuit breaker
TRAINING_OPTIONS_DELAY_MS=5000           # Задержка показа вариантов
TRAINING_WRONG_ANSWER_DELAY_SECONDS=5    # Задержка после ошибки
```

### Веб-приложение

```env
WEBAPP_JWT_TTL_HOURS=24                  # TTL access token
WEBAPP_REFRESH_TTL_HOURS=720             # TTL refresh token
WEBAPP_OTP_TTL_SECONDS=300              # TTL OTP кода
WEBAPP_PUBLIC_URL=https://your-domain.com  # Публичный URL (для CORS)
```

Полный список переменных см. в `env.example`.

## Структура проекта

```
english-bot/
├── cmd/bot/              # Точка входа
├── internal/
│   ├── ai/              # AI сервис
│   ├── bot/             # Telegram бот
│   ├── config/           # Конфигурация
│   ├── database/         # Инициализация БД
│   ├── models/           # Модели данных
│   ├── repository/       # Репозитории (работа с БД)
│   ├── service/          # Бизнес-логика
│   ├── utils/            # Утилиты (Markdown)
│   └── web/              # Веб-приложение (API + статика)
├── webapp/               # Vue SPA исходники
│   ├── src/
│   └── dist/             # Собранный фронтенд (embed в Go)
├── prompts/              # Промпты для AI
└── data/                 # SQLite БД (создается автоматически)
```

## Разработка

```bash
make setup      # Настройка проекта
make build      # Сборка
make run        # Запуск
make dev        # Разработка
make test       # Тесты
make lint       # Линтинг
```

### Фронтенд

```bash
cd webapp
npm install
npm run dev     # Vite dev server (прокси на :8184)
```

## Docker

```bash
make docker-build
make docker-run
```

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

## Деплой

См. [DEPLOYMENT.md](DEPLOYMENT.md)

## Лицензия

MIT License
