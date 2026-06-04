# Complaints automation (agent skill first)

## Рекомендуемый контур (2026)

- **Триаж:** Cursor skill [`.cursor/skills/content-complaints-triage/`](../.cursor/skills/content-complaints-triage/SKILL.md) + [reference.md](../.cursor/skills/content-complaints-triage/reference.md).
- **Fetch без LLM:** `make complaints-fetch-en` / `make complaints-fetch-es` → `tools-local/complaints-triage/fetch_reports.py`.
- **Prod credentials:** `secrets/complaints-prod.env` (шаблон `secrets/complaints-prod.env.example`).

## Release internal API

Авторизация: header `X-Service-Token` (`COMPLAINTS_SERVICE_TOKEN` или `WEBAPP_INTERNAL_SERVICE_TOKENS_JSON`).

| Method | Path | Описание |
|--------|------|----------|
| GET | `/api/internal/content-reports` | Все типы: `source_type`, `course`, `category`, `status`, `cursor`, `limit` |
| GET | `/api/internal/content-reports/summary` | Агрегаты по category/chapter |
| GET | `/api/internal/content-reports/{id}` | Контекст: cards, TTS, training_pack_relpath |
| POST | `/api/internal/content-reports/resolve-bulk` | Массовый resolve |
| PUT | `/api/internal/training/card/{id}` | Правка training card (JSON) |
| POST | `/api/internal/tts/regenerate` | `{"word":"..."}` |
| GET | `/api/internal/tts/status?word=` | Статус TTS |

Legacy (совместимость):

- `GET /api/internal/content-reports/grammar`
- `POST /api/internal/content-reports/grammar/resolve-bulk`

## UI

Попап жалобы: категории (chips) + опциональные подробности. Поле БД: `content_reports.report_category`.

## БД

- Миграция `000010_content_reports_report_category.sql`
- Индекс `(source_type, report_category, status)`

## Grammar sync

Источник правды — `courses/*-grammar`. Релиз: `make grammar-bundle` → deploy → `kubectl exec ... /app/import_learning_content --commit`. См. `docs/DB_FIRST_CONTENT_MIGRATION.md` и skill reference § Grammar sync.

## Deprecated: llama worker

`tools-local/complaints-worker/worker.py`, launchd, `make complaints-dry-*` / `complaints-apply-*` — **не использовать для триажа**. Оставлены для обратной совместимости и отдельных prompt-autofix задач. См. `docs/LOCAL_COMPLAINTS_WORKER.md`.

## Файлы

- `internal/web/content_reports.go`, `content_reports_internal.go`, `content_reports_internal_write.go`
- `internal/repository/content_report_repository.go`
- `internal/database/migrations/000010_content_reports_report_category.sql`
- `webapp/src/components/ContentReportDialog.vue`
- `webapp/src/constants/contentReportCategories.ts`

## Тесты

- `internal/repository/content_report_repository_test.go`
- `internal/web/content_reports_internal_test.go`
- `internal/models/content_report_category_test.go`
