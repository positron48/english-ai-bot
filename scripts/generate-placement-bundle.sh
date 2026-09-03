#!/bin/sh
set -eu
exec python3 "$(dirname "$0")/build-placement-bank.py" "${1:-all}"
