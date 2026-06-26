#!/usr/bin/env bash
# Stop local ComfyUI / Comfy Desktop processes (by name and known API ports).
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"

set -a
[ -f .env ] && . ./.env
[ -f .env.comfy ] && . ./.env.comfy
set +a

stopped=0

is_comfy_cmd() {
  echo "$1" | grep -qiE 'comfyui|ComfyUI|/comfy/|comfy_extras|comfy_api|ComfyUI\.app'
}

kill_pid() {
  local pid="$1"
  local cmd
  cmd="$(ps -p "$pid" -o command= 2>/dev/null || true)"
  [ -z "$cmd" ] && return 1
  if ! is_comfy_cmd "$cmd"; then
    return 1
  fi
  echo "Stopping ComfyUI (pid $pid): $cmd"
  kill "$pid" 2>/dev/null || true
  stopped=1
  return 0
}

for pattern in 'ComfyUI' 'comfyui/main.py' 'ComfyUI/main.py' 'comfy_extras' 'comfy_api'; do
  while IFS= read -r pid; do
    [ -z "$pid" ] && continue
    kill_pid "$pid" || true
  done < <(pgrep -f "$pattern" 2>/dev/null || true)
done

ports="8000 8188"
if [ -n "${COMFYUI_URL:-}" ]; then
  extra_port="$(echo "$COMFYUI_URL" | sed -nE 's#.*:([0-9]+)(/|$).*#\1#p')"
  if [ -n "$extra_port" ]; then
    ports="$extra_port $ports"
  fi
fi

for port in $ports; do
  for pid in $(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true); do
    kill_pid "$pid" || true
  done
done

if [ "$stopped" -eq 0 ]; then
  echo "No ComfyUI processes found."
  exit 0
fi

sleep 0.5
for pattern in 'ComfyUI' 'comfyui/main.py' 'ComfyUI/main.py' 'comfy_extras' 'comfy_api'; do
  while IFS= read -r pid; do
    [ -z "$pid" ] && continue
    cmd="$(ps -p "$pid" -o command= 2>/dev/null || true)"
    if is_comfy_cmd "$cmd"; then
      echo "Force stop ComfyUI (pid $pid)"
      kill -9 "$pid" 2>/dev/null || true
    fi
  done < <(pgrep -f "$pattern" 2>/dev/null || true)
done
echo "✅ ComfyUI stopped"
