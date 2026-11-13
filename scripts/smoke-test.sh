#!/usr/bin/env bash
# Quick check of the public DNS and NTP endpoints from your laptop.
#   ./scripts/smoke-test.sh <node-public-ip>
set -euo pipefail
HOST="${1:?usage: smoke-test.sh <node-public-ip>}"

echo "== internal zone"
dig +short @"$HOST" status.platform.internal A

echo "== recursive lookup"
dig +short @"$HOST" ekeneeche.com A

echo "== DNS over TCP"
dig +tcp +short @"$HOST" time.platform.internal A

echo "== NTP"
if command -v ntpdate >/dev/null; then
  ntpdate -q "$HOST"
elif command -v sntp >/dev/null; then
  sntp "$HOST"
else
  echo "install ntpdate or sntp, or run: go run ./services/status-api/cmd/statusctl -dns $HOST:53 -ntp $HOST:123"
fi

echo "== status API"
curl -fsS https://status.ekeneeche.com/api/status | jq '{overall, checks: [.results[] | {name, status, latency_ms}]}'
