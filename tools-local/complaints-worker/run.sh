#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
WORKER="$ROOT_DIR/tools-local/complaints-worker/worker.py"

if [[ ! -f "$WORKER" ]]; then
  echo "worker.py not found: $WORKER" >&2
  exit 1
fi

exec python3 "$WORKER" "$@"

