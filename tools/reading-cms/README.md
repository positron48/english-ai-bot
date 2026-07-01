# Reading Texts CMS (local)

Dev-only authoring tool for Linglow reading texts. Not deployed to production.

Detailed runbook: [docs/READING_TEXTS_LOCAL_SETUP.md](../../docs/READING_TEXTS_LOCAL_SETUP.md).

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

- `GET /api/drafts` — list drafts (filters: `course_code`, `level`, `status`, `audio`, `cover`, `search`)
- `POST /api/drafts/generate` — LLM batch generation
- `POST /api/drafts/import-json` — ready JSON → voices + TTS + optional publish (no LLM)
- `POST /api/drafts/import-json-batch` — ready JSON batch → voices + TTS + optional publish (no LLM)
- `POST /api/prompts/reading` — course-specific generate/transform prompt text for external LLM
- `GET /api/drafts/:id` — draft + document
- `POST /api/drafts/:id/approve|reject|audio|publish`
- `DELETE /api/drafts/:id`
- `GET /api/published` — texts in course catalog (filters: `course_code`, `level`, `cover`, `git`, `search`; audio/cover/git status, thumb paths)
- `POST /api/published/sync` — `{course_code, level?, cover?, git?, search?, force?}` import course texts into CMS drafts
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
- `tools/reading-cms/batches/` — import-ready batch JSON files

## Text generation / import

CMS supports three input paths:

1. `Generate` creates texts with the course LLM prompt in `courses/*/scripts/generate-reading-text.py`.
2. `Import plain text` sends pasted prose/dialogue through the LLM transform prompt, then voices/TTS it.
3. `Import JSON` skips the LLM. Paste one JSON object, a JSON array, consecutive JSON objects, or load a `.json` file; CMS assigns voices, generates TTS, and can publish/sync.

Import-ready JSON uses the generator schema:

```json
{
  "title_short": "La canción de Lucía",
  "level": "A1",
  "segments": [
    {
      "segment_id": "s1",
      "speaker_id": "speaker_a",
      "speaker_gender": "female",
      "text": "Lucía canta en la plaza.",
      "text_translation_ru": "Лусия поёт на площади."
    }
  ],
  "vocab_focus": ["canta", "plaza"],
  "questions": [
    {
      "id": "q1",
      "prompt": "Лусия поёт на площади.",
      "type": "true_false",
      "correct_answer": "true",
      "explanation": "В тексте сказано: canta en la plaza."
    }
  ]
}
```

Example batch file: [batches/es_ru_npc_stories.json](./batches/es_ru_npc_stories.json).
