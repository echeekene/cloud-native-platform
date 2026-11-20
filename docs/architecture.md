# Architecture

## Goals

1. Run real DNS and NTP services, not mocks, with the same concerns a cloud provider has: availability, latency, correctness and abuse.
2. Keep cost near zero, so a single node instead of a multi node cluster.
3. Make every change go through code review and an automated pipeline.

## Components

**CoreDNS** runs two replicas. It serves the `platform.internal` zone from a zone file and answers everything else from a cache, forwarding misses to Cloudflare over DNS over TLS so upstream lookups cannot be read or tampered with on the wire. It listens on 1053 inside the container so it needs no Linux capabilities; the Service maps 53 to it.

**chrony** runs two replicas. It syncs from Amazon Time Sync (a link local, highly accurate source inside AWS) with public NTS pools as backup, and serves time to clients. It runs with `-x` so it never changes the node clock, which removes the need for the `SYS_TIME` capability.

**status-api** is a Go service. Every 15 seconds it:

- resolves `status.platform.internal` against CoreDNS (authoritative path)
- resolves `ekeneeche.com` against CoreDNS (recursive and upstream path)
- sends an SNTP query to chrony and measures clock offset

Results are served three ways: an HTML status page, `/api/status` JSON, and `/metrics` for Prometheus. The SNTP client and the metrics output are written with the standard library only, so the binary has no third party dependencies.

**Prometheus and Grafana** run in a separate `monitoring` namespace. Prometheus discovers pods by annotation and evaluates the alert rules in `k8s/monitoring/prometheus/alerts.yml`. Grafana is not exposed publicly.

## Why single node

A three node cluster across availability zones is the right answer in production, and the manifests already spread replicas with topology constraints so they would do that unchanged. For a portfolio, one node plus pod level redundancy plus EC2 auto recovery covers the failure modes worth demonstrating at a fraction of the cost. The tradeoff is that a full node loss means a few minutes of downtime while AWS recovers the instance.

## Request flow

1. A visitor opens `https://status.ekeneeche.com`.
2. Cloudflare terminates TLS, then reconnects to the node over TLS using a Cloudflare Origin CA certificate (Full strict mode).
3. Traefik routes to a ready `status-api` pod.
4. The page polls `/api/status` every 15 seconds.

## Deploy flow

1. A PR runs Platform CI: Go tests, Terraform validate, kubeconform, promtool, ansible-lint, Trivy, image builds.
2. A merge to `main` reruns CI. On success, Platform deploy builds and pushes images tagged with the commit SHA to GHCR and scans them.
3. The deploy job gets short lived AWS credentials through OIDC. The role trusts only `main` of this repo and can only run `AWS-RunShellScript` on this one instance.
4. SSM runs `platform-deploy <sha>` on the node, which checks out that commit, pins the image tags with a Kustomize overlay, applies, waits for rollouts and runs `statusctl` as a smoke test.
5. The workflow finishes with a public smoke test against the status API.

No SSH keys or AWS access keys are stored in GitHub.
