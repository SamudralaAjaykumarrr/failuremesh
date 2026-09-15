# Historical mechanism comparison

The [machine-readable result](result.json) maps the public source claims, local representations, PostgreSQL observations, match statuses, and limitations. `MATCH` is a local historical comparison status, not an execution verdict for GitHub or a customer.

| Public-source claim | FailureMesh representation | Authoritative local observation | Status | Limitation |
| --- | --- | --- | --- | --- |
| Migration and production traffic overlapped | Two workloads in one domain | Migration 2 + production 2 at ticks 2–5 | MATCH | Synthetic units |
| Shared DB connection capacity saturated | Finite primary-db ledger | Capacity 4, occupied 4, available 0 at arrival | MATCH | No GitHub pool telemetry |
| Query/request contention | Bounded waiting requests | Three requests wait at tick 3 | MATCH | Logical wait, not physical query latency |
| Multiple services experienced timeouts | Three request classes and deadlines | Pull, secondary, and dependent requests cross deadlines | MATCH | Generic local classes |
| Shared dependencies caused degradation | Explicit dependency graph | All three affected classes use primary-db | MATCH | No exact GitHub topology |
| Migration pause preceded recovery | Release interval and later request | Capacity available after tick 6; tick-7 request completes | MATCH | No GitHub recovery timing |

**Applicability: APPLICABLE only at SERVICE-LEVEL CASCADE scope. Local historical comparison: MATCH; milestone recommendation: PASS**, conditional on the recorded PostgreSQL run. PASS means only that FailureMesh represented and deterministically reproduced the reported shared-database saturation and timeout-cascade mechanism within this documented bounded local model. See the [source record](source-record.md) for GitHub's reported observations versus causal analysis, follow-up, inference, and unknown telemetry. No exact GitHub reproduction, independent verification of GitHub telemetry, complete path graph, 5xx-rate reproduction, generalized portability, customer readiness, production readiness, PFC #4, or completed expert validation is claimed. Gate A remains REVISE at 8/10 qualifying incidents.
