# Reading cover images: локальный setup (ComfyUI + LLM)

Обложки для reading texts: **LLM (image prompt) → ComfyUI txt2img → WebP thumb/hero** в `courses/*/assets/reading/<text_id>/`.

См. также: [READING_TEXTS_LOCAL_SETUP.md](./READING_TEXTS_LOCAL_SETUP.md).

## Переменные окружения

| Variable | Default | Назначение |
|----------|---------|------------|
| `COMFYUI_URL` | auto (`8000` then `8188`) | ComfyUI HTTP API |
| `COMFYUI_WORKFLOW` | `scripts/comfyui/reading-cover-z-image-turbo.workflow.json` (if present) | workflow JSON |
| `COMFYUI_CHECKPOINT` | `v1-5-pruned-emaonly.safetensors` | только для legacy SD1.5 workflow |
| `READING_COVER_THUMB_W` / `_H` | `400` / `300` | thumbnail для списка |
| `READING_COVER_HERO_W` / `_H` | `1024` / `768` | hero для экрана текста (как у ComfyUI, без обрезки) |
| `READING_COVER_WEBP_QUALITY` | `85` | качество WebP |
| `READING_COVER_MAX_TOKENS` | `4096` | qwen3:30b cover JSON; reasoning eats tokens before `content` |
| `READING_COVER_STYLE_PREFIX` | watercolor / storybook / casual game (см. код) | English style prefix, добавляется к промпту автоматически |
| `READING_COVER_STYLE_SUFFIX` | _(пусто)_ | опциональный суффикс в конце промпта |
| LLM | `LLAMACPP_URL`, `LLAMACPP_START_CMD_READING` (или `LLAMACPP_START_CMD`) в `.env` / `.env.es` | локальный llama.cpp; CMS подхватывает эти файлы при генерации обложек |
| `LLAMACPP_START_CMD_READING` | — | автозапуск llama.cpp для cover prompt (из `.env.es`) |
| `READING_COVER_USE_AI_URL` | `0` | `1` — промпт обложки через `AI_URL` (OpenRouter), без local llama |

**Reading CMS** (`make reading-cms`) при каждом запуске Python-пайплайна подгружает `.env` + `.env.es` / `.env.en` — те же переменные, что и `make reading-covers-batch`. Перезапуск CMS после правки `.env` не обязателен.

Локальная LLM описывает **только сцену** (кто/где/что); стиль **акварель + storybook + casual game** дописывается в коде на английском — так стабильнее для слабых моделей.

Для тестов без LLM: `READING_COVER_PROMPT_MOCK=1`.

## 1) ComfyUI

**Comfy Desktop** обычно слушает `http://127.0.0.1:8000` (не 8188). Скрипт сам пробует `8000` → `8188`, если `COMFYUI_URL` не задан.

Для **Z-Image-Turbo** (дефолтный workflow в Comfy Desktop) в репо уже есть API workflow:
`scripts/comfyui/reading-cover-z-image-turbo.workflow.json` — он подхватывается автоматически, если файл существует.

Для legacy SD1.5 checkpoint: `scripts/comfyui/reading-cover.workflow.json` + `COMFYUI_CHECKPOINT`.

Проверка:

```bash
curl -s http://127.0.0.1:8188/system_stats | head
```

## 2) Зависимости Python

```bash
pip3 install --user Pillow --break-system-packages
```

LLM — локальный llama.cpp (тот же endpoint, что и `make reading-generate-free`).

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

`make reading-cms` при запуске cover-пайплайна подхватывает `.env` и `.env.es` / `.env.en` (как `make` для Spanish/English). Нужны:

- `LLAMACPP_URL` — URL локального llama-server (например `http://127.0.0.1:8090`)
- `LLAMACPP_START_CMD_READING` или `LLAMACPP_START_CMD` — автозапуск, если сервер ещё не поднят
- `COMFYUI_URL` — Comfy Desktop (`http://127.0.0.1:8000`)

В UI:

- **Сгенерировать** обложку в таблице текстов (`POST /api/published/cover`);
- **Generate cover** на черновике (`POST /api/drafts/:id/cover`);
- **Batch covers** в шапке (`POST /api/covers/batch`).

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
