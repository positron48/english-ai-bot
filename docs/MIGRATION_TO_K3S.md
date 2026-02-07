# Миграция english в k3s (qantrix.ru)

## 1. Что уже подготовлено

- В `english`:
  - добавлены `DATABASE_DRIVER`, `DATABASE_URL` в конфиг.
  - CI: добавлена сборка/публикация Docker-образа в GHCR.
  - удален legacy автодеплой по SSH.
- В `devops-time-host`:
  - добавлен `apps/english` (app + postgres + pvc + ingress).
  - добавлен Flux image automation для `ghcr.io/positron48/english`.
  - app включен в `clusters/prod/kustomization.yaml`.

## 2. Что сделать на машине с интернетом (финализация Postgres в коде)

Ниже команды для полного завершения перехода на Postgres runtime.

```bash
cd /var/www/my/k3s/english
git checkout feat/k3s-postgres-migration

# 1) добавить драйвер
go get github.com/jackc/pgx/v5/stdlib@latest

# 2) привести зависимости в порядок
go mod tidy

# 3) проверить сборку/тесты
GOCACHE=/tmp/go-build-cache go test ./internal/config ./internal/database ./internal/bot ./cmd/migrate_training_cards
GOCACHE=/tmp/go-build-cache go test ./... 

# 4) локальный smoke с postgres (через docker compose)
docker run --rm --name english-pg-smoke -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=english -e POSTGRES_USER=english -p 55432:5432 -d postgres:16-alpine

export DATABASE_DRIVER=postgres
export DATABASE_URL='postgres://english:postgres@127.0.0.1:55432/english?sslmode=disable'
export AI_URL='https://openrouter.ai/api/v1'
export AI_API_KEY='test'
export AI_PROMPT='test'
export WEBAPP_JWT_SECRET='test'

# проверка запуска (должен подняться HTTP сервер)
timeout 20s go run ./cmd/bot || true

# чистим smoke-бд
docker rm -f english-pg-smoke
```

Если эти шаги прошли, можно собирать и пушить образ в GHCR.

## 3. Миграция SQLite -> Postgres в k3s (через pod’ы)

Правильный поток:
- backup делается нативно на старом сервере (без k3s);
- перенос файла на новый сервер;
- импорт в Postgres выполняется в k3s (через pod’ы).

### 3.1. Найти pod’ы

```bash
kubectl -n english get pods -o wide

# Сохраняем имена в переменные
APP_POD=$(kubectl -n english get pod -l app=english -o jsonpath='{.items[0].metadata.name}')
PG_POD=$(kubectl -n english get pod -l app=english-postgres -o jsonpath='{.items[0].metadata.name}')
echo "$APP_POD"
echo "$PG_POD"
```

### 3.2. Сделать backup SQLite на старом сервере (native)

```bash
# старый сервер (native)
sqlite3 /path/to/words.db ".backup '/tmp/english-backup.db'"
ls -lh /tmp/english-backup.db
```

Скопировать backup на новый сервер (где k3s):

```bash
# выполняется со старого сервера
scp /tmp/english-backup.db <new-server>:/tmp/english-backup.db
```

### 3.3. Перенести backup в postgres pod на новом сервере

```bash
# на новом сервере
kubectl -n english cp /tmp/english-backup.db "$PG_POD":/tmp/english-backup.db
```

### 3.4. Импорт в Postgres внутри pod

Рекомендуемый путь: `pgloader` в отдельном utility pod.

```bash
# вариант A: временно поднять utility pod и скопировать в него backup
kubectl -n english run pgloader --image=dimitri/pgloader --restart=Never -- sleep 3600
kubectl -n english wait --for=condition=Ready pod/pgloader --timeout=120s
kubectl -n english cp /tmp/english-backup.db pgloader:/tmp/english-backup.db
kubectl -n english exec pgloader -- \
  pgloader /tmp/english-backup.db postgresql://english:<PASSWORD>@english-postgres:5432/english
kubectl -n english delete pod pgloader
```

### 3.5. Проверка после импорта

```bash
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT COUNT(*) FROM users;'"
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT COUNT(*) FROM word_cards;'"
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT COUNT(*) FROM training_cards;'"
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT COUNT(*) FROM user_cards;'"
```

## 4. Как повторно перенакатывать базу (re-import)

Когда нужно заново импортировать SQLite в Postgres:

1. Остановить приложение, чтобы не писало в БД во время reload.

```bash
kubectl -n english scale deployment/english --replicas=0
kubectl -n english rollout status deployment/english
```

2. Очистить схему в Postgres.

```bash
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'"
```

3. Повторить шаги из раздела 3:
   - на старом сервере сделать новый native backup;
   - скопировать его на новый сервер;
   - загрузить в pod и выполнить import.

4. Поднять приложение обратно.

```bash
kubectl -n english scale deployment/english --replicas=1
kubectl -n english rollout status deployment/english
kubectl -n english logs deploy/english --tail=200
```

## 5. Полная проверка после миграции (k3s)

```bash
# 1) состояние ресурсов
kubectl -n english get all
kubectl -n english get ingress

# 2) логи приложения и Postgres
kubectl -n english logs deploy/english --tail=200
kubectl -n english logs deploy/english-postgres --tail=200

# 3) health приложения изнутри pod
APP_POD=$(kubectl -n english get pod -l app=english -o jsonpath='{.items[0].metadata.name}')
kubectl -n english exec "$APP_POD" -- sh -lc "wget -qO- http://127.0.0.1:8080/health || true"

# 4) проверка, что приложение пишет в Postgres (пример: updated_at на users меняется)
kubectl -n english exec "$PG_POD" -- sh -lc "psql -U english -d english -c 'SELECT now();'"
```

Функциональные smoke-check (в UI/API):

1. login в webapp;
2. открыть dashboard;
3. открыть vocab;
4. сделать 1-2 ответа в training;
5. убедиться, что в `review_events` и/или `training_sessions` появились новые записи.

## 6. Локальная миграция без k3s (`make dev`)

Сценарий для локальной машины.

### 6.1. Поднять локальный Postgres

```bash
docker run --rm --name english-local-pg \
  -e POSTGRES_DB=english \
  -e POSTGRES_USER=english \
  -e POSTGRES_PASSWORD=english \
  -p 5432:5432 -d postgres:16-alpine
```

### 6.2. Подготовить SQLite backup

```bash
cd /var/www/my/k3s/english
sqlite3 ./data/words.db ".backup './data/english-backup.db'"
```

### 6.3. Импорт в локальный Postgres

Через `pgloader`:

```bash
pgloader ./data/english-backup.db postgresql://english:english@127.0.0.1:5432/english
```

### 6.4. Запустить приложение в dev-режиме на Postgres

```bash
export DATABASE_DRIVER=postgres
export DATABASE_URL='postgres://english:english@127.0.0.1:5432/english?sslmode=disable'

# остальные обязательные env для сервиса
export AI_URL='https://openrouter.ai/api/v1'
export AI_API_KEY='***'
export AI_PROMPT='***'
export WEBAPP_JWT_SECRET='***'

make dev
```

### 6.5. Повторный локальный re-import

```bash
# стопаем dev приложение
# затем чистим схему
psql 'postgresql://english:english@127.0.0.1:5432/english' -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'

# заново импорт
pgloader ./data/english-backup.db postgresql://english:english@127.0.0.1:5432/english

# снова make dev
```

## 7. Откат

- В k3s: `kubectl -n english scale deployment/english --replicas=0`.
- Вернуть DNS/трафик на старую площадку (если применимо).
- Использовать сохраненный SQLite backup для восстановления предыдущего состояния.
