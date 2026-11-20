# Runbook

Every alert in `k8s/monitoring/prometheus/alerts.yml` links to a section here.

Useful commands:

```bash
kubectl -n platform get pods -o wide
kubectl -n platform logs deploy/coredns --tail=100
kubectl -n platform exec deploy/status-api -- statusctl -dns coredns:53 -ntp chrony:123
kubectl -n platform rollout undo deploy/<name>
```

## DNSCheckFailing

**Meaning:** status-api cannot resolve a name through CoreDNS for 2 minutes.

1. Check which check is failing. `coredns-internal-zone` points at the zone file or CoreDNS itself; `coredns-recursive` points at upstream or egress.
2. `kubectl -n platform get pods -l app=coredns`. If pods are not ready, read their logs.
3. Internal zone only: a bad zone file fails to load. Look for `plugin/file` errors in logs and check the SOA serial was bumped.
4. Recursive only: test upstream from the node with `dig @1.1.1.1 ekeneeche.com` and `nc -vz 1.1.1.1 853`.
5. If a recent deploy caused it, `kubectl -n platform rollout undo deploy/coredns`.

## NTPOffsetHigh

**Meaning:** chrony reports more than 100 ms offset for 5 minutes.

1. `kubectl -n platform exec deploy/chrony -- chronyc -n sources` to see which sources are selected.
2. If Amazon Time Sync (169.254.169.123) is unreachable, check the security group egress and that IMDS hop limit is 2.
3. If only pool servers remain, expect higher jitter; offset should still settle under 100 ms within minutes.

## NTPServerDown

**Meaning:** chrony did not answer an SNTP query for 2 minutes.

1. Check pods and readiness. The readiness probe fails until chrony is synchronised, so a fresh pod with no reachable upstream stays unready.
2. `chronyc tracking` inside the pod. `Leap status: Not synchronised` means no upstream is reachable.

## CoreDNSServfailRate

**Meaning:** more than 5% of answers are SERVFAIL.

1. Usually upstream trouble. Check `coredns_forward_healthcheck_failures_total` in Prometheus.
2. Look for a single client hammering a broken name in the CoreDNS logs (`class denial error` logs these).

## CoreDNSLatencyHigh

**Meaning:** p99 latency above 250 ms for 10 minutes.

1. Check the cache hit ratio panel. A sudden drop means many unique names (possible abuse or a misbehaving client).
2. Check node CPU in CloudWatch and `kubectl top pods -n platform`.

## StatusAPIDown

**Meaning:** Prometheus cannot scrape any status-api pod.

1. `kubectl -n platform get pods -l app=status-api`.
2. If pods crash, read logs; a bad env var makes the service exit with `bad config`.
3. Roll back with `kubectl -n platform rollout undo deploy/status-api`.

## Node down

CloudWatch recovers the instance on hardware failure and reboots it on OS failure (see `infra/terraform/monitoring.tf`). Elastic IP, private IP and disk carry over, and k3s restarts all workloads on boot. If the instance is gone entirely, `terraform apply` then `ansible-playbook site.yml` rebuilds it.
