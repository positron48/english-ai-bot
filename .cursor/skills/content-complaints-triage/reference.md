# Content complaints triage — reference

## Prod credentials

Файл: `secrets/complaints-prod.env` (из `env.example.complaints-prod`).

```bash
kubectl -n english get secret english-secrets -o jsonpath='{.data.COMPLAINTS_SERVICE_TOKEN}' | base64 -d
kubectl -n spanish get secret spanish-secrets -o jsonpath='{.data.COMPLAINTS_SERVICE_TOKEN}' | base64 -d
```

## Internal API (X-Service-Token)

| Method | Path | Назначение |
|--------|------|------------|
| GET | `/api/internal/content-reports` | Список (source_type, course, category, cursor, limit) |
| GET | `/api/internal/content-reports/summary` | Агрегаты по category/chapter |
| GET | `/api/internal/content-reports/{id}` | Контекст для правки |
| POST | `/api/internal/content-reports/resolve-bulk` | Закрыть жалобы |
| PUT | `/api/internal/training/card/{id}` | JSON: word_ru, meaning_en, example_*, hint, pos, … |
| POST | `/api/internal/tts/regenerate` | `{"word":"..."}` |
| GET | `/api/internal/tts/status?word=` | Статус TTS |

Legacy (совместимость): `/api/internal/content-reports/grammar`, `.../grammar/resolve-bulk`.

## Report categories

**word_training:** wrong_translation, wrong_example, wrong_distractors, typo, bad_audio, unclear_question, other

**grammar_training:** wrong_answer, ambiguous, wrong_explanation, theory_mismatch, typo, too_hard, other

## Grammar sync (files → bundle → DB)

Источник правды: репозитории курсов (submodules):

- `courses/english-grammar`
- `courses/spanish-grammar`

Вопросы Grammar Training: `courses/{en|es}-grammar/training_pack/` + `training_pack/index.json` (`blocks`: `chapter_id::theory_block_id` → rel path).

В ответе `GET .../content-reports/{id}` для grammar есть:

- `training_pack_relpath` — путь в embedded pack
- `courses_training_pack_hint` — подсказка пути в courses repo

### Цепочка релиза

1. Правка JSON в **courses** submodule → commit + push courses repo.
2. В `english-ai-bot`:
   ```bash
   make grammar-bundle   # или ./scripts/generate-grammar-bundle.sh en|es
   ```
   Commit `internal/grammarbundle/*`, `internal/grammartrainingpack/*`.
3. Push `english-ai-bot` → CI → Flux rollout.
4. После Ready pod:
   ```bash
   # English
   kubectl exec -it deployment/english -n english -- /app/import_learning_content --commit
   # Spanish
   kubectl exec -it deployment/spanish -n spanish -- /app/import_learning_content --commit
   ```
   (env в pod уже задан: `GRAMMAR_BUNDLE_ID`, `LEARNING_*`.)

5. Smoke:
   ```bash
   curl -fsS "$URL/health"
   ```
   Проверить grammar categories / offline manifest. На prod ожидается `CONTENT_SOURCE=db` — без import пользователи увидят старый контент.

**Не правь** только embedded JSON в образе без courses — следующий `grammar-bundle` затрёт.

Документация: `docs/DB_FIRST_CONTENT_MIGRATION.md`.

## Word / TTS fixes

- Карточка: internal PUT training card (см. admin поля в `internal/web/admin.go`).
- Озвучка: `POST /api/internal/tts/regenerate`, затем status pending/ready.
- Массовые однотипные ошибки: grep training_cards / word_cards по word_category, pos.

## Журнал триажа

Полный runbook: **`docs/complaints/README.md`**.

### Текстовый (в git, по датам)

Канонический путь:

`docs/complaints/journal-YYYY-MM-DD-<slug>.md`

Создать перед apply:

```bash
make complaints-journal-new
make complaints-journal-new SLUG=en
make complaints-journal-new JOURNAL_DATE=2026-06-15 SLUG=hotfix
```

Шаблон: `docs/complaints/journal-TEMPLATE.md`. Индекс прогонов: таблица в `docs/complaints/README.md`.

В каждом блоке жалобы:

- **дата** (`created_at` из snapshot);
- **на что жалоба** — word/grammar, id, комментарий, question_id / word;
- **что изменено** — courses, prod API, «без правки».

Пример: `docs/complaints/journal-2026-06-04-triage.md`.

После релиза — в шапке журнала: тег (`make tag`), коммиты submodule.

### JSONL (локально, не в git)

`logs/complaints/triage-YYYY-MM.jsonl` — `append_triage_log.py` или вывод `resolve_all_active.py`.

### Локальные снимки (не в git)

`logs/complaints/snapshot-*.json`, `clusters-*.json` — из `make complaints-fetch-*`.

## Makefile helpers

```bash
make complaints-journal-new      # новый docs/complaints/journal-*.md
make complaints-triage-dry-en    # fetch + cluster EN
make complaints-triage-dry-es    # fetch + cluster ES
make complaints-fetch-en         # только snapshot
```

Prod apply:

```bash
set -a && . ./secrets/complaints-prod.env && set +a
python3 tools-local/complaints-triage/resolve_all_active.py en --report-ids "$VERIFIED_REPORT_IDS" --reason "Проверено после импорта"
```

## Полнота выборки и чтение

Сборщик всегда читает все страницы без серверного `course`: старые версии API отбрасывают `reading_text` с ID `free_es_*`. Курс определяется локально; неизвестные остаются видимыми в `unknown_course_report_ids`. Сводка строится из снимка. Проверять `complete`; legacy API не подтверждает отсутствие жалоб. Оба URL могут возвращать общую БД — не считать одинаковые ID дважды.

Чтение: канонические `courses/*-grammar/reading/texts`, индекс `reading/index.json` и `assets/reading`. Проверять текст, перевод, вопросы, ключи, аудио и соответствие обложки описанию; выполнить поиск аналогичных ошибок. `GET .../{id}` на обновлённом API отдаёт `reading_text` из текущей БД отдельно от исторического payload.

Для unified Linglow после здорового rollout оператор запускает `kubectl -n linglow exec deploy/linglow -- /app/import_learning_content --course-code es_ru --commit`. Локальный kubectl не использовать (AGENTS.md). Закрывать только явно указанные ID после проверки production; `resolve_all_active.py` без `--apply` ничего не изменяет.

Канонический production URL для обоих курсов: `https://linglow.qantrix.ru`. Старые домены перенаправляют GET, но POST resolve через urllib завершается HTTP 308. В `COMPLAINTS_SERVICE_URL_EN` и `COMPLAINTS_SERVICE_URL_ES` задавать канонический URL. Health endpoint: `/health`.
