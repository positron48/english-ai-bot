#!/bin/bash
cd /Users/antonfilatov/www/my/k3s/english-ai-bot
set -a
source .env
set +a
exec ./bin/english-ai-bot
