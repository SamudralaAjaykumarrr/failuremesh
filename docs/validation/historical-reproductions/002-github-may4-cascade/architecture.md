# Incident architecture and AC

The incident-specific AC `historical.github-may4-cascade@1` is scoped to **SERVICE-LEVEL CASCADE** in a local synthetic build. The AC uses the existing `Fact` evidence, source, scope, quality, and sensitivity fields to separate two prerequisite groups:

| Support | AC prerequisites | Classification |
| --- | --- | --- |
| GitHub first-party causal account | `shared_critical_db_dependency`, `finite_capacity_domain`, `online_migration_consumes_capacity`, `production_traffic_consumes_same_capacity`, `migration_traffic_overlap`, `capacity_can_saturate`, `latency_can_exceed_deadline`, `multiple_services_share_dependency`, `migration_can_be_paused` | `quality: reported`, public source URL, qualitative service-level scope; no direct GitHub telemetry |
| FailureMesh local implementation | `requests_wait_on_capacity`, `timeout_state_observable`, `capacity_state_observable`, `local_dependency_graph` | `quality: direct`, local model/schema source, synthetic sensitivity; directly supported fixture capabilities chosen through FailureMesh inference |

The matcher requires both the reported qualitative incident prerequisites and the separately supported local experiment/observability capabilities. Missing, stale, conflicting, insufficient, or wrongly classified facts yield UNKNOWN; evidenced absence yields NOT_APPLICABLE. Service-level applicability remains a FailureMesh inference. Exact public pool size, occupancy count, queue lengths, deadlines, complete per-request topology, complete traces, and retry policy remain visible UNKNOWNs in the AC. Local numbers, waiting representation, graph, and state observability do not become GitHub-reported facts.

The local graph is synthetic:

```text
migration-M ---------\
normal-production ----> primary-db (one finite capacity domain)
                          |-- pull-request-service
                          |-- secondary-service
                          `-- dependent-service
```

Edges mean only that the local actors use the same modeled capacity domain. They do not assert GitHub's internal service graph or a Pull Requests → Copilot edge. PostgreSQL stores the domain, workloads, history, graph edges, request states, and trace. The authoritative capacity ledger and request rows are checked together. No real connection pool is exhausted.

**SYNTHETIC LOCAL MODEL PARAMETERS:** capacity 4 units; migration occupancy 2 from logical tick 1 until release at 6; normal production occupancy 2 from tick 2 through 7; three affected requests arrive at tick 3; logical deadlines are ticks 4 or 5; a recovery request arrives and completes at tick 7. Units and ticks are arbitrary logical values. No GitHub count, clock, deadline, or physical latency is inferred from them. Safety is level 1 with ten fixed trace steps, four fixed requests, no sleeps, no random scheduler, and no external endpoints.
