# Миграция english: native SQLite -> k3s Postgres

Ниже только актуальная последовательность команд.

## 1) Установка секретов для pod'ов (на новом сервере с k3s)

```bash
# если приложение еще не применено в кластер
kubectl apply -k /var/www/my/k3s/devops-time-host/apps/english/prod

# секрет приложения
kubectl -n english create secret generic english-secrets \
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

# секрет Postgres
kubectl -n english create secret generic english-postgres \
  --from-literal=POSTGRES_DB='english' \
  --from-literal=POSTGRES_USER='english' \
  --from-literal=POSTGRES_PASSWORD='***' \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n english get secrets
kubectl -n english get pods
```

Если pod падает с ошибкой `failed to read prompt file prompts/english-teacher.txt`:

```bash
# быстрый обходной путь: задать AI_PROMPT напрямую (без файла)
kubectl -n english create secret generic english-secrets \
  --from-literal=AI_URL='https://openrouter.ai/api/v1' \
  --from-literal=AI_API_KEY='***' \
  --from-literal=AI_MODEL='openai/gpt-5-mini' \
  --from-literal=AI_MODEL_HIGH='openai/gpt-5-mini' \
  --from-literal=AI_PROMPT='You are a helpful English teacher.' \
  --from-literal=WEBAPP_JWT_SECRET='***' \
  --from-literal=WEBAPP_SESSION_SECRET='***' \
  --from-literal=TELEGRAM_TOKEN='***' \
  --from-literal=ADMIN_TELEGRAM_ID='0' \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n english rollout restart deployment/english
kubectl -n english rollout status deployment/english
```

### 1.1) Доступ к приватному GHCR-образу (image pull secret)

Нужен GitHub PAT с правами минимум `read:packages` (и доступом к репозиторию, если он private).

```bash
# создать/обновить docker-registry secret для ghcr.io
kubectl -n english create secret docker-registry ghcr-creds \
  --docker-server=ghcr.io \
  --docker-username='<GITHUB_USERNAME>' \
  --docker-password='<GITHUB_PAT_WITH_read:packages>' \
  --docker-email='<EMAIL>' \
  --dry-run=client -o yaml | kubectl apply -f -

# привязать secret к deployment
kubectl -n english patch deployment english \
  --type='merge' \
  -p '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"ghcr-creds"}]}}}}'

# перезапустить pod
kubectl -n english rollout restart deployment/english
kubectl -n english rollout status deployment/english
```

Проверка причины, если pull снова упал:

```bash
kubectl -n english describe pod -l app=english | sed -n '/Events/,$p'
```

## 2) Дамп SQLite на старом (native) сервере + перенос на новый сервер с k3s

### На старом сервере (где сервис работает нативно)

```bash
sqlite3 /path/to/words.db ".backup '/tmp/english-backup.db'"
ls -lh /tmp/english-backup.db
```

### Скопировать на новый сервер (k3s)

```bash
scp /tmp/english-backup.db root@<NEW_SERVER_IP>:/tmp/english-backup.db
```

### На новом сервере проверить файл

```bash
ls -lh /tmp/english-backup.db
```

## 3) Переключение на Postgres + импорт дампа в pod + миграция в Postgres

### 3.1. Остановить приложение на время импорта

```bash
kubectl -n english scale deployment/english --replicas=0
kubectl -n english get pods
```

### 3.2. Найти postgres pod

```bash
PG_POD=$(kubectl -n english get pod -l app=english-postgres -o jsonpath='{.items[0].metadata.name}')
echo "$PG_POD"
```

### 3.3. Очистить БД перед импортом (если нужен чистый повторный накат)

```bash
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'"
```

### 3.4. Импорт SQLite -> Postgres внутри k3s

```bash
# временный pod с pgloader
kubectl -n english run pgloader --image=dimitri/pgloader --restart=Never -- sleep 3600
kubectl -n english wait --for=condition=Ready pod/pgloader --timeout=120s

# копируем backup в pod pgloader
kubectl -n english cp /tmp/english-backup.db pgloader:/tmp/english-backup.db

# запускаем миграцию в сервис postgres внутри namespace english
kubectl -n english exec pgloader -- \
  pgloader /tmp/english-backup.db postgresql://english:<POSTGRES_PASSWORD>@english-postgres:5432/english

# удаляем временный pod
kubectl -n english delete pod pgloader
```

### 3.5. Проверка данных в Postgres

```bash
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT COUNT(*) FROM users;'"
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT COUNT(*) FROM word_cards;'"
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT COUNT(*) FROM training_cards;'"
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT COUNT(*) FROM user_cards;'"
```

### 3.6. Поднять приложение обратно

```bash
kubectl -n english scale deployment/english --replicas=1
kubectl -n english rollout status deployment/english
kubectl -n english logs deployment/english --tail=200
```

Готово: приложение работает в k3s с Postgres и данными из старого SQLite.
