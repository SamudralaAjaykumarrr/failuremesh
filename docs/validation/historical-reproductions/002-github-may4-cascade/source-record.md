# Historical Reproduction #002 — public source record

Primary source: [GitHub availability report: May 2026](https://github.blog/news-insights/company-news/github-availability-report-may-2026/), published June 11, 2026. This is a first-party public availability report about May 4, not FailureMesh telemetry. The existing [family-8 Gate A record](../../gate-a/incidents/08-dependency-latency-cascade.md) remains unchanged.

## A. Reported observations

GitHub reports customer impact approximately 15:34–16:40 UTC, elevated github.com latency and request failures, pull requests most significantly affected, and elevated latency or intermittent errors for Issues, Actions, webhooks, and Git operations. It reports varying degradation of Codespaces, Pages, Packages, OAuth and GitHub Apps, Marketplace, and Copilot due to shared data dependencies. Reported peak 5xx rate was approximately 1.3%, with an average of approximately 0.46%. Responders paused the contributing migration; dependent services recovered shortly afterward. GitHub reports about 33 minutes to mitigation and full resolution about 30 minutes later. These are GitHub-reported service-level observations, not independently audited request traces.

## B. GitHub's reported causal analysis

A routine online schema migration of a large, heavily accessed table had run for several hours. Traffic then increased toward a weekly peak. GitHub attributes the disruption to combined migration and production load saturating database connection capacity, causing query contention on a primary database and cascading timeouts across dependent services. This is a first-party causal account, not a FailureMesh direct observation of GitHub infrastructure.

## C. GitHub's follow-up actions

GitHub reports tighter low-traffic scheduling for large-table migrations, dynamic throttling by live cluster load, circuit breakers for migration-induced latency or connection utilization, migration-pressure monitoring, and review of connection-pool headroom. These are reported plans/actions, not controls implemented by this reproduction.

## D. FailureMesh inference

A bounded incident-consistent experiment needs a finite shared capacity domain, overlapping migration and production demand, waiting requests from multiple service classes sharing that domain, logical deadline crossing, and recovery after migration release. The local graph, capacity counts, actors, requests, and deadlines are **SYNTHETIC LOCAL MODEL PARAMETERS**. Service-level applicability is an inference from GitHub's reported mechanism; it does not establish an exact request path.

## E. Unknown / unsupported details

Exact pool size, occupied connections, migration demand, production demand, per-request queue lengths and deadlines, retry policies, raw pool telemetry, complete per-hop traces, full service graph, exact interleaving, and proof that every affected request used an identical dependency path are UNKNOWN. The local model never treats its numbers or graph as GitHub facts.
