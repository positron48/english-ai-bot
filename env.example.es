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

# В общем .env должны быть AI_URL, AI_API_KEY, AI_MODEL (и при желании TTS_* — см. env.example)
AI_PROMPT_FILE=prompts/teacher-ru-es.txt
TRAINING_PROMPT_FILE=prompts/training-card-ru-es.txt

TTS_ENABLED=true
TTS_PROVIDER=openrouter
TTS_EXTERNAL_ONLY=false
TTS_AUDIO_DIR=./data/tts-es
TTS_DICTIONARY_ENABLED=false
# Промпт для chat+audio: {word} → слово в обратных кавычках (как раньше в коде)
TTS_CHAT_PRONUNCIATION_PROMPT="You are a pronunciation machine. Say ONLY the exact Spanish word below as audio, with natural Spanish pronunciation. One word, no greeting, no pause, no repetition. Word: {word}"
TTS_INTERNAL_ENABLED=false
TTS_INTERNAL_TOKENS_JSON=
TTS_INTERNAL_MAX_PENDING_LIMIT=500
TTS_INTERNAL_MAX_UPLOAD_MB=10

TRAINING_WORKER_ENABLED=true
TRAINING_WORKER_INTERVAL=30s
TRAINING_LLM_WORKERS=4

# Тренировка спряжений (отдельный SRS): /api/verb-training/* и экран «Формы»
SPANISH_VERB_FORMS_ENABLED=true
# Сколько карточек в одной сессии и сколько из них могут быть в состоянии new (остальное — просроченные review/learning).
# По умолчанию 30/30; новые карточки набираются вперемешку по разным словам (round-robin по word_card_id).
# VERB_FORMS_MAX_CARDS_PER_SESSION=30
# VERB_FORMS_MAX_NEW_PER_SESSION=30
# После стольких успешных повторов (reps) без состояния learning — карточка допускает режим ввода формы целиком
VERB_FORMS_TYPED_MIN_REPS=2
# При допуске к вводу: с какой вероятностью (0–100) показывать ввод, а не варианты; 50 ≈ как spell/type в тренировке слов
VERB_FORMS_TYPED_CHANCE_PERCENT=50

# Отдельный JWT для второго инстанса (или тот же для локалки)
WEBAPP_JWT_SECRET=change_me_local_es
COMPLAINTS_SERVICE_TOKEN=
COMPLAINTS_SERVICE_URL=http://127.0.0.1:8284
LLAMACPP_URL=http://127.0.0.1:8080
LLAMACPP_MODEL=local-model
# Optional auto-start command if LLAMACPP_URL is down:
# LLAMACPP_START_CMD='nohup llama-server -m "/path/to/model.gguf" --host 127.0.0.1 --port 8080 -ngl 999 -c 8192 >/tmp/llama-complaints.log 2>&1 &'
# Verb forms (large ctx): LLAMACPP_START_CMD_VERB='... -c 32768 -n 16384 ...'
# Reading on Mac (qwen3:30b): use smaller ctx in courses/spanish-grammar/.env.local:
# LLAMACPP_START_CMD_READING='pkill -f "llama-server.*--port 8080" || true; sleep 2; nohup llama-server ... -c 8192 -n 2048 >/tmp/llama-reading.log 2>&1 &'
# READING_CTX_TOKENS=8192  READING_RESTART_ON_OOM=1
WEBAPP_JWT_TTL_HOURS=24
WEBAPP_REFRESH_TTL_HOURS=720
WEBAPP_OTP_TTL_SECONDS=300
