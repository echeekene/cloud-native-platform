# Security decisions

| Decision | Why |
|---|---|
| DNS and NTP reachable only from `admin_cidrs` | Open resolvers and open NTP servers are abused for reflection and amplification DDoS |
| chrony `ratelimit` and `cmdport 0` | Limits abuse and removes the remote control port |
| Upstream DNS over TLS | Recursive lookups cannot be read or modified in transit |
| IMDSv2 required | Blocks SSRF style theft of instance credentials |
| Encrypted EBS and TLS only state bucket | Data at rest and in transit protection |
| GitHub OIDC role scoped to `main` and one instance | No long lived keys; a forked repo or other branch cannot deploy |
| SSM instead of SSH for deploys | No inbound port needed for CI, every command logged by AWS |
| SSH key only, no root login, fail2ban | Basic host hardening |
| Distroless non root images, read only root FS, all capabilities dropped | Small attack surface inside containers |
| CoreDNS on port 1053 internally | Runs with zero capabilities |
| NetworkPolicies with default deny | Pods only accept the traffic they need |
| Pod Security Standards labels | Restricted for monitoring, baseline for platform (chrony binds 123 as root then drops) |
| k3s secrets encryption and API audit log | Secrets encrypted in etcd, every change to the cluster recorded |
| VPC flow logs (rejected traffic) | Evidence of scanning and blocked access attempts |
| Grafana not exposed | Reached only through `kubectl port-forward` |
| Trivy config and image scans in CI | Catches misconfigurations and critical CVEs before deploy |
