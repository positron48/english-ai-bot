# Настройка произношения TTS (Dictionary + external worker; OpenRouter TTS opt-in)

В проекте используется генерация/кэш аудио произношения по отдельным словам.

По умолчанию **встроенный OpenRouter/OpenAI TTS выключен** (`TTS_OPENROUTER_ENABLED=false`). Активные пути:
1. `Free Dictionary API` для английских слов (когда `TTS_DICTIONARY_ENABLED=true`)
2. Внешний `tts-worker` через internal API (`TTS_INTERNAL_ENABLED=true`): cron на Mac берёт pending-слова, генерирует локально и заливает MP3 обратно

При `TTS_PROVIDER=auto` и `TTS_OPENROUTER_ENABLED=true` (legacy/opt-in) приоритет такой:
1. `Free Dictionary API` (основной, бесплатный)
2. Fallback TTS: если `TTS_BASE_URL` — **OpenRouter** (`openrouter.ai`), приложение дергает `/chat/completions` с `modalities: ["text","audio"]`, стрим, парсит `delta.audio.data` (PCM), конвертирует в MP3 через **ffmpeg** (должен быть в образе). Если базовый URL — **OpenAI** (`api.openai.com`), используется `/audio/speech` и в ответе уже MP3.

Если у Dictionary есть аудио для слова, до TTS fallback запрос не дойдёт.

Важно по доменной логике:
- произношение кэшируется и запрашивается только для канонического `word`/lemma;
- `training_card.display_word` (например `to spy`) не является отдельной сущностью произношения.

## 1) Локальная настройка `.env`

Возьми за основу `env.example` и задай ключи.

Обязательно для фичи произношения:

```env
TTS_ENABLED=true
TTS_PROVIDER=auto
TTS_OPENROUTER_ENABLED=false
TTS_AUDIO_DIR=/app/data/tts
TTS_PUBLIC_BASE_PATH=/media/tts

TTS_DICTIONARY_ENABLED=true
TTS_DICTIONARY_BASE_URL=https://api.dictionaryapi.dev/api/v2/entries/en
TTS_DICTIONARY_MIN_DELAY=100ms

TTS_INTERNAL_ENABLED=true
TTS_INTERNAL_TOKENS_JSON={"en":"..."}

TTS_PREFETCH_ENABLED=true
TTS_PREFETCH_WORKERS=2
TTS_BACKFILL_INTERVAL=10m
TTS_BACKFILL_BATCH_SIZE=200
TTS_RETRY_BASE_DELAY=1m
TTS_RETRY_MAX_DELAY=24h
TTS_MAX_RETRIES=8
```

Опционально для legacy fallback TTS (только при `TTS_OPENROUTER_ENABLED=true`):

```env
TTS_OPENROUTER_ENABLED=true
TTS_API_KEY=...          # ключ OpenRouter или OpenAI
TTS_MODEL=openai/gpt-4o-audio-preview   # для OpenRouter; для OpenAI — tts-1, tts-1-hd, gpt-4o-mini-tts
TTS_BASE_URL=https://openrouter.ai/api/v1   # или https://api.openai.com/v1
TTS_VOICE=alloy
TTS_REQUEST_TIMEOUT=45s
```

Важно:
- **OpenRouter** (`TTS_BASE_URL` с `openrouter.ai`): приложение вызывает `/chat/completions` с жёстким промптом «только слово», стрим, собирает PCM из `delta.audio.data`, конвертирует в MP3 через **ffmpeg** (в Docker-образе ffmpeg уже добавлен).
- **OpenAI** (`TTS_BASE_URL=https://api.openai.com/v1`): используется `/audio/speech`, в ответе сразу MP3, ffmpeg не нужен.
- `TTS_REQUEST_TIMEOUT` — общий таймаут (по умолчанию 45s). При таймаутах увеличь до 60s.
- Если `TTS_API_KEY` пустой или `TTS_OPENROUTER_ENABLED=false`, встроенный TTS fallback выключен.
- Сервис работает только с английскими (latin) словами; не-latin отбрасываются на нормализации.

## 2) Настройка в k3s / Flux

### 2.1 ConfigMap (не секреты)

В `devops-time-host/apps/english/base/configmap.yaml` должны быть:

```yaml
TTS_ENABLED: "true"
TTS_PROVIDER: "auto"
TTS_EXTERNAL_ONLY: "false"
TTS_OPENROUTER_ENABLED: "false"
TTS_AUDIO_DIR: "/app/data/tts"
TTS_PUBLIC_BASE_PATH: "/media/tts"
TTS_DICTIONARY_ENABLED: "true"
TTS_DICTIONARY_BASE_URL: "https://api.dictionaryapi.dev/api/v2/entries/en"
TTS_DICTIONARY_MIN_DELAY: "100ms"
TTS_INTERNAL_ENABLED: "true"
TTS_BASE_URL: "https://openrouter.ai/api/v1" # legacy, only if TTS_OPENROUTER_ENABLED=true
TTS_MODEL: "openai/gpt-audio-mini"
TTS_VOICE: "alloy"
TTS_REQUEST_TIMEOUT: "15s"
TTS_PREFETCH_ENABLED: "true"
TTS_PREFETCH_WORKERS: "2"
TTS_BACKFILL_INTERVAL: "10m"
TTS_BACKFILL_BATCH_SIZE: "200"
TTS_RETRY_BASE_DELAY: "1m"
TTS_RETRY_MAX_DELAY: "24h"
TTS_MAX_RETRIES: "8"
```

### 2.2 Secret (только чувствительные значения)

В `english-secrets` (namespace `english`):

- обязательные для старта приложения:
  - `DATABASE_URL`
  - `AI_API_KEY`
  - `WEBAPP_JWT_SECRET` (или `WEBAPP_SESSION_SECRET`)
- опциональные:
  - `TELEGRAM_TOKEN` (если используешь Telegram-бот)
  - `TTS_API_KEY` (если включаешь fallback OpenRouter)

Не-секретные параметры (`AI_URL`, `AI_MODEL`, `AI_MODEL_HIGH`, `AI_PROMPT_FILE`, а также `TTS_*` кроме ключа) задаются в GitOps ConfigMap.

Пример безопасного обновления основного секрета (`create --dry-run | apply`):

```bash
kubectl -n english create secret generic english-secrets \
  --from-literal=DATABASE_URL='postgres://english:<POSTGRES_PASSWORD>@english-postgres:5432/english?sslmode=disable' \
  --from-literal=AI_API_KEY='***' \
  --from-literal=WEBAPP_JWT_SECRET='***' \
  --dry-run=client -o yaml | kubectl apply -f -
```

Если включаешь fallback OpenRouter, повтори ту же команду, добавив `TTS_API_KEY`:

```bash
kubectl -n english create secret generic english-secrets \
  --from-literal=DATABASE_URL='postgres://english:<POSTGRES_PASSWORD>@english-postgres:5432/english?sslmode=disable' \
  --from-literal=AI_API_KEY='***' \
  --from-literal=WEBAPP_JWT_SECRET='***' \
  --from-literal=TTS_API_KEY='***' \
  --dry-run=client -o yaml | kubectl apply -f -
```

Полный релизный runbook для k3s находится в GitOps-репозитории:
`devops-time-host/apps/english/RELEASE_K3S.md`.

### 2.3 Постоянное хранилище (уже подключено)

- PVC: `english-tts-data`
- Путь в контейнере: `/app/data/tts`
- Кэш переживает restart/rollout.

### 2.4 Бэкап (уже подключён)

В `k3s-backup` уже включён архив кэша TTS:
- `/app/data/tts` -> `english/tts.tar.gz`

## 3) Чеклист выката

1. Закоммить изменения в:
   - `english-ai-bot`
   - `devops-time-host`
2. Запушь оба репозитория.
3. Дождись сборки образа и Flux reconcile.

Опционально ускорить:

```bash
flux reconcile image repository english -n flux-system
flux reconcile image policy english -n flux-system
flux reconcile image update flux-system -n flux-system
flux reconcile kustomization flux-system -n flux-system --with-source
```

4. Проверка:

```bash
kubectl get pods -n english -l app=english
kubectl logs -n english deploy/english --tail=200 | rg -i "tts|pronunciation|dictionary|openrouter"
```

5. Проверка в UI:
- Открой карточку английского слова с транскрипцией.
- Кнопка произношения показывается только когда файл уже есть в кэше.
- При первом запросе файл создаётся в фоне, затем кнопка становится активной.

## 4) Как заполняются уже существующие слова

Есть два механизма:

1. On-demand:
- при открытии карточки/тренировки фронт вызывает `GET /api/tts/word?word=...`;
- если файла нет, сервис ставит слово в очередь фоновой генерации.

2. Периодический backfill:
- при старте и далее раз в `TTS_BACKFILL_INTERVAL` сервис берёт кандидатов из `word_cards.word` (канонические lemma);
- размер порции регулируется `TTS_BACKFILL_BATCH_SIZE`.

Для ускоренного «прогрева» существующей базы временно выставь:

```yaml
TTS_PREFETCH_ENABLED: "true"
TTS_PREFETCH_WORKERS: "4"
TTS_BACKFILL_INTERVAL: "1m"
TTS_BACKFILL_BATCH_SIZE: "2000"
```

После прогрева верни более щадящие значения (`10m`/`200`/`2`).

Примечание: текущий backfill ориентирован на свежие слова (по `created_at`), поэтому при очень большом историческом объёме старые слова будут прогреваться медленнее.

## 5) Поведение API

- `GET /api/tts/word?word=<word>`
  - возвращает `{ available, url, word }`
  - если файла нет, ставит генерацию в фон
- `GET /media/tts/...`
  - отдаёт кэшированный mp3 с длинными immutable cache headers

## 6) Backoff/ретраи

При ошибках генерации применяется экспоненциальный backoff:
- базовая задержка: `TTS_RETRY_BASE_DELAY`
- рост: x2 на каждую попытку
- потолок: `TTS_RETRY_MAX_DELAY`
- максимум попыток: `TTS_MAX_RETRIES`

Пока не наступит `nextTry`, слово повторно в очередь не попадёт.

## 7) OpenRouter для fallback

Fallback при `TTS_BASE_URL` с `openrouter.ai`:
- Запрос: `POST /chat/completions`, `modalities: ["text","audio"]`, `stream: true`, `max_tokens: 150`, один `user` prompt (без `system`) с инструкцией «озвучить только слово».
- Ответ: SSE-стрим; приложение собирает `choices[0].delta.audio.data` (base64 PCM 16-bit 24 kHz), декодирует, конвертирует в MP3 через **ffmpeg** и сохраняет в кэш.

Рекомендованные модели (audio output): `openai/gpt-audio-mini`, `openai/gpt-4o-audio-preview`. Для стоимости и лимитов смотри [openrouter.ai/models](https://openrouter.ai/models) (фильтр output = audio).

## 8) Reading Texts: сегментное multi-voice аудио

Для блока `reading_passage` аудио хранится как сегменты (`audio_rel_path`) в `courses/*/assets/reading/...` и копируется в `internal/grammarbundle/*/assets` при `make grammar-bundle`.

Генерация делается локально через:
- `courses/english-grammar/scripts/generate-reading-text.py`
- `courses/spanish-grammar/scripts/generate-reading-text.py`

Профили голосов:
- `courses/english-grammar/config/reading-voices.json`
- `courses/spanish-grammar/config/reading-voices.json`

Минимальный preflight:
1. `make -C courses/english-grammar reading-validate`
2. `make -C courses/spanish-grammar reading-validate`
3. Убедиться, что для каждого сегмента `reading_passage.segments[].audio_rel_path` существует файл.
