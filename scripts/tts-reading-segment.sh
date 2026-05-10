#!/usr/bin/env bash
set -euo pipefail

VOICE_ID=""
TEXT=""
OUTPUT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --voice-id)
      VOICE_ID="$2"
      shift 2
      ;;
    --text)
      TEXT="$2"
      shift 2
      ;;
    --output)
      OUTPUT="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 2
      ;;
  esac
done

if [[ -z "$VOICE_ID" || -z "$TEXT" || -z "$OUTPUT" ]]; then
  echo "Usage: $0 --voice-id <voice_id> --text <text> --output <output.mp3>" >&2
  exit 2
fi

PIPER_BIN="${PIPER_BIN:-$HOME/bin/piper}"
VOICE_DIR="${READING_VOICE_DIR:-$HOME/tts/voices}"
VOICE_MODEL="$VOICE_DIR/$VOICE_ID.onnx"
ESPEAK_DATA_PATH="${ESPEAK_DATA_PATH:-$HOME/tts/local/espeak-ng-data}"

if [[ ! -x "$PIPER_BIN" ]]; then
  echo "piper binary not found or not executable: $PIPER_BIN" >&2
  exit 1
fi

if [[ ! -f "$VOICE_MODEL" ]]; then
  echo "voice model not found: $VOICE_MODEL" >&2
  exit 1
fi

if [[ ! -f "${VOICE_MODEL}.json" ]]; then
  echo "voice config not found: ${VOICE_MODEL}.json" >&2
  exit 1
fi

if ! command -v ffmpeg >/dev/null 2>&1; then
  echo "ffmpeg not found in PATH" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUTPUT")"
TMP_WAV="$(mktemp "/tmp/reading-tts-XXXXXX.wav")"
trap 'rm -f "$TMP_WAV"' EXIT

printf '%s' "$TEXT" | ESPEAK_DATA_PATH="$ESPEAK_DATA_PATH" "$PIPER_BIN" \
  --model "$VOICE_MODEL" \
  --output_file "$TMP_WAV" >/dev/null

ffmpeg -y -loglevel error -i "$TMP_WAV" -codec:a libmp3lame -qscale:a 4 "$OUTPUT"

