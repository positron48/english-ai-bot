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
   curl -fsS "$URL/api/health"
   ```
   Проверить grammar categories / offline manifest. На prod ожидается `CONTENT_SOURCE=db` — без import пользователи увидят старый контент.

**Не правь** только embedded JSON в образе без courses — следующий `grammar-bundle` затрёт.

Документация: `docs/DB_FIRST_CONTENT_MIGRATION.md`.

## Word / TTS fixes

- Карточка: internal PUT training card (см. admin поля в `internal/web/admin.go`).
- Озвучка: `POST /api/internal/tts/regenerate`, затем status pending/ready.
- Массовые однотипные ошибки: grep training_cards / word_cards по word_category, pos.

## JSONL journal (apply mode)

`logs/complaints/triage-YYYY-MM.jsonl`:

```json
{"run_id":"...","course":"en","cluster_key":"...","category":"bad_audio","action":"tts_regenerate","report_ids":[1,2],"files_changed":[],"resolve_status":"ok"}
```

## Makefile helpers

```bash
make complaints-triage-dry-en   # fetch + cluster EN
make complaints-triage-dry-es   # fetch + cluster ES
make complaints-fetch-en        # только snapshot
python3 tools-local/complaints-triage/append_triage_log.py --help
```
