#!/usr/bin/env bash
# Failover drill: query DNS twice a second while killing a CoreDNS pod,
# then report how many queries failed. Run on the node (or anywhere with
# kubectl access) and keep the output for docs/failover-demo.md.
#   ./scripts/failover-demo.sh <dns-ip> [seconds]
set -uo pipefail
DNS="${1:?usage: failover-demo.sh <dns-ip> [seconds]}"
DURATION="${2:-60}"
NAME="status.platform.internal"
LOG="failover-$(date +%Y%m%d-%H%M%S).log"

ok=0; fail=0
end=$(( $(date +%s) + DURATION ))

(
  sleep 10
  victim=$(kubectl -n platform get pod -l app=coredns -o jsonpath='{.items[0].metadata.name}')
  echo "$(date +%T) deleting pod $victim" | tee -a "$LOG"
  kubectl -n platform delete pod "$victim" --wait=false >/dev/null
) &

echo "querying $NAME via $DNS for ${DURATION}s, log: $LOG"
while [ "$(date +%s)" -lt "$end" ]; do
  if answer=$(dig +short +time=1 +tries=1 @"$DNS" "$NAME" A) && [ -n "$answer" ]; then
    ok=$((ok + 1)); echo "$(date +%T.%3N) ok $answer" >> "$LOG"
  else
    fail=$((fail + 1)); echo "$(date +%T.%3N) FAIL" | tee -a "$LOG"
  fi
  sleep 0.5
done
wait

total=$((ok + fail))
echo
echo "queries: $total  ok: $ok  failed: $fail"
awk -v o="$ok" -v t="$total" 'BEGIN { printf "success rate: %.2f%%\n", (o / t) * 100 }'
