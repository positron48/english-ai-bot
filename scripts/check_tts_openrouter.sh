#!/usr/bin/env bash
# Читает ключ из .env в корне репо и дергает OpenRouter. Запуск: из корня репо ./scripts/check_tts_openrouter.sh
set -e
cd "$(dirname "$0")/.."
if [ ! -f .env ]; then
  echo "Нет .env в $(pwd). Создай из env.example и задай TTS_API_KEY или OPENROUTER_API_KEY."
  exit 1
fi
set -a
. ./.env
set +a
KEY="${TTS_API_KEY:-$OPENROUTER_API_KEY}"
if [ -z "$KEY" ]; then
  echo "В .env нет TTS_API_KEY и нет OPENROUTER_API_KEY."
  exit 1
fi
echo "Ключ найден (длина ${#KEY}). Делаю запросы..."

mkdir -p tmp

echo ""
echo "1) POST /audio/speech (озвучка одного слова)..."
HTTP=$(curl -sS -w "%{http_code}" -o tmp/or_audio_speech.bin -X POST "https://openrouter.ai/api/v1/audio/speech" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $KEY" \
  -d '{"model":"tts-1","input":"hello","voice":"alloy","response_format":"mp3"}')
echo "HTTP $HTTP, размер ответа: $(wc -c < tmp/or_audio_speech.bin) байт"
if [ "$HTTP" = "200" ] && [ -s tmp/or_audio_speech.bin ]; then
  file tmp/or_audio_speech.bin
  echo "Сохранено: tmp/or_audio_speech.bin — можно послушать."
else
  head -c 500 tmp/or_audio_speech.bin | head -c 300
  echo ""
fi

echo ""
echo "2) POST /chat/completions с modalities text,audio (stream)..."
curl -sS -w "\nHTTP %{http_code}\n" -X POST "https://openrouter.ai/api/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $KEY" \
  -H "Accept: text/event-stream" \
  -d '{
    "model": "openai/gpt-4o-audio-preview",
    "messages": [
      {"role": "system", "content": "You are a pronunciation machine. Say ONLY the exact word the user sent, as audio. One word, no greeting, no pause, no repetition."},
      {"role": "user", "content": "hello"}
    ],
    "modalities": ["text", "audio"],
    "audio": {"voice": "alloy", "format": "pcm16"},
    "stream": true,
    "max_tokens": 150
  }' 2>&1 | tee tmp/or_chat_audio.txt | tail -20

echo ""
echo "3) GET /models — модели с output_modalities audio (TTS через chat)..."
curl -sS "https://openrouter.ai/api/v1/models" | jq -r '
  .data[] | select(.architecture.output_modalities | index("audio")) | "\(.id) (audio out)"
' 2>/dev/null | head -30 || echo "(jq не установлен или ответ не JSON — проверь https://openrouter.ai/models?output_modalities=audio)"

echo ""
echo "Готово. Проверь tmp/or_audio_speech.bin и tmp/or_chat_audio.txt."
