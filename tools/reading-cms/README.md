# Reading Texts CMS (local)

Dev-only authoring tool for Linglow reading texts. Not deployed to production.

## Start

```bash
make reading-cms
# or
go run ./cmd/reading_cms -port 8791
```

`make reading-cms` stops a previous CMS instance on the same port before starting.
To stop without restart: `make reading-cms-stop` (port via `READING_CMS_PORT=8791`).

Open http://127.0.0.1:8791/

## API (local)

- `GET /api/drafts` — list drafts (filters: `course_code`, `level`, `status`, `audio`, `search`)
- `POST /api/drafts/generate` — LLM batch generation
- `POST /api/drafts/import-json` — ready JSON → voices + TTS + optional publish (no LLM)
- `POST /api/prompts/reading` — course-specific generate/transform prompt text for external LLM
- `GET /api/drafts/:id` — draft + document
- `POST /api/drafts/:id/approve|reject|audio|publish`
- `DELETE /api/drafts/:id`
- `GET /api/published` — texts in course catalog (audio/cover status, thumb paths)
- `POST /api/published/sync` — `{course_code, level?, cover?, search?, force?}` import course texts into CMS drafts
- `POST /api/published/cover` — `{course_code, text_id, force}` LLM + ComfyUI for one published text
- `DELETE /api/published/cover` — `{course_code, text_id}` удалить обложку в каталоге
- `POST /api/drafts/:id/cover` — cover for draft staging
- `DELETE /api/drafts/:id/cover` — удалить обложку черновика
- `POST /api/covers/batch` — async batch covers; returns `{ batch_id, total }`; poll `GET /api/cover-batch-progress?batch_id=…`
- `GET /api/course-images/{course}/{textId}/{file}` — cover assets from course dir
- `GET /api/images/{textId}/{file}` — cover assets from draft staging
- `DELETE /api/published?course_code=&text_id=`

## Code layout

- `cmd/reading_cms/` — HTTP server entrypoint
- `internal/readingcms/` — store, pipeline, publish/delete
- `tools/reading-cms/web/` — static UI
