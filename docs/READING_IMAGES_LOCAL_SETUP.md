# Reading cover images: локальный setup (ComfyUI + LLM)

Обложки для reading texts: **LLM (image prompt) → ComfyUI txt2img → WebP thumb/hero** в `courses/*/assets/reading/<text_id>/`.

См. также: [READING_TEXTS_LOCAL_SETUP.md](./READING_TEXTS_LOCAL_SETUP.md).

## Переменные окружения

| Variable | Default | Назначение |
|----------|---------|------------|
| `COMFYUI_URL` | `http://127.0.0.1:8188` | ComfyUI HTTP API |
| `COMFYUI_WORKFLOW` | `scripts/comfyui/reading-cover.workflow.json` | workflow JSON |
| `COMFYUI_CHECKPOINT` | `v1-5-pruned-emaonly.safetensors` | имя checkpoint в `ComfyUI/models/checkpoints/` |
| `READING_COVER_THUMB_W` / `_H` | `400` / `300` | thumbnail для списка |
| `READING_COVER_HERO_W` / `_H` | `1200` / `480` | hero для экрана текста |
| `READING_COVER_WEBP_QUALITY` | `85` | качество WebP |
| `READING_COVER_BATCH_SLEEP_SEC` | `2` | пауза между текстами в batch |
| LLM | `LLAMACPP_URL` / `AI_URL` | как для reading text generation |

Для тестов без LLM: `READING_COVER_PROMPT_MOCK=1`.

## 1) ComfyUI

1. Установите [ComfyUI](https://github.com/comfyanonymous/ComfyUI) локально.
2. Положите checkpoint в `ComfyUI/models/checkpoints/` и задайте `COMFYUI_CHECKPOINT`.
3. Запустите сервер (по умолчанию порт `8188`).
4. При необходимости отредактируйте `scripts/comfyui/reading-cover.workflow.json` (размер latent, steps, negative prompt).

Проверка:

```bash
curl -s http://127.0.0.1:8188/system_stats | head
```

## 2) Зависимости Python

```bash
pip install Pillow
```

LLM — тот же llama.cpp / OpenAI-compatible endpoint, что и для `make reading-generate-free`.

## 3) Один текст

```bash
cd courses/spanish-grammar
set -a; [ -f ../../.env ] && . ../../.env; set +a
python3 scripts/generate-reading-cover.py --text-id <text_id>
```

Или thin wrapper ComfyUI (только PNG, без LLM/JSON):

```bash
bash scripts/generate-reading-cover.sh \
  --prompt "flat illustration of friends in a cafe, warm colors, no text" \
  --output /tmp/cover_raw.png
```

## 4) Batch EN + ES

```bash
cd /path/to/english-ai-bot
set -a; [ -f .env ] && . ./.env; [ -f .env.es ] && . ./.env.es; set +a

# все тексты обоих курсов, skip если cover уже есть
make reading-covers-batch

# force regen одного курса
make -C courses/spanish-grammar reading-covers-batch FORCE=1

# лимит для прогона
make reading-covers-batch LIMIT=3
```

После batch:

```bash
make -C courses/spanish-grammar reading-validate
make -C courses/english-grammar reading-validate
make grammar-bundle
make check
```

Валидация по умолчанию в Makefile использует `--covers-optional` (пока не все тексты с обложками). Для strict-проверки:

```bash
python3 courses/spanish-grammar/scripts/validate-reading-artifacts.py
```

## 5) JSON schema (top-level поля)

```json
{
  "cover_thumb_rel_path": "assets/reading/<text_id>/cover_thumb.webp",
  "cover_hero_rel_path": "assets/reading/<text_id>/cover_hero.webp",
  "cover_image_prompt": "optional English SD prompt for regen/debug"
}
```

Файлы:

```
courses/<course>/assets/reading/<text_id>/
  cover_raw.png      # staging (опционально, можно не коммитить)
  cover_thumb.webp
  cover_hero.webp
```

## 6) Reading CMS

В `make reading-cms` UI:

- **Generate cover** на черновике (`POST /api/drafts/:id/cover`);
- preview thumb/hero + `cover_image_prompt`;
- **Generate all covers** для опубликованных текстов курса (`POST /api/covers/batch`).

Publish копирует `assets/reading/<text_id>/` вместе с аудио.

## 7) Release checklist

1. `courses/*/reading/texts/*.json` + `courses/*/assets/reading/<text_id>/cover_*.webp`
2. `make grammar-bundle`
3. `make check`
4. Git commit + push → CI → Flux rollout `english` / `spanish`

`cover_raw.png` и `.local/reading-cms/` не коммитить.

## 8) Runtime API / webapp

- `GET /api/learning/reading/image?path=assets/reading/...&course_code=es_ru`
- Список текстов и детальный экран отдают `cover_thumb_rel_path` / `cover_hero_rel_path`
- UI: thumbnail в списке, hero над текстом (graceful fallback без обложки)
