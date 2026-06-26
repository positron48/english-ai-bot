#!/usr/bin/env bash
set -euo pipefail

LLAMA_URL="${LLAMACPP_URL:-http://127.0.0.1:8090}"
START_CMD="${LLAMACPP_START_CMD:-}"
MAX_WAIT_SEC="${LLAMACPP_START_MAX_WAIT_SEC:-45}"

check_ready() {
  curl -fsS "${LLAMA_URL%/}/v1/models" >/dev/null 2>&1 && return 0
  curl -fsS "${LLAMA_URL%/}/health" >/dev/null 2>&1 && return 0
  return 1
}

if check_ready; then
  echo "✅ llama.cpp already running at ${LLAMA_URL}"
  exit 0
fi

if [[ -z "$START_CMD" ]]; then
  echo "❌ llama.cpp is not reachable at ${LLAMA_URL}" >&2
  echo "   Set LLAMACPP_URL to a running server, or define LLAMACPP_START_CMD for auto-start." >&2
  exit 1
fi

echo "⚙️  llama.cpp is not reachable at ${LLAMA_URL}, starting via LLAMACPP_START_CMD..."
bash -lc "$START_CMD"

for ((i=1; i<=MAX_WAIT_SEC; i++)); do
  if check_ready; then
    echo "✅ llama.cpp is ready at ${LLAMA_URL}"
    exit 0
  fi
  sleep 1
done

echo "❌ llama.cpp did not become ready within ${MAX_WAIT_SEC}s at ${LLAMA_URL}" >&2
exit 1

