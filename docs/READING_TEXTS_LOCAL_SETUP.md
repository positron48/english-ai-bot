# Reading Texts: локальный запуск (EN + ES)

Этот runbook описывает:

- **локальную CMS** для authoring (`make reading-cms`);
- как подготовить локальные голоса;
- как запускать генерацию reading-текста через `make`;
- как работает выбор голоса (роль LLM vs конфиг голосов);
- как проверить, что артефакты корректные и попадут в bundle/deploy.
- **обложки reading texts** (ComfyUI + LLM): см. [READING_IMAGES_LOCAL_SETUP.md](./READING_IMAGES_LOCAL_SETUP.md).

## 0) Локальная CMS (отдельно от `/app/admin`)

Отдельное dev-only приложение для управляемого конвейера: генерация → ревью → озвучка → публикация в `courses/*/reading/`.

Запуск:

```bash
cd /Users/antonfilatov/www/my/k3s/english-ai-bot
make reading-cms
```

Откроется `http://127.0.0.1:8791/` (порт можно сменить: `go run ./cmd/reading_cms -port 8792`).

Данные CMS:

- состояние черновиков: `.local/reading-cms/drafts/`
- staging аудио: `.local/reading-cms/staging/<text_id>/assets/reading/...`

Workflow:

1. **Generate** — batch LLM (`COUNT`, `LEVEL`, `FORMAT`); опционально TTS сразу.
2. **Import plain text** — вставить готовый текст; LLM (`--input-text`) преобразует в сегменты/переводы/вопросы/диалоги, по умолчанию сразу TTS и **auto-publish** в course (можно отключить чекбоксами в UI).
3. **Import JSON** — готовый JSON **без LLM**: один объект, массив объектов, несколько объектов подряд или `.json` файл; CMS назначает голоса, делает TTS и publish (кнопки «Скопировать промпт» — course-specific промпт для внешней модели).
4. **Preview** — открыть черновик, при необходимости править JSON.
5. **Approve / Generate audio / Publish** — для generate-flow или JSON-import без автопубликации.
6. **Delete** — удаление из course/bundle.

После публикации:

```bash
make -C courses/english-grammar reading-validate   # или spanish-grammar
make grammar-bundle
make check
```

CMS **не** встроена в prod runtime и **не** требует admin login.

### 0.1 Как CMS генерит / импортирует reading-тексты

Есть три режима:

1. **Generate**: UI вызывает `POST /api/drafts/generate`; backend запускает `courses/<course>/scripts/generate-reading-text.py` без входного текста. Скрипт строит course-specific LLM prompt, получает JSON (`title_short`, `segments`, `vocab_focus`, `questions`), назначает голоса и пишет временный reading catalog. CMS переносит результат в `.local/reading-cms/drafts/`.
2. **Import plain text**: UI вызывает `POST /api/drafts/import-text`; backend запускает тот же `generate-reading-text.py --input-text`. LLM сохраняет смысл исходного текста, но раскладывает его в сегменты, переводы, вопросы и словарь.
3. **Import JSON / batch JSON**: UI вызывает `POST /api/drafts/import-json` или `POST /api/drafts/import-json-batch`; LLM не используется. CMS принимает уже готовый JSON, запускает `generate-reading-text.py --input-json`, чтобы назначить `voice_id`, `audio_rel_path`, токены, сгенерировать TTS и опционально сразу опубликовать.

Batch JSON в UI принимает:

- один JSON-объект;
- JSON-массив объектов;
- несколько JSON-объектов подряд;
- `.json` файл через file picker.

Готовый пример для NPC: `tools/reading-cms/batches/es_ru_npc_stories.json`.

Минимальная схема входного JSON:

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

Важно:

- `text` всегда на изучаемом языке курса (`es` для `es_ru`, `en` для `en_ru`);
- `text_translation_ru`, `questions.prompt`, `questions.explanation` — на русском;
- для диалогов используйте стабильные `speaker_id` (`speaker_a`, `speaker_b`) и `speaker_gender`;
- для нарратива можно использовать `speaker_id: "narrator"` и импортировать с `Format = narrative`;
- итоговый опубликованный документ уже будет иметь полный формат `reading/texts/<text_id>.json`: `id`, `category_id`, `reading_passage.segments[].voice_id`, `audio_rel_path`, `tokens`.

## 1) Как устроен выбор голоса

Коротко: **LLM выбирает пол/роль реплики, а не путь к модели**.

- LLM генерирует сегменты с полями:
  - `speaker_id` (кто говорит);
  - `speaker_gender` (`female|male|neutral`).
- Оркестратор `generate-reading-text.py` дальше:
  1. проверяет `speakers` map в `reading-voices.json`;
  2. если явного mapping нет — выбирает `voice_id` из gender-пула:
    - `female_voice_ids`
    - `male_voice_ids`
    - `neutral_voice_ids`
- Для одного и того же `speaker_id` фиксируется один и тот же `voice_id` (консистентный голос по всему тексту).
- Если в profile есть `*_2` голоса, но файлов пока нет, генератор автоматически их пропустит (по `READING_VOICE_DIR`).

Файлы профилей:

- `courses/english-grammar/config/reading-voices.json`
- `courses/spanish-grammar/config/reading-voices.json`

## 2) Рекомендуемая схема мужских/женских голосов

Актуальный профиль (после перехода на gender-aware):

- Для каждого языка задаются пулы:
  - `female_voice_ids`
  - `male_voice_ids`
  - `neutral_voice_ids` (обычно narrator)
- LLM в сегменте указывает `speaker_gender`, а генератор выбирает голос из нужного пула и закрепляет его за `speaker_id`.
- В `speakers` стоит держать только явные фиксированные роли (обычно `narrator`), а диалоговые роли (`speaker_a`, `speaker_b`, ...) распределять автоматически через gender-пулы.

Если нужно больше вариативности, добавляйте `*_2`, `*_3` в пулы и новые роли (`speaker_c`, `speaker_d`) в LLM-диалог.

## 3) Установка Piper и голосов

Базовый runbook:

- `../tts-worker/LOCAL_TTS_WORKER_SETUP.md`

Быстрый путь:

1. Установить зависимости:

```bash
brew install ffmpeg cmake automake autoconf libtool pkg-config onnxruntime espeak-ng
```

1. Установить/собрать `piper` (как в runbook выше).
2. Создать папку голосов:

```bash
mkdir -p "$HOME/tts/voices"
```

1. Скачать набор голосов (EN/ES, male/female + second voices) одной командой:

```bash
cd /Users/antonfilatov/www/my/k3s/english-ai-bot
./scripts/download-reading-voices.sh
```

По умолчанию скачивает в `~/tts/voices`.

## 4) Именование голосов под Reading

Скрипт `scripts/tts-reading-segment.sh` ожидает модель по шаблону:

- `"$READING_VOICE_DIR/<voice_id>.onnx"`
- `"$READING_VOICE_DIR/<voice_id>.onnx.json"`

Скрипт `download-reading-voices.sh` уже сохраняет файлы сразу в нужных именах (`voice_id.onnx`), так что симлинки не нужны.

Alias-голоса, которые используются в profile:

- EN: `en_narrator`, `en_female_1`, `en_female_2`, `en_male_1`, `en_male_2`
- ES: `es_narrator`, `es_female_1`, `es_female_2`, `es_male_1`, `es_male_2`

## 5) Переменные окружения для генератора

Минимум для LLM-генерации:

```bash
export AI_URL="http://localhost:11434/v1"
export AI_API_KEY="dummy"
export AI_MODEL="qwen2.5:14b-instruct"
```

Для TTS helper:

```bash
export PIPER_BIN="$HOME/bin/piper"
export ESPEAK_DATA_PATH="$HOME/tts/local/espeak-ng-data"
export READING_VOICE_DIR="$HOME/tts/voices"
```

## 6) Запуск через Makefile

### 6.0 Быстрый preflight

```bash
export PIPER_BIN="$HOME/bin/piper"
export ESPEAK_DATA_PATH="$HOME/tts/local/espeak-ng-data"
export READING_VOICE_DIR="$HOME/tts/voices"

ls "$READING_VOICE_DIR"/en_narrator.onnx "$READING_VOICE_DIR"/en_female_1.onnx "$READING_VOICE_DIR"/en_male_1.onnx
ls "$READING_VOICE_DIR"/es_narrator.onnx "$READING_VOICE_DIR"/es_female_1.onnx "$READING_VOICE_DIR"/es_male_1.onnx

# Optional voices for richer variety (safe to skip):
ls "$READING_VOICE_DIR"/en_female_2.onnx "$READING_VOICE_DIR"/en_male_2.onnx "$READING_VOICE_DIR"/es_female_2.onnx "$READING_VOICE_DIR"/es_male_2.onnx || true
```

### 6.1 Быстрая произвольная генерация (standalone, без chapter_id)

`TITLE` опционален. Если не задан, будет автосгенерирован `text_id` вида `free_<lang>_<level>_reading_<timestamp>`.

```bash
cd courses/spanish-grammar

make reading-generate-free \
  TARGET_LANG=es \
  LEVEL=A2 \
  FORMAT=narrative
```

Артефакт будет сохранен в standalone-каталоге:

- `reading/texts/<generated_id>.json`
- `reading/index.json` (обновляется автоматически)

### 6.2 Пример standalone для EN

```bash
cd courses/english-grammar

make reading-generate-free \
  TARGET_LANG=en \
  LEVEL=A2 \
  FORMAT=dialogue \
  TITLE="A weekend in London"
```

### 6.3 Режим без LLM (входной JSON, standalone)

```bash
make reading-generate-free \
  TARGET_LANG=en \
  LEVEL=A2 \
  INPUT_JSON=/abs/path/to/reading-input.json
```

`TTS_CMD_TEMPLATE` уже зашит в `Makefile` по умолчанию (через `scripts/tts-reading-segment.sh`).
Переопределять его нужно только если хотите использовать альтернативный TTS-движок.

## 7) Что генерируется

После `reading-generate-free`:

- `reading/texts/<text_id>.json` — самостоятельный reading-текст;
- `reading/index.json` — индекс categories/texts для Reading API;
- `assets/reading/<text_id>/*.mp3` — сегментное аудио;
- `reports/reading-build-report.json` — отчет по генерации.

## 8) Проверки перед коммитом

В каждом курсе:

```bash
make reading-validate
```

В корне репозитория:

```bash
make check
```

`make check` уже включает `reading-validate` для EN и ES.

## 9) Попадание в приложение

Путь доставки standalone Reading:

1. `courses/*` обновлены (`reading/index.json`, `reading/texts/*`, `assets/reading/*`);
2. `make grammar-bundle` копирует `reading/` и `assets/` в embed bundle;
3. `make build` / CI image / deploy;
4. раздел Reading доступен в UI через `reading` API.
