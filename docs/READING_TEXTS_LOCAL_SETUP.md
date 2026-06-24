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
3. **Import JSON** — готовый JSON **без LLM**: назначение голосов, TTS, publish (кнопки «Скопировать промпт» — course-specific промпт для внешней модели).
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