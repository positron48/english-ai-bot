#!/usr/bin/env bash
# Проверка OpenRouter TTS с payload, максимально близким к runtime-коду.
# Не требует запуска всего приложения: только curl + jq (+ ffmpeg/sox для mp3).
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
  echo "Нет .env. Создай из env.example и задай TTS_API_KEY или OPENROUTER_API_KEY."
  exit 1
fi

set -a
. ./.env
set +a

KEY="${TTS_API_KEY:-${OPENROUTER_API_KEY:-}}"
if [ -z "$KEY" ]; then
  echo "В .env нет TTS_API_KEY и нет OPENROUTER_API_KEY."
  exit 1
fi

BASE_URL="${TTS_BASE_URL:-https://openrouter.ai/api/v1}"
MODEL="${TTS_MODEL:-openai/gpt-4o-audio-preview}"
VOICE="${TTS_VOICE:-alloy}"
WORDS="${*:-hello world lettuce amble}"
OUT_DIR="tmp"
mkdir -p "$OUT_DIR"
USER_INSTRUCTION_PREFIX="You are a pronunciation machine. Say ONLY the exact word below as audio. One word, no greeting, no pause, no repetition. Word:"

# PCM 16-bit 24kHz mono (gpt-4o-audio)
RATE=24000
TRIM_FILTER="silenceremove=start_periods=1:start_threshold=-45dB,areverse,silenceremove=start_periods=1:start_threshold=-45dB,areverse"

normalize_like_runtime() {
  # Приближенно к normalizePronunciationTranscript в Go-коде:
  # lowercase + trim quotes + оставить a-z, 0-9, ' и -
  printf '%s' "$1" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E "s/^[\"'\`]+|[\"'\`]+$//g" \
    | tr -cd "a-z0-9'-"
}

for word in $WORDS; do
  echo "--- $word ---"
  echo "  model: $MODEL"
  echo "  voice: $VOICE"
  echo "  base_url: $BASE_URL"
  quoted_word="\`$word\`"
  user_prompt="${USER_INSTRUCTION_PREFIX} ${quoted_word}"
  echo "  user_prompt: $user_prompt"
  echo "  trim_filter: $TRIM_FILTER"
  stream_file="$OUT_DIR/or_stream_${word}.txt"
  pcm_file="$OUT_DIR/openrouter_${word}.pcm"
  mp3_file="$OUT_DIR/openrouter_${word}.mp3"
  transcript_file="$OUT_DIR/or_transcript_${word}.txt"

  curl -sS -X POST "${BASE_URL%/}/chat/completions" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $KEY" \
    -H "Accept: text/event-stream" \
    -d "{
      \"model\": \"${MODEL}\",
      \"messages\": [
        {\"role\": \"user\", \"content\": \"$user_prompt\"}
      ],
      \"modalities\": [\"text\", \"audio\"],
      \"audio\": {\"voice\": \"${VOICE}\", \"format\": \"pcm16\"},
      \"stream\": true,
      \"max_tokens\": 150
    }" -o "$stream_file"

  if [ ! -s "$stream_file" ]; then
    echo "  стрим не получен (пустой или ошибка), пропуск"
    continue
  fi

  : > "$pcm_file"
  : > "$transcript_file"
  while IFS= read -r line; do
    case "$line" in
      data:\ *) ;;
      *) continue ;;
    esac
    json="${line#data: }"
    [ "$json" = "[DONE]" ] && continue

    b64=$(printf '%s' "$json" | jq -r '.choices[0].delta.audio.data // empty' 2>/dev/null)
    [ -n "$b64" ] && printf '%s' "$b64" | base64 -d 2>/dev/null >> "$pcm_file"

    txt_audio=$(printf '%s' "$json" | jq -r '.choices[0].delta.audio.transcript // empty' 2>/dev/null || true)
    [ -n "$txt_audio" ] && printf '%s' "$txt_audio" >> "$transcript_file"
    txt_content=$(printf '%s' "$json" | jq -r '.choices[0].delta.content[]?.text // empty' 2>/dev/null || true)
    [ -n "$txt_content" ] && printf '%s' "$txt_content" >> "$transcript_file"
  done < "$stream_file"

  if [ ! -s "$pcm_file" ]; then
    echo "  нет аудио в стриме"
    # Если вместо SSE пришёл JSON-ошибки от роутера/провайдера, покажем её сразу.
    if jq -e '.error' "$stream_file" >/dev/null 2>&1; then
      err_msg="$(jq -r '.error.message // "unknown error"' "$stream_file" 2>/dev/null || true)"
      err_code="$(jq -r '.error.code // "n/a"' "$stream_file" 2>/dev/null || true)"
      provider="$(jq -r '.error.metadata.provider_name // "n/a"' "$stream_file" 2>/dev/null || true)"
      raw="$(jq -r '.error.metadata.raw // empty' "$stream_file" 2>/dev/null || true)"
      echo "  error: code=${err_code} provider=${provider} message=${err_msg}"
      [ -n "$raw" ] && echo "  provider_raw: $raw"
    fi
    if [ -s "$transcript_file" ]; then
      echo "  transcript: $(cat "$transcript_file")"
    fi
    echo "  raw stream saved: $stream_file"
    rm -f "$pcm_file" "$transcript_file"
    continue
  fi

  transcript="$(cat "$transcript_file" 2>/dev/null || true)"
  if [ -z "$(printf '%s' "$transcript" | tr -d '[:space:]')" ]; then
    echo "  missing transcript for validation (skip)"
    echo "  raw stream saved: $stream_file"
    rm -f "$pcm_file" "$transcript_file"
    continue
  fi
  expected_norm="$(normalize_like_runtime "$word")"
  transcript_norm="$(normalize_like_runtime "$transcript")"
  if [ -n "$transcript" ] && [ "$expected_norm" != "$transcript_norm" ]; then
    echo "  transcript mismatch: expected='$word' got='$transcript' (skip)"
    echo "  raw stream saved: $stream_file"
    rm -f "$pcm_file" "$transcript_file"
    continue
  fi
  rm -f "$stream_file"
  rm -f "$transcript_file"

  echo "  PCM: $(wc -c < "$pcm_file") байт"
  if command -v ffmpeg >/dev/null 2>&1; then
    ffmpeg -y -loglevel error -f s16le -ar $RATE -ac 1 -i "$pcm_file" -af "$TRIM_FILTER" "$mp3_file"
    rm -f "$pcm_file"
    echo "  MP3: $mp3_file"
  elif command -v sox >/dev/null 2>&1; then
    sox -r $RATE -e signed-integer -b 16 -c 1 "$pcm_file" "$mp3_file"
    rm -f "$pcm_file"
    echo "  MP3: $mp3_file"
  else
    echo "  ffmpeg/sox нет — только PCM: $pcm_file (воспроизведение: ffplay -f s16le -ar $RATE -ac 1 $pcm_file)"
  fi
done

echo ""
echo "Готово. OpenRouter: $OUT_DIR/openrouter_*.mp3"
