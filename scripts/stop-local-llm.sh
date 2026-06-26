#!/usr/bin/env bash
# Stop local llama.cpp server processes (llama-server on any port).
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"

set -a
[ -f .env ] && . ./.env
[ -f .env.es ] && . ./.env.es
[ -f .env.en ] && . ./.env.en
set +a

stopped=0

kill_pid() {
  local pid="$1"
  local cmd
  cmd="$(ps -p "$pid" -o command= 2>/dev/null || true)"
  [ -z "$cmd" ] && return 0
  echo "Stopping llama.cpp (pid $pid): $cmd"
  kill "$pid" 2>/dev/null || true
  stopped=1
}

for pattern in 'llama-server' 'llama_cpp\.server'; do
  while IFS= read -r pid; do
    [ -z "$pid" ] && continue
    kill_pid "$pid"
  done < <(pgrep -f "$pattern" 2>/dev/null || true)
done

url="${LLAMACPP_URL:-http://127.0.0.1:8090}"
port="$(echo "$url" | sed -nE 's#.*:([0-9]+)(/|$).*#\1#p')"
port="${port:-8090}"

for pid in $(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true); do
  cmd="$(ps -p "$pid" -o command= 2>/dev/null || true)"
  if echo "$cmd" | grep -qiE 'llama'; then
    kill_pid "$pid"
  fi
done

if [ "$stopped" -eq 0 ]; then
  echo "No llama.cpp processes found."
  exit 0
fi

sleep 0.5
for pattern in 'llama-server' 'llama_cpp\.server'; do
  while IFS= read -r pid; do
    [ -z "$pid" ] && continue
    echo "Force stop llama.cpp (pid $pid)"
    kill -9 "$pid" 2>/dev/null || true
  done < <(pgrep -f "$pattern" 2>/dev/null || true)
done
echo "✅ llama.cpp stopped"
