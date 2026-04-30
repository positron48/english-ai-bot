# Local Complaints Worker

Этот документ описывает локальный hourly-процесс обработки grammar-жалоб, вынесенный из релизного Go-сервиса.

## Что находится в релизном сервисе

- `GET /api/internal/content-reports/grammar`
- `POST /api/internal/content-reports/grammar/resolve-bulk`
- Авторизация через `X-Service-Token` (`WEBAPP_INTERNAL_SERVICE_TOKENS_JSON`).

В сервисе нет LLM-анализа, нет редактирования `courses/`, нет локального scheduler.

## Локальный инструмент

- Скрипт: `tools-local/complaints-worker/worker.py`
- Обертка: `tools-local/complaints-worker/run.sh`
- `dry-run` по умолчанию.
- Для реального удаления вопросов и resolve нужен флаг `--apply`.

## Переменные окружения

- `COMPLAINTS_SERVICE_URL` - базовый URL сервиса, например `http://127.0.0.1:8184`
- `COMPLAINTS_SERVICE_TOKEN` - service token для internal API
- `LLAMACPP_URL` - URL локального `llama.cpp` OpenAI-compatible API, например `http://127.0.0.1:8080`
- `LLAMACPP_MODEL` - имя модели
- `COURSE_SCOPE` - `english|spanish|both`
- `WORKSPACE_ROOT` - опционально, путь к workspace

## Ручной запуск

Dry-run:

```bash
python3 tools-local/complaints-worker/worker.py
```

Боевой запуск:

```bash
python3 tools-local/complaints-worker/worker.py --apply
```

## JSONL журнал

Файл:

- `logs/complaints/complaints-YYYY-MM.jsonl`

Основные поля записи:

- `timestamp`
- `run_id`
- `course`
- `chapter_id`
- `theory_block_id`
- `question_ids`
- `report_ids`
- `llm_diagnosis`
- `action` (`dry_run|removed|noop|error`)
- `error`
- `resolve_status`
- `hash_before`
- `hash_after`
- `training_pack_relpath`
- `dry_run`

Список затронутых блоков:

- `logs/complaints/changed-theory-blocks-YYYYMMDDHH.json`

Этот файл используется как вход для последующей пакетной ревалидации/перегенерации блоков.

## launchd (каждый час)

Шаблон plist:

- `tools-local/complaints-worker/launchd/com.englishai.complaints-worker.plist`

Установка:

```bash
mkdir -p "$HOME/Library/LaunchAgents"
cp tools-local/complaints-worker/launchd/com.englishai.complaints-worker.plist "$HOME/Library/LaunchAgents/"
launchctl unload "$HOME/Library/LaunchAgents/com.englishai.complaints-worker.plist" 2>/dev/null || true
launchctl load "$HOME/Library/LaunchAgents/com.englishai.complaints-worker.plist"
```

Проверка:

```bash
launchctl list | rg complaints-worker
```

Логи:

- `logs/complaints/launchd.stdout.log`
- `logs/complaints/launchd.stderr.log`

## Smoke-check

1. Запустить dry-run и убедиться, что:
  - journal создан;
  - changed-blocks создан;
  - запросы в internal API проходят.
2. Запустить `--apply` на тестовом наборе жалоб.
3. Проверить:
  - вопросы удалены из `courses/*/training_pack/chapters/*.questions.json`;
  - API вернул успешный bulk resolve;
  - в `content_reports` статус сменился на `resolved`.

