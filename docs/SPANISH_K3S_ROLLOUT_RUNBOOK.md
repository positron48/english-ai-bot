# Запуск Spanish (RU→ES) в k3s: что сделать после мержа кода

Кратко: один репозиторий приложения (`english-ai-bot`), два образа в GHCR (`english`, `spanish`), второй инстанс в кластере — namespace `spanish`, отдельная БД и секреты. GitOps-манифесты лежат в `devops-time-host`.

## 1. Репозиторий `english-ai-bot`

1. Смержить ветку с multilang/Spanish в default branch (`master`).
2. Создать и запушить **git tag** (релизный workflow публикует Docker только на tag, см. `.github/workflows/ci.yml`, job `docker-image` с matrix `english` / `spanish`).
3. Дождаться успешного CI: в GHCR должны обновиться `ghcr.io/<owner>/english:latest` и `ghcr.io/<owner>/spanish:latest`.

Если GitHub owner не `positron48`, в `devops-time-host` поправь образы в `apps/spanish/base/deployment.yaml` и `clusters/prod/infra/image-automation/spanish-image.yaml` под свой `ghcr.io/<owner>/spanish`.

## 2. Репозиторий `devops-time-host`

1. Закоммитьть и запушить изменения (уже подготовлены в дереве):
   - `apps/spanish/` — Deployment, Postgres, PVC, Service, Ingress, ConfigMap;
   - `clusters/prod/kustomization.yaml` — подключён `../../apps/spanish/prod`;
   - `clusters/prod/infra/image-automation/spanish-image.yaml` + запись в `.../kustomization.yaml`;
   - `apps/k3s-backup/` — дамп БД и TTS для `spanish`;
   - `clusters/prod/infra/observability/alloy-config.yaml` — логи namespace `spanish`;
   - `clusters/prod/infra/observability/grafana.yaml` — datasource **Postgres Spanish** (нужны ключи в секрете, см. ниже).
2. Flux подхватит коммит (или выполни reconcile, см. §5).

### Домен

По умолчанию в манифестах: **https://es.qantrix.ru** (`ingress` + `WEBAPP_PUBLIC_URL` в `spanish-config`). Если нужен другой host — одним коммитом поменяй `apps/spanish/base/ingress.yaml` и `WEBAPP_PUBLIC_URL` в `apps/spanish/base/configmap.yaml`, настрой DNS на ingress.

## 3. Не-секретные переменные (уже в GitOps)

Файл `devops-time-host/apps/spanish/base/configmap.yaml` задаёт профиль RU→ES:

| Переменная | Значение (как в репо) |
|------------|----------------------|
| `LEARNING_PAIR` | `ru-es` |
| `LEARNING_NATIVE_LANG` | `ru` |
| `LEARNING_TARGET_LANG` | `es` |
| `LEARNING_APP_CODE` | `spanish` |
| `GRAMMAR_BUNDLE_ID` | `es` |
| `AI_URL` | `https://openrouter.ai/api/v1` |
| `AI_MODEL` / `AI_MODEL_HIGH` | `openai/gpt-5-mini` |
| `AI_PROMPT_FILE` | `prompts/teacher-ru-es.txt` |
| `TRAINING_PROMPT_FILE` | `prompts/training-card-ru-es.txt` |
| `TRAINING_WORKER_ENABLED` | `true` |
| `TRAINING_WORKER_INTERVAL` | `30s` |
| `TRAINING_LLM_WORKERS` | `4` |
| `TTS_ENABLED` | `true` |
| `TTS_PROVIDER` | `auto` |
| `TTS_DICTIONARY_ENABLED` | `false` |
| `TTS_BASE_URL` | `https://openrouter.ai/api/v1` |
| `TTS_MODEL` | `openai/gpt-audio-mini` |
| `TTS_VOICE` | `alloy` |
| `TTS_AUDIO_DIR` | `/app/data/tts` |
| `TTS_PUBLIC_BASE_PATH` | `/media/tts` |
| `TTS_PREFETCH_*` / `TTS_BACKFILL_*` / `TTS_RETRY_*` | как в ConfigMap (см. файл) |
| `TTS_CHAT_PRONUNCIATION_PROMPT` | испанский one-word prompt (в ConfigMap) |
| `TELEGRAM_SOCKS5_PROXY_ADDR` | как у English (`51.254.98.124:1080` в репо) |
| `WEBAPP_PUBLIC_URL` | `https://es.qantrix.ru` |

Секреты в ConfigMap не кладём: `DATABASE_URL`, `AI_API_KEY`, `WEBAPP_JWT_SECRET`, `TELEGRAM_TOKEN`, `TTS_API_KEY` — только в `spanish-secrets`.

## 4. Секреты на сервере (kubectl)

Полный текст команд — в `devops-time-host/apps/spanish/RELEASE_K3S.md`. Минимум:

1. Namespace `spanish`.
2. `spanish-postgres`: `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` (в примерах БД/пользователь `spanish`).
3. `spanish-secrets`: `DATABASE_URL` на `spanish-postgres:5432`, `AI_API_KEY`, `WEBAPP_JWT_SECRET`; для бота — `TELEGRAM_TOKEN`; для TTS через OpenRouter — `TTS_API_KEY` (часто тот же ключ, что и `AI_API_KEY`).
4. При приватном GHCR — `ghcr-creds` в namespace `spanish` (как у English).

### Grafana

Добавь в существующий секрет `observability/grafana-db-datasources` ключи (сохрани все уже существующие ключи budget/english/positroid в той же команде):

- `SPANISH_POSTGRES_DB=spanish`
- `SPANISH_POSTGRES_USER=spanish`
- `SPANISH_POSTGRES_PASSWORD=<пароль из spanish-postgres>`

Пример шаблона (подставь реальные значения для **всех** баз):

```bash
kubectl -n observability create secret generic grafana-db-datasources \
  --from-literal=BUDGET_POSTGRES_PASSWORD='…' \
  --from-literal=ENGLISH_POSTGRES_DB='english' \
  --from-literal=ENGLISH_POSTGRES_USER='english' \
  --from-literal=ENGLISH_POSTGRES_PASSWORD='…' \
  --from-literal=SPANISH_POSTGRES_DB='spanish' \
  --from-literal=SPANISH_POSTGRES_USER='spanish' \
  --from-literal=SPANISH_POSTGRES_PASSWORD='…' \
  --from-literal=POSITROID_DB_NAME='…' \
  --from-literal=POSITROID_DB_USER='…' \
  --from-literal=POSITROID_DB_PASSWORD='…' \
  --dry-run=client -o yaml | kubectl apply -f -
```

После изменения секрета при необходимости перезапусти под Grafana.

## 5. Flux

```bash
flux reconcile image repository spanish -n flux-system
flux reconcile image policy spanish -n flux-system
flux reconcile image update flux-system -n flux-system
flux reconcile kustomization flux-system -n flux-system --with-source
```

## 6. Проверки

```bash
kubectl get pods -n spanish
kubectl get ingress -n spanish
kubectl logs -n spanish deploy/spanish --tail=200
curl -fsS https://es.qantrix.ru/health
```

## 7. См. также

- `docs/SPANISH_K3S_ROLLOUT_CHECKLIST.md` — короткий чеклист.
- `docs/MULTILANG_SPANISH_LAUNCH_PLAN.md` — архитектура и этапы кода.
- `devops-time-host/apps/spanish/RELEASE_K3S.md` — секреты и Flux одной страницей.
