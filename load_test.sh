#!/usr/bin/env bash
set -euo pipefail

N=${1:-5000}   # сколько запросов
C=${2:-100}    # параллельность
URL=${3:-http://localhost:8080/}

echo "Load test: $N requests, parallel $C -> $URL"

# Генерируем N чисел и для каждого запускаем curl в отдельном воркере.
# Важно: используем 'bash -c' чтобы $(date) и $RANDOM вычислялись в момент вызова.
seq 1 "$N" | xargs -P "$C" -n 1 -I{} bash -c '
  curl -sS -X POST -H "Content-Type: text/plain" \
    --data "https://example.com/{}-$(date +%s%N)-$RANDOM" "'"$URL"'" > /dev/null
'
echo "Done."