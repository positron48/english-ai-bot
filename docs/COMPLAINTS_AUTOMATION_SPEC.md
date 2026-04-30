# Complaints Automation Spec (Release + Local)

## 1) Разделение контуров

- **Release (Go-сервис):** только internal API для чтения активных grammar-жалоб и массового `resolve`.
- **Local (Mac):** отдельный инструмент для группировки, анализа через `llama.cpp`, удаления вопросов из `courses/`, ведения журнала и hourly запуска.

---

## 2) Что реализовано в release-контуре

### Internal API

- `GET /api/internal/content-reports/grammar`
  - Авторизация: `X-Service-Token`
  - Возвращает активные жалобы `source_type=grammar_training`
  - Поддерживает фильтры:
    - `course` (`en|english|es|spanish`)
    - `chapter_id`
    - `theory_block_id`
    - `cursor` (по `id`, для пагинации)
    - `limit`
- `POST /api/internal/content-reports/grammar/resolve-bulk`
  - Авторизация: `X-Service-Token`
  - Тело:
    - `report_ids: []int64`
    - `reason: string`
  - Массово переводит жалобы в `resolved`.

### Репозиторий и БД

- Добавлены методы:
  - `ListActiveGrammarReports(...)`
  - `ResolveBulk(...)`
- Добавлена миграция индексов:
  - `internal/database/migrations/000008_content_reports_internal_indexes.sql`

### Конфиг токенов internal API

- Новый конфиг:
  - `webapp.complaints_service_token`
  - `COMPLAINTS_SERVICE_TOKEN` (основной для релиза)
  - `webapp.internal_service_tokens_json`
- Дополнительно (backward compatible):
  - `WEBAPP_INTERNAL_SERVICE_TOKENS_JSON`
- Формат: JSON map токенов (например `{"default":"token123","en":"...","es":"..."}`).

### Файлы release-части

- `internal/web/content_reports.go`
- `internal/web/router.go`
- `internal/repository/content_report_repository.go`
- `internal/config/config.go`
- `internal/database/migrations/000008_content_reports_internal_indexes.sql`

### Тесты release-части

- `internal/repository/content_report_repository_test.go`
- `internal/web/content_reports_internal_test.go`

---

## 3) Что реализовано в local-контуре

### Локальный инструмент

- `tools-local/complaints-worker/worker.py`
- Обертка запуска: `tools-local/complaints-worker/run.sh`

### Что делает worker

1. Получает жалобы через internal API.
2. Группирует по `course -> chapter_id -> theory_block_id`.
3. Вызывает `llama.cpp` (OpenAI-compatible endpoint `/v1/chat/completions`).
4. Удаляет вопросы из:
  - `courses/*-grammar/training_pack/chapters/*.questions.json`
5. Отправляет bulk resolve в API (в боевом режиме).
6. Пишет журнал и артефакты.

### Режимы

- По умолчанию: **dry-run**
- Боевой режим: `--apply`

---

## 4) Журнал и артефакты

### JSONL журнал

- Путь: `logs/complaints/complaints-YYYY-MM.jsonl`

Поля записи:

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

### Snapshot затронутых блоков

- Путь: `logs/complaints/changed-theory-blocks-YYYYMMDDHH.json`
- Используется как вход для последующей пакетной ревалидации/перегенерации.

---

## 5) Hourly запуск на Mac (launchd)

Шаблон:

- `tools-local/complaints-worker/launchd/com.englishai.complaints-worker.plist`

Ключевые параметры:

- `StartInterval = 3600`
- запуск `tools-local/complaints-worker/run.sh`
- stdout/stderr в `logs/complaints/`

---

## 6) Минимальные команды

### Dry-run

```bash
python3 tools-local/complaints-worker/worker.py
```

### Боевой запуск

```bash
python3 tools-local/complaints-worker/worker.py --apply
```

### Проверка launchd

```bash
launchctl list | rg complaints-worker
```

---

## 7) Где смотреть подробный runbook

- `docs/LOCAL_COMPLAINTS_WORKER.md`