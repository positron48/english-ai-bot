#!/usr/bin/env bash
set -euo pipefail

PROMPT=""
OUTPUT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prompt)
      PROMPT="$2"
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

if [[ -z "$PROMPT" || -z "$OUTPUT" ]]; then
  echo "Usage: $0 --prompt <text> --output <cover_raw.png>" >&2
  exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
READING_COVER_PROMPT="$PROMPT" READING_COVER_OUTPUT="$OUTPUT" python3 -c "
import os, pathlib, sys
sys.path.insert(0, '${SCRIPT_DIR}')
from reading_cover_comfy import generate_png
generate_png(os.environ['READING_COVER_PROMPT'], pathlib.Path(os.environ['READING_COVER_OUTPUT']))
"
