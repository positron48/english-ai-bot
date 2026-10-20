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
| DELETE | `/api/internal/word-cards/{id}` | Предпросмотр или удаление одного слова с проверкой курса и написания |
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

## Удаление слов через internal API

`DELETE /api/internal/word-cards/{id}` использует тот же `X-Service-Token`, что и internal content reports.
Тело JSON обязательно: `course_code` и точное `expected_word` из word card; ID — положительное число.
`dry_run` по умолчанию `true`. Опечатки в полях и несколько JSON-объектов отклоняются.

Предпросмотр для Extremadura (после rollout версии с endpoint):

```http
DELETE /api/internal/word-cards/298413
X-Service-Token: <service-token>
Content-Type: application/json

{"course_code":"es_ru","expected_word":"extremadura"}
```

Ответ 200: `word_card_id`, `word`, `course_code`, `training_cards`, `user_cards`, `dry_run:true`, `deleted:false`.
Числа показывают найденные связанные карточки до удаления. Для выполнения повторить с `"dry_run":false`.

Удаляется ровно один word_card с совпадающими ID, курсом и словом. FK каскадно удаляют связанные training/user cards, повторения и другие зависимые записи. Content reports сохраняют исторический payload и получают NULL в ссылках; статус жалоб автоматически не меняется. Файлы TTS этим endpoint не удаляются.
Перед удалением снимаются только ссылки жалоб на удаляемое слово и его карточки; payload и статус сохраняются. Это работает также с ограничивающими legacy FK. Удаление атомарно: ошибка любого шага откатывает и снятие ссылок, и удаление. В журнал приложения пишутся ID, курс и слово.

Ошибки: 400 — неверный ID/JSON или нет обязательных полей; 401 — неверный токен;
404 — ID не существует (в том числе при повторном удалении); 409 — не совпали курс/написание либо цель изменилась перед DELETE;
405 — другой HTTP-метод; 500 — ошибка БД без раскрытия деталей клиенту.

Перед закрытием жалобы проверить detail API: соответствующие карточки отсутствуют, историческая жалоба сохранена.
Для Extremadura отдельная миграция больше не требуется: удаление производится этим API после rollout и явного apply.
