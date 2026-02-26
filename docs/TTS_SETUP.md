# Pronunciation TTS Setup (Dictionary first, OpenAI fallback)

This project uses word-level pronunciation audio with filesystem cache.

Priority order in `TTS_PROVIDER=auto`:
1. `Free Dictionary API` (primary, free)
2. `OpenAI /audio/speech` (fallback, only if API key + model are set)

If dictionary audio exists for a word, OpenAI is not called.

## 1) Local `.env` setup

Use `env.example` as base and set these keys.

Required for pronunciation feature:

```env
TTS_ENABLED=true
TTS_PROVIDER=auto
TTS_AUDIO_DIR=/app/data/tts
TTS_PUBLIC_BASE_PATH=/media/tts

TTS_DICTIONARY_ENABLED=true
TTS_DICTIONARY_BASE_URL=https://api.dictionaryapi.dev/api/v2/entries/en

TTS_PREFETCH_ENABLED=true
TTS_PREFETCH_WORKERS=2
TTS_BACKFILL_INTERVAL=10m
TTS_BACKFILL_BATCH_SIZE=200
TTS_RETRY_BASE_DELAY=1m
TTS_RETRY_MAX_DELAY=24h
TTS_MAX_RETRIES=8
```

Optional (OpenAI fallback):

```env
TTS_API_KEY=...
TTS_MODEL=gpt-4o-mini-tts
TTS_BASE_URL=https://api.openai.com/v1
TTS_VOICE=alloy
TTS_REQUEST_TIMEOUT=15s
```

Important:
- If `TTS_API_KEY` or `TTS_MODEL` is empty, OpenAI fallback is disabled automatically.
- English-only behavior is enforced in service normalization; non-Latin words are ignored.

## 2) k3s / Flux setup

### 2.1 ConfigMap (non-secret)

In `devops-time-host/apps/english/base/configmap.yaml` keep:

```yaml
TTS_ENABLED: "true"
TTS_PROVIDER: "auto"
TTS_AUDIO_DIR: "/app/data/tts"
TTS_PUBLIC_BASE_PATH: "/media/tts"
TTS_DICTIONARY_ENABLED: "true"
TTS_DICTIONARY_BASE_URL: "https://api.dictionaryapi.dev/api/v2/entries/en"
TTS_REQUEST_TIMEOUT: "15s"
TTS_PREFETCH_ENABLED: "true"
TTS_PREFETCH_WORKERS: "2"
TTS_BACKFILL_INTERVAL: "10m"
TTS_BACKFILL_BATCH_SIZE: "200"
```

### 2.2 Secret (optional fallback keys)

`english-secrets` in namespace `english`:

- required for TTS fallback only:
  - `TTS_API_KEY`
  - `TTS_MODEL`
- optional:
  - `TTS_BASE_URL`
  - `TTS_VOICE`

Dictionary-only mode does **not** require TTS secrets.

Example update command (safe apply):

```bash
kubectl -n english create secret generic english-secrets \
  --from-literal=DATABASE_URL='postgres://english:<POSTGRES_PASSWORD>@english-postgres:5432/english?sslmode=disable' \
  --from-literal=AI_URL='https://openrouter.ai/api/v1' \
  --from-literal=AI_API_KEY='***' \
  --from-literal=AI_MODEL='openai/gpt-5-mini' \
  --from-literal=AI_MODEL_HIGH='openai/gpt-5-mini' \
  --from-literal=AI_PROMPT_FILE='prompts/english-teacher.txt' \
  --from-literal=WEBAPP_JWT_SECRET='***' \
  --from-literal=WEBAPP_SESSION_SECRET='***' \
  --from-literal=TELEGRAM_TOKEN='***' \
  --from-literal=ADMIN_TELEGRAM_ID='0' \
  --dry-run=client -o yaml | kubectl apply -f -
```

To enable OpenAI fallback, add keys:

```bash
kubectl -n english create secret generic english-secrets \
  --from-literal=TTS_API_KEY='***' \
  --from-literal=TTS_MODEL='gpt-4o-mini-tts' \
  --from-literal=TTS_BASE_URL='https://api.openai.com/v1' \
  --from-literal=TTS_VOICE='alloy' \
  --dry-run=client -o yaml | kubectl apply -f -
```

### 2.3 Persistent storage (already wired)

- PVC: `english-tts-data`
- Mount path in pod: `/app/data/tts`
- Files survive pod restart/rollout.

### 2.4 Backup (already wired)

`k3s-backup` includes TTS cache archive:
- `/app/data/tts` -> `english/tts.tar.gz`

## 3) Rollout checklist

1. Commit changes in:
   - `english-ai-bot`
   - `devops-time-host`
2. Push both repos.
3. Wait for image build + Flux reconcile.
   Optional acceleration:

```bash
flux reconcile image repository english -n flux-system
flux reconcile image policy english -n flux-system
flux reconcile image update flux-system -n flux-system
flux reconcile kustomization flux-system -n flux-system --with-source
```

4. Verify:

```bash
kubectl get pods -n english -l app=english
kubectl logs -n english deploy/english --tail=200 | rg -i "tts|pronunciation|dictionary|openai"
```

5. Runtime check:
   - Open web app, open any English card with transcription.
   - Pronounce button appears only when file is already cached.
   - First requests schedule background generation; button appears after cache file exists.

## 4) API behavior summary

- `GET /api/tts/word?word=<word>`
  - returns `{ available, url, word }`
  - triggers background generation when cache is missing
- `GET /media/tts/...`
  - serves cached mp3 with long immutable cache headers

## 5) Cost note for OpenAI fallback

Check current pricing before enabling fallback in production:
- OpenAI pricing page: https://openai.com/api/pricing/
- TTS model docs: https://platform.openai.com/docs/models/tts-1-hd

At the time of writing, TTS docs show:
- `TTS-1`: `$15.00` per 1M tokens
- `TTS-1 HD`: `$30.00` per 1M tokens

Treat these values as subject to change and validate before rollout.
