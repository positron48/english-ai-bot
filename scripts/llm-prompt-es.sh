#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [[ -z "${AI_URL:-}" || -z "${AI_API_KEY:-}" ]]; then
  echo "AI_URL and AI_API_KEY are required."
  echo "Example:"
  echo "  AI_URL=https://openrouter.ai/api/v1 AI_API_KEY=... scripts/llm-prompt-es.sh"
  exit 1
fi

echo "Running Spanish prompt regression tests (word + training)..."
go test -tags=integration -v -run '^TestLLM_WordCards_ES$|^TestLLM_TrainingCards_ES$' ./internal/integration/llm/...
echo "Done."

