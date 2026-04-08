#!/usr/bin/env bash
set -euo pipefail

# Soft repair for invalid data by running cmd/revalidate_training_cards.
# - invalid training cards: delete + reset corresponding word_card for regenerate
# - duplicate training cards: delete duplicate rows only (no word_card reset)
# - invalid tts statuses: reset tts status to pending only for invalid words
#
# Defaults:
# - target lang: es
# - mode: dry-run
#
# Examples:
#   scripts/requeue_invalid_training_cards.sh
#   scripts/requeue_invalid_training_cards.sh --commit
#   scripts/requeue_invalid_training_cards.sh --commit --limit 500
#   scripts/requeue_invalid_training_cards.sh --commit --only-word castellano

TARGET_LANG="${TARGET_LANG:-es}"
COMMIT=false
LIMIT=""
ONLY_WORD=""
CHECK_TTS=true

while [[ $# -gt 0 ]]; do
  case "$1" in
    --commit)
      COMMIT=true
      shift
      ;;
    --limit)
      LIMIT="${2:-}"
      shift 2
      ;;
    --only-word)
      ONLY_WORD="${2:-}"
      shift 2
      ;;
    --target-lang)
      TARGET_LANG="${2:-}"
      shift 2
      ;;
    --no-tts)
      CHECK_TTS=false
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

ENV_LANG_FILE=".env.${TARGET_LANG}"
if [[ ! -f "${ENV_LANG_FILE}" ]]; then
  echo "Missing ${ENV_LANG_FILE}. Create it from env.example.${TARGET_LANG}" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "${ENV_LANG_FILE}"
set +a

CMD=(go run ./cmd/revalidate_training_cards)
if [[ "${COMMIT}" == "true" ]]; then
  CMD+=(--commit)
fi
if [[ -n "${LIMIT}" ]]; then
  CMD+=(--limit "${LIMIT}")
fi
if [[ -n "${ONLY_WORD}" ]]; then
  CMD+=(--only-word "${ONLY_WORD}")
fi
if [[ "${CHECK_TTS}" != "true" ]]; then
  CMD+=(--check-tts=false)
fi

echo "Running: ${CMD[*]}"
"${CMD[@]}"
