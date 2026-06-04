# DB-first content migration

This runbook covers the first DB-first content step for grammar course content, grammar training packs, reading catalog, and speaking catalog.

## What changed

- Runtime keeps using embedded bundles by default: `CONTENT_SOURCE=bundle`.
- Imported DB content can be enabled with `CONTENT_SOURCE=db` after a successful import.
- The importer is idempotent per `GRAMMAR_BUNDLE_ID`: it replaces imported grammar content rows for that bundle and refreshes reading/speaking catalog rows, but does not touch user progress or publish state.

Existing user data stays in the current tables:

- `grammar_published_items`
- `grammar_progress`
- `grammar_test_attempts`
- `grammar_theory_memory`
- `grammar_attempts`
- `reading_text_progress`
- `speaking_sessions`
- `speaking_attempts`

## Dry-run

English:

```bash
go run ./cmd/import_learning_content
```

Spanish:

```bash
LEARNING_PAIR=ru-es \
LEARNING_NATIVE_LANG=ru \
LEARNING_TARGET_LANG=es \
LEARNING_APP_CODE=spanish \
GRAMMAR_BUNDLE_ID=es \
go run ./cmd/import_learning_content
```

Expected current source counts:

- English: `sections=19`, `chapters=130`, `training_questions=6901`
- Spanish: `sections=23`, `chapters=152`, `training_questions=5375`

## Commit import

Load the same env chain used by the target instance, then run with `--commit`.

English local example:

```bash
set -a; [ -f .env ] && . ./.env; set +a
set -a && . ./.env.en && set +a
go run ./cmd/import_learning_content --commit
```

Spanish local example:

```bash
set -a; [ -f .env ] && . ./.env; set +a
set -a && . ./.env.es && set +a
go run ./cmd/import_learning_content --commit
```

In k3s, run the binary inside the deployed pod:

English:

```bash
kubectl exec -it deployment/english -n english -- /app/import_learning_content --commit
```

Spanish:

```bash
kubectl exec -it deployment/spanish -n spanish -- /app/import_learning_content --commit
```

## Switch runtime

Only after commit import succeeds and API smoke checks match the bundle-backed behavior:

```env
CONTENT_SOURCE=db
```

Recommended checks before and after switching:

```bash
curl -fsS "$WEBAPP_PUBLIC_URL/api/health"
curl -fsS -H "Authorization: Bearer $TOKEN" "$WEBAPP_PUBLIC_URL/api/learning/grammar/categories"
curl -fsS -H "Authorization: Bearer $TOKEN" "$WEBAPP_PUBLIC_URL/api/learning/grammar/offline/manifest"
```

If DB content is missing or invalid, switch back to:

```env
CONTENT_SOURCE=bundle
```
