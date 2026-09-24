#!/usr/bin/env bash
# Run the whole stack locally with Docker and check it works.
set -euo pipefail
cd "$(dirname "$0")/.."

docker compose up --build -d
echo "waiting for chrony to sync..."
sleep 20

dig +short @127.0.0.1 -p 1053 status.platform.internal
curl -fsS http://localhost:8080/api/status | jq .
echo "status page: http://localhost:8080"
