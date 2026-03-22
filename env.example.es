# Локальный профиль RU→ES (скопировать в .env.es).
# Общие секреты — в .env (подхватывается перед .env.es в `make up-es`).
# Нужен отдельный Telegram-бот и отдельная БД `spanish` (см. make postgres-dev-init-dbs).
#
#   cp env.example.es .env.es
#   make postgres-up && make postgres-dev-init-dbs

SERVICE_NAME=ai-bot-es

LEARNING_PAIR=ru-es
LEARNING_NATIVE_LANG=ru
LEARNING_TARGET_LANG=es
LEARNING_APP_CODE=spanish
GRAMMAR_BUNDLE_ID=es

TELEGRAM_TOKEN=your_dev_spanish_bot_token_here
TELEGRAM_DEBUG=false

# Другой порт, чтобы параллельно с up-en
SERVER_ADDRESS=:8284
SERVER_PORT=8284
LOG_LEVEL=info

DATABASE_DRIVER=postgres
DATABASE_URL=postgres://english:english@127.0.0.1:5433/spanish?sslmode=disable

AI_URL=https://openrouter.ai/api/v1
AI_MODEL=qwen/qwen3-coder:free
AI_MODEL_HIGH=openai/gpt-5-mini
AI_API_KEY=your_openrouter_api_key_here

AI_PROMPT_FILE=prompts/teacher-ru-es.txt
TRAINING_PROMPT_FILE=prompts/training-card-ru-es.txt

TTS_ENABLED=true
# Только OpenRouter (ключ обычно в .env: TTS_API_KEY / TTS_MODEL / TTS_BASE_URL)
TTS_PROVIDER=openrouter
TTS_BASE_URL=https://openrouter.ai/api/v1
TTS_API_KEY=your_openrouter_key_same_as_ai_ok
TTS_MODEL=openai/gpt-audio-mini
TTS_VOICE=alloy
TTS_AUDIO_DIR=./data/tts-es
TTS_PUBLIC_BASE_PATH=/media/tts
TTS_DICTIONARY_ENABLED=false
# Промпт для chat+audio: {word} → слово в обратных кавычках (как раньше в коде)
TTS_CHAT_PRONUNCIATION_PROMPT=You are a pronunciation machine. Say ONLY the exact Spanish word below as audio, with natural Spanish pronunciation. One word, no greeting, no pause, no repetition. Word: {word}

TRAINING_WORKER_ENABLED=true
TRAINING_WORKER_INTERVAL=30s
TRAINING_LLM_WORKERS=4

# Отдельный JWT для второго инстанса (или тот же для локалки)
WEBAPP_JWT_SECRET=change_me_local_es
WEBAPP_JWT_TTL_HOURS=24
WEBAPP_REFRESH_TTL_HOURS=720
WEBAPP_OTP_TTL_SECONDS=300
