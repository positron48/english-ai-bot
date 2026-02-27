#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

BASE_URL_DEFAULT="https://api.dictionaryapi.dev/api/v2/entries/en"
TIMEOUT_DEFAULT=15
TOTAL_WORDS_DEFAULT=40
DELAYS_DEFAULT="0,100,200,300,400,500,600,700,800,900,1000,1200"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

BASE_URL="${TTS_DICTIONARY_BASE_URL:-$BASE_URL_DEFAULT}"
TIMEOUT="${TIMEOUT_SECONDS:-$TIMEOUT_DEFAULT}"
TOTAL_WORDS="${TOTAL_WORDS:-$TOTAL_WORDS_DEFAULT}"
DELAYS_CSV="${DELAYS_MS:-$DELAYS_DEFAULT}"
VERBOSE="${VERBOSE:-1}"

WORDS=(
  servant chaos powder slice occupation addition bacteria dragon sacrifice fitness
  turkey cliff shore mommy gulf kit pastor stimulus spectrum transmission
)

IFS=',' read -r -a DELAYS <<< "$DELAYS_CSV"

sleep_ms() {
  local ms="$1"
  if [[ "$ms" -le 0 ]]; then
    return 0
  fi
  local sec=$((ms / 1000))
  local rem=$((ms % 1000))
  sleep "$(printf "%d.%03d" "$sec" "$rem")"
}

test_delay() {
  local delay_ms="$1"
  local total=0
  local lookup_429=0
  local lookup_timeout=0
  local lookup_other=0
  local audio_429=0
  local audio_timeout=0
  local audio_other=0
  local no_audio=0

  echo ""
  echo "== delay=${delay_ms}ms =="

  for ((i=0; i<TOTAL_WORDS; i++)); do
    local word="${WORDS[$((i % ${#WORDS[@]}))]}"
    local tmp_json
    tmp_json="$(mktemp)"

    local code
    local curl_status
    set +e
    local lookup_meta
    lookup_meta="$(curl -sS --max-time "$TIMEOUT" -o "$tmp_json" -w "%{http_code} %{time_total}" "${BASE_URL%/}/$word")"
    curl_status=$?
    set -e
    code="${lookup_meta%% *}"
    lookup_time="${lookup_meta##* }"
    total=$((total + 1))

    if [[ "$curl_status" -ne 0 ]]; then
      lookup_timeout=$((lookup_timeout + 1))
      [[ "$VERBOSE" == "1" ]] && echo "  lookup_timeout word=$word"
      rm -f "$tmp_json"
      sleep_ms "$delay_ms"
      continue
    fi
    if [[ "$code" == "429" ]]; then
      lookup_429=$((lookup_429 + 1))
      [[ "$VERBOSE" == "1" ]] && echo "  lookup_429 word=$word t=${lookup_time}s"
      rm -f "$tmp_json"
      sleep_ms "$delay_ms"
      continue
    fi
    if [[ "$code" != "200" ]]; then
      lookup_other=$((lookup_other + 1))
      [[ "$VERBOSE" == "1" ]] && echo "  lookup_http_${code} word=$word t=${lookup_time}s"
      rm -f "$tmp_json"
      sleep_ms "$delay_ms"
      continue
    fi
    [[ "$VERBOSE" == "1" ]] && echo "  lookup_ok word=$word t=${lookup_time}s"

    local audio_url
    audio_url="$(jq -r 'if type=="array" then ([.[0].phonetics[]?.audio // empty] | map(select(length>0)) | .[0] // "") else "" end' "$tmp_json" 2>/dev/null || true)"
    rm -f "$tmp_json"
    sleep_ms "$delay_ms"

    if [[ -z "$audio_url" ]]; then
      no_audio=$((no_audio + 1))
      [[ "$VERBOSE" == "1" ]] && echo "  no_audio word=$word"
      continue
    fi

    local audio_code
    local audio_curl_status
    local audio_meta
    set +e
    audio_meta="$(curl -sS --max-time "$TIMEOUT" -o /dev/null -w "%{http_code} %{time_total}" "$audio_url")"
    audio_curl_status=$?
    set -e
    audio_code="${audio_meta%% *}"
    audio_time="${audio_meta##* }"
    total=$((total + 1))
    if [[ "$audio_curl_status" -ne 0 ]]; then
      audio_timeout=$((audio_timeout + 1))
      [[ "$VERBOSE" == "1" ]] && echo "  audio_timeout word=$word"
      sleep_ms "$delay_ms"
      continue
    fi
    if [[ "$audio_code" == "429" ]]; then
      audio_429=$((audio_429 + 1))
      [[ "$VERBOSE" == "1" ]] && echo "  audio_429 word=$word t=${audio_time}s"
    elif [[ "$audio_code" != "200" ]]; then
      audio_other=$((audio_other + 1))
      [[ "$VERBOSE" == "1" ]] && echo "  audio_http_${audio_code} word=$word t=${audio_time}s"
    else
      [[ "$VERBOSE" == "1" ]] && echo "  audio_ok word=$word t=${audio_time}s"
    fi
    sleep_ms "$delay_ms"
  done

  local rl=$((lookup_429 + audio_429))
  local to=$((lookup_timeout + audio_timeout))
  local other=$((lookup_other + audio_other))

  echo "total_http=$total lookup_429=$lookup_429 audio_429=$audio_429 lookup_timeout=$lookup_timeout audio_timeout=$audio_timeout other_http=$other no_audio_entries=$no_audio"
  if [[ "$rl" -eq 0 && "$to" -eq 0 ]]; then
    echo "PASS"
    return 0
  fi
  echo "FAIL"
  return 1
}

best=""
for d in "${DELAYS[@]}"; do
  if test_delay "$d"; then
    best="$d"
    break
  fi
done

echo ""
if [[ -n "$best" ]]; then
  echo "Recommended dictionary throttle: ${best}ms"
  exit 0
fi
echo "No safe delay found in tested range: $DELAYS_CSV"
exit 2
