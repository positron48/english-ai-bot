# Speaking Mode — env и rollout

## Env (Spanish instance)

```bash
SPEAKING_MODE_ENABLED=true
SPEAKING_EVAL_MODEL=openai/gpt-audio-mini
SPEAKING_EVAL_BASE_URL=https://openrouter.ai/api/v1
SPEAKING_EVAL_API_KEY=...          # optional: falls back to TTS_API_KEY / AI_API_KEY
SPEAKING_EVAL_TIMEOUT=60s
SPEAKING_MAX_AUDIO_MB=2
SPEAKING_MAX_ATTEMPTS_DEFAULT=3
SPEAKING_ACCEPT_MEANING_SCORE=3
SPEAKING_SESSION_TASK_COUNT=5
WEBAPP_RATE_LIMIT_SPEAKING_PER_USER=30
```

## PRO / tier

Пользовательский tier: `users.subscription_tier` (`free` | `pro` | `pro_plus`).

Speaking доступен при `TierAllowsFeature(tier, "speaking")` — то есть `pro` и выше.

Admin: **Пользователи → Подписка → Tier** или `PUT /api/admin/users/{id}/subscription-tier` body `{ "tier": "pro" }`.

## Контент

Источник: `courses/spanish-grammar/speaking/` → `make grammar-bundle` / `./scripts/generate-grammar-bundle.sh es`.

Seed: `python3 courses/spanish-grammar/speaking/generate_seed.py`.

При старте приложения каталог синхронизируется в БД (`speaking_categories`, `speaking_tasks`).

## API

- `GET /api/learning/speaking/availability` — без tier gate (для UI)
- Остальные `/api/learning/speaking/*` — требуют tier `pro+`
