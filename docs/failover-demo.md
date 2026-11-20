# Failover drill

Goal: show that losing a CoreDNS pod does not interrupt DNS.

## How to run

```bash
./scripts/failover-demo.sh <node_public_ip> 60
```

The script queries `status.platform.internal` twice a second for 60 seconds and deletes one CoreDNS pod after 10 seconds.

## Results

Fill this in after running the drill against the live cluster.

| Date | Queries | Failed | Success rate | Notes |
|---|---|---|---|---|
| | | | | |

Add the Grafana screenshot of the DNS panels during the drill to `docs/images/`.

## What to look for

- With two replicas, `maxUnavailable: 0` and readiness probes, failed queries should be zero or very close to it.
- Any failures should line up with the moment the pod was deleted; if they continue, the Service is still routing to the terminating pod and the `lameduck` setting in the Corefile needs a longer value.
