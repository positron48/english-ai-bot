# Reading Texts CMS (local)

Dev-only authoring tool for Linglow reading texts. Not deployed to production.

## Start

```bash
make reading-cms
# or
go run ./cmd/reading_cms -port 8791
```

Open http://127.0.0.1:8791/

## API (local)

- `GET /api/drafts` — list drafts (filters: `course_code`, `level`, `status`, `audio`, `search`)
- `POST /api/drafts/generate` — LLM batch generation
- `POST /api/drafts/import-json` — ready JSON → voices + TTS + optional publish (no LLM)
- `POST /api/prompts/reading` — course-specific generate/transform prompt text for external LLM
- `GET /api/drafts/:id` — draft + document
- `POST /api/drafts/:id/approve|reject|audio|publish`
- `DELETE /api/drafts/:id`
- `GET /api/published` — texts in course catalog
- `DELETE /api/published?course_code=&text_id=`

## Code layout

- `cmd/reading_cms/` — HTTP server entrypoint
- `internal/readingcms/` — store, pipeline, publish/delete
- `tools/reading-cms/web/` — static UI
